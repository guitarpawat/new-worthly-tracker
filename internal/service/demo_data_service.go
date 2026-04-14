package service

import (
	"context"
	"errors"
	"fmt"
)

var ErrDemoDataRequiresEmptyDatabase = errors.New("demo data can only be inserted into an empty database")

type DemoDataStore interface {
	HasAnyUserData(ctx context.Context) (bool, error)
	SeedDevData(ctx context.Context) error
}

type DemoDataService struct {
	store DemoDataStore
}

func NewDemoDataService(store DemoDataStore) *DemoDataService {
	return &DemoDataService{store: store}
}

func (s *DemoDataService) InsertOnboardingData(ctx context.Context) error {
	hasData, err := s.store.HasAnyUserData(ctx)
	if err != nil {
		return fmt.Errorf("check existing data: %w", err)
	}
	if hasData {
		return ErrDemoDataRequiresEmptyDatabase
	}

	if err := s.store.SeedDevData(ctx); err != nil {
		return fmt.Errorf("seed demo data: %w", err)
	}

	return nil
}
