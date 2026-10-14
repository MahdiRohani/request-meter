package outbound

import (
	"sync"

	"github.com/MahdiRohani/request-meter/internal/core/domain"
)

type InMemoryRepository struct {
	mu       sync.RWMutex
	counters map[string]*domain.Counter
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		counters: make(map[string]*domain.Counter),
	}
}

func (r *InMemoryRepository) Increment(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.counters[name]; !exists {
		r.counters[name] = domain.NewCounter(name)
	}
	r.counters[name].Increment()
	return nil
}

func (r *InMemoryRepository) Get(name string) (*domain.Counter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counter, exists := r.counters[name]
	if !exists {
		return domain.NewCounter(name), nil
	}
	return &domain.Counter{
		Name:  counter.Name,
		Count: counter.Count,
	}, nil
}
