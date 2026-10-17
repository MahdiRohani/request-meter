package inbound

import "github.com/MahdiRohani/request-meter/internal/core/domain"

type CounterService interface {
	Increment(name string) error
	GetStats(name string) (*domain.Counter, error)
}
