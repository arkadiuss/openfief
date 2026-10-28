package app_test

import (
	"errors"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
)

// importFixture holds synthetic transactions in acc-1 that the import tests
// combine.
type importFixture struct {
	deposit, buy, buyCorrected, buyReordered, elsewhere, undated ledger.Transaction
}

func newImportFixture(t *testing.T) importFixture {
	t.Helper()
	a := instrument(t, isinA, "TSTA")
	f := importFixture{
		deposit:      must(t)(ledger.NewDeposit(header(t, acc1, "d1", "2026-03-02"), money(t, "1000", "USD"))),
		buy:          must(t)(ledger.NewBuy(header(t, acc1, "b1", "2026-03-03"), a, qty(t, "10"), money(t, "500", "USD"), money(t, "1", "USD"))),
		buyCorrected: must(t)(ledger.NewBuy(header(t, acc1, "b1", "2026-03-03"), a, qty(t, "12"), money(t, "600", "USD"), money(t, "1", "USD"))),
		elsewhere:    must(t)(ledger.NewDeposit(header(t, "acc-9", "d9", "2026-03-02"), money(t, "5", "USD"))),
	}
	f.buyReordered = f.buy
	f.buyReordered.Postings = slices.Clone(f.buy.Postings)
	slices.Reverse(f.buyReordered.Postings)
	f.buyReordered.Postings[0].Amount = money(t, "-1.00", "USD")
	f.undated = f.deposit
	f.undated.Key.ExternalID = "d2"
	f.undated.Date = kernel.Date{}
	return f
}

func versioned(t *ledger.Transaction, version int) ledger.VersionedTransaction {
	return ledger.VersionedTransaction{Transaction: *t, Version: version}
}

// checkSummary fails t unless got is want.
func checkSummary(t *testing.T, want, got *app.ImportSummary) {
	t.Helper()
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("ImportSummary (-want +got):\n%s", diff)
	}
}

// checkLog fails t unless store's log of acc-1 is want.
func checkLog(t *testing.T, store *memStore, want []ledger.VersionedTransaction) {
	t.Helper()
	log, err := store.Log(t.Context(), acc1)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if diff := cmp.Diff(want, log, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("stored log (-want +got):\n%s", diff)
	}
}

func TestImportAppendsVersions(t *testing.T) {
	t.Parallel()
	f := newImportFixture(t)
	tests := []struct {
		name    string
		imports [][]ledger.Transaction
		want    app.ImportSummary // of the last import
		wantErr error
		wantLog []ledger.VersionedTransaction
	}{
		{
			name:    "new keys become version 1",
			imports: [][]ledger.Transaction{{f.deposit, f.buy}},
			want:    app.ImportSummary{New: 2},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1), versioned(&f.deposit, 1)},
		},
		{
			name:    "re-import changes nothing",
			imports: [][]ledger.Transaction{{f.deposit, f.buy}, {f.deposit, f.buy}},
			want:    app.ImportSummary{Unchanged: 2},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1), versioned(&f.deposit, 1)},
		},
		{
			name:    "posting order and number form don't matter",
			imports: [][]ledger.Transaction{{f.buy}, {f.buyReordered}},
			want:    app.ImportSummary{Unchanged: 1},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1)},
		},
		{
			name:    "a changed row supersedes and the old version stays",
			imports: [][]ledger.Transaction{{f.buy}, {f.buyCorrected}},
			want:    app.ImportSummary{NewVersions: 1},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1), versioned(&f.buyCorrected, 2)},
		},
		{
			name:    "changing a row back is another version",
			imports: [][]ledger.Transaction{{f.buy}, {f.buyCorrected}, {f.buy}},
			want:    app.ImportSummary{NewVersions: 1},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1), versioned(&f.buyCorrected, 2), versioned(&f.buy, 3)},
		},
		{
			name:    "one key twice in one import",
			imports: [][]ledger.Transaction{{f.buy, f.buyCorrected}},
			want:    app.ImportSummary{New: 1, NewVersions: 1},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1), versioned(&f.buyCorrected, 2)},
		},
		{
			name:    "an unknown account writes nothing",
			imports: [][]ledger.Transaction{{f.deposit, f.elsewhere}},
			wantErr: app.ErrUnknownAccount,
		},
		{
			name:    "an invalid transaction writes nothing",
			imports: [][]ledger.Transaction{{f.deposit, f.undated}},
			wantErr: ledger.ErrInvalidTransaction,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, store := seeded(t)
			var got app.ImportSummary
			var err error
			for _, txs := range tt.imports {
				got, err = l.Import(t.Context(), txs, nil)
			}
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Import error (-want +got):\n%s", diff)
			}
			checkSummary(t, &tt.want, &got)
			checkLog(t, store, tt.wantLog)
		})
	}
}

