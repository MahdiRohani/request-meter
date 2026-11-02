package outbound

import (
	"context"

	"github.com/MahdiRohani/request-meter/internal/domain"
)

type Closable interface {
	Close(ctx context.Context) error
}

type FastRepositoryPort interface {
	Increment(ctx context.Context, name string) error
	Get(ctx context.Context, name string) (*domain.Counter, error)
	GetAll(ctx context.Context) []*domain.Counter
	SetBatch(ctx context.Context, counters []*domain.Counter)
	ResetFlushed(ctx context.Context, flushedCounters []*domain.Counter)
}

type SlowRepositoryPort interface {
	Closable
	SaveBatch(ctx context.Context, counters []*domain.Counter) error
	GetAll(ctx context.Context) []*domain.Counter
}
