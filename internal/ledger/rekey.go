package ledger

import "slices"

// Rekeyed returns t with every posting to a holding of from moved to the
// same holding of to, and whether it posts to from at all (D-214): the new
// version of t that re-keys an instrument whose ISIN was provisional to the
// registry's. Its key, account, kind, dates, description and amounts stay,
// so importing it supersedes t (D-18) and changes nothing else; a deletion,
// which has no postings, never posts to from.
func Rekeyed(t *Transaction, from, to Instrument) (Transaction, bool) {
	out := *t
	out.Postings = slices.Clone(t.Postings)
	moved := false
	for i := range out.Postings {
		b := &out.Postings[i].Book
		if b.Kind == BookHolding && b.Instrument == from {
			b.Instrument = to
			moved = true
		}
	}
	if !moved {
		return *t, false
	}
	return out, true
}

// RekeyedNames returns the display names to append so that to shows the
// names from shows (D-197, D-214): the first one a connector delivered of
// from, while to has none delivered, and the owner's latest rename of
// from, while the owner never renamed to. records are every stored name,
// in the order stored. Appending them again finds to named already, so it
// returns nothing.
func RekeyedNames(records []InstrumentName, from, to ISIN) []InstrumentName {
	var out []InstrumentName
	for _, byOwner := range []bool{false, true} {
		if slices.ContainsFunc(records, func(r InstrumentName) bool { return r.ISIN.Equal(to) && r.ByOwner == byOwner }) {
			continue
		}
		if n, ok := nameOf(records, from, byOwner); ok {
			n.ISIN = to
			out = append(out, n)
		}
	}
	return out
}

// nameOf returns, of records, the first name a connector delivered of
// isin, or with byOwner the owner's latest, and whether there is one.
func nameOf(records []InstrumentName, isin ISIN, byOwner bool) (InstrumentName, bool) {
	var (
		out   InstrumentName
		found bool
	)
	for _, r := range records {
		if r.ISIN.Equal(isin) && r.ByOwner == byOwner && (byOwner || !found) {
			out, found = r, true
		}
	}
	return out, found
}
