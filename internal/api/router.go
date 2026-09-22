package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/example/jobparser/internal/search"
	"github.com/example/jobparser/internal/storage"
)

func NewRouter(db *storage.DB, se *search.SearchEngine) *fiber.App {
	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	h := &Handlers{DB: db, Search: se}

	app.Get("/health", h.Health)
	api := app.Group("/api")
	api.Get("/vacancies", h.GetVacancies)
	api.Get("/vacancies/:id", h.GetVacancy)

	return app
}
