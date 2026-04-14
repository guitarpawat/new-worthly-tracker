package dto

type ProgressFilter struct {
	StartDate string
	EndDate   string
}

type ProgressSnapshotItem struct {
	SnapshotID    int64   `db:"snapshot_id"`
	SnapshotDate  string  `db:"snapshot_date"`
	AssetName     string  `db:"asset_name"`
	AssetTypeName string  `db:"asset_type_name"`
	CurrentPrice  float64 `db:"current_price"`
	BoughtPrice   float64 `db:"bought_price"`
	IsCash        bool    `db:"is_cash"`
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
	TotalBought  float64
	TotalCurrent float64
	TotalProfit  float64
	ProfitRate   float64
	TotalCash    float64
	TotalNonCash float64
	CashRatio    float64
}

type ProjectionPoint struct {
	SnapshotDate string
	TotalCurrent float64
}

type ProgressSummary struct {
	CurrentNetWorth float64
	CurrentProfit   float64
	ProfitRate      float64
	CashRatio       float64
}

type AllocationSnapshot struct {
	SnapshotDate string
	ByAssetType  []AllocationSlice
	ByAsset      []AllocationSlice
	ByCategory   []AllocationSlice
}

type AllocationSlice struct {
	Name  string
	Value float64
}

type GoalEstimate struct {
	GoalID         int64
	Name           string
	TargetAmount   float64
	TargetDate     string
	EstimatedDate  string
	Status         string
	RemainingValue float64
}
