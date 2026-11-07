package views_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// decimal parses s for a test.
func decimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", s, err)
	}
	return d
}

// TestCheckBalance checks the balance check's tolerance (D-200, criterion
// 1): cash passes within 0.01 of its currency, either way, and a holding
// only with equal units. Amounts are synthetic.
func TestCheckBalance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		holding                    bool
		reported, derived, wantDif string
		wantPassed                 bool
	}{
		{name: "cash that matches", reported: "1250.40", derived: "1250.40", wantDif: "0", wantPassed: true},
		{name: "cash 0.01 over", reported: "1250.41", derived: "1250.40", wantDif: "0.01", wantPassed: true},
		{name: "cash 0.01 under", reported: "1250.39", derived: "1250.40", wantDif: "-0.01", wantPassed: true},
		{name: "cash just past 0.01", reported: "1250.411", derived: "1250.40", wantDif: "0.011", wantPassed: false},
		{name: "cash a unit off", reported: "1251.40", derived: "1250.40", wantDif: "1", wantPassed: false},
		{name: "negative cash that matches", reported: "-12.30", derived: "-12.30", wantDif: "0", wantPassed: true},
		{name: "holding that matches", holding: true, reported: "15", derived: "15", wantDif: "0", wantPassed: true},
		{name: "holding a hundredth of a unit off", holding: true, reported: "15.01", derived: "15", wantDif: "0.01", wantPassed: false},
		{name: "holding with a unit missing", holding: true, reported: "15", derived: "16", wantDif: "-1", wantPassed: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := ledger.CashAssertion("acc-1", date(t, "2026-03-02"), money(t, tt.reported, "USD"))
			if tt.holding {
				a = ledger.HoldingAssertion("acc-1", date(t, "2026-03-02"), instrument(t, isinIE, "ETF1"),
					kernel.NewQuantity(decimal(t, tt.reported)))
			}
			got := views.CheckBalance(&a, decimal(t, tt.derived))
			want := views.BalanceCheck{
				Assertion: a, Derived: decimal(t, tt.derived), Difference: decimal(t, tt.wantDif), Passed: tt.wantPassed,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("CheckBalance (-want +got):\n%s", diff)
			}
		})
	}
}

// TestLatestBalanceChecks checks that the latest assertion of each account
// and ledger account stands for what the institution reported last, in
// account, date and ledger account order, and that every check counts in
// Failed.
func TestLatestBalanceChecks(t *testing.T) {
	t.Parallel()
	check := func(account ledger.AccountID, day string, cash bool, reported, derived string) views.BalanceCheck {
		a := ledger.CashAssertion(account, date(t, day), money(t, reported, "PLN"))
		if !cash {
			a = ledger.HoldingAssertion(account, date(t, day), instrument(t, isinPL, "EDO0334"),
				kernel.NewQuantity(decimal(t, reported)))
		}
		return views.CheckBalance(&a, decimal(t, derived))
	}
	oldCash := check("acc-1", "2026-03-01", true, "100", "90")
	newCash := check("acc-1", "2026-03-02", true, "100", "100")
	bonds := check("acc-1", "2026-02-27", false, "10", "10")
	other := check("acc-0", "2026-03-02", true, "5", "4")
	all := []views.BalanceCheck{newCash, other, oldCash, bonds}

	got := views.LatestBalanceChecks(all)
	want := []views.BalanceCheck{other, bonds, newCash}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("LatestBalanceChecks (-want +got):\n%s", diff)
	}
	if n := (&views.BalanceReport{Checks: got}).Failed(); n != 1 {
		t.Errorf("Failed of the latest = %d, want 1", n)
	}
	if n := (&views.BalanceReport{Checks: all}).Failed(); n != 2 {
		t.Errorf("Failed of all = %d, want 2", n)
	}
	views.SortBalanceChecks(all)
	if diff := cmp.Diff([]views.BalanceCheck{other, bonds, oldCash, newCash}, all); diff != "" {
		t.Errorf("SortBalanceChecks (-want +got):\n%s", diff)
	}
}

// TestLatestUnsupported checks that only each instance's latest completed
// batch counts, so open XTB CFDs closed since are no longer listed (D-24).
func TestLatestUnsupported(t *testing.T) {
	t.Parallel()
	at := func(hour int) time.Time { return time.Date(2026, 3, 2, hour, 0, 0, 0, time.UTC) }
	batch := func(id, instance string, hour, unsupported int) ingest.BatchRecord {
		return ingest.BatchRecord{
			ID: ingest.BatchID(id), Instance: instance, Institution: "xtb", Finished: at(hour),
			Accounts: []ledger.AccountID{"acc-x"}, Problems: unsupported + 1, Unsupported: unsupported,
		}
	}
	tests := []struct {
		name    string
		batches []ingest.BatchRecord
		want    []views.UnsupportedRows
	}{
		{name: "no batches", want: []views.UnsupportedRows{}},
		{
			name:    "open CFDs in the latest upload",
			batches: []ingest.BatchRecord{batch("b-1", "xtb-main", 9, 0), batch("b-2", "xtb-main", 10, 2)},
			want: []views.UnsupportedRows{{
				Instance: "xtb-main", Institution: "xtb", Batch: "b-2", Finished: at(10),
				Accounts: []ledger.AccountID{"acc-x"}, Rows: 2,
			}},
		},
		{
			name:    "CFDs closed since",
			batches: []ingest.BatchRecord{batch("b-2", "xtb-main", 11, 0), batch("b-1", "xtb-main", 10, 2)},
			want:    []views.UnsupportedRows{},
		},
		{
			name:    "the greater ID breaks a tie",
			batches: []ingest.BatchRecord{batch("b-2", "xtb-main", 10, 1), batch("b-1", "xtb-main", 10, 0)},
			want: []views.UnsupportedRows{{
				Instance: "xtb-main", Institution: "xtb", Batch: "b-2", Finished: at(10),
				Accounts: []ledger.AccountID{"acc-x"}, Rows: 1,
			}},
		},
		{
			name:    "instances in name order",
			batches: []ingest.BatchRecord{batch("b-1", "xtb-two", 9, 1), batch("b-2", "xtb-one", 9, 3)},
			want: []views.UnsupportedRows{
				{Instance: "xtb-one", Institution: "xtb", Batch: "b-2", Finished: at(9), Accounts: []ledger.AccountID{"acc-x"}, Rows: 3},
				{Instance: "xtb-two", Institution: "xtb", Batch: "b-1", Finished: at(9), Accounts: []ledger.AccountID{"acc-x"}, Rows: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, views.LatestUnsupported(tt.batches)); diff != "" {
				t.Errorf("LatestUnsupported (-want +got):\n%s", diff)
			}
		})
	}
}
