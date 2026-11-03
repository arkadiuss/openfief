package app_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// Synthetic account IDs; nothing real is implied.
const (
	acc1 ledger.AccountID = "acc-1"
	acc2 ledger.AccountID = "acc-2"
)

var errNoAccount = errors.New("no such account")

// memRegistry is app.RegistryStore in memory, with the Postgres store's
// rules: a kind keeps its default role, a sleeve needs a stored kind, an
// account a stored sleeve, and nothing in use is deleted.
type memRegistry struct {
	kinds   map[portfolio.KindKey]portfolio.Kind
	sleeves map[portfolio.SleeveID]portfolio.Sleeve
	in      map[ledger.AccountID]portfolio.SleeveID
}

func newMemRegistry() *memRegistry {
	return &memRegistry{
		kinds:   map[portfolio.KindKey]portfolio.Kind{},
		sleeves: map[portfolio.SleeveID]portfolio.Sleeve{},
		in:      map[ledger.AccountID]portfolio.SleeveID{},
	}
}

func (m *memRegistry) SaveKind(_ context.Context, k portfolio.Kind) error {
	if old, ok := m.kinds[k.Key]; ok {
		k.DefaultRole = old.DefaultRole
	}
	m.kinds[k.Key] = k
	return nil
}

func (m *memRegistry) Kinds(context.Context) ([]portfolio.Kind, error) {
	out := slices.Collect(maps.Values(m.kinds))
	slices.SortFunc(out, func(a, b portfolio.Kind) int { return strings.Compare(string(a.Key), string(b.Key)) })
	return out, nil
}

func (m *memRegistry) SaveSleeve(_ context.Context, s portfolio.Sleeve) error {
	if _, ok := m.kinds[s.Kind]; !ok {
		return fmt.Errorf("sleeve %s: kind %q: %w", s.ID, s.Kind, app.ErrNotFound)
	}
	m.sleeves[s.ID] = s
	return nil
}

func (m *memRegistry) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	out := slices.Collect(maps.Values(m.sleeves))
	slices.SortFunc(out, func(a, b portfolio.Sleeve) int { return strings.Compare(string(a.ID), string(b.ID)) })
	return out, nil
}

func (m *memRegistry) DeleteSleeve(_ context.Context, id portfolio.SleeveID) error {
	for _, s := range m.in {
		if s == id {
			return fmt.Errorf("sleeve %s: %w", id, portfolio.ErrSleeveHasAccounts)
		}
	}
	delete(m.sleeves, id)
	return nil
}

func (m *memRegistry) DeleteKind(_ context.Context, key portfolio.KindKey) error {
	for _, s := range m.sleeves {
		if s.Kind == key {
			return fmt.Errorf("kind %s: %w", key, portfolio.ErrKindInUse)
		}
	}
	delete(m.kinds, key)
	return nil
}

