package dto

import "github.com/shopspring/decimal"

// CustomAllocationChart is an asset setting shared by all snapshot dates.
type CustomAllocationChart struct {
	ExcludedAssetIDs []int64
	ID               int64
	Name             string
	Entries          []CustomAllocationEntry
}

type CustomAllocationEntry struct {
	AssetID    int64           `db:"asset_id"`
	Category   string          `db:"category"`
	Percentage decimal.Decimal `db:"percentage" ts_type:"string"`
}

type CustomAllocationBreakdown struct {
	ID   int64
	Name string
	Rows []AllocationSlice
}

type AssetChartAllocation struct {
	Excluded bool
	ChartID  int64
	Entries  []CustomAllocationEntry
}
