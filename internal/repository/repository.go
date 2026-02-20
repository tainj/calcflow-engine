package repository

import (
	"context"

	"github.com/tainj/distributed_calculator2/internal/models"
)

type VariableRepository interface {
	SetResult(ctx context.Context, variable string, result float64) error
}

type ExampleRepository interface {
	SaveExample(ctx context.Context, example *models.Example) error
	UpdateExample(ctx context.Context, exampleId string, result float64) error
	UpdateExampleWithError(ctx context.Context, exampleID, errorMsg string) error
	GetResult(ctx context.Context, exampleID string) (float64, error)
	GetExamplesByUserID(ctx context.Context, userID string) ([]models.Example, error)
}

type UserRepository interface {
	Register(ctx context.Context, user *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}
