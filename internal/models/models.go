package models

import (
	"time"

	"github.com/lib/pq"
)

// Source — источник вакансии
type Source string

const (
	SourceHH     Source = "hh"
	SourceHabr   Source = "habr"
	SourceQyzmet Source = "qyzmet"
)

// Vacancy — единая модель вакансии для всех источников
type Vacancy struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ExternalID  string         `json:"external_id" gorm:"index:idx_ext_source,unique"`
	Source      string         `json:"source" gorm:"index:idx_ext_source,unique"`
	Title       string         `json:"title" gorm:"index"`
	Company     string         `json:"company"`
	City        string         `json:"city" gorm:"index"`
	SalaryFrom  *int           `json:"salary_from"`
	SalaryTo    *int           `json:"salary_to"`
	Currency    string         `json:"currency"`
	URL         string         `json:"url" gorm:"uniqueIndex"`
	Description string         `json:"description"`
	Skills      pq.StringArray `json:"skills" gorm:"type:text[]"`
	PublishedAt time.Time      `json:"published_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TelegramSubscription — подписка пользователя ТГ на ключевые слова
type TelegramSubscription struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	ChatID     int64          `json:"chat_id" gorm:"index"`
	Query      string         `json:"query"`
	City       string         `json:"city"`
	Sources    pq.StringArray `json:"sources" gorm:"type:text[]"`
	CreatedAt  time.Time      `json:"created_at"`
	LastNotify time.Time      `json:"last_notify"`
}
