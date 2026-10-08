package app

import (
	"context"
	"io"
	"log/slog"
	"testing"

	dbfiles "github.com/guitarpawat/worthly-tracker/db"
	adapterdb "github.com/guitarpawat/worthly-tracker/internal/adapter/db"
	"github.com/guitarpawat/worthly-tracker/internal/adapter/repository"
	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/guitarpawat/worthly-tracker/internal/service"
	"github.com/shopspring/decimal"
)

// The embedded SQLite database is isolated and needs no external test service.
func TestCustomAllocationsAppFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := adapterdb.Open(adapterdb.SQLiteConfig{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := adapterdb.ApplyMigrations(ctx, database, dbfiles.FS); err != nil {
		t.Fatal(err)
	}
	progressRepo := repository.NewProgressRepository(database)
	goalRepo := repository.NewGoalRepository(database)
	app := New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		service.NewRecordService(repository.NewRecordSnapshotRepository(database)),
		service.NewAssetManagementService(repository.NewAssetManagementRepository(database)),
		service.NewProgressService(progressRepo, goalRepo),
		service.NewGoalService(goalRepo),
		service.NewDemoDataService(repository.NewDemoDataRepository(database, dbfiles.FS)),
		service.NewCustomAllocationService(repository.NewCustomAllocationRepository(database), progressRepo),
	)
	app.Startup(ctx)
	id, err := app.SaveCustomAllocationChart(dto.CustomAllocationChart{Name: " Country "})
	if err != nil {
		t.Fatal(err)
	}
	page, err := app.GetAssetManagementPage()
	if err != nil || len(page.CustomAllocationCharts) != 1 || page.CustomAllocationCharts[0].Name != "Country" {
		t.Fatalf("chart not returned to editor: %+v, %v", page, err)
	}
	if _, err := app.SaveCustomAllocationChart(dto.CustomAllocationChart{Name: "country"}); err == nil {
		t.Fatal("duplicate name accepted")
	}
	if hasData, err := repository.NewDemoDataRepository(database, dbfiles.FS).HasAnyUserData(ctx); err != nil || !hasData {
		t.Fatalf("chart-only database treated as empty: %v, %v", hasData, err)
	}
	if _, err := database.Exec(`
		INSERT INTO asset_types (id, name) VALUES (1, 'Funds');
		INSERT INTO assets (id, asset_type_id, name) VALUES (1, 1, 'Fund 1');
		INSERT INTO record_snapshots (id, record_date) VALUES (1, '2026-01-01');
		INSERT INTO record_items (snapshot_id, asset_id, bought_price, current_price) VALUES (1, 1, '500', '1000');
	`); err != nil {
		t.Fatal(err)
	}
	chart := dto.CustomAllocationChart{ID: id, Name: "Country", Entries: []dto.CustomAllocationEntry{
		{AssetID: 1, Category: "USA", Percentage: decimal.NewFromInt(70)},
	}}
	if _, err := app.SaveCustomAllocationChart(chart); err != nil {
		t.Fatal(err)
	}
	chart.Entries[0].Percentage = decimal.NewFromInt(101)
	if _, err := app.SaveCustomAllocationChart(chart); err == nil {
		t.Fatal("invalid percentage accepted at boundary")
	}
	home, err := app.GetHomePage(0)
	if err != nil {
		t.Fatal(err)
	}
	progress, err := app.GetProgressPage(dto.ProgressFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(home.CustomAllocations) != 1 || len(progress.AllocationSnapshots) != 1 {
		t.Fatalf("missing custom allocations: home=%+v progress=%+v", home.CustomAllocations, progress.AllocationSnapshots)
	}
	for _, breakdowns := range [][]dto.CustomAllocationBreakdown{home.CustomAllocations, progress.AllocationSnapshots[0].CustomAllocations} {
		if len(breakdowns) != 1 || len(breakdowns[0].Rows) != 2 {
			t.Fatalf("wrong breakdown: %+v", breakdowns)
		}
		if !breakdowns[0].Rows[0].Value.Equal(decimal.NewFromInt(700)) || !breakdowns[0].Rows[1].Value.Equal(decimal.NewFromInt(300)) {
			t.Fatalf("wrong weighted values: %+v", breakdowns)
		}
	}
	if err := app.DeleteCustomAllocationChart(id); err != nil {
		t.Fatal(err)
	}
	home, err = app.GetHomePage(0)
	if err != nil || len(home.CustomAllocations) != 0 {
		t.Fatalf("deleted chart still visible: %+v %v", home.CustomAllocations, err)
	}
}
