package app

import "fmt"

func (a *App) DeleteAsset(id int64) error {
	if id <= 0 {
		return fmt.Errorf("asset id must be positive")
	}
	return a.assetManagementService.DeleteAsset(a.ctx, id)
}

func (a *App) DeleteAssetType(id int64) error {
	if id <= 0 {
		return fmt.Errorf("type id must be positive")
	}
	return a.assetManagementService.DeleteAssetType(a.ctx, id)
}
