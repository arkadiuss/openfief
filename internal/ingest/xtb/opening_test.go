package xtb_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/xtb"
	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// money returns a synthetic amount in currency code.
func money(t *testing.T, amount, code string) kernel.Money {
	t.Helper()
	c, err := kernel.ParseCurrency(code)
	if err != nil {
		t.Fatal(err)
	}
	m, err := kernel.ParseMoney(amount, c)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// current is a synthetic current export of account 0000000 in PLN, taken
// on 2026-06-30, whose Open Positions hold two positions of one stock, one
// of them with a float artifact in its open price, an ETF quoted in
// another currency and a CFD. Its Cash Operations reach back before them.
func current() *xtbtest.Export {
	e := fixture()
	e.Open = [][]xtbtest.Cell{
		xtbtest.OpenRow(&xtbtest.Position{
			Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "7", OpenPrice: "123.45",
			Opened: at(2025, 11, 4, 10, 0), Value: "864.15",
		}),
		xtbtest.OpenRow(&xtbtest.Position{
			Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "3", OpenPrice: "130.10000000000001",
			Opened: at(2026, 2, 10, 9, 30), Value: "420", CurrentPrice: "140",
		}),
		xtbtest.OpenRow(&xtbtest.Position{
			Name: "Test Beta PLC", Ticker: "TESTB.UK", Category: "ETF", Volume: "5", OpenPrice: "20.00",
			Opened: at(2025, 11, 5, 23, 30), Value: "540.12", CurrentPrice: "21",
		}),
		xtbtest.OpenRow(&xtbtest.Position{Name: "Test Gold", Ticker: "TESTCMD", Category: "CMD", Volume: "0.1", OpenPrice: "2000", Opened: at(2026, 6, 1, 10, 0)}),
	}
	return e
}

// mapOpening maps e in opening mode with the owner's opening cash in PLN,
// none when cash is empty.
func mapOpening(t *testing.T, e *xtbtest.Export, fileName, cash string) ingest.Mapping {
	t.Helper()
	var c kernel.Money
	if cash != "" {
		c = money(t, cash, "PLN")
	}
	m, err := xtb.Mapper{}.Map(e.Bytes(), fileName, ingest.ModeOpening, c)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	return m
}

func TestMapOpeningGolden(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cash string
	}{
		{name: "export.opening.golden", cash: "2500.50"},
		{name: "nocash.opening.golden", cash: "0"},
		{name: "unentered.opening.golden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := current()
			m := mapOpening(t, e, e.FileName(), tt.cash)
			checkGolden(t, tt.name, render(&m))
		})
	}
}

// TestMapOpeningAgain checks that a re-upload of the same current export,
// even renamed by the owner, maps to the same keys and amounts, so the
// ledger finds the seeding unchanged (D-18, D-23).
func TestMapOpeningAgain(t *testing.T) {
	t.Parallel()
	e := current()
	first := mapOpening(t, e, e.FileName(), "2500.50")
	again := mapOpening(t, current(), "renamed.xlsx", "2500.50")
	if diff := cmp.Diff(render(&first), render(&again)); diff != "" {
		t.Errorf("the same export maps otherwise (-want +got):\n%s", diff)
	}
}

// TestMapOpeningSkipsHistory checks that an opening batch books nothing of
// Cash Operations and Closed Positions: the opening lots and the opening
// cash hold their effect (D-23), and the transfers' comments are never
// read.
func TestMapOpeningSkipsHistory(t *testing.T) {
	t.Parallel()
	e := current()
	m := mapOpening(t, e, e.FileName(), "2500.50")
	out := render(&m)
	for _, gone := range []string{"9000000001", "Cash Operations", "Closed Positions", "dividend", "withdrawal", "Bank Testowy"} {
		if strings.Contains(out, gone) {
			t.Errorf("the opening mapping holds %q:\n%s", gone, out)
		}
	}
}

// TestMapOpeningPositionKeys checks the key of an open position: its
// position ID when the sheet has one, else its ticker, open time and its
// count among positions of both.
func TestMapOpeningPositionKeys(t *testing.T) {
	t.Parallel()
	same := xtbtest.Position{
		Name: "Test Alpha SA", Ticker: "TESTA.PL", Category: "STOCK", Volume: "1", OpenPrice: "100",
		Opened: at(2026, 3, 2, 9, 0), Value: "100",
	}
	numbered := same
	numbered.Name = "9000000201"
	tests := []struct {
		name string
		rows []*xtbtest.Position
		want []string
	}{
		{
			name: "two positions of one ticker opened together",
			rows: []*xtbtest.Position{&same, &same},
			want: []string{"xtb/open-position:TESTA.PL@20260302T090000Z#1", "xtb/open-position:TESTA.PL@20260302T090000Z#2"},
		},
		{
			name: "a position number in Instrument/Position",
			rows: []*xtbtest.Position{&numbered},
			want: []string{"xtb/open-position:9000000201"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := fixture()
			e.Open = nil
			for _, p := range tt.rows {
				e.Open = append(e.Open, xtbtest.OpenRow(p))
			}
			m := mapOpening(t, e, e.FileName(), "")
			var got []string
			for i := range m.ByBrokerAccount {
				for _, te := range m.ByBrokerAccount[i].TickerEntries {
					got = append(got, te.Header.Key.String())
				}
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("keys (-want +got):\n%s", diff)
			}
		})
	}
}
