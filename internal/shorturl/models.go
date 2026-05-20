package shorturl

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type URL struct {
	ID           uint       `json:"id"           gorm:"primaryKey;autoIncrement"`
	OriginalURL  string     `json:"original_url" gorm:"not null"`
	ShortID      string     `json:"short_id"     gorm:"uniqueIndex;not null"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	Hits         int        `json:"hits" gorm:"default:0"`
	PasswordHash string     `json:"password_hash,omitempty" gorm:"default:null"`
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&URL{})
}

var ErrCacheMiss = errors.New("cache miss")
var ErrNotFound = errors.New("short url not found")
var ErrExpired = errors.New("short url expired")
var ErrIncorrectPassword = errors.New("incorrect password")
