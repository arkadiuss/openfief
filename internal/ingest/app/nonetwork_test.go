package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// inChangeKey marks a context as the one a change's transaction runs with.
type inChangeKey struct{}

// inChange reports whether ctx is a change's.
func inChange(ctx context.Context) bool {
	marked, ok := ctx.Value(inChangeKey{}).(bool)
	return ok && marked
}

// transactionAuditor runs each change with a context marked as its
// transaction's, as the core's outbox does.
type transactionAuditor struct{ changes int }

func (a *transactionAuditor) Audited(ctx context.Context, _, _, _ string, _ []string, change func(context.Context) error) error {
	a.changes++
	return change(context.WithValue(ctx, inChangeKey{}, true))
}

// callProbe is a connector, with its manifest declarations, that notes whether
// each call was made inside a change; the calls go on to the fakes.
type callProbe struct {
	*fakeCaller

	calls  int
	inside []string
}

func (p *callProbe) ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file app.NamedFile) (app.Delivery, error) {
	p.note(ctx, "ImportFile")
	return p.fakeCaller.ImportFile(ctx, instance, id, mode, file)
}

func (p *callProbe) ImportBondLots(ctx context.Context, instance string, id ingest.BatchID, e *ingest.BondLotEntries) (app.Delivery, error) {
	p.note(ctx, "ImportBondLots")
	return p.fakeCaller.ImportBondLots(ctx, instance, id, e)
}

func (p *callProbe) Declarations(ctx context.Context, instance string) (ingest.Declarations, error) {
	p.note(ctx, "Declarations")
	return fakeDeclarations{deliveries: true}.Declarations(ctx, instance)
}

func (p *callProbe) note(ctx context.Context, what string) {
	p.calls++
	if inChange(ctx) {
		p.inside = append(p.inside, what)
	}
}

// writeProbe is the ledger that notes whether it was written inside a change.
type writeProbe struct {
	*fakeLedger

	inside int
}

func (w *writeProbe) Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	if inChange(ctx) {
		w.inside++
	}
	return w.fakeLedger.Import(ctx, txs, assertions)
}

func (w *writeProbe) Seed(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	return w.Import(ctx, txs, assertions)
}

func (w *writeProbe) AddOpeningLots(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	return w.Import(ctx, txs, assertions)
}

// batchProbe is the batch history that notes whether a batch was recorded
// inside a change.
type batchProbe struct {
	*fakeRecorder

	inside int
}

func (b *batchProbe) RecordBatch(ctx context.Context, instance string, batch *ingest.Batch) error {
	if inChange(ctx) {
		b.inside++
	}
	return b.fakeRecorder.RecordBatch(ctx, instance, batch)
}

// probes are what a change of the owner's is run over, and what noted of it.
type probes struct {
	audit  *transactionAuditor
	caller *callProbe
	ledger *writeProbe
	batch  *batchProbe
}

// newProbes returns probes whose connector answers delivery.
func newProbes(delivery *app.Delivery) *probes {
	return &probes{
		audit:  &transactionAuditor{},
		caller: &callProbe{fakeCaller: &fakeCaller{delivery: *delivery}},
		ledger: &writeProbe{fakeLedger: &fakeLedger{summary: ingest.ImportSummary{New: 1}}},
		batch:  &batchProbe{fakeRecorder: &fakeRecorder{}},
	}
}

// wire makes c, the connectors of a use case, call and write through the probes.
func (p *probes) wire(c *app.Connectors) {
	c.Caller, c.Ledger, c.Declarations = p.caller, p.ledger, p.caller
}

// check fails t unless the change asked the connector and read its manifest
// outside its transaction, and ran the import and recorded the batch inside it.
func (p *probes) check(t *testing.T) {
	t.Helper()
	if p.audit.changes != 1 || p.caller.calls < 2 {
		t.Fatalf("%d changes and %d calls to the connector, want one change that asked the connector and read its manifest", p.audit.changes, p.caller.calls)
	}
	if len(p.caller.inside) != 0 {
		t.Errorf("calls made inside the change's transaction: %v", p.caller.inside)
	}
	if p.ledger.inside != 1 || p.batch.inside != 1 {
		t.Errorf("the import ran inside the change %d times and the batch was recorded inside it %d times, want once each", p.ledger.inside, p.batch.inside)
	}
}

// The origin and session of the owner's changes in these tests.
const (
	probeOrigin  = "https://ui.example.com"
	probeSession = "owner"
)

// TestStatementUploadMakesNoNetworkCallInsideItsTransaction pins the rule that
// no connector call, and no read of a connector's manifest, is made inside a
// change's transaction (D-211): a statement upload asks the connector first and
// hands its answer to the change, whose transaction then holds only the import
// and the record of the batch.
func TestStatementUploadMakesNoNetworkCallInsideItsTransaction(t *testing.T) {
	t.Parallel()
	delivery := fileDelivery(t)
	p := newProbes(&delivery)
	u := uploads(t, p.caller.fakeCaller, p.ledger.fakeLedger, p.batch.fakeRecorder)
	u.Audit, u.Batches = p.audit, p.batch
	p.wire(u.Connectors)
	file := app.NamedFile{Name: "flex.xml", Contents: strings.NewReader("<FlexQueryResponse/>")}
	if _, err := u.Upload(t.Context(), probeOrigin, probeSession, "acc-1", ingest.ModeTransactions, file); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	p.check(t)
}

// TestBondLotEntryMakesNoNetworkCallInsideItsTransaction is the same for a bond
// lot the owner enters.
func TestBondLotEntryMakesNoNetworkCallInsideItsTransaction(t *testing.T) {
	t.Parallel()
	delivery := bondDelivery(t, bondsAccount)
	p := newProbes(&delivery)
	b := bondLots(t, p.caller.fakeCaller, p.ledger.fakeLedger, &fakeStores{}, []string{"bonds-main"})
	b.Audit, b.Batches = p.audit, p.batch
	p.wire(b.Connectors)
	lot := edoLot(t)
	if _, err := b.Add(t.Context(), probeOrigin, probeSession, bondsAccount, &lot); err != nil {
		t.Fatalf("Add: %v", err)
	}
	p.check(t)
}

// TestBondLotDeletionMakesNoNetworkCallInsideItsTransaction is the same for a
// bond lot the owner deletes, which also holds the instance's sync claim.
func TestBondLotDeletionMakesNoNetworkCallInsideItsTransaction(t *testing.T) {
	t.Parallel()
	delivery := bondDelivery(t, bondsAccount)
	p := newProbes(&delivery)
	b := bondLots(t, p.caller.fakeCaller, p.ledger.fakeLedger, &fakeStores{}, []string{"bonds-main"})
	b.Audit, b.Batches = p.audit, p.batch
	p.wire(b.Connectors)
	lot := edoLot(t)
	if _, err := b.Delete(t.Context(), probeOrigin, probeSession, bondsAccount, lot.Key()); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	p.check(t)
}
