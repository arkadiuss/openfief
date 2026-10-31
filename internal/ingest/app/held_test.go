package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// heldLedger holds one open lot of one instrument per account in lots,
// and counts the day it was asked about.
type heldLedger struct {
	lots  map[ledger.AccountID]ledger.Lot
	asked kernel.Date
	err   error
}

func (h *heldLedger) HoldingsOn(_ context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Holding, error) {
	h.asked = on
	l, ok := h.lots[account]
	if !ok {
		return nil, h.err
	}
	return []ledger.Holding{{Account: account, Instrument: l.Instrument, Date: on, Quantity: l.Quantity}}, h.err
}

func (h *heldLedger) LotsOn(_ context.Context, account ledger.AccountID, _ kernel.Date) ([]ledger.Lot, error) {
	l, ok := h.lots[account]
	if !ok {
		return nil, nil
	}
	return []ledger.Lot{l}, nil
}

// liveSince says every account in days is seeded on its date.
type liveSince map[ledger.AccountID]kernel.Date

func (l liveSince) GoLive(_ context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	d, ok := l[account]
	return d, ok, nil
}

// wrappers are the recorded tax wrappers.
type wrappers map[ledger.AccountID]jurisdiction.Wrapper

func (w wrappers) TaxWrapper(_ context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	v, ok := w[account]
	return v, ok, nil
}

func TestHoldingsScope(t *testing.T) {
	t.Parallel()
	plnCurrency, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	bond := pricingInstrument(t, "PLEDO0536000", "EDO0536")
	stock := pricingInstrument(t, "US0378331005", "AAPL")
	today := march2(t)
	lot := func(in ledger.Instrument) ledger.Lot {
		return ledger.Lot{Instrument: in, Opened: today, Quantity: kernel.NewQuantity(kernel.NewDecimal(4, 0)), Cost: kernel.NewMoney(kernel.NewDecimal(400, 0), plnCurrency)}
	}
	book := &heldLedger{lots: map[ledger.AccountID]ledger.Lot{"acc-bonds": lot(bond), "acc-ike": lot(bond), "acc-stocks": lot(stock)}}
	h := &app.HoldingsScope{
		Ledger: book, Seeding: liveSince{"acc-bonds": today}, Bonds: bondISINs{"PLEDO0536000": true},
		Wrappers: wrappers{"acc-ike": "ike"}, Default: "taxable", Today: func() kernel.Date { return today },
	}
	held, terms, err := h.HeldAccounts(t.Context(), []ledger.AccountID{"acc-stocks", "acc-ike", "acc-bonds", "acc-ike"})
	if err != nil {
		t.Fatalf("HeldAccounts: %v", err)
	}
	type row struct {
		Account, GoLive, Wrapper string
		Holdings, Lots           int
	}
	got := make([]row, 0, len(held))
	for _, a := range held {
		goLive := ""
		if !a.GoLive.IsZero() {
			goLive = a.GoLive.String()
		}
		got = append(got, row{string(a.Account), goLive, a.Wrapper, len(a.Holdings), len(a.Lots)})
	}
	// In ID order and each once; go-live only for a seeded account; the
	// pack's default wrapper unless one is recorded.
	want := []row{
		{"acc-bonds", today.String(), "taxable", 1, 1},
		{"acc-ike", "", "ike", 1, 1},
		{"acc-stocks", "", "taxable", 1, 1},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("HeldAccounts (-want +got):\n%s", diff)
	}
	if len(terms) != 1 || terms[0].Instrument != bond {
		t.Errorf("bond terms = %v, want EDO0536's once", terms)
	}
	if !book.asked.Equal(today) {
		t.Errorf("holdings asked on %s, want today %s", book.asked, today)
	}

	failing := &app.HoldingsScope{
		Ledger: &heldLedger{err: errCaller}, Seeding: liveSince{}, Bonds: bondISINs{}, Wrappers: wrappers{},
		Today: func() kernel.Date { return today },
	}
	if _, _, err := failing.HeldAccounts(t.Context(), []ledger.AccountID{"acc-x"}); !errors.Is(err, errCaller) {
		t.Errorf("HeldAccounts with a failing ledger = %v, want %v", err, errCaller)
	}
}

// fixedHeld is a HeldSource answering held for any accounts, and records
// the accounts asked.
type fixedHeld struct {
	held  []ingest.HeldAccount
	asked []ledger.AccountID
}

func (f *fixedHeld) HeldAccounts(_ context.Context, accounts []ledger.AccountID) ([]ingest.HeldAccount, []ingest.BondTerms, error) {
	f.asked = accounts
	return f.held, nil, nil
}

// TestSchedulerSendsWhatBoundAccountsHold checks that every sync of an
// instance bound to accounts carries what they hold (D-117), and a
// market-data connector, bound to none, gets none.
func TestSchedulerSendsWhatBoundAccountsHold(t *testing.T) {
	t.Parallel()
	held := []ingest.HeldAccount{{Account: "acc-1", Wrapper: "taxable"}}
	for _, tt := range []struct {
		name  string
		bound []ledger.AccountID
		want  []ingest.HeldAccount
	}{
		{name: "a connector bound to accounts", bound: []ledger.AccountID{"acc-1"}, want: held},
		{name: "a market-data connector"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			caller := &fakeCaller{delivery: app.Delivery{Institution: "treasury-bonds"}}
			conns := connectors(t, caller, &fakeLedger{}, &fakeStores{})
			conns.Bindings = &fakeBindings{accounts: tt.bound}
			source := &fixedHeld{held: held}
			s := &app.Scheduler{
				Connectors: conns, Instances: listed{names: []string{"bonds"}}, Store: newMemStore(), Held: source,
				Policy:  ingest.SyncPolicy{Interval: time.Hour, RetryFirst: time.Minute, RetryMax: time.Hour},
				Timeout: time.Minute, Now: t0, NewBatchID: func() ingest.BatchID { return "b-1" },
				Jitter: func(time.Duration) time.Duration { return 0 },
			}
			if _, err := s.SyncNow(t.Context(), "bonds", ingest.ModeTransactions); err != nil {
				t.Fatalf("SyncNow: %v", err)
			}
			if diff := cmp.Diff(tt.want, caller.held); diff != "" {
				t.Errorf("held sent (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.bound, source.asked); diff != "" {
				t.Errorf("accounts asked (-want +got):\n%s", diff)
			}
		})
	}
}
