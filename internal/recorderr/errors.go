package recorderr

import "errors"

var (
	ErrSnapshotNotFound          = errors.New("snapshot not found")
	ErrSnapshotDateAlreadyExists = errors.New("snapshot date already exists")
	ErrAssetUnavailable          = errors.New("asset is unavailable")
	ErrSnapshotHasNoRows         = errors.New("snapshot must contain at least one asset")
	ErrAssetTypeNotFound         = errors.New("asset type not found")
	ErrAssetNotFound             = errors.New("asset not found")
	ErrAssetTypeInactive         = errors.New("asset type is inactive")
	ErrAssetTypeNameExists       = errors.New("asset type name already exists")
	ErrAssetNameExists           = errors.New("asset name already exists in asset type")
	ErrGoalNotFound              = errors.New("goal not found")
)
