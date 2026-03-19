package shorturl

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, originalUrl, shortID string) error
	GetUrl(ctx context.Context, shortID string) (string, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db}
}

func (r *postgresRepository) Create(ctx context.Context, originalUrl, shortID string) error {
	url := &URL{
		OriginalURL: originalUrl,
		ShortID:     shortID,
	}
	return r.db.WithContext(ctx).Create(url).Error
}

func (r *postgresRepository) GetUrl(ctx context.Context, shortID string) (string, error) {
	var url URL
	result := r.db.WithContext(ctx).
		Where("short_id = ?", shortID).
		First(&url)

	if result.Error == gorm.ErrRecordNotFound {
		return "", ErrNotFound
	}
	if result.Error != nil {
		return "", result.Error
	}

	return url.OriginalURL, nil
}
