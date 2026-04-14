package validator

import (
	"fmt"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
)

type ProgressValidator struct{}

func (ProgressValidator) ValidateProgressFilter(filter dto.ProgressFilter) error {
	if filter.StartDate == "" && filter.EndDate == "" {
		return nil
	}
	if filter.StartDate == "" || filter.EndDate == "" {
		return fmt.Errorf("progress filter requires both start date and end date")
	}

	startDate, err := time.Parse("2006-01-02", filter.StartDate)
	if err != nil {
		return fmt.Errorf("progress start date must use YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", filter.EndDate)
	if err != nil {
		return fmt.Errorf("progress end date must use YYYY-MM-DD")
	}
	if endDate.Before(startDate) {
		return fmt.Errorf("progress end date must not be earlier than start date")
	}

	return nil
}
