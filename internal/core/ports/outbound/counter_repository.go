package outbound

import "github.com/MahdiRohani/request-meter/internal/core/domain"

type FastRepositoryPort interface {
	Increment(name string) error
	GetCachedTotal(name string) (*domain.Counter, error)
	GetAll() []*domain.Counter
	SetBatchCachedTotals(counters []*domain.Counter)
	ResetFlushed(flushedCounters []*domain.Counter)
}

type SlowRepositoryPort interface {
	SaveBatch(counters []*domain.Counter) error
	GetAll() []*domain.Counter
}
