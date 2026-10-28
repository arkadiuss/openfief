package pl

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrSeriesMismatch reports a purchase date whose maturity does not fall in
// the month the series code names, such as an EDO0536 bought in June 2026.
var ErrSeriesMismatch = errors.New("purchase date does not match the bond series")

// ErrBeforePurchase reports a valuation date before the purchase date.
var ErrBeforePurchase = errors.New("valuation date before purchase")

// addMonths returns the date n months after d, on the same day of the month,
// or on the month's last day when that month is shorter (Jan 31 + 1 month is
// Feb 28 or 29). It is always counted from d itself, never chained, so a bond
// bought on the 31st returns to the 31st after a short month. This is the
// schedule the ROR and DOR issue letters print in Załącznik nr 3 ("Wykaz
// terminów rozpoczęcia i zakończenia poszczególnych okresów odsetkowych"):
// a ROR bought on 31.10.2026 has periods ending 30.11.2026, 31.12.2026,
// 31.01.2027, 28.02.2027 and 31.03.2027 (list emisyjny nr 95/2026, ROR1027,
// https://www.obligacjeskarbowe.pl/oferta-obligacji/obligacje-roczne-ror/ror1027/).
// The interest tables at https://www.obligacjeskarbowe.pl/tabela-odsetkowa/
// follow the same rule for bonds bought on the 29th to 31st.
func addMonths(d kernel.Date, n int) (kernel.Date, error) {
	total := d.Year()*monthsPerYear + int(d.Month()) - 1 + n
	year, month := total/monthsPerYear, time.Month(total%monthsPerYear+1)
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	r, err := kernel.NewDate(year, month, min(d.Day(), last))
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%s + %d months: %w", d, n, err)
	}
	return r, nil
}

// Maturity returns the redemption date of a bond of s bought on purchased:
// its term runs from its own purchase date. It fails with ErrSeriesMismatch
// when that date is not in the month the series code names, which catches a
// lot entered under the wrong series.
func Maturity(s Series, purchased kernel.Date) (kernel.Date, error) {
	st, err := s.Type.Structure()
	if err != nil {
		return kernel.Date{}, err
	}
	m, err := addMonths(purchased, st.TenorMonths)
	if err != nil {
		return kernel.Date{}, err
	}
	if m.Year() != s.MaturityYear || m.Month() != s.MaturityMonth {
		return kernel.Date{}, fmt.Errorf("%w: %s bought on %s matures on %s", ErrSeriesMismatch, s, purchased, m)
	}
	return m, nil
}

// PeriodStart returns the first day of interest period n (1-based) of a bond
// of type b bought on purchased. Period n+1 starts the day period n ends, so
// the day count of period n is PeriodStart(n).DaysUntil(PeriodStart(n+1)).
func PeriodStart(b BondType, purchased kernel.Date, n int) (kernel.Date, error) {
	st, err := b.Structure()
	if err != nil {
		return kernel.Date{}, err
	}
	return addMonths(purchased, (n-1)*st.PeriodMonths)
}

// CPIMonth returns the month whose CPI (year on year, from GUS) sets the rate
// of an inflation-indexed interest period starting on start. The issue
// letters take the reading "ogłaszanej przez Prezesa Głównego Urzędu
// Statystycznego w miesiącu poprzedzającym pierwszy miesiąc danego okresu
// odsetkowego" (EDO1036 letter ust. 16 and Załącznik nr 1); GUS announces a
// month's final CPI in the following month, so that is the reading for the
// month two months before the period starts. The published period rates bear
// this out: EDO1133 (margin 1.50%) earns 6.40% from November 2024 and 4.40%
// from November 2025, which is September CPI of 4.9% and 2.9% plus the margin
// (interest tables at https://www.obligacjeskarbowe.pl/tabela-odsetkowa/).
// It returns that month's year and month.
func CPIMonth(start kernel.Date) (int, time.Month) {
	const lag = 2
	total := start.Year()*monthsPerYear + int(start.Month()) - 1 - lag
	return total / monthsPerYear, time.Month(total%monthsPerYear + 1)
}