func (m *memRegistry) PlaceAccount(_ context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error {
	if _, ok := m.sleeves[sleeve]; !ok {
		return fmt.Errorf("sleeve %s: %w", sleeve, app.ErrNotFound)
	}
	m.in[account] = sleeve
	return nil
}

func (m *memRegistry) AccountSleeves(context.Context) (map[ledger.AccountID]portfolio.SleeveID, error) {
	return maps.Clone(m.in), nil
}

func (m *memRegistry) RemoveAccount(_ context.Context, account ledger.AccountID) error {
	delete(m.in, account)
	return nil
}

func (m *memRegistry) InTx(_ context.Context, fn func(tx app.RegistryStore) error) error {
	return fn(m)
}

// ledgerAccounts are the accounts the ledger holds.
type ledgerAccounts []ledger.AccountID

func (l ledgerAccounts) Account(_ context.Context, id ledger.AccountID) (ledger.Account, error) {
	if !slices.Contains(l, id) {
		return ledger.Account{}, fmt.Errorf("%w: %s", errNoAccount, id)
	}
	return ledger.Account{ID: id, Institution: "test-broker", Name: string(id)}, nil
}

// coreSleeve is the sleeve core that a seeded registry holds.
func coreSleeve() portfolio.Sleeve {
	return portfolio.Sleeve{ID: "core", Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
}

// seededRegistry returns a registry over the built-in kinds and the sleeve
// core, with the ledger holding acc-1 and acc-2.
func seededRegistry(t *testing.T) (*app.Registry, *memRegistry) {
	t.Helper()
	store := newMemRegistry()
	r := app.NewRegistry(store, ledgerAccounts{acc1, acc2})
	for _, k := range portfolio.DefaultKinds() {
		if err := r.SaveKind(t.Context(), k); err != nil {
			t.Fatalf("SaveKind(%s): %v", k.Key, err)
		}
	}
	if err := r.SaveSleeve(t.Context(), coreSleeve()); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	return r, store
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

func TestRegistrySaveSleeve(t *testing.T) {
	t.Parallel()
	play := func(kind portfolio.KindKey, role portfolio.Role, override bool) portfolio.Sleeve {
		return portfolio.Sleeve{ID: "play", Name: "Play", Kind: kind, Role: role, RoleOverride: override}
	}
	tests := []struct {
		name    string
		sleeve  portfolio.Sleeve
		wantErr error
		// stored is the sleeve as stored; zero when saving fails.
		stored portfolio.Sleeve
	}{
		{
			name:   "offensive",
			sleeve: play(portfolio.KindSatellite, portfolio.RoleOffensive, false),
			stored: play(portfolio.KindSatellite, portfolio.RoleOffensive, false),
		},
		{
			name:   "role from the kind",
			sleeve: play(portfolio.KindAgentPool, 0, false),
			stored: play(portfolio.KindAgentPool, portfolio.RoleOffensive, false),
		},
		{
			name:   "defensive in an offensive kind with the override",
			sleeve: play(portfolio.KindSatellite, portfolio.RoleDefensive, true),
			stored: play(portfolio.KindSatellite, portfolio.RoleDefensive, true),
		},
		{
			name:    "defensive in an offensive kind without the override",
			sleeve:  play(portfolio.KindSatellite, portfolio.RoleDefensive, false),
			wantErr: portfolio.ErrRoleOverride,
		},
		{
			name:   "offensive in a defensive kind with the override",
			sleeve: play(portfolio.KindRetirement, portfolio.RoleOffensive, true),
			stored: play(portfolio.KindRetirement, portfolio.RoleOffensive, true),
		},
		{
			name:    "offensive in a defensive kind without the override",
			sleeve:  play(portfolio.KindRetirement, portfolio.RoleOffensive, false),
			wantErr: portfolio.ErrRoleOverride,
		},
		{name: "no kind", sleeve: play("", portfolio.RoleOffensive, false), wantErr: portfolio.ErrUnknownKind},
		{name: "kind not stored", sleeve: play("crypto", portfolio.RoleOffensive, false), wantErr: portfolio.ErrUnknownKind},
		{name: "invalid", sleeve: portfolio.Sleeve{ID: "play", Kind: portfolio.KindCore}, wantErr: portfolio.ErrInvalidSleeve},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, _ := seededRegistry(t)
			err := r.SaveSleeve(t.Context(), tt.sleeve)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SaveSleeve error (-want +got):\n%s", diff)
			}
			want := []portfolio.Sleeve{coreSleeve()}
			if tt.wantErr == nil {
				want = append(want, tt.stored)
			}
			got, err := r.Sleeves(t.Context())
			checkResult(t, "Sleeves", want, got, err)
		})
	}
}

func TestRegistrySaveKind(t *testing.T) {
	t.Parallel()
	defaults := portfolio.DefaultKinds()
	crypto := portfolio.Kind{Key: "crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive}
	renamed := portfolio.Kind{Key: portfolio.KindRetirement, Name: "IKE and IKZE", DefaultRole: portfolio.RoleDefensive}
	tests := []struct {
		name    string
		kind    portfolio.Kind
		wantErr error
		want    []portfolio.Kind
	}{
		{
			name: "custom kind",
			kind: crypto,
			want: []portfolio.Kind{defaults[0], defaults[1], defaults[2], crypto, defaults[3], defaults[4], defaults[5]},
		},
		{
			name: "rename a built-in kind",
			kind: renamed,
			want: []portfolio.Kind{defaults[0], defaults[1], defaults[2], defaults[3], renamed, defaults[5]},
		},
		{
			name:    "change a kind's default role",
			kind:    portfolio.Kind{Key: portfolio.KindCore, Name: "Core", DefaultRole: portfolio.RoleOffensive},
			wantErr: portfolio.ErrKindRoleChange,
			want:    defaults,
		},
		{
			name:    "invalid kind",
			kind:    portfolio.Kind{Key: "Crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive},
			wantErr: portfolio.ErrInvalidKind,
			want:    defaults,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, _ := seededRegistry(t)
			err := r.SaveKind(t.Context(), tt.kind)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SaveKind error (-want +got):\n%s", diff)
			}
			got, err := r.Kinds(t.Context())
			checkResult(t, "Kinds", tt.want, got, err)
		})
	}
}

