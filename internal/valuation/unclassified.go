package valuation

import (
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
)

// Unclassified returns the ISINs of the holdings priced at market among
// rows that pack can't classify in wrapper w and that owner, the owner's
// classes by ISIN, doesn't hold, in ISIN order without repeats: the
// instruments the owner is asked to classify for a yearly tax's exemptions
// (D-132, D-148).
func Unclassified(pack jurisdiction.Pack, w jurisdiction.Wrapper, rows []HoldingValuation, owner map[string]jurisdiction.ExemptionClass) []string {
	var out []string
	for i := range rows {
		v := &rows[i]
		if v.Basis != Market || !v.Valued() {
			continue
		}
		isin := v.Instrument.ISIN.String()
		asset := jurisdiction.Asset{Kind: jurisdiction.AssetListed, Instrument: isin, Currency: v.Value.Currency()}
		if _, known := pack.ExemptionClass(w, &asset); known {
			continue
		}
		if _, set := owner[isin]; set || slices.Contains(out, isin) {
			continue
		}
		out = append(out, isin)
	}
	slices.Sort(out)
	return out
}
