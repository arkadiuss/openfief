package ingest_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestConfirmedName(t *testing.T) {
	t.Parallel()
	proposed, err := ledger.ParseISIN(isinA)
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	other, err := ledger.ParseISIN("IETEST000026")
	if err != nil {
		t.Fatalf("ParseISIN: %v", err)
	}
	tests := []struct {
		name   string
		ticker string
		lookup string
		isin   ledger.ISIN
		want   string
	}{
		{name: "the institution's name first", ticker: "Synthetic SA", lookup: "SYNTHETIC SA", isin: proposed, want: "Synthetic SA"},
		{name: "the lookup's name for the proposed ISIN", lookup: "SYNTHETIC SA", isin: proposed, want: "SYNTHETIC SA"},
		{name: "no lookup name for another ISIN", lookup: "SYNTHETIC SA", isin: other},
		{name: "no name at all", isin: proposed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := ingest.PendingTicker{
				Instance: "xtb-main", Ticker: ingest.Ticker{Symbol: "SYNT.PL", Name: tt.ticker},
				Proposal: ingest.ISINProposal{Symbol: "SYNT.PL", ISIN: proposed, Name: tt.lookup},
			}
			got, ok := ingest.ConfirmedName(&p, tt.isin)
			if got != tt.want || ok != (tt.want != "") {
				t.Errorf("ConfirmedName = %q, %v; want %q", got, ok, tt.want)
			}
		})
	}
}
