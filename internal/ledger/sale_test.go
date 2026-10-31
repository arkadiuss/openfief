package ledger_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestSales checks the sales a realized gain is counted from (D-128): 10
// units bought at 100 and 10 at 120, then 15 sold for 1 950 with a 5 fee
// takes the first lot whole and 5 of the second (cost 600 of 1 200); a
// same-day buy and sell consumes the day's buy; a withdrawal is no sale.
func TestSales(t *testing.T) {
	t.Parallel()
	a := instrument(t, isinA, "TSTA")
	m := must(t)
	usd := kernel.NewMoney(kernel.Decimal{}, cur(t, "USD"))
	txs := []ledger.Transaction{
		m(ledger.NewSell(header(t, "s1", "2026-03-10"), a, qty(t, "15"), money(t, "1950", "USD"), money(t, "5", "USD"))),
		m(ledger.NewBuy(header(t, "b1", "2026-03-01"), a, qty(t, "10"), money(t, "1000", "USD"), usd)),
		m(ledger.NewBuy(header(t, "b2", "2026-03-02"), a, qty(t, "10"), money(t, "1200", "USD"), usd)),
		m(ledger.NewSell(header(t, "s2", "2026-03-11"), a, qty(t, "5"), money(t, "650", "USD"), usd)),
		m(ledger.NewWithdrawal(header(t, "w1", "2026-03-11"), money(t, "10", "USD"))),
	}
	got, err := ledger.Sales(txs)
	if err != nil {
		t.Fatalf("Sales: %v", err)
	}
	lot := func(id, opened, q, cost string) ledger.Lot {
		return ledger.Lot{Instrument: a, Opened: date(t, opened), Quantity: qty(t, q), Cost: money(t, cost, "USD"), Origin: ledger.TransactionKey{Source: "test", ExternalID: id}}
	}
	want := []ledger.Sale{
		{
			Key: ledger.TransactionKey{Source: "test", ExternalID: "s1"}, Date: date(t, "2026-03-10"), Instrument: a,
			Proceeds: money(t, "1950", "USD"), Fees: []kernel.Money{money(t, "5", "USD")},
			Consumed: []ledger.Lot{lot("b1", "2026-03-01", "10", "1000"), lot("b2", "2026-03-02", "5", "600.00")},
		},
		{
			Key: ledger.TransactionKey{Source: "test", ExternalID: "s2"}, Date: date(t, "2026-03-11"), Instrument: a,
			Proceeds: money(t, "650", "USD"), Consumed: []ledger.Lot{lot("b2", "2026-03-02", "5", "600.00")},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestSalesOversell(t *testing.T) {
	t.Parallel()
	a := instrument(t, isinA, "TSTA")
	txs := []ledger.Transaction{
		must(t)(ledger.NewSell(header(t, "s1", "2026-03-10"), a, qty(t, "1"), money(t, "10", "USD"), money(t, "0", "USD"))),
	}
	var oversell *ledger.OversellError
	if _, err := ledger.Sales(txs); !errors.As(err, &oversell) || oversell.Disposal.ExternalID != "s1" {
		t.Errorf("Sales = %v, want an *OversellError for s1", err)
	}
}
