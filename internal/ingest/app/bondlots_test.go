package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// bondsAccount is the synthetic account the bond lots go into.
const bondsAccount ledger.AccountID = "acc-bonds"

// fakeOwners accepts the session "owner" and refuses any other with err.
type fakeOwners struct {
	err error
}

func (f fakeOwners) CheckOwner(_ context.Context, _, session string) error {
	if session == "owner" {
		return nil
	}
	return f.err
}

// fakeAccountBindings binds instances to every account, or fails with err.
type fakeAccountBindings struct {
	instances []string
	err       error
}

func (f fakeAccountBindings) BoundInstances(context.Context, ledger.AccountID) ([]string, error) {
	return f.instances, f.err
}

// fakeRecorder records the batches it is given, or fails with err.
type fakeRecorder struct {
	err       error
	instances []string
	batches   []ingest.Batch
}

func (f *fakeRecorder) RecordBatch(_ context.Context, instance string, b *ingest.Batch) error {
	if f.err != nil {
		return f.err
	}
	f.instances, f.batches = append(f.instances, instance), append(f.batches, *b)
	return nil
}

// edoLot is a synthetic lot of 40 EDO0536 bonds bought on 2026-05-04.
func edoLot(t *testing.T) ingest.BondLot {
	t.Helper()
	series, err := pl.ParseSeries("EDO0536")
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	purchased, err := kernel.ParseDate("2026-05-04")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	fee, err := kernel.ParseMoney("2.00", pl.Nominal().Currency())
	if err != nil {
		t.Fatalf("ParseMoney: %v", err)
	}
	return ingest.BondLot{
		Terms:    pl.Terms{Series: series, FirstPeriodRate: kernel.NewDecimal(6, -2), Margin: kernel.NewDecimal(2, -2), EarlyRedemptionFee: fee},
		Quantity: kernel.NewQuantity(kernel.NewDecimal(40, 0)), Purchased: purchased,
	}
}

// bondDelivery is what a bond connector delivers for edoLot in account: a
// deposit and a buy, and the series' terms.
func bondDelivery(t *testing.T, account ledger.AccountID) app.Delivery {
	t.Helper()
	l := edoLot(t)
	isin, err := ledger.ParseISIN("PLEDO0536000")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	in := ledger.Instrument{ISIN: isin, Symbol: "EDO0536"}
	header := func(id string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "bonds", ExternalID: id}, Account: account, Date: l.Purchased}
	}
	deposit, err1 := ledger.NewDeposit(header("deposit"), l.Cost())
	buy, err2 := ledger.NewBuy(header("buy"), in, l.Quantity, l.Cost(), kernel.Money{})
	if err := errors.Join(err1, err2); err != nil {
		t.Fatalf("building the delivery: %v", err)
	}
	return app.Delivery{
		Institution: "treasury-bonds",
		Mapping:     ingest.Mapping{Transactions: []ledger.Transaction{deposit, buy}},
		BondTerms:   []ingest.BondTerms{{Instrument: in, Terms: l.Terms}},
	}
}

// bondLots returns BondLots over the fakes, with today 2026-10-29 and
// batch IDs b-1, b-2, ...
func bondLots(t *testing.T, caller *fakeCaller, l *fakeLedger, stores *fakeStores, instances []string) *app.BondLots {
	t.Helper()
	today, err := kernel.ParseDate("2026-10-29")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	c := connectors(t, caller, l, stores)
	c.Bindings = &fakeBindings{accounts: []ledger.AccountID{bondsAccount}}
	n := 0
	return &app.BondLots{
		Owners: fakeOwners{err: app.ErrNoSession}, Bindings: fakeAccountBindings{instances: instances}, Connectors: c,
		Batches:    &fakeRecorder{},
		Today:      func() kernel.Date { return today },
		NewBatchID: func() ingest.BatchID { n++; return ingest.BatchID(fmt.Sprintf("b-%d", n)) },
	}
}

