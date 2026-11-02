package postgres_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/postgres"
)

func TestUnsupportedBatches(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	s := postgres.New(pool)
	ctx := t.Context()
	if got, err := s.UnsupportedBatches(ctx); err != nil || len(got) != 0 {
		t.Fatalf("UnsupportedBatches on an empty history = %v, %v", got, err)
	}
	upload := &ingest.Batch{
		ID: "b-7", Institution: "ibkr", Source: app.SourceUpload, Mode: ingest.ModeTransactions,
		Problems: []ingest.Problem{
			ingest.Skipped("row 3", "summary row"), ingest.Unsupported("row 4", "option trade"), ingest.Unsupported("row 5", "interest"),
		},
	}
	clean := &ingest.Batch{
		ID: "b-8", Institution: "ibkr", Source: app.SourceUpload, Mode: ingest.ModeTransactions,
		Problems: []ingest.Problem{ingest.Skipped("row 1", "summary row")},
	}
	for _, b := range []*ingest.Batch{upload, clean} {
		if err := s.RecordBatch(ctx, "flex-main", b); err != nil {
			t.Fatalf("RecordBatch %s: %v", b.ID, err)
		}
	}
	claim(t, s, "xtb-main", t0())
	if err := s.StartBatch(ctx, "xtb-main", app.OpenBatch{ID: "b-1", Mode: ingest.ModeTransactions}, t0()); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}
	synced := &ingest.Batch{
		ID: "b-1", Institution: "xtb", Source: app.SourceSync, Mode: ingest.ModeTransactions,
		Problems: []ingest.Problem{ingest.Unsupported("row 9", "dividend adjustment")},
	}
	if err := s.CompleteSync(ctx, "xtb-main", synced, "w-1", t0().Add(time.Hour), t0().Add(time.Minute), time.Time{}); err != nil {
		t.Fatalf("CompleteSync: %v", err)
	}
	got, err := s.UnsupportedBatches(ctx)
	if err != nil {
		t.Fatalf("UnsupportedBatches: %v", err)
	}
	want := []ingest.BatchProblems{{Batch: "b-7", Instance: "flex-main", Unsupported: 2}, {Batch: "b-1", Instance: "xtb-main", Unsupported: 1}}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("UnsupportedBatches (-want +got):\n%s", diff)
	}
}
