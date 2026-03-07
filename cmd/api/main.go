package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/ishansaini194/customurls/internal/shorturl"
)

func main() {
	app := fiber.New()

	dbURL := os.Getenv("DATABASE_URL")

	repo, err := shorturl.NewPostgresRepository(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	service := shorturl.NewService(repo)
	handler := shorturl.NewHandler(service)

	app.Post("/shorten", handler.CreateShortUrl)
	app.Get("/:custom", handler.Redirect)

	log.Fatal(app.Listen(":8080"))
}
