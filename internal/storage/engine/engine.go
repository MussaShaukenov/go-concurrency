package engine

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

var ErrKeyNotFound = errors.New("key not found")

type Engine struct {
	logger *zap.SugaredLogger

	mu      sync.Mutex
	storage map[any]any
}

func NewEngine(logger *zap.SugaredLogger) *Engine {
	return &Engine{
		storage: make(map[any]any),
		logger:  logger,
	}
}

func (e *Engine) Set(key, value any) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.storage[key] = value
	e.logger.Infof("storage size: %d entries", len(e.storage))
}

func (e *Engine) Get(key any) (any, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	value, ok := e.storage[key]
	if !ok {
		e.logger.Warnf("key '%v' not found", key)
		return nil, ErrKeyNotFound
	}

	return value, nil
}

func (e *Engine) Delete(key any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.storage, key)

	e.logger.Infof("key '%v' deleted. Storage size: %d entries", key, len(e.storage))

	return nil
}
