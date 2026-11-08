package delivery

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// Provisional reads the instruments a sync request names as provisional
// (D-214): the treasury bond series the bound accounts post to under a
// provisional ISIN. One that doesn't parse fails with an error wrapping
// ErrMalformedHoldings naming the field.
func Provisional(req *connectorv1.SyncRequest) ([]ledger.Instrument, error) {
	out := make([]ledger.Instrument, 0, len(req.GetProvisionalInstruments()))
	for i, in := range req.GetProvisionalInstruments() {
		parsed, err := instrumentFrom(in)
		if err != nil {
			return nil, fmt.Errorf("%w: provisional_instruments[%d]: %w", ErrMalformedHoldings, i, err)
		}
		out = append(out, parsed)
	}
	return out, nil
}

// RegistryISINs writes the registry's ISINs of provisional instruments
// (D-214), in order: each instrument as the request named it, with its
// registry ISIN.
func RegistryISINs(isins []ingest.RegistryISIN) []*connectorv1.RegistryIsin {
	out := make([]*connectorv1.RegistryIsin, 0, len(isins))
	for _, r := range isins {
		out = append(out, &connectorv1.RegistryIsin{
			Instrument: &connectorv1.Instrument{Isin: r.Instrument.ISIN.String(), Symbol: r.Instrument.Symbol},
			Isin:       r.ISIN.String(),
		})
	}
	return out
}
