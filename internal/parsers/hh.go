package parsers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/example/jobparser/internal/models"
)

// HHParser использует официальный публичный API hh.ru (https://github.com/hhru/api).
// Работает без ключа. area — код региона (159 = Казахстан целиком, 40 = Алматы, 160 = Астана и т.д.)
// Полный список: GET https://api.hh.ru/areas
type HHParser struct {
	BaseURL    string // https://api.hh.ru
	Area       string
	HTTPClient *http.Client
}

func NewHHParser(area string) *HHParser {
	return &HHParser{
		BaseURL:    "https://api.hh.ru",
		Area:       area,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (p *HHParser) Name() string { return string(models.SourceHH) }

type hhResponse struct {
	Items []hhItem `json:"items"`
	Pages int      `json:"pages"`
	Page  int      `json:"page"`
}

type hhItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Area struct {
		Name string `json:"name"`
	} `json:"area"`
	Employer struct {
		Name string `json:"name"`
	} `json:"employer"`
	Salary *struct {
		From     *int   `json:"from"`
		To       *int   `json:"to"`
		Currency string `json:"currency"`
	} `json:"salary"`
	Snippet struct {
		Requirement    string `json:"requirement"`
		Responsibility string `json:"responsibility"`
	} `json:"snippet"`
	AlternateURL string    `json:"alternate_url"`
	PublishedAt  time.Time `json:"published_at"`
	KeySkills    []struct {
		Name string `json:"name"`
	} `json:"key_skills"`
}

// Fetch собирает вакансии по запросу query, проходя постранично (макс. 5 страниц за вызов,
// чтобы не перегружать API — при необходимости увеличьте).
func (p *HHParser) Fetch(ctx context.Context, query string) ([]models.Vacancy, error) {
	var result []models.Vacancy

	const perPage = 50
	const maxPages = 5

	for page := 0; page < maxPages; page++ {
		q := url.Values{}
		q.Set("text", query)
		q.Set("area", p.Area)
		q.Set("per_page", strconv.Itoa(perPage))
		q.Set("page", strconv.Itoa(page))
		q.Set("order_by", "publication_time")

		reqURL := p.BaseURL + "/vacancies?" + q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return result, err
		}
		req.Header.Set("User-Agent", "jobparser/1.0 (contact: admin@example.com)")

		resp, err := p.HTTPClient.Do(req)
		if err != nil {
			return result, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return result, fmt.Errorf("hh api status %d", resp.StatusCode)
		}

		var hr hhResponse
		if err := json.NewDecoder(resp.Body).Decode(&hr); err != nil {
			resp.Body.Close()
			return result, err
		}
		resp.Body.Close()

		for _, item := range hr.Items {
			v := models.Vacancy{
				ExternalID:  item.ID,
				Source:      string(models.SourceHH),
				Title:       item.Name,
				Company:     item.Employer.Name,
				City:        item.Area.Name,
				URL:         item.AlternateURL,
				Description: item.Snippet.Requirement + " " + item.Snippet.Responsibility,
				PublishedAt: item.PublishedAt,
			}
			if item.Salary != nil {
				v.SalaryFrom = item.Salary.From
				v.SalaryTo = item.Salary.To
				v.Currency = item.Salary.Currency
			}
			for _, sk := range item.KeySkills {
				v.Skills = append(v.Skills, sk.Name)
			}
			result = append(result, v)
		}

		if page+1 >= hr.Pages {
			break
		}
		time.Sleep(300 * time.Millisecond) // вежливая задержка между запросами
	}

	return result, nil
}
