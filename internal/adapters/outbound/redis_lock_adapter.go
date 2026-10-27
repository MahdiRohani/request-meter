package outbound

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLockAdapter struct {
	client *redis.Client
}

func NewRedisLockAdapter(client *redis.Client) *RedisLockAdapter {
	return &RedisLockAdapter{
		client: client,
	}
}

func (r *RedisLockAdapter) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	success, err := r.client.SetNX(ctx, key, "locked", ttl).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (r *RedisLockAdapter) Release(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
