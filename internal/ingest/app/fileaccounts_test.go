package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// perAccountBindings binds each account to its instances.
type perAccountBindings map[ledger.AccountID][]string

func (p perAccountBindings) BoundInstances(_ context.Context, account ledger.AccountID) ([]string, error) {
	return p[account], nil
}

// perInstanceBindings binds each instance to its accounts.
type perInstanceBindings map[string][]ledger.AccountID

func (p perInstanceBindings) BoundAccounts(_ context.Context, instance string) ([]ledger.AccountID, error) {
	return p[instance], nil
}

// TestUploadsChooseTheMode checks that the core picks an upload's mode,
// opening for an account not yet seeded and transactions for a seeded one,
// and that the owner's mode overrides it (D-109); and that an upload whose
// account's seeding can't be read imports nothing.
func TestUploadsChooseTheMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		seeded   bool
		override ingest.Mode
		want     ingest.Mode
	}{
		{name: "an account not yet seeded is seeded", want: ingest.ModeOpening},
		{name: "a seeded account takes transactions", seeded: true, want: ingest.ModeTransactions},
		{name: "the owner seeds a seeded account again", seeded: true, override: ingest.ModeOpening, want: ingest.ModeOpening},
		{name: "the owner sends transactions first", override: ingest.ModeTransactions, want: ingest.ModeTransactions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			caller := &fakeCaller{delivery: fileDelivery(t)}
			u := uploads(t, caller, &fakeLedger{}, &fakeRecorder{})
			if tt.seeded {
				u.Connectors.Seeding = fixedSeeding{goLive: march2(t)}
			}
			got, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", tt.override,
				app.NamedFile{Name: "s.xml", Contents: strings.NewReader("statement")})
			if err != nil {
				t.Fatalf("Upload: %v", err)
			}
			if got.Mode != tt.want || caller.mode != tt.want {
				t.Errorf("batch mode %s, handed %s; want %s", got.Mode, caller.mode, tt.want)
			}
		})
	}
	l := &fakeLedger{}
	u := uploads(t, &fakeCaller{delivery: fileDelivery(t)}, l, &fakeRecorder{})
	u.Connectors.Seeding = fixedSeeding{err: errStore}
	if _, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", 0,
		app.NamedFile{Name: "s.xml", Contents: strings.NewReader("statement")}); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) || l.calls != 0 {
		t.Errorf("Upload with the seeding unreadable = %v after %d imports, want %v and none", err, l.calls, errStore)
	}
}

// TestUploadsFileAccounts checks that the owner learns which accounts take
// statement files: those whose one bound, running instance declares it,
// with the mode of their next upload (D-109); and that only the owner may
// ask.
func TestUploadsFileAccounts(t *testing.T) {
	t.Parallel()
	u := uploads(t, &fakeCaller{}, &fakeLedger{}, &fakeRecorder{})
	u.Bindings = perAccountBindings{"acc-1": {"flex-main"}, "acc-2": {"flex-main"}, "acc-3": {"flex-main", "flex-b"}, "acc-4": {"bonds"}}
	u.Connectors.Bindings = perInstanceBindings{"flex-main": {"acc-2", "acc-1", "acc-3"}, "bonds": {"acc-4"}, "stopped": {"acc-5"}}
	u.Connectors.Declarations = fakeDeclarations{byInstance: map[string]ingest.Declarations{
		"flex-main": {StatementFiles: true, CrossCheckPrices: true}, "stopped": {StatementFiles: true},
	}}
	u.Connectors.Seeding = fixedSeeding{goLive: march2(t)}
	u.Instances = listed{names: []string{"bonds", "flex-main"}}
	got, err := u.FileAccounts(t.Context(), "https://ui.example.com", "owner")
	if err != nil {
		t.Fatalf("FileAccounts: %v", err)
	}
	want := []app.FileAccount{
		{Account: "acc-1", Instance: "flex-main", Mode: ingest.ModeTransactions},
		{Account: "acc-2", Instance: "flex-main", Mode: ingest.ModeOpening},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FileAccounts (-want +got):\n%s", diff)
	}
	u.Owners = fakeOwners{err: app.ErrNotOwner}
	if _, err := u.FileAccounts(t.Context(), "https://ui.example.com", "reader"); !cmp.Equal(err, app.ErrNotOwner, cmpopts.EquateErrors()) {
		t.Errorf("FileAccounts for a reader = %v, want ErrNotOwner", err)
	}
	none := uploads(t, &fakeCaller{}, &fakeLedger{}, &fakeRecorder{})
	if got, err := none.FileAccounts(t.Context(), "https://ui.example.com", "owner"); err != nil || len(got) != 0 {
		t.Errorf("FileAccounts with no instances found = %v, %v; want none", got, err)
	}
	u.Owners, u.Instances = fakeOwners{}, listed{err: errStore}
	if _, err := u.FileAccounts(t.Context(), "https://ui.example.com", "owner"); !cmp.Equal(err, errStore, cmpopts.EquateErrors()) {
		t.Errorf("FileAccounts with the instances unreadable = %v, want %v", err, errStore)
	}
}
