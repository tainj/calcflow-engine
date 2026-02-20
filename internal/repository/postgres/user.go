package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/tainj/distributed_calculator2/internal/models"
	"github.com/tainj/distributed_calculator2/pkg/db/postgres"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

type AuthUserRepository struct {
	db     *postgres.DB
	logger logger.Logger
}

func NewAuthUserRepository(db *postgres.DB, logger logger.Logger) *AuthUserRepository {
	return &AuthUserRepository{db: db, logger: logger}
}

func (r *AuthUserRepository) Register(ctx context.Context, user *models.User) error {
	// собираем запрос на вставку пользователя
	query := sq.Insert("users").
		Columns("id", "email", "password_hash", "role").
		Values(user.ID, user.Email, user.PasswordHash, string(user.Role)).
		PlaceholderFormat(sq.Dollar)

	// генерируем sql и аргументы
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	// выполняем запрос
	_, err = r.db.Db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	r.logger.Debug(ctx, "the user has been successfully registered", "userId", user.ID)

	return nil
}

func (r *AuthUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	// ищем пользователя по email
	query := sq.Select("id", "email", "password_hash", "role", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"email": email}).
		PlaceholderFormat(sq.Dollar)

	// готовим sql
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user models.User
	// забираем данные
	err = r.db.Db.QueryRowContext(ctx, sql, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	r.logger.Debug(ctx, "successful receiving a user by mail", "userId", user.ID)

	return &user, nil
}

func (r *AuthUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	// получаем пользователя по id
	query := sq.Select("id", "username", "email", "password_hash", "role", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	// формируем запрос
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var user models.User
	// читаем строку
	err = r.db.Db.QueryRowContext(ctx, sql, args...).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	r.logger.Debug(ctx, "successful receiving a user by id", "userId", user.ID)

	return &user, nil
}
