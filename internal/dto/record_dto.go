package dto

import "time"

type RecordInput struct {
	Date string
}

type SnapshotItem struct {
	AssetID           int64
	AssetName         string
	AssetTypeID       int64
	AssetTypeName     string
	AssetTypeOrdering int
	AssetOrdering     int
	Broker            string
	IsCash            bool
	BoughtPrice       float64
	CurrentPrice      float64
	Remarks           string
}

type Snapshot struct {
	ID         int64
	RecordDate time.Time
	Items      []SnapshotItem
}

type HomePage struct {
	SnapshotID         int64
	HasSnapshot        bool
	SnapshotDate       time.Time
	PreviousSnapshot   *time.Time
	SnapshotOptions    []SnapshotOption
	Groups             []HomeAssetGroup
	Summary            HomeSummary
	Comparison         *HomeSummaryDelta
	CanNavigateBack    bool
	CanNavigateForward bool
}

type SnapshotOption struct {
	Offset int
	Label  string
}

type HomeAssetGroup struct {
	AssetTypeName string
	Rows          []HomeAssetRow
}

type HomeAssetRow struct {
	AssetID          int64
	AssetName        string
	Broker           string
	BoughtPrice      float64
	CurrentPrice     float64
	Profit           float64
	ProfitPercent    float64
	ProfitApplicable bool
	Remarks          string
}

type HomeSummary struct {
	TotalBought     float64
	TotalCurrent    float64
	TotalProfit     float64
	TotalProfitRate float64
	TotalCash       float64
	TotalNonCash    float64
	CashRatio       float64
}

type HomeSummaryDelta struct {
	PreviousSnapshotDate time.Time
	BoughtChange         float64
	CurrentChange        float64
	ProfitChange         float64
	ProfitRateChange     float64
	CashChange           float64
	NonCashChange        float64
	CashRatioChange      float64
}

type EditableSnapshot struct {
	ID              int64
	RecordDate      time.Time
	Items           []EditableSnapshotItem
	AvailableAssets []EditableAssetOption
}

type EditableSnapshotItem struct {
	AssetID           int64
	AssetName         string
	AssetTypeID       int64
	AssetTypeName     string
	AssetTypeOrdering int
	AssetOrdering     int
	Broker            string
	IsCash            bool
	IsActive          bool
	BoughtPrice       float64
	CurrentPrice      float64
	Remarks           string
}

type EditableAssetOption struct {
	AssetID           int64
	AssetName         string
	AssetTypeID       int64
	AssetTypeName     string
	AssetTypeOrdering int
	AssetOrdering     int
	Broker            string
	IsCash            bool
	IsActive          bool
}

type EditSnapshotPage struct {
	Mode                 string
	SnapshotID           int64
	SnapshotDate         string
	Groups               []EditSnapshotGroup
	AvailableAssetGroups []EditAssetOptionGroup
}

type EditSnapshotGroup struct {
	AssetTypeName string
	Rows          []EditSnapshotRow
}

type EditSnapshotRow struct {
	AssetTypeOrdering int
	AssetOrdering     int
	AssetID           int64
	AssetName         string
	Broker            string
	IsCash            bool
	IsActive          bool
	BoughtPrice       float64
	CurrentPrice      float64
	Remarks           string
}

type EditAssetOptionGroup struct {
	AssetTypeName string
	Options       []EditAssetOption
}

type EditAssetOption struct {
	AssetTypeOrdering int
	AssetOrdering     int
	AssetID           int64
	AssetName         string
	Broker            string
	IsCash            bool
	IsActive          bool
}

type SaveSnapshotInput struct {
	SnapshotID   int64
	SnapshotDate string
	Items        []SaveSnapshotItemInput
}

type SaveSnapshotItemInput struct {
	AssetID      int64
	BoughtPrice  float64
	CurrentPrice float64
	Remarks      string
}

type SaveSnapshotResult struct {
	Offset int
}

type CreateSnapshotInput struct {
	SnapshotDate string
	Items        []SaveSnapshotItemInput
}

type CreateSnapshotResult struct {
	Offset int
}

type DeleteSnapshotInput struct {
	SnapshotID int64
	Offset     int
}

type DeleteSnapshotResult struct {
	Offset       int
	HasSnapshots bool
}
