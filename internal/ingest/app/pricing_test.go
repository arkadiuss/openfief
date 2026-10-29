package app_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// heldBook holds fixed holdings per account, or fails with err.
type heldBook struct {
	holdings map[ledger.AccountID][]ledger.Instrument
	err      error
	asked    kernel.Date
}

func (h *heldBook) Accounts(context.Context) ([]ledger.Account, error) {
	if h.err != nil {
		return nil, h.err
	}
	return []ledger.Account{{ID: "acc-1"}, {ID: "acc-2"}}, nil
}

func (h *heldBook) HoldingsOn(_ context.Context, account ledger.AccountID, on kernel.Date) ([]ledger.Holding, error) {
	h.asked = on
	out := make([]ledger.Holding, 0, len(h.holdings[account]))
	for _, in := range h.holdings[account] {
		out = append(out, ledger.Holding{Account: account, Instrument: in, Date: on, Quantity: kernel.NewQuantity(kernel.NewDecimal(1, 0))})
	}
	return out, nil
}

// bondISINs are the ISINs with stored bond terms.
type bondISINs map[string]bool

func (b bondISINs) BondTerms(_ context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	return pl.Terms{}, b[isin.String()], nil
}

// modelTargets is a model portfolio's targets, or err.
type modelTargets struct {
	instruments []ledger.Instrument
	indices     []string
	err         error
}

func (m modelTargets) MarketDataTargets(context.Context) ([]ledger.Instrument, []string, error) {
	return m.instruments, m.indices, m.err
}

// pricingInstrument is a synthetic instrument.
func pricingInstrument(t *testing.T, isin, symbol string) ledger.Instrument {
	t.Helper()
	i, err := ledger.ParseISIN(isin)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", isin, err)
	}
	return ledger.Instrument{ISIN: i, Symbol: symbol}
}

func TestPricingScope(t *testing.T) {
	t.Parallel()
	etf := pricingInstrument(t, "IETEST000026", "TSTB")
	stock := pricingInstrument(t, "XX00000SYNT3", "SYNT")
	bond := pricingInstrument(t, "PL0000000006", "EDO0536")
	tests := []struct {
		name    string
		model   app.ModelTargets
		holdErr error
		want    ingest.MarketDataScope
		wantErr error
	}{
		{name: "held instruments, each once, bonds left out", want: ingest.MarketDataScope{Instruments: []ledger.Instrument{etf, stock}}},
		{
			name:  "with the model portfolio's targets",
			model: modelTargets{instruments: []ledger.Instrument{stock}, indices: []string{"FAKE20"}},
			want:  ingest.MarketDataScope{Instruments: []ledger.Instrument{etf, stock}, Indices: []string{"FAKE20"}},
		},
		{name: "holdings fail", holdErr: errStore, wantErr: errStore},
		{name: "the model fails", model: modelTargets{err: errStore}, wantErr: errStore},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			book := &heldBook{err: tt.holdErr, holdings: map[ledger.AccountID][]ledger.Instrument{
				"acc-1": {stock, bond}, "acc-2": {etf, stock},
			}}
			p := &app.PricingScope{Holdings: book, Bonds: bondISINs{"PL0000000006": true}, Model: tt.model, Today: func() kernel.Date { return march2(t) }}
			got, err := p.MarketDataScope(t.Context())
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("MarketDataScope error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got, cmp.Comparer(func(x, y ledger.ISIN) bool { return x.Equal(y) })); diff != "" {
				t.Errorf("MarketDataScope (-want +got):\n%s", diff)
			}
			if tt.wantErr == nil && !book.asked.Equal(march2(t)) {
				t.Errorf("holdings asked on %s, want today", book.asked)
			}
		})
	}
}
