package app_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestOwnerServicesNeedTheChange checks that every owner service that
// changes an account asks the authorizer first (D-163): with the owner
// refused a change to the account, each one fails with ErrNotOwner and
// changes nothing.
func TestOwnerServicesNeedTheChange(t *testing.T) {
	t.Parallel()
	const origin, owner = "https://ui.example.com", "owner"
	denied := fakeOwners{err: app.ErrNotOwner, denied: []ledger.AccountID{bondsAccount, "acc-1"}}
	a, l, bindings, _ := newAccounts()
	a.Owners = denied
	caller := &fakeCaller{}
	b := bondLots(t, caller, &fakeLedger{}, &fakeStores{}, []string{"bonds-main"})
	b.Owners = denied
	u := uploads(t, caller, &fakeLedger{}, &fakeRecorder{})
	u.Owners = denied
	ctx := t.Context()
	lot := edoLot(t)
	_, errUpdate := a.Update(ctx, origin, owner, bondsAccount, &app.AccountChange{Name: "X", Sleeve: "core"})
	_, errAdd := b.Add(ctx, origin, owner, bondsAccount, &lot)
	_, errDeleteLot := b.Delete(ctx, origin, owner, bondsAccount, ingest.BondLotKey{})
	_, errUpload := u.Upload(ctx, origin, owner, "acc-1", 0, app.NamedFile{Name: "s.csv", Contents: strings.NewReader("x")})
	checks := map[string]error{
		"Update":                      errUpdate,
		"Bind":                        a.Bind(ctx, origin, owner, bondsAccount, "ibkr-main", ""),
		"Delete":                      a.Delete(ctx, origin, owner, bondsAccount),
		"Remap onto the account":      a.Remap(ctx, origin, owner, "ibkr-main", "U0000001", bondsAccount),
		"Remap away from the account": a.Remap(ctx, origin, owner, "ibkr-main", "U0000002", ""),
		"BondLots.Add":                errAdd,
		"BondLots.Delete":             errDeleteLot,
		"Uploads.Upload":              errUpload,
	}
	for name, err := range checks {
		if !cmp.Equal(app.ErrNotOwner, err, cmpopts.EquateErrors()) {
			t.Errorf("%s = %v, want %v", name, err, app.ErrNotOwner)
		}
	}
	if l.saves != 0 || len(bindings.accounts) != 0 || caller.id != "" {
		t.Errorf("refused changes saved %d accounts and %d bindings, and called connector %q",
			l.saves, len(bindings.accounts), caller.instance)
	}
}

func TestFileAccountsOnlyChangeable(t *testing.T) {
	t.Parallel()
	u := uploads(t, &fakeCaller{}, &fakeLedger{}, &fakeRecorder{})
	u.Bindings = perAccountBindings{"acc-1": {"flex-main"}, "acc-2": {"flex-main"}}
	u.Connectors.Bindings = perInstanceBindings{"flex-main": {"acc-2", "acc-1"}}
	u.Connectors.Declarations = fakeDeclarations{byInstance: map[string]ingest.Declarations{"flex-main": {StatementFiles: true}}}
	u.Connectors.Seeding = fixedSeeding{goLive: march2(t)}
	u.Instances = listed{names: []string{"flex-main"}}
	u.Owners = fakeOwners{denied: []ledger.AccountID{"acc-2"}}
	got, err := u.FileAccounts(t.Context(), "https://ui.example.com", "owner")
	if err != nil {
		t.Fatalf("FileAccounts: %v", err)
	}
	want := []app.FileAccount{{Account: "acc-1", Instance: "flex-main", Mode: ingest.ModeTransactions}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FileAccounts (-want +got):\n%s", diff)
	}
}
