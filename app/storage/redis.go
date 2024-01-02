package storage

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis() *Redis {
	return &Redis{}
}

func (r *Redis) Name() string {
	return "redis"
}

func (r *Redis) Connect() error {
	redisHost, exists := os.LookupEnv("REDIS_HOST")
	if !exists {
		return errors.New("redis: no host provided")
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:        redisHost,
		Password:    "",
		DB:          0,
		PoolTimeout: 4 * time.Second,
	})
	cpErr := r.client.Ping(context.Background()).Err()
	return cpErr
}

func (r *Redis) Store(key string, value interface{}) error {
	return r.client.Set(context.Background(), key, value, 0).Err()
}

func (r *Redis) Retrieve(key string) ([]byte, error) {
	return r.client.Get(context.Background(), key).Bytes()
}
