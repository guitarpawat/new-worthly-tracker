package model

type RecordItem struct {
	ID           int64
	SnapshotID   int64
	AssetID      int64
	BoughtPrice  float64
	CurrentPrice float64
	Remarks      string
	SoftDelete
}
