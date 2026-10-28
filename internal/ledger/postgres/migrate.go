package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// ErrBadMigration reports an embedded migration whose name does not start
// with a version number, or two migrations with one number.
var ErrBadMigration = errors.New("bad migration")

// migrationLock is the advisory lock key that serializes Migrate across
// processes: "ledger" in ASCII.
const migrationLock int64 = 0x6c6564676572

// setupSQL creates the ledger schema and its migration table if missing.
const setupSQL = `
CREATE SCHEMA IF NOT EXISTS ledger;
CREATE TABLE IF NOT EXISTS ledger.schema_migrations (
    version    integer PRIMARY KEY,
    name       text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT now()
);`

// migration is one embedded SQL file, NNNN_name.sql.
type migration struct {
	version int
	name    string
	sql     string
}

// Migrate applies every embedded migration that the database has not
// applied yet, in version order, all in one database transaction: on any
// error nothing is applied. An advisory lock makes concurrent calls, such as
// two replicas starting at once, apply each migration once.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	all, err := migrations()
	if err != nil {
		return err
	}
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return migrateTx(ctx, tx, all) })
	if err != nil {
		return fmt.Errorf("migrating the ledger schema: %w", err)
	}
	return nil
}

// migrateTx applies the migrations among all that tx's database has not
// applied, holding the migration lock until tx ends.
func migrateTx(ctx context.Context, tx pgx.Tx, all []migration) error {
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", migrationLock); err != nil {
		return fmt.Errorf("locking: %w", err)
	}
	if _, err := tx.Exec(ctx, setupSQL); err != nil {
		return fmt.Errorf("creating the migration table: %w", err)
	}
	applied, err := appliedVersions(ctx, tx)
	if err != nil {
		return err
	}
	for _, m := range all {
		if applied[m.version] {
			continue
		}
		if _, err := tx.Exec(ctx, m.sql); err != nil {
			return fmt.Errorf("applying %s: %w", m.name, err)
		}
		if _, err := tx.Exec(ctx, "INSERT INTO ledger.schema_migrations (version, name) VALUES ($1, $2)", m.version, m.name); err != nil {
			return fmt.Errorf("recording %s: %w", m.name, err)
		}
	}
	return nil
}

// appliedVersions returns the versions recorded in ledger.schema_migrations.
func appliedVersions(ctx context.Context, tx pgx.Tx) (map[int]bool, error) {
	rows, err := tx.Query(ctx, "SELECT version FROM ledger.schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("reading applied migrations: %w", err)
	}
	versions, err := pgx.CollectRows(rows, pgx.RowTo[int32])
	if err != nil {
		return nil, fmt.Errorf("reading applied migrations: %w", err)
	}
	applied := make(map[int]bool, len(versions))
	for _, v := range versions {
		applied[int(v)] = true
	}
	return applied, nil
}

// migrations returns the embedded migrations in version order.
func migrations() ([]migration, error) {
	names, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("listing migrations: %w", err)
	}
	out := make([]migration, 0, len(names))
	seen := map[int]bool{}
	for _, name := range names {
		base := path.Base(name)
		prefix, _, _ := strings.Cut(base, "_")
		v, err := strconv.Atoi(prefix)
		if err != nil || v < 1 || seen[v] {
			return nil, fmt.Errorf("%w: %s", ErrBadMigration, base)
		}
		seen[v] = true
		sql, err := migrationFiles.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", base, err)
		}
		out = append(out, migration{version: v, name: base, sql: string(sql)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].version < out[j].version })
	return out, nil
}
