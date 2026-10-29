package app_test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// fakeSleeveStore keeps kinds and sleeves in memory and checks them with
// the portfolio rules, as the ledger does.
type fakeSleeveStore struct {
	kinds   []portfolio.Kind
	sleeves []portfolio.Sleeve
	saves   int
	err     error
}

func (f *fakeSleeveStore) Kinds(context.Context) ([]portfolio.Kind, error) {
	return slices.Clone(f.kinds), f.err
}

func (f *fakeSleeveStore) SaveKind(_ context.Context, k portfolio.Kind) error {
	if err := portfolio.CheckKind(f.kinds, k); err != nil {
		return fmt.Errorf("checking the kind: %w", err)
	}
	f.saves++
	f.kinds = slices.DeleteFunc(f.kinds, func(old portfolio.Kind) bool { return old.Key == k.Key })
	f.kinds = append(f.kinds, k)
	slices.SortFunc(f.kinds, func(a, b portfolio.Kind) int { return strings.Compare(string(a.Key), string(b.Key)) })
	return nil
}

func (f *fakeSleeveStore) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	return slices.Clone(f.sleeves), f.err
}

func (f *fakeSleeveStore) SaveSleeve(_ context.Context, s portfolio.Sleeve) error {
	checked, err := portfolio.CheckSleeve(f.kinds, s)
	if err != nil {
		return fmt.Errorf("checking the sleeve: %w", err)
	}
	f.saves++
	f.sleeves = slices.DeleteFunc(f.sleeves, func(old portfolio.Sleeve) bool { return old.ID == checked.ID })
	f.sleeves = append(f.sleeves, checked)
	slices.SortFunc(f.sleeves, func(a, b portfolio.Sleeve) int { return strings.Compare(string(a.ID), string(b.ID)) })
	return nil
}

// newSleeves returns the sleeve entries over a fresh install's kinds and
// sleeves, with owners refusing other tokens with ErrNotOwner.
func newSleeves() (*app.Sleeves, *fakeSleeveStore) {
	store := &fakeSleeveStore{kinds: portfolio.DefaultKinds(), sleeves: portfolio.DefaultSleeves()}
	return &app.Sleeves{Owners: fakeOwners{err: app.ErrNotOwner}, Store: store}, store
}

func TestSleevesAddKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kind    string
		role    portfolio.Role
		want    portfolio.Kind
		wantErr error
	}{
		{
			name: "a kind of the owner's", kind: "Krypto główne", role: portfolio.RoleOffensive,
			want: portfolio.Kind{Key: "krypto-glowne", Name: "Krypto główne", DefaultRole: portfolio.RoleOffensive},
		},
		{name: "a key that is taken", kind: "Core", role: portfolio.RoleDefensive, wantErr: app.ErrExists},
		{name: "no key from the name", kind: "!!", role: portfolio.RoleDefensive, wantErr: portfolio.ErrInvalidKind},
		{name: "no role", kind: "Gold", wantErr: portfolio.ErrInvalidKind},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, store := newSleeves()
			got, err := s.AddKind(t.Context(), origin, owner, tt.kind, tt.role)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("AddKind error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AddKind (-want +got):\n%s", diff)
			}
			if tt.wantErr == nil && !slices.Contains(store.kinds, tt.want) {
				t.Errorf("kinds = %v, want %v stored", store.kinds, tt.want)
			}
		})
	}
}

func TestSleevesRenameKind(t *testing.T) {
	t.Parallel()
	s, store := newSleeves()
	got, err := s.RenameKind(t.Context(), origin, owner, portfolio.KindSatellite, "Play money")
	if err != nil {
		t.Fatalf("RenameKind: %v", err)
	}
	want := portfolio.Kind{Key: portfolio.KindSatellite, Name: "Play money", DefaultRole: portfolio.RoleOffensive}
	if diff := cmp.Diff(want, got); diff != "" || !slices.Contains(store.kinds, want) {
		t.Errorf("RenameKind (-want +got):\n%s\nstored %v", diff, store.kinds)
	}
	if _, err := s.RenameKind(t.Context(), origin, owner, "gold", "Gold"); !cmp.Equal(err, app.ErrNotFound, cmpopts.EquateErrors()) {
		t.Errorf("renaming an unknown kind = %v, want ErrNotFound", err)
	}
	if _, err := s.RenameKind(t.Context(), origin, owner, portfolio.KindCore, ""); !cmp.Equal(err, portfolio.ErrInvalidKind, cmpopts.EquateErrors()) {
		t.Errorf("renaming to nothing = %v, want ErrInvalidKind", err)
	}
}

