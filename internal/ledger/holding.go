package ledger

import (
	"sort"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Holding is the quantity of one instrument in one account on a date,
// derived from postings.
type Holding struct {
	Account    AccountID
	Instrument Instrument
	Date       kernel.Date
	Quantity   kernel.Quantity
}

// HoldingOn returns account's holding of the instrument in at the end of
// on: the sum of the holding postings of the current versions of account's
// transactions dated on or before on.
func HoldingOn(log []VersionedTransaction, account AccountID, in Instrument, on kernel.Date) (Holding, error) {
	txs, err := currentInAccount(log, account, on)
	if err != nil {
		return Holding{}, err
	}
	h := Holding{Account: account, Instrument: in, Date: on}
	for i := range txs {
		for j := range txs[i].Postings {
			if p := &txs[i].Postings[j]; p.Book == HoldingBook(in) {
				h.Quantity = h.Quantity.Add(p.Quantity)
			}
		}
	}
	return h, nil
}

// HoldingsOn returns every non-zero holding in account at the end of on,
// ordered by ISIN and then venue symbol.
func HoldingsOn(log []VersionedTransaction, account AccountID, on kernel.Date) ([]Holding, error) {
	txs, err := currentInAccount(log, account, on)
	if err != nil {
		return nil, err
	}
	units := map[Instrument]kernel.Quantity{}
	for i := range txs {
		for j := range txs[i].Postings {
			if p := &txs[i].Postings[j]; p.Book.Kind == BookHolding {
				units[p.Book.Instrument] = units[p.Book.Instrument].Add(p.Quantity)
			}
		}
	}
	out := make([]Holding, 0, len(units))
	for in, q := range units {
		if !q.IsZero() {
			out = append(out, Holding{Account: account, Instrument: in, Date: on, Quantity: q})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Instrument.String() < out[j].Instrument.String() })
	return out, nil
}

// currentInAccount returns the current versions of account's transactions
// dated on or before on, in date then key order.
func currentInAccount(log []VersionedTransaction, account AccountID, on kernel.Date) ([]Transaction, error) {
	all, err := CurrentTransactions(log)
	if err != nil {
		return nil, err
	}
	out := all[:0]
	for i := range all {
		if all[i].Account == account && !all[i].Date.After(on) {
			out = append(out, all[i])
		}
	}
	return out, nil
}
