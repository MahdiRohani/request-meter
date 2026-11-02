package outbound

import (
	"context"
	"time"
)

type DistributedLockPort interface {
	Closable
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
}
