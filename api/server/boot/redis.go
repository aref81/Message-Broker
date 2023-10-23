package boot

import (
	"context"
	"github.com/go-redis/redis/v8"
)

const (
	address  = "redis-yes"
	port     = "6379"
	password = ""
	db       = 0
	poolSize = 10000
)

func InitRedis() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     address + ":" + port,
		Password: password,
		PoolSize: poolSize,
		DB:       db,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return &redis.Client{}, err
	}
	return redisClient, nil
}
