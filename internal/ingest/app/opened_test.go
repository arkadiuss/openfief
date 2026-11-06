package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestAccountOpenedOnSeedingDay checks the owner's mark of an account
// opened on its seeding day (D-168): set when the account is created, kept
// by a change that doesn't name it, cleared and set again by one that
// does, stored with the account, and refused, with nothing stored, on an
// account the owner may not change (D-163).
func TestAccountOpenedOnSeedingDay(t *testing.T) {
	t.Parallel()
	a, l, _, _ := newAccounts()
	ctx := t.Context()
	created, err := a.Create(ctx, "https://ui.example.com", "owner", &app.NewAccount{
		Name: "XTB", Institution: "xtb", Sleeve: "core", OpenedOnSeedingDay: true,
	})
	if err != nil || !created.OpenedOnSeedingDay {
		t.Fatalf("Create = %+v, %v; want the account marked", created, err)
	}
	unmark, mark := false, true
	steps := []struct {
		name   string
		change app.AccountChange
		want   bool
	}{
		{name: "a rename keeps the mark", change: app.AccountChange{Name: "XTB renamed", Sleeve: "core"}, want: true},
		{name: "unmarked", change: app.AccountChange{Name: "XTB renamed", Sleeve: "core", OpenedOnSeedingDay: &unmark}, want: false},
		{name: "a rename keeps it unmarked", change: app.AccountChange{Name: "XTB", Sleeve: "core"}, want: false},
		{name: "marked again", change: app.AccountChange{Name: "XTB", Sleeve: "core", OpenedOnSeedingDay: &mark}, want: true},
	}
	for _, s := range steps {
		got, updateErr := a.Update(ctx, "https://ui.example.com", "owner", created.ID, &s.change)
		if updateErr != nil || got.OpenedOnSeedingDay != s.want {
			t.Errorf("%s: Update = %+v, %v; want marked %t", s.name, got.Account, updateErr, s.want)
		}
		checkStoredMark(t, l, created.ID, s.want)
	}
	saves := l.saves
	_, err = a.Update(ctx, "https://ui.example.com", "read", created.ID, &app.AccountChange{Name: "XTB", Sleeve: "core", OpenedOnSeedingDay: &unmark})
	if diff := cmp.Diff(app.ErrNotOwner, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("a read session's change (-want +got):\n%s", diff)
	}
	if l.saves != saves {
		t.Errorf("a refused change was saved")
	}
	if _, err := a.Update(ctx, "https://ui.example.com", "owner", ledger.AccountID("acc-none"), &app.AccountChange{Name: "X", Sleeve: "core", OpenedOnSeedingDay: &mark}); !cmp.Equal(app.ErrNoAccount, err, cmpopts.EquateErrors()) {
		t.Errorf("an account that is not stored: error = %v, want ErrNoAccount", err)
	}
}

// checkStoredMark fails t unless the stored account id is marked as want.
func checkStoredMark(t *testing.T, l *memAccounts, id ledger.AccountID, want bool) {
	t.Helper()
	stored, err := l.Accounts(t.Context())
	if err != nil {
		t.Fatalf("Accounts: %v", err)
	}
	for i := range stored {
		if stored[i].ID == id && stored[i].OpenedOnSeedingDay != want {
			t.Errorf("%s stored marked %t, want %t", id, stored[i].OpenedOnSeedingDay, want)
		}
	}
}
