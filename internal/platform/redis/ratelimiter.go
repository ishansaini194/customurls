package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

func (rl *RateLimiter) Check(ctx context.Context, ip string, limit int, window time.Duration) (bool, int, int, error) {
	val, err := rl.client.Get(ctx, rl.prefixKey(ip)).Result()
	if err == redis.Nil {
		if err := rl.client.Set(ctx, rl.prefixKey(ip), limit-1, window).Err(); err != nil {
			return false, 0, 0, fmt.Errorf("ratelimit set: %w", err)
		}
		return true, limit - 1, int(window.Seconds()), nil
	}
	if err != nil {
		return false, 0, 0, fmt.Errorf("ratelimit get: %w", err)
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return false, 0, 0, fmt.Errorf("ratelimit parse: %w", err)
	}

	ttl, _ := rl.client.TTL(ctx, rl.prefixKey(ip)).Result()

	if count <= 0 {
		return false, 0, int(ttl.Seconds()), nil
	}

	rl.client.Decr(ctx, rl.prefixKey(ip))
	return true, count - 1, int(ttl.Seconds()), nil
}

func (rl *RateLimiter) prefixKey(ip string) string {
	return "ratelimit:" + ip
}
