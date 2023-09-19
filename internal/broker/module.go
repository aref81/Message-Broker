package broker

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"sync"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/utils/db"
)

type Module struct {
	subsMu      sync.Mutex
	dbms        db.Dbms
	isClosed    bool
	subscribers map[string]map[chan model.Message]struct{}
}

func NewModule(dbType int) broker.Broker {
	logger := logrus.New()

	var dbms db.Dbms
	var err error

	switch dbType {
	case db.INMEM:
		dbms, err = db.InitRAM()
	case db.POSTGRES:
		dbms, err = db.InitPostgresql()
	case db.CASSANDRA:
		dbms, err = db.InitCassandra()
	case db.SCYLLA:
		dbms, err = db.InitScylla()
	}

	if err != nil {
		logger.Error(err.Error())
		return nil
	}
	logger.Println("Connected to DBMS Successfully")

	return &Module{
		subsMu:      sync.Mutex{},
		dbms:        dbms,
		isClosed:    false,
		subscribers: make(map[string]map[chan model.Message]struct{}),
	}
}

func (m *Module) Close() error {
	m.subsMu.Lock()
	defer m.subsMu.Unlock()

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

	span, _ := opentracing.StartSpanFromContext(ctx, "publish: module")
	defer span.Finish()
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

	m.subsMu.Lock()
	defer m.subsMu.Unlock()

	ch := make(chan model.Message, 100)
	subscribers, ok := m.subscribers[subject]
	if !ok {
		subscribers = make(map[chan model.Message]struct{})
		m.subscribers[subject] = subscribers
	}
	subscribers[ch] = struct{}{}

	go func() {
		<-ctx.Done()
		m.subsMu.Lock()
		defer m.subsMu.Unlock()
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
