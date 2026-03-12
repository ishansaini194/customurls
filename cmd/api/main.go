package main

import (
	"log"

	"github.com/ishansaini194/customurls/config"
	"github.com/ishansaini194/customurls/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	srv, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("server starting on :%s", cfg.Port)
	log.Fatal(srv.Start(":" + cfg.Port))
}
