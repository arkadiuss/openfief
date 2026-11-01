package valuation

import (
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of market-data series, their sources and their gaps.
var (
	// ErrInvalidSeries reports a series whose key names no instrument, no
	// currency other than PLN, or no index kind.
	ErrInvalidSeries = errors.New("invalid market-data series")
	// ErrInvalidGapThreshold reports a gap threshold outside 1 to 3660
	// days.
	ErrInvalidGapThreshold = errors.New("invalid gap threshold")
)

// SourceRulesVersion is the version of the rules by which a series takes
// its values from its sources. Raise it whenever they change, so every
// account is revalued at start-up (app.InputRecorder.ApplyRules). Version
// 2 is D-94 as rewritten: one source per series, chosen on first need;
// version 1 took the highest-ranked source with a value on each date.
const SourceRulesVersion = "2"

// maxSeriesKey is the longest key a series may have: an instrument's
// ISIN, a space and its venue symbol fit well within it.
const maxSeriesKey = 200

// Series is one market-data series: one instrument's prices, one
// currency's FX rates to PLN, or one period index's readings, which take
// every value from exactly one source (D-94). Key is the instrument
// (ledger.Instrument.String()), the currency's code, or the index kind's
// name (IndexKind.String). Market index levels are no series: they have no
// source of their own, and each index and date keeps the level delivered
// last (D-124).
type Series struct {
	Kind MarketDataKind
	Key  string
}

// PriceSeries is the series of instrument's prices.
func PriceSeries(instrument string) Series { return Series{Kind: KindPrices, Key: instrument} }

// FXSeries is the series of currency's FX rates to PLN.
func FXSeries(currency kernel.Currency) Series {
	return Series{Kind: KindFXRates, Key: currency.String()}
}

// IndexSeries is the series of an index kind's readings.
func IndexSeries(kind IndexKind) Series { return Series{Kind: KindIndexReadings, Key: kind.String()} }

// String returns the kind and the key, such as prices/XX00000SYNT3/SYNT or
// fx_rates/USD.
func (s Series) String() string { return s.Kind.String() + "/" + s.Key }

// Validate reports whether s names a series: a known kind and a key of 1
// to 200 characters that, for FX rates, is a currency
// other than PLN and, for index readings, an index kind.
func (s Series) Validate() error {
	if s.Key == "" || len(s.Key) > maxSeriesKey {
		return fmt.Errorf("%w: %s key must have 1 to %d characters: %q", ErrInvalidSeries, s.Kind, maxSeriesKey, s.Key)
	}
	switch s.Kind {
	case KindPrices:
		return nil
	case KindFXRates:
		c, err := kernel.ParseCurrency(s.Key)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidSeries, err)
		}
		if c.Equal(PLN()) {
			return fmt.Errorf("%w: PLN has no FX rate to PLN", ErrInvalidSeries)
		}
		return nil
	case KindIndexReadings:
		if _, err := ParseIndexKind(s.Key); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidSeries, err)
		}
		return nil
	}
	return fmt.Errorf("%w: %w: %d", ErrInvalidSeries, ErrUnknownMarketDataKind, int(s.Kind))
}

// SeriesSource is the source a series takes every value from (D-94), and
// whether the owner chose it; otherwise it was chosen from the kind's
// ranking when the series first needed a value. Available lists, by name,
// every source with a value of the series, which the owner may choose
// from instead.
type SeriesSource struct {
	Series    Series
	Source    string
	ByOwner   bool
	Available []string
}

// ChooseSource returns the source of a series that first needs a value
// (D-94): of available, the sources with a value of the series, the one
// rank places highest, rank being a kind's ranking with the highest first;
// a source rank doesn't name ranks after every named one, by name. It
// returns false when no source has a value.
func ChooseSource(rank, available []string) (string, bool) {
	if len(available) == 0 {
		return "", false
	}
	return slices.MinFunc(available, rankOf(rank)), true
}

// BySource is the values of one series split by source (D-94): Own holds
// those of the series' source, Others those of every other source, which
// only cross-check it (D-22); both in input order.
type BySource[T any] struct {
	Own    []T
	Others []T
}

// bySource splits values by whether source names own.
func bySource[T any](values []T, source func(*T) string, own string) BySource[T] {
	var out BySource[T]
	for i := range values {
		if source(&values[i]) == own {
			out.Own = append(out.Own, values[i])
		} else {
			out.Others = append(out.Others, values[i])
		}
	}
	return out
}

