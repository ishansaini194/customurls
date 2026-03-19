package shorturl

import (
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

type request struct {
	Url         string `json:"url"`
	CustomShort string `json:"customShort"`
	Alias       string `json:"alias"`
}

type response struct {
	URL             string `json:"url"`
	CustomShort     string `json:"short"`
	XRateRemaining  int    `json:"rate-limit"`
	XRateLimitReset int    `json:"rate-limit-reset"`
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

	short, err := h.service.CreateShortUrl(ctx.Context(), body.Url, body.CustomShort)
	if err != nil {
		if err.Error() == "short url already exists" {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	remaining, _ := ctx.Locals("rate-limit-remaining").(int)
	reset, _ := ctx.Locals("rate-limit-reset").(int)
	domain := os.Getenv("DOMAIN")

	return ctx.Status(fiber.StatusOK).JSON(response{
		URL:             body.Url,
		CustomShort:     domain + "/" + short,
		XRateRemaining:  remaining,
		XRateLimitReset: reset,
	})
}

func (h *Handler) Redirect(ctx *fiber.Ctx) error {
	custom := ctx.Params("custom")

	url, err := h.service.GetOriginalUrl(ctx.Context(), custom)
	if err != nil || url == "" {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short url not found"})
	}

	return ctx.Redirect(url, fiber.StatusMovedPermanently)
}
