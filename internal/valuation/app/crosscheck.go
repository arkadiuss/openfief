package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrUnknownPriceCheck reports an acknowledgement of a price check that
// doesn't exist: no cross-check price of that instrument, date and source
// is compared with a price.
var ErrUnknownPriceCheck = errors.New("no such price check")

// CrossCheckStore stores cross-check prices, the owner's price check
// thresholds and acknowledgements, and reads them back with the prices the
// cross-check prices are compared with. *postgres.Store implements it.
type CrossCheckStore interface {
	// SaveCrossCheckPrices upserts cross-check prices in one transaction,
	// keyed by instrument, date and source. No holding is valued from them,
	// so they enqueue no revaluation.
	SaveCrossCheckPrices(ctx context.Context, prices []valuation.CrossCheckPrice) error
	// CrossCheckedPrices returns the winning price of each date of every
	// instrument that has a cross-check price or more than one source's
	// price, and every cross-check price, a lower-ranked source's prices
	// included (D-94).
	CrossCheckedPrices(ctx context.Context) ([]valuation.Price, []valuation.CrossCheckPrice, error)
	// PriceCheckThresholds returns the owner's threshold of each instrument
	// that has one (D-110).
	PriceCheckThresholds(ctx context.Context) (valuation.Thresholds, error)
	// SavePriceCheckThreshold sets instrument's threshold, or removes it
	// when threshold is nil.
	SavePriceCheckThreshold(ctx context.Context, instrument string, threshold *kernel.Decimal) error
	// PriceCheckAcknowledgements returns every acknowledgement recorded.
	PriceCheckAcknowledgements(ctx context.Context) ([]valuation.PriceCheckAcknowledgement, error)
	// AppendPriceCheckAcknowledgement records an acknowledgement; nothing
	// recorded is ever changed or deleted (D-112).
	AppendPriceCheckAcknowledgement(ctx context.Context, ack *valuation.PriceCheckAcknowledgement) error
}

// CrossChecks records institutions' own prices as cross-check prices and
// compares them with the stored prices (D-22, D-110..D-112): they are
// readings beside the prices, never valuation inputs. The owner sets a
// threshold for a single instrument and acknowledges flagged checks.
type CrossChecks struct {
	store  CrossCheckStore
	owners Owners
}

// NewCrossChecks returns the cross-check use cases over store, with the
// owner's sessions checked by owners; nil owners serves no owner, as the
// connector commands need none.
func NewCrossChecks(store CrossCheckStore, owners Owners) *CrossChecks {
	return &CrossChecks{store: store, owners: owners}
}

// Threshold returns the relative difference beyond which a price check is
// flagged unless the owner set another for its instrument: one percent
// (D-110).
func (c *CrossChecks) Threshold() kernel.Decimal { return valuation.DefaultPriceCheckThreshold() }

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

// PriceChecks compares every stored cross-check price with the latest
// stored winning price of its instrument on or before its date, at the
// instrument's threshold, with the owner's acknowledgements, in date,
// instrument and source order (valuation.CheckPrices).
func (c *CrossChecks) PriceChecks(ctx context.Context) ([]valuation.PriceCheck, error) {
	prices, crossChecks, err := c.store.CrossCheckedPrices(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading cross-checked prices: %w", err)
	}
	thresholds, err := c.store.PriceCheckThresholds(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading price check thresholds: %w", err)
	}
	acks, err := c.store.PriceCheckAcknowledgements(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading price check acknowledgements: %w", err)
	}
	checks, err := valuation.CheckPrices(prices, crossChecks, thresholds, acks)
	if err != nil {
		return nil, fmt.Errorf("checking prices: %w", err)
	}
	return checks, nil
}

// SetThreshold sets, for the owner, the threshold of the price checks of
// instrument (D-110), written as Price writes it; nil puts the instrument
// back on the default. An empty instrument or a threshold outside (0, 1)
// is an error wrapping ErrInvalidInput, and nothing changes.
func (c *CrossChecks) SetThreshold(ctx context.Context, origin, sessionToken, instrument string, threshold *kernel.Decimal) error {
	if err := c.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if instrument == "" {
		return fmt.Errorf("%w: a threshold needs an instrument", ErrInvalidInput)
	}
	if threshold != nil {
		if _, err := valuation.CheckThreshold(*threshold); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
	}
	if err := c.store.SavePriceCheckThreshold(ctx, instrument, threshold); err != nil {
		return fmt.Errorf("setting the price check threshold of %s: %w", instrument, err)
	}
	return nil
}

// Acknowledge records, for the owner, the acknowledgement of the price
// check of instrument's cross-check price of day from source, which must
// be flagged now (D-112), and returns the check acknowledged. A check that
// doesn't exist is ErrUnknownPriceCheck; one that isn't flagged wraps
// ErrInvalidInput and valuation.ErrNotFlagged. Acknowledging it again
// records another acknowledgement and changes nothing else.
func (c *CrossChecks) Acknowledge(ctx context.Context, origin, sessionToken, instrument string, day kernel.Date, source string) (valuation.PriceCheck, error) {
	if err := c.checkOwner(ctx, origin, sessionToken); err != nil {
		return valuation.PriceCheck{}, err
	}
	checks, err := c.PriceChecks(ctx)
	if err != nil {
		return valuation.PriceCheck{}, err
	}
	for i := range checks {
		check := &checks[i]
		if check.CrossCheck.Instrument != instrument || !check.CrossCheck.Date.Equal(day) || check.CrossCheck.Source != source {
			continue
		}
		ack, err := valuation.Acknowledge(check)
		if err != nil {
			return valuation.PriceCheck{}, fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
		if err := c.store.AppendPriceCheckAcknowledgement(ctx, &ack); err != nil {
			return valuation.PriceCheck{}, fmt.Errorf("acknowledging the price check: %w", err)
		}
		check.Acknowledged = true
		return *check, nil
	}
	return valuation.PriceCheck{}, fmt.Errorf("%w: %s on %s from %s", ErrUnknownPriceCheck, instrument, day, source)
}

// checkOwner checks that sessionToken is the owner's on origin; with no
// Owners, nobody is.
func (c *CrossChecks) checkOwner(ctx context.Context, origin, sessionToken string) error {
	if c.owners == nil {
		return ErrNotOwner
	}
	if err := c.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("checking the session: %w", err)
	}
	return nil
}
