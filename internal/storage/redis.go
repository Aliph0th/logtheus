package storage

import (
	"context"
	"fmt"
	"logtheus/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDatabase struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedisClient(cfg *config.AppConfig) (*RedisDatabase, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
	})
	var ctx = context.Background()
	return &RedisDatabase{rdb, ctx}, nil
}

func (r *RedisDatabase) Get(key string) (string, error) {
	return r.rdb.Get(r.ctx, key).Result()
}

func (r *RedisDatabase) Set(key string, value any, ttlSeconds *time.Duration) error {
	var expiration time.Duration = 0
	if ttlSeconds != nil {
		expiration = *ttlSeconds
	}
	return r.rdb.Set(r.ctx, key, value, expiration).Err()
}

func (r *RedisDatabase) Del(key string) error {
	return r.rdb.Del(r.ctx, key).Err()
}
