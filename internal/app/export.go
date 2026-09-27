package app

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/guitarpawat/worthly-tracker/internal/dto"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportSnapshotCSV exports the selected record after the user chooses a destination.
func (a *App) ExportSnapshotCSV(offset int, snapshotID int64) error {
	if snapshotID <= 0 {
		return fmt.Errorf("select a snapshot to export")
	}
	page, err := a.GetHomePage(offset)
	if err != nil {
		return err
	}
	if !page.HasSnapshot || page.SnapshotID != snapshotID {
		return fmt.Errorf("snapshot changed; reload the record before exporting")
	}
	return saveSnapshotCSV(page, func(filename string) (string, error) {
		return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title:           "Export Record as CSV",
			DefaultFilename: filename,
			Filters:         []runtime.FileFilter{{DisplayName: "CSV files", Pattern: "*.csv"}},
		})
	})
}

func saveSnapshotCSV(page dto.HomePage, choosePath func(string) (string, error)) error {
	data, err := snapshotCSV(page)
	if err != nil {
		return fmt.Errorf("prepare CSV: %w", err)
	}
	path, err := choosePath("worthly-record-" + page.SnapshotDate.Format("2006-01-02") + ".csv")
	if err != nil {
		return fmt.Errorf("choose CSV destination: %w", err)
	}
	if path == "" {
		return nil
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("save CSV: %w", err)
	}
	return nil
}

func snapshotCSV(page dto.HomePage) ([]byte, error) {
	var buffer bytes.Buffer
	// A UTF-8 BOM lets spreadsheet apps recognize non-ASCII asset names.
	buffer.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&buffer)
	rows := [][]string{{
		"Snapshot Date", "Asset Type", "Name", "Broker", "Cash", "Liability",
		"Bought Price", "Current Price", "Profit", "Profit (%)", "Notes",
	}}
	for _, group := range page.Groups {
		for _, row := range group.Rows {
			profit, percent := "", ""
			if row.ProfitApplicable {
				profit = row.Profit.String()
				percent = row.ProfitPercent.Shift(2).String()
			}
			rows = append(rows, []string{
				page.SnapshotDate.Format("2006-01-02"), csvText(group.AssetTypeName),
				csvText(row.AssetName), csvText(row.Broker), strconv.FormatBool(row.IsCash),
				strconv.FormatBool(row.IsLiability), row.BoughtPrice.String(), row.CurrentPrice.String(),
				profit, percent, csvText(row.Remarks),
			})
		}
	}
	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Prevent user-entered text from being interpreted as a spreadsheet formula.
// Numeric columns are generated separately, preserving valid negative amounts.
func csvText(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(value, "\t") || strings.HasPrefix(value, "\r") || strings.HasPrefix(value, "\n") {
		return "'" + value
	}
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
