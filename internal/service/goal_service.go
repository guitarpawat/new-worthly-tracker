package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/recorderr"
)

type GoalMutator interface {
	CreateGoal(ctx context.Context, input dto.CreateGoalInput) (dto.GoalMutationResult, error)
	UpdateGoal(ctx context.Context, input dto.UpdateGoalInput) (dto.GoalMutationResult, error)
	DeleteGoal(ctx context.Context, input dto.DeleteGoalInput) error
}

type GoalService struct {
	repository GoalMutator
}

func NewGoalService(repository GoalMutator) *GoalService {
	return &GoalService{repository: repository}
}

func (s *GoalService) CreateGoal(ctx context.Context, input dto.CreateGoalInput) (dto.GoalMutationResult, error) {
	result, err := s.repository.CreateGoal(ctx, input)
	if err != nil {
		return dto.GoalMutationResult{}, fmt.Errorf("create goal: %w", err)
	}
	return result, nil
}

func (s *GoalService) UpdateGoal(ctx context.Context, input dto.UpdateGoalInput) (dto.GoalMutationResult, error) {
	result, err := s.repository.UpdateGoal(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, recorderr.ErrGoalNotFound):
			return dto.GoalMutationResult{}, recorderr.ErrGoalNotFound
		default:
			return dto.GoalMutationResult{}, fmt.Errorf("update goal: %w", err)
		}
	}
	return result, nil
}

func (s *GoalService) DeleteGoal(ctx context.Context, input dto.DeleteGoalInput) error {
	if err := s.repository.DeleteGoal(ctx, input); err != nil {
		switch {
		case errors.Is(err, recorderr.ErrGoalNotFound):
			return recorderr.ErrGoalNotFound
		default:
			return fmt.Errorf("delete goal: %w", err)
		}
	}
	return nil
}
