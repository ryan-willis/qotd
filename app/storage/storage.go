package storage

import (
	"os"

	"github.com/ryan-willis/qotd/app"
)

func Connect() (app.StorageProvider, error) {
	_, exists := os.LookupEnv("REDIS_HOST")
	if exists {
		redis := NewRedis()
		return redis, redis.Connect()
	}
	return NewMemory(), nil
}
