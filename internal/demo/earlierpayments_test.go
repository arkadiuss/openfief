package demo_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestBondsEarlierPayments checks what the demo's owner enters as paid into
// the bond IKE before go-live (D-149): the cost of its opening lots plus its
// opening cash, as the opening batch books them.
func TestBondsEarlierPayments(t *testing.T) {
	t.Parallel()
	today, err := kernel.ParseDate("2026-03-02")
	if err != nil {
		t.Fatal(err)
	}
	ds, err := demo.Generate(1, today)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	var seeded kernel.Decimal
	for i := range ds.BondsOpening.Transactions {
		for _, p := range ds.BondsOpening.Transactions[i].Postings {
			if p.Book.Kind == ledger.BookEquity {
				seeded = seeded.Sub(p.Amount.Amount())
			}
		}
	}
	paid, err := demo.BondsEarlierPayments()
	if err != nil || !paid.Amount().Equal(seeded) || paid.Currency().String() != "PLN" || !paid.Equal(ds.EarlierPayments) {
		t.Errorf("BondsEarlierPayments = %s, %v; want the %s PLN the opening batch seeds, as the dataset holds (%s)", paid, err, seeded, ds.EarlierPayments)
	}
}
