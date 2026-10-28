package ingest_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestCheckBound(t *testing.T) {
	t.Parallel()
	tx := func(account ledger.AccountID) ledger.Transaction { return ledger.Transaction{Account: account} }
	assertion := func(account ledger.AccountID) ledger.BalanceAssertion {
		return ledger.BalanceAssertion{Account: account}
	}
	tests := []struct {
		name  string
		bound []ledger.AccountID
		m     ingest.Mapping
		want  []ledger.AccountID
	}{
		{name: "empty mapping needs no binding"},
		{
			name: "every account bound", bound: []ledger.AccountID{"acc-1", "acc-2"},
			m: ingest.Mapping{Transactions: []ledger.Transaction{tx("acc-1"), tx("acc-2")}, Assertions: []ledger.BalanceAssertion{assertion("acc-2")}},
		},
		{
			name: "a transaction in another account", bound: []ledger.AccountID{"acc-1"},
			m:    ingest.Mapping{Transactions: []ledger.Transaction{tx("acc-1"), tx("acc-9")}},
			want: []ledger.AccountID{"acc-9"},
		},
		{
			name: "an assertion in another account", bound: []ledger.AccountID{"acc-1"},
			m:    ingest.Mapping{Transactions: []ledger.Transaction{tx("acc-1")}, Assertions: []ledger.BalanceAssertion{assertion("acc-8")}},
			want: []ledger.AccountID{"acc-8"},
		},
		{
			name: "no binding at all lists each account once, sorted",
			m:    ingest.Mapping{Transactions: []ledger.Transaction{tx("acc-2"), tx("acc-1"), tx("acc-2")}, Assertions: []ledger.BalanceAssertion{assertion("acc-1")}},
			want: []ledger.AccountID{"acc-1", "acc-2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ingest.CheckBound("xtb-main", tt.bound, &tt.m)
			if tt.want == nil {
				if err != nil {
					t.Fatalf("CheckBound: %v", err)
				}
				return
			}
			unbound, ok := errors.AsType[*ingest.UnboundAccountsError](err)
			if !ok || !errors.Is(err, ingest.ErrAccountNotBound) {
				t.Fatalf("CheckBound = %v, want an *UnboundAccountsError", err)
			}
			want := &ingest.UnboundAccountsError{Instance: "xtb-main", Accounts: tt.want}
			if diff := cmp.Diff(want, unbound); diff != "" {
				t.Errorf("unbound (-want +got):\n%s", diff)
			}
		})
	}
}

func TestValidateInstanceName(t *testing.T) {
	t.Parallel()
	long := "a" + strings.Repeat("b", 55)
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "simple", in: "ibkr-main"},
		{name: "digits", in: "nbp2"},
		{name: "56 characters", in: long},
		{name: "empty", in: "", wantErr: true},
		{name: "57 characters", in: long + "c", wantErr: true},
		{name: "starts with a digit", in: "2nbp", wantErr: true},
		{name: "ends with a hyphen", in: "nbp-", wantErr: true},
		{name: "upper case", in: "nBp", wantErr: true},
		{name: "dot", in: "nbp.main", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ingest.ValidateInstanceName(tt.in)
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, ingest.ErrInvalidInstanceName)) {
				t.Errorf("ValidateInstanceName(%q) = %v, want error %v", tt.in, err, tt.wantErr)
			}
		})
	}
}

func TestUnboundAccountsErrorMessage(t *testing.T) {
	t.Parallel()
	err := &ingest.UnboundAccountsError{Instance: "xtb-main", Accounts: []ledger.AccountID{"acc-1", "acc-2"}}
	if got, want := err.Error(), "connector instance xtb-main is not bound to account acc-1, acc-2"; got != want {
		t.Errorf("Error = %q, want %q", got, want)
	}
}
