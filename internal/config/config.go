package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения, читаемые из окружения / .env
type Config struct {
	// Postgres
	PostgresDSN string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Meilisearch
	MeiliHost   string
	MeiliAPIKey string

	// Telegram
	TelegramToken string

	// Web
	HTTPPort string

	// Парсер
	ParseInterval time.Duration
	HHAreaID      string // код региона hh.ru/hh.kz, напр. "159" — Казахстан, "40" — Алматы
	SearchQueries []string
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load читает .env (если есть) и переменные окружения
func Load() *Config {
	_ = godotenv.Load()

	interval, err := time.ParseDuration(getEnv("PARSE_INTERVAL", "15m"))
	if err != nil {
		interval = 15 * time.Minute
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	return &Config{
		PostgresDSN:   getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=jobparser port=5432 sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		MeiliHost:     getEnv("MEILI_HOST", "http://localhost:7700"),
		MeiliAPIKey:   getEnv("MEILI_API_KEY", ""),
		TelegramToken: getEnv("TELEGRAM_TOKEN", ""),
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		ParseInterval: interval,
		HHAreaID:      getEnv("HH_AREA_ID", "159"), // Казахстан
		SearchQueries: []string{
			getEnv("SEARCH_QUERY_1", "Golang"),
			getEnv("SEARCH_QUERY_2", "Backend"),
			getEnv("SEARCH_QUERY_3", "Python"),
		},
	}
}
