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
}

func NewService(r Repository) Service {
	return &service{repository: r}
}

func (s *service) CreateShortUrl(ctx context.Context, originalUrl, customUrl string) (string, error) {

	originalUrl = helpers.EnforceHTTP(originalUrl)

	if !helpers.RemoveDomainError(originalUrl) {
		return "", errors.New("cannot shorten own domain")
	}

	if customUrl == "" {
		customUrl = generateShort()
	}

	err := s.repository.Create(ctx, originalUrl, customUrl)
	if err != nil {
		return "", err
	}

	return customUrl, nil
}

func (s *service) GetOriginalUrl(ctx context.Context, customUrl string) (string, error) {
	return s.repository.GetUrl(ctx, customUrl)
}

func generateShort() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}

	return string(b)
}
