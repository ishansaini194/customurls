package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ishansaini194/customurls/internal/shorturl"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "shorturl:"

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string) (*shorturl.URL, error) {
	val, err := c.client.Get(ctx, c.prefixKey(key)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, shorturl.ErrCacheMiss // domain-defined error, not redis.Nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var url shorturl.URL
	if err := json.Unmarshal([]byte(val), &url); err != nil {
		return nil, fmt.Errorf("redis unmarshal: %w", err)
	}

	return &url, nil
}

func (c *Cache) Set(ctx context.Context, key string, url *shorturl.URL, ttl time.Duration) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("redis marshal: %w", err)
	}

	if err := c.client.Set(ctx, c.prefixKey(key), data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, c.prefixKey(key)).Err(); err != nil {
		return fmt.Errorf("redis delete: %w", err)
	}
	return nil
}

func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.prefixKey(key)).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists: %w", err)
	}
	return n > 0, nil
}

func (c *Cache) prefixKey(key string) string {
	return keyPrefix + key
}