func TestBondLotsAdd(t *testing.T) {
	t.Parallel()
	caller := &fakeCaller{delivery: bondDelivery(t, bondsAccount)}
	l := &fakeLedger{summary: ingest.ImportSummary{New: 2}}
	stores := &fakeStores{}
	lot := edoLot(t)
	b := bondLots(t, caller, l, stores, []string{"bonds-main"})
	got, err := b.Add(t.Context(), "https://ui.example.com", "owner", bondsAccount, &lot)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	want := ingest.Batch{
		ID: "b-1", Institution: "treasury-bonds", Source: app.SourceEntry, Mode: ingest.ModeTransactions, Started: march2(t),
		Transactions: 2, Imported: ingest.ImportSummary{New: 2}, Accounts: []ledger.AccountID{bondsAccount},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Add batch (-want +got):\n%s", diff)
	}
	rec, isFake := b.Batches.(*fakeRecorder)
	if !isFake {
		t.Fatalf("Batches is %T, want the fake recorder", b.Batches)
	}
	if diff := cmp.Diff([]ingest.Batch{want}, rec.batches, cmpopts.EquateEmpty()); diff != "" || len(rec.instances) != 1 || rec.instances[0] != "bonds-main" {
		t.Errorf("recorded %v for %v; want the batch once for bonds-main (-want +got):\n%s", rec.batches, rec.instances, diff)
	}
	checkHandedOver(t, caller, l, stores)
}

// checkHandedOver checks that the lot went to bonds-main as batch b-1 and
// was imported (checkImported).
func checkHandedOver(t *testing.T, caller *fakeCaller, l *fakeLedger, stores *fakeStores) {
	t.Helper()
	if caller.instance != "bonds-main" || caller.id != "b-1" || len(caller.lots) != 1 || caller.lots[0].Series().String() != "EDO0536" {
		t.Errorf("handed %v as %s to %s; want the lot as b-1 to the bound instance bonds-main", caller.lots, caller.id, caller.instance)
	}
	// The account is not seeded, so it goes live today with this lot (D-98).
	if !caller.entries.GoLive.Equal(march2(t)) {
		t.Errorf("go-live sent = %s, want today %s for an account not seeded", caller.entries.GoLive, march2(t))
	}
	checkImported(t, l, stores)
}

// checkImported checks that the lot's transactions were imported once,
// unseeded, with keys scoped to the instance, and that EDO0536's terms were
// stored.
func checkImported(t *testing.T, l *fakeLedger, stores *fakeStores) {
	t.Helper()
	if l.seeds != 0 || l.calls != 1 || l.openingLots != 1 {
		t.Errorf("ledger seeds = %d, calls = %d, opening-lot imports = %d; want one import that takes opening lots one at a time and no seeding", l.seeds, l.calls, l.openingLots)
	}
	for i := range l.txs {
		if l.txs[i].Key.Source != "plugin:bonds-main:bonds" {
			t.Errorf("imported key %s; want its source scoped to the instance (D-48)", l.txs[i].Key)
		}
	}
	if len(stores.terms) != 1 || stores.terms[0].Instrument.Symbol != "EDO0536" {
		t.Errorf("stored terms %v; want EDO0536's", stores.terms)
	}
}

// refusal is one way BondLots.Add refuses a lot: the session, the lot, the
// instances bound to the account, and what the connector does.
type refusal struct {
	name      string
	session   string
	owners    error
	lot       func(l *ingest.BondLot)
	instances []string
	bindErr   error
	delivered ledger.AccountID
	callErr   error
	wantErr   error
}

// ok is a refusal's starting point: everything as in a successful Add.
func ok(name string) *refusal {
	return &refusal{name: name, session: "owner", lot: func(*ingest.BondLot) {}, instances: []string{"bonds-main"}, delivered: bondsAccount}
}

