package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestSaveSleeve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sleeve  portfolio.Sleeve
		wantErr error
	}{
		{name: "offensive", sleeve: portfolio.Sleeve{ID: "play", Name: "Play", Role: portfolio.RoleOffensive}},
		{name: "no role", sleeve: portfolio.Sleeve{ID: "play", Name: "Play"}, wantErr: portfolio.ErrInvalidSleeve},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, _ := seeded(t)
			err := l.SaveSleeve(t.Context(), tt.sleeve)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SaveSleeve error (-want +got):\n%s", diff)
			}
			want := []portfolio.Sleeve{{ID: core, Name: "Core", Role: portfolio.RoleDefensive}}
			if tt.wantErr == nil {
				want = append(want, tt.sleeve)
			}
			got, err := l.Sleeves(t.Context())
			checkResult(t, "Sleeves", want, got, err)
		})
	}
}

func TestSaveAccount(t *testing.T) {
	t.Parallel()
	account := func(id ledger.AccountID, sleeve portfolio.SleeveID) ledger.Account {
		return ledger.Account{ID: id, Institution: "test-broker", Name: string(id), Sleeve: sleeve}
	}
	tests := []struct {
		name    string
		account ledger.Account
		wantErr error
		want    []ledger.Account
	}{
		{
			name:    "new account",
			account: account("acc-3", core),
			want:    []ledger.Account{account(acc1, core), account(acc2, core), account("acc-3", core)},
		},
		{
			name:    "renamed account replaces the stored one",
			account: ledger.Account{ID: acc1, Institution: "test-broker", Name: "Renamed", Sleeve: core},
			want: []ledger.Account{
				{ID: acc1, Institution: "test-broker", Name: "Renamed", Sleeve: core},
				account(acc2, core),
			},
		},
		{
			name:    "unknown sleeve",
			account: account("acc-3", "nowhere"),
			wantErr: app.ErrUnknownSleeve,
			want:    []ledger.Account{account(acc1, core), account(acc2, core)},
		},
		{
			name:    "invalid account",
			account: ledger.Account{ID: "acc-3", Sleeve: core},
			wantErr: ledger.ErrInvalidAccount,
			want:    []ledger.Account{account(acc1, core), account(acc2, core)},
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
