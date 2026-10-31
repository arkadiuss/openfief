package app_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// fakeRemovals is the ledger, ingest and valuation side of deleting an
// account and remapping a broker account: busy holds the accounts with
// transactions, by the key source prefix each came from.
type fakeRemovals struct {
	busy      map[ledger.AccountID]string
	brokers   *fakeBrokers
	deleted   []ledger.AccountID
	forgotten []string
	remapped  []string
}

func (f *fakeRemovals) DeleteAccount(_ context.Context, id ledger.AccountID) error {
	if _, ok := f.busy[id]; ok {
		return fmt.Errorf("account %s: %w", id, ledger.ErrAccountHasTransactions)
	}
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *fakeRemovals) HasTransactionsFrom(_ context.Context, account ledger.AccountID, prefix string) (bool, error) {
	src, ok := f.busy[account]
	return ok && strings.HasPrefix(src, prefix), nil
}

func (f *fakeRemovals) AllBrokerAccounts(context.Context) ([]ingest.BrokerAccount, error) {
	return f.brokers.accounts, nil
}

func (f *fakeRemovals) ForgetAccount(_ context.Context, account ledger.AccountID) error {
	f.forgotten = append(f.forgotten, string(account))
	return nil
}

func (f *fakeRemovals) RemapBrokerAccount(_ context.Context, instance string, id ingest.BrokerAccountID, account ledger.AccountID, _ time.Time) error {
	f.remapped = append(f.remapped, instance+"/"+string(id)+"->"+string(account))
	return nil
}

// withRemovals returns newAccounts' use cases with the removal ports over
// a fakeRemovals in which the bond account holds transactions from
// ibkr-main.
func withRemovals() (*app.Accounts, *fakeRemovals, *memBindings) {
	a, _, bindings, brokers := newAccounts()
	r := &fakeRemovals{busy: map[ledger.AccountID]string{bondsAccount: "plugin:ibkr-main:flex"}, brokers: brokers}
	a.Deletions, a.Connections, a.Forget = r, r, r
	return a, r, bindings
}

func TestAccountsDelete(t *testing.T) {
	t.Parallel()
	a, r, _ := withRemovals()
	ctx, origin := t.Context(), "https://ui.example.com"
	// D-120: an account with transactions stays, and nothing of it is forgotten.
	if err := a.Delete(ctx, origin, "owner", bondsAccount); !errors.Is(err, ledger.ErrAccountHasTransactions) {
		t.Errorf("deleting an account with transactions = %v, want ErrAccountHasTransactions", err)
	}
	if len(r.forgotten) != 0 {
		t.Errorf("a refused delete forgot %v", r.forgotten)
	}
	// One without goes from the ledger first, then from ingest and valuation.
	if err := a.Delete(ctx, origin, "owner", "acc-empty"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(r.deleted) != 1 || r.deleted[0] != "acc-empty" || strings.Join(r.forgotten, ",") != "acc-empty,acc-empty" {
		t.Errorf("deleted %v, forgot %v; want acc-empty deleted, then forgotten by ingest and valuation", r.deleted, r.forgotten)
	}
	if err := a.Delete(ctx, origin, "owner", ""); !errors.Is(err, app.ErrNoAccount) {
		t.Errorf("deleting no account = %v, want ErrNoAccount", err)
	}
	if err := a.Delete(ctx, origin, "read", "acc-empty"); !errors.Is(err, app.ErrNotOwner) {
		t.Errorf("a read session's delete = %v, want ErrNotOwner", err)
	}
}

func TestAccountsRemap(t *testing.T) {
	t.Parallel()
	ctx, origin := t.Context(), "https://ui.example.com"
	tests := []struct {
		name         string
		busySource   string
		broker       ingest.BrokerAccountID
		to           ledger.AccountID
		wantErr      error
		wantRemapped string
		wantBound    int
	}{
		{name: "an unmapped broker account to an account", broker: "U0000001", to: bondsAccount, wantRemapped: "ibkr-main/U0000001->" + string(bondsAccount), wantBound: 1},
		{name: "booked from its instance into its account", busySource: "plugin:ibkr-main:flex", broker: "U0000002", to: "", wantErr: app.ErrBrokerAccountHasTransactions},
		{name: "its account holds only another instance's", busySource: "plugin:bonds-main:x", broker: "U0000002", wantRemapped: "ibkr-main/U0000002->"},
		{name: "a broker account never reported", broker: "U0000009", to: bondsAccount, wantErr: app.ErrUnknownBrokerAccount},
		{name: "to an account not stored", broker: "U0000001", to: "acc-nowhere", wantErr: app.ErrNoAccount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, r, bindings := withRemovals()
			r.busy[bondsAccount] = tt.busySource
			if tt.busySource == "" {
				delete(r.busy, bondsAccount)
			}
			err := a.Remap(ctx, origin, "owner", "ibkr-main", tt.broker, tt.to)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Remap = %v, want %v", err, tt.wantErr)
			}
			if got := strings.Join(r.remapped, ","); got != tt.wantRemapped {
				t.Errorf("remapped %q, want %q", got, tt.wantRemapped)
			}
			if len(bindings.accounts) != tt.wantBound {
				t.Errorf("bound %v, want %d bindings", bindings.accounts, tt.wantBound)
			}
		})
	}
}

func TestAccountsBrokerAccounts(t *testing.T) {
	t.Parallel()
	a, _, _ := withRemovals()
	ctx, origin := t.Context(), "https://ui.example.com"
	all, err := a.BrokerAccounts(ctx, origin, "owner", false)
	if err != nil || len(all) != 2 {
		t.Errorf("BrokerAccounts = %v, %v; want both", all, err)
	}
	unmapped, err := a.BrokerAccounts(ctx, origin, "owner", true)
	if err != nil || len(unmapped) != 1 || unmapped[0].ID != "U0000001" {
		t.Errorf("unmapped BrokerAccounts = %v, %v; want U0000001", unmapped, err)
	}
	// D-121: only the owner sees broker account numbers.
	if b, err := a.BrokerAccounts(ctx, origin, "read", false); !errors.Is(err, app.ErrNotOwner) || len(b) != 0 {
		t.Errorf("a read session's BrokerAccounts = %v, %v; want ErrNotOwner and none", b, err)
	}
}
