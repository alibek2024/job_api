package parsers

import (
	"context"
	"log"
	"sync"

	"github.com/example/jobparser/internal/models"
	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
)

// Manager оркестрирует все зарегистрированные парсеры
type Manager struct {
	Parsers []Parser
	DB      *storage.DB
	Redis   *storage.RedisClient
	Search  *search.SearchEngine

	// OnNewVacancy вызывается для каждой НОВОЙ (ранее не виденной) вакансии,
	// например для отправки уведомлений в Telegram
	OnNewVacancy func(v models.Vacancy)
}

func NewManager(db *storage.DB, redis *storage.RedisClient, se *search.SearchEngine) *Manager {
	return &Manager{
		Parsers: []Parser{},
		DB:      db,
		Redis:   redis,
		Search:  se,
	}
}

func (m *Manager) Register(p Parser) {
	m.Parsers = append(m.Parsers, p)
}

// RunAll запускает все парсеры параллельно по каждому ключевому запросу
func (m *Manager) RunAll(ctx context.Context, queries []string) {
	var wg sync.WaitGroup

	for _, p := range m.Parsers {
		for _, q := range queries {
			wg.Add(1)
			go func(p Parser, q string) {
				defer wg.Done()
				m.runOne(ctx, p, q)
			}(p, q)
		}
	}

	wg.Wait()
}

func (m *Manager) runOne(ctx context.Context, p Parser, query string) {
	vacancies, err := p.Fetch(ctx, query)
	if err != nil {
		log.Printf("[parser:%s] ошибка запроса %q: %v", p.Name(), query, err)
		return
	}
	log.Printf("[parser:%s] запрос %q: получено %d вакансий", p.Name(), query, len(vacancies))

	var fresh []models.Vacancy

	for _, v := range vacancies {
		key := p.Name() + ":" + v.ExternalID

		isNew, err := m.Redis.SeenVacancy(ctx, key)
		if err != nil {
			log.Printf("[redis] ошибка dedupe: %v", err)
		}

		if _, err := m.DB.UpsertVacancy(&v); err != nil {
			log.Printf("[db] ошибка сохранения вакансии %s: %v", v.URL, err)
			continue
		}

		if isNew {
			fresh = append(fresh, v)
			if m.OnNewVacancy != nil {
				m.OnNewVacancy(v)
			}
		}
	}

	if m.Search != nil && len(fresh) > 0 {
		if err := m.Search.IndexVacancies(fresh); err != nil {
			log.Printf("[meilisearch] ошибка индексации: %v", err)
		}
	}
}