// relabeled returns t with description d, so it differs from t.
func relabeled(t *ledger.Transaction, d string) *ledger.Transaction {
	c := *t
	c.Description = d
	return &c
}

func TestImportRetriesVersionConflicts(t *testing.T) {
	t.Parallel()
	f := newImportFixture(t)
	tests := []struct {
		name string
		// race appends for "another import" just before this one appends.
		race    func(s *memStore, v *ledger.VersionedTransaction) error
		want    app.ImportSummary
		wantErr *app.VersionConflictError
		wantLog []ledger.VersionedTransaction
	}{
		{
			name: "the other import appended the same content",
			race: func(s *memStore, v *ledger.VersionedTransaction) error {
				s.beforeAppend = nil
				return s.appendVersion(v)
			},
			want:    app.ImportSummary{Unchanged: 1},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buy, 1)},
		},
		{
			name: "the other import appended other content",
			race: func(s *memStore, _ *ledger.VersionedTransaction) error {
				s.beforeAppend = nil
				other := versioned(&f.buyCorrected, 1)
				return s.appendVersion(&other)
			},
			want:    app.ImportSummary{NewVersions: 1},
			wantLog: []ledger.VersionedTransaction{versioned(&f.buyCorrected, 1), versioned(&f.buy, 2)},
		},
		{
			name: "every attempt loses",
			race: func(s *memStore, v *ledger.VersionedTransaction) error {
				other := versioned(relabeled(&f.buy, "other "+strconv.Itoa(v.Version)), v.Version)
				return s.appendVersion(&other)
			},
			wantErr: &app.VersionConflictError{Key: f.buy.Key, Attempts: app.MaxAppendAttempts},
			wantLog: []ledger.VersionedTransaction{
				versioned(relabeled(&f.buy, "other 1"), 1),
				versioned(relabeled(&f.buy, "other 2"), 2),
				versioned(relabeled(&f.buy, "other 3"), 3),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, store := seeded(t)
			store.beforeAppend = tt.race
			got, err := l.Import(t.Context(), []ledger.Transaction{f.buy}, nil)
			checkConflict(t, tt.wantErr, err)
			checkSummary(t, &tt.want, &got)
			checkLog(t, store, tt.wantLog)
		})
	}
}

// checkConflict fails t unless err is want, or nil when want is nil.
func checkConflict(t *testing.T, want *app.VersionConflictError, err error) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("Import: %v", err)
		}
		return
	}
	var conflict *app.VersionConflictError
	if !errors.As(err, &conflict) || !errors.Is(err, app.ErrVersionConflict) {
		t.Fatalf("Import error = %v, want a *VersionConflictError", err)
	}
	if diff := cmp.Diff(want, conflict); diff != "" {
		t.Errorf("VersionConflictError (-want +got):\n%s", diff)
	}
}

func TestConcurrentImportsAppendOneVersion(t *testing.T) {
	t.Parallel()
	const imports = 8
	f := newImportFixture(t)
	l, store := seeded(t)
	if _, err := l.Import(t.Context(), []ledger.Transaction{f.buy}, nil); err != nil {
		t.Fatalf("Import: %v", err)
	}
	sums := make([]app.ImportSummary, imports)
	errs := make([]error, imports)
	var wg sync.WaitGroup
	for i := range imports {
		wg.Go(func() {
			sums[i], errs[i] = l.Import(t.Context(), []ledger.Transaction{f.buyCorrected}, nil)
		})
	}
	wg.Wait()
	var total app.ImportSummary
	for i := range imports {
		if errs[i] != nil {
			t.Fatalf("Import %d: %v", i, errs[i])
		}
		total.Unchanged += sums[i].Unchanged
		total.NewVersions += sums[i].NewVersions
	}
	checkSummary(t, &app.ImportSummary{Unchanged: imports - 1, NewVersions: 1}, &total)
	checkLog(t, store, []ledger.VersionedTransaction{versioned(&f.buy, 1), versioned(&f.buyCorrected, 2)})
}
