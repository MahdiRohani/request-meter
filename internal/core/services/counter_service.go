package services

import (
	"context"
	"log"
	"time"

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

func (c *counterService) StartSync(cnt context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-cnt.Done():
				log.Printf("Background worker started. Syncing every %v...\n", interval)
				return
			case <-ticker.C:
				c.syncToSlowDB()
			}
		}
	}()

}

func (c *counterService) syncToSlowDB() {
	snapShot := c.fastRepo.GetAll()

	if len(snapShot) == 0 {
		return
	}

	for _, counter := range snapShot {
		err := c.slowRepo.Save(counter)
		if err != nil {
			log.Printf("Failed to save counter %s: %v\n", counter.Name, err)
		}
	}
}
