package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// InstrumentNames reads instruments' display names (D-197);
// *ledger/app.Names implements it.
type InstrumentNames interface {
	// Current returns the display name of each named instrument.
	Current(ctx context.Context) (map[ledger.ISIN]string, error)
}

// WithInstrumentNames reads instruments' display names from n; without it
// no instrument has a name.
func WithInstrumentNames(n InstrumentNames) Option {
	return func(v *Views) { v.names = n }
}

// InstrumentNames returns the display name of each named instrument, by
// ISIN (D-197). A name is not tied to an account: the read API serves it
// only with an instrument an answer already names, so no caller learns of
// an instrument it may not see.
func (v *Views) InstrumentNames(ctx context.Context) (map[string]string, error) {
	out := map[string]string{}
	if v.names == nil {
		return out, nil
	}
	names, err := v.names.Current(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading instrument names: %w", err)
	}
	for isin, name := range names {
		out[isin.String()] = name
	}
	return out, nil
}
