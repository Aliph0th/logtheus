package repository

import (
	"logtheus/shared/pkg/storages"
	"time"

	"gorm.io/gorm"
)

type TokenRepository struct {
	db    *gorm.DB
	redis *storages.RedisDatabase
}

func NewTokenRepository(db *storages.Database, redis *storages.RedisDatabase) *TokenRepository {
	return &TokenRepository{db: db.DB, redis: redis}
}

func (r *TokenRepository) CreateInRedis(key, token string, ttl time.Duration) error {
	return r.redis.Set(key, token, &ttl)
}

func (r *TokenRepository) GetFromRedis(key string) (string, error) {
	return r.redis.Get(key)
}

func (r *TokenRepository) DeleteFromRedis(key string) error {
	return r.redis.Del(key)
}
