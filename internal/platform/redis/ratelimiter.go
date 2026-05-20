package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

var rateLimitScript = redis.NewScript(`
local current = redis.call("GET", KEYS[1])
if current == false then
    redis.call("SET", KEYS[1], ARGV[1] - 1, "EX", ARGV[2])
    local ttl = redis.call("TTL", KEYS[1])
    return {tonumber(ARGV[1]) - 1, ttl}
end
if tonumber(current) <= 0 then
    local ttl = redis.call("TTL", KEYS[1])
    return {-1, ttl}
end
local remaining = redis.call("DECR", KEYS[1])
local ttl = redis.call("TTL", KEYS[1])
return {remaining, ttl}
`)

func (rl *RateLimiter) Check(ctx context.Context, ip string, limit int, window time.Duration) (bool, int, int, error) {
	res, err := rateLimitScript.Run(
		ctx,
		rl.client,
		[]string{rl.prefixKey(ip)},
		limit,
		int(window.Seconds()),
	).Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("ratelimit script: %w", err)
	}

	vals, ok := res.([]interface{})
	if !ok || len(vals) != 2 {
		return false, 0, 0, fmt.Errorf("ratelimit: unexpected response")
	}

	remaining, _ := vals[0].(int64)
	ttl, _ := vals[1].(int64)

	if remaining < 0 {
		return false, 0, int(ttl), nil
	}
	return true, int(remaining), int(ttl), nil
}

// func (rl *RateLimiter) Check(ctx context.Context, ip string, limit int, window time.Duration) (bool, int, int, error) {
// 	val, err := rl.client.Get(ctx, rl.prefixKey(ip)).Result()
// 	if err == redis.Nil {
// 		if err := rl.client.Set(ctx, rl.prefixKey(ip), limit-1, window).Err(); err != nil {
// 			return false, 0, 0, fmt.Errorf("ratelimit set: %w", err)
// 		}
// 		return true, limit - 1, int(window.Seconds()), nil
// 	}
// 	if err != nil {
// 		return false, 0, 0, fmt.Errorf("ratelimit get: %w", err)
// 	}

// 	count, err := strconv.Atoi(val)
// 	if err != nil {
// 		return false, 0, 0, fmt.Errorf("ratelimit parse: %w", err)
// 	}

// 	ttl, _ := rl.client.TTL(ctx, rl.prefixKey(ip)).Result()

// 	if count <= 0 {
// 		return false, 0, int(ttl.Seconds()), nil
// 	}

// 	rl.client.Decr(ctx, rl.prefixKey(ip))
// 	return true, count - 1, int(ttl.Seconds()), nil
// }

func (rl *RateLimiter) prefixKey(ip string) string {
	return "ratelimit:" + ip
}
