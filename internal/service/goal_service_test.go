package service

import (
	"context"
	"errors"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type goalMutatorStub struct {
	updateErr error
	deleteErr error
}

func (s goalMutatorStub) CreateGoal(context.Context, dto.CreateGoalInput) (dto.GoalMutationResult, error) {
	return dto.GoalMutationResult{ID: 1}, nil
}

func (s goalMutatorStub) UpdateGoal(context.Context, dto.UpdateGoalInput) (dto.GoalMutationResult, error) {
	if s.updateErr != nil {
		return dto.GoalMutationResult{}, s.updateErr
	}
	return dto.GoalMutationResult{ID: 1}, nil
}

func (s goalMutatorStub) DeleteGoal(context.Context, dto.DeleteGoalInput) error {
	return s.deleteErr
}

func TestGoalService_UpdateGoalPreservesNotFoundError(t *testing.T) {
	t.Parallel()

	service := NewGoalService(goalMutatorStub{updateErr: recorderr.ErrGoalNotFound})

	_, err := service.UpdateGoal(context.Background(), dto.UpdateGoalInput{ID: 1, Name: "Goal", TargetAmount: 1000})
	if !errors.Is(err, recorderr.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}

func TestGoalService_DeleteGoalPreservesNotFoundError(t *testing.T) {
	t.Parallel()

	service := NewGoalService(goalMutatorStub{deleteErr: recorderr.ErrGoalNotFound})

	err := service.DeleteGoal(context.Background(), dto.DeleteGoalInput{ID: 1})
	if !errors.Is(err, recorderr.ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound, got %v", err)
	}
}
