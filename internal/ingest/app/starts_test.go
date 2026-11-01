package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// memStarts keeps the go-live opening uploads recorded, first one wins.
type memStarts struct {
	starts map[ledger.AccountID]kernel.Date
}

func (m *memStarts) RecordStarts(_ context.Context, starts map[ledger.AccountID]kernel.Date) error {
	for a, d := range starts {
		if _, ok := m.starts[a]; !ok {
			m.starts[a] = d
		}
	}
	return nil
}

func (m *memStarts) Start(_ context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	d, ok := m.starts[account]
	return d, ok, nil
}

// waitingTickers confirms no ticker, so every ticker entry waits.
type waitingTickers struct {
	app.TickerStore

	held int
}

func (w *waitingTickers) ConfirmedTickers(context.Context, string) (map[string]ingest.Listing, error) {
	return map[string]ingest.Listing{}, nil
}

func (w *waitingTickers) HoldTickers(_ context.Context, _ string, _ []ingest.Ticker, waiting []ingest.TickerEntry) ([]ingest.Ticker, error) {
	w.held += len(waiting)
	return nil, nil
}

// TestUploadsOpeningStartsAccount checks that an opening upload whose
// positions all wait for their ticker's ISIN starts the account at once
// (D-160): its go-live is recorded with the upload, the next upload is in
// transactions mode, and that upload books nothing dated on or before the
// go-live.
func TestUploadsOpeningStartsAccount(t *testing.T) {
	t.Parallel()
	goLive := march2(t)
	lot := ingest.TickerEntry{
		Symbol: "TESTA.PL", Kind: ledger.KindOpeningLot, Quantity: kernel.NewQuantity(kernel.NewDecimal(7, 0)),
		Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb", ExternalID: "open-position:1"}, Account: "acc-1", Date: goLive},
		Opened: goLive, OpenPrice: kernel.NewDecimal(12345, -2), Description: "opening lot TESTA.PL", Ref: "Open Positions#1",
	}
	caller := &fakeCaller{delivery: app.Delivery{Institution: "xtb", Mapping: ingest.Mapping{
		Tickers: []ingest.Ticker{{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL"}}, TickerEntries: []ingest.TickerEntry{lot},
	}}}
	l := &fakeLedger{}
	u := uploads(t, caller, l, &fakeRecorder{})
	tickers, starts := &waitingTickers{}, &memStarts{starts: map[ledger.AccountID]kernel.Date{}}
	u.Owners, u.History = fakeOwners{}, fakeHistory{}
	u.Connectors.Tickers, u.Connectors.Starts = tickers, starts
	first, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", 0, app.NamedFile{Name: "s.xlsx", Contents: strings.NewReader("export")})
	if err != nil || first.Mode != ingest.ModeOpening || tickers.held != 1 {
		t.Fatalf("first Upload = %s, %v, %d entries held; want an opening batch holding the position", first.Mode, err, tickers.held)
	}
	if diff := cmp.Diff(map[ledger.AccountID]kernel.Date{"acc-1": goLive}, starts.starts); diff != "" {
		t.Errorf("recorded starts (-want +got):\n%s", diff)
	}
	caller.delivery = fileDelivery(t)
	second, err := u.Upload(t.Context(), "https://ui.example.com", "owner", "acc-1", 0, app.NamedFile{Name: "s.xlsx", Contents: strings.NewReader("export")})
	if err != nil || second.Mode != ingest.ModeTransactions {
		t.Fatalf("second Upload = %s, %v; want a transactions batch", second.Mode, err)
	}
	if len(l.txs) != 0 {
		t.Errorf("the second upload booked %d transactions dated on or before go-live, want none", len(l.txs))
	}
}
