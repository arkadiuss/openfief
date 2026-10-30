package ghostfolio_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// withID returns a with Ghostfolio's ID id.
func withID(a *ghostfolio.Activity, id string) ghostfolio.Activity {
	out := *a
	out.ID = id
	return out
}

// inA2 returns a moved to the owned account of account a2.
func inA2(a *ghostfolio.Activity) ghostfolio.Activity {
	out := *a
	out.AccountID = ghostfolio.AccountUUID(instance, "a2")
	return out
}

// TestMakePlan checks what a sync changes for each difference between
// what the owned accounts should hold and what they hold (D-4, D-64).
func TestMakePlan(t *testing.T) {
	t.Parallel()
	a1, a2 := ghostfolio.AccountUUID(instance, "a1"), ghostfolio.AccountUUID(instance, "a2")
	accounts := []ghostfolio.Account{{ID: a1, Name: "One", Currency: cur(t, "PLN")}, {ID: a2, Name: "Two", Currency: cur(t, "PLN")}}
	buy := activity(t, "a1", "b1", ghostfolio.TypeBuy, "2025-01-03", "USD", yahoo, "ETF1.L", "3", "420.5", "1.25")
	sell := activity(t, "a1", "s1", ghostfolio.TypeSell, "2025-02-03", "USD", yahoo, "ETF1.L", "1", "430", "0")
	edited := withID(&buy, "g-edited")
	edited.Quantity = dec(t, "4")
	foreign := withID(&buy, "g-foreign")
	sellInA2 := inA2(&sell)
	foreign.AccountID = "a7c2a4c8-0000-4000-8000-000000000001"
	assets := []ghostfolio.AssetProfile{{Symbol: bondSym("EDO0136"), Name: "EDO0136", Currency: cur(t, "PLN")}}
	bond := activity(t, "a1", "e1", ghostfolio.TypeBuy, "2025-01-02", "PLN", ghostfolio.DataSourceManual, bondSym("EDO0136"), "10", "100", "0")
	editedBond := withID(&bond, "g-bond")
	editedBond.Quantity = dec(t, "11")
	// legacyBond is bond as an earlier version stored it, under a symbol
	// derived from the series code alone.
	legacyBond := withID(&bond, "g-legacy")
	legacyBond.Symbol = "4512a0f0-0f02-594a-8add-87da055fe65d"
	bal := func(id, account, day, value string) ghostfolio.Balance {
		return ghostfolio.Balance{ID: id, AccountID: account, Date: date(t, day), Value: dec(t, value)}
	}
	tests := []struct {
		name       string
		existing   []string
		desired    []ghostfolio.Activity
		actual     []ghostfolio.Activity
		desiredBal []ghostfolio.Balance
		actualBal  []ghostfolio.Balance
		want       ghostfolio.Plan
	}{
		{
			name:    "first sync: create the accounts, import everything",
			desired: []ghostfolio.Activity{buy, sell},
			want:    ghostfolio.Plan{CreateAccounts: accounts, Import: []ghostfolio.Activity{buy, sell}},
		},
		{
			name:     "nothing changed",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{buy, sell},
			actual:   []ghostfolio.Activity{withID(&sell, "g2"), withID(&buy, "g1")},
		},
		{
			name:     "a missing activity is imported",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{buy, sell},
			actual:   []ghostfolio.Activity{withID(&buy, "g1")},
			want:     ghostfolio.Plan{Import: []ghostfolio.Activity{sell}},
		},
		{
			name:     "a manual edit is reverted: import the desired one, delete the edited one",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{buy},
			actual:   []ghostfolio.Activity{edited},
			want:     ghostfolio.Plan{Import: []ghostfolio.Activity{buy}, Delete: []ghostfolio.Activity{edited}},
		},
		{
			name:     "a stale activity and a second copy are deleted one by one",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{buy},
			actual:   []ghostfolio.Activity{withID(&buy, "g1"), withID(&buy, "g2"), withID(&sell, "g3")},
			want:     ghostfolio.Plan{Delete: []ghostfolio.Activity{withID(&buy, "g2"), withID(&sell, "g3")}},
		},
		{
			name:     "an owned account that should hold nothing is emptied with a filtered bulk delete",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{buy},
			actual:   []ghostfolio.Activity{withID(&buy, "g1"), withID(&sellInA2, "g4")},
			want:     ghostfolio.Plan{Clear: []string{a2}},
		},
		{
			name:     "an activity moved to another owned account is deleted before the import it would block",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{inA2(&buy)},
			actual:   []ghostfolio.Activity{withID(&buy, "g1")},
			want:     ghostfolio.Plan{PreDelete: []ghostfolio.Activity{withID(&buy, "g1")}, Import: []ghostfolio.Activity{inA2(&buy)}},
		},
		{
			name:     "an activity on a leftover MANUAL asset is deleted before the import, which would reuse that asset by name",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{bond},
			actual:   []ghostfolio.Activity{legacyBond},
			want:     ghostfolio.Plan{PreDelete: []ghostfolio.Activity{legacyBond}, Import: []ghostfolio.Activity{bond}},
		},
		{
			name:     "an edited activity on a desired MANUAL asset is deleted after the import, keeping the asset's prices",
			existing: []string{a1, a2},
			desired:  []ghostfolio.Activity{bond},
			actual:   []ghostfolio.Activity{editedBond},
			want:     ghostfolio.Plan{Import: []ghostfolio.Activity{bond}, Delete: []ghostfolio.Activity{editedBond}},
		},
		{
			name:       "balances: kept, set when missing or different, deleted on a date with none, another account's untouched",
			existing:   []string{a1, a2, foreign.AccountID},
			desiredBal: []ghostfolio.Balance{bal("", a1, "2025-01-02", "100"), bal("", a1, "2025-01-03", "50"), bal("", a2, "2025-01-02", "7")},
			actualBal: []ghostfolio.Balance{
				bal("b1", a1, "2025-01-02", "100"), bal("b2", a1, "2025-01-03", "55"), bal("b3", a1, "2025-01-05", "1"),
				bal("b4", foreign.AccountID, "2025-01-02", "9"),
			},
			want: ghostfolio.Plan{
				SetBalances:    []ghostfolio.Balance{bal("", a1, "2025-01-03", "50"), bal("", a2, "2025-01-02", "7")},
				DeleteBalances: []ghostfolio.Balance{bal("b3", a1, "2025-01-05", "1")},
			},
		},
		{
			name:     "an activity outside the owned accounts is never touched",
			existing: []string{a1, a2, foreign.AccountID},
			desired:  []ghostfolio.Activity{buy},
			actual:   []ghostfolio.Activity{withID(&buy, "g1"), foreign},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			desired := ghostfolio.Desired{Accounts: accounts, Activities: tt.desired, Assets: assets, Balances: tt.desiredBal}
			got := ghostfolio.MakePlan(&desired, &ghostfolio.Actual{Accounts: tt.existing, Activities: tt.actual, Balances: tt.actualBal})
			want := tt.want
			want.Assets = assets
			if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("MakePlan (-want +got):\n%s", diff)
			}
			if got.IsEmpty() != (want.Import == nil && want.Delete == nil && want.PreDelete == nil && want.Clear == nil &&
				want.CreateAccounts == nil && want.SetBalances == nil && want.DeleteBalances == nil) {
				t.Errorf("IsEmpty = %v", got.IsEmpty())
			}
		})
	}
}

