package shorturl

import (
	"errors"
	"time"
)

type URL struct {
	ID          int
	OriginalURL string
	CustomURL   string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

var ErrCacheMiss = errors.New("cache miss")
var ErrNotFound = errors.New("short url not found")
