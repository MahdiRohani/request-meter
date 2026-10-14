package services

import (
	"github.com/MahdiRohani/request-meter/internal/core/domain"
	"github.com/MahdiRohani/request-meter/internal/core/ports"
)

type counterService struct {
	fastRepo ports.CounterRepository
}

func NewCounterService(fastRepo ports.CounterRepository) ports.CounterService {
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
