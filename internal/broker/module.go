package broker

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
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
	redisClient *redis.Client
	isClosed    bool
	//subscribers map[string]map[chan model.Message]struct{}
}

func NewModule(redisClient *redis.Client, dbType int) broker.Broker {
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
		redisClient: redisClient,
		isClosed:    false,
		//subscribers: make(map[string]map[chan model.Message]struct{}),
	}
}

func (m *Module) Close() error {
	m.subsMu.Lock()
	defer m.subsMu.Unlock()

	if err := m.redisClient.Close(); err != nil {
		return err
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

	//span, _ := opentracing.StartSpanFromContext(ctx, "publish: module")
	//defer span.Finish()

	messageID, err := m.dbms.SendMessage(msg, subject)
	if err != nil {
		return -1, err
	}

	msg.Id = messageID

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return -1, err
	}

	err = m.redisClient.Publish(ctx, subject, msgJSON).Err()
	if err != nil {
		return -1, err
	}

	return messageID, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan model.Message, error) {
	if m.isClosed {
		return nil, broker.ErrUnavailable
	}

	span, _ := opentracing.StartSpanFromContext(ctx, "subscribe: module")
	defer span.Finish()

	m.subsMu.Lock()
	defer m.subsMu.Unlock()

	ch := make(chan model.Message, 100)

	pubsub := m.redisClient.Subscribe(ctx, subject)
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		defer pubsub.Close()

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				close(ch)
				return
			}

			var message model.Message
			err = json.Unmarshal([]byte(msg.Payload), &message)
			if err != nil {
				continue
			}

			select {
			case ch <- message:
			default:
				// Skip subscribers that are not ready to receive (buffer full)
			}
		}
	}()

	return ch, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (model.Message, error) {
	if m.isClosed {
		return model.Message{}, broker.ErrUnavailable
	}

	span, _ := opentracing.StartSpanFromContext(ctx, "fetch: module")
	defer span.Finish()

	msg, err := m.dbms.FetchMessage(id, subject)

	if err != nil {
		return msg, err
	}

	return msg, nil
}
