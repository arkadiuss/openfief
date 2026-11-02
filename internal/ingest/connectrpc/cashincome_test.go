package connectrpc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestCashIncomeRoundTrip checks that cash income (D-152) a connector
// writes, such as XTB's free-funds interest with its tax and a closed
// CFD's loss, reaches the core unchanged: its kind and its income ledger
// account survive the contract.
func TestCashIncomeRoundTrip(t *testing.T) {
	t.Parallel()
	day, err := kernel.ParseDate("2026-02-01")
	if err != nil {
		t.Fatal(err)
	}
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	h := func(id string) ledger.TransactionHeader {
		return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "xtb", ExternalID: id}, Account: "acc-1", Date: day}
	}
	interest, err := ledger.NewCashIncome(h("9000000007"), kernel.NewMoney(kernel.NewDecimal(123, -2), pln), kernel.NewMoney(kernel.NewDecimal(23, -2), pln))
	if err != nil {
		t.Fatal(err)
	}
	loss, err := ledger.NewCashIncome(h("closed-position:9000000102"), kernel.NewMoney(kernel.NewDecimal(-10, 0), pln), kernel.Money{})
	if err != nil {
		t.Fatal(err)
	}
	want := ingest.Mapping{Transactions: []ledger.Transaction{interest, loss}}
	d, err := delivery.FromMapping("b-1", "xtb", &want)
	if err != nil {
		t.Fatalf("FromMapping: %v", err)
	}
	got, err := deliveryFrom(d, "b-1", "")
	if err != nil {
		t.Fatalf("deliveryFrom: %v", err)
	}
	if diff := cmp.Diff(want, got.Mapping, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("mapping (-want +got):\n%s", diff)
	}
}
