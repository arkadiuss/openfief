package xtb_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
)

// TestMapCashIncome checks that free-funds interest is booked as cash
// income with the tax of its date withheld, and a closed CFD position's
// result as cash income of either sign (D-152, D-24), and which rows stay
// ingest problems.
func TestMapCashIncome(t *testing.T) {
	t.Parallel()
	feb := at(2026, 2, 1, 6, 0)
	interest := func(amount, id string) []xtbtest.Cell {
		return xtbtest.CashRow("Free funds interest", "", "", "", feb, amount, id, "Free-funds Interest 2026-01")
	}
	tax := func(amount, id string) []xtbtest.Cell {
		return xtbtest.CashRow("Free funds interest tax", "", "", "", feb, amount, id, "Free-funds Interest Tax 2026-01")
	}
	cfd := func(profit, id string) []xtbtest.Cell {
		return xtbtest.ClosedRow(&xtbtest.Position{
			Name: "Test Index", Ticker: "TESTCFD", Category: "IND", Volume: "1", OpenPrice: "100", Opened: at(2026, 1, 5, 10, 0),
			ClosePrice: "90", Closed: at(2026, 1, 6, 10, 0), Profit: profit, PositionID: id,
		})
	}
	tests := []struct {
		name   string
		cash   [][]xtbtest.Cell
		closed [][]xtbtest.Cell
		want   []string
	}{
		{
			name: "interest with two tax rows",
			cash: [][]xtbtest.Cell{tax("-0.10", "9000000003"), tax("-0.13", "9000000002"), interest("1.23", "9000000001")},
			want: []string{
				"tx xtb/9000000001 cash-income 2026-02-01 \"free-funds interest\" income:PLN -1.23 PLN|withholding-tax:PLN 0.23 PLN|cash:PLN 1 PLN",
			},
		},
		{
			name: "interest without tax",
			cash: [][]xtbtest.Cell{interest("0.50", "9000000001")},
			want: []string{"tx xtb/9000000001 cash-income 2026-02-01 \"free-funds interest\" income:PLN -0.5 PLN|cash:PLN 0.5 PLN"},
		},
		{
			name: "tax without its interest",
			cash: [][]xtbtest.Cell{tax("-0.23", "9000000002")},
			want: []string{"problem unsupported Cash Operations#6[ID=9000000002]: free-funds interest tax without its free-funds interest in this file"},
		},
		{
			name: "tax refund",
			cash: [][]xtbtest.Cell{tax("0.23", "9000000002"), interest("1.23", "9000000001")},
			want: []string{
				"tx xtb/9000000001 cash-income 2026-02-01 \"free-funds interest\" income:PLN -1.23 PLN|cash:PLN 1.23 PLN",
				"problem unsupported Cash Operations#6[ID=9000000002]: a free-funds interest tax refund is not supported",
			},
		},
		{
			name: "tax matching two interest rows",
			cash: [][]xtbtest.Cell{tax("-0.23", "9000000003"), interest("1", "9000000002"), interest("1", "9000000001")},
			want: []string{
				"tx xtb/9000000001 cash-income 2026-02-01 \"free-funds interest\" income:PLN -1 PLN|cash:PLN 1 PLN",
				"tx xtb/9000000002 cash-income 2026-02-01 \"free-funds interest\" income:PLN -1 PLN|cash:PLN 1 PLN",
				"problem unsupported Cash Operations#6[ID=9000000003]: free-funds interest tax matches 2 free-funds interest rows",
			},
		},
		{
			name: "negative interest",
			cash: [][]xtbtest.Cell{interest("-1", "9000000001")},
			want: []string{"problem unsupported Cash Operations#6[ID=9000000001]: negative free-funds interest is not supported"},
		},
		{
			name:   "closed CFD's profit and loss",
			closed: [][]xtbtest.Cell{cfd("12.5", "9000000101"), cfd("-7.25", "9000000102")},
			want: []string{
				"tx xtb/closed-position:9000000101 cash-income 2026-01-06 \"closed CFD position TESTCFD\" income:PLN -12.5 PLN|cash:PLN 12.5 PLN",
				"tx xtb/closed-position:9000000102 cash-income 2026-01-06 \"closed CFD position TESTCFD\" income:PLN 7.25 PLN|cash:PLN -7.25 PLN",
			},
		},
		{
			name:   "closed CFD of no result",
			closed: [][]xtbtest.Cell{cfd("0", "9000000101")},
			want:   []string{"problem skipped Closed Positions#6[Position ID=9000000101]: a closed CFD position of no result"},
		},
		{
			name:   "closed CFD without a position ID",
			closed: [][]xtbtest.Cell{cfd("1", "")},
			want:   []string{"problem unsupported Closed Positions#6: a closed CFD position without a position ID is not supported"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := fixture()
			e.Cash, e.Closed, e.Open = tt.cash, tt.closed, nil
			m := mapExport(t, e)
			if diff := cmp.Diff(tt.want, renderLines(&m)); diff != "" {
				t.Errorf("mapping (-want +got):\n%s", diff)
			}
		})
	}
}

// renderLines writes m's transactions, each with its postings on one line,
// then its problems.
func renderLines(m *ingest.Mapping) []string {
	var out []string
	for i := range m.ByBrokerAccount {
		for j := range m.ByBrokerAccount[i].Transactions {
			tx := &m.ByBrokerAccount[i].Transactions[j]
			postings := make([]string, 0, len(tx.Postings))
			for _, p := range tx.Postings {
				postings = append(postings, p.Book.String()+" "+p.Amount.String())
			}
			out = append(out, "tx "+tx.Key.String()+" "+tx.Kind.String()+" "+tx.Date.String()+
				" \""+tx.Description+"\" "+strings.Join(postings, "|"))
		}
	}
	for _, p := range m.Problems {
		out = append(out, "problem "+p.String())
	}
	return out
}
