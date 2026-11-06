package ingest

import (
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// TickerRows are the ticker entries and ticker assertions of one connector
// instance: those of a delivery that wait for their tickers' ISINs, or
// those it booked, which a correction books anew (D-153).
type TickerRows struct {
	Instance   string
	Entries    []TickerEntry
	Assertions []TickerAssertion
}

// Accounts returns the accounts r's entries and assertions are in, each
// once, in ID order.
func (r *TickerRows) Accounts() []ledger.AccountID {
	out := make([]ledger.AccountID, 0, len(r.Entries)+len(r.Assertions))
	for i := range r.Entries {
		out = append(out, r.Entries[i].Header.Account)
	}
	for i := range r.Assertions {
		out = append(out, r.Assertions[i].Account)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// BookedTickerRows returns the ticker entries of m whose ticker confirmed
// maps to a listing and that book a valid transaction with it, and the
// ticker assertions of those tickers: what a delivery booked of instance's
// tickers, kept so that a correction of a ticker's ISIN books it anew
// (D-153). m's ticker entries and assertions must have their account.
func BookedTickerRows(instance string, m *Mapping, confirmed map[string]Listing) TickerRows {
	out := TickerRows{Instance: instance}
	for i := range m.TickerEntries {
		e := &m.TickerEntries[i]
		l, ok := confirmed[e.Symbol]
		if !ok {
			continue
		}
		if _, err := e.Transaction(l); err == nil {
			out.Entries = append(out.Entries, *e)
		}
	}
	for i := range m.TickerAssertions {
		if _, ok := confirmed[m.TickerAssertions[i].Symbol]; ok {
			out.Assertions = append(out.Assertions, m.TickerAssertions[i])
		}
	}
	return out
}

// ErrTickerNotConfirmed reports a correction of a ticker whose ISIN was
// never confirmed: there is nothing to correct.
var ErrTickerNotConfirmed = errors.New("the ticker's ISIN is not confirmed")

// TickerBooking is what the owner's confirmation or correction of a
// ticker's ISIN books (D-116, D-153): the listing it settles, the one a
// correction replaces (zero for a confirmation), and the rows of the
// ticker, at every instance of the institution, to book with it.
type TickerBooking struct {
	Listing  Listing
	Previous Listing
	Rows     []TickerRows
}

// TickerLedger is what a TickerBooking books into the ledger: the
// transactions, the balance assertions, and those it replaces.
type TickerLedger struct {
	Transactions []ledger.Transaction
	Asserted     []ledger.BalanceAssertion
	Replaced     []ledger.BalanceAssertion
}

// Ledger returns what b books: each row's transaction with b's listing,
// its key scoped to its instance (ScopeToConnector), so a correction's
// transactions are new versions that supersede those booked under the old
// ISIN (D-18); each ticker assertion as the balance assertion of the
// listing's instrument; and, for a correction to another ISIN, the balance
// assertions of the previous instrument those replace, which no longer
// hold. An entry that books no valid transaction fails.
func (b *TickerBooking) Ledger() (TickerLedger, error) {
	var out TickerLedger
	moved := !b.Previous.ISIN.IsZero() && b.Previous.ISIN != b.Listing.ISIN
	for i := range b.Rows {
		r := &b.Rows[i]
		own := make([]ledger.Transaction, 0, len(r.Entries))
		for j := range r.Entries {
			t, err := r.Entries[j].Transaction(b.Listing)
			if err != nil {
				return TickerLedger{}, fmt.Errorf("booking %s: %w", r.Entries[j].Header.Key, err)
			}
			own = append(own, t)
		}
		out.Transactions = append(out.Transactions, ScopeToConnector(r.Instance, own)...)
		for j := range r.Assertions {
			a := &r.Assertions[j]
			out.Asserted = append(out.Asserted, a.Assertion(b.Listing))
			if moved {
				out.Replaced = append(out.Replaced, a.Assertion(b.Previous))
			}
		}
	}
	return out, nil
}

// TickerHold is what a delivery of a connector instance leaves with the
// ticker store (D-116, D-153, D-155): the institution the instance
// reports, every ticker the delivery names, its rows that wait for their
// ticker's ISIN and those it booked.
type TickerHold struct {
	Institution string
	Tickers     []Ticker
	Waiting     TickerRows
	Booked      TickerRows
}

// ConfirmedTicker is a ticker whose ISIN the owner confirmed for every
// account at its institution (D-153): a connector instance of the
// institution that reported it, the confirmed ISIN, the trading currency
// code of its listing (ListingCode, "" for none) and how many of its
// statement rows are booked with it.
type ConfirmedTicker struct {
	Institution     string
	Instance        string
	Ticker          Ticker
	ISIN            ledger.ISIN
	TradingCurrency string
	Booked          int
}

// CurrencyChoices returns the trading currency codes the owner picks one
// of when correcting c to an ISIN whose listing names no currency
// (D-170): its exchange's, when that exchange quotes in more than one,
// and none otherwise.
func (c *ConfirmedTicker) CurrencyChoices() []string {
	if codes := ExchangeCurrencies(c.Ticker.Country); len(codes) > 1 {
		return codes
	}
	return nil
}
