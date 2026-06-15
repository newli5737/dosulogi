package db

import (
	"context"

	"github.com/dosu-logi/logistics-erp/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
}

func PingRedis(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
