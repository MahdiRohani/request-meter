package outbound

import "github.com/MahdiRohani/request-meter/internal/core/domain"

type CounterRepository interface {
	Increment(name string) error
	Get(name string) (*domain.Counter, error)
	GetAll() []*domain.Counter
	Save(counter *domain.Counter) error
}
