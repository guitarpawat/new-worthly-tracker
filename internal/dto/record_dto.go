package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

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
	IsLiability       bool
	BoughtPrice       decimal.Decimal `ts_type:"string"`
	CurrentPrice      decimal.Decimal `ts_type:"string"`
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
	Summary       HomeAssetGroupSummary
	Rows          []HomeAssetRow
}

type HomeAssetGroupSummary struct {
	AssetCount   int
	TotalCurrent decimal.Decimal `ts_type:"string"`
}

type HomeAssetRow struct {
	AssetID          int64
	AssetName        string
	Broker           string
	IsCash           bool
	IsLiability      bool
	BoughtPrice      decimal.Decimal `ts_type:"string"`
	CurrentPrice     decimal.Decimal `ts_type:"string"`
	Profit           decimal.Decimal `ts_type:"string"`
	ProfitPercent    decimal.Decimal `ts_type:"string"`
	ProfitApplicable bool
	Remarks          string
}

type HomeSummary struct {
	TotalBought     decimal.Decimal `ts_type:"string"`
	TotalCurrent    decimal.Decimal `ts_type:"string"`
	TotalProfit     decimal.Decimal `ts_type:"string"`
	TotalProfitRate decimal.Decimal `ts_type:"string"`
	TotalCash       decimal.Decimal `ts_type:"string"`
	TotalNonCash    decimal.Decimal `ts_type:"string"`
	CashRatio       decimal.Decimal `ts_type:"string"`
}

type HomeSummaryDelta struct {
	PreviousSnapshotDate time.Time
	BoughtChange         decimal.Decimal `ts_type:"string"`
	CurrentChange        decimal.Decimal `ts_type:"string"`
	ProfitChange         decimal.Decimal `ts_type:"string"`
	ProfitRateChange     decimal.Decimal `ts_type:"string"`
	CashChange           decimal.Decimal `ts_type:"string"`
	NonCashChange        decimal.Decimal `ts_type:"string"`
	CashRatioChange      decimal.Decimal `ts_type:"string"`
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
	BoughtPrice       decimal.Decimal `ts_type:"string"`
	CurrentPrice      decimal.Decimal `ts_type:"string"`
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
	BoughtPrice       decimal.Decimal `ts_type:"string"`
	CurrentPrice      decimal.Decimal `ts_type:"string"`
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
	BoughtPrice  decimal.Decimal `ts_type:"string"`
	CurrentPrice decimal.Decimal `ts_type:"string"`
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
