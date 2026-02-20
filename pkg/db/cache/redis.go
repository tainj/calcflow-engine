package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

type Config struct {
	Host string `env:"REDIS_HOST" env-default:"localhost"`
	Port string `env:"REDIS_PORT" env-default:"6379"`
}

type CACHE struct {
	Client *redis.Client
}

func New(cfg Config, l logger.Logger) *CACHE {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
	})

	// Simple check without timeout
	if _, err := client.Ping(context.Background()).Result(); err != nil {
		l.Warn(context.Background(), "redis connection warning", "error", err) // non-fatal error
	} else {
		l.Info(context.Background(), "redis connected")

	}

	return &CACHE{Client: client}
}

func (s *CACHE) GetByKey(ctx context.Context, key string, dest interface{}) error {
	data, err := s.Client.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get key from Redis: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return nil
}

func (s *CACHE) SetByKey(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	if err := s.Client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set key in Redis: %w", err)
	}
	return nil
}
