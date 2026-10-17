package ports

import "github.com/MahdiRohani/request-meter/internal/core/domain"

type CounterService interface {
	Increment(name string) error
	GetStats(name string) (*domain.Counter, error)
}

type CounterRepository interface {
	Increment(name string) error
	Get(name string) (*domain.Counter, error)
	GetAll() ([]*domain.Counter, error)
	Save(counter *domain.Counter) error
}
