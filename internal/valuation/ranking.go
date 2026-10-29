package valuation

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
)

// Errors of market-data sources and their rankings.
var (
	// ErrUnknownMarketDataKind reports a MarketDataKind outside the
	// declared constants, or text that names none of them.
	ErrUnknownMarketDataKind = errors.New("unknown kind of market data")
	// ErrInvalidRanking reports a source ranking that names a source twice
	// or names one with an empty or overlong name.
	ErrInvalidRanking = errors.New("invalid source ranking")
)

// maxSourceName is the longest name a source may have: a connector
// instance's name is at most 56 characters (D-48).
const maxSourceName = 64

// MarketDataKind is a kind of market data whose sources the owner ranks
// (D-94): prices, FX rates and index readings.
type MarketDataKind int

const (
	// KindPrices are end-of-day prices of instruments (D-21).
	KindPrices MarketDataKind = iota + 1
	// KindFXRates are FX rates to PLN (D-20).
	KindFXRates
	// KindIndexReadings are period index readings: CPI and the NBP
	// reference rate.
	KindIndexReadings
)

// MarketDataKinds returns every kind, in display order.
func MarketDataKinds() []MarketDataKind {
	return []MarketDataKind{KindPrices, KindFXRates, KindIndexReadings}
}

// String returns the kind's name: prices, fx_rates or index_readings.
func (k MarketDataKind) String() string {
	switch k {
	case KindPrices:
		return "prices"
	case KindFXRates:
		return "fx_rates"
	case KindIndexReadings:
		return "index_readings"
	}
	return "MarketDataKind(" + strconv.Itoa(int(k)) + ")"
}

