package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/portfolio/postgres"
)

// SQLSTATEs the sleeve constraints raise.
const (
	foreignKeyViolation = "23503"
	checkViolation      = "23514"
	integrityViolation  = "23000"
)

// Synthetic account IDs; nothing real is implied.
const (
	acc1 ledger.AccountID = "acc-1"
	acc2 ledger.AccountID = "acc-2"
)

// sqlState returns err's PostgreSQL SQLSTATE, or "" if it has none.
func sqlState(err error) string {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		return pgErr.Code
	}
	return ""
}

// checkResult fails t on err, or unless got is want; empty and nil are
// equal.
func checkResult[T any](t *testing.T, name string, want, got T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("%s (-want +got):\n%s", name, diff)
	}
}

// knownAccounts is the ledger's accounts as the registry reads them: every
// ID is stored.
type knownAccounts struct{}

func (knownAccounts) Account(_ context.Context, id ledger.AccountID) (ledger.Account, error) {
	return ledger.Account{ID: id, Institution: "test-broker", Name: string(id)}, nil
}

func TestMigrateStoresTheDefaultKinds(t *testing.T) {
	t.Parallel()
	store := postgres.New(newPool(t))
	kinds, err := store.Kinds(t.Context())
	checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
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
	_, err := pool.Exec(ctx, `UPDATE portfolio.sleeve_kinds SET default_role = 'offensive' WHERE key = 'core'`)
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

// TestSleevesAndAccountsRoundTrip stores sleeves and which account is in
// which, keyed by the ledger's account ID with no foreign key into the
// ledger schema, and moves and removes an account.
func TestSleevesAndAccountsRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	store := postgres.New(newPool(t))
	core := portfolio.Sleeve{ID: "core", Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
	play := portfolio.Sleeve{ID: "play", Name: "Play", Kind: portfolio.KindRetirement, Role: portfolio.RoleOffensive, RoleOverride: true}
	for _, s := range []portfolio.Sleeve{play, core} {
		if err := store.SaveSleeve(ctx, s); err != nil {
			t.Fatalf("SaveSleeve(%s): %v", s.ID, err)
		}
	}
	sleeves, err := store.Sleeves(ctx)
	checkResult(t, "Sleeves", []portfolio.Sleeve{core, play}, sleeves, err)
	for _, step := range []struct {
		account ledger.AccountID
		sleeve  portfolio.SleeveID
	}{{acc1, "core"}, {acc2, "core"}, {acc2, "play"}} {
		if placeErr := store.PlaceAccount(ctx, step.account, step.sleeve); placeErr != nil {
			t.Fatalf("PlaceAccount(%s, %s): %v", step.account, step.sleeve, placeErr)
		}
	}
	in, err := store.AccountSleeves(ctx)
	checkResult(t, "AccountSleeves", map[ledger.AccountID]portfolio.SleeveID{acc1: "core", acc2: "play"}, in, err)
	if placeErr := store.PlaceAccount(ctx, acc1, "nowhere"); sqlState(placeErr) != foreignKeyViolation {
		t.Errorf("PlaceAccount in a sleeve not stored: SQLSTATE %q (%v), want %s", sqlState(placeErr), placeErr, foreignKeyViolation)
	}
	for range 2 { // removing again changes nothing
		if removeErr := store.RemoveAccount(ctx, acc2); removeErr != nil {
			t.Fatalf("RemoveAccount: %v", removeErr)
		}
	}
	in, err = store.AccountSleeves(ctx)
	checkResult(t, "AccountSleeves after a removal", map[ledger.AccountID]portfolio.SleeveID{acc1: "core"}, in, err)
}

// TestDeleteSleeves checks D-120 in PostgreSQL: a sleeve goes only once no
// account is in it, and a kind once no sleeve has it.
func TestDeleteSleeves(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	store := postgres.New(newPool(t))
	core := portfolio.Sleeve{ID: "core", Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
	if err := store.SaveSleeve(ctx, core); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := store.PlaceAccount(ctx, acc1, core.ID); err != nil {
		t.Fatalf("PlaceAccount: %v", err)
	}
	steps := []struct {
		name    string
		del     func() error
		wantErr error
	}{
		{"a sleeve with accounts", func() error { return store.DeleteSleeve(ctx, core.ID) }, portfolio.ErrSleeveHasAccounts},
		{"a kind a sleeve has", func() error { return store.DeleteKind(ctx, portfolio.KindCore) }, portfolio.ErrKindInUse},
		{"an unused built-in kind", func() error { return store.DeleteKind(ctx, portfolio.KindSatellite) }, nil},
		{"the account leaves the sleeve", func() error { return store.RemoveAccount(ctx, acc1) }, nil},
		{"the sleeve with none", func() error { return store.DeleteSleeve(ctx, core.ID) }, nil},
		{"the kind no sleeve has", func() error { return store.DeleteKind(ctx, portfolio.KindCore) }, nil},
	}
	for _, s := range steps {
		if err := s.del(); !errors.Is(err, s.wantErr) {
			t.Errorf("deleting %s = %v, want %v", s.name, err, s.wantErr)
		}
	}
	sleeves, err := store.Sleeves(ctx)
	checkResult(t, "Sleeves", []portfolio.Sleeve(nil), sleeves, err)
}

func TestEnsureDefaultsOnPostgres(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	r := app.NewRegistry(postgres.New(newPool(t)), knownAccounts{})
	const runs = 4
	errs := make(chan error, runs)
	for range runs {
		go func() { errs <- r.EnsureDefaults(ctx) }()
	}
	for range runs {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent EnsureDefaults: %v", err)
		}
	}
	if err := r.SaveSleeve(ctx, portfolio.Sleeve{ID: "core", Name: "Long term", Kind: portfolio.KindRetirement}); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := r.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults again: %v", err)
	}
	want := portfolio.DefaultSleeves()
	want[1] = portfolio.Sleeve{ID: "core", Name: "Long term", Kind: portfolio.KindRetirement, Role: portfolio.RoleDefensive}
	sleeves, err := r.Sleeves(ctx)
	checkResult(t, "Sleeves", want, sleeves, err)
	kinds, err := r.Kinds(ctx)
	checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
}