func TestBondLotsAddRefuses(t *testing.T) {
	t.Parallel()
	tests := []*refusal{
		with(ok("no session"), func(r *refusal) { r.session, r.owners, r.wantErr = "nope", app.ErrNoSession, app.ErrNoSession }),
		with(ok("a read-scoped session"), func(r *refusal) { r.session, r.owners, r.wantErr = "reader", app.ErrNotOwner, app.ErrNotOwner }),
		with(ok("a lot bought after today"), func(r *refusal) {
			r.lot = func(l *ingest.BondLot) {
				l.Terms.Series.MaturityMonth, l.Terms.Series.MaturityYear = 11, 2036
				l.Purchased = l.Purchased.AddDays(184) // 2026-11-04
			}
			r.wantErr = ingest.ErrInvalidBondLot
		}),
		with(ok("a lot under the wrong series"), func(r *refusal) {
			r.lot, r.wantErr = func(l *ingest.BondLot) { l.Purchased = l.Purchased.AddDays(40) }, ingest.ErrInvalidBondLot
		}),
		with(ok("no connector bound to the account"), func(r *refusal) { r.instances, r.wantErr = nil, app.ErrNoConnector }),
		with(ok("two connectors bound to the account"), func(r *refusal) {
			r.instances, r.wantErr = []string{"bonds-a", "bonds-b"}, app.ErrNoConnector
		}),
		with(ok("bindings that can't be read"), func(r *refusal) { r.bindErr, r.wantErr = errStore, errStore }),
		with(ok("the connector refuses the lot"), func(r *refusal) {
			r.callErr, r.wantErr = fmt.Errorf("%w: synthetic", app.ErrEntryRefused), app.ErrEntryRefused
		}),
		with(ok("the connector takes no bond lots"), func(r *refusal) {
			r.callErr, r.wantErr = fmt.Errorf("%w: synthetic", app.ErrNoEntries), app.ErrNoEntries
		}),
		with(ok("a delivery into another account"), func(r *refusal) { r.delivered, r.wantErr = "acc-other", ingest.ErrAccountNotBound }),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			caller := &fakeCaller{delivery: bondDelivery(t, tt.delivered), err: tt.callErr}
			l, stores := &fakeLedger{}, &fakeStores{}
			b := bondLots(t, caller, l, stores, tt.instances)
			b.Bindings = fakeAccountBindings{instances: tt.instances, err: tt.bindErr}
			b.Owners = fakeOwners{err: tt.owners}
			lot := edoLot(t)
			tt.lot(&lot)
			if _, err := b.Add(t.Context(), "https://ui.example.com", tt.session, bondsAccount, &lot); !errors.Is(err, tt.wantErr) {
				t.Errorf("Add error = %v, want %v", err, tt.wantErr)
			}
			if l.calls != 0 || stores.calls != 0 {
				t.Errorf("a refused lot reached the ledger %d times and the stores %d times; want neither", l.calls, stores.calls)
			}
		})
	}
}

// with returns r changed by change.
func with(r *refusal, change func(r *refusal)) *refusal {
	change(r)
	return r
}

func TestBondLotsDelete(t *testing.T) {
	t.Parallel()
	caller := &fakeCaller{delivery: bondDelivery(t, bondsAccount)}
	l := &fakeLedger{summary: ingest.ImportSummary{NewVersions: 2}}
	b := bondLots(t, caller, l, &fakeStores{}, []string{"bonds-main"})
	lot := edoLot(t)
	got, err := b.Delete(t.Context(), "https://ui.example.com", "owner", bondsAccount, lot.Key())
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got.ID != "b-1" || got.Source != app.SourceEntry || got.Imported.NewVersions != 2 {
		t.Errorf("Delete batch = %+v, want b-1 from the owner's entry with the new versions", got)
	}
	want := ingest.BondLotEntries{GoLive: march2(t), Deleted: []ingest.BondLotKey{lot.Key()}}
	if diff := cmp.Diff(want, caller.entries, cmpopts.EquateEmpty()); diff != "" || caller.instance != "bonds-main" {
		t.Errorf("handed to %s (-want +got):\n%s", caller.instance, diff)
	}
	if l.openingLots != 1 {
		t.Errorf("opening-lot imports = %d, want the deletion imported once", l.openingLots)
	}

	if _, err := b.Delete(t.Context(), "https://ui.example.com", "nobody", bondsAccount, lot.Key()); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Delete without a session: %v, want ErrNoSession", err)
	}
	if _, err := b.Delete(t.Context(), "https://ui.example.com", "owner", bondsAccount, ingest.BondLotKey{Series: lot.Series()}); !errors.Is(err, ingest.ErrInvalidBondLot) {
		t.Errorf("Delete without a date: %v, want ErrInvalidBondLot", err)
	}
}
