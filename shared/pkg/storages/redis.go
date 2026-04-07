package storages

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDatabase struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedisClient(host string, port int, password string, database int) (*RedisDatabase, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

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

func (r *RedisDatabase) Close() error {
	return r.rdb.Close()
}
