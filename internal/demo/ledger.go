package demo

import (
	"fmt"
	"strconv"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// Batch is what one call of the ledger's API imports: transactions with the
// balance assertions to check after them.
type Batch struct {
	Transactions []ledger.Transaction
	Assertions   []ledger.BalanceAssertion
}

// history is an account's seeding at go-live and its history since, for
// the ledger's API.
type history struct {
	opening  Batch
	activity Batch
}

// direct writes an account's simulated history as ledger transactions, for
// the accounts the demo books through the ledger's own API.
type direct struct {
	account ledger.AccountID
	source  string
}

// header returns the transaction header for externalID on d.
func (w direct) header(externalID string, d kernel.Date) ledger.TransactionHeader {
	return ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: w.source, ExternalID: externalID},
		Account: w.account,
		Date:    d,
	}
}

// opening returns the seeding of the account at go-live (D-23): its opening
// lots and opening cash against equity (D-33), with balance assertions of
// both on go-live.
func (w direct) opening(goLive kernel.Date, lots []lot, cash map[kernel.Currency]int64) (Batch, error) {
	var b Batch
	units := map[*Instrument]int64{}
	for i := range lots {
		l := &lots[i]
		id := "opening-lot:" + l.in.Instrument.ISIN.String() + ":" + l.opened.String() + ":" + l.id
		t, err := ledger.NewOpeningLot(w.header(id, goLive), l.in.Instrument, quantity(l.units),
			kernel.NewMoney(cents(l.costCents), l.in.Currency), l.opened)
		if err != nil {
			return Batch{}, fmt.Errorf("demo opening lot of %s: %w", w.account, err)
		}
		t.Description = "opening lot"
		b.Transactions = append(b.Transactions, t)
		units[l.in] += l.units
	}
	for _, c := range sortedCurrencies(cash) {
		m := kernel.NewMoney(cents(cash[c]), c)
		t, err := ledger.NewOpeningCash(w.header(ledger.OpeningCashID(c, goLive), goLive), m)
		if err != nil {
			return Batch{}, fmt.Errorf("demo opening cash of %s: %w", w.account, err)
		}
		t.Description = "opening cash"
		b.Transactions = append(b.Transactions, t)
	}
	b.Assertions = w.assertions(snapshot{date: goLive, cash: cash, units: units})
	return b, nil
}

// activity returns ops as ledger transactions, with balance assertions of
// the account at end.
func (w direct) activity(ops []op, end snapshot) (Batch, error) {
	var b Batch
	for i := range ops {
		t, err := w.transaction(&ops[i])
		if err != nil {
			return Batch{}, fmt.Errorf("demo transaction %s of %s: %w", ops[i].id, w.account, err)
		}
		b.Transactions = append(b.Transactions, t)
	}
	b.Assertions = w.assertions(end)
	return b, nil
}

// transaction returns the ledger transaction of o.
func (w direct) transaction(o *op) (ledger.Transaction, error) {
	h := w.header(o.id, o.date)
	m := kernel.NewMoney(cents(o.cents), o.currency)
	fee := kernel.NewMoney(cents(o.fee), o.currency)
	var t ledger.Transaction
	var err error
	switch o.kind {
	case opDeposit:
		t, err = ledger.NewDeposit(h, m)
	case opFee:
		t, err = ledger.NewFee(h, m)
	case opBuy:
		t, err = ledger.NewBuy(h, o.in.Instrument, quantity(o.units), m, fee)
	case opSell:
		t, err = ledger.NewSell(h, o.in.Instrument, quantity(o.units), m, fee)
	case opDividend:
		t, err = ledger.NewDividend(h, o.in.Instrument, m, kernel.NewMoney(cents(o.withheld), o.currency))
	case opFX:
		bought := kernel.NewMoney(kernel.NewDecimal(o.baseUnits, 0), o.base)
		t, err = ledger.NewFXConversion(h, m, bought, fee)
	}
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("demo %s: %w", o.kind, err)
	}
	t.Description = o.kind.String()
	if o.in != nil {
		t.Description += " " + o.in.Instrument.Symbol
	}
	return t, nil
}

// assertions returns the balance assertions of snap: its cash per currency
// and its holdings.
func (w direct) assertions(snap snapshot) []ledger.BalanceAssertion {
	out := make([]ledger.BalanceAssertion, 0, len(snap.cash)+len(snap.units))
	for _, c := range sortedCurrencies(snap.cash) {
		out = append(out, ledger.CashAssertion(w.account, snap.date, kernel.NewMoney(cents(snap.cash[c]), c)))
	}
	for _, in := range sortedInstruments(snap.units) {
		out = append(out, ledger.HoldingAssertion(w.account, snap.date, in.Instrument, quantity(snap.units[in])))
	}
	return out
}

// quantity returns n whole units.
func quantity(n int64) kernel.Quantity { return kernel.NewQuantity(kernel.NewDecimal(n, 0)) }

// String returns the operation kind's name, such as "deposit".
func (k opKind) String() string {
	switch k {
	case opDeposit:
		return "deposit"
	case opBuy:
		return "buy"
	case opSell:
		return "sell"
	case opDividend:
		return "dividend"
	case opFee:
		return "fee"
	case opFX:
		return "fx"
	}
	return "opKind(" + strconv.Itoa(int(k)) + ")"
}
