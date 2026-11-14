package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views"
)

// ErrNothingHeld reports a portfolio with nothing recorded, so no date it is
// valued through.
var ErrNothingHeld = errors.New("nothing is recorded to value")

// HeldValues is what the accounts the caller may view hold, valued gross in
// the reporting currency (D-77) at the end of the date every one of them is
// valued through: the input of the model portfolio's weights (D-204).
type HeldValues struct {
	On kernel.Date
	// Total is the whole portfolio's value, cash included.
	Total kernel.Money
	// ByInstrument is the value of each instrument held, summed over
	// accounts, by "ISIN/symbol".
	ByInstrument map[string]kernel.Money
	// Unvalued counts the holdings and cash balances left out of Total for
	// want of an input, such as a price.
	Unvalued int
}

// HeldValues returns what the caller's accounts hold, valued gross as of the
// date they are all valued through, or an error wrapping ErrNothingHeld when
// no transaction is recorded.
func (v *Views) HeldValues(ctx context.Context) (HeldValues, error) {
	accounts, err := v.Accounts(ctx)
	if err != nil {
		return HeldValues{}, err
	}
	rec, err := v.recorded(ctx, accounts)
	if err != nil {
		return HeldValues{}, err
	}
	if len(rec.txs) == 0 {
		return HeldValues{}, ErrNothingHeld
	}
	items, unvalued, err := v.itemsOn(ctx, rec.through, views.Gross, false)
	if err != nil {
		return HeldValues{}, err
	}
	out := HeldValues{On: rec.through, Total: kernel.NewMoney(kernel.Decimal{}, v.reporting), ByInstrument: map[string]kernel.Money{}, Unvalued: len(unvalued)}
	for i := range items {
		it := &items[i]
		if out.Total, err = out.Total.Add(it.Reporting); err != nil {
			return HeldValues{}, fmt.Errorf("totalling the portfolio: %w", err)
		}
		if it.AssetClass == views.AssetCash {
			continue
		}
		key := it.Instrument.String()
		sum, ok := out.ByInstrument[key]
		if !ok {
			sum = kernel.NewMoney(kernel.Decimal{}, v.reporting)
		}
		if out.ByInstrument[key], err = sum.Add(it.Reporting); err != nil {
			return HeldValues{}, fmt.Errorf("totalling %s: %w", key, err)
		}
	}
	return out, nil
}