func TestSleevesAddAndUpdateSleeve(t *testing.T) {
	t.Parallel()
	sleeve := func(id portfolio.SleeveID, name string, kind portfolio.KindKey, role portfolio.Role, override bool) portfolio.Sleeve {
		return portfolio.Sleeve{ID: id, Name: name, Kind: kind, Role: role, RoleOverride: override}
	}
	tests := []struct {
		name    string
		add     bool
		in      portfolio.Sleeve
		want    portfolio.Sleeve
		wantErr error
	}{
		{
			name: "the kind's role", add: true, in: sleeve("", "Agents", portfolio.KindAgentPool, 0, false),
			want: sleeve("agents", "Agents", portfolio.KindAgentPool, portfolio.RoleOffensive, false),
		},
		{
			name: "a defensive sleeve of an offensive kind, chosen explicitly", add: true, in: sleeve("ignored", "Cautious play", portfolio.KindSatellite, portfolio.RoleDefensive, true),
			want: sleeve("cautious-play", "Cautious play", portfolio.KindSatellite, portfolio.RoleDefensive, true),
		},
		{
			name: "another role without the explicit choice", add: true, in: sleeve("", "Play", portfolio.KindCore, portfolio.RoleOffensive, false),
			wantErr: portfolio.ErrRoleOverride,
		},
		{name: "a name whose ID is taken", add: true, in: sleeve("", "Bonds", portfolio.KindBonds, 0, false), wantErr: app.ErrExists},
		{name: "an unknown kind", add: true, in: sleeve("", "Gold", "gold", 0, false), wantErr: portfolio.ErrUnknownKind},
		{
			name: "rename and move a default sleeve", in: sleeve("core", "Long term", portfolio.KindRetirement, 0, false),
			want: sleeve("core", "Long term", portfolio.KindRetirement, portfolio.RoleDefensive, false),
		},
		{name: "an unknown sleeve", in: sleeve("gold", "Gold", portfolio.KindCore, 0, false), wantErr: app.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, store := newSleeves()
			do := s.UpdateSleeve
			if tt.add {
				do = s.AddSleeve
			}
			got, err := do(t.Context(), origin, owner, tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("sleeve (-want +got):\n%s", diff)
			}
			if tt.wantErr != nil && store.saves != 0 {
				t.Errorf("a refused sleeve stored %d times", store.saves)
			}
		})
	}
}

func TestSleevesOwnerOnly(t *testing.T) {
	t.Parallel()
	for _, refusal := range []error{app.ErrNoSession, app.ErrNotOwner, errStoreDown} {
		store := &fakeSleeveStore{kinds: portfolio.DefaultKinds(), sleeves: portfolio.DefaultSleeves()}
		s := &app.Sleeves{Owners: fakeOwners{err: refusal}, Store: store}
		ctx, read := t.Context(), "ofsess_read"
		_, errAddKind := s.AddKind(ctx, origin, read, "Gold", portfolio.RoleDefensive)
		_, errRename := s.RenameKind(ctx, origin, read, portfolio.KindCore, "Long")
		_, errAdd := s.AddSleeve(ctx, origin, read, portfolio.Sleeve{Name: "Gold", Kind: portfolio.KindCore})
		_, errUpdate := s.UpdateSleeve(ctx, origin, read, portfolio.Sleeve{ID: "core", Name: "Long", Kind: portfolio.KindCore})
		for _, err := range []error{errAddKind, errRename, errAdd, errUpdate} {
			if !cmp.Equal(refusal, err, cmpopts.EquateErrors()) {
				t.Errorf("a refused session = %v, want %v", err, refusal)
			}
		}
		if store.saves != 0 {
			t.Errorf("a refused session stored %d entries", store.saves)
		}
	}
}

func TestSleevesStoreFailure(t *testing.T) {
	t.Parallel()
	s, store := newSleeves()
	store.err = errStoreDown
	if _, err := s.AddKind(t.Context(), origin, owner, "Gold", portfolio.RoleDefensive); !cmp.Equal(err, errStoreDown, cmpopts.EquateErrors()) {
		t.Errorf("AddKind with the store down = %v, want %v", err, errStoreDown)
	}
	if _, err := s.AddSleeve(t.Context(), origin, owner, portfolio.Sleeve{Name: "Gold", Kind: portfolio.KindCore}); !cmp.Equal(err, errStoreDown, cmpopts.EquateErrors()) {
		t.Errorf("AddSleeve with the store down = %v, want %v", err, errStoreDown)
	}
}
