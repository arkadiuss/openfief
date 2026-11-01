package xtb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// The Open Positions columns an opening batch reads (R-24).
const (
	colPositionName = "Instrument/Position"
	colPositionType = "Type"
	colVolume       = "Volume"
	colValue        = "Value"
	colCurrentPrice = "Current price"
	colOpenPrice    = "Open price"
	colOpenTime     = "Open time (UTC)"
)

// openingColumns are the Open Positions columns an opening batch needs.
func openingColumns() []string {
	return []string{colTicker, colCategory, colVolume, colValue, colCurrentPrice, colOpenPrice, colOpenTime}
}

// openingLotPrefix starts the external ID of every opening lot the mapper
// writes, so it never meets an operation ID of Cash Operations.
const openingLotPrefix = "open-position:"

// pricePlaces are the decimal places a price is rounded to when it differs
// from that rounding by a floating-point artifact only (R-24).
const pricePlaces = 8

// openingDate returns the account's go-live date in opening mode: the
// date of Open Positions' report, in UTC (D-136). The export must have
// Open Positions with the columns an opening lot needs.
func openingDate(x *export) (kernel.Date, error) {
	if x.open == nil {
		return kernel.Date{}, fmt.Errorf("%w: no sheet %q", ErrOpeningMode, sheetOpen)
	}
	if !hasAll(x.open.cols, openingColumns()) {
		return kernel.Date{}, fmt.Errorf("%w: sheet %q has no header row with %s", ErrOpeningMode, sheetOpen, strings.Join(openingColumns(), ", "))
	}
	day, ok := reportDate(x.open.sheet, x.open.header, x.date1904)
	if !ok {
		return kernel.Date{}, fmt.Errorf("%w: sheet %q has no %q date", ErrOpeningMode, sheetOpen, labelReport)
	}
	return day, nil
}

// opening maps an opening batch (D-136): each open stock or ETF position
// to an opening lot on the go-live date, at its open time and price, and
// cash, the balance the owner entered, to the account's opening cash. Open
// CFD positions are flagged (D-24). Cash Operations and Closed Positions
// are not read: the positions and the cash hold their effect.
func (m *mapping) opening(cash kernel.Money) error {
	if err := m.openingCash(cash); err != nil {
		return err
	}
	seen := map[string]int{}
	for _, rec := range m.x.open.records() {
		if c := rec.text(colCategory); !stockCategory(c) {
			m.problem(ingest.Unsupported(m.positionRef(sheetOpen, rec),
				fmt.Sprintf("open CFD position (category %q) is not supported (D-24)", shown(c))))
			continue
		}
		e, err := m.openingLot(rec, seen)
		if err == nil {
			err = m.addEntry(rec.text(colPositionName), &e)
		}
		if err != nil {
			m.problem(problemOf(m.positionRef(sheetOpen, rec), err))
		}
	}
	return nil
}

// openingCash books cash, the cash balance the owner entered, as the
// account's opening cash on the go-live date (ledger.NewOpeningCash). None
// entered (the zero Money) and a zero balance book nothing; a balance in
// another currency than the account's, or with more decimals than its
// minor unit, fails with ErrOpeningCash.
func (m *mapping) openingCash(cash kernel.Money) error {
	if cash.Currency().IsZero() {
		return nil
	}
	if !cash.Currency().Equal(m.x.currency) {
		return fmt.Errorf("%w: it must be in the account's currency, %s", ErrOpeningCash, m.x.currency)
	}
	rounded, err := cash.Round(m.x.places, kernel.RoundDown)
	if err != nil || !rounded.Equal(cash) {
		return fmt.Errorf("%w: it has more decimals than %s's minor unit", ErrOpeningCash, m.x.currency)
	}
	if cash.IsZero() {
		return nil
	}
	h := ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: BrokerSource, ExternalID: ledger.OpeningCashID(cash.Currency(), m.x.goLive)},
		Account: standIn, Date: m.x.goLive,
	}
	t, err := ledger.NewOpeningCash(h, cash)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpeningCash, err)
	}
	t.Account, t.Description = "", "opening cash"
	m.entries.Transactions = append(m.entries.Transactions, t)
	return nil
}

