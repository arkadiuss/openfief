package ghostfolio

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ActivityType is the type of a Ghostfolio activity. The view writes only
// these four.
type ActivityType string

// The activity types the view writes.
const (
	TypeBuy      ActivityType = "BUY"
	TypeSell     ActivityType = "SELL"
	TypeDividend ActivityType = "DIVIDEND"
	TypeFee      ActivityType = "FEE"
)

// DataSourceManual is Ghostfolio's data source for an asset priced only by
// the market data sent to it: bonds and fees here (D-64).
const DataSourceManual = "MANUAL"

// CommentPrefix starts the comment of every activity the view writes; the
// transaction's key follows it (D-4).
const CommentPrefix = "openfief:"

// Account is a Ghostfolio account the view owns: the UUID it derived for an
// account of the record (AccountUUID), its name and its currency.
type Account struct {
	ID       string
	Name     string
	Currency kernel.Currency
}

// Activity is a Ghostfolio activity, as the view writes and compares it. ID
// is Ghostfolio's own; a desired activity has none.
type Activity struct {
	ID         string
	AccountID  string
	Comment    string
	Type       ActivityType
	Date       kernel.Date
	Currency   kernel.Currency
	DataSource string
	Symbol     string
	Quantity   kernel.Decimal
	UnitPrice  kernel.Decimal
	Fee        kernel.Decimal
}

// sameAs reports whether a and b agree in every field the view writes: the
// test for an activity that needs no change. Edits keep the comment, so the
// comment alone can't tell (D-64).
func (a *Activity) sameAs(b *Activity) bool {
	return a.AccountID == b.AccountID && a.duplicateOf(b)
}

// duplicateOf reports whether Ghostfolio's import would skip a as a
// duplicate of b: every field but the account and the ID agree.
func (a *Activity) duplicateOf(b *Activity) bool {
	return a.Comment == b.Comment && a.Type == b.Type && a.Date == b.Date &&
		a.Currency.Equal(b.Currency) && a.DataSource == b.DataSource && a.Symbol == b.Symbol &&
		a.Quantity.Equal(b.Quantity) && a.UnitPrice.Equal(b.UnitPrice) && a.Fee.Equal(b.Fee)
}

// AssetProfile is a MANUAL asset the view prices itself: a bond series or
// the fees of one currency, with its full price history.
type AssetProfile struct {
	Symbol        string
	Name          string
	ISIN          string
	Currency      kernel.Currency
	AssetClass    string
	AssetSubClass string
	MarketData    []Price
}

// Price is an asset's price on a date.
type Price struct {
	Date  kernel.Date
	Price kernel.Decimal
}

// namespace returns the UUID namespace of every UUID the view derives,
// f9c83d88-aa46-4ff2-b76f-4a8c1828ab50. It was generated once, at random,
// for OpenFief's Ghostfolio view.
func namespace() uuid.UUID {
	return uuid.UUID{0xf9, 0xc8, 0x3d, 0x88, 0xaa, 0x46, 0x4f, 0xf2, 0xb7, 0x6f, 0x4a, 0x8c, 0x18, 0x28, 0xab, 0x50}
}

// AccountUUID returns the ID of the Ghostfolio account the view instance
// named instance owns for account: a UUID derived from both, so every sync
// finds the same account and no storage is needed (D-64).
func AccountUUID(instance string, account ledger.AccountID) string {
	return uuidV5("account\x00" + instance + "\x00" + string(account))
}

// BondSymbol returns the symbol of the MANUAL asset for the bond series
// with the given code, such as "EDO0536": a UUID derived from the code
// (D-64).
func BondSymbol(series string) string {
	return uuidV5("bond\x00" + series)
}

// FeeSymbol returns the symbol of the MANUAL asset that fees in currency c
// are booked against: a UUID derived from the currency.
func FeeSymbol(c kernel.Currency) string {
	return uuidV5("fee\x00" + c.String())
}

// uuidV5 returns the name-based UUID, version 5 (RFC 9562), of name in
// namespace.
func uuidV5(name string) string {
	return uuid.NewSHA1(namespace(), []byte(name)).String()
}

// wireDigits is the most significant digits a number the view sends may
// have. Ghostfolio stores numbers as IEEE 754 doubles, which hold any
// decimal of up to 15 significant digits exactly enough to print it back
// unchanged, so a number of at most 15 digits compares equal after a round
// trip.
const wireDigits = 15

// ErrNumberTooLarge reports a number with more than 15 digits before the
// decimal point, which no double prints back unchanged.
var ErrNumberTooLarge = errors.New("number too large for Ghostfolio")

// toWire rounds d, half to even, to at most 15 significant digits, so it
// survives Ghostfolio's doubles unchanged.
func toWire(d kernel.Decimal) (kernel.Decimal, error) {
	s := d.String()
	s = strings.TrimPrefix(s, "-")
	intPart, frac, _ := strings.Cut(s, ".")
	var places int
	if intPart != "0" {
		places = wireDigits - len(intPart)
	} else {
		zeros := len(frac) - len(strings.TrimLeft(frac, "0"))
		places = wireDigits + zeros
	}
	if places < 0 {
		return kernel.Decimal{}, fmt.Errorf("%w: %d integer digits", ErrNumberTooLarge, len(intPart))
	}
	if len(frac) <= places {
		return d, nil
	}
	r, err := d.Round(int32(places), kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("rounding for Ghostfolio: %w", err)
	}
	return r, nil
}
