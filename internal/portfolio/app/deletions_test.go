package app_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

// fakeDeletions deletes from store, refusing the sleeves in withAccounts
// and every kind a stored sleeve has.
type fakeDeletions struct {
	store        *fakeSleeveStore
	withAccounts map[portfolio.SleeveID]bool
}

func (f *fakeDeletions) DeleteSleeve(_ context.Context, id portfolio.SleeveID) error {
	if f.withAccounts[id] {
		return fmt.Errorf("sleeve %s: %w", id, portfolio.ErrSleeveHasAccounts)
	}
	f.store.sleeves = slices.DeleteFunc(f.store.sleeves, func(s portfolio.Sleeve) bool { return s.ID == id })
	return nil
}

func (f *fakeDeletions) DeleteKind(_ context.Context, key portfolio.KindKey) error {
	if slices.ContainsFunc(f.store.sleeves, func(s portfolio.Sleeve) bool { return s.Kind == key }) {
		return fmt.Errorf("kind %s: %w", key, portfolio.ErrKindInUse)
	}
	f.store.kinds = slices.DeleteFunc(f.store.kinds, func(k portfolio.Kind) bool { return k.Key == key })
	return nil
}

func TestSleevesDelete(t *testing.T) {
	t.Parallel()
	s, store := newSleeves()
	s.Deletions = &fakeDeletions{store: store, withAccounts: map[portfolio.SleeveID]bool{"core": true}}
	ctx := t.Context()
	// D-120: in order, so the satellite sleeve's going frees its kind.
	steps := []struct {
		name    string
		del     func() error
		wantErr error
	}{
		{"a sleeve with accounts", func() error { return s.DeleteSleeve(ctx, origin, owner, "core") }, portfolio.ErrSleeveHasAccounts},
		{"a kind a sleeve has", func() error { return s.DeleteKind(ctx, origin, owner, portfolio.KindSatellite) }, portfolio.ErrKindInUse},
		{"a sleeve with none", func() error { return s.DeleteSleeve(ctx, origin, owner, "satellite") }, nil},
		{"its kind, now unused", func() error { return s.DeleteKind(ctx, origin, owner, portfolio.KindSatellite) }, nil},
		{"an unknown sleeve", func() error { return s.DeleteSleeve(ctx, origin, owner, "nowhere") }, app.ErrNotFound},
		{"an unknown kind", func() error { return s.DeleteKind(ctx, origin, owner, "nothing") }, app.ErrNotFound},
		{"a read session's", func() error { return s.DeleteSleeve(ctx, origin, "read", "bonds") }, app.ErrNotOwner},
	}
	for _, st := range steps {
		if err := st.del(); !errors.Is(err, st.wantErr) {
			t.Errorf("deleting %s = %v, want %v", st.name, err, st.wantErr)
		}
	}
	if slices.ContainsFunc(store.sleeves, func(sl portfolio.Sleeve) bool { return sl.ID == "satellite" }) ||
		!slices.ContainsFunc(store.sleeves, func(sl portfolio.Sleeve) bool { return sl.ID == "bonds" }) {
		t.Errorf("sleeves = %v, want satellite gone and bonds kept", store.sleeves)
	}
}
