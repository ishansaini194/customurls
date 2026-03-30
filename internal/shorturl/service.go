package shorturl

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/ishansaini194/customurls/internal/helpers"
)

type Service interface {
	CreateShortUrl(ctx context.Context, originalUrl, alias string, expiry time.Duration) (string, error)
	GetOriginalUrl(ctx context.Context, shortID string) (string, error)
	IncrementHits(ctx context.Context, shortID string) error
	GetHits(ctx context.Context, shortID string) (int, error)
}

type service struct {
	repository Repository
	cache      Cache
}

func NewService(r Repository, c Cache) Service {
	return &service{repository: r, cache: c}
}

func (s *service) CreateShortUrl(ctx context.Context, originalUrl, alias string, expiry time.Duration) (string, error) {
	originalUrl = helpers.EnforceHTTP(originalUrl)

	if !helpers.RemoveDomainError(originalUrl) {
		return "", errors.New("cannot shorten own domain")
	}

	shortID := generateShort(0, "")

	exists, err := s.cache.Exists(ctx, shortID)
	if err == nil && exists {
		return "", errors.New("short url already exists")
	}

	// default 3 days
	if expiry == 0 {
		expiry = 72
	}
	expiresAt := time.Now().Add(expiry * time.Hour)

	if err := s.repository.Create(ctx, originalUrl, shortID, &expiresAt); err != nil {
		return "", err
	}

	// cache with same TTL
	_ = s.cache.Set(ctx, shortID, &URL{
		OriginalURL: originalUrl,
		ShortID:     shortID,
		ExpiresAt:   &expiresAt,
	}, expiry*time.Hour)

	return buildShortURL(alias, shortID), nil
}

func buildShortURL(alias, shortID string) string {
	domain := os.Getenv("DOMAIN")

	if alias != "" {
		return "http://" + domain + "/" + alias + "/" + shortID
	}

	return "http://" + domain + "/" + shortID
}

func (s *service) GetOriginalUrl(ctx context.Context, shortID string) (string, error) {
	// 1. check cache first
	cached, err := s.cache.Get(ctx, shortID)
	if err == nil {
		// check expiry
		if cached.ExpiresAt != nil && time.Now().After(*cached.ExpiresAt) {
			return "", errors.New("short url expired")
		}
		return cached.OriginalURL, nil
	}

	// 2. cache miss, hit postgres
	url, err := s.repository.GetUrl(ctx, shortID)
	if err != nil {
		return "", err
	}

	// 3. check expiry
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return "", errors.New("short url expired")
	}

	// 4. populate cache safely
	var ttl time.Duration
	if url.ExpiresAt != nil {
		ttl = time.Until(*url.ExpiresAt)
	} else {
		ttl = 72 * time.Hour
	}
	_ = s.cache.Set(ctx, shortID, url, ttl)

	return url.OriginalURL, nil
}

func (s *service) IncrementHits(ctx context.Context, shortID string) error {
	return s.repository.IncrementHits(ctx, shortID)
}

func (s *service) GetHits(ctx context.Context, shortID string) (int, error) {
	return s.repository.GetHits(ctx, shortID)
}

const defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateShort(length int, charset string) string {
	if length <= 0 {
		length = 6
	}
	if charset == "" {
		charset = defaultCharset
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[globalRand.Intn(len(charset))]
	}

	return string(b)
}
