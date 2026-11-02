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
	Increment(cnx context.Context, cmd IncrementCommand) error
	GetStats(cnt context.Context, query GetStatsQuery) (GetStatsResult, error)
	StartSync(cnx context.Context, wg *sync.WaitGroup, interval time.Duration)
}
