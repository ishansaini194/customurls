package shorturl

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type URL struct {
	ID          uint       `json:"id"           gorm:"primaryKey;autoIncrement"`
	OriginalURL string     `json:"original_url" gorm:"not null"`
	CustomURL   string     `json:"custom_url"   gorm:"uniqueIndex;not null"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&URL{})
}

var ErrCacheMiss = errors.New("cache miss")
var ErrNotFound = errors.New("short url not found")
