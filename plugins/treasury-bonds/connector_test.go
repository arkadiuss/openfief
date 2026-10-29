package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// coreToken is what the fake verifier takes for the core's token.
const coreToken = "synthetic-core-token"

type fakeVerifier struct{}

func (fakeVerifier) VerifyCore(_ context.Context, token string) error {
	if token != coreToken {
		return openfief.ErrNotCore
	}
	return nil
}

// coreClient is the generated client, as a struct so helpers return no
// interface.
type coreClient struct {
	connectorv1connect.ConnectorServiceClient
}

// served serves a connector for account acc-bonds, whose today is
// 2026-10-29, with the SDK as the plugin does, and returns a client that
// calls it as the core.
func served(t *testing.T) coreClient {
	t.Helper()
	today, err := kernel.ParseDate("2026-10-29")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	c := &connector{mapper: treasurybonds.Mapper{Account: "acc-bonds"}, today: func() kernel.Date { return today }}
	server := openfief.NewConnectorServer("", c, fakeVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	return coreClient{connectorv1connect.NewConnectorServiceClient(&http.Client{Transport: &http.Transport{Protocols: &p}}, srv.URL)}
}

// asCore returns msg as a request carrying the core's token.
func asCore[T any](msg *T) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+coreToken)
	return req
}

// edoLot is a synthetic lot of 40 EDO0536 bonds bought on 2026-05-04.
func edoLot() *connectorv1.BondLot {
	return &connectorv1.BondLot{
		Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04", FirstPeriodRate: "0.06", Margin: "0.02",
		EarlyRedemptionFee: &connectorv1.Money{Amount: "2.00", Currency: "PLN"},
	}
}

func TestImportBondLots(t *testing.T) {
	t.Parallel()
	client := served(t)
	resp, err := client.ImportBondLots(t.Context(), asCore(&connectorv1.ImportBondLotsRequest{BatchId: "b-1", Lots: []*connectorv1.BondLot{edoLot()}}))
	if err != nil {
		t.Fatalf("ImportBondLots: %v", err)
	}
	edo := &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"}
	pln := func(amount string) *connectorv1.Money { return &connectorv1.Money{Amount: amount, Currency: "PLN"} }
	cash := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "PLN"}
	key := func(id string) *connectorv1.TransactionKey {
		return &connectorv1.TransactionKey{Source: "treasury-bonds:acc-bonds", ExternalId: "lot:EDO0536:2026-05-04:" + id}
	}
	want := &connectorv1.Delivery{
		BatchId: "b-1", Institution: "treasury-bonds",
		Transactions: []*connectorv1.Transaction{
			{
				Key: key("deposit"), AccountId: "acc-bonds", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-05-04",
				Description: "Cash for treasury bonds EDO0536",
				Postings: []*connectorv1.Posting{
					{Book: cash, Amount: pln("4000")},
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "PLN"}, Amount: pln("-4000")},
				},
			},
			{
				Key: key("buy"), AccountId: "acc-bonds", Kind: connectorv1.TransactionKind_TRANSACTION_KIND_BUY, Date: "2026-05-04",
				Description: "Treasury bonds EDO0536",
				Postings: []*connectorv1.Posting{
					{Book: &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_HOLDING, Instrument: edo}, Amount: pln("4000"), Quantity: "40"},
					{Book: cash, Amount: pln("-4000")},
				},
			},
		},
		BondTerms: []*connectorv1.BondTerms{{Instrument: edo, FirstPeriodRate: "0.06", Margin: "0.02", EarlyRedemptionFee: pln("2")}},
	}
	if diff := cmp.Diff(want, resp.Msg.GetDelivery(), protocmp.Transform()); diff != "" {
		t.Errorf("ImportBondLots delivery (-want +got):\n%s", diff)
	}
}

func TestImportBondLotsRefuses(t *testing.T) {
	t.Parallel()
	client := served(t)
	with := func(change func(l *connectorv1.BondLot)) []*connectorv1.BondLot {
		l := edoLot()
		change(l)
		return []*connectorv1.BondLot{l}
	}
	tests := []struct {
		name string
		lots []*connectorv1.BondLot
	}{
		{name: "an unknown bond type", lots: with(func(l *connectorv1.BondLot) { l.Series = "XYZ0536" })},
		{name: "a purchase date of another series", lots: with(func(l *connectorv1.BondLot) { l.PurchaseDate = "2026-06-04" })},
		{name: "bought after today", lots: with(func(l *connectorv1.BondLot) { l.Series, l.PurchaseDate = "EDO1136", "2026-11-04" })},
		{name: "a fraction of a bond", lots: with(func(l *connectorv1.BondLot) { l.Quantity = "2.5" })},
		{name: "no fee", lots: with(func(l *connectorv1.BondLot) { l.EarlyRedemptionFee = nil })},
		{name: "a rate that isn't a number", lots: with(func(l *connectorv1.BondLot) { l.FirstPeriodRate = "6%" })},
		{name: "one series and day twice", lots: []*connectorv1.BondLot{edoLot(), edoLot()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.ImportBondLots(t.Context(), asCore(&connectorv1.ImportBondLotsRequest{BatchId: "b-1", Lots: tt.lots}))
			if connect.CodeOf(err) != connect.CodeInvalidArgument {
				t.Errorf("ImportBondLots = %v, want InvalidArgument", err)
			}
		})
	}
	if _, err := client.ImportBondLots(t.Context(), connect.NewRequest(&connectorv1.ImportBondLotsRequest{BatchId: "b-2"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("ImportBondLots without the core's token: %v, want Unauthenticated", err)
	}
}

func TestSyncDeliversNothing(t *testing.T) {
	t.Parallel()
	resp, err := served(t).Sync(t.Context(), asCore(&connectorv1.SyncRequest{BatchId: "b-1", Mode: connectorv1.Mode_MODE_OPENING, Watermark: "w"}))
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	want := &connectorv1.SyncResponse{Delivery: &connectorv1.Delivery{BatchId: "b-1", Institution: "treasury-bonds"}, Watermark: "w"}
	if diff := cmp.Diff(want, resp.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("Sync (-want +got):\n%s", diff)
	}
}
