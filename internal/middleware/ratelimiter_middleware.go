package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/ishansaini194/customurls/internal/platform/redis"
)

func RateLimit(limiter *redis.RateLimiter, limit int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		allowed, remaining, reset, err := limiter.Check(c.Context(), c.IP(), limit, 30*time.Second)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "rate limit error"})
		}
		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate-limit-reset": reset,
			})
		}

		c.Locals("rate-limit-remaining", remaining)
		c.Locals("rate-limit-reset", reset)

		return c.Next()
	}
}
