package ibkrflex

import (
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// IBKR's cash transaction types that this adapter books, and the level of
// detail of one cash transaction.
const (
	typeDividends   = "Dividends"
	typeWithholding = "Withholding Tax"
	typeDeposits    = "Deposits/Withdrawals"
	typeDepositsOld = "Deposits & Withdrawals"
	typeOtherFees   = "Other Fees"
	levelDetail     = "DETAIL"
	sectionCash     = "CashTransactions"
	cashIDAttr      = "transactionID"
)

// cashTransactions maps the CashTransactions section, keyed by IBKR's
// transactionID: deposits and withdrawals, other fees, and dividends. IBKR
// reports a dividend and the tax withheld from it as separate rows, so each
// Withholding Tax row is paired with its dividend (see pairWithholding) and
// booked with it as one dividend transaction under the dividend's
// transactionID. In opening mode every row is skipped (D-23).
func (s *stmt) cashTransactions(rows []row) {
	var dividends, withholdings []*row
	for i := range rows {
		r := &rows[i]
		ref := s.ref(sectionCash, r, cashIDAttr)
		if err := s.needActivityRow(r, "CashTransaction", levelDetail); err != nil {
			s.problem(ref, err)
			continue
		}
		switch r.text("type") {
		case typeDividends:
			dividends = append(dividends, r)
		case typeWithholding:
			withholdings = append(withholdings, r)
		}
	}
	withheld := s.pairWithholding(dividends, withholdings)
	for i := range rows {
		r := &rows[i]
		if s.needActivityRow(r, "CashTransaction", levelDetail) != nil || r.text("type") == typeWithholding {
			continue
		}
		t, err := s.cashTransaction(r, withheld[r])
		s.add(s.ref(sectionCash, r, cashIDAttr), &t, err)
	}
}

// cashTransaction translates one cash transaction row other than
// withholding tax; withheld is the tax paired with a dividend.
func (s *stmt) cashTransaction(r *row, withheld kernel.Money) (ledger.Transaction, error) {
	kind := r.text("type")
	switch kind {
	case typeDividends, typeDeposits, typeDepositsOld, typeOtherFees:
	default:
		return ledger.Transaction{}, unsupported("cash transaction type %q is not supported", kind)
	}
	id := r.text(cashIDAttr)
	if id == "" {
		return ledger.Transaction{}, unsupported("cash transaction without a transactionID")
	}
	on, err := r.date("dateTime", "reportDate")
	if err != nil {
		return ledger.Transaction{}, err
	}
	amount, err := r.money("amount")
	if err != nil {
		return ledger.Transaction{}, err
	}
	h := s.header(id, on)
	desc := r.text("description")
	var t ledger.Transaction
	switch {
	case kind == typeDividends:
		return dividend(h, r, amount, withheld)
	case kind == typeOtherFees:
		return fee(h, amount, desc)
	case amount.Sign() > 0:
		t, err = ledger.NewDeposit(h, amount)
	default:
		t, err = ledger.NewWithdrawal(h, amount.Neg())
	}
	return booked(&t, err, desc)
}

// fee returns the fee of an Other Fees row, which IBKR reports as a
// negative amount.
func fee(h ledger.TransactionHeader, amount kernel.Money, desc string) (ledger.Transaction, error) {
	if amount.Sign() >= 0 {
		return ledger.Transaction{}, unsupported("a fee refund is not supported")
	}
	t, err := ledger.NewFee(h, amount.Neg())
	return booked(&t, err, desc)
}

// dividend returns the dividend of a Dividends row with the tax withheld from
// it.
func dividend(h ledger.TransactionHeader, r *row, gross, withheld kernel.Money) (ledger.Transaction, error) {
	if gross.Sign() <= 0 {
		return ledger.Transaction{}, unsupported("a dividend reversal is not supported")
	}
	in, err := r.instrument()
	if err != nil {
		return ledger.Transaction{}, err
	}
	t, err := ledger.NewDividend(h, in, gross, withheld)
	return booked(&t, err, r.text("description"))
}

// pairWithholding returns the tax withheld from each dividend, from the
// Withholding Tax rows. A withholding row belongs to the dividend with the
// same actionID when both have one, and otherwise to the dividend with the
// same ISIN, currency and date. Several withholding rows of one dividend
// add up. A withholding row with no dividend, more than one, or a positive
// amount (a refund) becomes an unsupported ingest problem.
func (s *stmt) pairWithholding(dividends, withholdings []*row) map[*row]kernel.Money {
	withheld := map[*row]kernel.Money{}
	for _, w := range withholdings {
		ref := s.ref(sectionCash, w, cashIDAttr)
		d, err := matchDividend(w, dividends)
		if err != nil {
			s.problem(ref, err)
			continue
		}
		tax, err := w.money("amount")
		if err != nil {
			s.problem(ref, err)
			continue
		}
		if tax.Sign() >= 0 {
			s.problem(ref, unsupported("a withholding tax refund is not supported"))
			continue
		}
		sum, ok := withheld[d]
		if !ok {
			withheld[d] = tax.Neg()
			continue
		}
		if withheld[d], err = sum.Add(tax.Neg()); err != nil {
			s.problem(ref, unsupported("withholding tax in another currency than earlier withholding of its dividend"))
		}
	}
	return withheld
}

// matchDividend returns the one dividend that withholding row w belongs to.
func matchDividend(w *row, dividends []*row) (*row, error) {
	var found []*row
	for _, d := range dividends {
		if sameAction(w, d) {
			found = append(found, d)
		}
	}
	switch len(found) {
	case 0:
		return nil, unsupported("withholding tax without its dividend in this statement")
	case 1:
		return found[0], nil
	}
	return nil, unsupported("withholding tax matches %d dividends", len(found))
}

// sameAction reports whether withholding row w and dividend row d record one
// corporate action: the same actionID when both have one, and otherwise the
// same ISIN, currency and date.
func sameAction(w, d *row) bool {
	if w.text("actionID") != "" && d.text("actionID") != "" {
		return w.text("actionID") == d.text("actionID")
	}
	wd, werr := w.date("dateTime", "reportDate")
	dd, derr := d.date("dateTime", "reportDate")
	return werr == nil && derr == nil && wd.Equal(dd) &&
		w.text("isin") != "" && w.text("isin") == d.text("isin") && w.text("currency") == d.text("currency")
}
