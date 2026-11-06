package views

import (
	"cmp"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// PostsTo reports whether t posts to a holding of an instrument with ISIN
// isin.
func PostsTo(t *ledger.Transaction, isin string) bool {
	return slices.ContainsFunc(t.Postings, func(p ledger.Posting) bool {
		return p.Book.Kind == ledger.BookHolding && p.Book.Instrument.ISIN.String() == isin
	})
}

// InstrumentTransactions returns those of txs that post to a holding of an
// instrument with ISIN isin, in their order: the instrument page's
// transactions across accounts (D-195).
func InstrumentTransactions(txs []ledger.Transaction, isin string) []ledger.Transaction {
	var out []ledger.Transaction
	for i := range txs {
		if PostsTo(&txs[i], isin) {
			out = append(out, txs[i])
		}
	}
	return out
}

// InstrumentListings returns the listings of the instrument with ISIN isin
// that txs post to, each once, in symbol order: the venue symbols it is
// held under, such as VWCE.
func InstrumentListings(txs []ledger.Transaction, isin string) []ledger.Instrument {
	seen := map[string]bool{}
	var out []ledger.Instrument
	for i := range txs {
		for j := range txs[i].Postings {
			b := txs[i].Postings[j].Book
			if b.Kind == ledger.BookHolding && b.Instrument.ISIN.String() == isin && !seen[b.Instrument.Symbol] {
				seen[b.Instrument.Symbol] = true
				out = append(out, b.Instrument)
			}
		}
	}
	slices.SortFunc(out, func(a, b ledger.Instrument) int { return cmp.Compare(a.Symbol, b.Symbol) })
	return out
}

// InstrumentClass is an instrument's exemption class in one wrapper's
// yearly tax, such as OKI's (D-127, D-132): the owner's choice when they
// made one, otherwise the pack's. Known is false when the pack can't tell
// and the owner hasn't chosen: the instrument then counts in Class, the
// one the pack falls back to, until the owner classifies it.
type InstrumentClass struct {
	Wrapper     jurisdiction.Wrapper
	Class       jurisdiction.ExemptionClass
	OwnerChosen bool
	Known       bool
}

// InstrumentClasses returns asset's exemption class in every wrapper of
// pack whose yearly tax tells assets apart, in the pack's wrapper order;
// owner holds the owner's class of each instrument they classified, by
// ISIN.
func InstrumentClasses(pack jurisdiction.Pack, asset *jurisdiction.Asset, owner map[string]jurisdiction.ExemptionClass) []InstrumentClass {
	var out []InstrumentClass
	for _, w := range pack.Wrappers() {
		if len(pack.ExemptionClasses(w)) == 0 {
			continue
		}
		if c, ok := owner[asset.Instrument]; ok {
			out = append(out, InstrumentClass{Wrapper: w, Class: c, OwnerChosen: true, Known: true})
			continue
		}
		c, known := pack.ExemptionClass(w, asset)
		out = append(out, InstrumentClass{Wrapper: w, Class: c, Known: known})
	}
	return out
}

// InstrumentPrice is a stored price of one listing of an instrument, from
// one source (D-94), with whether that source is the one its listing is
// valued at.
type InstrumentPrice struct {
	valuation.Price

	Primary bool
}

// MarkPrimary returns prices, each marked primary when its source is the
// one sources names for its listing (valuation.Price.Instrument); a listing
// sources doesn't name has no primary source.
func MarkPrimary(prices []valuation.Price, sources map[string]string) []InstrumentPrice {
	out := make([]InstrumentPrice, len(prices))
	for i := range prices {
		src, ok := sources[prices[i].Instrument]
		out[i] = InstrumentPrice{Price: prices[i], Primary: ok && src == prices[i].Source}
	}
	return out
}

// ListingOf returns the listing a price names in valuation.Price.Instrument,
// "<ISIN>/<symbol>"; false when it names none.
func ListingOf(s string) (ledger.Instrument, bool) {
	code, symbol, found := strings.Cut(s, "/")
	isin, err := ledger.ParseISIN(code)
	if !found || err != nil || symbol == "" {
		return ledger.Instrument{}, false
	}
	return ledger.Instrument{ISIN: isin, Symbol: symbol}, true
}

// InstrumentHolding is one account's holding of one listing of an
// instrument on a date, with its open lots of that listing.
type InstrumentHolding struct {
	Account ledger.AccountID
	Holding ledger.Holding
	Lots    []ledger.Lot
}

// HoldingsOf returns account's holdings of the listings of isin among
// holdings, in their order, each with those of lots that are of its
// listing, in their order.
func HoldingsOf(account ledger.AccountID, holdings []ledger.Holding, lots []ledger.Lot, isin string) []InstrumentHolding {
	var out []InstrumentHolding
	for i := range holdings {
		h := holdings[i]
		if h.Instrument.ISIN.String() != isin {
			continue
		}
		ih := InstrumentHolding{Account: account, Holding: h}
		for j := range lots {
			if lots[j].Instrument == h.Instrument {
				ih.Lots = append(ih.Lots, lots[j])
			}
		}
		out = append(out, ih)
	}
	return out
}

// AssetOf returns what a pack classifies for a yearly tax (D-127) of the
// instrument held as in and priced at price: an accrued asset for a
// treasury bond, a listed one otherwise, in the price's currency.
func AssetOf(in ledger.Instrument, price *valuation.Price) jurisdiction.Asset {
	kind := jurisdiction.AssetListed
	if AssetClassOf(in) == AssetTreasuryBond {
		kind = jurisdiction.AssetAccrued
	}
	return jurisdiction.Asset{Kind: kind, Instrument: in.ISIN.String(), Currency: price.Close.Currency()}
}
