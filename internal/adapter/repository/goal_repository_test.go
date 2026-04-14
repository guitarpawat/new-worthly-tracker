package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

func TestGoalRepository_CreateListUpdateDeleteGoal(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	repo := NewGoalRepository(database)

	createResult, err := repo.CreateGoal(context.Background(), dto.CreateGoalInput{
		Name:         "Emergency Fund",
		TargetAmount: 500000,
		TargetDate:   "2027-12-31",
	})
	if err != nil {
		t.Fatalf("CreateGoal returned error: %v", err)
	}

	goals, err := repo.ListGoals(context.Background())
	if err != nil {
		t.Fatalf("ListGoals returned error: %v", err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(goals))
	}
	if goals[0].ID != createResult.ID || goals[0].Name != "Emergency Fund" {
		t.Fatalf("unexpected goal row: %+v", goals[0])
	}

	updated, err := repo.UpdateGoal(context.Background(), dto.UpdateGoalInput{
		ID:           createResult.ID,
		Name:         "Coast FIRE",
		TargetAmount: 750000,
		TargetDate:   "",
	})
	if err != nil {
		t.Fatalf("UpdateGoal returned error: %v", err)
	}
	if updated.ID != createResult.ID {
		t.Fatalf("expected same id after update, got %d", updated.ID)
	}

	goals, err = repo.ListGoals(context.Background())
	if err != nil {
		t.Fatalf("ListGoals returned error after update: %v", err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected 1 goal after update, got %d", len(goals))
	}
	if goals[0].Name != "Coast FIRE" || goals[0].TargetDate != "" {
		t.Fatalf("unexpected updated goal: %+v", goals[0])
	}

	if err := repo.DeleteGoal(context.Background(), dto.DeleteGoalInput{ID: createResult.ID}); err != nil {
		t.Fatalf("DeleteGoal returned error: %v", err)
	}

	goals, err = repo.ListGoals(context.Background())
	if err != nil {
		t.Fatalf("ListGoals returned error after delete: %v", err)
	}
	if len(goals) != 0 {
		t.Fatalf("expected no goals after soft delete, got %d", len(goals))
	}
}

func TestGoalRepository_UpdateGoalReturnsNotFoundForDeletedGoal(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	repo := NewGoalRepository(database)

	_, err := repo.UpdateGoal(context.Background(), dto.UpdateGoalInput{
		ID:           999,
		Name:         "Missing",
		TargetAmount: 1000,
	})
	if !errors.Is(err, recorderr.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}

func TestGoalRepository_DeleteGoalReturnsNotFoundForMissingGoal(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	repo := NewGoalRepository(database)

	err := repo.DeleteGoal(context.Background(), dto.DeleteGoalInput{ID: 999})
	if !errors.Is(err, recorderr.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}
