package pl

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrUnknownTaxWrapper reports a TaxWrapper outside the declared constants.
var ErrUnknownTaxWrapper = errors.New("unknown tax wrapper")

// TaxWrapper is the tax regime of the account a lot is held in.
type TaxWrapper int

const (
	// Taxable is an ordinary account: Belka tax is withheld from bond interest.
	Taxable TaxWrapper = iota + 1
	// IKE is an individual retirement account; its income is exempt from
	// Belka tax (art. 21 ust. 1 pkt 58 of the PIT act) when the withdrawal
	// conditions are met, which valuation assumes.
	IKE
	// IKZE is an individual retirement security account; no Belka tax is
	// withheld inside it. The flat tax on IKZE withdrawals applies to the
	// whole account, not per bond, and is not part of a bond's value.
	IKZE
)

// BelkaRate returns the 19% flat tax rate on interest (art. 30a ust. 1 pkt 2
// of the PIT act).
func BelkaRate() kernel.Decimal {
	const percent = 19
	return kernel.NewDecimal(percent, -2)
}

// BelkaTax returns the Belka tax withheld from interest earned in wrapper:
// 19% of interest, rounded half up to whole grosze (art. 63 § 1a of the tax
// ordinance); zero when interest is not positive or inside IKE and IKZE.
func BelkaTax(interest kernel.Money, wrapper TaxWrapper) (kernel.Money, error) {
	tax, err := belkaTax(interest.Amount(), wrapper)
	if err != nil {
		return kernel.Money{}, err
	}
	return kernel.NewMoney(tax, interest.Currency()), nil
}

func belkaTax(interest kernel.Decimal, wrapper TaxWrapper) (kernel.Decimal, error) {
	switch wrapper {
	case IKE, IKZE:
		return kernel.Decimal{}, nil
	case Taxable:
		if interest.Sign() <= 0 {
			return kernel.Decimal{}, nil
		}
		tax, err := interest.Mul(BelkaRate()).Round(grosze, kernel.RoundHalfUp)
		if err != nil {
			return kernel.Decimal{}, fmt.Errorf("belka tax: %w", err)
		}
		return tax, nil
	}
	return kernel.Decimal{}, fmt.Errorf("%w: %d", ErrUnknownTaxWrapper, wrapper)
}

// TaxRateCutoff returns the latest date whose NBP table A rate may convert
// income received on event for Polish tax: art. 11a ust. 1 of the PIT act
// takes the NBP average rate of the last business day before the day of the
// income. NBP publishes tables on business days only, so the latest table
// dated on or before the returned day is that rate.
func TaxRateCutoff(event kernel.Date) kernel.Date {
	return event.AddDays(-1)
}
