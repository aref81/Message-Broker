package broker

import (
	"context"
	"sync"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

type Module struct {
	mu          sync.Mutex
	isClosed    bool
	subscribers map[string]map[chan model.Message]struct{}
	messages    map[string][]model.Pair
	nextID      map[string]int
}

func NewModule() broker.Broker {
	return &Module{
		mu:          sync.Mutex{},
		isClosed:    false,
		subscribers: make(map[string]map[chan model.Message]struct{}),
		messages:    make(map[string][]model.Pair),
		nextID:      make(map[string]int),
	}
}

func (m *Module) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for subject, subscribers := range m.subscribers {
		for ch := range subscribers {
			close(ch)
		}
		delete(m.subscribers, subject)
	}
	m.isClosed = true

	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg model.Message) (int, error) {
	if m.isClosed {
		return -1, broker.ErrUnavailable
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// this line caused trouble in test : TestPublishShouldPreserveOrder
	//msg.Id = m.nextID[subject]
	m.nextID[subject]++

	pair := model.Pair{
		msg,
		time.Now(),
	}

	m.messages[subject] = append(m.messages[subject], pair)

	subscribers := m.subscribers[subject]
	for ch := range subscribers {
		select {
		case ch <- msg:
		default:
			// skip subscribers that are not ready to receive -> do nothing
		}
	}

	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan model.Message, error) {
	if m.isClosed {
		return nil, broker.ErrUnavailable
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan model.Message, 100)
	subscribers, ok := m.subscribers[subject]
	if !ok {
		subscribers = make(map[chan model.Message]struct{})
		m.subscribers[subject] = subscribers
	}
	subscribers[ch] = struct{}{}

	go func() {
		<-ctx.Done()
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(subscribers, ch)
		close(ch)
	}()

	return ch, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (model.Message, error) {
	if m.isClosed {
		return model.Message{}, broker.ErrUnavailable
	}

	messages, ok := m.messages[subject]
	if !ok || id < 0 || id >= len(messages) {
		return model.Message{}, broker.ErrInvalidID
	}

	pair := messages[id]
	if pair.Message.Expiration > 0 && time.Since(pair.Sent) > pair.Message.Expiration {
		return model.Message{}, broker.ErrExpiredID
	}

	return pair.Message, nil
}
