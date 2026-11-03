package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestSaveAccount(t *testing.T) {
	t.Parallel()
	account := func(id ledger.AccountID) ledger.Account {
		return ledger.Account{ID: id, Institution: "test-broker", Name: string(id)}
	}
	tests := []struct {
		name    string
		account ledger.Account
		wantErr error
		want    []ledger.Account
	}{
		{
			name:    "new account",
			account: account("acc-3"),
			want:    []ledger.Account{account(acc1), account(acc2), account("acc-3")},
		},
		{
			name:    "renamed account replaces the stored one",
			account: ledger.Account{ID: acc1, Institution: "test-broker", Name: "Renamed"},
			want: []ledger.Account{
				{ID: acc1, Institution: "test-broker", Name: "Renamed"},
				account(acc2),
			},
		},
		{
			name:    "invalid account",
			account: ledger.Account{ID: "acc-3"},
			wantErr: ledger.ErrInvalidAccount,
			want:    []ledger.Account{account(acc1), account(acc2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, _ := seeded(t)
			err := l.SaveAccount(t.Context(), tt.account)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SaveAccount error (-want +got):\n%s", diff)
			}
			got, err := l.Accounts(t.Context())
			checkResult(t, "Accounts", tt.want, got, err)
			if tt.wantErr == nil {
				one, err := l.Account(t.Context(), tt.account.ID)
				checkResult(t, "Account", tt.account, one, err)
			}
		})
	}
}
