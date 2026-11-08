package pl

import "strings"

// provisionalFiller pads a seven-character series code to the nine
// characters of an ISIN's national number in a provisional ISIN.
const provisionalFiller = "00"

// provisionalPrefixLen is how many characters of an ISIN its provisional
// prefix (ProvisionalPrefix) fixes: all but the check digit.
const provisionalPrefixLen = 11

// ProvisionalPrefix returns the first eleven characters of series s's
// provisional ISIN: "PL", the series code and "00", such as "PLEDO053600"
// for EDO0536. The check digit that makes it a valid ISIN completes it.
//
// A treasury bond is identified by the registry's ISIN, from the Ministry
// of Finance's data, never by one derived from its series code (D-214).
// The provisional ISIN only holds a series' place in the ledger, which
// needs an ISIN for every holding, until the registry's arrives: it is
// what a lot of a series the data doesn't list yet is booked under, and
// what versions before D-214 booked every series under (D-99). The core
// then re-keys everything stored under it to the registry's ISIN.
func ProvisionalPrefix(s Series) string {
	return polishISINPrefix + s.String() + provisionalFiller
}

// ProvisionalSeries returns the series whose provisional ISIN isin is, a
// valid ISIN, and false for any other ISIN, such as the registry's. A
// registry ISIN of a Polish retail treasury bond, such as PL0000103594,
// never is one: its national number holds no series code.
func ProvisionalSeries(isin string) (Series, bool) {
	if len(isin) != provisionalPrefixLen+1 || !strings.HasPrefix(isin, polishISINPrefix) {
		return Series{}, false
	}
	s, err := ParseSeries(isin[len(polishISINPrefix) : len(polishISINPrefix)+seriesCodeLen])
	if err != nil || ProvisionalPrefix(s) != isin[:provisionalPrefixLen] {
		return Series{}, false
	}
	return s, true
}

// IsProvisional reports whether isin, a valid ISIN, is the provisional
// ISIN of the series whose code is symbol (ProvisionalPrefix): an
// instrument whose ISIN waits for the registry's (D-214).
func IsProvisional(isin, symbol string) bool {
	s, ok := ProvisionalSeries(isin)
	return ok && s.String() == symbol
}
