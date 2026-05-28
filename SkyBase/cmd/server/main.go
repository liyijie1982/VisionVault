package main

import (
	"log"

	"skybase/internal/app"
	"skybase/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	log.Printf("skybase starting on %s", cfg.HTTP.Addr)
	if err := application.Run(); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
