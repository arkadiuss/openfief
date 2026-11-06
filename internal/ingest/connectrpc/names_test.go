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

// TestInstrumentNamesRoundTrip checks that a connector's instrument names
// (D-197) travel from its mapping, each ISIN once, to the core's delivery,
// and that a name with a malformed ISIN refuses the delivery.
func TestInstrumentNamesRoundTrip(t *testing.T) {
	t.Parallel()
	isin, err := ledger.ParseISIN(testISIN)
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	m := ingest.Mapping{Names: []ledger.InstrumentName{{ISIN: isin, Name: "SYNTHETIC ETF"}, {ISIN: isin, Name: "Again"}}}
	d, err := delivery.FromMapping("b-1", "ibkr", &m)
	if err != nil {
		t.Fatalf("FromMapping: %v", err)
	}
	got, err := deliveryFrom(d, "b-1", "")
	if err != nil {
		t.Fatalf("deliveryFrom: %v", err)
	}
	want := []ledger.InstrumentName{{ISIN: isin, Name: "SYNTHETIC ETF"}}
	if diff := cmp.Diff(want, got.InstrumentNames, cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
		t.Errorf("names (-want +got):\n%s", diff)
	}
	d.InstrumentNames = append(d.InstrumentNames, &connectorv1.InstrumentName{Isin: "XX123", Name: "Bad"})
	if _, err := deliveryFrom(d, "b-1", ""); !errors.Is(err, app.ErrInvalidDelivery) {
		t.Errorf("a malformed ISIN: error = %v, want ErrInvalidDelivery", err)
	}
}
