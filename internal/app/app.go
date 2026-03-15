package app

import (
	"fmt"

	"github.com/ishansaini194/customurls/config"
	"github.com/ishansaini194/customurls/internal/middleware"
	"github.com/ishansaini194/customurls/internal/platform/redis"
	"github.com/ishansaini194/customurls/internal/server"
	"github.com/ishansaini194/customurls/internal/shorturl"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg *config.Config) (*server.Server, error) {
	// postgres
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	// auto migrate
	if err := shorturl.Migrate(db); err != nil {
		return nil, fmt.Errorf("migration: %w", err)
	}

	// redis
	redisClient := redis.NewClient(cfg.RedisAddr)
	cache := redis.NewCache(redisClient)
	limiter := redis.NewRateLimiter(redisClient)

	// service + handler
	repo := shorturl.NewPostgresRepository(db)
	service := shorturl.NewService(repo, cache)
	handler := shorturl.NewHandler(service)

	// server
	srv := server.New()
	srv.App.Use(middleware.RateLimit(limiter, cfg.APIQuota))
	srv.App.Post("/shorten", handler.CreateShortUrl)
	srv.App.Get("/:custom", handler.Redirect)

	return srv, nil
}
