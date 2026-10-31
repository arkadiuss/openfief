package ingest

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidBondLot reports a bond lot that cannot be booked: no purchase
// date, a purchase date after today, a negative margin, or anything
// pl.ValueLot refuses for a lot bought at its price on that date (an
// unknown bond type, a purchase date that doesn't match the series, a
// quantity that is not a positive whole number, a negative rate or fee), a
// purchase neither cash nor exchange, or a deleted lot without a series or
// purchase date.
var ErrInvalidBondLot = errors.New("invalid bond lot")

// BondPurchase is how the bonds of a lot were bought (D-100). The zero
// purchase is a cash purchase.
type BondPurchase int

// The bond purchases.
const (
	// PurchaseCash is a purchase for cash at the nominal, 100 PLN per bond.
	PurchaseCash BondPurchase = iota
	// PurchaseExchange is a purchase by exchange (zamiana) of maturing bonds
	// at the series' exchange price, 99.90 PLN per bond (pl.ExchangePrice).
	PurchaseExchange
)

// String returns "cash" or "exchange", or a placeholder for an invalid
// purchase.
func (p BondPurchase) String() string {
	switch p {
	case PurchaseCash:
		return "cash"
	case PurchaseExchange:
		return "exchange"
	}
	return fmt.Sprintf("BondPurchase(%d)", int(p))
}

// BondLot is a lot of Polish retail treasury bonds the owner entered by hand
// (D-9): its series, how many bonds and when and how they were bought, plus
// the series terms from the issue letter, which valuing the lot needs. The
// bonds were bought for cash at the nominal (pl.Nominal) or by exchange at
// the exchange price (pl.ExchangePrice, D-100), so the lot cost Quantity
// times that price.
type BondLot struct {
	// Terms are the series and its terms.
	Terms     pl.Terms
	Quantity  kernel.Quantity
	Purchased kernel.Date
	Purchase  BondPurchase
}

// BondLotKey names one entered lot by its series and purchase date, which
// make it unique in its account (D-98).
type BondLotKey struct {
	Series    pl.Series
	Purchased kernel.Date
}

// BondLotEntries is what the owner hands a treasury bond connector in one
// batch: the lots to add and the lots to delete (D-100), with the account's
// go-live date, which decides how each lot is booked (D-98).
type BondLotEntries struct {
	// GoLive is the account's go-live date: the date it was seeded on, or
	// today for an account not seeded yet.
	GoLive  kernel.Date
	Lots    []BondLot
	Deleted []BondLotKey
}

// Key returns the lot's key: its series and purchase date.
func (l *BondLot) Key() BondLotKey { return BondLotKey{Series: l.Series(), Purchased: l.Purchased} }

// Series returns the lot's bond series.
func (l *BondLot) Series() pl.Series { return l.Terms.Series }

// Price returns what one bond of the lot cost: the nominal for a cash
// purchase, the exchange price for an exchange (D-100), or false for an
// invalid purchase.
func (l *BondLot) Price() (kernel.Money, bool) {
	switch l.Purchase {
	case PurchaseCash:
		return pl.Nominal(), true
	case PurchaseExchange:
		return pl.ExchangePrice(l.Series().Type), true
	}
	return kernel.Money{}, false
}

// Cost returns what the lot cost: its price per bond times the quantity, in
// PLN; zero for an invalid purchase, which Validate refuses.
func (l *BondLot) Cost() kernel.Money {
	price, _ := l.Price()
	return price.Mul(l.Quantity.Decimal())
}

// Validate reports whether l can be booked as of today: it was bought on a
// date no later than today, for cash or by exchange, its margin is not
// negative, and pl.ValueLot values it on its purchase date at its price.
// Otherwise it returns an error wrapping ErrInvalidBondLot that names the
// rule broken; the error carries the series and dates, never an amount.
func (l *BondLot) Validate(today kernel.Date) error {
	switch {
	case l.Purchased.IsZero():
		return fmt.Errorf("%w: %s has no purchase date", ErrInvalidBondLot, l.Series())
	case l.Purchased.After(today):
		return fmt.Errorf("%w: %s bought on %s, after today %s", ErrInvalidBondLot, l.Series(), l.Purchased, today)
	case l.Terms.Margin.Sign() < 0:
		return fmt.Errorf("%w: %s has a negative margin", ErrInvalidBondLot, l.Series())
	}
	price, ok := l.Price()
	if !ok {
		return fmt.Errorf("%w: %s bought by %s", ErrInvalidBondLot, l.Series(), l.Purchase)
	}
	p := pl.Purchase{Date: l.Purchased, Price: price}
	if _, err := pl.ValueLot(&l.Terms, nil, l.Quantity, p, pl.Taxable, l.Purchased); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidBondLot, err)
	}
	return nil
}

// BondTerms are a treasury bond series' terms as a connector delivers them:
// the instrument they belong to and the terms. Recording them makes the
// core value holdings of the instrument as bonds (valuation's
// RecordBondTerms).
type BondTerms struct {
	Instrument ledger.Instrument
	Terms      pl.Terms
}

// Validate reports whether k names a lot: a known bond type and a purchase
// date. Otherwise it returns an error wrapping ErrInvalidBondLot.
func (k BondLotKey) Validate() error {
	if _, err := k.Series.Type.Structure(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidBondLot, err)
	}
	if k.Purchased.IsZero() {
		return fmt.Errorf("%w: %s has no purchase date", ErrInvalidBondLot, k.Series)
	}
	return nil
}
