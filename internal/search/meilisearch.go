package search

import (
	"github.com/example/jobparser/internal/models"
	"github.com/meilisearch/meilisearch-go"
)

const indexName = "vacancies"

type SearchEngine struct {
	client *meilisearch.Client
}

func NewSearchEngine(host, apiKey string) *SearchEngine {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   host,
		APIKey: apiKey,
	})
	return &SearchEngine{client: client}
}

// EnsureIndex создаёт индекс и настраивает атрибуты фильтрации/поиска
func (s *SearchEngine) EnsureIndex() error {
	_, err := s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        indexName,
		PrimaryKey: "id",
	})
	if err != nil {
		// индекс может уже существовать — это не ошибка
	}

	idx := s.client.Index(indexName)
	_, _ = idx.UpdateFilterableAttributes(&[]string{"source", "city", "salary_from", "salary_to"})
	_, _ = idx.UpdateSortableAttributes(&[]string{"published_at", "salary_from"})
	_, _ = idx.UpdateSearchableAttributes(&[]string{"title", "company", "description", "skills"})
	return nil
}

// IndexVacancy добавляет/обновляет вакансию в поисковом индексе
func (s *SearchEngine) IndexVacancy(v *models.Vacancy) error {
	idx := s.client.Index(indexName)
	_, err := idx.AddDocuments([]models.Vacancy{*v})
	return err
}

// IndexVacancies — батч индексация
func (s *SearchEngine) IndexVacancies(vs []models.Vacancy) error {
	if len(vs) == 0 {
		return nil
	}
	idx := s.client.Index(indexName)
	_, err := idx.AddDocuments(vs)
	return err
}

type SearchParams struct {
	Query  string
	City   string
	Source string
	Limit  int64
	Offset int64
}

func (s *SearchEngine) Search(p SearchParams) (*meilisearch.SearchResponse, error) {
	idx := s.client.Index(indexName)

	filter := ""
	if p.City != "" {
		filter += "city = '" + p.City + "'"
	}
	if p.Source != "" {
		if filter != "" {
			filter += " AND "
		}
		filter += "source = '" + p.Source + "'"
	}

	req := &meilisearch.SearchRequest{
		Limit:  p.Limit,
		Offset: p.Offset,
	}
	if filter != "" {
		req.Filter = filter
	}

	return idx.Search(p.Query, req)
}
