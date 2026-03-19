package shorturl

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/ishansaini194/customurls/internal/helpers"
)

type Service interface {
	CreateShortUrl(ctx context.Context, originalUrl, shortID string) (string, error)
	GetOriginalUrl(ctx context.Context, shortID string) (string, error)
}

type service struct {
	repository Repository
	cache      Cache
}

func NewService(r Repository, c Cache) Service {
	return &service{repository: r, cache: c}
}

func (s *service) CreateShortUrl(ctx context.Context, originalUrl, shortID string) (string, error) {
	originalUrl = helpers.EnforceHTTP(originalUrl)

	if !helpers.RemoveDomainError(originalUrl) {
		return "", errors.New("cannot shorten own domain")
	}

	if shortID == "" {
		shortID = generateShort(0, "")
	}

	// check duplicate
	exists, err := s.cache.Exists(ctx, shortID)
	if err == nil && exists {
		return "", errors.New("short url already exists")
	}

	if err := s.repository.Create(ctx, originalUrl, shortID); err != nil {
		return "", err
	}

	return shortID, nil
}

func (s *service) GetOriginalUrl(ctx context.Context, shortID string) (string, error) {
	// 1. check cache first
	cached, err := s.cache.Get(ctx, shortID)
	if err == nil {
		return cached.OriginalURL, nil
	}

	// 2. cache miss, hit postgres
	originalUrl, err := s.repository.GetUrl(ctx, shortID)
	if err != nil {
		return "", err
	}

	// 3. populate cache for next time
	_ = s.cache.Set(ctx, shortID, &URL{
		OriginalURL: originalUrl,
		ShortID:     shortID,
	}, 24*time.Hour)

	return originalUrl, nil
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
