package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
)

// memCrossChecks keeps the cross-check prices it is given; it holds no
// prices.
type memCrossChecks struct {
	saved []valuation.CrossCheckPrice
}

func (m *memCrossChecks) SaveCrossCheckPrices(_ context.Context, prices []valuation.CrossCheckPrice) error {
	m.saved = append(m.saved, prices...)
	return nil
}

func (m *memCrossChecks) CrossCheckedPrices(context.Context) ([]valuation.Price, []valuation.CrossCheckPrice, error) {
	return nil, m.saved, nil
}

func (*memCrossChecks) PriceCheckThresholds(context.Context) (valuation.Thresholds, error) {
	return valuation.Thresholds{}, nil
}

func (*memCrossChecks) SavePriceCheckThreshold(context.Context, string, *kernel.Decimal) error {
	return nil
}

func (*memCrossChecks) PriceCheckAcknowledgements(context.Context) ([]valuation.PriceCheckAcknowledgement, error) {
	return nil, nil
}

func (*memCrossChecks) AppendPriceCheckAcknowledgement(context.Context, *valuation.PriceCheckAcknowledgement) error {
	return nil
}

// TestCrossCheckInputs checks that the core hands a delivery's cross-check
// prices to valuation's cross-check use cases, which store valid ones and
// refuse a batch with an invalid one.
func TestCrossCheckInputs(t *testing.T) {
	t.Parallel()
	store := &memCrossChecks{}
	checks := valuationapp.NewCrossChecks(store, nil)
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		t.Fatalf("ParseCurrency: %v", err)
	}
	day, err := kernel.ParseDate("2026-03-06")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	mark := valuation.CrossCheckPrice{
		Instrument: "PLTEST000010/TSTA", Date: day, Close: kernel.NewMoney(kernel.NewDecimal(10125, -2), usd), Source: "ibkr-main",
	}
	inputs := crossCheckInputs{checks: checks}
	if storeErr := inputs.StoreCrossCheckPrices(t.Context(), "b-1", []valuation.CrossCheckPrice{mark}); storeErr != nil {
		t.Fatalf("StoreCrossCheckPrices: %v", storeErr)
	}
	if diff := cmp.Diff([]valuation.CrossCheckPrice{mark}, store.saved); diff != "" {
		t.Errorf("stored (-want +got):\n%s", diff)
	}
	bad := mark
	bad.Close = kernel.NewMoney(kernel.Decimal{}, usd)
	err = inputs.StoreCrossCheckPrices(t.Context(), "b-2", []valuation.CrossCheckPrice{bad})
	if diff := cmp.Diff(valuationapp.ErrInvalidInput, err, cmpopts.EquateErrors()); diff != "" {
		t.Errorf("StoreCrossCheckPrices(zero close) error (-want +got):\n%s", diff)
	}
}
