package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// changingOwners is the owner's session, ownerToken, which may change
// every account but denied, as the authorizer answers (D-163); it records
// the accounts it was asked about.
type changingOwners struct {
	denied ledger.AccountID
	asked  []ledger.AccountID
}

func (o *changingOwners) CheckOwner(_ context.Context, _, token string) error {
	if token != ownerToken {
		return app.ErrNotOwner
	}
	return nil
}

func (o *changingOwners) Changeable(ctx context.Context, origin, token string, accounts []ledger.AccountID) ([]ledger.AccountID, error) {
	if err := o.CheckOwner(ctx, origin, token); err != nil {
		return nil, err
	}
	o.asked = append(o.asked, accounts...)
	var out []ledger.AccountID
	for _, a := range accounts {
		if a != o.denied {
			out = append(out, a)
		}
	}
	return out, nil
}

// memStarts keeps the stored charts' starts in memory, by account.
type memStarts struct {
	stored map[ledger.AccountID]kernel.Date
	err    error
}

func (m *memStarts) ChartStart(_ context.Context, account ledger.AccountID) (kernel.Date, error) {
	return m.stored[account], m.err
}

func (m *memStarts) SaveChartStart(_ context.Context, account ledger.AccountID, start kernel.Date) error {
	if m.err != nil {
		return m.err
	}
	if start.IsZero() {
		delete(m.stored, account)
		return nil
	}
	m.stored[account] = start
	return nil
}

// TestSetChartStart checks the owner's setting of the charts' start
// (D-167): it is stored for the portfolio and for each account, going back
// to the default clears it, an account's needs the authorizer's leave to
// change it (D-163), and nothing is stored for a read session, a refused
// account or an account that is not stored.
func TestSetChartStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		token   string
		account ledger.AccountID
		start   string
		// before and want are the stored starts by account, the
		// portfolio's under "".
		before, want map[ledger.AccountID]string
		wantErr      error
	}{
		{
			name: "the portfolio's start is stored", token: ownerToken, start: "2026-03-02",
			want: map[ledger.AccountID]string{"": "2026-03-02"},
		},
		{
			name: "an account's start is stored apart from the portfolio's", token: ownerToken, account: acc1, start: "2026-03-05",
			before: map[ledger.AccountID]string{"": "2026-03-02"},
			want:   map[ledger.AccountID]string{"": "2026-03-02", acc1: "2026-03-05"},
		},
		{
			name: "back to the default clears the portfolio's", token: ownerToken,
			before: map[ledger.AccountID]string{"": "2026-03-02", acc1: "2026-03-05"},
			want:   map[ledger.AccountID]string{acc1: "2026-03-05"},
		},
		{
			name: "back to the default clears an account's", token: ownerToken, account: acc1,
			before: map[ledger.AccountID]string{"": "2026-03-02", acc1: "2026-03-05"},
			want:   map[ledger.AccountID]string{"": "2026-03-02"},
		},
		{
			name: "a read session may not", token: "ofsess_read", start: "2026-03-02",
			want: map[ledger.AccountID]string{}, wantErr: app.ErrNotOwner,
		},
		{
			name: "an account the owner may not change", token: ownerToken, account: acc2, start: "2026-03-02",
			want: map[ledger.AccountID]string{}, wantErr: app.ErrNotOwner,
		},
		{
			name: "an account that is not stored", token: ownerToken, account: "acc-9", start: "2026-03-02",
			want: map[ledger.AccountID]string{}, wantErr: app.ErrUnknownAccount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &memStarts{stored: dates(t, tt.before)}
			starts := app.NewChartStarts(&changingOwners{denied: acc2}, store, twoAccounts(t))
			var start kernel.Date
			if tt.start != "" {
				start = date(t, tt.start)
			}
			err := starts.Set(t.Context(), "https://fief.example.org", tt.token, tt.account, start)
			if !cmp.Equal(tt.wantErr, err, cmpopts.EquateErrors()) {
				t.Fatalf("Set: error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(dates(t, tt.want), store.stored); diff != "" {
				t.Errorf("stored starts (-want +got):\n%s", diff)
			}
		})
	}
}

// TestSetChartStartAsksBeforeReading checks that a refused account's
// setting fails the same whether the account is stored or not: the
// authorizer is asked before the accounts are listed (D-163), and a store
// or ledger failure is passed on.
func TestSetChartStartAsksBeforeReading(t *testing.T) {
	t.Parallel()
	owners := &changingOwners{denied: "acc-9"}
	l := twoAccounts(t)
	l.err = errStoreDown
	starts := app.NewChartStarts(owners, &memStarts{stored: map[ledger.AccountID]kernel.Date{}}, l)
	day := date(t, "2026-03-02")
	if err := starts.Set(t.Context(), "", ownerToken, "acc-9", day); !cmp.Equal(app.ErrNotOwner, err, cmpopts.EquateErrors()) {
		t.Errorf("a refused account that is not stored: error = %v, want ErrNotOwner", err)
	}
	if diff := cmp.Diff([]ledger.AccountID{"acc-9"}, owners.asked); diff != "" {
		t.Errorf("accounts asked about (-want +got):\n%s", diff)
	}
	if err := starts.Set(t.Context(), "", ownerToken, acc1, day); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("the ledger down: error = %v, want errStoreDown", err)
	}
	failing := app.NewChartStarts(owners, &memStarts{err: errStoreDown}, twoAccounts(t))
	if err := failing.Set(t.Context(), "", ownerToken, "", day); !cmp.Equal(errStoreDown, err, cmpopts.EquateErrors()) {
		t.Errorf("the store down: error = %v, want errStoreDown", err)
	}
}

// dates reads days by account as dates; nil reads as an empty map.
func dates(t *testing.T, days map[ledger.AccountID]string) map[ledger.AccountID]kernel.Date {
	t.Helper()
	out := map[ledger.AccountID]kernel.Date{}
	for acc, d := range days {
		out[acc] = date(t, d)
	}
	return out
}
