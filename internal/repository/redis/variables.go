package repository

import (
	"context"
	"fmt"

	"github.com/tainj/distributed_calculator2/pkg/db/cache"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

type RedisResultRepository struct {
	cache  *cache.CACHE
	logger logger.Logger
}

func NewRedisResultRepository(cache *cache.CACHE, logger logger.Logger) *RedisResultRepository {
	return &RedisResultRepository{cache: cache, logger: logger}
}

func (r *RedisResultRepository) SetResult(ctx context.Context, variable string, result float64) error {
	err := r.cache.SetByKey(ctx, fmt.Sprintf("result:%s", variable), result)
	if err != nil {
		return fmt.Errorf("repository.SetResult: %w", err)
	}

	r.logger.Debug(ctx, "set result", "variable", variable, "result", result)
	return nil
}
