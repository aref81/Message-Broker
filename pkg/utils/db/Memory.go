package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"therealbroker/internal/broker/model"
	"therealbroker/pkg/broker"
	"time"
)

const (
	address       = "redis-yes"
	RedisPort     = "6379"
	RedisPassword = ""
	poolSize      = 10000
	db            = 1
)

type Memory struct {
	Dbms
	client *redis.Client
}

func InitRAM() (*Memory, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     address + ":" + RedisPort,
		Password: RedisPassword,
		PoolSize: poolSize,
		DB:       db,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return &Memory{
		client: redisClient,
	}, nil
}

func (m *Memory) SendMessage(message model.Message, subject string) (int, error) {
	// this line caused trouble in test : TestPublishShouldPreserveOrder
	//msg.Id = m.nextID[subject]

	item := model.RedisMessageItem{
		Body:       message.Body,
		Expiration: time.Now().Add(message.Expiration),
	}

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return 0, err
	}

	_, err = m.client.LPush(context.Background(), subject, itemJSON).Result()
	if err != nil {
		return 0, err
	}

	length, err := m.client.LLen(context.Background(), subject).Result()
	if err != nil {
		return 0, err
	}

	id := int(length) - 1

	return id, nil
}

func (m *Memory) FetchMessage(messageId int, subject string) (model.Message, error) {
	reply, err := m.client.LIndex(context.Background(), subject, int64(messageId)).Result()
	if err == redis.Nil {
		return model.Message{}, broker.ErrInvalidID
	} else if err != nil {
		return model.Message{}, err
	}

	var item model.RedisMessageItem
	if err := json.Unmarshal([]byte(reply), &item); err != nil {
		return model.Message{}, err
	}

	if time.Now().After(item.Expiration) {
		return model.Message{}, broker.ErrExpiredID
	}

	msg := model.Message{
		Id:         messageId,
		Body:       item.Body,
		Expiration: 0,
	}

	return msg, nil
}
