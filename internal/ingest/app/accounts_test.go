package app_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// memAccounts is app.AccountLedger in memory, over the default sleeves.
type memAccounts struct {
	accounts []ledger.Account
	saves    int
	err      error
}

func (m *memAccounts) Accounts(context.Context) ([]ledger.Account, error) {
	return slices.Clone(m.accounts), m.err
}

func (m *memAccounts) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	return portfolio.DefaultSleeves(), m.err
}

func (m *memAccounts) SaveAccount(_ context.Context, a ledger.Account) error {
	m.saves++
	m.accounts = slices.DeleteFunc(m.accounts, func(old ledger.Account) bool { return old.ID == a.ID })
	m.accounts = append(m.accounts, a)
	return nil
}

// newAccounts returns the account entries over in-memory fakes holding
// one stored account, acc-bonds, and the synthetic broker accounts
// U0000001, unmapped, and U0000002, mapped to acc-bonds, reported by
// ibkr-main.
func newAccounts() (*app.Accounts, *memAccounts, *memBindings, *fakeBrokers) {
	ledgerFake := &memAccounts{accounts: []ledger.Account{{ID: bondsAccount, Institution: "treasury-bonds", Name: "Bonds", Sleeve: "bonds"}}}
	bindings := &memBindings{}
	brokers := &fakeBrokers{accounts: []ingest.BrokerAccount{
		{Instance: "ibkr-main", ID: "U0000001", Institution: "ibkr"},
		{Instance: "ibkr-main", ID: "U0000002", Institution: "ibkr", Account: bondsAccount},
	}}
	now := time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC)
	a := &app.Accounts{
		Owners: fakeOwners{err: app.ErrNotOwner}, Ledger: ledgerFake, Bindings: bindings, Brokers: brokers,
		Now: func() time.Time { return now },
	}
	return a, ledgerFake, bindings, brokers
}

func TestAccountsCreate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		in          app.NewAccount
		want        ledger.Account
		wantErr     error
		wantBound   []ledger.AccountID
		wantMapping ledger.AccountID
	}{
		{
			name:      "a broker account the connector reported, its institution taken from it",
			in:        app.NewAccount{Name: "IBKR main", Sleeve: "satellite", Instance: "ibkr-main", BrokerAccount: "U0000001"},
			want:      ledger.Account{ID: "ibkr-main", Institution: "ibkr", Name: "IBKR main", Sleeve: "satellite"},
			wantBound: []ledger.AccountID{"ibkr-main"}, wantMapping: "ibkr-main",
		},
		{
			name:      "an account of a connector that names its accounts itself",
			in:        app.NewAccount{Name: "Obligacje skarbowe", Institution: "treasury-bonds", Sleeve: "bonds", Instance: "bonds-main"},
			want:      ledger.Account{ID: "obligacje-skarbowe", Institution: "treasury-bonds", Name: "Obligacje skarbowe", Sleeve: "bonds"},
			wantBound: []ledger.AccountID{"obligacje-skarbowe"},
		},
		{
			name: "an account with no connector yet",
			in:   app.NewAccount{Name: "XTB", Institution: "xtb", Sleeve: "core"},
			want: ledger.Account{ID: "xtb", Institution: "xtb", Name: "XTB", Sleeve: "core"},
		},
		{name: "a name whose ID is taken", in: app.NewAccount{Name: "acc bonds", Institution: "x", Sleeve: "core"}, wantErr: app.ErrAccountExists},
		{name: "no ID from the name", in: app.NewAccount{Name: "--", Institution: "x", Sleeve: "core"}, wantErr: ledger.ErrInvalidAccount},
		{name: "no institution", in: app.NewAccount{Name: "Cash", Sleeve: "core"}, wantErr: ledger.ErrInvalidAccount},
		{name: "an unknown sleeve", in: app.NewAccount{Name: "Cash", Institution: "x", Sleeve: "gold"}, wantErr: app.ErrNoSleeve},
		{name: "a malformed instance", in: app.NewAccount{Name: "Cash", Institution: "x", Sleeve: "core", Instance: "IBKR"}, wantErr: ingest.ErrInvalidInstanceName},
		{name: "a broker account without its instance", in: app.NewAccount{Name: "Cash", Sleeve: "core", BrokerAccount: "U0000001"}, wantErr: app.ErrBrokerAccountNeedsInstance},
		{name: "a broker account the instance never reported", in: app.NewAccount{Name: "Cash", Sleeve: "core", Instance: "ibkr-main", BrokerAccount: "U0000009"}, wantErr: app.ErrUnknownBrokerAccount},
		{name: "a broker account mapped to another account", in: app.NewAccount{Name: "Cash", Sleeve: "core", Instance: "ibkr-main", BrokerAccount: "U0000002"}, wantErr: app.ErrBrokerAccountMapped},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, l, bindings, brokers := newAccounts()
			got, err := a.Create(t.Context(), "https://ui.example.com", "owner", &tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Create error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Create (-want +got):\n%s", diff)
			}
			checkStored(t, tt.wantErr == nil, l, bindings, brokers, tt.wantBound, tt.wantMapping)
		})
	}
}

