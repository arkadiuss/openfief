package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// CrossCheckStore stores cross-check prices and reads them back with the
// prices they are compared with. *postgres.Store implements it.
type CrossCheckStore interface {
	// SaveCrossCheckPrices upserts cross-check prices in one transaction,
	// keyed by instrument, date and source. No holding is valued from them,
	// so they enqueue no revaluation.
	SaveCrossCheckPrices(ctx context.Context, prices []valuation.CrossCheckPrice) error
	// CrossCheckedPrices returns every stored cross-check price, and every
	// stored price of an instrument and date that has one.
	CrossCheckedPrices(ctx context.Context) ([]valuation.Price, []valuation.CrossCheckPrice, error)
}

// CrossChecks records institutions' own prices as cross-check prices and
// compares them with the market-data connector's prices (D-22): they are
// readings beside the prices, never valuation inputs.
type CrossChecks struct {
	store     CrossCheckStore
	threshold kernel.Decimal
}

// NewCrossChecks returns the cross-check use cases over store, flagging a
// price check whose relative difference exceeds threshold. A zero threshold
// takes valuation.DefaultPriceCheckThreshold; any other must be a fraction
// greater than 0 and less than 1.
func NewCrossChecks(store CrossCheckStore, threshold kernel.Decimal) (*CrossChecks, error) {
	if threshold.IsZero() {
		threshold = valuation.DefaultPriceCheckThreshold()
	}
	t, err := valuation.CheckThreshold(threshold)
	if err != nil {
		return nil, fmt.Errorf("cross checks: %w", err)
	}
	return &CrossChecks{store: store, threshold: t}, nil
}

// Threshold returns the relative difference beyond which a price check is
// flagged.
func (c *CrossChecks) Threshold() kernel.Decimal { return c.threshold }

// RecordCrossCheckPrices stores cross-check prices. Every one must be valid,
// or nothing is stored. Recording one again for the same instrument, date
// and source replaces it.
func (c *CrossChecks) RecordCrossCheckPrices(ctx context.Context, prices []valuation.CrossCheckPrice) error {
	for i := range prices {
		if err := prices[i].Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
	}
	if err := c.store.SaveCrossCheckPrices(ctx, prices); err != nil {
		return fmt.Errorf("recording cross-check prices: %w", err)
	}
	return nil
}

// PriceChecks compares every stored cross-check price with the stored price
// of the same instrument, date and currency, in date, instrument and source
// order (valuation.CheckPrices).
func (c *CrossChecks) PriceChecks(ctx context.Context) ([]valuation.PriceCheck, error) {
	prices, crossChecks, err := c.store.CrossCheckedPrices(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading cross-checked prices: %w", err)
	}
	checks, err := valuation.CheckPrices(prices, crossChecks, c.threshold)
	if err != nil {
		return nil, fmt.Errorf("checking prices: %w", err)
	}
	return checks, nil
}
