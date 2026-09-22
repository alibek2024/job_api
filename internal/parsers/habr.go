package parsers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/example/jobparser/internal/models"
	"github.com/gocolly/colly/v2"
)

// HabrParser — скрейпер career.habr.com. У сервиса нет официального публичного API,
// поэтому используется парсинг HTML через colly.
//
// ВАЖНО: CSS-селекторы ниже подобраны по текущей вёрстке career.habr.com на момент
// написания и МОГУТ измениться. Если парсер перестанет находить вакансии — откройте
// https://career.habr.com/vacancies?q=<query> в браузере, посмотрите DevTools и
// поправьте селекторы (helper-функции selector* ниже).
type HabrParser struct {
	BaseURL   string
	Collector *colly.Collector
	MaxPages  int
}

func NewHabrParser() *HabrParser {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (compatible; JobParserBot/1.0; +https://example.com/bot)"),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*habr.com*", Delay: 1 * time.Second, Parallelism: 2})

	return &HabrParser{
		BaseURL:   "https://career.habr.com",
		Collector: c,
		MaxPages:  3,
	}
}

func (p *HabrParser) Name() string { return string(models.SourceHabr) }

func (p *HabrParser) Fetch(ctx context.Context, query string) ([]models.Vacancy, error) {
	var result []models.Vacancy

	c := p.Collector.Clone()

	// Карточка вакансии в списке /vacancies
	c.OnHTML("div.vacancy-card", func(e *colly.HTMLElement) {
		title := strings.TrimSpace(e.ChildText("a.vacancy-card__title-link"))
		relURL := e.ChildAttr("a.vacancy-card__title-link", "href")
		company := strings.TrimSpace(e.ChildText("div.vacancy-card__company-title"))
		city := strings.TrimSpace(e.ChildText("div.vacancy-card__meta"))
		salaryText := strings.TrimSpace(e.ChildText("div.basic-salary"))

		if title == "" || relURL == "" {
			return
		}

		fullURL := relURL
		if strings.HasPrefix(relURL, "/") {
			fullURL = p.BaseURL + relURL
		}

		// извлекаем id вакансии из URL вида /vacancies/1000123456
		externalID := relURL
		parts := strings.Split(strings.Trim(relURL, "/"), "/")
		if len(parts) > 0 {
			externalID = parts[len(parts)-1]
		}

		from, to, currency := parseHabrSalary(salaryText)

		v := models.Vacancy{
			ExternalID:  externalID,
			Source:      string(models.SourceHabr),
			Title:       title,
			Company:     company,
			City:        city,
			URL:         fullURL,
			SalaryFrom:  from,
			SalaryTo:    to,
			Currency:    currency,
			PublishedAt: time.Now(), // список не всегда отдаёт точную дату — уточняется на странице вакансии при необходимости
		}
		result = append(result, v)
	})

	for page := 1; page <= p.MaxPages; page++ {
		listURL := p.BaseURL + "/vacancies?type=all&page=" + strconv.Itoa(page)
		if query != "" {
			listURL += "&q=" + strings.ReplaceAll(query, " ", "+")
		}
		if err := c.Visit(listURL); err != nil {
			// продолжаем со следующей страницей / логируем на уровне вызывающего кода
			continue
		}
	}
	c.Wait()

	return result, nil
}

// parseHabrSalary — грубый парсинг строк вида "от 500 000 ₸ до 800 000 ₸" / "$1500-2500"
func parseHabrSalary(s string) (from, to *int, currency string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil, ""
	}
	switch {
	case strings.Contains(s, "₸"):
		currency = "KZT"
	case strings.Contains(s, "$"):
		currency = "USD"
	case strings.Contains(s, "₽"):
		currency = "RUR"
	case strings.Contains(s, "€"):
		currency = "EUR"
	}

	var nums []int
	cur := ""
	for _, r := range s {
		if r >= '0' && r <= '9' {
			cur += string(r)
		} else if cur != "" {
			if n, err := strconv.Atoi(cur); err == nil {
				nums = append(nums, n)
			}
			cur = ""
		}
	}
	if cur != "" {
		if n, err := strconv.Atoi(cur); err == nil {
			nums = append(nums, n)
		}
	}

	if len(nums) == 1 {
		from = &nums[0]
	} else if len(nums) >= 2 {
		from = &nums[0]
		to = &nums[1]
	}
	return
}
