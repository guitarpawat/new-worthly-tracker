package validator

import (
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

func TestGoalValidator_ValidateCreateGoalInputRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	validator := GoalValidator{}
	testCases := []struct {
		name  string
		input dto.CreateGoalInput
	}{
		{
			name:  "missing name",
			input: dto.CreateGoalInput{TargetAmount: 1000},
		},
		{
			name:  "non positive amount",
			input: dto.CreateGoalInput{Name: "Goal", TargetAmount: 0},
		},
		{
			name:  "bad date",
			input: dto.CreateGoalInput{Name: "Goal", TargetAmount: 1000, TargetDate: "12/31/2026"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if err := validator.ValidateCreateGoalInput(testCase.input); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestProgressValidator_ValidateProgressFilterRejectsInvalidRange(t *testing.T) {
	t.Parallel()

	validator := ProgressValidator{}
	testCases := []struct {
		name   string
		filter dto.ProgressFilter
	}{
		{
			name:   "missing end date",
			filter: dto.ProgressFilter{StartDate: "2026-01-01"},
		},
		{
			name:   "bad start format",
			filter: dto.ProgressFilter{StartDate: "01/01/2026", EndDate: "2026-02-01"},
		},
		{
			name:   "end before start",
			filter: dto.ProgressFilter{StartDate: "2026-03-01", EndDate: "2026-02-01"},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if err := validator.ValidateProgressFilter(testCase.filter); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
