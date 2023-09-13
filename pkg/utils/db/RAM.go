package db

import (
	"github.com/sirupsen/logrus"
	"os"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

type RAM struct {
	Dbms
	messages map[string][]model.Pair
	nextID   map[string]int
}

func InitRAM() (*RAM, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Configure logrus to output to console
	logger.Out = os.Stdout

	return &RAM{
		messages: make(map[string][]model.Pair),
		nextID:   make(map[string]int),
	}, nil
}

func (r *RAM) SendMessage(message model.Message, subject string) (int, error) {
	// this line caused trouble in test : TestPublishShouldPreserveOrder
	//msg.Id = m.nextID[subject]

	Id := r.nextID[subject]

	pair := model.Pair{
		Message: message,
		Sent:    time.Now(),
	}
	r.nextID[subject]++
	r.messages[subject] = append(r.messages[subject], pair)

	return Id, nil
}

func (r *RAM) FetchMessage(messageId int, subject string) (model.Message, error) {
	pairs, ok := r.messages[subject]
	if !ok || messageId < 0 || messageId >= len(pairs) {
		return model.Message{}, broker.ErrInvalidID
	}

	pair := pairs[messageId]
	if pair.Message.Expiration > 0 && time.Since(pair.Sent) > pair.Message.Expiration {
		return model.Message{}, broker.ErrExpiredID
	}

	return pair.Message, nil
}
