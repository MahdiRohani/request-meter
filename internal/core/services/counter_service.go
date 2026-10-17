package services

import (
	"github.com/MahdiRohani/request-meter/internal/core/domain"
	"github.com/MahdiRohani/request-meter/internal/core/ports/inbound"
	"github.com/MahdiRohani/request-meter/internal/core/ports/outbound"
)

type counterService struct {
	fastRepo outbound.CounterRepository
}

func NewCounterService(fastRepo outbound.CounterRepository) inbound.CounterService {
	return &counterService{
		fastRepo: fastRepo,
	}
}

func (c *counterService) Increment(name string) error {
	return c.fastRepo.Increment(name)
}

func (c *counterService) GetStats(name string) (*domain.Counter, error) {
	return c.fastRepo.Get(name)
}
