package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// SQLSTATEs the sleeve kind constraints raise.
const (
	foreignKeyViolation = "23503"
	checkViolation      = "23514"
	integrityViolation  = "23000"
)

func TestMigrateStoresTheDefaultKinds(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	kinds, err := store.Kinds(t.Context())
	checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
}

func TestMigrateGivesStoredSleevesAKind(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	pool := newDatabase(t)
	if err := postgres.MigrateThrough(ctx, pool, 1); err != nil {
		t.Fatalf("MigrateThrough(1): %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO ledger.sleeves (id, name, role) VALUES
		('long', 'Long', 'defensive'), ('play', 'Play', 'offensive')`); err != nil {
		t.Fatalf("storing sleeves before kinds: %v", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	sleeves, err := postgres.New(pool).Sleeves(ctx)
	want := []portfolio.Sleeve{
		{ID: "long", Name: "Long", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive},
		{ID: "play", Name: "Play", Kind: portfolio.KindSatellite, Role: portfolio.RoleOffensive},
	}
	checkResult(t, "Sleeves", want, sleeves, err)
}

func TestMigrateGivesDifferingSleevesARoleOverride(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	pool := newDatabase(t)
	if err := postgres.MigrateThrough(ctx, pool, 2); err != nil {
		t.Fatalf("MigrateThrough(2): %v", err)
	}
	// Before D-71 a defensive sleeve of an offensive kind needed no override.
	if _, err := pool.Exec(ctx, `INSERT INTO ledger.sleeves (id, name, kind, role, offensive_override) VALUES
		('calm', 'Calm', 'satellite', 'defensive', false), ('core', 'Core', 'core', 'defensive', false),
		('play', 'Play', 'retirement', 'offensive', true)`); err != nil {
		t.Fatalf("storing sleeves before the role override: %v", err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	sleeves, err := postgres.New(pool).Sleeves(ctx)
	want := []portfolio.Sleeve{
		{ID: "calm", Name: "Calm", Kind: portfolio.KindSatellite, Role: portfolio.RoleDefensive, RoleOverride: true},
		{ID: "core", Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive},
		{ID: "play", Name: "Play", Kind: portfolio.KindRetirement, Role: portfolio.RoleOffensive, RoleOverride: true},
	}
	checkResult(t, "Sleeves", want, sleeves, err)
}

func TestKindRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	pool := newPool(t)
	store := postgres.New(pool)
	crypto := portfolio.Kind{Key: "crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive}
	renamed := portfolio.Kind{Key: portfolio.KindRetirement, Name: "IKE and IKZE", DefaultRole: portfolio.RoleDefensive}
	for _, k := range []portfolio.Kind{crypto, renamed} {
		if err := store.SaveKind(ctx, k); err != nil {
			t.Fatalf("SaveKind(%s): %v", k.Key, err)
		}
	}
	// The store keeps a stored kind's default role even when asked to
	// change it, and the trigger refuses a direct update.
	if err := store.SaveKind(ctx, portfolio.Kind{Key: portfolio.KindBonds, Name: "Bonds", DefaultRole: portfolio.RoleOffensive}); err != nil {
		t.Fatalf("SaveKind(bonds): %v", err)
	}
	_, err := pool.Exec(ctx, `UPDATE ledger.sleeve_kinds SET default_role = 'offensive' WHERE key = 'core'`)
	if got := sqlState(err); got != integrityViolation {
		t.Errorf("changing a default role: SQLSTATE %q (%v), want %s", got, err, integrityViolation)
	}
	defaults := portfolio.DefaultKinds()
	want := []portfolio.Kind{defaults[0], defaults[1], defaults[2], crypto, defaults[3], renamed, defaults[5]}
	kinds, err := store.Kinds(ctx)
	checkResult(t, "Kinds", want, kinds, err)
}

func TestSleeveKindConstraints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		sleeve portfolio.Sleeve
		want   string
	}{
		{
			name:   "kind not stored",
			sleeve: portfolio.Sleeve{ID: "s", Name: "S", Kind: "crypto", Role: portfolio.RoleOffensive},
			want:   foreignKeyViolation,
		},
		{
			name:   "override on a sleeve whose role is its kind's",
			sleeve: portfolio.Sleeve{ID: "s", Name: "S", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive, RoleOverride: true},
			want:   checkViolation,
		},
		{
			name:   "defensive in an offensive kind without the override",
			sleeve: portfolio.Sleeve{ID: "s", Name: "S", Kind: portfolio.KindSatellite, Role: portfolio.RoleDefensive},
			want:   checkViolation,
		},
		{
			name:   "offensive in a defensive kind without the override",
			sleeve: portfolio.Sleeve{ID: "s", Name: "S", Kind: portfolio.KindBonds, Role: portfolio.RoleOffensive},
			want:   checkViolation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := postgres.New(newPool(t)).SaveSleeve(t.Context(), tt.sleeve)
			if got := sqlState(err); got != tt.want {
				t.Errorf("SaveSleeve: SQLSTATE %q (%v), want %s", got, err, tt.want)
			}
		})
	}
}

func TestEnsureDefaultsOnPostgres(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	l := app.New(postgres.New(newPool(t)))
	const runs = 4
	errs := make(chan error, runs)
	for range runs {
		go func() { errs <- l.EnsureDefaults(ctx) }()
	}
	for range runs {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent EnsureDefaults: %v", err)
		}
	}
	if err := l.SaveSleeve(ctx, portfolio.Sleeve{ID: "core", Name: "Long term", Kind: portfolio.KindRetirement}); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := l.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults again: %v", err)
	}
	want := portfolio.DefaultSleeves()
	want[1] = portfolio.Sleeve{ID: "core", Name: "Long term", Kind: portfolio.KindRetirement, Role: portfolio.RoleDefensive}
	sleeves, err := l.Sleeves(ctx)
	checkResult(t, "Sleeves", want, sleeves, err)
	kinds, err := l.Kinds(ctx)
	checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
}
