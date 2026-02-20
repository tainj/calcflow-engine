package repository

import (
	postgresRepo "github.com/tainj/distributed_calculator2/internal/repository/postgres"
	redisRepo "github.com/tainj/distributed_calculator2/internal/repository/redis"
	"github.com/tainj/distributed_calculator2/pkg/db/cache"
	"github.com/tainj/distributed_calculator2/pkg/db/postgres"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

type RepositoryFactory struct {
	postgresDB *postgres.DB
	redisCache *cache.CACHE
	logger     logger.Logger
}

func NewRepositoryFactory(postgresDB *postgres.DB, redisCache *cache.CACHE, logger logger.Logger) *RepositoryFactory {
	return &RepositoryFactory{
		postgresDB: postgresDB,
		redisCache: redisCache,
		logger:     logger,
	}
}

// CreateUserRepository creates repository for working with users
func (f *RepositoryFactory) CreateUserRepository() UserRepository {
	return postgresRepo.NewAuthUserRepository(f.postgresDB, f.logger.With("layer", "repo"))
}

// CreateExampleRepository creates repository for working with examples
func (f *RepositoryFactory) CreateExampleRepository() ExampleRepository {
	return postgresRepo.NewPostgresResultRepository(f.postgresDB, f.logger.With("layer", "repo"))
}

// CreateVariableRepository создает репозиторий для работы с переменными в Redis
func (f *RepositoryFactory) CreateVariableRepository() VariableRepository {
	return redisRepo.NewRedisResultRepository(f.redisCache, f.logger.With("layer", "repo"))
}