func TestRegistryEnsureDefaults(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	r := app.NewRegistry(newMemRegistry(), ledgerAccounts{})
	for range 2 {
		if err := r.EnsureDefaults(ctx); err != nil {
			t.Fatalf("EnsureDefaults: %v", err)
		}
		kinds, err := r.Kinds(ctx)
		checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
		sleeves, err := r.Sleeves(ctx)
		checkResult(t, "Sleeves", portfolio.DefaultSleeves(), sleeves, err)
	}

	// The owner renames a kind and moves a default sleeve to another kind;
	// ensuring the defaults again keeps both changes.
	renamed := portfolio.Kind{Key: portfolio.KindCore, Name: "Long term", DefaultRole: portfolio.RoleDefensive}
	if err := r.SaveKind(ctx, renamed); err != nil {
		t.Fatalf("SaveKind: %v", err)
	}
	moved := portfolio.Sleeve{ID: "satellite", Name: "Satellite", Kind: portfolio.KindAgentPool, Role: portfolio.RoleOffensive}
	if err := r.SaveSleeve(ctx, moved); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := r.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults after changes: %v", err)
	}
	wantKinds := portfolio.DefaultKinds()
	wantKinds[2] = renamed
	kinds, err := r.Kinds(ctx)
	checkResult(t, "Kinds after changes", wantKinds, kinds, err)
	wantSleeves := portfolio.DefaultSleeves()
	wantSleeves[2] = moved
	sleeves, err := r.Sleeves(ctx)
	checkResult(t, "Sleeves after changes", wantSleeves, sleeves, err)
}

func TestRegistryEnsureDefaultsOnlyOnAFreshInstall(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	// The demo dataset, or the owner, stored sleeves of their own before
	// the defaults were ensured: the install has started, so only the
	// built-in kinds are added (D-71: an install starts with the defaults).
	own := portfolio.Sleeve{ID: "long-term", Name: "Long term", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
	fresh := app.NewRegistry(newMemRegistry(), ledgerAccounts{})
	if err := fresh.SaveKind(ctx, portfolio.DefaultKinds()[2]); err != nil {
		t.Fatalf("SaveKind: %v", err)
	}
	if err := fresh.SaveSleeve(ctx, own); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := fresh.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults: %v", err)
	}
	kinds, err := fresh.Kinds(ctx)
	checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
	sleeves, err := fresh.Sleeves(ctx)
	checkResult(t, "Sleeves", []portfolio.Sleeve{own}, sleeves, err)
}

// TestRegistryPlaceAccount puts accounts the ledger holds in stored
// sleeves, one sleeve each, and refuses an account the ledger doesn't hold
// or a sleeve that isn't stored.
func TestRegistryPlaceAccount(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	r, _ := seededRegistry(t)
	play := portfolio.Sleeve{ID: "play", Name: "Play", Kind: portfolio.KindSatellite}
	if err := r.SaveSleeve(ctx, play); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	steps := []struct {
		account ledger.AccountID
		sleeve  portfolio.SleeveID
		wantErr error
	}{
		{account: acc1, sleeve: "core"},
		{account: acc2, sleeve: "core"},
		{account: acc2, sleeve: "play"},
		{account: "acc-9", sleeve: "core", wantErr: errNoAccount},
		{account: acc1, sleeve: "gold", wantErr: app.ErrNotFound},
	}
	for _, s := range steps {
		if err := r.PlaceAccount(ctx, s.account, s.sleeve); !errors.Is(err, s.wantErr) {
			t.Errorf("PlaceAccount(%s, %s) = %v, want %v", s.account, s.sleeve, err, s.wantErr)
		}
	}
	in, err := r.AccountSleeves(ctx)
	checkResult(t, "AccountSleeves", map[ledger.AccountID]portfolio.SleeveID{acc1: "core", acc2: "play"}, in, err)
}

// TestRegistryDeletes checks D-120: a sleeve goes once no account is in
// it, and a kind once no sleeve has it.
func TestRegistryDeletes(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	r, store := seededRegistry(t)
	if err := r.PlaceAccount(ctx, acc1, "core"); err != nil {
		t.Fatalf("PlaceAccount: %v", err)
	}
	steps := []struct {
		name    string
		del     func() error
		wantErr error
	}{
		{"a sleeve with accounts", func() error { return r.DeleteSleeve(ctx, "core") }, portfolio.ErrSleeveHasAccounts},
		{"a kind a sleeve has", func() error { return r.DeleteKind(ctx, portfolio.KindCore) }, portfolio.ErrKindInUse},
		{"an unused built-in kind", func() error { return r.DeleteKind(ctx, portfolio.KindSatellite) }, nil},
		{"the account leaves its sleeve", func() error { return r.RemoveAccount(ctx, acc1) }, nil},
		{"an account in no sleeve", func() error { return r.RemoveAccount(ctx, acc1) }, nil},
		{"the sleeve with none", func() error { return r.DeleteSleeve(ctx, "core") }, nil},
	}
	for _, s := range steps {
		if err := s.del(); !errors.Is(err, s.wantErr) {
			t.Errorf("deleting %s = %v, want %v", s.name, err, s.wantErr)
		}
	}
	if _, ok := store.kinds[portfolio.KindSatellite]; ok {
		t.Error("the unused kind is still stored")
	}
	if len(store.sleeves) != 0 {
		t.Errorf("sleeves %v left, want none", store.sleeves)
	}
}
