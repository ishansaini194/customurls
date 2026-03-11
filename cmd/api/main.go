package main

import (
	"log"
	"os"

	"github.com/ishansaini194/customurls/internal/server"
	"github.com/ishansaini194/customurls/internal/shorturl"
)

func main() {
	httpServer := server.New()

	dbURL := os.Getenv("DATABASE_URL")

	repo, err := shorturl.NewPostgresRepository(dbURL)
	if err != nil {
		log.Fatal(err)
	}

	service := shorturl.NewService(repo)
	handler := shorturl.NewHandler(service)

	httpServer.App.Post("/shorten", handler.CreateShortUrl)
	httpServer.App.Get("/:custom", handler.Redirect)

	log.Fatal(httpServer.Start(":8080"))
}
