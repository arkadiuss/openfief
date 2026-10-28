package demo

import (
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// BondSeries is a Polish retail treasury bond series the demo holds: a real
// series with its real, public issue terms, under a synthetic ISIN.
type BondSeries struct {
	Instrument ledger.Instrument
	Terms      pl.Terms
}

// bondSpec is a bond series with the demo's opening lot of it.
type bondSpec struct {
	number    int
	series    string
	first     string
	margin    string
	fee       string
	purchased string
	units     int64
}

// bondSpecs are the demo's bonds. The series and their terms are public:
// first-year rate, margin and early-redemption fee per bond as the issue
// letters fix them, published with the interest tables at
// https://www.obligacjeskarbowe.pl/tabela-odsetkowa/ and checked against
// those tables in internal/jurisdiction/pl/official_test.go. Each lot is
// bought in its series' sale month. The ISINs are not the series' real
// ones: they are synthetic (IsSynthetic), so that no demo instrument can be
// mistaken for a real holding.
func bondSpecs() []bondSpec {
	return []bondSpec{
		{9001, "EDO1133", "0.07", "0.015", "2.00", "2023-11-15", 40},
		{9002, "COI0128", "0.0665", "0.0125", "0.70", "2024-01-10", 30},
		{9003, "TOS0128", "0.0595", "0", "0.70", "2025-01-20", 20},
	}
}

// cpiTenths are GUS's consumer price index, the same month of the previous
// year = 100, in tenths (1049 is 104.9), from January 2023 to September 2026
// (Wskaźnik cen towarów i usług konsumpcyjnych, "Analogiczny miesiąc
// poprzedniego roku = 100", Polska: the CSV "miesięczne wskaźniki cen
// towarów i usług konsumpcyjnych od 1982 roku" at https://stat.gov.pl/
// obszary-tematyczne/ceny-handel/wskazniki-cen/wskazniki-cen-towarow-i-
// uslug-konsumpcyjnych-pot-inflacja-/, read on 2026-10-28). The bond
// periods use September and November readings (pl.CPIMonth); the rest make
// the series complete.
func cpiTenths() [][12]int64 {
	return [][12]int64{
		{1166, 1184, 1161, 1147, 1130, 1115, 1108, 1101, 1082, 1066, 1066, 1062}, // 2023
		{1037, 1028, 1020, 1024, 1025, 1026, 1042, 1043, 1049, 1050, 1047, 1047}, // 2024
		{1049, 1049, 1049, 1043, 1040, 1041, 1031, 1029, 1029, 1028, 1025, 1024}, // 2025
		{1021, 1021, 1030, 1032, 1031, 1025, 1030, 1034, 1041},                   // 2026, to September
	}
}

// cpiFirstYear is the year of cpiTenths' first row.
const cpiFirstYear = 2023

// CPIReadings returns the GUS readings of cpiTenths as yearly fractions:
// 104.9 is 0.049.
func CPIReadings() []valuation.IndexReading {
	const base, exp = 1000, -3
	var out []valuation.IndexReading
	for y, months := range cpiTenths() {
		for m, v := range months {
			if v == 0 {
				continue
			}
			out = append(out, valuation.IndexReading{
				Kind: valuation.CPI, Year: cpiFirstYear + y, Month: time.Month(m + 1),
				Value: kernel.NewDecimal(v-base, exp), Source: "gus-cpi-yoy",
			})
		}
	}
	return out
}

// bondLot is one opening lot of a bond series.
type bondLot struct {
	series    BondSeries
	units     int64
	purchased kernel.Date
}

// Bonds returns the demo's bond series with their terms.
func Bonds() ([]BondSeries, error) {
	lots, err := bondLots()
	if err != nil {
		return nil, err
	}
	out := make([]BondSeries, 0, len(lots))
	for i := range lots {
		out = append(out, lots[i].series)
	}
	return out, nil
}

// bondLots returns the demo's bond lots from bondSpecs.
func bondLots() ([]bondLot, error) {
	specs := bondSpecs()
	out := make([]bondLot, 0, len(specs))
	for i := range specs {
		s := &specs[i]
		l, err := s.lot()
		if err != nil {
			return nil, fmt.Errorf("demo bond %s: %w", s.series, err)
		}
		out = append(out, l)
	}
	return out, nil
}

// lot parses s.
func (s *bondSpec) lot() (bondLot, error) {
	isin, err := SyntheticISIN("PL", s.number)
	if err != nil {
		return bondLot{}, err
	}
	series, err := pl.ParseSeries(s.series)
	if err != nil {
		return bondLot{}, fmt.Errorf("demo bond: %w", err)
	}
	first, err1 := kernel.ParseDecimal(s.first)
	margin, err2 := kernel.ParseDecimal(s.margin)
	fee, err3 := kernel.ParseMoney(s.fee, pl.Nominal().Currency())
	purchased, err4 := kernel.ParseDate(s.purchased)
	if err := firstError(err1, err2, err3, err4); err != nil {
		return bondLot{}, err
	}
	terms := pl.Terms{Series: series, FirstPeriodRate: first, Margin: margin, EarlyRedemptionFee: fee}
	return bondLot{
		series:    BondSeries{Instrument: ledger.Instrument{ISIN: isin, Symbol: s.series}, Terms: terms},
		units:     s.units,
		purchased: purchased,
	}, nil
}

// firstError returns the first non-nil error of errs.
func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
