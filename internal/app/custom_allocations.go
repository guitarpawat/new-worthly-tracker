package app

import (
	"fmt"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/validator"
)

func (a *App) SaveCustomAllocationChart(input dto.CustomAllocationChart) (int64, error) {
	normalized, err := validator.NormalizeCustomAllocation(input)
	if err != nil {
		return 0, err
	}
	return a.customAllocations.SaveChart(a.ctx, normalized)
}

func (a *App) DeleteCustomAllocationChart(id int64) error {
	if id <= 0 {
		return fmt.Errorf("chart id must be positive")
	}
	return a.customAllocations.DeleteChart(a.ctx, id)
}
