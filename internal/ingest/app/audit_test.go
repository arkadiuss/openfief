package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// auditCall is one entry a use case asked for.
type auditCall struct {
	Action string
	Refs   []string
}

// recordingAuditor enters the changes it is asked to, as the core does, and
// runs each change after it; with err set it refuses every entry, as an audit
// log that is down does, and calls no change.
type recordingAuditor struct {
	calls []auditCall
	err   error
}

func (r *recordingAuditor) Audited(ctx context.Context, _, _, action string, refs []string, change func(context.Context) error) error {
	if r.err != nil {
		return r.err
	}
	r.calls = append(r.calls, auditCall{Action: action, Refs: refs})
	return change(ctx)
}

// confirmingStore is a TickerStore that settles every confirmation and
// correction with nothing to book, and counts them.
type confirmingStore struct {
	app.TickerStore

	confirmed, corrected int
}

func (*confirmingStore) TickerAccounts(context.Context, string, string, bool) ([]ledger.AccountID, error) {
	return nil, nil
}

func (s *confirmingStore) ConfirmTicker(context.Context, ingest.TickerConfirmation) (ingest.TickerBooking, error) {
	s.confirmed++
	return ingest.TickerBooking{}, nil
}

func (s *confirmingStore) CorrectTicker(context.Context, ingest.TickerConfirmation) (ingest.TickerBooking, error) {
	s.corrected++
	return ingest.TickerBooking{}, nil
}

// ingestChange is an owner's change that makes the change once and returns
// how many times the change itself was made, and the error.
type ingestChange struct {
	name string
	run  func(t *testing.T, a *recordingAuditor) (made int, err error)
	want auditCall
}

const changeOrigin = "https://ui.example.com"

// confirmation is the owner's confirmation of a synthetic ticker.
func confirmation(t *testing.T) ingest.TickerConfirmation {
	t.Helper()
	isin, err := ledger.ParseISIN("PLTEST000010")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	return ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TSTB", ISIN: isin}
}

func confirmTicker(t *testing.T, a *recordingAuditor) (int, error) {
	t.Helper()
	store := &confirmingStore{}
	tk := &app.Tickers{Owners: fakeOwners{}, Store: store, Audit: a}
	if _, err := tk.Confirm(t.Context(), changeOrigin, "owner", confirmation(t)); err != nil {
		return store.confirmed, fmt.Errorf("confirming: %w", err)
	}
	return store.confirmed, nil
}

func correctTicker(t *testing.T, a *recordingAuditor) (int, error) {
	t.Helper()
	store := &confirmingStore{}
	tk := &app.Tickers{Owners: fakeOwners{}, Store: store, Audit: a}
	if _, err := tk.Correct(t.Context(), changeOrigin, "owner", confirmation(t)); err != nil {
		return store.corrected, fmt.Errorf("correcting: %w", err)
	}
	return store.corrected, nil
}

func retrySync(t *testing.T, a *recordingAuditor) (int, error) {
	t.Helper()
	tr := &recordingTrigger{}
	r := &app.Retries{Owners: fakeOwners{}, Bindings: &fakeBindings{}, Trigger: tr, Now: t0, Audit: a}
	err := r.RetrySync(t.Context(), changeOrigin, "owner", "xtb-main")
	return len(tr.calls), err
}

// TestIngestOwnerChangesAreAudited checks the owner's changes in the ingest
// context that the core's own tests can't drive without a connector or a
// running scheduler (D-211): the confirmation and the correction of a ticker
// and the retry of a sync each enter exactly one entry, naming the record by
// ID, before the change is made, and a log that is down means the change is
// not made.
func TestIngestOwnerChangesAreAudited(t *testing.T) {
	t.Parallel()
	tests := []ingestChange{
		{"a ticker confirmed", confirmTicker, auditCall{Action: "ticker_confirmed", Refs: []string{"instrument:PLTEST000010"}}},
		{"a ticker corrected", correctTicker, auditCall{Action: "ticker_corrected", Refs: []string{"instrument:PLTEST000010"}}},
		{"a sync retried", retrySync, auditCall{Action: "sync_retried", Refs: []string{"plugin:xtb-main"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := &recordingAuditor{}
			made, err := tt.run(t, a)
			if err != nil || made != 1 {
				t.Fatalf("the change = %d times, %v; want once and no error", made, err)
			}
			if diff := cmp.Diff([]auditCall{tt.want}, a.calls); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
			down := &recordingAuditor{err: kernel.ErrAuditUnavailable}
			made, err = tt.run(t, down)
			if !errors.Is(err, kernel.ErrAuditUnavailable) || made != 0 {
				t.Errorf("with the log down: the change = %d times, %v; want no change and ErrAuditUnavailable", made, err)
			}
		})
	}
}

// TestRetryOfAnInstanceNoPluginCanHaveEntersNothing checks that a retry naming
// something no plugin instance can be called is refused as no instance before
// any entry, so the log never takes a name that isn't one (GR-12).
func TestRetryOfAnInstanceNoPluginCanHaveEntersNothing(t *testing.T) {
	t.Parallel()
	a := &recordingAuditor{}
	tr := &recordingTrigger{}
	r := &app.Retries{Owners: fakeOwners{}, Bindings: &fakeBindings{}, Trigger: tr, Now: t0, Audit: a}
	err := r.RetrySync(t.Context(), "https://ui.example.com", "owner", "Not A Plugin 12 3456")
	if !errors.Is(err, app.ErrNoInstance) || len(a.calls) != 0 || len(tr.calls) != 0 {
		t.Errorf("RetrySync = %v with %d entries and %d syncs, want ErrNoInstance and neither", err, len(a.calls), len(tr.calls))
	}
}
