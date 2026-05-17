package shorturl

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/ishansaini194/customurls/internal/helpers"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateShortUrl(ctx context.Context, originalUrl, alias, password string, expiry time.Duration) (string, error)
	GetOriginalUrl(ctx context.Context, shortID string) (string, error)
	IncrementHits(ctx context.Context, shortID string) error
	GetHits(ctx context.Context, shortID string) (int, error)
	GetStats(ctx context.Context, shortID string) (*URL, error)
	VerifyPassword(ctx context.Context, shortID, password string) (string, error)
	GetURLCached(ctx context.Context, shortID string) (*URL, error)
}

type service struct {
	repository Repository
	cache      Cache
}

func NewService(r Repository, c Cache) Service {
	return &service{repository: r, cache: c}
}

func (s *service) CreateShortUrl(ctx context.Context, originalUrl, alias, password string, expiry time.Duration) (string, error) {
	originalUrl = helpers.EnforceHTTP(originalUrl)

	if !helpers.RemoveDomainError(originalUrl) {
		return "", errors.New("cannot shorten own domain")
	}

	shortID := generateShort(0, "")

	exists, err := s.cache.Exists(ctx, shortID)
	if err == nil && exists {
		return "", errors.New("short url already exists")
	}

	if expiry == 0 {
		expiry = 72
	}
	expiresAt := time.Now().Add(expiry * time.Hour)

	// hash password if provided
	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", errors.New("could not secure password")
		}
		passwordHash = string(hash)
	}

	if err := s.repository.Create(ctx, originalUrl, shortID, passwordHash, &expiresAt); err != nil {
		return "", err
	}

	_ = s.cache.Set(ctx, shortID, &URL{
		OriginalURL:  originalUrl,
		ShortID:      shortID,
		ExpiresAt:    &expiresAt,
		PasswordHash: passwordHash,
	}, expiry*time.Hour)

	return buildShortURL(alias, shortID), nil
}

func (s *service) VerifyPassword(ctx context.Context, shortID, password string) (string, error) {
	url, err := s.repository.GetUrl(ctx, shortID)
	if err != nil {
		return "", err
	}
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return "", errors.New("short url expired")
	}
	if url.PasswordHash == "" {
		// not protected — just return the URL
		return url.OriginalURL, nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(url.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("incorrect password")
	}
	return url.OriginalURL, nil
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

// GetStats returns the full URL record for a short ID.
func (s *service) GetStats(ctx context.Context, shortID string) (*URL, error) {
	return s.repository.GetUrl(ctx, shortID)
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

// GetURLCached returns the URL record, checking cache first then Postgres.
// Used by the redirect handler so it can see PasswordHash without
// always hitting the database.
func (s *service) GetURLCached(ctx context.Context, shortID string) (*URL, error) {
	// 1. cache first
	cached, err := s.cache.Get(ctx, shortID)
	if err == nil {
		if cached.ExpiresAt != nil && time.Now().After(*cached.ExpiresAt) {
			return nil, errors.New("short url expired")
		}
		return cached, nil
	}

	// 2. cache miss → Postgres
	url, err := s.repository.GetUrl(ctx, shortID)
	if err != nil {
		return nil, err
	}
	if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
		return nil, errors.New("short url expired")
	}

	// 3. populate cache
	var ttl time.Duration
	if url.ExpiresAt != nil {
		ttl = time.Until(*url.ExpiresAt)
	} else {
		ttl = 72 * time.Hour
	}
	_ = s.cache.Set(ctx, shortID, url, ttl)

	return url, nil
}
