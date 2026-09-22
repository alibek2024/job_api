package main

import (
	"log"

	"github.com/example/jobparser/internal/config"
	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
	"github.com/example/jobparser/internal/telegram"
)

func main() {
	cfg := config.Load()

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN не задан")
	}

	db, err := storage.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	se := search.NewSearchEngine(cfg.MeiliHost, cfg.MeiliAPIKey)

	bot, err := telegram.New(cfg.TelegramToken, db, se)
	if err != nil {
		log.Fatalf("telegram: %v", err)
	}

	bot.Start()
}
