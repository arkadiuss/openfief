package ledger

import (
	"errors"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Sale is one sell transaction's disposal of an instrument: its proceeds,
// its fees and the lots it consumed, FIFO, which a realized gain is counted
// from (D-128).
type Sale struct {
	Key        TransactionKey
	Date       kernel.Date
	Instrument Instrument
	Proceeds   kernel.Money
	// Fees are the sale's fees, one per currency.
	Fees     []kernel.Money
	Consumed []Lot
}

// Sales returns the sales (KindSell) among txs, one account's current
// transactions, in date then key order, each with the lots it consumed as
// LotsOn consumes them: on one date acquisitions first, then disposals
// FIFO. A disposal of more units than held fails with an *OversellError.
func Sales(txs []Transaction) ([]Sale, error) {
	txs = slices.Clone(txs)
	slices.SortStableFunc(txs, func(a, b Transaction) int {
		if c := a.Date.Compare(b.Date); c != 0 {
			return c
		}
		return compareKeys(a.Key, b.Key)
	})
	var out []Sale
	for _, in := range soldInstruments(txs) {
		sales, err := salesOf(txs, in)
		if err != nil {
			return nil, err
		}
		out = append(out, sales...)
	}
	slices.SortStableFunc(out, func(a, b Sale) int {
		if c := a.Date.Compare(b.Date); c != 0 {
			return c
		}
		return compareKeys(a.Key, b.Key)
	})
	return out, nil
}

// soldInstruments returns the instruments a sell in txs disposes of, in
// first-sold order.
func soldInstruments(txs []Transaction) []Instrument {
	var out []Instrument
	for i := range txs {
		if txs[i].Kind != KindSell {
			continue
		}
		for _, p := range txs[i].Postings {
			if p.Book.Kind == BookHolding && p.Quantity.Sign() < 0 && !slices.Contains(out, p.Book.Instrument) {
				out = append(out, p.Book.Instrument)
			}
		}
	}
	return out
}

// salesOf walks txs, which are in date order, keeping in's lots, and
// returns the sales of in with the lots each consumed.
func salesOf(txs []Transaction, in Instrument) ([]Sale, error) {
	var lots []Lot
	var out []Sale
	for start := 0; start < len(txs); {
		end := start
		for end < len(txs) && txs[end].Date.Equal(txs[start].Date) {
			end++
		}
		lots = append(lots, acquisitions(txs[start:end], in)...)
		for i := start; i < end; i++ {
			var sales []Sale
			var err error
			if lots, sales, err = disposeOf(lots, &txs[i], in); err != nil {
				return nil, err
			}
			out = append(out, sales...)
		}
		start = end
	}
	return out, nil
}

// disposeOf consumes lots FIFO for each posting of t that removes units of
// in, and returns the lots left and, when t is a sell, its sales.
func disposeOf(lots []Lot, t *Transaction, in Instrument) ([]Lot, []Sale, error) {
	var out []Sale
	for _, p := range t.Postings {
		if p.Book != HoldingBook(in) || p.Quantity.Sign() >= 0 {
			continue
		}
		taken, err := ConsumeFIFO(lots, p.Quantity.Neg())
		if oversell, ok := errors.AsType[*OversellError](err); ok {
			oversell.Account, oversell.Instrument, oversell.Date, oversell.Disposal = t.Account, in, t.Date, t.Key
			return nil, nil, oversell
		}
		if err != nil {
			return nil, nil, err
		}
		lots = taken.Remaining
		if t.Kind == KindSell {
			out = append(out, Sale{Key: t.Key, Date: t.Date, Instrument: in, Proceeds: p.Amount.Neg(), Fees: fees(t), Consumed: taken.Consumed})
		}
	}
	return lots, out, nil
}

// fees returns t's fee postings' amounts.
func fees(t *Transaction) []kernel.Money {
	var out []kernel.Money
	for _, p := range t.Postings {
		if p.Book.Kind == BookFee {
			out = append(out, p.Amount)
		}
	}
	return out
}
