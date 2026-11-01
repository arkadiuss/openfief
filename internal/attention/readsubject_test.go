package attention_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/attention"
)

// TestReadSubject checks that read access sees an unmapped broker
// account's item by its connector instance alone, never the broker
// account number (D-121, D-137), and every other subject as stored.
func TestReadSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kind    attention.Kind
		subject string
		want    string
	}{
		{name: "unmapped broker account", kind: attention.KindBrokerAccountUnmapped, subject: "flex-main/U0000001", want: "flex-main"},
		{name: "unmapped broker account without an instance part", kind: attention.KindBrokerAccountUnmapped, subject: "flex-main", want: "flex-main"},
		{name: "ticker", kind: attention.KindTickerUnconfirmed, subject: "xtb-main/TESTA.PL", want: "xtb-main/TESTA.PL"},
		{name: "batch", kind: attention.KindImportProblems, subject: "b-7", want: "b-7"},
		{name: "instrument", kind: attention.KindPriceCheckFlagged, subject: "XX00000SYNT3/SYN", want: "XX00000SYNT3/SYN"},
		{name: "assertion", kind: attention.KindAssertionFlagged, subject: "acc-1/2026-03-02/cash:PLN", want: "acc-1/2026-03-02/cash:PLN"},
		{name: "setting", kind: attention.KindSettingMissing, subject: "marginal-rate", want: "marginal-rate"},
		{name: "transfer pair", kind: attention.KindTransferBroken, subject: "acc-1/b-1+acc-2/b-2", want: "acc-1/b-1+acc-2/b-2"},
		// Kinds other contexts raise: valuation's market-data gap, whose
		// subject is the OpenFief account and the series, and its
		// unclassified instrument, an ISIN; neither holds a broker account.
		{name: "market-data gap", kind: attention.Kind("market-data-gap"), subject: "acc-1/price/XX00000SYNT3", want: "acc-1/price/XX00000SYNT3"},
		{name: "unclassified instrument", kind: attention.Kind("instrument-unclassified"), subject: "XX00000SYNT3", want: "XX00000SYNT3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := attention.ReadSubject(tt.kind, tt.subject); got != tt.want {
				t.Errorf("ReadSubject(%s, %q) = %q, want %q", tt.kind, tt.subject, got, tt.want)
			}
		})
	}
}
