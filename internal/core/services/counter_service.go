package services

import (
	"github.com/MahdiRohani/request-meter/internal/core/domain"
	"github.com/MahdiRohani/request-meter/internal/core/ports/inbound"
	"github.com/MahdiRohani/request-meter/internal/core/ports/outbound"
)

type counterService struct {
	fastRepo outbound.FastStoragePort
	slowRepo outbound.SlowStoragePort
}

func NewCounterService(fastRepo outbound.FastStoragePort, slowRepo outbound.SlowStoragePort) inbound.CounterService {
	return &counterService{
		fastRepo: fastRepo,
		slowRepo: slowRepo,
	}
}

func (c *counterService) Increment(name string) error {
	return c.fastRepo.Increment(name)
}

func (c *counterService) GetStats(name string) (*domain.Counter, error) {
	return c.fastRepo.Get(name)
}
