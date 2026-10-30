package views

import (
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// RemovedBehindHoldings returns, for each instrument held in txs at the end
// of on, the removed connector instances behind its holding (D-91): the
// instances of the transactions dated on or before on that post to its
// holding ledger account and that in knows are no longer installed, in
// name order. An instrument with none is left out; so is everything while
// the installed instances are unknown. txs are one account's current
// transactions.
func RemovedBehindHoldings(txs []ledger.Transaction, in Installed, on kernel.Date) map[ledger.Instrument][]string {
	out := map[ledger.Instrument][]string{}
	for i := range txs {
		t := &txs[i]
		if t.Date.After(on) {
			continue
		}
		origin := in.OriginOf(t.Key.Source)
		if !origin.Removed {
			continue
		}
		for j := range t.Postings {
			if p := &t.Postings[j]; p.Book.Kind == ledger.BookHolding {
				out[p.Book.Instrument] = addName(out[p.Book.Instrument], origin.Instance)
			}
		}
	}
	return out
}

// RemovedBehindLots returns, for each open lot of instrument at the end of
// on, keyed by the transaction that opened it, the removed connector
// instances behind it (D-91): those of the transaction that opened it and
// of every disposal that consumed part of it, replayed as ledger.LotsOn
// does (acquisitions before disposals on a date, FIFO). A lot with none is
// left out. txs are one account's current transactions in date order; a
// disposal of more units than held is an error.
func RemovedBehindLots(txs []ledger.Transaction, instrument ledger.Instrument, in Installed, on kernel.Date) (map[ledger.TransactionKey][]string, error) {
	book := ledger.HoldingBook(instrument)
	shaped := map[ledger.TransactionKey][]ledger.TransactionKey{}
	var lots []ledger.Lot
	for start := 0; start < len(txs) && !txs[start].Date.After(on); {
		end := start
		for end < len(txs) && txs[end].Date.Equal(txs[start].Date) {
			end++
		}
		lots = append(lots, openedLots(txs[start:end], book, shaped)...)
		var err error
		if lots, err = consumeLots(lots, txs[start:end], book, shaped); err != nil {
			return nil, err
		}
		start = end
	}
	out := map[ledger.TransactionKey][]string{}
	for i := range lots {
		for _, key := range shaped[lots[i].Origin] {
			if origin := in.OriginOf(key.Source); origin.Removed {
				out[lots[i].Origin] = addName(out[lots[i].Origin], origin.Instance)
			}
		}
	}
	return out, nil
}

// openedLots returns the lots txs open in book, each shaped by its own
// transaction.
func openedLots(txs []ledger.Transaction, book ledger.Book, shaped map[ledger.TransactionKey][]ledger.TransactionKey) []ledger.Lot {
	var lots []ledger.Lot
	for i := range txs {
		t := &txs[i]
		for j := range t.Postings {
			if p := &t.Postings[j]; p.Book == book && p.Quantity.Sign() > 0 {
				lot := ledger.Lot{Instrument: book.Instrument, Opened: t.Date, Quantity: p.Quantity, Cost: p.Amount, Origin: t.Key}
				if t.Kind == ledger.KindOpeningLot {
					lot.Opened, lot.Opening = t.LotOpened, true
				}
				lots = append(lots, lot)
				shaped[t.Key] = append(shaped[t.Key], t.Key)
			}
		}
	}
	return lots
}

// consumeLots consumes lots FIFO for every posting in txs that removes
// units from book, adding the disposal to the transactions behind each lot
// it consumed part of and left open.
func consumeLots(lots []ledger.Lot, txs []ledger.Transaction, book ledger.Book, shaped map[ledger.TransactionKey][]ledger.TransactionKey) ([]ledger.Lot, error) {
	for i := range txs {
		t := &txs[i]
		for j := range t.Postings {
			p := &t.Postings[j]
			if p.Book != book || p.Quantity.Sign() >= 0 {
				continue
			}
			taken, err := ledger.ConsumeFIFO(lots, p.Quantity.Neg())
			if err != nil {
				return nil, fmt.Errorf("replaying lots of %s: %w", book.Instrument, err)
			}
			markConsumed(&taken, t.Key, shaped)
			lots = taken.Remaining
		}
	}
	return lots, nil
}

// markConsumed adds disposal to the transactions behind each lot it
// consumed part of and left open.
func markConsumed(taken *ledger.LotConsumption, disposal ledger.TransactionKey, shaped map[ledger.TransactionKey][]ledger.TransactionKey) {
	for k := range taken.Remaining {
		origin := taken.Remaining[k].Origin
		if slices.ContainsFunc(taken.Consumed, func(c ledger.Lot) bool { return c.Origin == origin }) {
			shaped[origin] = append(shaped[origin], disposal)
		}
	}
}

// addName adds name to the sorted names, once.
func addName(names []string, name string) []string {
	i, found := slices.BinarySearch(names, name)
	if found {
		return names
	}
	return slices.Insert(names, i, name)
}
