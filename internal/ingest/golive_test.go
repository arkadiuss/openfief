package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// goLiveDate parses a synthetic go-live date.
func goLiveDate(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func TestPlanSync(t *testing.T) {
	t.Parallel()
	opening, transactions := ingest.ModeOpening, ingest.ModeTransactions
	tests := []struct {
		name   string
		bound  []ledger.AccountID
		seeded []ledger.AccountID
		want   []ingest.Mode
	}{
		{name: "no bound account, such as a market-data connector", want: []ingest.Mode{transactions}},
		{name: "every bound account seeded", bound: []ledger.AccountID{"acc-1", "acc-2"}, seeded: []ledger.AccountID{"acc-1", "acc-2"}, want: []ingest.Mode{transactions}},
		{name: "no bound account seeded", bound: []ledger.AccountID{"acc-1"}, want: []ingest.Mode{opening}},
		{name: "a seeded account of another instance doesn't count", bound: []ledger.AccountID{"acc-1"}, seeded: []ledger.AccountID{"acc-9"}, want: []ingest.Mode{opening}},
		{name: "one seeded, one newly mapped", bound: []ledger.AccountID{"acc-1", "acc-2"}, seeded: []ledger.AccountID{"acc-1"}, want: []ingest.Mode{opening, transactions}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			goLive := map[ledger.AccountID]kernel.Date{}
			for _, a := range tt.seeded {
				goLive[a] = goLiveDate(t, "2026-03-31")
			}
			if diff := cmp.Diff(tt.want, ingest.PlanSync(tt.bound, goLive)); diff != "" {
				t.Errorf("PlanSync (-want +got):\n%s", diff)
			}
		})
	}
}

// goLiveTx is a synthetic transaction of account, keyed by id, on day.
func goLiveTx(t *testing.T, account ledger.AccountID, kind ledger.TransactionKind, id, day string) ledger.Transaction {
	t.Helper()
	return ledger.Transaction{
		Key: ledger.TransactionKey{Source: "flex:" + string(account), ExternalID: id}, Account: account,
		Kind: kind, Date: goLiveDate(t, day),
	}
}

func TestKeepFromGoLive(t *testing.T) {
	t.Parallel()
	lot := goLiveTx(t, "acc-1", ledger.KindOpeningLot, "opening-lot:1", "2026-03-31")
	cash := goLiveTx(t, "acc-1", ledger.KindDeposit, "opening-cash:USD:2026-03-31", "2026-03-31")
	before := goLiveTx(t, "acc-1", ledger.KindBuy, "t-1", "2026-03-30")
	onGoLive := goLiveTx(t, "acc-1", ledger.KindBuy, "t-2", "2026-03-31")
	after := goLiveTx(t, "acc-1", ledger.KindBuy, "t-3", "2026-04-01")
	newLot := goLiveTx(t, "acc-2", ledger.KindOpeningLot, "opening-lot:2", "2026-04-02")
	newBuy := goLiveTx(t, "acc-2", ledger.KindBuy, "t-4", "2026-03-01")
	assertion := ledger.BalanceAssertion{Account: "acc-1"}
	earlier := ingest.Skipped("row#1", "summary row")
	m := ingest.Mapping{
		Transactions: []ledger.Transaction{lot, cash, before, onGoLive, after, newLot, newBuy},
		Assertions:   []ledger.BalanceAssertion{assertion},
		Problems:     []ingest.Problem{earlier},
	}
	goLive := map[ledger.AccountID]kernel.Date{"acc-1": goLiveDate(t, "2026-03-31")}
	const gone = "a transaction before go-live 2026-03-31"
	tests := []struct {
		name string
		mode ingest.Mode
		want ingest.Mapping
	}{
		{
			name: "opening: a seeded account's seeding is left out, the new account's kept",
			mode: ingest.ModeOpening,
			want: ingest.Mapping{
				Transactions: []ledger.Transaction{before, onGoLive, after, newLot, newBuy},
				Assertions:   []ledger.BalanceAssertion{assertion},
				Problems: []ingest.Problem{
					earlier,
					ingest.Skipped("flex:acc-1/opening-lot:1", "the account is already seeded"),
					ingest.Skipped("flex:acc-1/opening-cash:USD:2026-03-31", "the account is already seeded"),
				},
			},
		},
		{
			name: "transactions: a seeded account's transactions up to go-live are left out",
			mode: ingest.ModeTransactions,
			want: ingest.Mapping{
				Transactions: []ledger.Transaction{after, newLot, newBuy},
				Assertions:   []ledger.BalanceAssertion{assertion},
				Problems: []ingest.Problem{
					earlier,
					ingest.Skipped("flex:acc-1/opening-lot:1", gone),
					ingest.Skipped("flex:acc-1/opening-cash:USD:2026-03-31", gone),
					ingest.Skipped("flex:acc-1/t-1", gone),
					ingest.Skipped("flex:acc-1/t-2", gone),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ingest.KeepFromGoLive(&m, tt.mode, goLive)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("KeepFromGoLive (-want +got):\n%s", diff)
			}
			if len(m.Problems) != 1 || len(m.Transactions) != 7 {
				t.Errorf("KeepFromGoLive changed its input: %d problems, %d transactions", len(m.Problems), len(m.Transactions))
			}
		})
	}
}
