package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// PriceChecks compares the stored prices with cross-check prices (D-22).
// *valuation/app.CrossChecks implements it.
type PriceChecks interface {
	// PriceChecks returns every price check, in date, instrument and
	// cross-check source order.
	PriceChecks(ctx context.Context) ([]valuation.PriceCheck, error)
	// Threshold returns the relative difference beyond which a check is
	// flagged.
	Threshold() kernel.Decimal
}

// WithPriceChecks serves price checks from checks. Without it there are
// none, and no threshold.
func WithPriceChecks(checks PriceChecks) Option {
	return func(v *Views) { v.priceChecks = checks }
}

// PriceCheckList is the price checks the read API serves, with the
// threshold they were flagged at; the zero threshold when no prices are
// compared.
type PriceCheckList struct {
	Threshold kernel.Decimal
	Checks    []valuation.PriceCheck
}

// PriceChecks returns the price checks, only those with status when status
// is not zero, and only those of the listings of the instrument with ISIN
// isin, its check history, when isin is not empty (D-112, D-195).
func (v *Views) PriceChecks(ctx context.Context, status valuation.PriceCheckStatus, isin string) (PriceCheckList, error) {
	if v.priceChecks == nil {
		return PriceCheckList{}, nil
	}
	checks, err := v.priceChecks.PriceChecks(ctx)
	if err != nil {
		return PriceCheckList{}, fmt.Errorf("loading price checks: %w", err)
	}
	out := PriceCheckList{Threshold: v.priceChecks.Threshold(), Checks: checks[:0:0]}
	for i := range checks {
		if (status == 0 || checks[i].Status == status) && (isin == "" || ofISIN(checks[i].CrossCheck.Instrument, isin)) {
			out.Checks = append(out.Checks, checks[i])
		}
	}
	return out, nil
}

// ofISIN reports whether instrument, written as ledger.Instrument.String()
// ("<ISIN>/<symbol>"), is a listing of isin.
func ofISIN(instrument, isin string) bool {
	return instrument == isin || strings.HasPrefix(instrument, isin+"/")
}
