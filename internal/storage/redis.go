package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedis(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisClient{Client: rdb}
}

// SeenVacancy помечает вакансию как обработанную (дедупликация между запусками парсера).
// Возвращает true, если вакансия НОВАЯ (ранее не встречалась).
func (r *RedisClient) SeenVacancy(ctx context.Context, key string) (isNew bool, err error) {
	ok, err := r.Client.SetNX(ctx, "seen:"+key, 1, 30*24*time.Hour).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// RateLimit — простой лимитер запросов к источникам (N запросов в окно)
func (r *RedisClient) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	count, err := r.Client.Incr(ctx, "rl:"+key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		r.Client.Expire(ctx, "rl:"+key, window)
	}
	return count <= int64(limit), nil
}
