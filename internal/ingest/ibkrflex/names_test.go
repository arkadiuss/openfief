package ibkrflex_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestMapInstrumentNames checks that a trade's and an open position's
// description is the instrument's display name (D-197), and that a cash
// transaction's, a currency trade's and a row without one name nothing.
func TestMapInstrumentNames(t *testing.T) {
	t.Parallel()
	sections := `<Trades>
<Trade ` + stkAttrs + `description="SYNTHETIC WORLD ETF" tradeID="1" tradeDate="20260310" buySell="BUY" quantity="2" proceeds="-200.5" />
<Trade accountId="U0000000" currency="USD" assetCategory="CASH" symbol="EUR.USD" description="EUR.USD" levelOfDetail="EXECUTION" transactionType="ExchTrade" tradeID="2" tradeDate="20260310" buySell="BUY" quantity="10" proceeds="-11" />
</Trades>
<OpenPositions>
<OpenPosition accountId="U0000000" currency="EUR" assetCategory="STK" subCategory="COMMON" symbol="TSTB" isin="IETEST000026" description="SYNTHETIC BOND FUND" multiplier="1" reportDate="2026-03-13" markPrice="107" position="3" costBasisMoney="300" levelOfDetail="SUMMARY" side="Long" />
</OpenPositions>
<CashTransactions>
<CashTransaction ` + cashTx + `currency="USD" type="Dividends" assetCategory="STK" symbol="TSTA" isin="PLTEST000010" description="TSTA CASH DIVIDEND" amount="7.5" dateTime="2026-03-11" transactionID="9" />
</CashTransactions>`
	m, err := testMapper().Map(strings.NewReader(statementXML(sections)), ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	got := map[string]string{}
	for _, n := range m.Names {
		got[n.ISIN.String()] = n.Name
	}
	want := map[string]string{"PLTEST000010": "SYNTHETIC WORLD ETF", "IETEST000026": "SYNTHETIC BOND FUND"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("names (-want +got):\n%s", diff)
	}
	for _, n := range m.Names {
		if n.ByOwner || n.ISIN == (ledger.ISIN{}) {
			t.Errorf("name %+v: want a delivered name of an ISIN", n)
		}
	}
}
