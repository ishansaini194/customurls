package app

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	srv.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST",
	}))
	rl := middleware.RateLimit(limiter, cfg.APIQuota)

	srv.App.Post("/shorten", rl, handler.CreateShortUrl)
	srv.App.Post("/verify/:shortID", rl, handler.VerifyPassword)
	srv.App.Get("/qr", handler.GenerateQR)
	srv.App.Get("/stats/:shortID", handler.GetStats)
	srv.App.Get("/qr/:shortID", handler.GetQR)

	srv.App.Static("/", "./frontend", fiber.Static{
		Index:  "index.html",
		Browse: false,
	})

	srv.App.Get("/:shortID", func(c *fiber.Ctx) error {
		shortID := c.Params("shortID")

		// Skip static assets like .css, .js, .png, etc.
		if strings.Contains(shortID, ".") {
			return c.Next()
		}

		return handler.Redirect(c)
	})

	srv.App.Get("/:alias/:shortID", handler.Redirect)

	return srv, nil
}
