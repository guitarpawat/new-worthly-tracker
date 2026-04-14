package model

type AssetType struct {
	ID       int64
	Name     string
	IsActive bool
	Ordering int
	SoftDelete
}