// TestCheckPreview checks that an import goes ahead only when its dry run
// shows every activity stored as sent.
func TestCheckPreview(t *testing.T) {
	t.Parallel()
	buy := activity(t, "a1", "b1", ghostfolio.TypeBuy, "2025-01-03", "PLN", ghostfolio.DataSourceManual, bondSym("EDO0136"), "3", "100", "0")
	fee := activity(t, "a1", "f1", ghostfolio.TypeFee, "2025-01-03", "PLN", ghostfolio.DataSourceManual, feeSym(t, "PLN"), "0", "0", "1")
	renamed := buy
	renamed.Symbol = "5a0c4d34-0000-4000-8000-000000000002"
	tests := []struct {
		name    string
		preview []ghostfolio.ImportedActivity
		wantErr error
	}{
		{"stored as sent, in another order", []ghostfolio.ImportedActivity{{Activity: fee}, {Activity: buy}}, nil},
		{"one would be skipped as a duplicate", []ghostfolio.ImportedActivity{{Activity: buy}, {Activity: fee, Duplicate: true}}, ghostfolio.ErrPreviewMismatch},
		{"one under the owner's own asset", []ghostfolio.ImportedActivity{{Activity: renamed}, {Activity: fee}}, ghostfolio.ErrPreviewMismatch},
		{"one under another account", []ghostfolio.ImportedActivity{{Activity: inA2(&buy)}, {Activity: fee}}, ghostfolio.ErrPreviewMismatch},
		{"one missing", []ghostfolio.ImportedActivity{{Activity: buy}}, ghostfolio.ErrPreviewMismatch},
		{"one replaced by another", []ghostfolio.ImportedActivity{{Activity: buy}, {Activity: buy}}, ghostfolio.ErrPreviewMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ghostfolio.CheckPreview([]ghostfolio.Activity{buy, fee}, tt.preview)
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil) != (err == nil) {
				t.Errorf("CheckPreview = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestBatches checks how an import is split: each batch with the assets its
// activities use, the first also with the assets no activity uses.
func TestBatches(t *testing.T) {
	t.Parallel()
	bond, other := bondSym("EDO0136"), bondSym("COI0128")
	acts := []ghostfolio.Activity{
		activity(t, "a1", "1", ghostfolio.TypeBuy, "2025-01-02", "PLN", ghostfolio.DataSourceManual, bond, "1", "100", "0"),
		activity(t, "a1", "2", ghostfolio.TypeBuy, "2025-01-03", "USD", yahoo, "ETF1.L", "1", "100", "0"),
		activity(t, "a1", "3", ghostfolio.TypeBuy, "2025-01-04", "PLN", ghostfolio.DataSourceManual, bond, "1", "100", "0"),
	}
	assets := []ghostfolio.AssetProfile{{Symbol: bond, Name: "EDO0136"}, {Symbol: other, Name: "COI0128"}}
	tests := []struct {
		name string
		acts []ghostfolio.Activity
		size int
		want []ghostfolio.Batch
	}{
		{"one batch", acts, 5, []ghostfolio.Batch{{Activities: acts, Assets: []ghostfolio.AssetProfile{assets[0], assets[1]}}}},
		{"two batches", acts, 2, []ghostfolio.Batch{
			{Activities: acts[:2], Assets: []ghostfolio.AssetProfile{assets[0], assets[1]}},
			{Activities: acts[2:], Assets: []ghostfolio.AssetProfile{assets[0]}},
		}},
		{"no activities: the assets alone", nil, 2, []ghostfolio.Batch{{Assets: assets}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, ghostfolio.Batches(tt.acts, assets, tt.size), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Batches (-want +got):\n%s", diff)
			}
		})
	}
}
