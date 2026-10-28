package ledger

import (
	"errors"
	"fmt"
	"sort"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// LotCostPlaces is the number of decimal places the cost of a partly
// consumed lot is split at. The consumed part's cost is the lot's cost times
// the consumed share of its units, rounded half-even to this many places; the
// part left keeps the exact rest, and the last units consumed take whatever
// cost is left. So no cost is ever created or lost: the consumed and remaining
// costs always add up to the lot's original cost.
const LotCostPlaces int32 = 2

// Lot is a quantity of an instrument acquired on a date at a cost. Lots are
// consumed FIFO. An opening lot was seeded from a broker's current positions
// at go-live (D-23) instead of from full history.
type Lot struct {
	Instrument Instrument
	// Opened is the date the units were acquired; for an opening lot, the
	// broker-reported open date.
	Opened kernel.Date
	// Quantity is the number of units still in the lot.
	Quantity kernel.Quantity
	// Cost is the cost basis of those units, in the currency they were paid in.
	Cost kernel.Money
	// Opening reports an opening lot.
	Opening bool
	// Origin is the transaction that acquired the lot.
	Origin TransactionKey
}

// OversellError reports a disposal of more units than the account holds.
// Account, Date and Disposal are set when it comes from LotsOn.
type OversellError struct {
	Account    AccountID
	Instrument Instrument
	Date       kernel.Date
	Disposal   TransactionKey
	Requested  kernel.Quantity
	Available  kernel.Quantity
}

// Error names the instrument and both quantities.
func (e *OversellError) Error() string {
	return fmt.Sprintf("disposal %s of %s units of %s on %s, but only %s held",
		e.Disposal, e.Requested, e.Instrument, e.Date, e.Available)
}

// LotConsumption is what ConsumeFIFO took from lots and what it left, both
// oldest first. A lot consumed in part appears in both.
type LotConsumption struct {
	Consumed  []Lot
	Remaining []Lot
}

// ConsumeFIFO takes qty units from lots of one instrument, oldest first by
// open date, then by origin key. A lot consumed in part has its cost split as
// LotCostPlaces describes. Taking more units than the lots hold fails with an
// *OversellError and consumes nothing.
func ConsumeFIFO(lots []Lot, qty kernel.Quantity) (LotConsumption, error) {
	if qty.Sign() <= 0 {
		return LotConsumption{}, fmt.Errorf("%w: disposal of %s units", ErrInvalidTransaction, qty)
	}
	remaining := sortedLots(lots)
	var available kernel.Quantity
	for i := range remaining {
		available = available.Add(remaining[i].Quantity)
	}
	if qty.Cmp(available) > 0 {
		e := &OversellError{Requested: qty, Available: available}
		if len(lots) > 0 {
			e.Instrument = lots[0].Instrument
		}
		return LotConsumption{}, e
	}
	var consumed []Lot
	left := qty
	for len(remaining) > 0 && left.Sign() > 0 {
		lot := &remaining[0]
		if lot.Quantity.Cmp(left) <= 0 {
			consumed = append(consumed, *lot)
			left = left.Sub(lot.Quantity)
			remaining = remaining[1:]
			continue
		}
		share, err := consumedCost(lot, left)
		if err != nil {
			return LotConsumption{}, err
		}
		part := *lot
		part.Quantity, part.Cost = left, share
		consumed = append(consumed, part)
		lot.Quantity = lot.Quantity.Sub(left)
		lot.Cost = kernel.NewMoney(lot.Cost.Amount().Sub(share.Amount()), lot.Cost.Currency())
		left = kernel.Quantity{}
	}
	return LotConsumption{Consumed: consumed, Remaining: remaining}, nil
}

// consumedCost returns the cost of units taken from lot, which holds more
// than units, rounded as LotCostPlaces describes. The caller leaves the exact
// rest in the lot.
func consumedCost(lot *Lot, units kernel.Quantity) (kernel.Money, error) {
	share, err := lot.Cost.Amount().Mul(units.Decimal()).Quo(lot.Quantity.Decimal(), LotCostPlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Money{}, fmt.Errorf("splitting lot %s: %w", lot.Origin, err)
	}
	return kernel.NewMoney(share, lot.Cost.Currency()), nil
}

// sortedLots returns a copy of lots, oldest first by open date, then by
// origin key.
func sortedLots(lots []Lot) []Lot {
	out := append([]Lot(nil), lots...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := &out[i], &out[j]
		if c := a.Opened.Compare(b.Opened); c != 0 {
			return c < 0
		}
		if a.Origin.Source != b.Origin.Source {
			return a.Origin.Source < b.Origin.Source
		}
		return a.Origin.ExternalID < b.Origin.ExternalID
	})
	return out
}

// LotsOn returns account's open lots of the instrument in at the end of on,
// oldest first, derived from the current versions of account's transactions: each
// holding posting that adds units opens a lot at its amount, and each that
// removes units consumes lots FIFO. On one date, acquisitions are applied
// before disposals, so the order of a day's rows never causes a false
// oversell. A disposal of more units than held fails with an *OversellError.
func LotsOn(log []VersionedTransaction, account AccountID, in Instrument, on kernel.Date) ([]Lot, error) {
	txs, err := currentInAccount(log, account, on)
	if err != nil {
		return nil, err
	}
	var lots []Lot
	for start := 0; start < len(txs); {
		end := start
		for end < len(txs) && txs[end].Date.Equal(txs[start].Date) {
			end++
		}
		lots = append(lots, acquisitions(txs[start:end], in)...)
		if lots, err = applyDisposals(lots, txs[start:end], in); err != nil {
			return nil, err
		}
		start = end
	}
	return sortedLots(lots), nil
}

// acquisitions returns the lots that txs open for in.
func acquisitions(txs []Transaction, in Instrument) []Lot {
	var lots []Lot
	for i := range txs {
		t := &txs[i]
		for j := range t.Postings {
			if p := &t.Postings[j]; p.Book == HoldingBook(in) && p.Quantity.Sign() > 0 {
				lot := Lot{Instrument: in, Opened: t.Date, Quantity: p.Quantity, Cost: p.Amount, Origin: t.Key}
				if t.Kind == KindOpeningLot {
					lot.Opened, lot.Opening = t.LotOpened, true
				}
				lots = append(lots, lot)
			}
		}
	}
	return lots
}

// applyDisposals consumes lots FIFO for every posting in txs that removes
// units of in.
func applyDisposals(lots []Lot, txs []Transaction, in Instrument) ([]Lot, error) {
	for i := range txs {
		t := &txs[i]
		for j := range t.Postings {
			p := &t.Postings[j]
			if p.Book != HoldingBook(in) || p.Quantity.Sign() >= 0 {
				continue
			}
			taken, err := ConsumeFIFO(lots, p.Quantity.Neg())
			if oversell, ok := errors.AsType[*OversellError](err); ok {
				oversell.Account, oversell.Instrument, oversell.Date, oversell.Disposal = t.Account, in, t.Date, t.Key
				return nil, oversell
			}
			if err != nil {
				return nil, err
			}
			lots = taken.Remaining
		}
	}
	return lots, nil
}
