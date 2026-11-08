package ingest

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidRegistryISIN reports a registry ISIN a connector delivered that
// the core may not re-key with: its instrument isn't a treasury bond series
// under its provisional ISIN, or the ISIN isn't a Polish one other than a
// provisional ISIN, or two name one instrument (D-214).
var ErrInvalidRegistryISIN = errors.New("invalid registry ISIN")

// RegistryISIN is the registry's ISIN of an instrument the core holds under
// a provisional ISIN (D-214), as a connector delivers it in a sync: a
// treasury bond series booked before the Ministry of Finance's data listed
// its ISIN, or under the ISIN versions before D-214 derived from its series
// code. The core re-keys everything it stores under Instrument to Rekeyed.
type RegistryISIN struct {
	Instrument ledger.Instrument
	ISIN       ledger.ISIN
}

// Rekeyed returns the instrument r re-keys Instrument to: the registry's
// ISIN, with the same symbol, the series code.
func (r RegistryISIN) Rekeyed() ledger.Instrument {
	return ledger.Instrument{ISIN: r.ISIN, Symbol: r.Instrument.Symbol}
}

// Validate reports whether r may re-key its instrument: one whose ISIN is
// the provisional ISIN of the series its symbol names (pl.IsProvisional),
// to a Polish ISIN (pl.CheckBondISIN) that is no provisional one. Otherwise
// it returns an error wrapping ErrInvalidRegistryISIN.
func (r RegistryISIN) Validate() error {
	in := r.Instrument
	if err := in.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRegistryISIN, err)
	}
	if !pl.IsProvisional(in.ISIN.String(), in.Symbol) {
		return fmt.Errorf("%w: %s is no treasury bond series under its provisional ISIN", ErrInvalidRegistryISIN, in)
	}
	if r.ISIN.IsZero() {
		return fmt.Errorf("%w: %s: no ISIN", ErrInvalidRegistryISIN, in)
	}
	if err := pl.CheckBondISIN(r.ISIN.String()); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrInvalidRegistryISIN, in, err)
	}
	if _, provisional := pl.ProvisionalSeries(r.ISIN.String()); provisional {
		return fmt.Errorf("%w: %s: %s is a provisional ISIN", ErrInvalidRegistryISIN, in, r.ISIN)
	}
	return nil
}

// CheckRegistryISINs validates each of isins (RegistryISIN.Validate) and
// returns them in the order of their instruments, each once; two that name
// one instrument, with different ISINs or the same, fail with an error
// wrapping ErrInvalidRegistryISIN.
func CheckRegistryISINs(isins []RegistryISIN) ([]RegistryISIN, error) {
	out := slices.Clone(isins)
	for i := range out {
		if err := out[i].Validate(); err != nil {
			return nil, fmt.Errorf("registry ISIN %d: %w", i, err)
		}
	}
	slices.SortFunc(out, func(a, b RegistryISIN) int { return strings.Compare(a.Instrument.String(), b.Instrument.String()) })
	for i := 1; i < len(out); i++ {
		if out[i].Instrument == out[i-1].Instrument {
			return nil, fmt.Errorf("%w: %s is named twice", ErrInvalidRegistryISIN, out[i].Instrument)
		}
	}
	return out, nil
}

// ProvisionalInstruments returns the instruments txs post to a holding of
// whose ISIN is provisional (pl.IsProvisional, D-214): treasury bond series
// waiting for the registry's ISIN, each once, in ISIN order. A sync request
// names them, held or not, so a connector can deliver their registry ISIN.
func ProvisionalInstruments(txs []ledger.Transaction) []ledger.Instrument {
	seen := map[ledger.Instrument]bool{}
	var out []ledger.Instrument
	for i := range txs {
		for _, p := range txs[i].Postings {
			in := p.Book.Instrument
			if p.Book.Kind != ledger.BookHolding || seen[in] || !pl.IsProvisional(in.ISIN.String(), in.Symbol) {
				continue
			}
			seen[in] = true
			out = append(out, in)
		}
	}
	slices.SortFunc(out, func(a, b ledger.Instrument) int { return strings.Compare(a.String(), b.String()) })
	return out
}
