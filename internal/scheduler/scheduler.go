package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/example/jobparser/internal/parsers"
)

// Run запускает менеджер парсеров сразу и затем каждые interval
func Run(ctx context.Context, m *parsers.Manager, queries []string, interval time.Duration) {
	run := func() {
		log.Println("[scheduler] запуск сбора вакансий...")
		start := time.Now()
		m.RunAll(ctx, queries)
		log.Printf("[scheduler] сбор завершён за %s", time.Since(start))
	}

	run() // первый запуск сразу при старте

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
