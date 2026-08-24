package dto

import "github.com/shopspring/decimal"

type ProgressFilter struct {
	StartDate string
	EndDate   string
}

type ProgressSnapshotItem struct {
	SnapshotID    int64           `db:"snapshot_id"`
	AssetID       int64           `db:"asset_id"`
	SnapshotDate  string          `db:"snapshot_date"`
	AssetName     string          `db:"asset_name"`
	AssetTypeName string          `db:"asset_type_name"`
	CurrentPrice  decimal.Decimal `db:"current_price" ts_type:"string"`
	BoughtPrice   decimal.Decimal `db:"bought_price" ts_type:"string"`
	IsCash        bool            `db:"is_cash"`
	IsLiability   bool            `db:"is_liability"`
}

type ProgressPage struct {
	HasData             bool
	Filter              ProgressFilter
	AvailableDates      []string
	TrendPoints         []ProgressPoint
	ProjectionPoints    []ProjectionPoint
	AllocationSnapshots []AllocationSnapshot
	Goals               []GoalRow
	GoalEstimates       []GoalEstimate
	Summary             ProgressSummary
}

type ProgressPoint struct {
	SnapshotDate string
	TotalBought  decimal.Decimal `ts_type:"string"`
	TotalCurrent decimal.Decimal `ts_type:"string"`
	TotalProfit  decimal.Decimal `ts_type:"string"`
	ProfitRate   decimal.Decimal `ts_type:"string"`
	TotalCash    decimal.Decimal `ts_type:"string"`
	TotalNonCash decimal.Decimal `ts_type:"string"`
	CashRatio    decimal.Decimal `ts_type:"string"`
}

type ProjectionPoint struct {
	SnapshotDate string
	TotalCurrent decimal.Decimal `ts_type:"string"`
	TotalCash    decimal.Decimal `ts_type:"string"`
	TotalNonCash decimal.Decimal `ts_type:"string"`
	Liabilities  decimal.Decimal `ts_type:"string"`
}

type ProgressSummary struct {
	CurrentNetWorth decimal.Decimal `ts_type:"string"`
	CurrentProfit   decimal.Decimal `ts_type:"string"`
	ProfitRate      decimal.Decimal `ts_type:"string"`
	CashRatio       decimal.Decimal `ts_type:"string"`
}

type AllocationSnapshot struct {
	SnapshotDate string
	ByAssetType  []AllocationSlice
	ByAsset      []AllocationSlice
	ByCategory   []AllocationSlice
}

type AllocationSlice struct {
	Name  string
	Value decimal.Decimal `ts_type:"string"`
}

type GoalEstimate struct {
	GoalID         int64
	Name           string
	TargetAmount   decimal.Decimal `ts_type:"string"`
	TargetDate     string
	EstimatedDate  string
	Status         string
	RemainingValue decimal.Decimal `ts_type:"string"`
}
