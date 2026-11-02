package outbound

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLockAdapter struct {
	client *redis.Client
}

func NewRedisLockAdapter(redisAddr string) *RedisLockAdapter {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &RedisLockAdapter{
		client: client,
	}
}

func (r *RedisLockAdapter) Close(ctx context.Context) error {
	log.Println("Closing redis connection...")
	return r.client.Close()
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