// PricesBySource splits an instrument's prices into those of its source,
// own, and every other source's. A day own has no price for takes the
// latest earlier one of own's (PriceOnOrBefore), never another source's.
func PricesBySource(prices []Price, own string) BySource[Price] {
	return bySource(prices, func(p *Price) string { return p.Source }, own)
}

// FXRatesBySource splits a currency's FX rates into those of its source,
// own, and every other source's.
func FXRatesBySource(rates []FXRate, own string) BySource[FXRate] {
	return bySource(rates, func(r *FXRate) string { return r.Source }, own)
}

// IndexReadingsBySource splits an index kind's readings into those of its
// source, own, and every other source's.
func IndexReadingsBySource(readings []IndexReading, own string) BySource[IndexReading] {
	return bySource(readings, func(r *IndexReading) string { return r.Source }, own)
}

// OtherSourceChecks returns the prices of sources other than their
// instrument's own as cross-check prices of the same instrument, date,
// close and source: another source's value only checks the series' own
// (D-94, D-22).
func OtherSourceChecks(others []Price) []CrossCheckPrice {
	out := make([]CrossCheckPrice, 0, len(others))
	for i := range others {
		p := &others[i]
		out = append(out, CrossCheckPrice{Instrument: p.Instrument, Date: p.Date, Close: p.Close, Source: p.Source})
	}
	return out
}

// maxGapThreshold is the longest gap threshold the owner may set: about
// ten years, in days.
const maxGapThreshold = 3660

// Default gap thresholds, in days.
const (
	// defaultDailyGap covers a long weekend with holidays around it, such
	// as Christmas, for the daily prices and FX rates.
	defaultDailyGap = 7
	// defaultReadingGap covers a monthly reading published in the middle
	// of the month after the one it measures, as GUS publishes CPI.
	defaultReadingGap = 92
)

// GapThresholds are the owner's gap threshold of each kind of market data,
// in days (D-94): a series whose source's latest value is more days older
// than the day valued raises an attention item. A kind Days doesn't hold
// takes DefaultGapThreshold.
type GapThresholds struct {
	Days map[MarketDataKind]int
}

// DefaultGapThreshold returns kind's gap threshold unless the owner sets
// one: 7 days for prices and FX rates, 92 for index readings.
func DefaultGapThreshold(kind MarketDataKind) int {
	if kind == KindIndexReadings {
		return defaultReadingGap
	}
	return defaultDailyGap
}

// For returns kind's gap threshold: the owner's, or the default.
func (t GapThresholds) For(kind MarketDataKind) int {
	if d, ok := t.Days[kind]; ok {
		return d
	}
	return DefaultGapThreshold(kind)
}

// ValidateGapThreshold reports whether days can be kind's gap threshold:
// a known kind and 1 to 3660 days.
func ValidateGapThreshold(kind MarketDataKind, days int) error {
	if DefaultInstitution(kind) == "" {
		return fmt.Errorf("%w: %d", ErrUnknownMarketDataKind, int(kind))
	}
	if days < 1 || days > maxGapThreshold {
		return fmt.Errorf("%w: %s: %d days; it must be 1 to %d", ErrInvalidGapThreshold, kind, days, maxGapThreshold)
	}
	return nil
}

// SeriesGap is a series whose source has no value on a day valued and
// none for longer than its gap threshold before it (D-94).
type SeriesGap struct {
	Series Series
	Source string
	// Latest is the date of the source's latest value on or before Day,
	// zero when it has none.
	Latest kernel.Date
	Day    kernel.Date
}

// Days returns the days from Latest to Day, or -1 when the source has no
// value on or before Day.
func (g *SeriesGap) Days() int {
	if g.Latest.IsZero() {
		return -1
	}
	return g.Latest.DaysUntil(g.Day)
}

// Detail says what the gap is, for the owner: since when the source has
// no value. It names no day valued, so it stays the same while the gap
// lasts and an item raised with it isn't opened again each day.
func (g *SeriesGap) Detail() string {
	if g.Latest.IsZero() {
		return "no value from " + g.Source + " yet"
	}
	return "no value from " + g.Source + " since " + g.Latest.String()
}

// GapOf returns the gap of series on day, its source being source and
// latest the date of source's latest value on or before day (zero when it
// has none), and whether it is longer than threshold days (D-94). A gap of
// exactly threshold days is no gap yet.
func GapOf(series Series, source string, latest, day kernel.Date, threshold int) (SeriesGap, bool) {
	g := SeriesGap{Series: series, Source: source, Latest: latest, Day: day}
	return g, latest.IsZero() || g.Days() > threshold
}
