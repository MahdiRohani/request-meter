package outbound

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/MahdiRohani/request-meter/internal/domain"
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

func (s *SlowDatabaseRepository) GetAll() []*domain.Counter {
	s.simulateLatency()

	s.mu.RLock()
	defer s.mu.RUnlock()

	var snapshot []*domain.Counter
	for _, val := range s.counters {
		snapshot = append(snapshot, &domain.Counter{
			Name:  val.Name,
			Count: val.Count,
		})
	}
	return snapshot
}

func (s *SlowDatabaseRepository) SaveBatch(counters []*domain.Counter) error {
	s.simulateLatency()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, counter := range counters {
		s.counters[counter.Name] = &domain.Counter{
			Name:  counter.Name,
			Count: counter.Count,
		}
		log.Printf("[Database] Saved: %s = %d\n", counter.Name, counter.Count)
	}

	return nil
}
