package ibkrflex

import (
	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// baseSummary is the CashReport row that totals all currencies in the base
// currency; it is not cash in any one currency.
const baseSummary = "BASE_SUMMARY"

// cashReport maps the CashReport section: each currency's endingCash becomes
// a cash balance assertion on the statement's last day (D-18). In opening
// mode it also becomes the opening cash: a deposit, or a withdrawal for a
// negative balance, dated on the go-live date, so that the cash assertion
// holds from go-live.
func (s *stmt) cashReport(rows []row) {
	for i := range rows {
		r := &rows[i]
		if r.text("currency") == baseSummary {
			continue
		}
		ref := s.ref("CashReport", r, "currency")
		cash, err := endingCash(r)
		if err != nil {
			s.problem(ref, err)
			continue
		}
		s.out.Assertions = append(s.out.Assertions, ledger.CashAssertion(s.account, s.to, cash))
		if s.mode == ingest.ModeOpening && !cash.IsZero() {
			t, err := s.openingCash(cash)
			s.add(ref, &t, err)
		}
	}
}

// endingCash reads a CashReportCurrency row's ending cash.
func endingCash(r *row) (kernel.Money, error) {
	if r.elem != "CashReportCurrency" {
		return kernel.Money{}, skipped("<%s> is not a cash report row", r.elem)
	}
	return r.money("endingCash")
}

// openingCash returns the transaction that brings cash into the account at
// go-live: a deposit, or a withdrawal for a negative balance, against equity
// (D-33), keyed so that the ledger knows the account is seeded.
func (s *stmt) openingCash(cash kernel.Money) (ledger.Transaction, error) {
	h := s.header(ledger.OpeningCashID(cash.Currency(), s.to), s.to)
	t, err := ledger.NewOpeningCash(h, cash)
	return booked(&t, err, "opening cash")
}
