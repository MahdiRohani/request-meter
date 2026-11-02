package outbound

import (
	"context"
	"sync"

	"github.com/MahdiRohani/request-meter/internal/domain"
)

type InMemoryRepository struct {
	mu           sync.RWMutex
	deltas       map[string]*domain.Counter
	cachedTotals map[string]*domain.Counter
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		deltas:       make(map[string]*domain.Counter),
		cachedTotals: make(map[string]*domain.Counter),
	}
}

func (r *InMemoryRepository) Increment(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.deltas[name]; !exists {
		r.deltas[name] = domain.NewCounter(name)
	}
	r.deltas[name].Increment()
	return nil
}

func (r *InMemoryRepository) Get(ctx context.Context, name string) (*domain.Counter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counter, exists := r.cachedTotals[name]
	if !exists {
		return nil, nil
	}
	return &domain.Counter{
		Name:  counter.Name,
		Count: counter.Count,
	}, nil
}

func (r *InMemoryRepository) GetAll(ctx context.Context) []*domain.Counter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var snapshot []*domain.Counter
	for _, counter := range r.deltas {
		snapshot = append(snapshot, &domain.Counter{
			Name:  counter.Name,
			Count: counter.Count,
		})
	}
	return snapshot
}

func (r *InMemoryRepository) SetBatch(ctx context.Context, counters []*domain.Counter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, counter := range counters {
		r.cachedTotals[counter.Name] = &domain.Counter{
			Name:  counter.Name,
			Count: counter.Count,
		}
	}

}

func (r *InMemoryRepository) ResetFlushed(ctx context.Context, flushed []*domain.Counter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, f := range flushed {

		if current, exists := r.deltas[f.Name]; exists {

			current.Count -= f.Count

			if current.Count <= 0 {
				delete(r.deltas, f.Name)
			}
		}
	}
}
