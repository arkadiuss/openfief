// Package cli is the views context's part of the owner's command line:
// `openfief check balances`, the owner's balance check (D-200, criterion
// 1). It runs `core check balances` in the core's pod through Kubernetes
// exec with the owner's own kubeconfig, as `openfief enroll` runs the
// core's enrollment command (D-80), so only someone who may exec into the
// core's pod gets it: never a viewer, a script token or a view plugin. The
// core writes the report as JSON to the exec's standard output, which
// reaches only the owner's terminal; this package reads it back and prints
// it as a table, or as JSON with --json. The command fails when any
// balance assertion fails, so a script can act on its exit status.
// cmd/core writes the report with WriteJSON; cmd/openfief wires the
// Kubernetes adapter, internal/identity/kube's CoreAccess.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// Usage is the command's synopsis.
const Usage = `usage:
  openfief check balances [--all] [--json] [--kubeconfig PATH] [--context NAME]`

// command is the one thing `openfief check` checks.
const command = "balances"

// tablePadding is the space between the table's columns.
const tablePadding = 2

// ReportVersion is the version of the report's JSON shape, which the core
// writes and the command reads: a command refuses a report of another
// version rather than misread it.
const ReportVersion = 1

// Errors of the command.
var (
	// ErrUsage reports arguments the command does not accept.
	ErrUsage = errors.New("bad arguments")
	// ErrBalancesFailed reports a report with at least one failed balance
	// assertion: the command's output is complete, and its exit status
	// says that something failed.
	ErrBalancesFailed = errors.New("balance assertions failed")
	// ErrBadReport reports output of the core that is no report this
	// command reads.
	ErrBadReport = errors.New("unexpected report from the core")
)

// CoreArgs returns the arguments of the core's command that writes the
// report: `core check balances`, with --all for every balance assertion.
func CoreArgs(all bool) []string {
	if all {
		return []string{"check", command, "--all"}
	}
	return []string{"check", command}
}

// Core is the core's pod, reached with the owner's kubeconfig.
type Core interface {
	// Exec runs the core's binary with args in the core's pod and returns
	// what it wrote to standard output.
	Exec(ctx context.Context, args []string) ([]byte, error)
}

// Env is what the command needs from outside it.
type Env struct {
	Out io.Writer
	// Connect reaches the core with a kubeconfig path and context, either
	// empty for the defaults.
	Connect func(kubeconfig, context string) (Core, error)
}

// options are the parsed flags.
type options struct {
	kubeconfig, context string
	all, json           bool
}

// parse parses the arguments after "check": "balances" and its flags, in
// any order.
func parse(args []string) (options, error) {
	fs := flag.NewFlagSet("openfief check balances", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var o options
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "kubeconfig file")
	fs.StringVar(&o.context, "context", "", "kubeconfig context")
	fs.BoolVar(&o.all, "all", false, "every balance assertion, not only the latest of each position and cash balance")
	fs.BoolVar(&o.json, "json", false, "write the report as JSON")
	var positional []string
	for rest := args; ; {
		if err := fs.Parse(rest); err != nil {
			return options{}, fmt.Errorf("%w: %w\n%s", ErrUsage, err, Usage)
		}
		if fs.NArg() == 0 {
			break
		}
		positional = append(positional, fs.Arg(0))
		rest = fs.Args()[1:]
	}
	if len(positional) != 1 || positional[0] != command {
		return options{}, fmt.Errorf("%w: openfief check takes balances\n%s", ErrUsage, Usage)
	}
	return o, nil
}

// Run runs `openfief check` with args, the arguments after "check": it has
// the core write the report, prints it, and returns an error wrapping
// ErrBalancesFailed when any balance assertion failed.
func Run(ctx context.Context, args []string, env *Env) error {
	o, err := parse(args)
	if err != nil {
		return err
	}
	core, err := env.Connect(o.kubeconfig, o.context)
	if err != nil {
		return fmt.Errorf("connecting to the cluster: %w", err)
	}
	out, err := core.Exec(ctx, CoreArgs(o.all))
	if err != nil {
		return fmt.Errorf("running the check in the core: %w", err)
	}
	r, err := ReadJSON(out)
	if err != nil {
		return err
	}
	if o.json {
		err = writeReport(env.Out, &r)
	} else {
		err = WriteTable(env.Out, &r)
	}
	if err != nil {
		return err
	}
	if r.Failed > 0 {
		return fmt.Errorf("%w: %d of %d", ErrBalancesFailed, r.Failed, r.Checked)
	}
	return nil
}

// Report is the report's JSON shape, version ReportVersion: what the core
// writes, and what --json prints. Amounts and units are decimal strings,
// never JSON numbers, which most readers hold as binary floats.
type Report struct {
	Version int `json:"version"`
	// Tolerance is views.CashTolerance, the cash tolerance in each
	// assertion's own currency; holdings must match exactly.
	Tolerance   string            `json:"tolerance"`
	Checked     int               `json:"checked"`
	Failed      int               `json:"failed"`
	Checks      []Check           `json:"checks"`
	Unsupported []UnsupportedRows `json:"unsupported"`
}

// Check is one balance assertion and its check.
type Check struct {
	Account string `json:"account"`
	Date    string `json:"date"`
	// Kind is "cash", with Currency, or "holding", with ISIN and Symbol.
	Kind     string `json:"kind"`
	Currency string `json:"currency,omitempty"`
	ISIN     string `json:"isin,omitempty"`
	Symbol   string `json:"symbol,omitempty"`
	// Entered marks a cash checkpoint the owner entered (D-158).
	Entered    bool   `json:"entered"`
	Reported   string `json:"reported"`
	Derived    string `json:"derived"`
	Difference string `json:"difference"`
	Passed     bool   `json:"passed"`
}

