package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestBrokenTransferLinks(t *testing.T) {
	t.Parallel()
	confirmed := ledger.TransferLink{Pair: pair("w", "d"), Decision: ledger.TransferConfirmed, Version: 1}
	rejected := ledger.TransferLink{Pair: pair("w", "d"), Decision: ledger.TransferRejected, Version: 2}
	tests := []struct {
		name    string
		deposit string // the deposit's current amount
		day     string // the deposit's current date
		drop    bool   // the deposit has no current version
		links   []ledger.TransferLink
		want    []ledger.TransferLink
	}{
		{name: "matching legs keep counting", deposit: "100", day: "2026-03-03", links: []ledger.TransferLink{confirmed}},
		{
			name: "a corrected amount breaks the link", deposit: "99.50", day: "2026-03-03",
			links: []ledger.TransferLink{confirmed}, want: []ledger.TransferLink{confirmed},
		},
		{
			name: "a corrected date out of the window breaks the link", deposit: "100", day: "2026-03-09",
			links: []ledger.TransferLink{confirmed}, want: []ledger.TransferLink{confirmed},
		},
		{
			name: "a deleted leg breaks the link", deposit: "100", day: "2026-03-03", drop: true,
			links: []ledger.TransferLink{confirmed}, want: []ledger.TransferLink{confirmed},
		},
		{name: "a rejected pair is never broken", deposit: "99.50", day: "2026-03-03", links: []ledger.TransferLink{rejected}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			txs := []ledger.Transaction{leg(t, "a", "w", "2026-03-02", "-100", "PLN")}
			if !tt.drop {
				txs = append(txs, leg(t, "b", "d", tt.day, tt.deposit, "PLN"))
			}
			got := ledger.BrokenTransferLinks(tt.links, txs)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BrokenTransferLinks (-want +got):\n%s", diff)
			}
		})
	}
}
