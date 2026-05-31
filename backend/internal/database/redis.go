package database

import (
	"context"
	"fmt"
	"log"
	"time"

	appconfig "github.com/time/card/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg *appconfig.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	log.Printf("redis connected: %s", cfg.Addr())
	return client, nil
}
