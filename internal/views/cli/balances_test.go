package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/cli"
)

var errExecRefused = errors.New(`pods "openfief-core-0" is forbidden: cannot create resource "pods/exec"`)

// fakeCore answers Exec with out, or fails with err, and records the
// arguments it ran.
type fakeCore struct {
	out  []byte
	err  error
	args []string
}

func (f *fakeCore) Exec(_ context.Context, args []string) ([]byte, error) {
	f.args = args
	return f.out, f.err
}

// syntheticReport is a report over synthetic accounts: cash that matches
// within 0.01, a holding a unit off when failing is set, and an XTB
// instance's latest upload with two open CFD positions left out (D-24).
func syntheticReport(t *testing.T, failing bool) views.BalanceReport {
	t.Helper()
	usd, err := kernel.ParseCurrency("USD")
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
	cash := ledger.CashAssertion("acc-ibkr", day, kernel.NewMoney(kernel.NewDecimal(125040, -2), usd))
	units := ledger.HoldingAssertion("acc-ibkr", day, ledger.Instrument{ISIN: isin, Symbol: "ETF1"}, kernel.NewQuantity(kernel.NewDecimal(15, 0)))
	derived := kernel.NewDecimal(15, 0)
	if failing {
		derived = kernel.NewDecimal(14, 0)
	}
	return views.BalanceReport{
		Checks: []views.BalanceCheck{
			views.CheckBalance(&cash, kernel.NewDecimal(125039, -2)),
			views.CheckBalance(&units, derived),
		},
		Unsupported: []views.UnsupportedRows{{
			Instance: "xtb-main", Institution: "xtb", Batch: "batch-7", Finished: time.Date(2026, 3, 2, 18, 4, 0, 0, time.UTC),
			Accounts: []ledger.AccountID{"acc-xtb"}, Rows: 2,
		}},
	}
}

// coreJSON is what `core check balances` writes for r.
func coreJSON(t *testing.T, r *views.BalanceReport) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := cli.WriteJSON(&b, r); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	return b.Bytes()
}

const passingTable = `ACCOUNT   DATE        BALANCE            REPORTED  DERIVED  DIFFERENCE  RESULT
acc-ibkr  2026-03-02  cash USD           1250.4    1250.39  0.01        pass
acc-ibkr  2026-03-02  ETF1 IETEST000026  15        15       0           pass

FLAGGED, UNSUPPORTED: not checked (D-24); open XTB CFDs are among them
INSTANCE  INSTITUTION  LATEST BATCH  FINISHED              ACCOUNTS  ROWS
xtb-main  xtb          batch-7       2026-03-02T18:04:00Z  acc-xtb   2

2 balance assertions checked, 0 failed: cash passes within 0.01 of its currency, a holding with equal units.
`

// ran is what one run of the command did.
type ran struct {
	out      []byte
	coreArgs []string
	err      error
}

// runOn runs `openfief check` with args on a core that answers with the
// synthetic report, failing or not.
func runOn(t *testing.T, args []string, failing bool) (ran, views.BalanceReport) {
	t.Helper()
	r := syntheticReport(t, failing)
	core := &fakeCore{out: coreJSON(t, &r)}
	var out bytes.Buffer
	err := cli.Run(t.Context(), args, &cli.Env{Out: &out, Connect: func(string, string) (cli.Core, error) { return core, nil }})
	return ran{out: out.Bytes(), coreArgs: core.args, err: err}, r
}

