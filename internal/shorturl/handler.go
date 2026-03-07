package shorturl

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service Service
}

type request struct {
	Url         string `json:"url"`
	CustomShort string `json:"customShort"`
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

	short, err := h.service.CreateShortUrl(ctx.Context(), body.Url, body.CustomShort)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"Short-Url": short})
}

func (h *Handler) Redirect(ctx *fiber.Ctx) error {
	custom := ctx.Params("custom")

	url, err := h.service.GetOriginalUrl(ctx.Context(), custom)
	if err != nil || url == "" {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short url not found"})
	}

	return ctx.Redirect(url, fiber.StatusMovedPermanently)
}
