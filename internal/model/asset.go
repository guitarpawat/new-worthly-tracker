package model

type Asset struct {
	ID            int64
	AssetTypeID   int64
	Name          string
	Broker        string
	IsCash        bool
	IsActive      bool
	AutoIncrement float64
	Ordering      int
	SoftDelete
}
