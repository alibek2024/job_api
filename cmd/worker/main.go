package main

import (
	"context"
	"log"

	"github.com/example/jobparser/internal/config"
	"github.com/example/jobparser/internal/parsers"
	"github.com/example/jobparser/internal/scheduler"
	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
	"github.com/example/jobparser/internal/telegram"
)

func main() {
	cfg := config.Load()

	db, err := storage.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	rdb := storage.NewRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)

	se := search.NewSearchEngine(cfg.MeiliHost, cfg.MeiliAPIKey)
	if err := se.EnsureIndex(); err != nil {
		log.Printf("meilisearch index warning: %v", err)
	}

	mgr := parsers.NewManager(db, rdb, se)
	mgr.Register(parsers.NewHHParser(cfg.HHAreaID))
	mgr.Register(parsers.NewHabrParser())
	mgr.Register(parsers.NewQyzmetParser())

	// Если задан токен ТГ — подключаем рассылку новых вакансий подписчикам
	if cfg.TelegramToken != "" {
		bot, err := telegram.New(cfg.TelegramToken, db, se)
		if err != nil {
			log.Printf("telegram notify disabled: %v", err)
		} else {
			mgr.OnNewVacancy = bot.NotifyNewVacancy
		}
	}

	ctx := context.Background()
	scheduler.Run(ctx, mgr, cfg.SearchQueries, cfg.ParseInterval)
}
