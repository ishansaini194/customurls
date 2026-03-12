package shorturl

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/ishansaini194/customurls/internal/helpers"
)

type Service interface {
	CreateShortUrl(ctx context.Context, originalUrl, customUrl string) (string, error)
	GetOriginalUrl(ctx context.Context, customUrl string) (string, error)
}

type service struct {
	repository Repository
	cache      Cache
}

func NewService(r Repository, c Cache) Service {
	return &service{repository: r, cache: c}
}

func (s *service) CreateShortUrl(ctx context.Context, originalUrl, customUrl string) (string, error) {
	originalUrl = helpers.EnforceHTTP(originalUrl)

	if !helpers.RemoveDomainError(originalUrl) {
		return "", errors.New("cannot shorten own domain")
	}

	if customUrl == "" {
		customUrl = generateShort(0, "")
	}

	if err := s.repository.Create(ctx, originalUrl, customUrl); err != nil {
		return "", err
	}

	return customUrl, nil
}

func (s *service) GetOriginalUrl(ctx context.Context, customUrl string) (string, error) {
	// 1. check cache first
	cached, err := s.cache.Get(ctx, customUrl)
	if err == nil {
		return cached.OriginalURL, nil
	}

	// 2. cache miss, hit postgres
	originalUrl, err := s.repository.GetUrl(ctx, customUrl)
	if err != nil {
		return "", err
	}

	// 3. populate cache for next time
	_ = s.cache.Set(ctx, customUrl, &URL{
		OriginalURL: originalUrl,
		CustomURL:   customUrl,
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
