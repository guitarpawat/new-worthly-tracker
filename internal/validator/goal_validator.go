package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

type GoalValidator struct{}

func (GoalValidator) ValidateCreateGoalInput(input dto.CreateGoalInput) error {
	return validateGoalPayload(input.Name, input.TargetAmount, input.TargetDate)
}

func (GoalValidator) ValidateUpdateGoalInput(input dto.UpdateGoalInput) error {
	if input.ID <= 0 {
		return fmt.Errorf("goal id must be positive")
	}
	return validateGoalPayload(input.Name, input.TargetAmount, input.TargetDate)
}

func (GoalValidator) ValidateDeleteGoalInput(input dto.DeleteGoalInput) error {
	if input.ID <= 0 {
		return fmt.Errorf("goal id must be positive")
	}
	return nil
}

func validateGoalPayload(name string, targetAmount decimal.Decimal, targetDate string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("goal name is required")
	}
	if !targetAmount.IsPositive() {
		return fmt.Errorf("goal target amount must be greater than zero")
	}
	if hasMoreThanTwoDecimalPlaces(targetAmount) {
		return fmt.Errorf("goal target amount allows at most 2 decimal places")
	}
	if targetDate != "" {
		if _, err := time.Parse("2006-01-02", targetDate); err != nil {
			return fmt.Errorf("goal target date must use YYYY-MM-DD")
		}
	}
	return nil
}
