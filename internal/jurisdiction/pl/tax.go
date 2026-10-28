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
	// conditions are met, which valuation assumes. Bonds redeemed for an IKE
	// withdrawal pay no early-redemption fee (issue letters of ROR, DOR, TOS,
	// COI and EDO, the only types they offer for IKE and IKZE).
	IKE
	// IKZE is an individual retirement security account; no Belka tax is
	// withheld inside it and, as for IKE, no early-redemption fee is charged
	// on a withdrawal. The flat tax on IKZE withdrawals applies to the whole
	// account, not per bond, and is not part of a bond's value.
	IKZE
)

// BelkaRate returns the 19% flat tax rate on interest and discount of
// securities (art. 30a ust. 1 pkt 2 of the PIT act: "z odsetek i dyskonta od
// papierów wartościowych"). The discount of a bond bought by exchange below
// its nominal is taxed with its interest
// (https://www.obligacjeskarbowe.pl/faq/: "W przypadku, gdy obligacje były
// zakupione z dyskontem (np. w drodze zamiany) - opodatkowaniu 19-proc.
// podatkiem podlega również wartość dyskonta").
func BelkaRate() kernel.Decimal {
	const percent = 19
	return kernel.NewDecimal(percent, -2)
}

// BelkaTax returns the Belka tax withheld from interest earned in wrapper:
// 19% of interest, rounded half up to whole grosze; zero when interest is not
// positive or inside IKE and IKZE.
//
// The rounding follows the issuer's worked examples, which disagree with the
// statute's wording. Art. 63 § 1a of the tax ordinance (t.j. Dz. U. z 2026 r.
// poz. 622, https://api.sejm.gov.pl/eli/acts/DU/1997/926/text/U/D19970926Lj.pdf)
// says the bases and taxes of art. 30a ust. 1 pkt 1-3 of the PIT act are
// rounded "do pełnych groszy w górę", up. The examples at
// https://www.obligacjeskarbowe.pl/oferta-obligacji/ round half up: EDO0524
// withholds 0.41 PLN on 2.16 PLN (19% is 0.4104), ROR0623 0.08 PLN on a
// 0.44 PLN coupon (0.0836) and COI0423 1.85 PLN on 9.75 PLN (1.8525). Which
// one PKO BP applies is an open question for the author.
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
