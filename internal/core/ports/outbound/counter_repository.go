package outbound

import "github.com/MahdiRohani/request-meter/internal/core/domain"

type FastRepositoryPort interface {
	Increment(name string) error
	Get(name string) (*domain.Counter, error)
	GetAll() []*domain.Counter
	Save(counter *domain.Counter) error
	ResetFlushed(flushedCounters []*domain.Counter)
}

type SlowRepositoryPort interface {
	SaveBatch(counters []*domain.Counter) error
	GetAll() []*domain.Counter
}
