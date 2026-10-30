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
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
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

// memTerms is app.AccountTerms in memory.
type memTerms struct {
	wrappers map[ledger.AccountID]jurisdiction.Wrapper
	fees     map[ledger.AccountID]jurisdiction.FeeRule
}

func (m *memTerms) RecordTaxWrapper(_ context.Context, account ledger.AccountID, w jurisdiction.Wrapper) error {
	m.wrappers[account] = w
	return nil
}

func (m *memTerms) RecordFeeRule(_ context.Context, account ledger.AccountID, r *jurisdiction.FeeRule) error {
	m.fees[account] = *r
	return nil
}

func (m *memTerms) TaxWrapper(_ context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	w, ok := m.wrappers[account]
	return w, ok, nil
}

func (m *memTerms) FeeRule(_ context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error) {
	return m.fees[account], nil
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
		Terms: &memTerms{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{}, fees: map[ledger.AccountID]jurisdiction.FeeRule{}},
		Pack:  pl.NewPack(), Now: func() time.Time { return now },
	}
	return a, ledgerFake, bindings, brokers
}

func TestAccountsCreate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		in          app.NewAccount
		want        app.AccountEntry
		wantErr     error
		wantBound   []ledger.AccountID
		wantMapping ledger.AccountID
	}{
		{
			name:      "a broker account the connector reported, its institution taken from it",
			in:        app.NewAccount{Name: "IBKR main", Sleeve: "satellite", Instance: "ibkr-main", BrokerAccount: "U0000001"},
			want:      entry(ledger.Account{ID: "ibkr-main", Institution: "ibkr", Name: "IBKR main", Sleeve: "satellite"}, "taxable", nil),
			wantBound: []ledger.AccountID{"ibkr-main"}, wantMapping: "ibkr-main",
		},
		{
			name:      "an account of a connector that names its accounts itself",
			in:        app.NewAccount{Name: "Obligacje skarbowe", Institution: "treasury-bonds", Sleeve: "bonds", Instance: "bonds-main"},
			want:      entry(ledger.Account{ID: "obligacje-skarbowe", Institution: "treasury-bonds", Name: "Obligacje skarbowe", Sleeve: "bonds"}, "taxable", nil),
			wantBound: []ledger.AccountID{"obligacje-skarbowe"},
		},
		{
			name: "an account with no connector yet",
			in:   app.NewAccount{Name: "XTB", Institution: "xtb", Sleeve: "core"},
			want: entry(ledger.Account{ID: "xtb", Institution: "xtb", Name: "XTB", Sleeve: "core"}, "taxable", nil),
		},
		{
			name: "an IKZE with its broker's fee rule",
			in:   app.NewAccount{Name: "XTB IKZE", Institution: "xtb", Sleeve: "core", Wrapper: "ikze", Fee: brokerFee()},
			want: entry(ledger.Account{ID: "xtb-ikze", Institution: "xtb", Name: "XTB IKZE", Sleeve: "core"}, "ikze", brokerFee()),
		},
		{name: "a wrapper the pack doesn't have", in: app.NewAccount{Name: "PPK", Institution: "x", Sleeve: "core", Wrapper: "ppk"}, wantErr: ledger.ErrInvalidAccount},
		{
			name: "a fee rate of 100%", wantErr: ledger.ErrInvalidAccount,
			in: app.NewAccount{Name: "Cash", Institution: "x", Sleeve: "core", Fee: &jurisdiction.FeeRule{Rate: kernel.NewDecimal(1, 0)}},
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
			checkTerms(t, a, &tt.want)
		})
	}
}

// brokerFee is a synthetic fee rule: 0.29% at least 14 PLN.
func brokerFee() *jurisdiction.FeeRule {
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		return nil
	}
	return &jurisdiction.FeeRule{Rate: kernel.NewDecimal(29, -4), Minimum: kernel.NewMoney(kernel.NewDecimal(14, 0), c)}
}

// entry returns acc with wrapper and fee, the zero rule when fee is nil.
func entry(acc ledger.Account, wrapper jurisdiction.Wrapper, fee *jurisdiction.FeeRule) app.AccountEntry {
	e := app.AccountEntry{Account: acc, Wrapper: wrapper}
	if fee != nil {
		e.Fee = *fee
	}
	return e
}

// checkTerms fails t unless want's wrapper is recorded for its account,
// and its fee rule when it is not the zero rule; nothing for no account.
func checkTerms(t *testing.T, a *app.Accounts, want *app.AccountEntry) {
	t.Helper()
	terms, ok := a.Terms.(*memTerms)
	if !ok {
		t.Fatal("not memTerms")
	}
	if diff := cmp.Diff(want.Wrapper, terms.wrappers[want.ID]); diff != "" {
		t.Errorf("recorded wrapper (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(want.Fee, terms.fees[want.ID]); diff != "" {
		t.Errorf("recorded fee rule (-want +got):\n%s", diff)
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
	got, err := a.Update(ctx, "https://ui.example.com", "owner", bondsAccount, &app.AccountChange{Name: "Treasury bonds", Sleeve: "core"})
	want := entry(ledger.Account{ID: bondsAccount, Institution: "treasury-bonds", Name: "Treasury bonds", Sleeve: "core"}, "taxable", nil)
	if diff := cmp.Diff(want, got); err != nil || diff != "" {
		t.Errorf("Update = %v (-want +got):\n%s", err, diff)
	}
	got, err = a.Update(ctx, "https://ui.example.com", "owner", bondsAccount, &app.AccountChange{Name: "Treasury bonds", Sleeve: "core", Wrapper: "ike", Fee: brokerFee()})
	want = entry(want.Account, "ike", brokerFee())
	if diff := cmp.Diff(want, got); err != nil || diff != "" {
		t.Errorf("Update of the terms = %v (-want +got):\n%s", err, diff)
	}
	checkTerms(t, a, &want)
	refusals := []struct {
		id      ledger.AccountID
		change  app.AccountChange
		wantErr error
	}{
		{id: "acc-none", change: app.AccountChange{Name: "X", Sleeve: "core"}, wantErr: app.ErrNoAccount},
		{id: bondsAccount, change: app.AccountChange{Name: "X", Sleeve: "gold"}, wantErr: app.ErrNoSleeve},
		{id: bondsAccount, change: app.AccountChange{Name: "X", Sleeve: "core", Wrapper: "ppk"}, wantErr: ledger.ErrInvalidAccount},
	}
	for _, r := range refusals {
		if _, err := a.Update(ctx, "https://ui.example.com", "owner", r.id, &r.change); !cmp.Equal(err, r.wantErr, cmpopts.EquateErrors()) {
			t.Errorf("Update(%s, %+v) = %v, want %v", r.id, r.change, err, r.wantErr)
		}
	}
	if l.saves != 2 {
		t.Errorf("%d account saves, want the two updates", l.saves)
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
		_, errUpdate := a.Update(ctx, origin, read, bondsAccount, &app.AccountChange{Name: "X", Sleeve: "core"})
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
