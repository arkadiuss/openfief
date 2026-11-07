package connectrpc

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// TestRegistryISINsRoundTrip checks that the registry ISINs a connector
// delivers for provisional instruments (D-214) travel from the plugin's
// delivery to the core's, and that one with a malformed ISIN or
// instrument refuses the delivery; whether the core may re-key with them
// is the use case's to decide.
func TestRegistryISINsRoundTrip(t *testing.T) {
	t.Parallel()
	provisional, err := ledger.ParseISIN("PLEDO0536000")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	registry, err := ledger.ParseISIN("PL9999000018")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	want := []ingest.RegistryISIN{{Instrument: ledger.Instrument{ISIN: provisional, Symbol: "EDO0536"}, ISIN: registry}}
	d := &connectorv1.Delivery{BatchId: "b-1", Institution: "treasury-bonds", RegistryIsins: delivery.RegistryISINs(want)}
	got, err := deliveryFrom(d, "b-1", "")
	if err != nil {
		t.Fatalf("deliveryFrom: %v", err)
	}
	if diff := cmp.Diff(want, got.RegistryISINs, cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
		t.Errorf("registry ISINs (-want +got):\n%s", diff)
	}
	if n := deliveryItems(d); n != 1 {
		t.Errorf("deliveryItems = %d, want the registry ISIN counted", n)
	}
	for _, bad := range []*connectorv1.RegistryIsin{
		{Instrument: &connectorv1.Instrument{Isin: "PLEDO0536000", Symbol: "EDO0536"}, Isin: "XX123"},
		{Instrument: &connectorv1.Instrument{Isin: "XX123", Symbol: "EDO0536"}, Isin: "PL9999000018"},
		{Instrument: &connectorv1.Instrument{Isin: "PLEDO0536000"}, Isin: "PL9999000018"},
	} {
		d.RegistryIsins = []*connectorv1.RegistryIsin{bad}
		if _, err := deliveryFrom(d, "b-1", ""); !errors.Is(err, app.ErrInvalidDelivery) {
			t.Errorf("deliveryFrom(%v) error = %v, want ErrInvalidDelivery", bad, err)
		}
	}
}

// TestSyncRequestNamesProvisional checks that a sync request names the
// instruments the call says are provisional (D-214), as a plugin reads
// them back (delivery.Provisional).
func TestSyncRequestNamesProvisional(t *testing.T) {
	t.Parallel()
	isin, err := ledger.ParseISIN("PLEDO0536000")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	want := []ledger.Instrument{{ISIN: isin, Symbol: "EDO0536"}}
	req := &connectorv1.SyncRequest{ProvisionalInstruments: instrumentsTo(want)}
	got, err := delivery.Provisional(req)
	if err != nil {
		t.Fatalf("Provisional: %v", err)
	}
	if diff := cmp.Diff(want, got, cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
		t.Errorf("Provisional (-want +got):\n%s", diff)
	}
	req.ProvisionalInstruments = append(req.ProvisionalInstruments, &connectorv1.Instrument{Isin: "XX123", Symbol: "EDO0536"})
	if _, err := delivery.Provisional(req); !errors.Is(err, delivery.ErrMalformedHoldings) {
		t.Errorf("Provisional with a malformed ISIN: error = %v, want ErrMalformedHoldings", err)
	}
}
