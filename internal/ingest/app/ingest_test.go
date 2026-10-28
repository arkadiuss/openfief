package app_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

var (
	errMapper  = errors.New("mapper failed")
	errLedger  = errors.New("ledger failed")
	errFetcher = errors.New("fetch failed")
)

// fakeMapper returns its mapping or error, and records what it read.
type fakeMapper struct {
	mapping ingest.Mapping
	err     error
	read    string
	mode    ingest.Mode
}

func (f *fakeMapper) Map(src io.Reader, mode ingest.Mode) (ingest.Mapping, error) {
	b, err := io.ReadAll(src)
	if err != nil {
		return ingest.Mapping{}, fmt.Errorf("fake mapper: %w", err)
	}
	f.read, f.mode = string(b), mode
	return f.mapping, f.err
}

// fakeLedger records the transactions and assertions it imports.
type fakeLedger struct {
	summary    ingest.ImportSummary
	err        error
	calls      int
	txs        []ledger.Transaction
	assertions []ledger.BalanceAssertion
}

func (f *fakeLedger) Import(_ context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	f.calls++
	f.txs, f.assertions = txs, assertions
	return f.summary, f.err
}

// fakeFetcher returns a fixed statement or error.
type fakeFetcher struct {
	statement string
	err       error
}

func (f fakeFetcher) FetchStatement(context.Context) ([]byte, error) {
	return []byte(f.statement), f.err
}

func march2(t *testing.T) kernel.Date {
	t.Helper()
	d, err := kernel.NewDate(2026, time.March, 2)
	if err != nil {
		t.Fatalf("NewDate: %v", err)
	}
	return d
}

// mapping returns a mapping with one deposit, one cash assertion and one
// problem.
func mapping(t *testing.T) ingest.Mapping {
	t.Helper()
	day := march2(t)
	usd, err := kernel.ParseMoney("100", mustCurrency(t, "USD"))
	if err != nil {
		t.Fatalf("ParseMoney: %v", err)
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "test", ExternalID: "1"}, Account: "acc-1", Date: day}
	dep, err := ledger.NewDeposit(h, usd)
	if err != nil {
		t.Fatalf("NewDeposit: %v", err)
	}
	return ingest.Mapping{
		Transactions: []ledger.Transaction{dep},
		Assertions:   []ledger.BalanceAssertion{ledger.CashAssertion("acc-1", day, usd)},
		Problems:     []ingest.Problem{ingest.Unsupported("row 2", "broker interest")},
	}
}

func mustCurrency(t *testing.T, code string) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	return c
}

// service returns a Service with mapper for "ibkr" and the ledger.
func service(t *testing.T, mapper app.Mapper, l app.LedgerImporter) *app.Service {
	t.Helper()
	day := march2(t)
	return &app.Service{
		Ledger:     l,
		Mappers:    map[ingest.InstitutionID]app.Mapper{"ibkr": mapper},
		Today:      func() kernel.Date { return day },
		NewBatchID: func() ingest.BatchID { return "batch-1" },
	}
}

func TestIngestFile(t *testing.T) {
	t.Parallel()
	m := mapping(t)
	mapper := &fakeMapper{mapping: m}
	l := &fakeLedger{summary: ingest.ImportSummary{New: 1}}
	got, err := service(t, mapper, l).IngestFile(context.Background(), "ibkr", ingest.ModeTransactions, strings.NewReader("statement"))
	if err != nil {
		t.Fatalf("IngestFile: %v", err)
	}
	want := ingest.Batch{
		ID: "batch-1", Institution: "ibkr", Source: app.SourceFile, Mode: ingest.ModeTransactions, Started: march2(t),
		Transactions: 1, Assertions: 1, Imported: ingest.ImportSummary{New: 1}, Problems: m.Problems,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("batch (-want +got):\n%s", diff)
	}
	if mapper.read != "statement" || mapper.mode != ingest.ModeTransactions {
		t.Errorf("mapper read %q in %s, want the statement in transactions mode", mapper.read, mapper.mode)
	}
	if diff := cmp.Diff(m.Transactions, l.txs); diff != "" {
		t.Errorf("imported transactions (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(m.Assertions, l.assertions); diff != "" {
		t.Errorf("imported assertions (-want +got):\n%s", diff)
	}
}

func TestIngestFileFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		institution ingest.InstitutionID
		mode        ingest.Mode
		mapErr      error
		ledgerErr   error
		wantErr     error
		wantImports int
	}{
		{name: "unknown institution", institution: "xtb", mode: ingest.ModeTransactions, wantErr: app.ErrUnknownInstitution},
		{name: "invalid mode", institution: "ibkr", wantErr: ingest.ErrInvalidBatch},
		{name: "unreadable statement imports nothing", institution: "ibkr", mode: ingest.ModeOpening, mapErr: errMapper, wantErr: errMapper},
		{name: "ledger failure", institution: "ibkr", mode: ingest.ModeOpening, ledgerErr: errLedger, wantErr: errLedger, wantImports: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &fakeLedger{err: tt.ledgerErr}
			svc := service(t, &fakeMapper{mapping: mapping(t), err: tt.mapErr}, l)
			_, err := svc.IngestFile(context.Background(), tt.institution, tt.mode, strings.NewReader("x"))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("IngestFile error (-want +got):\n%s", diff)
			}
			if l.calls != tt.wantImports {
				t.Errorf("ledger imports = %d, want %d", l.calls, tt.wantImports)
			}
		})
	}
}

func TestPull(t *testing.T) {
	t.Parallel()
	mapper := &fakeMapper{mapping: mapping(t)}
	l := &fakeLedger{summary: ingest.ImportSummary{Unchanged: 1}}
	got, err := service(t, mapper, l).Pull(context.Background(), "ibkr", ingest.ModeTransactions, fakeFetcher{statement: "pulled"})
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if got.Source != app.SourcePull || got.Imported.Unchanged != 1 || mapper.read != "pulled" {
		t.Errorf("Pull = source %q, imported %+v, mapper read %q; want a pull batch of the fetched statement", got.Source, got.Imported, mapper.read)
	}
}

func TestPullFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		institution ingest.InstitutionID
		fetchErr    error
		wantErr     error
	}{
		{name: "unknown institution", institution: "xtb", wantErr: app.ErrUnknownInstitution},
		{name: "fetch failure imports nothing", institution: "ibkr", fetchErr: errFetcher, wantErr: errFetcher},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &fakeLedger{}
			svc := service(t, &fakeMapper{mapping: mapping(t)}, l)
			_, err := svc.Pull(context.Background(), tt.institution, ingest.ModeTransactions, fakeFetcher{err: tt.fetchErr})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Pull error (-want +got):\n%s", diff)
			}
			if l.calls != 0 {
				t.Errorf("ledger imports = %d, want 0", l.calls)
			}
		})
	}
}
