package ingest

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Errors of tickers and ticker entries.
var (
	// ErrInvalidTicker reports a ticker without a valid symbol or country.
	ErrInvalidTicker = errors.New("invalid ticker")
	// ErrInvalidTickerEntry reports a ticker entry that books no valid
	// transaction, such as a dividend with a quantity or a buy for nothing.
	ErrInvalidTickerEntry = errors.New("invalid ticker entry")
)

// Bounds of a ticker's fields.
const (
	maxTickerSymbolLen = 32
	maxTickerNameLen   = 200
	countryCodeLen     = 2
)

// Ticker is an instrument as an institution that reports no ISIN names it,
// such as a ticker of an xStation export (D-116): its symbol there, which
// the ledger keeps as the instrument's venue symbol, its code on its market,
// the ISO 3166-1 alpha-2 code of the market's country when the symbol says,
// and its display name. The owner confirms its ISIN once (TickerConfirmation).
type Ticker struct {
	Symbol  string
	Code    string
	Country string
	Name    string
}

// Validate reports whether t has a symbol of 1 to 32 printable characters
// without spaces, a country of two upper-case letters or none, and a name
// of at most 200 characters.
func (t *Ticker) Validate() error {
	if err := validTickerSymbol(t.Symbol); err != nil {
		return err
	}
	if t.Country != "" && (len(t.Country) != countryCodeLen || strings.ToUpper(t.Country) != t.Country ||
		!unicode.IsLetter(rune(t.Country[0])) || !unicode.IsLetter(rune(t.Country[1]))) {
		return fmt.Errorf("%w %s: country must be two upper-case letters", ErrInvalidTicker, t.Symbol)
	}
	if len(t.Name) > maxTickerNameLen || len(t.Code) > maxTickerSymbolLen {
		return fmt.Errorf("%w %s: code or name too long", ErrInvalidTicker, t.Symbol)
	}
	return nil
}

// validTickerSymbol reports whether s can be a ticker's symbol.
func validTickerSymbol(s string) error {
	if s == "" || len(s) > maxTickerSymbolLen {
		return fmt.Errorf("%w: symbol must have 1 to %d characters", ErrInvalidTicker, maxTickerSymbolLen)
	}
	for _, c := range s {
		if !unicode.IsPrint(c) || unicode.IsSpace(c) {
			return fmt.Errorf("%w: symbol must be printable characters without spaces", ErrInvalidTicker)
		}
	}
	return nil
}

// TickerEntry is a buy, sell, dividend or opening lot of an instrument an
// institution named by ticker only (D-116), such as a position of an
// xStation export that seeds an account (D-136). The core books it with the
// ISIN the owner confirmed for the ticker (Transaction), and until then
// holds it as an ingest problem. Amount is a buy's cost, a sale's proceeds,
// a dividend's gross amount or an opening lot's cost; Charge is a buy's or
// sale's fee, or the tax withheld from a dividend, zero for none and always
// zero for an opening lot. Opened is the date an opening lot's units were
// first acquired, and zero for any other kind. An opening lot whose
// institution gives its open price in its instrument's trading currency,
// which it doesn't name, has no Amount and its OpenPrice per unit: its cost
// is booked in the trading currency of the confirmed listing (D-159).
// Ref names its statement row, as Problem.Ref does.
type TickerEntry struct {
	Symbol      string
	Header      ledger.TransactionHeader
	Kind        ledger.TransactionKind
	Quantity    kernel.Quantity
	Amount      kernel.Money
	Charge      kernel.Money
	Opened      kernel.Date
	OpenPrice   kernel.Decimal
	Description string
	Ref         string
}

// errNotTickerKind reports a ticker entry of a kind other than a buy, sell
// or dividend.
var errNotTickerKind = errors.New("not a ticker entry kind")

// checkISIN is a well-formed ISIN of no country (XX), never a real
// instrument's: Validate books an entry with it to check its amounts.
const checkISIN = "XX0000000002"

