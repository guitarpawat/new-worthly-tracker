package app

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/shopspring/decimal"
)

func TestSnapshotCSV(t *testing.T) {
	t.Parallel()
	page := dto.HomePage{
		SnapshotDate: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		Groups: []dto.HomeAssetGroup{{AssetTypeName: "株式", Rows: []dto.HomeAssetRow{
			{AssetName: "Fund, \"A\"", Broker: "=1+1", Remarks: "line one\nline two",
				BoughtPrice: decimal.RequireFromString("100.01"), CurrentPrice: decimal.RequireFromString("-10.02"),
				Profit: decimal.RequireFromString("-110.03"), ProfitPercent: decimal.RequireFromString("-1.10019"),
				ProfitApplicable: true, IsLiability: true},
			{AssetName: "Cash", IsCash: true, CurrentPrice: decimal.RequireFromString("9007199254740993.01")},
		}}},
	}
	data, err := snapshotCSV(page)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "\xef\xbb\xbf") {
		t.Fatal("missing UTF-8 BOM")
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(string(data), "\xef\xbb\xbf"))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || len(rows[0]) != 11 {
		t.Fatalf("unexpected CSV dimensions: %v", rows)
	}
	want := []string{
		"2025-04-01", "株式", "Fund, \"A\"", "'=1+1", "false", "true", "100.01", "-10.02", "-110.03", "-110.019", "line one\nline two",
	}
	for i, value := range want {
		if rows[1][i] != value {
			t.Errorf("column %d: got %q, want %q", i, rows[1][i], value)
		}
	}
	if rows[2][7] != "9007199254740993.01" || rows[2][8] != "" || rows[2][9] != "" {
		t.Fatalf("cash precision or profit changed: %v", rows[2])
	}
}

func TestCSVText(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"=SUM(A1)", "+1", "-1", "@command", "  =1", "\ttext", "\rtext", "\ntext"} {
		t.Run(value, func(t *testing.T) {
			if got := csvText(value); got != "'"+value {
				t.Fatalf("unsafe text: %q", got)
			}
		})
	}
	if csvText("ordinary text") != "ordinary text" || csvText("") != "" {
		t.Fatal("plain text changed")
	}
}

func TestSaveSnapshotCSV(t *testing.T) {
	t.Parallel()
	page := dto.HomePage{SnapshotDate: time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)}
	path := filepath.Join(t.TempDir(), "record.csv")
	err := saveSnapshotCSV(page, func(name string) (string, error) {
		if name != "worthly-record-2025-04-01.csv" {
			t.Fatalf("filename: %s", name)
		}
		return path, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(path); err != nil || !strings.Contains(string(data), "Snapshot Date") {
		t.Fatalf("CSV not saved: %v", err)
	}
	if err := saveSnapshotCSV(page, func(string) (string, error) { return "", nil }); err != nil {
		t.Fatal(err)
	}
	dialogErr := errors.New("dialog failed")
	if err := saveSnapshotCSV(page, func(string) (string, error) { return "", dialogErr }); !errors.Is(err, dialogErr) {
		t.Fatalf("missing dialog error: %v", err)
	}
	if err := saveSnapshotCSV(page, func(string) (string, error) { return t.TempDir(), nil }); err == nil {
		t.Fatal("expected write error")
	}
}