// UnsupportedRows is what an instance's latest import batch left out as
// unsupported (views.UnsupportedRows).
type UnsupportedRows struct {
	Instance    string   `json:"instance"`
	Institution string   `json:"institution"`
	Batch       string   `json:"batch"`
	Finished    string   `json:"finished"`
	Accounts    []string `json:"accounts"`
	Rows        int      `json:"rows"`
}

// NewReport returns r in the report's JSON shape.
func NewReport(r *views.BalanceReport) Report {
	out := Report{
		Version: ReportVersion, Tolerance: views.CashTolerance().String(), Checked: len(r.Checks), Failed: r.Failed(),
		Checks: make([]Check, 0, len(r.Checks)), Unsupported: make([]UnsupportedRows, 0, len(r.Unsupported)),
	}
	for i := range r.Checks {
		c := &r.Checks[i]
		a := &c.Assertion
		check := Check{
			Account: string(a.Account), Date: a.Date.String(), Kind: a.Book.Kind.String(), Entered: a.Entered,
			Reported: a.Balance.String(), Derived: c.Derived.String(), Difference: c.Difference.String(), Passed: c.Passed,
		}
		if a.Book.Kind == ledger.BookHolding {
			check.ISIN, check.Symbol = a.Book.Instrument.ISIN.String(), a.Book.Instrument.Symbol
		} else {
			check.Currency = a.Book.Currency.String()
		}
		out.Checks = append(out.Checks, check)
	}
	for i := range r.Unsupported {
		u := &r.Unsupported[i]
		accounts := make([]string, 0, len(u.Accounts))
		for _, a := range u.Accounts {
			accounts = append(accounts, string(a))
		}
		out.Unsupported = append(out.Unsupported, UnsupportedRows{
			Instance: u.Instance, Institution: string(u.Institution), Batch: string(u.Batch),
			Finished: u.Finished.UTC().Format(time.RFC3339), Accounts: accounts, Rows: u.Rows,
		})
	}
	return out
}

// WriteJSON writes r to w in the report's JSON shape: what `core check
// balances` writes to its standard output.
func WriteJSON(w io.Writer, r *views.BalanceReport) error {
	rep := NewReport(r)
	return writeReport(w, &rep)
}

// writeReport writes r as indented JSON.
func writeReport(w io.Writer, r *Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("writing the report: %w", err)
	}
	return nil
}

// ReadJSON reads a report the core wrote, refusing unknown fields, another
// version and counts that disagree with the checks.
func ReadJSON(b []byte) (Report, error) {
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	var r Report
	if err := dec.Decode(&r); err != nil {
		return Report{}, fmt.Errorf("%w: %w", ErrBadReport, err)
	}
	if r.Version != ReportVersion {
		return Report{}, fmt.Errorf("%w: version %d, this command reads %d; use the CLI of the core's release",
			ErrBadReport, r.Version, ReportVersion)
	}
	failed := 0
	for i := range r.Checks {
		if !r.Checks[i].Passed {
			failed++
		}
	}
	if r.Checked != len(r.Checks) || r.Failed != failed {
		return Report{}, fmt.Errorf("%w: its counts disagree with its checks", ErrBadReport)
	}
	return r, nil
}

// WriteTable writes r as the command's table: each check with its result,
// the instances' unsupported rows, flagged and not checked, and a summary.
func WriteTable(w io.Writer, r *Report) error {
	var b strings.Builder
	b.WriteString("ACCOUNT\tDATE\tBALANCE\tREPORTED\tDERIVED\tDIFFERENCE\tRESULT\n")
	for i := range r.Checks {
		c := &r.Checks[i]
		balance := "cash " + c.Currency
		if c.Kind == ledger.BookHolding.String() {
			balance = c.Symbol + " " + c.ISIN
		}
		if c.Entered {
			balance += " (entered)"
		}
		result := "pass"
		if !c.Passed {
			result = "FAIL"
		}
		b.WriteString(row(c.Account, c.Date, balance, c.Reported, c.Derived, c.Difference) + "\t" + result + "\n")
	}
	if len(r.Unsupported) > 0 {
		b.WriteString("\nFLAGGED, UNSUPPORTED: not checked (D-24); open XTB CFDs are among them\n")
		b.WriteString("INSTANCE\tINSTITUTION\tLATEST BATCH\tFINISHED\tACCOUNTS\tROWS\n")
		for i := range r.Unsupported {
			u := &r.Unsupported[i]
			b.WriteString(row(u.Instance, u.Institution, u.Batch, u.Finished, strings.Join(u.Accounts, ",")) +
				"\t" + strconv.Itoa(u.Rows) + "\n")
		}
	}
	b.WriteString("\n" + strconv.Itoa(r.Checked) + " balance assertions checked, " + strconv.Itoa(r.Failed) +
		" failed: cash passes within " + shown(r.Tolerance) + " of its currency, a holding with equal units.\n")
	tw := tabwriter.NewWriter(w, 0, 0, tablePadding, ' ', 0)
	if _, err := io.WriteString(tw, b.String()); err != nil {
		return fmt.Errorf("writing the report: %w", err)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("writing the report: %w", err)
	}
	return nil
}

// row joins cells with tabs, each as shown.
func row(cells ...string) string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = shown(c)
	}
	return strings.Join(out, "\t")
}

// shown returns s as it is when every character prints as itself, and
// Go-quoted otherwise, so no text in the report can rewrite the owner's
// terminal.
func shown(s string) string {
	for _, r := range s {
		if !unicode.IsPrint(r) || unicode.Is(unicode.Bidi_Control, r) {
			return strconv.Quote(s)
		}
	}
	return s
}
