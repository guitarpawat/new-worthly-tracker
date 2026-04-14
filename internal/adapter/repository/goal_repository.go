package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type GoalRepository struct {
	db *sqlx.DB
}

type goalRow struct {
	ID           int64          `db:"id"`
	Name         string         `db:"name"`
	TargetAmount float64        `db:"target_amount"`
	TargetDate   sql.NullString `db:"target_date"`
}

func NewGoalRepository(db *sqlx.DB) *GoalRepository {
	return &GoalRepository{db: db}
}

func (r *GoalRepository) ListGoals(ctx context.Context) ([]dto.GoalRow, error) {
	rows := []goalRow{}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT
			id,
			name,
			target_amount,
			CAST(target_date AS TEXT) AS target_date
		FROM goals
		WHERE deleted_at IS NULL
		ORDER BY target_amount, name
	`); err != nil {
		return nil, fmt.Errorf("select goals: %w", err)
	}

	goals := make([]dto.GoalRow, 0, len(rows))
	for _, row := range rows {
		goals = append(goals, dto.GoalRow{
			ID:           row.ID,
			Name:         row.Name,
			TargetAmount: row.TargetAmount,
			TargetDate:   row.TargetDate.String,
		})
	}

	return goals, nil
}

func (r *GoalRepository) CreateGoal(ctx context.Context, input dto.CreateGoalInput) (dto.GoalMutationResult, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO goals (name, target_amount, target_date)
		VALUES (?, ?, NULLIF(?, ''))
	`, input.Name, input.TargetAmount, input.TargetDate)
	if err != nil {
		return dto.GoalMutationResult{}, fmt.Errorf("insert goal: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return dto.GoalMutationResult{}, fmt.Errorf("goal last insert id: %w", err)
	}

	return dto.GoalMutationResult{ID: id}, nil
}

func (r *GoalRepository) UpdateGoal(ctx context.Context, input dto.UpdateGoalInput) (dto.GoalMutationResult, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE goals
		SET name = ?,
		    target_amount = ?,
		    target_date = NULLIF(?, '')
		WHERE id = ?
		  AND deleted_at IS NULL
	`, input.Name, input.TargetAmount, input.TargetDate, input.ID)
	if err != nil {
		return dto.GoalMutationResult{}, fmt.Errorf("update goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dto.GoalMutationResult{}, fmt.Errorf("goal rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return dto.GoalMutationResult{}, recorderr.ErrGoalNotFound
	}

	return dto.GoalMutationResult{ID: input.ID}, nil
}

func (r *GoalRepository) DeleteGoal(ctx context.Context, input dto.DeleteGoalInput) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE goals
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND deleted_at IS NULL
	`, input.ID)
	if err != nil {
		return fmt.Errorf("soft delete goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("goal delete rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return recorderr.ErrGoalNotFound
	}

	return nil
}
