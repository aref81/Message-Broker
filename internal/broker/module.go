package broker

import (
	"context"
	"sync"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/utils/db"
)

type Module struct {
	mu          sync.Mutex
	dbms        db.Dbms
	isClosed    bool
	subscribers map[string]map[chan model.Message]struct{}
	//messages    map[string][]model.Message
	//nextID      map[string]int
}

func NewModule() broker.Broker {
	dbms, err := db.InitCassandra()
	if err != nil {
		return nil
	}

	return &Module{
		mu:          sync.Mutex{},
		dbms:        dbms,
		isClosed:    false,
		subscribers: make(map[string]map[chan model.Message]struct{}),
		//messages:    make(map[string][]model.Message),
		//nextID:      make(map[string]int),
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

	if err := m.dbms.Close(); err != nil {
		return err
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
	//m.nextID[subject]++

	//m.messages[subject] = append(m.messages[subject], pair)
	messageID, err := m.dbms.SendMessage(msg, subject)
	if err != nil {
		return -1, err
	}

	subscribers := m.subscribers[subject]
	for ch := range subscribers {
		select {
		case ch <- msg:
		default:
			// skip subscribers that are not ready to receive -> do nothing
		}
	}

	return messageID, nil
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

	msg, err := m.dbms.FetchMessage(id, subject)

	if err != nil {
		return msg, err
	}

	return msg, nil
}
