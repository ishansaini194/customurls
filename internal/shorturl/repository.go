package shorturl

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, originalUrl, customUrl string) error
	GetUrl(ctx context.Context, customUrl string) (string, error)
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
	return &postgresRepository{db}
}

func (r *postgresRepository) Create(ctx context.Context, originalUrl, customUrl string) error {
	url := &URL{
		OriginalURL: originalUrl,
		CustomURL:   customUrl,
	}
	return r.db.WithContext(ctx).Create(url).Error
}

func (r *postgresRepository) GetUrl(ctx context.Context, customUrl string) (string, error) {
	var url URL
	result := r.db.WithContext(ctx).
		Where("custom_url = ?", customUrl).
		First(&url)

	if result.Error == gorm.ErrRecordNotFound {
		return "", ErrNotFound
	}
	if result.Error != nil {
		return "", result.Error
	}

	return url.OriginalURL, nil
}
