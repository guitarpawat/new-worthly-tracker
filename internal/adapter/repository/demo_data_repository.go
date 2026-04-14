package repository

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jmoiron/sqlx"
)

type DemoDataRepository struct {
	db     *sqlx.DB
	seedFS fs.FS
}

func NewDemoDataRepository(
	db *sqlx.DB,
	seedFS fs.FS,
) *DemoDataRepository {
	return &DemoDataRepository{
		db:     db,
		seedFS: seedFS,
	}
}

func (r *DemoDataRepository) HasAnyUserData(ctx context.Context) (bool, error) {
	queries := []string{
		`SELECT COUNT(1) FROM asset_types`,
		`SELECT COUNT(1) FROM assets`,
		`SELECT COUNT(1) FROM record_snapshots`,
		`SELECT COUNT(1) FROM record_items`,
	}

	for _, query := range queries {
		var count int
		if err := r.db.GetContext(ctx, &count, query); err != nil {
			return false, fmt.Errorf("count existing data: %w", err)
		}
		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

func (r *DemoDataRepository) SeedDevData(ctx context.Context) error {
	sqlBytes, err := fs.ReadFile(r.seedFS, "seeds/dev_seed.sql")
	if err != nil {
		return fmt.Errorf("read dev seed: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin dev seed: %w", err)
	}

	if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute dev seed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit dev seed: %w", err)
	}

	return nil
}
