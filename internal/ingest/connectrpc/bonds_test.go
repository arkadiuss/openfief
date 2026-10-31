package connectrpc_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// errRefusedLot is what the fake bond connector answers a lot of no bonds.
var errRefusedLot = errors.New("lot 1: invalid bond lot: quantity 0")

// refuse is the Connect error with code for err, as a connector answers.
func refuse(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// bondConnector records the lots it was handed and delivers the terms of
// their first series; it refuses a lot of no bonds.
type bondConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	mu   sync.Mutex
	auth string
	lots []*connectorv1.BondLot
	req  *connectorv1.ImportBondLotsRequest
}

func (b *bondConnector) ImportBondLots(_ context.Context, req *connect.Request[connectorv1.ImportBondLotsRequest]) (*connect.Response[connectorv1.ImportBondLotsResponse], error) {
	b.mu.Lock()
	b.auth, b.lots, b.req = req.Header().Get("Authorization"), req.Msg.GetLots(), req.Msg
	b.mu.Unlock()
	if req.Msg.GetLots()[0].GetQuantity() == "0" {
		return nil, refuse(connect.CodeInvalidArgument, errRefusedLot)
	}
	return connect.NewResponse(&connectorv1.ImportBondLotsResponse{Delivery: &connectorv1.Delivery{
		BatchId: req.Msg.GetBatchId(), Institution: "treasury-bonds",
		BondTerms: []*connectorv1.BondTerms{{
			Instrument:      &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"},
			FirstPeriodRate: "0.06", Margin: "0.02", EarlyRedemptionFee: &connectorv1.Money{Amount: "2", Currency: "PLN"},
		}},
	}}), nil
}

// lotOf returns a synthetic lot of quantity EDO0536 bonds bought on
// 2026-05-04.
func lotOf(t *testing.T, quantity int64) ingest.BondLot {
	t.Helper()
	series, err := pl.ParseSeries("EDO0536")
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	purchased, err := kernel.ParseDate("2026-05-04")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	return ingest.BondLot{
		Terms: pl.Terms{
			Series: series, FirstPeriodRate: kernel.NewDecimal(68, -3), Margin: kernel.NewDecimal(2, -2),
			EarlyRedemptionFee: kernel.NewMoney(kernel.NewDecimal(200, -2), pl.Nominal().Currency()),
		},
		Quantity: kernel.NewQuantity(kernel.NewDecimal(quantity, 0)), Purchased: purchased,
	}
}

func TestClientImportBondLots(t *testing.T) {
	t.Parallel()
	conn := &bondConnector{}
	d, err := newClient(t, conn).ImportBondLots(t.Context(), "bonds-main", "b-1", &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{lotOf(t, 40)}})
	if err != nil {
		t.Fatalf("ImportBondLots: %v", err)
	}
	want := []*connectorv1.BondLot{{
		Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04", FirstPeriodRate: "0.068", Margin: "0.02",
		EarlyRedemptionFee: &connectorv1.Money{Amount: "2", Currency: "PLN"}, Purchase: connectorv1.BondPurchase_BOND_PURCHASE_CASH,
	}}
	if diff := cmp.Diff(want, conn.lots, protocmp.Transform()); diff != "" {
		t.Errorf("lots sent (-want +got):\n%s", diff)
	}
	if conn.req.GetGoLive() != "2026-05-15" {
		t.Errorf("go_live sent = %q, want the account's go-live 2026-05-15", conn.req.GetGoLive())
	}
	if conn.auth != "Bearer synthetic-core-bonds-main" {
		t.Errorf("Authorization = %q, want the instance's core token", conn.auth)
	}
	if d.Institution != "treasury-bonds" || len(d.BondTerms) != 1 || d.BondTerms[0].Instrument.Symbol != "EDO0536" {
		t.Errorf("delivery = %+v, want treasury-bonds with EDO0536's terms", d)
	}
}

func TestClientImportBondLotsRefusals(t *testing.T) {
	t.Parallel()
	_, err := newClient(t, &bondConnector{}).ImportBondLots(t.Context(), "bonds-main", "b-1", &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{lotOf(t, 0)}})
	if !errors.Is(err, app.ErrEntryRefused) || !strings.Contains(err.Error(), errRefusedLot.Error()) {
		t.Errorf("a refused lot: %v; want ErrEntryRefused with the connector's reason", err)
	}
	_, err = newClient(t, &failingConnector{}).ImportBondLots(t.Context(), "bonds-main", "b-1", &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{lotOf(t, 40)}})
	if !errors.Is(err, app.ErrNoEntries) {
		t.Errorf("a connector without bond lots: %v; want ErrNoEntries", err)
	}
	if _, err = newClient(t, &bondConnector{}).ImportBondLots(t.Context(), "broken", "b-1", &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{lotOf(t, 40)}}); !errors.Is(err, errNoToken) || !errors.Is(err, app.ErrConnectorUnavailable) {
		t.Errorf("no core token: %v; want the token error as ErrConnectorUnavailable, and no call", err)
	}
	if _, err = newClient(t, &bondConnector{}).ImportBondLots(t.Context(), "nowhere", "b-1", &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{lotOf(t, 40)}}); !errors.Is(err, app.ErrConnectorUnavailable) {
		t.Errorf("an instance without an address: %v; want ErrConnectorUnavailable", err)
	}
	if _, err = newClient(t, &unavailableConnector{}).ImportBondLots(t.Context(), "bonds-main", "b-1", &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{lotOf(t, 40)}}); !errors.Is(err, app.ErrConnectorUnavailable) {
		t.Errorf("a connector answering UNAVAILABLE: %v; want ErrConnectorUnavailable", err)
	}
}

// errBusy is what unavailableConnector answers.
var errBusy = errors.New("synthetic: busy")

// unavailableConnector answers every bond lot call UNAVAILABLE.
type unavailableConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler
}

func (unavailableConnector) ImportBondLots(context.Context, *connect.Request[connectorv1.ImportBondLotsRequest]) (*connect.Response[connectorv1.ImportBondLotsResponse], error) {
	return nil, refuse(connect.CodeUnavailable, errBusy)
}

// goLive is the synthetic account's go-live date.
func goLive(t *testing.T) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate("2026-05-15")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	return d
}

func TestClientDeletesBondLots(t *testing.T) {
	t.Parallel()
	conn := &bondConnector{}
	l := lotOf(t, 40)
	exchanged := l
	exchanged.Purchase = ingest.PurchaseExchange
	e := &ingest.BondLotEntries{GoLive: goLive(t), Lots: []ingest.BondLot{exchanged}, Deleted: []ingest.BondLotKey{{Series: l.Series(), Purchased: goLive(t)}}}
	if _, err := newClient(t, conn).ImportBondLots(t.Context(), "bonds-main", "b-1", e); err != nil {
		t.Fatalf("ImportBondLots: %v", err)
	}
	want := []*connectorv1.BondLotKey{{Series: "EDO0536", PurchaseDate: "2026-05-15"}}
	if diff := cmp.Diff(want, conn.req.GetDeleted(), protocmp.Transform()); diff != "" {
		t.Errorf("deleted lots sent (-want +got):\n%s", diff)
	}
	if got := conn.req.GetLots()[0].GetPurchase(); got != connectorv1.BondPurchase_BOND_PURCHASE_EXCHANGE {
		t.Errorf("purchase sent = %s, want exchange", got)
	}
}
