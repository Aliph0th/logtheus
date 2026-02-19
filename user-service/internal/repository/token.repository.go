package repository

import (
	"logtheus/user/internal/storages"
	"time"

	"gorm.io/gorm"
)

type TokenRepository struct {
	db    *gorm.DB
	redis *storages.RedisDatabase
}

func NewTokenRepository(db *gorm.DB, redis *storages.RedisDatabase) *TokenRepository {
	return &TokenRepository{db: db, redis: redis}
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
