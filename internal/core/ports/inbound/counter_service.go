package inbound

import (
	"context"
	"time"

	"github.com/MahdiRohani/request-meter/internal/core/domain"
)

type CounterService interface {
	Increment(name string) error
	GetStats(name string) (*domain.Counter, error)
	StartSync(cnx context.Context, interval time.Duration)
}
