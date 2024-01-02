package app

import (
	"log/slog"
)

type StorageProvider interface {
	Name() string
	Retrieve(key string) ([]byte, error)
	Store(key string, value interface{}) error
}

type Context struct {
	Log          *slog.Logger
	Storage      StorageProvider
	IsProduction bool
	SentryDSN    string
}
