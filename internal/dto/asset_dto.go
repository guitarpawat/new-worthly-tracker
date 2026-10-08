package dto

import "github.com/shopspring/decimal"

type AssetManagementPage struct {
	CustomAllocationCharts []CustomAllocationChart
	AssetTypes             []AssetTypeRow
	Assets                 []AssetRow
	ActiveAssetTypes       []AssetTypeOption
}

type AssetTypeRow struct {
	ID         int64
	Name       string
	IsActive   bool
	Ordering   int
	AssetCount int
}

type AssetRow struct {
	CanDelete     bool
	ID            int64
	Name          string
	AssetTypeID   int64
	AssetTypeName string
	Broker        string
	IsCash        bool
	IsLiability   bool
	IsActive      bool
	Ordering      int
	AutoIncrement decimal.Decimal `ts_type:"string"`
}

type AssetTypeOption struct {
	ID   int64
	Name string
}

type CreateAssetTypeInput struct {
	Name     string
	IsActive bool
}

type UpdateAssetTypeInput struct {
	ID       int64
	Name     string
	IsActive bool
}

type AssetTypeMutationResult struct {
	ID int64
}

type CreateAssetInput struct {
	Name          string
	AssetTypeID   int64
	Broker        string
	IsCash        bool
	IsLiability   bool
	IsActive      bool
	AutoIncrement decimal.Decimal `ts_type:"string"`
}

type UpdateAssetInput struct {
	ChartAllocations []AssetChartAllocation
	ID               int64
	Name             string
	AssetTypeID      int64
	Broker           string
	IsCash           bool
	IsLiability      bool
	IsActive         bool
	AutoIncrement    decimal.Decimal `ts_type:"string"`
}

type AssetMutationResult struct {
	ID int64
}

type ReorderAssetTypesInput struct {
	OrderedIDs []int64
	ActiveOnly bool
}

type ReorderAssetInput struct {
	AssetTypeID int64
	OrderedIDs  []int64
	ActiveOnly  bool
}
