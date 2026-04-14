package repository

import (
	"context"
	"testing"

	dbfiles "github.com/guitarpawat/worthly-tracker/db"
)

func TestDemoDataRepository_SeedDevDataInsertsSeedRows(t *testing.T) {
	t.Parallel()

	database := openTestDB(t)
	repo := NewDemoDataRepository(database, dbfiles.FS)

	hasData, err := repo.HasAnyUserData(context.Background())
	if err != nil {
		t.Fatalf("HasAnyUserData returned error: %v", err)
	}
	if hasData {
		t.Fatal("expected empty test database before seeding")
	}

	if err := repo.SeedDevData(context.Background()); err != nil {
		t.Fatalf("SeedDevData returned error: %v", err)
	}

	hasData, err = repo.HasAnyUserData(context.Background())
	if err != nil {
		t.Fatalf("HasAnyUserData returned error after seed: %v", err)
	}
	if !hasData {
		t.Fatal("expected seeded database to contain user data")
	}
}