// openingLot returns the opening lot of an open stock or ETF position:
// its volume, first acquired on its open date, at its open price per unit.
// The open price is in the instrument's trading currency, which the export
// doesn't name (R-24), so the lot carries the price and no cost: the core
// books its cost, volume times price, in the trading currency of the
// listing the owner confirms for the ticker, or of the exchange its market
// suffix names (D-159). A position on a market with no exchange the core
// knows can't be priced and is unsupported. seen counts the positions
// keyed so far, for positionKey.
func (m *mapping) openingLot(rec record, seen map[string]int) (ingest.TickerEntry, error) {
	symbol := rec.text(colTicker)
	if symbol == "" {
		return ingest.TickerEntry{}, unsupported("an open position without a ticker is not supported")
	}
	if _, _, ok := ingest.ExchangeCurrency(splitSymbol(symbol).Country); !ok {
		return ingest.TickerEntry{}, unsupported("an open position on a market of unknown currency is not supported")
	}
	volume, err := boughtVolume(rec)
	if err != nil {
		return ingest.TickerEntry{}, err
	}
	price, err := m.price(rec, colOpenPrice)
	if err != nil {
		return ingest.TickerEntry{}, err
	}
	opened, err := rec.instant(colOpenTime, m.x.date1904)
	if err != nil {
		return ingest.TickerEntry{}, err
	}
	openDate, ok := utcDate(opened)
	if !ok {
		return ingest.TickerEntry{}, rec.errorf(colOpenTime)
	}
	key := ledger.TransactionKey{Source: BrokerSource, ExternalID: openingLotPrefix + m.positionKey(rec, symbol, opened, seen)}
	return ingest.TickerEntry{
		Symbol: symbol, Kind: ledger.KindOpeningLot, Quantity: kernel.NewQuantity(volume),
		OpenPrice: price, Opened: openDate,
		Header:      ledger.TransactionHeader{Key: key, Date: m.x.goLive},
		Description: "opening lot " + symbol, Ref: m.positionRef(sheetOpen, rec),
	}, nil
}

// boughtVolume returns the volume of a bought position, which must be
// positive; a sold (short) position is not supported.
func boughtVolume(rec record) (kernel.Decimal, error) {
	if side := strings.ToUpper(rec.text(colPositionType)); side != "" && side != "BUY" {
		return kernel.Decimal{}, unsupported("an open %s position is not supported", shown(side))
	}
	volume, err := rec.decimal(colVolume)
	if err != nil {
		return kernel.Decimal{}, err
	}
	if volume.Sign() <= 0 {
		return kernel.Decimal{}, unsupported("an open position without a positive volume is not supported")
	}
	return volume, nil
}

// price returns the non-negative price in column label, rounded to eight
// decimals when it differs from that rounding by a floating-point artifact
// only (R-24).
func (m *mapping) price(rec record, label string) (kernel.Decimal, error) {
	d, err := rec.decimal(label)
	if err != nil {
		return kernel.Decimal{}, err
	}
	if d.Sign() < 0 {
		return kernel.Decimal{}, rec.errorf(label)
	}
	rounded, err := d.Round(pricePlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, rec.errorf(label)
	}
	if artifact(d.Sub(rounded)) {
		return rounded, nil
	}
	return d, nil
}

// positionKey returns the key of an open position: its position ID when
// the sheet gives one, in a Position ID column or as a whole number in
// Instrument/Position, else its ticker, open time and how many positions
// of the same ticker and open time came before it in the sheet.
func (m *mapping) positionKey(rec record, symbol string, opened time.Time, seen map[string]int) string {
	if rec.t.has(colPositionID) {
		if id, err := rec.id(colPositionID); err == nil && id != "" {
			return id
		}
	}
	if id, err := rec.id(colPositionName); err == nil && id != "" && isDigits(id) {
		return id
	}
	base := symbol + "@" + opened.Format("20060102T150405Z")
	seen[base]++
	return base + "#" + strconv.Itoa(seen[base])
}

// isDigits reports whether s is only ASCII digits.
func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return s != ""
}
