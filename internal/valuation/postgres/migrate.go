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
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// ErrBadMigration reports an embedded migration whose name does not start
// with a version number, or two migrations with one number.
var ErrBadMigration = errors.New("bad migration")

// RiverSchema is the schema River's tables live in, apart from every
// context's own.
const RiverSchema = "river"

// migrationLock is the advisory lock key that serializes Migrate across
// processes: "valuat" in ASCII.
const migrationLock int64 = 0x76616c756174

// setupSQL creates the valuation and River schemas and the valuation
// migration table if missing.
const setupSQL = `
CREATE SCHEMA IF NOT EXISTS valuation;
CREATE SCHEMA IF NOT EXISTS river;
CREATE TABLE IF NOT EXISTS valuation.schema_migrations (
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

// Migrate applies every embedded valuation migration that the database has
// not applied yet, in version order, in one database transaction: on any
// error none of them is applied. An advisory lock makes concurrent calls,
// such as two replicas starting at once, apply each migration once. It is the
// ledger store's runner (internal/ledger/postgres), copied so the contexts
// share no adapter code, with its own version table and lock key. Then it
// applies River's own migrations into RiverSchema with rivermigrate, which
// runs each in a transaction of its own, as River requires.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	all, err := migrations()
	if err != nil {
		return err
	}
	if err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error { return migrateTx(ctx, tx, all) }); err != nil {
		return fmt.Errorf("migrating the valuation schema: %w", err)
	}
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), &rivermigrate.Config{Schema: RiverSchema})
	if err != nil {
		return fmt.Errorf("migrating River: %w", err)
	}
	if _, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return fmt.Errorf("migrating River: %w", err)
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
		if _, err := tx.Exec(ctx, "INSERT INTO valuation.schema_migrations (version, name) VALUES ($1, $2)", m.version, m.name); err != nil {
			return fmt.Errorf("recording %s: %w", m.name, err)
		}
	}
	return nil
}

// appliedVersions returns the versions recorded in valuation.schema_migrations.
func appliedVersions(ctx context.Context, tx pgx.Tx) (map[int]bool, error) {
	rows, err := tx.Query(ctx, "SELECT version FROM valuation.schema_migrations")
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
