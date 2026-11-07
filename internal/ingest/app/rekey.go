package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// InstrumentRekeyer moves what one context stores under an instrument to
// another with the same symbol and the registry's ISIN (D-214), such as a
// treasury bond series booked under its provisional ISIN. It is idempotent:
// with nothing left under from, it changes nothing, so a re-key that fails
// part-way is simply run again by the next sync.
type InstrumentRekeyer interface {
	RekeyInstrument(ctx context.Context, from, to ledger.Instrument) error
}

// ProvisionalSource says which instruments the accounts bound to a
// connector instance post to under a provisional ISIN (D-214);
// HoldingsScope implements it.
type ProvisionalSource interface {
	ProvisionalInstruments(ctx context.Context, accounts []ledger.AccountID) ([]ledger.Instrument, error)
}

// ErrNoRekeyer reports registry ISINs delivered to a Connectors that has
// nothing to re-key instruments with.
var ErrNoRekeyer = errors.New("no way to re-key instruments")

// ProvisionalInstruments returns every instrument accounts post to, held or
// not, whose ISIN is provisional (ingest.ProvisionalInstruments), each once
// in ISIN order; none without Postings.
func (h *HoldingsScope) ProvisionalInstruments(ctx context.Context, accounts []ledger.AccountID) ([]ledger.Instrument, error) {
	if h.Postings == nil {
		return nil, nil
	}
	var txs []ledger.Transaction
	for _, id := range accounts {
		current, err := h.Postings.CurrentTransactions(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("transactions of %s: %w", id, err)
		}
		txs = append(txs, current...)
	}
	return ingest.ProvisionalInstruments(txs), nil
}

// rekey re-keys the instrument of each of isins, which must be valid and
// name each instrument once (ingest.CheckRegistryISINs), in instrument
// order, to the one with its registry ISIN (D-214):
// each Rekeyer in turn moves what its context stores, the ledger's last,
// one instrument at a time. Each step is idempotent, and the ledger's is
// what takes the provisional instrument off the next sync request, so a
// failure anywhere leaves the instrument provisional and the next sync
// delivers its ISIN and re-keys it again.
func (c *Connectors) rekey(ctx context.Context, isins []ingest.RegistryISIN) error {
	if len(isins) == 0 {
		return nil
	}
	if len(c.Rekeyers) == 0 {
		return ErrNoRekeyer
	}
	ordered, err := ingest.CheckRegistryISINs(isins)
	if err != nil {
		return fmt.Errorf("refused: %w", err)
	}
	for _, r := range ordered {
		for _, k := range c.Rekeyers {
			if err := k.RekeyInstrument(ctx, r.Instrument, r.Rekeyed()); err != nil {
				return fmt.Errorf("re-keying %s to %s: %w", r.Instrument, r.ISIN, err)
			}
		}
	}
	return nil
}
