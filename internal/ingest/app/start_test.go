package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// fakeHistory says whether every account holds transactions, or fails.
type fakeHistory struct {
	booked bool
	err    error
}

func (f fakeHistory) HasTransactionsFrom(context.Context, ledger.AccountID, string) (bool, error) {
	return f.booked, f.err
}

// syntheticCash is a synthetic opening cash balance in PLN.
func syntheticCash(t *testing.T, amount string) kernel.Money {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	m, err := kernel.ParseMoney(amount, pln)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// TestUploadsStart checks how an account starts (D-136): one that holds
// transactions but isn't seeded started from its whole history, so it
// takes transactions and is never seeded; the opening cash the owner
// enters reaches the connector with an opening upload only.
func TestUploadsStart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		booked   bool
		seeded   bool
		override ingest.Mode
		cash     string
		wantMode ingest.Mode
		wantErr  error
	}{
		{name: "a new account starts from its positions and the cash entered", cash: "2500.50", wantMode: ingest.ModeOpening},
		{name: "a new account starts from its whole history", override: ingest.ModeTransactions, wantMode: ingest.ModeTransactions},
		{name: "an account started from its whole history continues", booked: true, wantMode: ingest.ModeTransactions},
		{name: "an account started from its whole history is never seeded", booked: true, override: ingest.ModeOpening, wantErr: ingest.ErrStartedFromHistory},
		{name: "opening cash with transactions", override: ingest.ModeTransactions, cash: "1", wantErr: ingest.ErrOpeningCashOutsideOpening},
		{name: "opening cash for an account started from its whole history", booked: true, cash: "1", wantErr: ingest.ErrOpeningCashOutsideOpening},
		{name: "opening cash entered again for a seeded account", seeded: true, booked: true, override: ingest.ModeOpening, cash: "1", wantErr: ingest.ErrOpeningCashEntered},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u, caller := startUploads(t, tt.booked, tt.seeded)
			file := startFile(t, tt.cash)
			got, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", tt.override, file)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Upload() error = %v, want %v", err, tt.wantErr)
			}
			wantCash := file.OpeningCash
			if tt.wantErr != nil {
				wantCash = nil
			}
			if got.Mode != tt.wantMode || caller.cash != wantCash {
				t.Errorf("Upload() mode %s, cash handed %v; want %s, %v", got.Mode, caller.cash, tt.wantMode, wantCash)
			}
		})
	}
}

// startUploads returns the uploads of acc-1, for the owner, whose
// transactions booked says it holds and seeded says it is seeded, with the
// connector it calls.
func startUploads(t *testing.T, booked, seeded bool) (*app.Uploads, *fakeCaller) {
	t.Helper()
	caller := &fakeCaller{delivery: fileDelivery(t)}
	u := uploads(t, caller, &fakeLedger{}, &fakeRecorder{})
	u.Owners, u.History = fakeOwners{}, fakeHistory{booked: booked}
	if seeded {
		u.Connectors.Seeding = fixedSeeding{goLive: march2(t)}
	}
	return u, caller
}

// startFile is a statement the owner uploads with cash entered, none when
// cash is empty.
func startFile(t *testing.T, cash string) app.NamedFile {
	t.Helper()
	file := app.NamedFile{Name: "s.xlsx", Contents: strings.NewReader("statement")}
	if cash != "" {
		m := syntheticCash(t, cash)
		file.OpeningCash = &m
	}
	return file
}

// TestUploadsStartHistoryUnreadable checks that an upload whose account's
// transactions can't be read imports nothing.
func TestUploadsStartHistoryUnreadable(t *testing.T) {
	t.Parallel()
	u := uploads(t, &fakeCaller{delivery: fileDelivery(t)}, &fakeLedger{}, &fakeRecorder{})
	u.Owners, u.History = fakeOwners{}, fakeHistory{err: errStore}
	if _, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", 0,
		app.NamedFile{Name: "s.xlsx", Contents: strings.NewReader("statement")}); !errors.Is(err, errStore) {
		t.Errorf("Upload with the history unreadable = %v, want %v", err, errStore)
	}
}

// fakeTickerStore holds the waiting entries of one ticker, the listing
// confirming it settles, and records the keys booked.
type fakeTickerStore struct {
	app.TickerStore

	waiting []ingest.TickerEntry
	listing ingest.Listing
	booked  []ledger.TransactionKey
}

func (f *fakeTickerStore) ConfirmTicker(_ context.Context, c ingest.TickerConfirmation) ([]ingest.TickerEntry, ingest.Listing, error) {
	l := f.listing
	l.ISIN = c.ISIN
	return f.waiting, l, nil
}

func (f *fakeTickerStore) BookedTickerEntries(_ context.Context, _ string, keys []ledger.TransactionKey) error {
	f.booked = append(f.booked, keys...)
	return nil
}

// TestTickersConfirmOpeningLots checks that the opening lots of an
// account seeded from an xStation export's open positions, waiting for
// their ticker's ISIN, join the account's seeding one at a time once the
// owner confirms it (D-136), while other entries are imported.
func TestTickersConfirmOpeningLots(t *testing.T) {
	t.Parallel()
	day, err := kernel.ParseDate("2026-06-30")
	if err != nil {
		t.Fatal(err)
	}
	opened, err := kernel.ParseDate("2025-11-04")
	if err != nil {
		t.Fatal(err)
	}
	entry := func(kind ledger.TransactionKind, id string, lotOpened kernel.Date) ingest.TickerEntry {
		return ingest.TickerEntry{
			Symbol: "TESTA.PL", Kind: kind, Quantity: kernel.NewQuantity(kernel.NewDecimal(7, 0)), Amount: syntheticCash(t, "864.15"),
			Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb", ExternalID: id}, Account: "acc-1", Date: day},
			Opened: lotOpened,
		}
	}
	isin, err := ledger.ParseISIN("PLTESTA00002")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		waiting     []ingest.TickerEntry
		openingLots int
		imports     int
	}{
		{name: "an opening lot joins the seeding", waiting: []ingest.TickerEntry{entry(ledger.KindOpeningLot, "open-position:1", opened), entry(ledger.KindBuy, "9000000003", kernel.Date{})}, openingLots: 1, imports: 1},
		{name: "a buy is imported", waiting: []ingest.TickerEntry{entry(ledger.KindBuy, "9000000003", kernel.Date{})}, imports: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &fakeLedger{}
			store := &fakeTickerStore{waiting: tt.waiting}
			tickers := app.Tickers{Owners: fakeOwners{}, Store: store, Ledger: l}
			if _, err := tickers.Confirm(t.Context(), "https://ui.example.com", "owner",
				ingest.TickerConfirmation{Instance: "xtb-main", Symbol: "TESTA.PL", ISIN: isin}); err != nil {
				t.Fatalf("Confirm: %v", err)
			}
			if l.openingLots != tt.openingLots || l.calls != tt.imports || len(l.txs) != len(tt.waiting) {
				t.Errorf("Confirm added opening lots %d times and imported %d times %d transactions; want %d, %d, %d",
					l.openingLots, l.calls, len(l.txs), tt.openingLots, tt.imports, len(tt.waiting))
			}
			if len(store.booked) != len(tt.waiting) {
				t.Errorf("booked %d keys, want %d", len(store.booked), len(tt.waiting))
			}
		})
	}
}
