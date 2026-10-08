package repository

import (
	"context"
	"testing"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/service"
	"github.com/guitarpawat/worthly-tracker/internal/validator"
	"github.com/shopspring/decimal"
)

func TestCustomAllocationLifecycleAndRollback(t *testing.T) {
	t.Parallel()
	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	ctx := context.Background()
	repo := NewCustomAllocationRepository(database)
	input := dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{
		{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(70)},
		{AssetID: 3, Category: "China", Percentage: decimal.NewFromInt(100)},
	}}
	id, err := repo.SaveChart(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	input.ID = id
	charts, err := repo.ListCharts(ctx)
	if err != nil || len(charts) != 1 || len(charts[0].Entries) != 2 {
		t.Fatalf("charts=%+v err=%v", charts, err)
	}

	duplicate := dto.CustomAllocationChart{Name: "country"}
	if _, err := repo.SaveChart(ctx, duplicate); err == nil {
		t.Fatal("duplicate chart accepted")
	}
	input.Name = "Renamed"
	input.Entries = append(input.Entries, dto.CustomAllocationEntry{AssetID: 9999, Category: "Japan", Percentage: decimal.NewFromInt(100)})
	if _, err := repo.SaveChart(ctx, input); err == nil {
		t.Fatal("unavailable asset accepted")
	}
	charts, err = repo.ListCharts(ctx)
	if err != nil || charts[0].Name != "Country" || len(charts[0].Entries) != 2 {
		t.Fatalf("save was not atomic: %+v %v", charts, err)
	}

	input.Entries = []dto.CustomAllocationEntry{}
	if _, err := repo.SaveChart(ctx, input); err != nil {
		t.Fatal(err)
	}
	var archived int
	if err := database.Get(&archived, `SELECT COUNT(*) FROM custom_allocation_entries WHERE deleted_at IS NOT NULL`); err != nil || archived != 2 {
		t.Fatalf("assignments not soft deleted: %d %v", archived, err)
	}
	if err := repo.DeleteChart(ctx, id); err != nil {
		t.Fatal(err)
	}
	charts, err = repo.ListCharts(ctx)
	if err != nil || len(charts) != 0 {
		t.Fatalf("deleted chart returned: %+v %v", charts, err)
	}
	if _, err := repo.SaveChart(ctx, input); err == nil {
		t.Fatal("deleted chart was updated")
	}
	if err := repo.DeleteChart(ctx, id); err == nil {
		t.Fatal("deleted chart was deleted again")
	}
	if _, err := repo.SaveChart(ctx, dto.CustomAllocationChart{Name: "Renamed"}); err != nil {
		t.Fatalf("deleted name not reusable: %v", err)
	}
}

func TestCustomAllocationBreakdownsUseHistoricalValuesAndCurrentSettings(t *testing.T) {
	t.Parallel()
	database := openTestDB(t)
	insertTestSnapshotData(t, database)
	ctx := context.Background()
	repo := NewCustomAllocationRepository(database)
	svc := service.NewCustomAllocationService(repo, NewProgressRepository(database))
	chart, err := validator.NormalizeCustomAllocation(dto.CustomAllocationChart{Name: "Country", Entries: []dto.CustomAllocationEntry{
		{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(50)},
		{AssetID: 3, Category: "China", Percentage: decimal.NewFromInt(100)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	chart.ID, err = svc.SaveChart(ctx, chart)
	if err != nil {
		t.Fatal(err)
	}
	for _, percentage := range []int64{50, 75} {
		chart.Entries[0].Percentage = decimal.NewFromInt(percentage)
		if _, err := svc.SaveChart(ctx, chart); err != nil {
			t.Fatal(err)
		}
		breakdowns, err := svc.Breakdowns(ctx, "2026-03-12", "2026-04-12")
		if err != nil {
			t.Fatal(err)
		}
		for date, value := range map[string]int64{"2026-03-12": 14000, "2026-04-12": 15000} {
			found := false
			for _, row := range breakdowns[date][0].Rows {
				if row.Name == "USA" {
					found = true
					if !row.Value.Equal(decimal.NewFromInt(value * percentage).Shift(-2)) {
						t.Fatalf("wrong historical value: %+v", row)
					}
				}
			}
			if !found {
				t.Fatalf("missing category for %s", date)
			}
		}
	}
	if _, err := database.Exec(`UPDATE assets SET deleted_at = CURRENT_TIMESTAMP WHERE id = 3`); err != nil {
		t.Fatal(err)
	}
	breakdowns, err := svc.Breakdowns(ctx, "2026-03-12", "2026-03-12")
	if err != nil {
		t.Fatal(err)
	}
	if rows := breakdowns["2026-03-12"][0].Rows; rows[0].Name != "China" || !rows[0].Value.Equal(decimal.NewFromInt(4500)) {
		t.Fatalf("historical deleted asset lost: %+v", rows)
	}
}
