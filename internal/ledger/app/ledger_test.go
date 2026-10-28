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
	play := func(kind portfolio.KindKey, role portfolio.Role, override bool) portfolio.Sleeve {
		return portfolio.Sleeve{ID: "play", Name: "Play", Kind: kind, Role: role, OffensiveOverride: override}
	}
	tests := []struct {
		name    string
		sleeve  portfolio.Sleeve
		wantErr error
		// stored is the sleeve as stored; zero when saving fails.
		stored portfolio.Sleeve
	}{
		{
			name:   "offensive",
			sleeve: play(portfolio.KindSatellite, portfolio.RoleOffensive, false),
			stored: play(portfolio.KindSatellite, portfolio.RoleOffensive, false),
		},
		{
			name:   "role from the kind",
			sleeve: play(portfolio.KindAgentPool, 0, false),
			stored: play(portfolio.KindAgentPool, portfolio.RoleOffensive, false),
		},
		{
			name:   "defensive in an offensive kind",
			sleeve: play(portfolio.KindSatellite, portfolio.RoleDefensive, false),
			stored: play(portfolio.KindSatellite, portfolio.RoleDefensive, false),
		},
		{
			name:   "offensive in a defensive kind with the override",
			sleeve: play(portfolio.KindRetirement, portfolio.RoleOffensive, true),
			stored: play(portfolio.KindRetirement, portfolio.RoleOffensive, true),
		},
		{
			name:    "offensive in a defensive kind without the override",
			sleeve:  play(portfolio.KindRetirement, portfolio.RoleOffensive, false),
			wantErr: portfolio.ErrRoleOverride,
		},
		{name: "no kind", sleeve: play("", portfolio.RoleOffensive, false), wantErr: portfolio.ErrUnknownKind},
		{name: "kind not stored", sleeve: play("crypto", portfolio.RoleOffensive, false), wantErr: portfolio.ErrUnknownKind},
		{name: "invalid", sleeve: portfolio.Sleeve{ID: "play", Kind: portfolio.KindCore}, wantErr: portfolio.ErrInvalidSleeve},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, _ := seeded(t)
			err := l.SaveSleeve(t.Context(), tt.sleeve)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SaveSleeve error (-want +got):\n%s", diff)
			}
			want := []portfolio.Sleeve{coreSleeve()}
			if tt.wantErr == nil {
				want = append(want, tt.stored)
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
