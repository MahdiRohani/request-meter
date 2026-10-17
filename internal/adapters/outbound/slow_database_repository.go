package outbound

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/MahdiRohani/request-meter/internal/core/domain"
)

type SlowDatabaseRepository struct {
	mu       sync.RWMutex
	counters map[string]*domain.Counter
}

func NewSlowDatabaseRepository() *SlowDatabaseRepository {
	return &SlowDatabaseRepository{
		counters: make(map[string]*domain.Counter),
	}
}

func (s *SlowDatabaseRepository) simulateLatency() {
	delay := time.Duration(rand.Intn(500)+500) * time.Millisecond
	time.Sleep(delay)
}

func (s *SlowDatabaseRepository) Increment(name string) error {
	s.simulateLatency()
	return nil
}

func (s *SlowDatabaseRepository) Get(name string) (*domain.Counter, error) {
	s.simulateLatency()
	return nil, nil
}

func (s *SlowDatabaseRepository) GetAll() []*domain.Counter {
	s.simulateLatency()
	return nil
}

func (s *SlowDatabaseRepository) Save(counter *domain.Counter) error {
	s.simulateLatency()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[counter.Name] = &domain.Counter{
		Name:  counter.Name,
		Count: counter.Count,
	}
	log.Printf("[Database] Saved: %s = %d\n", counter.Name, counter.Count)
	return nil
}
