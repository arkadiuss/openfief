package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// plnAmount returns amount in PLN.
func plnAmount(t *testing.T, amount string) *kernel.Money {
	t.Helper()
	d, err := kernel.ParseDecimal(amount)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", amount, err)
	}
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	m := kernel.NewMoney(d, c)
	return &m
}

// TestCreateAccountFormAndEarlierPayments checks D-147 and D-149 at
// creation: an OKI account needs its form, one OKI has, and is stored
// with it; no other wrapper takes one; an IKE takes its earlier payments
// in PLN, not below zero, and no other wrapper takes them.
func TestCreateAccountFormAndEarlierPayments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		in       app.NewAccount
		wantForm jurisdiction.AccountForm
		wantPaid *kernel.Money
		wantErr  error
	}{
		{
			name: "an OKI brokerage account", in: app.NewAccount{Name: "OKI", Institution: "xtb", Sleeve: "core", Wrapper: "oki", Form: pl.FormBrokerage},
			wantForm: pl.FormBrokerage,
		},
		{name: "an OKI account without its form", in: app.NewAccount{Name: "OKI", Institution: "xtb", Sleeve: "core", Wrapper: "oki"}, wantErr: app.ErrAccountForm},
		{name: "a form OKI doesn't have", in: app.NewAccount{Name: "OKI", Institution: "xtb", Sleeve: "core", Wrapper: "oki", Form: "savings-book"}, wantErr: app.ErrAccountForm},
		{name: "a form for a taxable account", in: app.NewAccount{Name: "Main", Institution: "xtb", Sleeve: "core", Form: pl.FormBank}, wantErr: app.ErrAccountForm},
		{
			name: "an IKE with its earlier payments", in: app.NewAccount{Name: "IKE", Institution: "xtb", Sleeve: "core", Wrapper: "ike", EarlierPayments: plnAmount(t, "40000")},
			wantPaid: plnAmount(t, "40000"),
		},
		{name: "earlier payments of a taxable account", in: app.NewAccount{Name: "Main", Institution: "xtb", Sleeve: "core", EarlierPayments: plnAmount(t, "1")}, wantErr: app.ErrEarlierPayments},
		{name: "earlier payments below zero", in: app.NewAccount{Name: "IKE", Institution: "xtb", Sleeve: "core", Wrapper: "ike", EarlierPayments: plnAmount(t, "-1")}, wantErr: app.ErrEarlierPayments},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, l, _, _ := newAccounts()
			got, err := a.Create(t.Context(), "https://ui.example.com", "owner", &tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Create error (-want +got):\n%s", diff)
			}
			if err != nil {
				if !cmp.Equal(err, ledger.ErrInvalidAccount, cmpopts.EquateErrors()) || l.saves != 0 {
					t.Errorf("refusal %v is no invalid account, or saved %d times", err, l.saves)
				}
				return
			}
			checkTaxDetails(t, a, &got, tt.wantForm, tt.wantPaid)
		})
	}
}

// checkTaxDetails fails t unless got, as created, and what a recorded for
// it hold form and paid.
func checkTaxDetails(t *testing.T, a *app.Accounts, got *app.AccountEntry, form jurisdiction.AccountForm, paid *kernel.Money) {
	t.Helper()
	terms, ok := a.Terms.(*memTerms)
	if !ok {
		t.Fatal("not memTerms")
	}
	if got.Form != form || terms.forms[got.ID] != form {
		t.Errorf("form = %q, recorded %q; want %q", got.Form, terms.forms[got.ID], form)
	}
	if diff := cmp.Diff(paid, got.EarlierPayments); diff != "" {
		t.Errorf("earlier payments (-want +got):\n%s", diff)
	}
	if recorded, set := terms.paid[got.ID]; set != (paid != nil) || (set && !recorded.Equal(*paid)) {
		t.Errorf("recorded earlier payments %s, %v; want %v", recorded, set, paid)
	}
}

// TestUpdateAccountForm checks that an account moved into OKI takes its form
// then (D-147), which never changes after, and that an IKE takes its
// earlier payments when they are changed (D-149).
func TestUpdateAccountForm(t *testing.T) {
	t.Parallel()
	a, l, _, _ := newAccounts()
	ctx := t.Context()
	steps := []struct {
		name     string
		change   app.AccountChange
		wantForm jurisdiction.AccountForm
		wantErr  error
	}{
		{name: "into OKI without a form", change: app.AccountChange{Name: "B", Sleeve: "core", Wrapper: "oki"}, wantErr: app.ErrAccountForm},
		{name: "into OKI as a treasury bond account", change: app.AccountChange{Name: "B", Sleeve: "core", Wrapper: "oki", Form: pl.FormTreasuryBonds}, wantForm: pl.FormTreasuryBonds},
		{name: "renamed, keeping its form", change: app.AccountChange{Name: "Bonds", Sleeve: "core"}, wantForm: pl.FormTreasuryBonds},
		{name: "the same form again", change: app.AccountChange{Name: "Bonds", Sleeve: "core", Form: pl.FormTreasuryBonds}, wantForm: pl.FormTreasuryBonds},
		{name: "another form", change: app.AccountChange{Name: "Bonds", Sleeve: "core", Form: pl.FormBank}, wantErr: app.ErrAccountForm},
		{name: "earlier payments in OKI", change: app.AccountChange{Name: "Bonds", Sleeve: "core", EarlierPayments: plnAmount(t, "1")}, wantErr: app.ErrEarlierPayments},
	}
	for _, s := range steps {
		got, err := a.Update(ctx, "https://ui.example.com", "owner", bondsAccount, &s.change)
		if !cmp.Equal(err, s.wantErr, cmpopts.EquateErrors()) {
			t.Fatalf("%s: Update error = %v, want %v", s.name, err, s.wantErr)
		}
		if err == nil && got.Form != s.wantForm {
			t.Errorf("%s: form = %q, want %q", s.name, got.Form, s.wantForm)
		}
	}
	got, err := a.Update(ctx, "https://ui.example.com", "owner", bondsAccount,
		&app.AccountChange{Name: "Bonds", Sleeve: "core", Wrapper: "ike", EarlierPayments: plnAmount(t, "12000.50")})
	if err != nil || got.EarlierPayments == nil || !got.EarlierPayments.Equal(*plnAmount(t, "12000.50")) {
		t.Errorf("Update into IKE with earlier payments = %+v, %v", got.EarlierPayments, err)
	}
	if l.saves != 4 {
		t.Errorf("%d account saves, want the four accepted changes; a refused one saves nothing", l.saves)
	}
}