// TestRunTable checks `openfief check balances`: it runs the core's check
// in its pod with --all only when asked, prints the table, and fails with
// ErrBalancesFailed, after printing everything, when any balance assertion
// failed, so a script can act on its exit status.
func TestRunTable(t *testing.T) {
	t.Parallel()
	failingTable := strings.NewReplacer(
		"15        15       0           pass", "15        14       1           FAIL",
		"2 balance assertions checked, 0 failed", "2 balance assertions checked, 1 failed",
	).Replace(passingTable)
	tests := []struct {
		name     string
		args     []string
		failing  bool
		wantArgs []string
		wantErr  error
		want     string
	}{
		{name: "the latest", args: []string{"balances"}, wantArgs: []string{"check", "balances"}, want: passingTable},
		{name: "all, flags first", args: []string{"--all", "balances"}, wantArgs: []string{"check", "balances", "--all"}, want: passingTable},
		{
			name: "a failed assertion", args: []string{"balances"}, failing: true, wantArgs: []string{"check", "balances"},
			wantErr: cli.ErrBalancesFailed, want: failingTable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := runOn(t, tt.args, tt.failing)
			if !errors.Is(got.err, tt.wantErr) || (tt.wantErr == nil) != (got.err == nil) {
				t.Errorf("Run: error %v, want %v", got.err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.wantArgs, got.coreArgs); diff != "" {
				t.Errorf("core args (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, string(got.out)); diff != "" {
				t.Errorf("output (-want +got):\n%s", diff)
			}
		})
	}
}

// TestRunJSON checks `openfief check balances --json`: it prints the
// report in its JSON shape, and fails with ErrBalancesFailed when any
// balance assertion failed.
func TestRunJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		failing bool
		wantErr error
	}{
		{name: "all pass", args: []string{"balances", "--json"}},
		{name: "a failed assertion", args: []string{"--json", "balances"}, failing: true, wantErr: cli.ErrBalancesFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, r := runOn(t, tt.args, tt.failing)
			if !errors.Is(got.err, tt.wantErr) || (tt.wantErr == nil) != (got.err == nil) {
				t.Errorf("Run: error %v, want %v", got.err, tt.wantErr)
			}
			printed, err := cli.ReadJSON(got.out)
			if err != nil {
				t.Fatalf("ReadJSON of the output: %v\n%s", err, got.out)
			}
			if diff := cmp.Diff(cli.NewReport(&r), printed); diff != "" {
				t.Errorf("JSON output (-want +got):\n%s", diff)
			}
		})
	}
}

// TestRunRefuses checks the command's refusals: arguments it doesn't take
// are usage errors before the cluster is reached, a refused exec is passed
// on, and output that is no report of this version is refused rather than
// misread.
func TestRunRefuses(t *testing.T) {
	t.Parallel()
	good := syntheticReport(t, false)
	tests := []struct {
		name    string
		args    []string
		out     string
		execErr error
		wantErr error
	}{
		{name: "nothing to check", args: nil, wantErr: cli.ErrUsage},
		{name: "something else to check", args: []string{"lots"}, wantErr: cli.ErrUsage},
		{name: "two things to check", args: []string{"balances", "balances"}, wantErr: cli.ErrUsage},
		{name: "an unknown flag", args: []string{"balances", "--yes"}, wantErr: cli.ErrUsage},
		{name: "exec refused", args: []string{"balances"}, execErr: errExecRefused, wantErr: errExecRefused},
		{name: "not JSON", args: []string{"balances"}, out: "core: usage", wantErr: cli.ErrBadReport},
		{name: "another version", args: []string{"balances"}, out: strings.Replace(string(coreJSON(t, &good)), `"version": 1`, `"version": 2`, 1), wantErr: cli.ErrBadReport},
		{name: "an unknown field", args: []string{"balances"}, out: strings.Replace(string(coreJSON(t, &good)), `"version": 1`, `"version": 1, "extra": true`, 1), wantErr: cli.ErrBadReport},
		{name: "counts that disagree", args: []string{"balances"}, out: strings.Replace(string(coreJSON(t, &good)), `"failed": 0`, `"failed": 1`, 1), wantErr: cli.ErrBadReport},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			core := &fakeCore{out: []byte(tt.out), err: tt.execErr}
			var out bytes.Buffer
			err := cli.Run(t.Context(), tt.args, &cli.Env{Out: &out, Connect: func(string, string) (cli.Core, error) { return core, nil }})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Run: error %v, want %v", err, tt.wantErr)
			}
			if out.Len() != 0 {
				t.Errorf("Run printed %q, want nothing", out.String())
			}
		})
	}
}

// TestWriteTableEscapes checks that text the report carries, such as an
// instrument's symbol a connector delivered, can't rewrite the owner's
// terminal: a control or bidirectional character makes it quoted.
func TestWriteTableEscapes(t *testing.T) {
	t.Parallel()
	r := cli.Report{Version: cli.ReportVersion, Tolerance: "0.01", Checked: 1, Checks: []cli.Check{{
		Account: "acc-1", Date: "2026-03-02", Kind: "holding", ISIN: "IETEST000026", Symbol: "ETF\x1b[2J\u202e",
		Reported: "1", Derived: "1", Difference: "0", Passed: true,
	}}}
	var out bytes.Buffer
	if err := cli.WriteTable(&out, &r); err != nil {
		t.Fatalf("WriteTable: %v", err)
	}
	if strings.ContainsAny(out.String(), "\x1b\u202e") || !strings.Contains(out.String(), strconv.Quote("ETF\x1b[2J\u202e IETEST000026")) {
		t.Errorf("WriteTable wrote %q, want the balance quoted", out.String())
	}
}
