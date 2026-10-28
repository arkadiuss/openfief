package pl

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrInvalidTerms reports series terms that cannot be used, such as a
// negative or non-PLN early-redemption fee.
var ErrInvalidTerms = errors.New("invalid series terms")

// Nominal returns the face value of one retail treasury bond: 100 PLN.
func Nominal() kernel.Money {
	const nominalPLN = 100
	return kernel.NewMoney(kernel.NewDecimal(nominalPLN, 0), pln())
}

func pln() kernel.Currency {
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		return kernel.Currency{}
	}
	return c
}

// Terms are what the issue letter of one series fixes. They change from
// issue to issue, so they are data supplied by the caller, never constants.
// Rates are yearly and written as fractions: 5.35% is 0.0535.
type Terms struct {
	Series Series
	// FirstPeriodRate is the fixed yearly rate of the first interest period;
	// for OTS and TOS it is the rate of the whole term.
	FirstPeriodRate kernel.Decimal
	// Margin is added to the period index from the second period on; it is
	// unused for fixed-rate types.
	Margin kernel.Decimal
	// EarlyRedemptionFee is charged per bond redeemed before maturity, in PLN.
	EarlyRedemptionFee kernel.Money
}

// Indices are the published readings that set the rates of interest periods
// after the first, keyed by period number (2, 3, ...): CPI year on year from
// GUS for inflation-indexed types (see CPIMonth), or the NBP reference rate
// for ROR and DOR. A reading is a yearly fraction: CPI of 104.9 (previous
// year = 100) is 0.049.
type Indices map[int]kernel.Decimal

// MissingIndexError reports an interest period whose index reading the
// caller did not supply.
type MissingIndexError struct {
	Series Series
	Period int
	Start  kernel.Date
}

func (e *MissingIndexError) Error() string {
	return fmt.Sprintf("no index reading for %s interest period %d starting %s", e.Series, e.Period, e.Start)
}

// PeriodRate returns the yearly rate of interest period n starting on start.
// Period 1 and every period of a fixed-rate type use FirstPeriodRate. Later
// periods of an inflation-indexed type use the CPI reading plus Margin, where
// a negative CPI (deflation) counts as zero, so the rate never falls below the
// margin. Later periods of ROR and DOR use the NBP reference rate plus Margin.
func (t *Terms) PeriodRate(idx Indices, n int, start kernel.Date) (kernel.Decimal, error) {
	st, err := t.Series.Type.Structure()
	if err != nil {
		return kernel.Decimal{}, err
	}
	if n == 1 || st.Indexation == Fixed {
		return t.FirstPeriodRate, nil
	}
	reading, ok := idx[n]
	if !ok {
		return kernel.Decimal{}, &MissingIndexError{Series: t.Series, Period: n, Start: start}
	}
	if st.Indexation == Inflation && reading.Sign() < 0 {
		reading = kernel.Decimal{}
	}
	return reading.Add(t.Margin), nil
}

func (t *Terms) validate() error {
	if !t.EarlyRedemptionFee.Currency().Equal(pln()) || t.EarlyRedemptionFee.Sign() < 0 {
		return fmt.Errorf("%w: %s: early-redemption fee %s must be a PLN amount of zero or more", ErrInvalidTerms, t.Series, t.EarlyRedemptionFee)
	}
	if t.FirstPeriodRate.Sign() < 0 {
		return fmt.Errorf("%w: %s: negative first-period rate %s", ErrInvalidTerms, t.Series, t.FirstPeriodRate)
	}
	return nil
}
