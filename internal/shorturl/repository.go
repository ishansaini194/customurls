package shorturl

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, originalUrl, shortID string, expiresAt *time.Time) error
	GetUrl(ctx context.Context, shortID string) (*URL, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db}
}

func (r *postgresRepository) Create(ctx context.Context, originalUrl, shortID string, expiresAt *time.Time) error {
	url := &URL{
		OriginalURL: originalUrl,
		ShortID:     shortID,
		ExpiresAt:   expiresAt,
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
