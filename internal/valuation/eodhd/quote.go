package eodhd

import (
	"slices"
	"strings"
)

// maxQuoteCandidates is the most listings a query that names several is
// answered with.
const maxQuoteCandidates = 10

// usVenues are the exchange names EODHD's search gives a US listing, all of
// which its end-of-day API takes as the exchange code "US".
func usVenues() []string {
	return []string{"US", "NASDAQ", "NYSE", "NYSE ARCA", "NYSE MKT", "AMEX", "BATS", "OTC", "OTCQB", "OTCQX", "PINK"}
}

// venueLSE is the London Stock Exchange's code.
const venueLSE = "LSE"

// venueWAR is the Warsaw Stock Exchange's code.
const venueWAR = "WAR"

// venueXETRA is Xetra's code.
const venueXETRA = "XETRA"

// ExchangeCode returns the exchange code EODHD's end-of-day API takes for the
// venue its search names: "US" for every US venue, the code as it is for any
// other.
func ExchangeCode(exchange string) string {
	if slices.Contains(usVenues(), strings.ToUpper(exchange)) {
		return "US"
	}
	return exchange
}

// QuoteTicker is the listing's end-of-day ticker, such as "NVDA.US", which
// differs from Ticker for a US listing whose venue is NASDAQ or NYSE.
func (l *Listing) QuoteTicker() string { return l.Code + "." + ExchangeCode(l.Exchange) }

// venueRank orders the venues a quote prefers when one instrument is listed
// on several: the US venues, then the supported exchanges in D-51's order,
// then every other venue by name.
func venueRank(exchange string) int {
	if ExchangeCode(exchange) == "US" {
		return 0
	}
	order := []string{venueWAR, venueLSE, venueXETRA, "PA", "AS", "BR", "LS", "IR", "MI", "OL"}
	if i := slices.Index(order, exchange); i >= 0 {
		return 1 + i
	}
	return len(order) + 1
}

// QuoteMatch is what MatchQuery found: the one listing a query names, or, for
// a ticker several ISINs share, the candidates and no listing.
type QuoteMatch struct {
	Listing    Listing
	Candidates []Listing
}

// MatchQuery resolves a ticker or an ISIN, query, among the listings a
// search found. A listing matches when its code, or its ISIN, equals query
// without regard to case; each ISIN and end-of-day ticker counts once. One instrument,
// one ISIN, listed on several venues is answered by its most preferred venue
// (venueRank), as the same security. Several ISINs under one ticker is
// ErrAmbiguous: no listing is chosen and the candidates, at most ten and one
// per ISIN, are returned for the person to choose by ISIN. No match is
// ErrNotListed.
func MatchQuery(listings []Listing, query string) (QuoteMatch, error) {
	var found []Listing
	for _, l := range listings {
		if !strings.EqualFold(l.Code, query) && !strings.EqualFold(l.ISIN.String(), query) {
			continue
		}
		if !slices.ContainsFunc(found, func(f Listing) bool { return f.ISIN == l.ISIN && f.QuoteTicker() == l.QuoteTicker() }) {
			found = append(found, l)
		}
	}
	if len(found) == 0 {
		return QuoteMatch{}, ErrNotListed
	}
	slices.SortStableFunc(found, func(a, b Listing) int { return venueRank(a.Exchange) - venueRank(b.Exchange) })
	var distinct []Listing
	for _, l := range found {
		if !slices.ContainsFunc(distinct, func(d Listing) bool { return d.ISIN == l.ISIN }) {
			distinct = append(distinct, l)
		}
	}
	if len(distinct) > 1 {
		return QuoteMatch{Candidates: distinct[:min(len(distinct), maxQuoteCandidates)]}, ErrAmbiguous
	}
	return QuoteMatch{Listing: found[0]}, nil
}
