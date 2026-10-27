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
	fastRepo outbound.FastRepositoryPort
	slowRepo outbound.SlowRepositoryPort
	lock     outbound.DistributedLockPort
}

func NewCounterService(
	fastRepo outbound.FastRepositoryPort,
	slowRepo outbound.SlowRepositoryPort,
	lock outbound.DistributedLockPort) inbound.CounterService {

	service := &counterService{
		fastRepo: fastRepo,
		slowRepo: slowRepo,
		lock:     lock,
	}

	existingCounters := service.slowRepo.GetAll()
	service.fastRepo.SetBatchCachedTotals(existingCounters)
	return service
}

func (c *counterService) Increment(name string) error {
	return c.fastRepo.Increment(name)
}

func (c *counterService) GetStats(name string) (*domain.Counter, error) {
	counter, err := c.fastRepo.GetCachedTotal(name)

	if counter == nil {
		return domain.NewCounter(name), nil
	}
	return counter, err
}

func (c *counterService) StartSync(cnt context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-cnt.Done():
				log.Println("Stopping background worker. Flushing final data")
				c.syncToSlowDB(cnt)
				return
			case <-ticker.C:
				c.syncToSlowDB(cnt)
			}
		}
	}()

}

func (c *counterService) syncToSlowDB(cnt context.Context) {
	snapShot := c.fastRepo.GetAll()

	if len(snapShot) == 0 {
		return
	}

	lockKey := "sync_counters_lock"
	ttl := 10 * time.Second

	acquire, err := c.lock.Acquire(cnt, lockKey, ttl)

	if err != nil || !acquire {
		return
	}
	defer c.lock.Release(context.Background(), lockKey)

	existingCounters := c.slowRepo.GetAll()
	existingMap := make(map[string]int64)

	for _, counter := range existingCounters {
		existingMap[counter.Name] = counter.Count
	}

	var finalTotals []*domain.Counter
	for _, memcounter := range snapShot {
		currentDBCount := existingMap[memcounter.Name]
		newAbsoluteTotal := currentDBCount + memcounter.Count

		finalTotals = append(finalTotals, &domain.Counter{
			Name:  memcounter.Name,
			Count: newAbsoluteTotal,
		})
	}

	err = c.slowRepo.SaveBatch(finalTotals)
	if err != nil {
		log.Printf("Failed to batch save counters: %v\n", err)
		return
	}
	c.fastRepo.ResetFlushed(snapShot)

	c.fastRepo.SetBatchCachedTotals(finalTotals)
}
