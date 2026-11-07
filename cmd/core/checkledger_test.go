package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
)

// TestParseCheckLedger checks check ledger's command line: --account is
// optional, and nothing else is taken.
func TestParseCheckLedger(t *testing.T) {
	t.Parallel()
	getenv := func(string) string { return "postgres://synthetic" }
	tests := []struct {
		name    string
		args    []string
		want    ledger.AccountID
		wantErr bool
	}{
		{name: "the whole log", args: []string{"check", "ledger"}},
		{name: "an account's versions", args: []string{"check", "ledger", "--account", "acc-xtb"}, want: "acc-xtb"},
		{name: "an argument", args: []string{"check", "ledger", "acc-xtb"}, wantErr: true},
		{name: "an unknown flag", args: []string{"check", "ledger", "--all"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, err := parseArgs(tt.args, getenv)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseArgs(%q) = %+v, want an error", tt.args, cmd)
				}
				return
			}
			if err != nil || cmd.Name != cmdCheckLedger || cmd.LedgerAccount != tt.want {
				t.Errorf("parseArgs(%q) = %s %q, %v; want %s %q", tt.args, cmd.Name, cmd.LedgerAccount, err, cmdCheckLedger, tt.want)
			}
		})
	}
}

// TestCheckLedger runs `core check ledger --account acc-xtb` on a real
// database holding seedBalances' deposit and buy, with the deposit then
// corrected to 1100.00 PLN: it writes a balanced report as JSON, with the
// superseded version counted and listed, and logs only that it ran, with
// counts, never a key, an account or an amount (GR-12).
func TestCheckLedger(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	_, _, pool := startCore(t, dsn)
	seedBalances(t, pool)
	correctDeposit(t, pool)
	cmd, err := parseArgs([]string{"check", "ledger", "--account", "acc-xtb"}, func(string) string { return dsn })
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	var out, logs bytes.Buffer
	if runErr := run(t.Context(), &cmd, &out, slog.New(slog.NewJSONHandler(&logs, nil))); runErr != nil {
		t.Fatalf("core check ledger: %v", runErr)
	}
	var got ledgerReport
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("decoding the report: %v\n%s", err, out.String())
	}
	want := ledgerReport{
		Balanced: true, Transactions: 2, Versions: 3, Superseded: 1,
		Accounts:   []ledgerAccountCount{{Account: "acc-xtb", Transactions: 2, Versions: 3, Superseded: 1}},
		Unbalanced: []ledgerUnbalanced{},
		Sums:       []ledgerAmount{{Amount: "0", Currency: "PLN"}},
		Logged: []ledgerVersion{
			{Source: "test", ExternalID: "buy-1", Version: 1, Kind: "buy"},
			{Source: "test", ExternalID: "dep-1", Version: 1, Kind: "deposit", Superseded: true},
			{Source: "test", ExternalID: "dep-1", Version: 2, Kind: "deposit"},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
	checkLedgerLogged(t, logs.String())
}

// correctDeposit imports seedBalances' deposit again at 1100.00 PLN, a
// second version that supersedes the first.
func correctDeposit(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	day, err := kernel.ParseDate("2026-03-02")
	if err != nil {
		t.Fatal(err)
	}
	header := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "dep-1"}, Account: "acc-xtb", Date: day}
	corrected, err := ledger.NewDeposit(header, kernel.NewMoney(kernel.NewDecimal(110000, -2), pln))
	if err != nil {
		t.Fatal(err)
	}
	if _, importErr := ledgerapp.New(ledgerpg.New(pool)).Import(t.Context(), []ledger.Transaction{corrected}, nil); importErr != nil {
		t.Fatalf("Import: %v", importErr)
	}
}

// checkLedgerLogged checks that logged says the ledger check ran with its
// counts, and names none of the log's keys, accounts or amounts (GR-12).
func checkLedgerLogged(t *testing.T, logged string) {
	t.Helper()
	if !strings.Contains(logged, `"msg":"ledger check ran"`) || !strings.Contains(logged, `"versions":3`) || !strings.Contains(logged, `"superseded":1`) {
		t.Errorf("log %q does not say the check ran with its counts", logged)
	}
	for _, value := range []string{"acc-xtb", "dep-1", "1100", "PLN"} {
		if strings.Contains(logged, value) {
			t.Errorf("log %q names %q", logged, value)
		}
	}
}
