package shorturl

import (
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	qrcode "github.com/skip2/go-qrcode"
)

type Handler struct {
	service Service
}

type request struct {
	Url    string        `json:"url"`
	Alias  string        `json:"alias"`
	Expiry time.Duration `json:"expiry"` // in hours
}

type response struct {
	URL             string        `json:"url"`
	ShortID         string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate-limit"`
	XRateLimitReset int           `json:"rate-limit-reset"`
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateShortUrl(ctx *fiber.Ctx) error {
	body := new(request)

	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	// URL validation
	if !govalidator.IsURL(body.Url) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}

	short, err := h.service.CreateShortUrl(ctx.Context(), body.Url, body.Alias, body.Expiry)
	if err != nil {
		if err.Error() == "short url already exists" {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	remaining, _ := ctx.Locals("rate-limit-remaining").(int)
	reset, _ := ctx.Locals("rate-limit-reset").(int)

	// show default if not set
	if body.Expiry == 0 {
		body.Expiry = 72
	}

	return ctx.Status(fiber.StatusOK).JSON(response{
		URL:             body.Url,
		ShortID:         short,
		Expiry:          body.Expiry,
		XRateRemaining:  remaining,
		XRateLimitReset: reset,
	})
}

func (h *Handler) GetStats(ctx *fiber.Ctx) error {
	shortID := ctx.Params("shortID")

	url, err := h.service.GetStats(ctx.Context(), shortID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short url not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"short_id":     url.ShortID,
		"hits":         url.Hits,
		"original_url": url.OriginalURL,
		"created_at":   url.CreatedAt,
		"expires_at":   url.ExpiresAt,
	})
}

// GetQR returns a PNG QR code that encodes the short URL for the given shortID.
// Optional query param ?size=NNN sets the pixel size (clamped to 64..1024).
func (h *Handler) GetQR(ctx *fiber.Ctx) error {
	shortID := ctx.Params("shortID")

	// Confirm the short URL exists (and isn't expired) before generating a QR.
	url, err := h.service.GetStats(ctx.Context(), shortID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short url not found"})
	}
	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
		return ctx.Status(fiber.StatusGone).JSON(fiber.Map{"error": "short url expired"})
	}

	// Size handling: default 256, clamp to a sane range.
	size := ctx.QueryInt("size", 256)
	if size < 64 {
		size = 64
	}
	if size > 1024 {
		size = 1024
	}

	// The QR encodes the canonical redirect URL: DOMAIN/shortID
	target := buildQRTarget(url.ShortID)

	png, err := qrcode.Encode(target, qrcode.Medium, size)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate qr"})
	}

	ctx.Set("Content-Type", "image/png")
	ctx.Set("Cache-Control", "public, max-age=86400")
	return ctx.Send(png)
}

// buildQRTarget builds the full short URL the QR should point to.
func buildQRTarget(shortID string) string {
	domain := os.Getenv("DOMAIN")
	scheme := "https://"
	// localhost is plain http during development
	if len(domain) >= 9 && domain[:9] == "localhost" {
		scheme = "http://"
	}
	return scheme + domain + "/" + shortID
}

func (h *Handler) Redirect(ctx *fiber.Ctx) error {
	shortID := ctx.Params("shortID")

	url, err := h.service.GetOriginalUrl(ctx.Context(), shortID)
	if err != nil {
		if err.Error() == "short url expired" {
			return ctx.Status(fiber.StatusGone).JSON(fiber.Map{"error": "short url expired"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short url not found"})
	}

	_ = h.service.IncrementHits(ctx.Context(), shortID)

	return ctx.Redirect(url, fiber.StatusMovedPermanently)
}
