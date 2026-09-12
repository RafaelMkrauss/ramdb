//defer helps to prevent deadlocks
//

package db

import (
	"errors"
	"sync"
)

var ErrKeyNotFound = errors.New("chave não encontrada")

type Engine struct {
	data map[string]string // data storage
	mu   sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		data: make(map[string]string),
	}
}

// set/update
func (e *Engine) Set(key string, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.data[key]; !exists {
		if len(e.data) >= 100000 {
			return errors.New("banco de dados cheio (OOM)")
		}
	}
	e.data[key] = value
	return nil
}

func (e *Engine) Get(key string) (string, error) {
	e.mu.RLock() //read
	defer e.mu.RUnlock()

	value, exists := e.data[key]
	if !exists {
		return "", ErrKeyNotFound
	}
	return value, nil
}

// Delete
func (e *Engine) Delete(key string) {
	e.mu.Lock() // write
	defer e.mu.Unlock()
	delete(e.data, key)
}
