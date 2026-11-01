package connectrpc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// tickerMapping is a mapping of a buy of TESTA.PL by broker account
// 0000000 and a dividend of it in account acc-1.
func tickerMapping(t *testing.T) ingest.Mapping {
	t.Helper()
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	day, err := kernel.ParseDate("2026-01-05")
	if err != nil {
		t.Fatal(err)
	}
	buy := ingest.TickerEntry{
		Symbol: "TESTA.PL", Kind: ledger.KindBuy, Quantity: kernel.NewQuantity(kernel.NewDecimal(4, 0)),
		Amount: kernel.NewMoney(kernel.NewDecimal(400, 0), pln), Description: "buy TESTA.PL", Ref: "Cash Operations#7",
		Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb", ExternalID: "9000000002"}, Date: day},
	}
	dividend := ingest.TickerEntry{
		Symbol: "TESTA.PL", Kind: ledger.KindDividend, Amount: kernel.NewMoney(kernel.NewDecimal(10, 0), pln),
		Charge: kernel.NewMoney(kernel.NewDecimal(19, -1), pln), Ref: "Cash Operations#8",
		Header: ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "acc-1", ExternalID: "9000000004"}, Account: "acc-1", Date: day},
	}
	return ingest.Mapping{
		BrokerAccounts: []ingest.BrokerAccountID{"0000000"},
		Tickers:        []ingest.Ticker{{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL", Name: "Test Alpha SA"}},
		TickerEntries:  []ingest.TickerEntry{dividend},
		ByBrokerAccount: []ingest.BrokerEntries{{
			Account: "0000000", TickerEntries: []ingest.TickerEntry{buy},
		}},
	}
}

// TestTickersRoundTrip checks that the tickers and ticker entries a
// connector writes (delivery.FromMapping) reach the core unchanged.
func TestTickersRoundTrip(t *testing.T) {
	t.Parallel()
	want := tickerMapping(t)
	d, err := delivery.FromMapping("b-1", "xtb", &want)
	if err != nil {
		t.Fatalf("FromMapping: %v", err)
	}
	got, err := deliveryFrom(d, "b-1", "")
	if err != nil {
		t.Fatalf("deliveryFrom: %v", err)
	}
	if diff := cmp.Diff(want, got.Mapping, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("mapping (-want +got):\n%s", diff)
	}
}

// TestTickersRefused checks that a delivery whose ticker transaction names
// no ticker of it, or books no valid transaction, is refused.
func TestTickersRefused(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(d *connectorv1.Delivery)
	}{
		{name: "unknown symbol", change: func(d *connectorv1.Delivery) { d.TickerTransactions[0].Symbol = "OTHER.PL" }},
		{name: "invalid ticker", change: func(d *connectorv1.Delivery) { d.Tickers[0].Country = "pl" }},
		{name: "buy for nothing", change: func(d *connectorv1.Delivery) { d.TickerTransactions[0].Amount.Amount = "-1" }},
		{name: "both accounts", change: func(d *connectorv1.Delivery) { d.TickerTransactions[0].AccountId = "acc-1" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := tickerMapping(t)
			d, err := delivery.FromMapping("b-1", "xtb", &m)
			if err != nil {
				t.Fatalf("FromMapping: %v", err)
			}
			d.TickerTransactions = []*connectorv1.TickerTransaction{d.GetTickerTransactions()[1]}
			tt.change(d)
			if _, err := deliveryFrom(d, "b-1", ""); err == nil {
				t.Error("deliveryFrom accepted the delivery")
			}
		})
	}
}
