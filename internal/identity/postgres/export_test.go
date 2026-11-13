package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrateThrough applies the embedded migrations up to and including
// version last, so a test can store what an older core stored and then
// migrate the rest.
func MigrateThrough(ctx context.Context, pool *pgxpool.Pool, last int) error {
	all, err := migrations()
	if err != nil {
		return err
	}
	var some []migration
	for _, m := range all {
		if m.version <= last {
			some = append(some, m)
		}
	}
	if err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return migrateTx(ctx, tx, some) }); err != nil {
		return fmt.Errorf("migrating through %d: %w", last, err)
	}
	return nil
}
