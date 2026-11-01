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

// TickerEntry is a buy, sell or dividend of an instrument an institution
// named by ticker only (D-116). The core books it with the ISIN the owner
// confirmed for the ticker (Transaction), and until then holds it as an
// ingest problem. Amount is a buy's cost, a sale's proceeds or a dividend's
// gross amount; Charge is a buy's or sale's fee, or the tax withheld from a
// dividend, zero for none. Ref names its statement row, as Problem.Ref does.
type TickerEntry struct {
	Symbol      string
	Header      ledger.TransactionHeader
	Kind        ledger.TransactionKind
	Quantity    kernel.Quantity
	Amount      kernel.Money
	Charge      kernel.Money
	Description string
	Ref         string
}

// errNotTickerKind reports a ticker entry of a kind other than a buy, sell
// or dividend.
var errNotTickerKind = errors.New("not a ticker entry kind")

// checkISIN is a well-formed ISIN of no country (XX), never a real
// instrument's: Validate books an entry with it to check its amounts.
const checkISIN = "XX0000000002"

// Transaction returns the transaction e books once its ticker's ISIN is
// confirmed as isin: the instrument is isin with the ticker's symbol.
func (e *TickerEntry) Transaction(isin ledger.ISIN) (ledger.Transaction, error) {
	in := ledger.Instrument{ISIN: isin, Symbol: e.Symbol}
	var (
		t   ledger.Transaction
		err error
	)
	switch e.Kind {
	case ledger.KindBuy:
		t, err = ledger.NewBuy(e.Header, in, e.Quantity, e.Amount, e.Charge)
	case ledger.KindSell:
		t, err = ledger.NewSell(e.Header, in, e.Quantity, e.Amount, e.Charge)
	case ledger.KindDividend:
		if !e.Quantity.IsZero() {
			return ledger.Transaction{}, fmt.Errorf("%w %s: a dividend has no quantity", ErrInvalidTickerEntry, e.Header.Key)
		}
		t, err = ledger.NewDividend(e.Header, in, e.Amount, e.Charge)
	case ledger.KindFee, ledger.KindFXConversion, ledger.KindDeposit, ledger.KindWithdrawal,
		ledger.KindOpeningLot, ledger.KindInterest, ledger.KindDeletion, ledger.KindRedemption:
		err = errNotTickerKind
	}
	if t.Kind == 0 && err == nil {
		err = errNotTickerKind
	}
	if errors.Is(err, errNotTickerKind) {
		return ledger.Transaction{}, fmt.Errorf("%w %s: kind %s is not a buy, sell or dividend", ErrInvalidTickerEntry, e.Header.Key, e.Kind)
	}
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("%w: %w", ErrInvalidTickerEntry, err)
	}
	t.Description = e.Description
	return t, nil
}

// Validate reports whether e has a valid symbol and books a valid
// transaction with any confirmed ISIN. Its account may still be empty when
// a broker account names it.
func (e *TickerEntry) Validate() error {
	if err := validTickerSymbol(e.Symbol); err != nil {
		return err
	}
	isin, err := ledger.ParseISIN(checkISIN)
	if err != nil {
		return fmt.Errorf("check ISIN: %w", err)
	}
	h := e.Header
	if h.Account == "" {
		h.Account = "broker-account"
	}
	c := *e
	c.Header = h
	_, err = c.Transaction(isin)
	return err
}

// TickerConfirmation is the ISIN the owner confirmed for a ticker of one
// connector instance (D-116): every entry of the ticker is booked with it.
type TickerConfirmation struct {
	Instance string
	Symbol   string
	ISIN     ledger.ISIN
}

// ISINProposal is the ISIN a market-data connector found for a ticker's
// code and market, with the instrument's name there, for the owner to
// confirm (D-116).
type ISINProposal struct {
	Symbol string
	ISIN   ledger.ISIN
	Name   string
}

// PendingTicker is a ticker of a connector instance whose ISIN the owner
// hasn't confirmed yet, with the market-data connector's proposal, if it
// found one (Proposal.ISIN not zero), and how many entries wait for it.
type PendingTicker struct {
	Instance string
	Ticker   Ticker
	Proposal ISINProposal
	Waiting  int
}

// BookTickerEntries returns m with each ticker entry whose symbol confirmed
// maps to an ISIN booked as a transaction, and the entries that wait for the
// owner to confirm their ticker's ISIN, each also an ingest problem in the
// result (D-116). An entry that books no valid transaction is an ingest
// problem too and doesn't wait. m's ticker entries must have their account.
func BookTickerEntries(m *Mapping, confirmed map[string]ledger.ISIN) (Mapping, []TickerEntry) {
	out := *m
	out.Transactions = append([]ledger.Transaction(nil), m.Transactions...)
	out.Problems = append([]Problem(nil), m.Problems...)
	out.TickerEntries = nil
	var waiting []TickerEntry
	for i := range m.TickerEntries {
		e := &m.TickerEntries[i]
		isin, ok := confirmed[e.Symbol]
		if !ok {
			waiting = append(waiting, *e)
			out.Problems = append(out.Problems, WaitingForISIN(e))
			continue
		}
		t, err := e.Transaction(isin)
		if err != nil {
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
