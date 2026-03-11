package shorturl

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (*URL, error)
	Set(ctx context.Context, key string, url *URL, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
