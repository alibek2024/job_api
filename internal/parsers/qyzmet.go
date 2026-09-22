package parsers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/example/jobparser/internal/models"
	"github.com/gocolly/colly/v2"
)

// QyzmetParser — скрейпер qyzmet.enbek.kz (государственный портал вакансий РК).
// Публичного API нет, используется парсинг HTML.
//
// ВАЖНО: как и для Habr, селекторы ниже — отправная точка. Портал построен на
// динамическом фронтенде (возможен SSR или подгрузка через XHR/JSON) — перед
// продакшен-использованием откройте страницу поиска в браузере, проверьте вкладку
// Network: если список вакансий приходит отдельным XHR-запросом с JSON, лучше
// дергать этот JSON-эндпоинт напрямую (это быстрее и надёжнее, чем парсить HTML).
type QyzmetParser struct {
	BaseURL   string
	Collector *colly.Collector
	MaxPages  int
}

func NewQyzmetParser() *QyzmetParser {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; JobParserBot/1.0; +https://example.com/bot)"),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*enbek.kz*", Delay: 1 * time.Second, Parallelism: 2})

	return &QyzmetParser{
		BaseURL:   "https://qyzmet.enbek.kz",
		Collector: c,
		MaxPages:  3,
	}
}

func (p *QyzmetParser) Name() string { return string(models.SourceQyzmet) }

func (p *QyzmetParser) Fetch(ctx context.Context, query string) ([]models.Vacancy, error) {
	var result []models.Vacancy

	c := p.Collector.Clone()

	// Пример структуры карточки — ПРОВЕРЬТЕ и поправьте под актуальную вёрстку сайта.
	c.OnHTML("div.vacancy-item, div.vacancy-card, tr.vacancy-row", func(e *colly.HTMLElement) {
		title := strings.TrimSpace(firstNonEmpty(
			e.ChildText("a.vacancy-title"),
			e.ChildText(".title"),
			e.ChildText("a"),
		))
		relURL := firstNonEmptyAttr(e, []string{"a.vacancy-title", "a"}, "href")
		company := strings.TrimSpace(firstNonEmpty(
			e.ChildText(".employer-name"),
			e.ChildText(".company"),
		))
		city := strings.TrimSpace(firstNonEmpty(
			e.ChildText(".region"),
			e.ChildText(".city"),
		))
		salaryText := strings.TrimSpace(firstNonEmpty(
			e.ChildText(".salary"),
			e.ChildText(".vacancy-salary"),
		))

		if title == "" || relURL == "" {
			return
		}

		fullURL := relURL
		if strings.HasPrefix(relURL, "/") {
			fullURL = p.BaseURL + relURL
		}

		externalID := relURL
		parts := strings.Split(strings.Trim(relURL, "/"), "/")
		if len(parts) > 0 {
			externalID = parts[len(parts)-1]
		}

		from, to, currency := parseHabrSalary(salaryText) // формат зарплаты похож, переиспользуем парсер

		v := models.Vacancy{
			ExternalID:  externalID,
			Source:      string(models.SourceQyzmet),
			Title:       title,
			Company:     company,
			City:        city,
			URL:         fullURL,
			SalaryFrom:  from,
			SalaryTo:    to,
			Currency:    currency,
			PublishedAt: time.Now(),
		}
		result = append(result, v)
	})

	for page := 1; page <= p.MaxPages; page++ {
		listURL := p.BaseURL + "/vacancy/search?page=" + strconv.Itoa(page)
		if query != "" {
			listURL += "&keyWord=" + strings.ReplaceAll(query, " ", "+")
		}
		if err := c.Visit(listURL); err != nil {
			continue
		}
	}
	c.Wait()

	return result, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func firstNonEmptyAttr(e *colly.HTMLElement, selectors []string, attr string) string {
	for _, sel := range selectors {
		if v := e.ChildAttr(sel, attr); v != "" {
			return v
		}
	}
	return ""
}
