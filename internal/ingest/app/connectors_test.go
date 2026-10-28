package app_test

import (
	"context"
	"errors"
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
	"github.com/arkadiuss/openfief/internal/valuation"
)

var (
	errCaller = errors.New("connector unreachable")
	errStore  = errors.New("store failed")
)

// fakeCaller delivers its delivery or fails, and records what it was asked.
type fakeCaller struct {
	delivery  app.Delivery
	err       error
	instance  string
	id        ingest.BatchID
	mode      ingest.Mode
	watermark string
	file      string
}

func (f *fakeCaller) Sync(_ context.Context, instance string, id ingest.BatchID, mode ingest.Mode, watermark string) (app.Delivery, error) {
	f.instance, f.id, f.mode, f.watermark = instance, id, mode, watermark
	return f.delivery, f.err
}

func (f *fakeCaller) ImportFile(_ context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file app.NamedFile) (app.Delivery, error) {
	f.instance, f.id, f.mode = instance, id, mode
	b, err := io.ReadAll(file.Contents)
	if err != nil {
		return app.Delivery{}, errors.Join(errCaller, err)
	}
	f.file = file.Name + ":" + string(b)
	return f.delivery, f.err
}

// fakeStores records what each valuation input store was given.
type fakeStores struct {
	err      error
	calls    int
	prices   []valuation.Price
	rates    []valuation.FXRate
	readings []ingest.IndexReading
}

func (f *fakeStores) StorePrices(_ context.Context, _ ingest.BatchID, prices []valuation.Price) error {
	f.calls++
	f.prices = prices
	return f.err
}

func (f *fakeStores) StoreFXRates(_ context.Context, _ ingest.BatchID, rates []valuation.FXRate) error {
	f.calls++
	f.rates = rates
	return f.err
}

func (f *fakeStores) StoreIndexReadings(_ context.Context, _ ingest.BatchID, readings []ingest.IndexReading) error {
	f.calls++
	f.readings = readings
	return f.err
}

// delivery returns a synthetic delivery from xtb: mapping(t) plus one price,
// FX rate and index reading, each with a source the core must overwrite.
func delivery(t *testing.T) app.Delivery {
	t.Helper()
	day := march2(t)
	usd := mustCurrency(t, "USD")
	cpi, err := kernel.NewDate(2026, time.February, 1)
	if err != nil {
		t.Fatalf("NewDate: %v", err)
	}
	return app.Delivery{
		Institution: "xtb", Mapping: mapping(t), Watermark: "w-2",
		Prices:        []valuation.Price{{Instrument: "PLTEST000010/TST", Date: day, Close: kernel.NewMoney(kernel.NewDecimal(1234, -2), usd), Source: "spoofed"}},
		FXRates:       []valuation.FXRate{{Currency: usd, Date: day, Rate: kernel.NewDecimal(39876, -4), Source: "spoofed"}},
		IndexReadings: []ingest.IndexReading{{Index: ingest.IndexCPI, Date: cpi, Value: kernel.NewDecimal(49, -3), Source: "spoofed"}},
	}
}

// fakeBindings binds every instance to its accounts, or fails with err.
type fakeBindings struct {
	accounts []ledger.AccountID
	err      error
	calls    int
}

func (f *fakeBindings) BoundAccounts(context.Context, string) ([]ledger.AccountID, error) {
	f.calls++
	return f.accounts, f.err
}

// connectors returns Connectors over the fakes, with the instance bound to
// acc-1, the account of mapping(t).
func connectors(t *testing.T, caller app.ConnectorCaller, l app.LedgerImporter, stores *fakeStores) *app.Connectors {
	t.Helper()
	day := march2(t)
	return &app.Connectors{
		Caller: caller, Ledger: l, Prices: stores, FXRates: stores, Indices: stores,
		Bindings: &fakeBindings{accounts: []ledger.AccountID{"acc-1"}},
		Today:    func() kernel.Date { return day },
	}
}

