package ghostfolio_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// fakeGhostfolio is Ghostfolio as ghostfolioModel models it, behind the
// Ghostfolio port. It logs every call, and can turn the restricted view on
// or answer every dry run with duplicates.
type fakeGhostfolio struct {
	ghostfolioModel

	restricted    bool
	duplicateDrys bool
	calls         []string
}

func (f *fakeGhostfolio) Login(context.Context) error {
	f.calls = append(f.calls, "login")
	return nil
}

func (f *fakeGhostfolio) User(context.Context) (ghostfolio.User, error) {
	f.calls = append(f.calls, "user")
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		return ghostfolio.User{}, fmt.Errorf("base currency: %w", err)
	}
	return ghostfolio.User{BaseCurrency: c, RestrictedView: f.restricted}, nil
}

func (f *fakeGhostfolio) Accounts(context.Context) ([]string, error) {
	f.calls = append(f.calls, "accounts")
	return slices.Clone(f.accounts), nil
}

func (f *fakeGhostfolio) CreateAccount(_ context.Context, a ghostfolio.Account) error {
	f.calls = append(f.calls, "create account")
	f.accounts = append(f.accounts, a.ID)
	return nil
}

func (f *fakeGhostfolio) Activities(_ context.Context, accounts []string) ([]ghostfolio.Activity, error) {
	f.calls = append(f.calls, "activities")
	var out []ghostfolio.Activity
	for i := range f.activities {
		if slices.Contains(accounts, f.activities[i].AccountID) {
			out = append(out, f.activities[i])
		}
	}
	return out, nil
}

func (f *fakeGhostfolio) Import(_ context.Context, b *ghostfolio.Batch, dryRun bool) ([]ghostfolio.ImportedActivity, error) {
	var out []ghostfolio.ImportedActivity
	if dryRun {
		f.calls = append(f.calls, fmt.Sprintf("dry run %d", len(b.Activities)))
		for i := range b.Activities {
			out = append(out, ghostfolio.ImportedActivity{Activity: b.Activities[i], Duplicate: f.duplicateDrys})
		}
		return out, nil
	}
	f.calls = append(f.calls, fmt.Sprintf("import %d", len(b.Activities)))
	for i := range b.Activities {
		a := b.Activities[i]
		if slices.ContainsFunc(f.activities, func(x ghostfolio.Activity) bool { return duplicate(&a, &x) }) {
			continue
		}
		f.next++
		a.ID = fmt.Sprintf("g-%03d", f.next)
		f.activities = append(f.activities, a)
		out = append(out, ghostfolio.ImportedActivity{Activity: a})
	}
	return out, nil
}

func (f *fakeGhostfolio) DeleteActivity(_ context.Context, id string) error {
	f.calls = append(f.calls, "delete")
	f.activities = slices.DeleteFunc(f.activities, func(a ghostfolio.Activity) bool { return a.ID == id })
	return nil
}

func (f *fakeGhostfolio) DeleteActivitiesInAccounts(_ context.Context, accounts []string) error {
	if len(accounts) == 0 {
		return ghostfolio.ErrNoAccountFilter
	}
	f.calls = append(f.calls, "bulk delete")
	f.activities = slices.DeleteFunc(f.activities, func(a ghostfolio.Activity) bool { return slices.Contains(accounts, a.AccountID) })
	return nil
}

// TestSync runs the reconcile loop over the synthetic record and a fake
// Ghostfolio: the first sync creates the owned accounts and imports, the
// second changes nothing but the prices, a manual edit is reverted, a
// removed transaction is deleted, and an account left with nothing is
// emptied with a filtered bulk delete (D-4, D-64).
func TestSync(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	record := syntheticRecord()
	g := &fakeGhostfolio{ghostfolioModel: ghostfolioModel{accounts: []string{foreignAccount}}}
	s := ghostfolio.NewSyncer(record, g, instance, nil)
	steps := []struct {
		name      string
		before    func()
		want      ghostfolio.Report
		wantCalls string
	}{
		{
			name: "first sync", want: ghostfolio.Report{AccountsCreated: 2, Imported: 3},
			wantCalls: "login user accounts create account create account dry run 3 import 3",
		},
		{
			name: "nothing changed: only the prices are sent", want: ghostfolio.Report{},
			wantCalls: "login user accounts activities dry run 0 import 0",
		},
		{
			name: "a manual edit is reverted",
			before: func() {
				i := slices.IndexFunc(g.activities, func(a ghostfolio.Activity) bool { return a.Comment == comment("bonds", "buy-1") })
				g.activities[i].Quantity = g.activities[i].Quantity.Add(dec(t, "1"))
			},
			want:      ghostfolio.Report{Imported: 1, Deleted: 1},
			wantCalls: "login user accounts activities dry run 1 import 1 delete",
		},
		{
			name:      "a removed transaction is deleted",
			before:    func() { record.withoutBond("fee-1") },
			want:      ghostfolio.Report{Deleted: 1},
			wantCalls: "login user accounts activities dry run 0 import 0 delete",
		},
		{
			name: "an account left with nothing is emptied",
			before: func() {
				record.withoutBond("open-1")
				record.withoutBond("buy-1")
			},
			want:      ghostfolio.Report{AccountsCleared: 1},
			wantCalls: "login user accounts activities bulk delete",
		},
	}
	for _, st := range steps {
		if st.before != nil {
			st.before()
		}
		g.calls = nil
		got, err := s.Sync(ctx)
		if err != nil {
			t.Fatalf("%s: Sync: %v", st.name, err)
		}
		if diff := cmp.Diff(st.want, got); diff != "" {
			t.Errorf("%s: report (-want +got):\n%s", st.name, diff)
		}
		if calls := strings.Join(g.calls, " "); calls != st.wantCalls {
			t.Errorf("%s: calls\n got %s\nwant %s", st.name, calls, st.wantCalls)
		}
	}
}

// TestSyncRefuses checks the syncs that change nothing: with the
// restricted view on, and when a dry run shows a duplicate.
func TestSyncRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		fake      fakeGhostfolio
		wantErr   error
		wantCalls string
	}{
		{"restricted view", fakeGhostfolio{restricted: true}, ghostfolio.ErrRestrictedView, "login user"},
		{"a dry run with a duplicate", fakeGhostfolio{duplicateDrys: true}, ghostfolio.ErrPreviewMismatch, "login user accounts create account create account dry run 3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := tt.fake
			_, err := ghostfolio.NewSyncer(syntheticRecord(), &g, instance, map[ledger.ISIN]ghostfolio.Listing{}).Sync(t.Context())
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Sync = %v, want %v", err, tt.wantErr)
			}
			if calls := strings.Join(g.calls, " "); calls != tt.wantCalls {
				t.Errorf("calls\n got %s\nwant %s", calls, tt.wantCalls)
			}
			if len(g.activities) != 0 {
				t.Errorf("%d activities written", len(g.activities))
			}
		})
	}
}
