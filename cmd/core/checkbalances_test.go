package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/ingest"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	viewscli "github.com/arkadiuss/openfief/internal/views/cli"
)

// seedBalances stores a synthetic XTB account with a deposit of 1000.00
// PLN and a buy of 4 units for 400.00 PLN, the broker's synthetic report
// of 600.00 PLN cash and 5 units, and an upload batch that left one open
// CFD position out (D-24).
func seedBalances(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := t.Context()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	day, err := kernel.ParseDate("2026-03-02")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	isin, err := ledger.ParseISIN("IETEST000026")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	etf := ledger.Instrument{ISIN: isin, Symbol: "ETF1"}
	l := ledgerapp.New(ledgerpg.New(pool))
	if saveErr := l.SaveAccount(ctx, ledger.Account{ID: "acc-xtb", Institution: "xtb", Name: "Synthetic XTB"}); saveErr != nil {
		t.Fatalf("SaveAccount: %v", saveErr)
	}
	header := func(id string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: id}, Account: "acc-xtb", Date: day}
	}
	deposit, err := ledger.NewDeposit(header("dep-1"), kernel.NewMoney(kernel.NewDecimal(100000, -2), pln))
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	buy, err := ledger.NewBuy(header("buy-1"), etf, kernel.NewQuantity(kernel.NewDecimal(4, 0)),
		kernel.NewMoney(kernel.NewDecimal(40000, -2), pln), kernel.NewMoney(kernel.Decimal{}, pln))
	if err != nil {
		t.Fatalf("NewBuy: %v", err)
	}
	assertions := []ledger.BalanceAssertion{
		ledger.CashAssertion("acc-xtb", day, kernel.NewMoney(kernel.NewDecimal(60000, -2), pln)),
		ledger.HoldingAssertion("acc-xtb", day, etf, kernel.NewQuantity(kernel.NewDecimal(5, 0))),
	}
	if _, importErr := l.Import(ctx, []ledger.Transaction{deposit, buy}, assertions); importErr != nil {
		t.Fatalf("Import: %v", importErr)
	}
	b, err := ingest.NewBatch("batch-1", "xtb", "upload", ingest.ModeTransactions, day)
	if err != nil {
		t.Fatalf("NewBatch: %v", err)
	}
	b.Problems = []ingest.Problem{ingest.Unsupported("Open Positions#2", "open CFD position (category \"CFD\") is not supported (D-24)")}
	b.Accounts = []ledger.AccountID{"acc-xtb"}
	if err := ingestpg.New(pool).RecordBatch(ctx, "xtb-main", &b); err != nil {
		t.Fatalf("RecordBatch: %v", err)
	}
}

// TestCheckBalances runs `core check balances` on a real database: it
// writes the owner's report as JSON, with the cash that matches passing,
// the holding a unit off failing and the open CFD flagged, and logs only
// that it ran, with counts, never an amount or a unit (GR-12).
func TestCheckBalances(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	_, _, pool := startCore(t, dsn)
	seedBalances(t, pool)
	cmd, err := parseArgs([]string{"check", "balances"}, func(string) string { return dsn })
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	var out, logs bytes.Buffer
	if runErr := run(t.Context(), &cmd, &out, slog.New(slog.NewJSONHandler(&logs, nil))); runErr != nil {
		t.Fatalf("core check balances: %v", runErr)
	}
	r, err := viewscli.ReadJSON(out.Bytes())
	if err != nil {
		t.Fatalf("ReadJSON: %v\n%s", err, out.String())
	}
	want := []viewscli.Check{
		{Account: "acc-xtb", Date: "2026-03-02", Kind: "cash", Currency: "PLN", Reported: "600", Derived: "600", Difference: "0", Passed: true},
		{Account: "acc-xtb", Date: "2026-03-02", Kind: "holding", ISIN: "IETEST000026", Symbol: "ETF1", Reported: "5", Derived: "4", Difference: "1"},
	}
	if diff := cmp.Diff(want, r.Checks); diff != "" {
		t.Errorf("checks (-want +got):\n%s", diff)
	}
	if r.Failed != 1 || len(r.Unsupported) != 1 || r.Unsupported[0].Instance != "xtb-main" || r.Unsupported[0].Rows != 1 {
		t.Errorf("report = %+v, want 1 failed and xtb-main's 1 unsupported row", r)
	}
	checkCountsOnly(t, logs.String())
}

// checkCountsOnly checks that logged says the balance check ran with its
// counts, and names none of seedBalances' values (GR-12).
func checkCountsOnly(t *testing.T, logged string) {
	t.Helper()
	if !strings.Contains(logged, `"msg":"balance check ran"`) || !strings.Contains(logged, `"checked":2`) || !strings.Contains(logged, `"failed":1`) {
		t.Errorf("log %q does not say the check ran with its counts", logged)
	}
	for _, value := range []string{"600", "acc-xtb", "IETEST000026", "ETF1"} {
		if strings.Contains(logged, value) {
			t.Errorf("log %q names %q", logged, value)
		}
	}
}
