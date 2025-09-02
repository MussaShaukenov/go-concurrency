package engine

import (
	"errors"
	"go.uber.org/zap"
)

var ErrKeyNotFound = errors.New("key not found")

type Engine struct {
	storage map[any]any
	logger  *zap.SugaredLogger
}

func NewEngine(logger *zap.SugaredLogger) *Engine {
	return &Engine{
		storage: make(map[any]any),
		logger:  logger,
	}
}

func (e *Engine) Set(key, value any) {
	e.storage[key] = value
	e.logger.Infof("storage size: %d entries", len(e.storage))
}

func (e *Engine) Get(key any) (any, error) {
	value, ok := e.storage[key]
	if !ok {
		e.logger.Warnf("key '%v' not found", key)
		return nil, ErrKeyNotFound
	}

	return value, nil
}

func (e *Engine) Delete(key any) error {
	_, ok := e.storage[key]
	if !ok {
		e.logger.Warnf("attempted to delete non-existent key '%v'", key)
		return ErrKeyNotFound
	}

	delete(e.storage, key)
	e.logger.Infof("key '%v' deleted. Storage size: %d entries", key, len(e.storage))

	return nil
}
