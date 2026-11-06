package xtb_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
)

// TestMapHoldings checks D-155: in both modes, Open Positions asserts each
// stock or ETF ticker's summed volume on the report's date, never a CFD's;
// a ticker one of whose positions can't be read is not asserted but
// reported; an export without the report's date asserts nothing.
func TestMapHoldings(t *testing.T) {
	t.Parallel()
	short := xtbtest.OpenRow(&xtbtest.Position{Name: "Test Beta PLC", Ticker: "TESTB.UK", Category: "ETF", Volume: "2", OpenPrice: "20", Opened: at(2026, 3, 1, 9, 0)})
	short[4] = xtbtest.Text("SELL")
	tests := []struct {
		name string
		edit func(*xtbtest.Export)
		mode ingest.Mode
		want []string
	}{
		{
			name: "two positions of a stock are summed, an ETF and a CFD", mode: ingest.ModeTransactions, edit: func(e *xtbtest.Export) { e.Open = current().Open },
			want: []string{"TESTA.PL 2026-06-30 10 Open Positions#10", "TESTB.UK 2026-06-30 5 Open Positions#12"},
		},
		{
			name: "the same in opening mode", mode: ingest.ModeOpening, edit: func(e *xtbtest.Export) { e.Open = current().Open },
			want: []string{"TESTA.PL 2026-06-30 10 Open Positions#10", "TESTB.UK 2026-06-30 5 Open Positions#12"},
		},
		{
			name: "a ticker with a short position is not asserted", mode: ingest.ModeTransactions,
			edit: func(e *xtbtest.Export) { e.Open = slices.Concat(current().Open, [][]xtbtest.Cell{short}) },
			want: []string{
				"TESTA.PL 2026-06-30 10 Open Positions#10",
				"problem unsupported Open Positions#12: the holding of TESTB.UK is not asserted: one of its open positions can't be read",
			},
		},
		{
			name: "no report date", mode: ingest.ModeTransactions,
			edit: func(e *xtbtest.Export) { e.Open, e.NoReportDate = current().Open, true },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := fixture()
			tt.edit(e)
			var m ingest.Mapping
			if tt.mode == ingest.ModeOpening {
				m = mapOpening(t, e, e.FileName(), "")
			} else {
				m = mapExport(t, e)
			}
			if diff := cmp.Diff(tt.want, holdingLines(&m)); diff != "" {
				t.Errorf("holdings (-want +got):\n%s", diff)
			}
		})
	}
}

// holdingLines writes m's ticker assertions and the problems of holdings
// not asserted.
func holdingLines(m *ingest.Mapping) []string {
	var out []string
	for i := range m.ByBrokerAccount {
		for _, a := range m.ByBrokerAccount[i].TickerAssertions {
			out = append(out, fmt.Sprintf("%s %s %s %s", a.Symbol, a.Date, a.Quantity, a.Ref))
		}
	}
	for _, p := range m.Problems {
		if strings.HasPrefix(p.Reason, "the holding of ") {
			out = append(out, "problem "+p.String())
		}
	}
	return out
}
