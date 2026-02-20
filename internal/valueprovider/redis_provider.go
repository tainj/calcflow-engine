package valueprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/tainj/distributed_calculator2/pkg/db/cache"
)

type Provider interface {
	Resolve(ctx context.Context, ref string) (float64, error)
}

type RedisValueProvider struct {
	cache *cache.CACHE
}

func NewRedisValueProvider(cache *cache.CACHE) *RedisValueProvider {
	return &RedisValueProvider{cache: cache}
}

func (r *RedisValueProvider) Resolve(ctx context.Context, ref string) (float64, error) {
	// try as number
	if n, err := strconv.ParseFloat(ref, 64); err == nil {
		return n, nil
	}

	// if not a number - assume it's a variable
	key := "result:" + ref
	var result float64
	if err := r.cache.GetByKey(ctx, key, &result); err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("variable %s not ready yet", ref) // ← not fatal, but error
		}
		return 0, fmt.Errorf("failed to resolve variable %s: %w", ref, err)
	}

	return result, nil
}
