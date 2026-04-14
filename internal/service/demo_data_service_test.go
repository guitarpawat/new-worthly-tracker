package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDemoDataService_InsertOnboardingData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		store       demoDataStoreStub
		wantErr     error
		wantSeedRun bool
	}{
		{
			name: "seeds when database is empty",
			store: demoDataStoreStub{
				hasAnyUserData: false,
			},
			wantSeedRun: true,
		},
		{
			name: "rejects when database already has user data",
			store: demoDataStoreStub{
				hasAnyUserData: true,
			},
			wantErr: ErrDemoDataRequiresEmptyDatabase,
		},
		{
			name: "returns storage error while checking emptiness",
			store: demoDataStoreStub{
				hasAnyUserDataErr: errors.New("count failed"),
			},
			wantErr: errors.New("count failed"),
		},
		{
			name: "returns storage error while seeding",
			store: demoDataStoreStub{
				hasAnyUserData: false,
				seedErr:        errors.New("seed failed"),
			},
			wantErr:     errors.New("seed failed"),
			wantSeedRun: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := NewDemoDataService(&testCase.store)
			err := service.InsertOnboardingData(context.Background())
			if testCase.wantErr == nil && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if testCase.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", testCase.wantErr)
				}
				if !errors.Is(err, testCase.wantErr) && err.Error() != testCase.wantErr.Error() && !strings.Contains(err.Error(), testCase.wantErr.Error()) {
					t.Fatalf("expected error containing %q, got %v", testCase.wantErr.Error(), err)
				}
			}
			if testCase.store.seedCalled != testCase.wantSeedRun {
				t.Fatalf("expected seedCalled=%v, got %v", testCase.wantSeedRun, testCase.store.seedCalled)
			}
		})
	}
}

type demoDataStoreStub struct {
	hasAnyUserData    bool
	hasAnyUserDataErr error
	seedErr           error
	seedCalled        bool
}

func (s *demoDataStoreStub) HasAnyUserData(context.Context) (bool, error) {
	if s.hasAnyUserDataErr != nil {
		return false, s.hasAnyUserDataErr
	}

	return s.hasAnyUserData, nil
}

func (s *demoDataStoreStub) SeedDevData(context.Context) error {
	s.seedCalled = true
	return s.seedErr
}
