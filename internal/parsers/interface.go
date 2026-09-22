package parsers

import (
	"context"

	"github.com/example/jobparser/internal/models"
)

// Parser — единый интерфейс для всех источников (hh, habr, qyzmet, ...).
// Чтобы добавить новый источник, достаточно реализовать этот интерфейс
// и зарегистрировать его в manager.go
type Parser interface {
	// Name — короткое имя источника, совпадает с models.Source
	Name() string

	// Fetch собирает вакансии по ключевому запросу (query может быть пустым — тогда "последние")
	Fetch(ctx context.Context, query string) ([]models.Vacancy, error)
}
