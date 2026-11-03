package connectrpc_test

import (
	"context"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// heldConnector reads what a sync says the bound accounts hold as a
// plugin would (delivery.Held).
type heldConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mu    sync.Mutex
	held  []ingest.HeldAccount
	terms []ingest.BondTerms
	err   error
}

func (h *heldConnector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.held, h.terms, h.err = delivery.Held(req.Msg)
	d := &connectorv1.Delivery{BatchId: req.Msg.GetBatchId(), Institution: "treasury-bonds"}
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: d, Watermark: "w"}), nil
}

// TestClientSyncSendsHoldings checks that what the core says the bound
// accounts hold (D-117) reaches a plugin intact: go-live, tax wrapper,
// payout setting (D-135), holdings, open lots and bond terms.
func TestClientSyncSendsHoldings(t *testing.T) {
	t.Parallel()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	isin, err := ledger.ParseISIN("PLEDO0536000")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	series, err := pl.ParseSeries("EDO0536")
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	in := ledger.Instrument{ISIN: isin, Symbol: "EDO0536"}
	forty := kernel.NewQuantity(kernel.NewDecimal(40, 0))
	held := []ingest.HeldAccount{
		{
			Account: "acc-bonds", GoLive: goLive(t), Wrapper: "taxable",
			Payout:   ingest.PayoutSetting{Payout: ingest.PayoutPaidOut, Since: goLive(t), Earlier: []ingest.PayoutPeriod{{Payout: ingest.PayoutCash}}},
			Holdings: []ledger.Holding{{Account: "acc-bonds", Instrument: in, Quantity: forty}},
			Lots: []ledger.Lot{{
				Instrument: in, Opened: goLive(t), Quantity: forty, Cost: kernel.NewMoney(kernel.NewDecimal(3996, 0), pln), Opening: true,
			}},
		},
		{Account: "acc-new", Wrapper: "ike"},
	}
	terms := []ingest.BondTerms{{Instrument: in, Terms: pl.Terms{
		Series: series, FirstPeriodRate: kernel.NewDecimal(68, -3), Margin: kernel.NewDecimal(2, -2),
		EarlyRedemptionFee: kernel.NewMoney(kernel.NewDecimal(2, 0), pln),
	}}}
	conn := &heldConnector{}
	call := &app.SyncCall{ID: "b-1", Mode: ingest.ModeTransactions, Held: held, BondTerms: terms}
	if _, err := newClient(t, conn).Sync(t.Context(), "bonds-main", call); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if conn.err != nil {
		t.Fatalf("delivery.Held: %v", conn.err)
	}
	if diff := cmp.Diff(held, conn.held); diff != "" {
		t.Errorf("held received (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(terms, conn.terms); diff != "" {
		t.Errorf("bond terms received (-want +got):\n%s", diff)
	}
}
