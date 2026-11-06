package postgres_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// TestAccountFormsAndEarlierPayments stores an OKI account's form (D-147)
// and an IKE's earlier payments (D-149): none before any is stored, each
// read back as stored and replaced when saved again, and both gone once the
// account is forgotten.
func TestAccountFormsAndEarlierPayments(t *testing.T) {
	t.Parallel()
	store, _ := newStore(t)
	ctx := t.Context()
	type form struct {
		Form jurisdiction.AccountForm
		Set  bool
	}
	type paid struct {
		Amount string
		Set    bool
	}
	readForm := func(what string, want form) {
		t.Helper()
		f, ok, err := store.AccountForm(ctx, acc1)
		check(t, what, want, form{f, ok}, err)
	}
	readPaid := func(what string, want paid) {
		t.Helper()
		m, ok, err := store.EarlierPayments(ctx, acc1)
		got := paid{Set: ok}
		if ok {
			got.Amount = m.String()
		}
		check(t, what, want, got, err)
	}
	readForm("form before any", form{})
	readPaid("earlier payments before any", paid{})
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []jurisdiction.AccountForm{"bank-account", "brokerage-account"} {
		if saveErr := store.SaveAccountForm(ctx, acc1, f); saveErr != nil {
			t.Fatalf("SaveAccountForm: %v", saveErr)
		}
		readForm("form after saving "+string(f), form{f, true})
	}
	for _, amount := range []int64{30000, 12500} {
		if saveErr := store.SaveEarlierPayments(ctx, acc1, kernel.NewMoney(kernel.NewDecimal(amount, 0), pln)); saveErr != nil {
			t.Fatalf("SaveEarlierPayments: %v", saveErr)
		}
	}
	readPaid("earlier payments after saving twice", paid{"12500 PLN", true})
	if forgetErr := store.ForgetAccount(ctx, acc1); forgetErr != nil {
		t.Fatalf("ForgetAccount: %v", forgetErr)
	}
	readForm("form after forgetting the account", form{})
	readPaid("earlier payments after forgetting the account", paid{})
}
