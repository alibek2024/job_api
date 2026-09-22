package storage

import (
	"log"

	"github.com/example/jobparser/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

// DB — обёртка над GORM
type DB struct {
	Conn *gorm.DB
}

// NewPostgres открывает соединение и (опционально) прогоняет автомиграцию.
// В проде рекомендуется использовать migrations/ через golang-migrate,
// AutoMigrate здесь — для удобства локальной разработки.
func NewPostgres(dsn string) (*DB, error) {
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}

	if err := conn.AutoMigrate(&models.Vacancy{}, &models.TelegramSubscription{}); err != nil {
		log.Printf("automigrate warning: %v", err)
	}

	return &DB{Conn: conn}, nil
}

// UpsertVacancy вставляет вакансию либо обновляет по конфликту (source, external_id)
func (d *DB) UpsertVacancy(v *models.Vacancy) (created bool, err error) {
	res := d.Conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}, {Name: "source"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "company", "city", "salary_from", "salary_to", "currency", "description", "skills", "updated_at"}),
	}).Create(v)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// SearchVacancies — простой поиск по БД (фолбэк, если Meilisearch недоступен)
func (d *DB) SearchVacancies(query, city, source string, limit, offset int) ([]models.Vacancy, error) {
	var vacancies []models.Vacancy
	tx := d.Conn.Model(&models.Vacancy{})
	if query != "" {
		tx = tx.Where("title ILIKE ?", "%"+query+"%")
	}
	if city != "" {
		tx = tx.Where("city ILIKE ?", "%"+city+"%")
	}
	if source != "" {
		tx = tx.Where("source = ?", source)
	}
	err := tx.Order("published_at desc").Limit(limit).Offset(offset).Find(&vacancies).Error
	return vacancies, err
}

func (d *DB) GetVacancyByID(id uint) (*models.Vacancy, error) {
	var v models.Vacancy
	if err := d.Conn.First(&v, id).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

func (d *DB) ListSubscriptions() ([]models.TelegramSubscription, error) {
	var subs []models.TelegramSubscription
	err := d.Conn.Find(&subs).Error
	return subs, err
}

func (d *DB) AddSubscription(sub *models.TelegramSubscription) error {
	return d.Conn.Create(sub).Error
}

func (d *DB) RemoveSubscriptions(chatID int64) error {
	return d.Conn.Where("chat_id = ?", chatID).Delete(&models.TelegramSubscription{}).Error
}