// Transaction returns the transaction e books once its ticker's listing
// l is confirmed (D-116): the instrument is l's ISIN with the ticker's
// symbol, and an opening lot priced by its open price costs its quantity
// times that price, in l's trading currency's major unit, rounded half to
// even to the currency's minor unit (D-159). Such a lot on a listing with
// no known trading currency fails with ErrUnknownTradingCurrency.
func (e *TickerEntry) Transaction(l Listing) (ledger.Transaction, error) {
	if reason := e.misfit(); reason != "" {
		return ledger.Transaction{}, fmt.Errorf("%w %s: %s", ErrInvalidTickerEntry, e.Header.Key, reason)
	}
	c := *e
	if e.PricedByListing() {
		cost, err := e.listingCost(l)
		if err != nil {
			return ledger.Transaction{}, err
		}
		c.Amount = cost
	}
	t, err := c.build(ledger.Instrument{ISIN: l.ISIN, Symbol: e.Symbol})
	if err != nil {
		return ledger.Transaction{}, err
	}
	t.Description = e.Description
	return t, nil
}

// PricedByListing reports whether e is an opening lot priced by its open
// price in its instrument's trading currency (D-159), with no Amount.
func (e *TickerEntry) PricedByListing() bool {
	return e.Kind == ledger.KindOpeningLot && e.Amount.Currency().IsZero()
}

// checkCurrency is the trading currency Validate books a lot priced by
// listing in, to check its amounts.
const checkCurrency = "PLN"

// Validate reports whether e has a valid symbol and books a valid
// transaction with any confirmed listing. Its account may still be empty
// when a broker account names it.
func (e *TickerEntry) Validate() error {
	if err := validTickerSymbol(e.Symbol); err != nil {
		return err
	}
	isin, err := ledger.ParseISIN(checkISIN)
	if err != nil {
		return fmt.Errorf("check ISIN: %w", err)
	}
	currency, err := kernel.ParseCurrency(checkCurrency)
	if err != nil {
		return fmt.Errorf("check currency: %w", err)
	}
	h := e.Header
	if h.Account == "" {
		h.Account = "broker-account"
	}
	c := *e
	c.Header = h
	_, err = c.Transaction(Listing{ISIN: isin, Currency: currency})
	return err
}

// misfit returns why e's fields don't fit its kind, or "": only an opening
// lot has a lot open date or an open price, a dividend has no quantity, an
// opening lot no charge, and an open price is not negative.
func (e *TickerEntry) misfit() string {
	switch {
	case e.Kind != ledger.KindOpeningLot && !e.OpenPrice.IsZero():
		return "only an opening lot has an open price"
	case e.OpenPrice.Sign() < 0:
		return "a negative open price"
	case e.Kind != ledger.KindOpeningLot && !e.Opened.IsZero():
		return "only an opening lot has a lot open date"
	case e.Kind == ledger.KindDividend && !e.Quantity.IsZero():
		return "a dividend has no quantity"
	case e.Kind == ledger.KindOpeningLot && !e.Charge.IsZero():
		return "an opening lot has no charge"
	}
	return ""
}

// build books e's kind of transaction of in. A kind a ticker entry can't
// be, or a transaction the ledger refuses, fails with ErrInvalidTickerEntry.
func (e *TickerEntry) build(in ledger.Instrument) (ledger.Transaction, error) {
	var (
		t   ledger.Transaction
		err = errNotTickerKind
	)
	switch e.Kind {
	case ledger.KindBuy:
		t, err = ledger.NewBuy(e.Header, in, e.Quantity, e.Amount, e.Charge)
	case ledger.KindSell:
		t, err = ledger.NewSell(e.Header, in, e.Quantity, e.Amount, e.Charge)
	case ledger.KindDividend:
		t, err = ledger.NewDividend(e.Header, in, e.Amount, e.Charge)
	case ledger.KindOpeningLot:
		t, err = ledger.NewOpeningLot(e.Header, in, e.Quantity, e.Amount, e.Opened)
	case ledger.KindFee, ledger.KindFXConversion, ledger.KindDeposit, ledger.KindWithdrawal,
		ledger.KindInterest, ledger.KindDeletion, ledger.KindRedemption:
	}
	if errors.Is(err, errNotTickerKind) {
		return ledger.Transaction{}, fmt.Errorf("%w %s: kind %s is not a buy, sell, dividend or opening lot", ErrInvalidTickerEntry, e.Header.Key, e.Kind)
	}
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("%w: %w", ErrInvalidTickerEntry, err)
	}
	return t, nil
}

// TickerConfirmation is the ISIN the owner confirmed for a ticker of one
// connector instance (D-116): every entry of the ticker is booked with it.
// Currency is the trading currency code the owner picked, such as "GBX",
// when the listing names none and the ticker's exchange quotes in more
// than one (CurrencyChoices, D-170), and "" otherwise.
type TickerConfirmation struct {
	Instance string
	Symbol   string
	ISIN     ledger.ISIN
	Currency string
}

