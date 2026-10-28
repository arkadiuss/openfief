package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrateThrough applies the embedded migrations up to and including
// version, so that a test can store data in an older schema and check what a
// later migration does with it.
func MigrateThrough(ctx context.Context, pool *pgxpool.Pool, version int) error {
	all, err := migrations()
	if err != nil {
		return err
	}
	var upTo []migration
	for _, m := range all {
		if m.version <= version {
			upTo = append(upTo, m)
		}
	}
	if err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return migrateTx(ctx, tx, upTo) }); err != nil {
		return fmt.Errorf("migrating through %d: %w", version, err)
	}
	return nil
}
