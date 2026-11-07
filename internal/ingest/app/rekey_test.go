package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// rekeyLog records the re-keys of each named fake rekeyer, in order, as
// "<name> <from> -> <to>".
type rekeyLog struct{ calls []string }

// fakeRekeyer is one context's re-key, recording into log; fail makes it
// fail.
type fakeRekeyer struct {
	name string
	log  *rekeyLog
	fail error
}

func (f fakeRekeyer) RekeyInstrument(_ context.Context, from, to ledger.Instrument) error {
	if f.fail != nil {
		return f.fail
	}
	f.log.calls = append(f.log.calls, f.name+" "+from.String()+" -> "+to.String())
	return nil
}

// errRekey is a synthetic failure of a re-key.
var errRekey = errors.New("synthetic re-key failure")

// registryISIN is the synthetic registry ISIN of series code, held under
// provisional.
func registryISIN(t *testing.T, provisional, code, registry string) ingest.RegistryISIN {
	t.Helper()
	isin, err := ledger.ParseISIN(registry)
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	return ingest.RegistryISIN{Instrument: pricingInstrument(t, provisional, code), ISIN: isin}
}

// TestConnectorsRekey checks D-214 at the ingest path: a sync's registry
// ISINs re-key each provisional instrument with every rekeyer in turn, the
// ledger's last, after the rest of the delivery is imported; a connector
// that may not deliver bond terms has them dropped with its undeclared
// item; an invalid one refuses the whole delivery before anything is
// imported; and a failing re-key fails the batch, so the next sync tries
// again.
func TestConnectorsRekey(t *testing.T) {
	t.Parallel()
	edo := registryISIN(t, "PLEDO0536000", "EDO0536", "PL9999000018")
	coi := registryISIN(t, "PLCOI1130004", "COI1130", "PL9999000034")
	tests := []struct {
		name      string
		isins     []ingest.RegistryISIN
		declared  bool
		fail      error
		wantCalls []string
		wantErr   bool
		imports   int
	}{
		{
			name: "each instrument re-keyed by every context, the ledger last", isins: []ingest.RegistryISIN{edo, coi}, declared: true, imports: 1,
			wantCalls: []string{
				"values PLCOI1130004/COI1130 -> PL9999000034/COI1130", "ledger PLCOI1130004/COI1130 -> PL9999000034/COI1130",
				"values PLEDO0536000/EDO0536 -> PL9999000018/EDO0536", "ledger PLEDO0536000/EDO0536 -> PL9999000018/EDO0536",
			},
		},
		{name: "undeclared: dropped, the rest imported", isins: []ingest.RegistryISIN{edo}, imports: 1},
		{
			name: "invalid: nothing imported", declared: true, wantErr: true,
			isins: []ingest.RegistryISIN{registryISIN(t, "PL9999000026", "EDO0536", "PL9999000018")},
		},
		{name: "a failing re-key fails the batch", isins: []ingest.RegistryISIN{edo}, declared: true, fail: errRekey, wantErr: true, imports: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := delivery(t)
			d.Institution, d.Prices, d.FXRates, d.IndexLevels, d.IndexReadings = "treasury-bonds", nil, nil, nil, nil
			d.RegistryISINs = tt.isins
			l, log := &fakeLedger{}, &rekeyLog{}
			c := connectors(t, &fakeCaller{delivery: d}, l, &fakeStores{})
			c.Declarations = fakeDeclarations{byInstance: map[string]ingest.Declarations{"bonds-1": {BondTerms: tt.declared}}}
			c.Rekeyers = []app.InstrumentRekeyer{fakeRekeyer{name: "values", log: log, fail: tt.fail}, fakeRekeyer{name: "ledger", log: log}}
			_, err := c.Sync(t.Context(), "bonds-1", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Sync = %v, want an error %t", err, tt.wantErr)
			}
			if diff := cmp.Diff([]any{tt.wantCalls, tt.imports}, []any{log.calls, l.calls}); diff != "" {
				t.Errorf("re-keys and ledger imports (-want +got):\n%s", diff)
			}
		})
	}
	c := connectors(t, &fakeCaller{delivery: app.Delivery{Institution: "treasury-bonds", RegistryISINs: []ingest.RegistryISIN{edo}}}, &fakeLedger{}, &fakeStores{})
	if _, err := c.Sync(t.Context(), "bonds-1", &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions}); !errors.Is(err, app.ErrNoRekeyer) {
		t.Errorf("Sync without rekeyers = %v, want ErrNoRekeyer", err)
	}
}

// postings is a TransactionReader of fixed current transactions by account.
type postings map[ledger.AccountID][]ledger.Transaction

func (p postings) CurrentTransactions(_ context.Context, account ledger.AccountID) ([]ledger.Transaction, error) {
	if account == "acc-broken" {
		return nil, errCaller
	}
	return p[account], nil
}

// TestHoldingsScopeProvisional names every instrument the asked accounts
// post to under a provisional ISIN, held or not, each once, and none
// without Postings (D-214); a sync carries them.
func TestHoldingsScopeProvisional(t *testing.T) {
	t.Parallel()
	edo := pricingInstrument(t, "PLEDO0536000", "EDO0536")
	registered := pricingInstrument(t, "PL9999000018", "EDO1136")
	pln := mustCurrency(t, "PLN")
	tx := func(in ledger.Instrument) ledger.Transaction {
		return ledger.Transaction{Postings: []ledger.Posting{{Book: ledger.HoldingBook(in), Amount: kernel.NewMoney(kernel.NewDecimal(100, 0), pln)}}}
	}
	h := &app.HoldingsScope{Postings: postings{"acc-1": {tx(edo), tx(registered)}, "acc-2": {tx(edo)}, "acc-3": {tx(registered)}}}
	got, err := h.ProvisionalInstruments(t.Context(), []ledger.AccountID{"acc-1", "acc-2", "acc-3"})
	if err != nil {
		t.Fatalf("ProvisionalInstruments: %v", err)
	}
	if diff := cmp.Diff([]ledger.Instrument{edo}, got); diff != "" {
		t.Errorf("ProvisionalInstruments (-want +got):\n%s", diff)
	}
	if _, err := h.ProvisionalInstruments(t.Context(), []ledger.AccountID{"acc-broken"}); !errors.Is(err, errCaller) {
		t.Errorf("ProvisionalInstruments of a failing read = %v, want its error", err)
	}
	if got, err := (&app.HoldingsScope{}).ProvisionalInstruments(t.Context(), []ledger.AccountID{"acc-1"}); got != nil || err != nil {
		t.Errorf("ProvisionalInstruments without Postings = %v, %v, want none", got, err)
	}

	caller := &fakeCaller{delivery: app.Delivery{Institution: "treasury-bonds"}}
	conns := connectors(t, caller, &fakeLedger{}, &fakeStores{})
	s := &app.Scheduler{
		Connectors: conns, Instances: listed{names: []string{"bonds"}}, Store: newMemStore(), Provisional: h,
		Policy:  ingest.SyncPolicy{Interval: time.Hour, RetryFirst: time.Minute, RetryMax: time.Hour},
		Timeout: time.Minute, Now: t0, NewBatchID: func() ingest.BatchID { return "b-1" },
		Jitter: func(time.Duration) time.Duration { return 0 },
	}
	if _, err := s.SyncNow(t.Context(), "bonds", ingest.ModeTransactions); err != nil {
		t.Fatalf("SyncNow: %v", err)
	}
	if diff := cmp.Diff([]ledger.Instrument{edo}, caller.provisional); diff != "" {
		t.Errorf("provisional instruments sent (-want +got):\n%s", diff)
	}
}
