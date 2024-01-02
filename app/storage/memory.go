package storage

import (
	"errors"
	"sync"
)

type Memory struct {
	data      map[string][]byte
	dataMutex sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		data:      make(map[string][]byte),
		dataMutex: sync.RWMutex{},
	}
}

func (m *Memory) Name() string {
	return "memory"
}

func (m *Memory) Store(key string, value interface{}) error {
	m.dataMutex.Lock()
	m.data[key] = value.([]byte)
	m.dataMutex.Unlock()
	return nil
}

func (m *Memory) Retrieve(key string) ([]byte, error) {
	m.dataMutex.RLock()
	if val, ok := m.data[key]; ok {
		m.dataMutex.RUnlock()
		return val, nil
	}
	m.dataMutex.RUnlock()
	// TODO: return a provider-generic error when key not found
	return nil, errors.New("redis: nil")
}
