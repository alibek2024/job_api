package api

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
)

type Handlers struct {
	DB     *storage.DB
	Search *search.SearchEngine
}

// GetVacancies — GET /api/vacancies?q=&city=&source=&limit=&offset=
func (h *Handlers) GetVacancies(c *fiber.Ctx) error {
	q := c.Query("q")
	city := c.Query("city")
	source := c.Query("source")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Основной путь — через Meilisearch (быстрый полнотекстовый поиск)
	if h.Search != nil {
		res, err := h.Search.Search(search.SearchParams{
			Query:  q,
			City:   city,
			Source: source,
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err == nil {
			return c.JSON(fiber.Map{
				"total": res.EstimatedTotalHits,
				"items": res.Hits,
			})
		}
	}

	// Фолбэк — прямой поиск по Postgres
	items, err := h.DB.SearchVacancies(q, city, source, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"items": items})
}

// GetVacancy — GET /api/vacancies/:id
func (h *Handlers) GetVacancy(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "неверный id"})
	}
	v, err := h.DB.GetVacancyByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "вакансия не найдена"})
	}
	return c.JSON(v)
}

func (h *Handlers) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
