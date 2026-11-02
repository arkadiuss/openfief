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
// (D-94): prices, FX rates and index readings. Market index levels have
// no ranking (D-124).
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
// NBP reference rate. A connector names its institution on every delivery,
// and the core refuses one naming an institution its signed manifest
// doesn't declare (D-182), so a source ranks first by default only when its
// manifest declares that institution.
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
// highest first (D-94): a series of that kind takes its source from it when
// it first needs a value (ChooseSource).
type Ranking struct {
	Kind    MarketDataKind
	Sources []RankedSource
	// GapThreshold is the kind's gap threshold in days (GapThresholds),
	// which the app fills in; RankSources leaves it zero.
	GapThreshold int
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

// rankOf returns the order of sources in rank: a source rank
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
