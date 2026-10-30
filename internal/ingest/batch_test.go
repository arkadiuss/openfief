package ingest_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func day(t *testing.T) kernel.Date {
	t.Helper()
	d, err := kernel.NewDate(2026, time.March, 2)
	if err != nil {
		t.Fatalf("NewDate: %v", err)
	}
	return d
}

func TestInstitutionValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      ingest.Institution
		wantErr error
	}{
		{name: "valid", in: ingest.Institution{ID: "ibkr", Name: "Interactive Brokers"}},
		{name: "no ID", in: ingest.Institution{Name: "Interactive Brokers"}, wantErr: ingest.ErrInvalidInstitution},
		{name: "no name", in: ingest.Institution{ID: "ibkr"}, wantErr: ingest.ErrInvalidInstitution},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.in.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewBatch(t *testing.T) {
	t.Parallel()
	started := day(t)
	tests := []struct {
		name        string
		id          ingest.BatchID
		institution ingest.InstitutionID
		source      string
		mode        ingest.Mode
		started     kernel.Date
		wantErr     error
	}{
		{name: "transactions batch", id: "b-1", institution: "ibkr", source: "upload", mode: ingest.ModeTransactions, started: started},
		{name: "opening batch", id: "b-1", institution: "ibkr", source: "upload", mode: ingest.ModeOpening, started: started},
		{name: "no ID", institution: "ibkr", source: "upload", mode: ingest.ModeOpening, started: started, wantErr: ingest.ErrInvalidBatch},
		{name: "no institution", id: "b-1", source: "upload", mode: ingest.ModeOpening, started: started, wantErr: ingest.ErrInvalidBatch},
		{name: "no source", id: "b-1", institution: "ibkr", mode: ingest.ModeOpening, started: started, wantErr: ingest.ErrInvalidBatch},
		{name: "no mode", id: "b-1", institution: "ibkr", source: "upload", started: started, wantErr: ingest.ErrInvalidBatch},
		{name: "unknown mode", id: "b-1", institution: "ibkr", source: "upload", mode: 9, started: started, wantErr: ingest.ErrInvalidBatch},
		{name: "no start date", id: "b-1", institution: "ibkr", source: "upload", mode: ingest.ModeOpening, wantErr: ingest.ErrInvalidBatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ingest.NewBatch(tt.id, tt.institution, tt.source, tt.mode, tt.started)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("NewBatch error (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			want := ingest.Batch{ID: tt.id, Institution: tt.institution, Source: tt.source, Mode: tt.mode, Started: tt.started}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("NewBatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBatchComplete(t *testing.T) {
	t.Parallel()
	b, err := ingest.NewBatch("b-1", "ibkr", "upload", ingest.ModeTransactions, day(t))
	if err != nil {
		t.Fatalf("NewBatch: %v", err)
	}
	problems := []ingest.Problem{ingest.Unsupported("Trades#1", "option trade")}
	m := ingest.Mapping{
		Transactions: make([]ledger.Transaction, 3),
		Assertions:   make([]ledger.BalanceAssertion, 2),
		Problems:     problems,
	}
	summary := ingest.ImportSummary{New: 2, Unchanged: 1, FailedAssertions: 1}
	want := b
	got := b
	got.Complete(&m, summary)
	want.Transactions, want.Assertions, want.Imported = 3, 2, summary
	want.Problems = []ingest.Problem{ingest.Unsupported("Trades#1", "option trade")}
	want.Accounts = []ledger.AccountID{""}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Complete (-want +got):\n%s", diff)
	}
	problems[0].Reason = "changed after Complete"
	if got.Problems[0].Reason != "option trade" {
		t.Errorf("Complete shares its problems with the mapping")
	}
}

func TestEnumStrings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		got, want string
	}{
		{ingest.ModeOpening.String(), "opening"},
		{ingest.ModeTransactions.String(), "transactions"},
		{ingest.Mode(0).String(), "Mode(0)"},
		{ingest.ProblemUnsupported.String(), "unsupported"},
		{ingest.ProblemSkipped.String(), "skipped"},
		{ingest.ProblemKind(7).String(), "ProblemKind(7)"},
		{ingest.Skipped("Trades#2", "summary row").String(), "skipped Trades#2: summary row"},
		{ingest.Unsupported("CashTransactions#1", "broker interest").String(), "unsupported CashTransactions#1: broker interest"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}