// checkStored fails t unless an account was saved only when saved, the
// bindings are wantBound and U0000001 is mapped to wantMapping.
func checkStored(t *testing.T, saved bool, l *memAccounts, bindings *memBindings, brokers *fakeBrokers, wantBound []ledger.AccountID, wantMapping ledger.AccountID) {
	t.Helper()
	if !saved && l.saves != 0 {
		t.Errorf("a refused account was saved %d times", l.saves)
	}
	if diff := cmp.Diff(wantBound, bindings.accounts, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("bindings (-want +got):\n%s", diff)
	}
	if got := brokers.accounts[0].Account; got != wantMapping {
		t.Errorf("U0000001 mapped to %q, want %q", got, wantMapping)
	}
}

func TestAccountsUpdate(t *testing.T) {
	t.Parallel()
	a, l, _, _ := newAccounts()
	ctx := t.Context()
	got, err := a.Update(ctx, "https://ui.example.com", "owner", bondsAccount, "Treasury bonds", "core")
	want := ledger.Account{ID: bondsAccount, Institution: "treasury-bonds", Name: "Treasury bonds", Sleeve: "core"}
	if diff := cmp.Diff(want, got); err != nil || diff != "" {
		t.Errorf("Update = %v (-want +got):\n%s", err, diff)
	}
	if _, err := a.Update(ctx, "https://ui.example.com", "owner", "acc-none", "X", "core"); !cmp.Equal(err, app.ErrNoAccount, cmpopts.EquateErrors()) {
		t.Errorf("Update of an unknown account = %v, want ErrNoAccount", err)
	}
	if _, err := a.Update(ctx, "https://ui.example.com", "owner", bondsAccount, "X", "gold"); !cmp.Equal(err, app.ErrNoSleeve, cmpopts.EquateErrors()) {
		t.Errorf("Update into an unknown sleeve = %v, want ErrNoSleeve", err)
	}
	if l.saves != 1 {
		t.Errorf("%d account saves, want the one update", l.saves)
	}
}

func TestAccountsBind(t *testing.T) {
	t.Parallel()
	a, _, bindings, brokers := newAccounts()
	ctx := t.Context()
	if err := a.Bind(ctx, "https://ui.example.com", "owner", bondsAccount, "ibkr-main", "U0000001"); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if err := a.Bind(ctx, "https://ui.example.com", "owner", bondsAccount, "ibkr-main", ""); err != nil {
		t.Fatalf("Bind again without a broker account: %v", err)
	}
	if diff := cmp.Diff([]ledger.AccountID{bondsAccount, bondsAccount}, bindings.accounts); diff != "" || brokers.accounts[0].Account != bondsAccount {
		t.Errorf("bindings (-want +got):\n%s\nU0000001 mapped to %q", diff, brokers.accounts[0].Account)
	}
	for _, tt := range []struct {
		account  ledger.AccountID
		instance string
		wantErr  error
	}{
		{account: "acc-none", instance: "ibkr-main", wantErr: app.ErrNoAccount},
		{account: bondsAccount, instance: "", wantErr: ingest.ErrInvalidInstanceName},
	} {
		if err := a.Bind(ctx, "https://ui.example.com", "owner", tt.account, tt.instance, ""); !cmp.Equal(err, tt.wantErr, cmpopts.EquateErrors()) {
			t.Errorf("Bind(%s, %q) = %v, want %v", tt.account, tt.instance, err, tt.wantErr)
		}
	}
}

func TestAccountsOwnerOnly(t *testing.T) {
	t.Parallel()
	for _, refusal := range []error{app.ErrNoSession, app.ErrNotOwner, errStore} {
		a, l, bindings, _ := newAccounts()
		a.Owners = fakeOwners{err: refusal}
		ctx, origin, read := t.Context(), "https://ui.example.com", "read"
		_, errCreate := a.Create(ctx, origin, read, &app.NewAccount{Name: "X", Institution: "x", Sleeve: "core"})
		_, errUpdate := a.Update(ctx, origin, read, bondsAccount, "X", "core")
		errBind := a.Bind(ctx, origin, read, bondsAccount, "ibkr-main", "")
		for _, err := range []error{errCreate, errUpdate, errBind} {
			if !cmp.Equal(refusal, err, cmpopts.EquateErrors()) {
				t.Errorf("a refused session = %v, want %v", err, refusal)
			}
		}
		if l.saves != 0 || len(bindings.accounts) != 0 {
			t.Errorf("a refused session saved %d accounts and %d bindings", l.saves, len(bindings.accounts))
		}
	}
}
