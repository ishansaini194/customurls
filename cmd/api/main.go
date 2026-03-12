package main

import (
	"log"

	"github.com/ishansaini194/customurls/config"
	"github.com/ishansaini194/customurls/internal/app"
)

func main() {
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