func TestConnectorsSync(t *testing.T) {
	t.Parallel()
	caller := &fakeCaller{delivery: delivery(t)}
	l := &fakeLedger{summary: ingest.ImportSummary{New: 1}}
	stores := &fakeStores{}
	got, err := connectors(t, caller, l, stores).Sync(t.Context(), "xtb-main", "b-1", ingest.ModeTransactions, "w-1")
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	want := app.SyncResult{Watermark: "w-2", Batch: ingest.Batch{
		ID: "b-1", Institution: "xtb", Source: app.SourceSync, Mode: ingest.ModeTransactions, Started: march2(t),
		Transactions: 1, Assertions: 1, Imported: ingest.ImportSummary{New: 1},
		Problems: []ingest.Problem{ingest.Unsupported("row 2", "broker interest")},
	}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Sync (-want +got):\n%s", diff)
	}
	if caller.instance != "xtb-main" || caller.id != "b-1" || caller.mode != ingest.ModeTransactions || caller.watermark != "w-1" {
		t.Errorf("caller asked %+v", caller)
	}
	wantTxs := mapping(t).Transactions
	wantTxs[0].Key.Source = "plugin:xtb-main:test"
	if diff := cmp.Diff(wantTxs, l.txs); diff != "" {
		t.Errorf("imported transactions (-want +got):\n%s", diff)
	}
	d := delivery(t)
	d.Prices[0].Source, d.FXRates[0].Source, d.IndexReadings[0].Source = "xtb-main", "xtb-main", "xtb-main"
	if diff := cmp.Diff([]any{d.Prices, d.FXRates, d.IndexReadings}, []any{stores.prices, stores.rates, stores.readings}); diff != "" {
		t.Errorf("stored inputs (-want +got):\n%s", diff)
	}
}

func TestConnectorsImportFile(t *testing.T) {
	t.Parallel()
	d := delivery(t)
	d.Prices, d.FXRates, d.IndexReadings, d.Watermark = nil, nil, nil, ""
	caller := &fakeCaller{delivery: d}
	stores := &fakeStores{}
	file := app.NamedFile{Name: "statement.csv", Contents: strings.NewReader("synthetic")}
	l := &fakeLedger{}
	got, err := connectors(t, caller, l, stores).ImportFile(t.Context(), "xtb-main", "b-2", ingest.ModeOpening, file)
	if err != nil {
		t.Fatalf("ImportFile: %v", err)
	}
	if l.seeds != 1 {
		t.Errorf("an opening batch seeded %d times, want 1", l.seeds)
	}
	if got.ID != "b-2" || got.Source != app.SourceFile || got.Mode != ingest.ModeOpening || got.Transactions != 1 {
		t.Errorf("ImportFile = %+v", got)
	}
	if caller.file != "statement.csv:synthetic" {
		t.Errorf("connector got %q", caller.file)
	}
	if stores.calls != 0 {
		t.Errorf("stores called %d times for a delivery without valuation inputs", stores.calls)
	}
}

func TestConnectorsFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		edit       func(d *app.Delivery)
		callerErr  error
		ledgerErr  error
		storeErr   error
		mode       ingest.Mode
		wantErr    error
		wantImport bool
	}{
		{name: "connector fails", callerErr: errCaller, mode: ingest.ModeTransactions, wantErr: errCaller},
		{name: "no institution", edit: func(d *app.Delivery) { d.Institution = "" }, mode: ingest.ModeTransactions, wantErr: ingest.ErrInvalidBatch},
		{name: "no mode", wantErr: ingest.ErrInvalidBatch},
		{
			name: "invalid index reading", mode: ingest.ModeTransactions, wantErr: ingest.ErrInvalidIndexReading,
			edit: func(d *app.Delivery) { d.IndexReadings[0].Date = d.IndexReadings[0].Date.AddDays(3) },
		},
		{name: "ledger fails", ledgerErr: errLedger, mode: ingest.ModeTransactions, wantErr: errLedger, wantImport: true},
		{name: "store fails", storeErr: errStore, mode: ingest.ModeTransactions, wantErr: errStore, wantImport: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := delivery(t)
			if tt.edit != nil {
				tt.edit(&d)
			}
			l := &fakeLedger{err: tt.ledgerErr}
			c := connectors(t, &fakeCaller{delivery: d, err: tt.callerErr}, l, &fakeStores{err: tt.storeErr})
			_, err := c.Sync(t.Context(), "xtb-main", "b-1", tt.mode, "")
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Sync error (-want +got):\n%s", diff)
			}
			if got := l.calls == 1; got != tt.wantImport {
				t.Errorf("ledger imported = %v, want %v", got, tt.wantImport)
			}
			_, err = c.ImportFile(t.Context(), "xtb-main", "b-1", tt.mode, app.NamedFile{Contents: strings.NewReader("")})
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("ImportFile error (-want +got):\n%s", diff)
			}
		})
	}
}
