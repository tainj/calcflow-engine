package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/tainj/distributed_calculator2/internal/models"
	"github.com/tainj/distributed_calculator2/pkg/db/postgres"
	"github.com/tainj/distributed_calculator2/pkg/logger"
)

func NewPostgresResultRepository(db *postgres.DB, logger logger.Logger) *PostgresResultRepository {
	return &PostgresResultRepository{db: db, logger: logger}
}

type PostgresResultRepository struct {
	db     *postgres.DB
	logger logger.Logger
}

func (r *PostgresResultRepository) SaveExample(ctx context.Context, example *models.Example) error {
	calculated := example.Error != nil

	query := sq.Insert("examples").
		Columns("id", "expression", "response", "user_id", "calculated", "error").
		Values(
			example.ID,
			example.Expression,
			example.Response,
			example.UserID,
			calculated,
			example.Error,
		).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db.Db)

	_, err := query.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("repository.SaveExample: failed to insert example: %w", err)
	}
	r.logger.Debug(ctx, "save example", "exampleId", example.ID)
	return nil
}

func (r *PostgresResultRepository) UpdateExample(ctx context.Context, exampleId string, result float64) error {
	// update status and result by id
	query := sq.Update("examples").
		Set("calculated", true).
		Set("result", result).
		Where(sq.Eq{"id": exampleId}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db.Db)

	// apply changes
	_, err := query.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("repository.UpdateExample: %w", err)
	}

	r.logger.Debug(ctx, "updating example", "exampleId", exampleId)

	return nil
}

func (r *PostgresResultRepository) UpdateExampleWithError(ctx context.Context, exampleId, errorMsg string) error {
	query := sq.Update("examples").
		Set("calculated", true).
		Set("error", errorMsg).
		Where(sq.Eq{"id": exampleId}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db.Db)

	_, err := query.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to update example with error: %w", err)
	}

	r.logger.Debug(ctx, "updating example with error", "exampleId", exampleId, "errorMsg", errorMsg)

	return nil
}

func (r *PostgresResultRepository) GetResult(ctx context.Context, exampleID string) (float64, error) {
	var calculated bool
	var result sql.NullFloat64
	var dbError sql.NullString

	query := sq.Select("calculated", "result", "error").
		From("examples").
		Where(sq.Eq{"id": exampleID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db.Db)

	err := query.QueryRowContext(ctx).Scan(&calculated, &result, &dbError)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("example not found")
		}
		return 0, fmt.Errorf("failed to query: %w", err)
	}

	if !calculated {
		return 0, fmt.Errorf("calculation not completed yet")
	}

	if dbError.Valid {
		return 0, fmt.Errorf("calculation failed: %s", dbError.String)
	}

	if !result.Valid {
		return 0, fmt.Errorf("result is not available")
	}

	r.logger.Debug(ctx, "successful receipt of the result", "exampleId", exampleID, "result", result)

	return result.Float64, nil
}

func (r *PostgresResultRepository) GetExamplesByUserID(ctx context.Context, userID string) ([]models.Example, error) {
	// build query with squirrel
	query := sq.Select("id", "expression", "calculated", "result", "error", "created_at").
		From("examples").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Dollar)

	// execute
	rows, err := query.RunWith(r.db.Db).QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query examples: %w", err)
	}
	defer rows.Close()

	var examples []models.Example
	for rows.Next() {
		var example models.Example
		var result sql.NullFloat64
		var dbError sql.NullString

		err := rows.Scan(
			&example.ID,
			&example.Expression,
			&example.Calculated,
			&result,
			&dbError,
			&example.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// if there's a result - save it
		if result.Valid {
			example.Result = &result.Float64
		}

		// if there's an error - save it
		if dbError.Valid {
			example.Error = &dbError.String
		}

		examples = append(examples, example)
	}

	// check if there was an error in rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	r.logger.Debug(ctx, "successful receipt of examples", "userId", userID)

	return examples, nil
}