// ParseMarketDataKind parses a kind's name as String writes it.
func ParseMarketDataKind(s string) (MarketDataKind, error) {
	for _, k := range MarketDataKinds() {
		if k.String() == s {
			return k, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownMarketDataKind, s)
}

// DefaultInstitution returns the institution of the first-party connector
// that ranks first for k unless the owner ranks otherwise (D-94): EODHD for
// prices, NBP for FX rates and the treasury bond connector for CPI and the
// NBP reference rate. A connector names its institution on every delivery.
func DefaultInstitution(k MarketDataKind) string {
	switch k {
	case KindPrices:
		return "eodhd"
	case KindFXRates:
		return "nbp"
	case KindIndexReadings:
		return "treasury-bonds"
	}
	return ""
}

// Source is a source of one kind of market data: the connector instance
// that delivered values of it, and the institution it named (D-48).
type Source struct {
	Kind        MarketDataKind
	Name        string
	Institution string
}

// RankedSource is one source in a ranking, with whether the owner placed
// it: an unplaced source ranks after every placed one, by default.
type RankedSource struct {
	Name        string
	Institution string
	Placed      bool
}

// Ranking is the order of the sources of one kind of market data, the
// highest first (D-94).
type Ranking struct {
	Kind    MarketDataKind
	Sources []RankedSource
}

// Names returns the sources' names, the highest-ranked first.
func (r *Ranking) Names() []string {
	out := make([]string, len(r.Sources))
	for i := range r.Sources {
		out[i] = r.Sources[i].Name
	}
	return out
}

// RankSources returns the ranking of kind's sources (D-94): first those
// the owner placed, in the owner's order, then every other known source of
// kind, those of DefaultInstitution(kind) first, then by name. A source
// the owner placed that never delivered keeps its place, with no
// institution. Sources of other kinds are ignored.
func RankSources(kind MarketDataKind, known []Source, placed []string) Ranking {
	institution := map[string]string{}
	var others []Source
	for _, s := range known {
		if s.Kind != kind {
			continue
		}
		institution[s.Name] = s.Institution
		if !slices.Contains(placed, s.Name) {
			others = append(others, s)
		}
	}
	out := Ranking{Kind: kind}
	for _, name := range placed {
		out.Sources = append(out.Sources, RankedSource{Name: name, Institution: institution[name], Placed: true})
	}
	first := DefaultInstitution(kind)
	slices.SortFunc(others, func(a, b Source) int {
		return cmp.Or(-cmp.Compare(btoi(a.Institution == first), btoi(b.Institution == first)), cmp.Compare(a.Name, b.Name))
	})
	others = slices.CompactFunc(others, func(a, b Source) bool { return a.Name == b.Name })
	for _, s := range others {
		out.Sources = append(out.Sources, RankedSource{Name: s.Name, Institution: s.Institution})
	}
	return out
}

// btoi is 1 for true and 0 for false.
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ValidateRanking reports whether sources can be the owner's ranking of
// kind: a known kind, and each source named once, by a name of 1 to 64
// characters. An empty ranking leaves every source in its default place.
func ValidateRanking(kind MarketDataKind, sources []string) error {
	if DefaultInstitution(kind) == "" {
		return fmt.Errorf("%w: %d", ErrUnknownMarketDataKind, int(kind))
	}
	seen := map[string]bool{}
	for _, s := range sources {
		switch {
		case s == "" || len(s) > maxSourceName:
			return fmt.Errorf("%w: a source name must have 1 to %d characters: %q", ErrInvalidRanking, maxSourceName, s)
		case seen[s]:
			return fmt.Errorf("%w: %q is ranked twice", ErrInvalidRanking, s)
		}
		seen[s] = true
	}
	return nil
}

// rankOf returns the position of each source in rank: a source rank
// doesn't name comes after every named one, and two of those are ordered
// by name.
func rankOf(rank []string) func(a, b string) int {
	pos := make(map[string]int, len(rank))
	for i, name := range rank {
		pos[name] = i
	}
	return func(a, b string) int {
		pa, oka := pos[a]
		pb, okb := pos[b]
		switch {
		case oka && okb:
			return cmp.Compare(pa, pb)
		case oka:
			return -1
		case okb:
			return 1
		}
		return cmp.Compare(a, b)
	}
}

// Split is values of one kind of market data split by rank (D-94):
// Winners holds the value of each key from the highest-ranked source with
// one, in the order the keys first appear; Lower every other value, in
// input order.
type Split[T any] struct {
	Winners []T
	Lower   []T
}

// split splits values, keyed by key and from source, by rank.
func split[T any](values []T, key, source func(*T) string, rank []string) Split[T] {
	better := rankOf(rank)
	best := map[string]int{}
	var order []string
	for i := range values {
		k := key(&values[i])
		j, seen := best[k]
		switch {
		case !seen:
			best[k] = i
			order = append(order, k)
		case better(source(&values[i]), source(&values[j])) < 0:
			best[k] = i
		}
	}
	won := make(map[int]bool, len(best))
	out := Split[T]{Winners: make([]T, 0, len(order))}
	for _, k := range order {
		won[best[k]] = true
		out.Winners = append(out.Winners, values[best[k]])
	}
	for i := range values {
		if !won[i] {
			out.Lower = append(out.Lower, values[i])
		}
	}
	return out
}

// RankedPrices splits prices into the winning price of each instrument and
// date, from the highest-ranked source with a value (D-94), and the
// lower-ranked ones, which serve as cross-checks (D-22).
func RankedPrices(prices []Price, rank []string) Split[Price] {
	return split(prices, func(p *Price) string { return p.Instrument + " " + p.Date.String() },
		func(p *Price) string { return p.Source }, rank)
}

// RankedFXRates returns the winning FX rate of each currency and date,
// from the highest-ranked source with a value (D-94).
func RankedFXRates(rates []FXRate, rank []string) []FXRate {
	return split(rates, func(r *FXRate) string { return r.Currency.String() + " " + r.Date.String() },
		func(r *FXRate) string { return r.Source }, rank).Winners
}

// RankedIndexReadings returns the winning reading of each index kind and
// month, from the highest-ranked source with a value (D-94).
func RankedIndexReadings(readings []IndexReading, rank []string) []IndexReading {
	return split(readings, func(r *IndexReading) string {
		return r.Kind.String() + " " + strconv.Itoa(r.Year) + " " + strconv.Itoa(int(r.Month))
	}, func(r *IndexReading) string { return r.Source }, rank).Winners
}

// LowerRankedChecks returns lower-ranked prices as cross-check prices of
// the same instrument, date, close and source: a lower-ranked source's
// value is compared with the winning price like an institution's own
// price (D-94, D-22).
func LowerRankedChecks(lower []Price) []CrossCheckPrice {
	out := make([]CrossCheckPrice, 0, len(lower))
	for i := range lower {
		p := &lower[i]
		out = append(out, CrossCheckPrice{Instrument: p.Instrument, Date: p.Date, Close: p.Close, Source: p.Source})
	}
	return out
}