// ISINProposal is the ISIN a market-data connector found for a ticker's
// code and market, with the instrument's name there, for the owner to
// confirm (D-116), and the currency the listing found trades in, as the
// connector names it, such as "GBX", or empty when it names none (D-159).
type ISINProposal struct {
	Symbol   string
	ISIN     ledger.ISIN
	Name     string
	Currency string
}

// ConfirmedName returns the display name an instrument takes when the
// owner confirms isin for p (D-197): the institution's own name of the
// ticker, or else the market-data lookup's name when isin is the one it
// proposed; ok is false when neither names it.
func ConfirmedName(p *PendingTicker, isin ledger.ISIN) (string, bool) {
	switch {
	case p.Ticker.Name != "":
		return p.Ticker.Name, true
	case p.Proposal.Name != "" && p.Proposal.ISIN.Equal(isin):
		return p.Proposal.Name, true
	}
	return "", false
}

// PendingTicker is a ticker of a connector instance whose ISIN the owner
// hasn't confirmed yet, with the market-data connector's proposal, if it
// found one (Proposal.ISIN not zero), and how many entries wait for it.
// Proposal.Currency is the proposed listing's trading currency, if it
// names one.
type PendingTicker struct {
	Instance string
	Ticker   Ticker
	Proposal ISINProposal
	Waiting  int
}

// CurrencyChoices returns the trading currency codes the owner picks one
// of when confirming p with an ISIN whose listing names no currency
// (D-170): its exchange's, when that exchange quotes in more than one, and
// none otherwise.
func (p *PendingTicker) CurrencyChoices() []string {
	if codes := ExchangeCurrencies(p.Ticker.Country); len(codes) > 1 {
		return codes
	}
	return nil
}

// BookTickerEntries returns m with each ticker entry whose symbol confirmed
// maps to a listing booked as a transaction (TickerEntry.Transaction), and
// the entries that wait for the owner to confirm their ticker's ISIN, each
// also an ingest problem in the result (D-116). An entry that books no
// valid transaction, or an opening lot whose trading currency is unknown
// (D-159), is an ingest problem too and doesn't wait. m's ticker entries must have their account.
func BookTickerEntries(m *Mapping, confirmed map[string]Listing) (Mapping, []TickerEntry) {
	out := *m
	out.Transactions = append([]ledger.Transaction(nil), m.Transactions...)
	out.Problems = append([]Problem(nil), m.Problems...)
	out.TickerEntries = nil
	var waiting []TickerEntry
	for i := range m.TickerEntries {
		e := &m.TickerEntries[i]
		listing, ok := confirmed[e.Symbol]
		if !ok {
			waiting = append(waiting, *e)
			out.Problems = append(out.Problems, WaitingForISIN(e))
			continue
		}
		t, err := e.Transaction(listing)
		switch {
		case errors.Is(err, ErrUnknownTradingCurrency):
			out.Problems = append(out.Problems, Unsupported(e.Ref, "the trading currency of "+e.Symbol+" is unknown"))
			continue
		case err != nil:
			out.Problems = append(out.Problems, Unsupported(e.Ref, "not a valid ledger transaction"))
			continue
		}
		out.Transactions = append(out.Transactions, t)
	}
	return out, waiting
}

// WaitingForISIN is the ingest problem of entry e while its ticker's ISIN
// is not confirmed.
func WaitingForISIN(e *TickerEntry) Problem {
	return Unsupported(e.Ref, "waits for the owner to confirm the ISIN of "+e.Symbol)
}

// listingCost returns the cost of e, an opening lot priced by listing, on
// listing l: its quantity times its open price in l's major unit, rounded
// half to even to the currency's minor unit.
func (e *TickerEntry) listingCost(l Listing) (kernel.Money, error) {
	if l.Currency.IsZero() {
		return kernel.Money{}, fmt.Errorf("%w: %s", ErrUnknownTradingCurrency, e.Symbol)
	}
	cost, err := l.Price(e.OpenPrice).Mul(e.Quantity.Decimal()).Round(CurrencyPlaces(l.Currency), kernel.RoundHalfEven)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("%w %s: cost: %w", ErrInvalidTickerEntry, e.Header.Key, err)
	}
	return kernel.NewMoney(cost, l.Currency), nil
}
