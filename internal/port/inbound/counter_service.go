package inbound

import (
	"context"
	"sync"
	"time"
)

type IncrementCommand struct {
	Name string
}

type GetStatsQuery struct {
	Name string
}

type GetStatsResult struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type CounterService interface {
	Increment(ctx context.Context, cmd IncrementCommand) error
	GetStats(ctx context.Context, query GetStatsQuery) (GetStatsResult, error)
	StartSync(ctx context.Context, wg *sync.WaitGroup, interval time.Duration)
}
