package main

import (
	"log"

	"github.com/example/jobparser/internal/api"
	"github.com/example/jobparser/internal/config"
	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
)

func main() {
	cfg := config.Load()

	db, err := storage.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	se := search.NewSearchEngine(cfg.MeiliHost, cfg.MeiliAPIKey)
	if err := se.EnsureIndex(); err != nil {
		log.Printf("meilisearch index warning: %v", err)
	}

	app := api.NewRouter(db, se)

	log.Printf("HTTP сервер запущен на :%s", cfg.HTTPPort)
	if err := app.Listen(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
