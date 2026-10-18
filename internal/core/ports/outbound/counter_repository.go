package outbound

import "github.com/MahdiRohani/request-meter/internal/core/domain"

type FastStoragePort interface {
	Increment(name string) error
	Get(name string) (*domain.Counter, error)
	GetAll() []*domain.Counter
	Save(counter *domain.Counter) error
}

type SlowStoragePort interface {
	Save(counter *domain.Counter) error
	GetAll() []*domain.Counter
}
