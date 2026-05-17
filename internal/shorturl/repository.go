package shorturl

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, originalUrl, shortID, passwordHash string, expiresAt *time.Time) error
	GetUrl(ctx context.Context, shortID string) (*URL, error)
	IncrementHits(ctx context.Context, shortID string) error
	GetHits(ctx context.Context, shortID string) (int, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db}
}

func (r *postgresRepository) Create(ctx context.Context, originalUrl, shortID, passwordHash string, expiresAt *time.Time) error {
	url := &URL{
		OriginalURL:  originalUrl,
		ShortID:      shortID,
		ExpiresAt:    expiresAt,
		PasswordHash: passwordHash,
	}
	return r.db.WithContext(ctx).Create(url).Error
}

func (r *postgresRepository) GetUrl(ctx context.Context, shortID string) (*URL, error) {
	var url URL
	result := r.db.WithContext(ctx).
		Where("short_id = ?", shortID).
		First(&url)

	if result.Error == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return &url, nil
}

func (r *postgresRepository) IncrementHits(ctx context.Context, shortID string) error {
	return r.db.WithContext(ctx).
		Model(&URL{}).
		Where("short_id = ?", shortID).
		UpdateColumn("hits", gorm.Expr("hits + 1")).
		Error
}

func (r *postgresRepository) GetHits(ctx context.Context, shortID string) (int, error) {
	var url URL
	result := r.db.WithContext(ctx).
		Select("hits").
		Where("short_id = ?", shortID).
		First(&url)

	if result.Error == gorm.ErrRecordNotFound {
		return 0, ErrNotFound
	}
	if result.Error != nil {
		return 0, result.Error
	}

	return url.Hits, nil
}
