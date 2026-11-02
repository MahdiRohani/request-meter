package services

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/MahdiRohani/request-meter/internal/domain"
	"github.com/MahdiRohani/request-meter/internal/port/inbound"
	"github.com/MahdiRohani/request-meter/internal/port/outbound"
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
	service.fastRepo.SetBatch(existingCounters)
	return service
}

func (c *counterService) Increment(cnt context.Context, cmd inbound.IncrementCommand) error {
	return c.fastRepo.Increment(cmd.Name)
}

func (c *counterService) GetStats(cnt context.Context, query inbound.GetStatsQuery) (inbound.GetStatsResult, error) {
	counter, err := c.fastRepo.Get(query.Name)

	if counter == nil {
		return inbound.GetStatsResult{Name: query.Name, Count: 0}, nil
	}
	return inbound.GetStatsResult{
		Name:  counter.Name,
		Count: counter.Count,
	}, err
}

func (c *counterService) StartSync(cnt context.Context, wg *sync.WaitGroup, interval time.Duration) {
	wg.Add(1)
	ticker := time.NewTicker(interval)

	go func() {
		defer wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-cnt.Done():
				log.Println("Stopping background worker. Flushing final data")
				rescueCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				c.syncToSlowDB(rescueCtx)
				return
			case <-ticker.C:
				c.syncToSlowDB(cnt)
			}
		}
	}()

}

func (c *counterService) syncToSlowDB(cnt context.Context) {
	instance := os.Getenv("INSTANCE_NAME")
	snapShot := c.fastRepo.GetAll()

	if len(snapShot) == 0 {
		existingCounters := c.slowRepo.GetAll()
		c.fastRepo.SetBatch(existingCounters)
		log.Printf("[Instance-%s] Read-Cache synced from DB (No local writes)", instance)
		return
	}

	lockKey := "sync_counters_lock"
	ttl := 10 * time.Second

	acquire, err := c.lock.Acquire(cnt, lockKey, ttl)

	if err != nil || !acquire {
		log.Printf("[Instance-%s] LOCKED OUT! Waiting for next cycle...", instance)
		return
	}
	defer c.lock.Release(context.Background(), lockKey)
	log.Printf("[Instance-%s] GOT LOCK! Flushing %v to DB...", instance, snapShot)

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

	c.fastRepo.SetBatch(finalTotals)
	log.Printf("[Instance-%s] DB Sync & Cache Update Complete! Current DB State: %v", instance, existingCounters)
}
