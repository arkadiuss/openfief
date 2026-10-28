package ibkrflex_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// statementXML wraps sections in a synthetic one-statement Flex response for
// account U0000000 over 2026-03-09 to 2026-03-13.
func statementXML(sections string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="t" type="AF"><FlexStatements count="1">
<FlexStatement accountId="U0000000" fromDate="20260309" toDate="20260313">` + sections + `</FlexStatement>
</FlexStatements></FlexQueryResponse>`
}

// mapRendered maps sections in mode and renders the result.
func mapRendered(t *testing.T, mode ingest.Mode, sections string) string {
	t.Helper()
	m, err := testMapper().Map(strings.NewReader(statementXML(sections)), mode)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	return render(&m)
}

const (
	stkAttrs = `accountId="U0000000" currency="USD" assetCategory="STK" symbol="TSTA" isin="PLTEST000010" multiplier="1" levelOfDetail="EXECUTION" transactionType="ExchTrade" `
	cashTx   = `accountId="U0000000" levelOfDetail="DETAIL" `
)

func TestMapTradeRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		row  string
		want string
	}{
		{
			name: "buy with commission and tax in the trade currency",
			row:  `<Trade ` + stkAttrs + `tradeID="1" tradeDate="20260310" buySell="BUY" quantity="2" proceeds="-200.5" ibCommission="-1" ibCommissionCurrency="USD" taxes="-0.5" />`,
			want: "tx ibkr-flex:acc-ibkr/1 buy 2026-03-10 account=acc-ibkr \"buy TSTA\"\n  holding:PLTEST000010/TSTA 200.5 USD units=2\n  cash:USD -200.5 USD\n  fee:USD 1.5 USD\n  cash:USD -1.5 USD\n",
		},
		{
			name: "sell without commission, ISO date-time",
			row:  `<Trade ` + stkAttrs + `tradeID="2" dateTime="2026-03-11;153000" buySell="SELL" quantity="-1.25" proceeds="130" ibCommission="0" />`,
			want: "tx ibkr-flex:acc-ibkr/2 sell 2026-03-11 account=acc-ibkr \"sell TSTA\"\n  holding:PLTEST000010/TSTA -130 USD units=-1.25\n  cash:USD 130 USD\n",
		},
		{
			name: "date with a T separator",
			row:  `<Trade ` + stkAttrs + `tradeID="3" tradeDate="2026-03-12T10:00:00" buySell="BUY" quantity="1" proceeds="-10" />`,
			want: "tx ibkr-flex:acc-ibkr/3 buy 2026-03-12 account=acc-ibkr \"buy TSTA\"\n  holding:PLTEST000010/TSTA 10 USD units=1\n  cash:USD -10 USD\n",
		},
		{
			name: "US date format is refused",
			row:  `<Trade ` + stkAttrs + `tradeID="4" tradeDate="03/12/2026" buySell="BUY" quantity="1" proceeds="-10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=4]: missing or unreadable attribute: tradeDate\n",
		},
		{
			name: "impossible date",
			row:  `<Trade ` + stkAttrs + `tradeID="5" tradeDate="20260230" buySell="BUY" quantity="1" proceeds="-10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=5]: missing or unreadable attribute: tradeDate\n",
		},
		{
			name: "buy with a negative quantity",
			row:  `<Trade ` + stkAttrs + `tradeID="6" tradeDate="20260310" buySell="BUY" quantity="-1" proceeds="10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=6]: trade side \"BUY\" with a quantity of that sign is not supported\n",
		},
		{
			name: "commission credit",
			row:  `<Trade ` + stkAttrs + `tradeID="7" tradeDate="20260310" buySell="BUY" quantity="1" proceeds="-10" ibCommission="0.3" ibCommissionCurrency="USD" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=7]: ibCommission is a credit, which is not supported\n",
		},
		{
			name: "commission and tax in different currencies",
			row:  `<Trade ` + stkAttrs + `tradeID="8" tradeDate="20260310" buySell="BUY" quantity="1" proceeds="-10" ibCommission="-1" ibCommissionCurrency="EUR" taxes="-0.1" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=8]: commission and taxes in different currencies are not supported\n",
		},
		{
			name: "no tradeID",
			row:  `<Trade ` + stkAttrs + `tradeDate="20260310" buySell="BUY" quantity="1" proceeds="-10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1: trade without a tradeID\n",
		},
		{
			name: "exponent notation is refused",
			row:  `<Trade ` + stkAttrs + `tradeID="9" tradeDate="20260310" buySell="BUY" quantity="1e999999999" proceeds="-10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=9]: missing or unreadable attribute: quantity\n",
		},
		{
			name: "contract multiplier",
			row:  `<Trade accountId="U0000000" currency="USD" assetCategory="STK" symbol="TSTA" isin="PLTEST000010" multiplier="10" levelOfDetail="EXECUTION" tradeID="10" tradeDate="20260310" buySell="BUY" quantity="1" proceeds="-10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=10]: contract multiplier \"10\" is not supported\n",
		},
		{
			name: "CFD",
			row:  `<Trade accountId="U0000000" currency="USD" assetCategory="CFD" symbol="TSTA" levelOfDetail="EXECUTION" tradeID="11" tradeDate="20260310" buySell="BUY" quantity="1" proceeds="-10" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=11]: asset category \"CFD\" is not supported\n",
		},
		{
			name: "FX pair not in the row's currency",
			row:  `<Trade accountId="U0000000" currency="EUR" assetCategory="CASH" symbol="EUR.USD" levelOfDetail="EXECUTION" tradeID="12" tradeDate="20260310" buySell="BUY" quantity="100" proceeds="-108" />`,
			want: "problem unsupported FlexStatement#1/Trades#1[tradeID=12]: currency pair \"EUR.USD\" is not BASE.QUOTE in the row's currency\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapRendered(t, ingest.ModeTransactions, "<Trades>"+tt.row+"</Trades>")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mapping (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapCashTransactionRows(t *testing.T) {
	t.Parallel()
	div := `<CashTransaction ` + cashTx + `type="Dividends" currency="USD" symbol="TSTA" isin="PLTEST000010" amount="10" dateTime="20260311;202000" transactionID="D1" actionID="A1" description="div" />`
	tests := []struct {
		name string
		rows string
		want string
	}{
		{
			name: "dividend without withholding",
			rows: div,
			want: "tx ibkr-flex:acc-ibkr/D1 dividend 2026-03-11 account=acc-ibkr \"div\"\n  holding:PLTEST000010/TSTA -10 USD\n  cash:USD 10 USD\n",
		},
		{
			name: "withholding before its dividend, paired by actionID",
			rows: `<CashTransaction ` + cashTx + `type="Withholding Tax" currency="USD" isin="PLTEST000010" amount="-1.5" dateTime="20260312" transactionID="W1" actionID="A1" />` + div,
			want: "tx ibkr-flex:acc-ibkr/D1 dividend 2026-03-11 account=acc-ibkr \"div\"\n  holding:PLTEST000010/TSTA -10 USD\n  withholding-tax:USD 1.5 USD\n  cash:USD 8.5 USD\n",
		},
		{
			name: "withholding with another actionID",
			rows: div + `<CashTransaction ` + cashTx + `type="Withholding Tax" currency="USD" isin="PLTEST000010" amount="-1.5" dateTime="20260311" transactionID="W1" actionID="A2" />`,
			want: "tx ibkr-flex:acc-ibkr/D1 dividend 2026-03-11 account=acc-ibkr \"div\"\n  holding:PLTEST000010/TSTA -10 USD\n  cash:USD 10 USD\nproblem unsupported FlexStatement#1/CashTransactions#2[transactionID=W1]: withholding tax without its dividend in this statement\n",
		},
		{
			name: "withholding refund",
			rows: div + `<CashTransaction ` + cashTx + `type="Withholding Tax" currency="USD" isin="PLTEST000010" amount="1.5" dateTime="20260311" transactionID="W1" actionID="A1" />`,
			want: "tx ibkr-flex:acc-ibkr/D1 dividend 2026-03-11 account=acc-ibkr \"div\"\n  holding:PLTEST000010/TSTA -10 USD\n  cash:USD 10 USD\nproblem unsupported FlexStatement#1/CashTransactions#2[transactionID=W1]: a withholding tax refund is not supported\n",
		},
		{
			name: "withholding matching two dividends by ISIN and date",
			rows: `<CashTransaction ` + cashTx + `type="Dividends" currency="USD" symbol="TSTA" isin="PLTEST000010" amount="1" dateTime="20260311" transactionID="D1" />` +
				`<CashTransaction ` + cashTx + `type="Dividends" currency="USD" symbol="TSTA" isin="PLTEST000010" amount="2" dateTime="20260311" transactionID="D2" />` +
				`<CashTransaction ` + cashTx + `type="Withholding Tax" currency="USD" isin="PLTEST000010" amount="-0.1" dateTime="20260311" transactionID="W1" />`,
			want: "tx ibkr-flex:acc-ibkr/D1 dividend 2026-03-11 account=acc-ibkr \"\"\n  holding:PLTEST000010/TSTA -1 USD\n  cash:USD 1 USD\n" +
				"tx ibkr-flex:acc-ibkr/D2 dividend 2026-03-11 account=acc-ibkr \"\"\n  holding:PLTEST000010/TSTA -2 USD\n  cash:USD 2 USD\n" +
				"problem unsupported FlexStatement#1/CashTransactions#3[transactionID=W1]: withholding tax matches 2 dividends\n",
		},
		{
			name: "withholding over the gross dividend",
			rows: div + `<CashTransaction ` + cashTx + `type="Withholding Tax" currency="USD" isin="PLTEST000010" amount="-11" dateTime="20260311" transactionID="W1" actionID="A1" />`,
			want: "problem unsupported FlexStatement#1/CashTransactions#1[transactionID=D1]: not a valid ledger transaction\n",
		},
		{
			name: "dividend reversal",
			rows: `<CashTransaction ` + cashTx + `type="Dividends" currency="USD" symbol="TSTA" isin="PLTEST000010" amount="-10" dateTime="20260311" transactionID="D9" />`,
			want: "problem unsupported FlexStatement#1/CashTransactions#1[transactionID=D9]: a dividend reversal is not supported\n",
		},
		{
			name: "deposit under the older type name",
			rows: `<CashTransaction ` + cashTx + `type="Deposits &amp; Withdrawals" currency="EUR" amount="50" dateTime="20260310" transactionID="C1" description="in" />`,
			want: "tx ibkr-flex:acc-ibkr/C1 deposit 2026-03-10 account=acc-ibkr \"in\"\n  cash:EUR 50 EUR\n  equity:EUR -50 EUR\n",
		},
		{
			name: "date falls back to reportDate",
			rows: `<CashTransaction ` + cashTx + `type="Other Fees" currency="USD" amount="-3" reportDate="20260313" transactionID="F1" description="fee" />`,
			want: "tx ibkr-flex:acc-ibkr/F1 fee 2026-03-13 account=acc-ibkr \"fee\"\n  fee:USD 3 USD\n  cash:USD -3 USD\n",
		},
		{
			name: "fee refund",
			rows: `<CashTransaction ` + cashTx + `type="Other Fees" currency="USD" amount="3" dateTime="20260313" transactionID="F2" />`,
			want: "problem unsupported FlexStatement#1/CashTransactions#1[transactionID=F2]: a fee refund is not supported\n",
		},
		{
			name: "no transactionID",
			rows: `<CashTransaction ` + cashTx + `type="Other Fees" currency="USD" amount="-3" dateTime="20260313" />`,
			want: "problem unsupported FlexStatement#1/CashTransactions#1: cash transaction without a transactionID\n",
		},
		{
			name: "no currency",
			rows: `<CashTransaction ` + cashTx + `type="Deposits/Withdrawals" amount="5" dateTime="20260313" transactionID="C2" />`,
			want: "problem unsupported FlexStatement#1/CashTransactions#1[transactionID=C2]: missing or unreadable attribute: currency\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapRendered(t, ingest.ModeTransactions, "<CashTransactions>"+tt.rows+"</CashTransactions>")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mapping (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapOpenPositionRows(t *testing.T) {
	t.Parallel()
	lot := `<OpenPosition accountId="U0000000" currency="USD" assetCategory="STK" symbol="TSTA" isin="PLTEST000010" levelOfDetail="LOT" `
	tests := []struct {
		name string
		mode ingest.Mode
		rows string
		want string
	}{
		{
			name: "lot without a summary row asserts the lot's units",
			mode: ingest.ModeTransactions,
			rows: lot + `position="4" costBasisMoney="40" openDateTime="20250101;100000" />`,
			want: "assert acc-ibkr 2026-03-13 holding:PLTEST000010/TSTA 4\n",
		},
		{
			name: "lot opened after the statement",
			mode: ingest.ModeOpening,
			rows: lot + `position="4" costBasisMoney="40" openDateTime="20260320;100000" />`,
			want: "assert acc-ibkr 2026-03-13 holding:PLTEST000010/TSTA 4\nproblem unsupported FlexStatement#1/OpenPositions#1: not a valid ledger transaction\n",
		},
		{
			name: "lot without an open date",
			mode: ingest.ModeOpening,
			rows: lot + `position="4" costBasisMoney="40" />`,
			want: "assert acc-ibkr 2026-03-13 holding:PLTEST000010/TSTA 4\nproblem unsupported FlexStatement#1/OpenPositions#1: missing or unreadable attribute: openDateTime\n",
		},
		{
			name: "unknown level of detail",
			mode: ingest.ModeTransactions,
			rows: `<OpenPosition accountId="U0000000" currency="USD" assetCategory="STK" symbol="TSTA" isin="PLTEST000010" levelOfDetail="CLOSED_LOT" position="4" />`,
			want: "problem skipped FlexStatement#1/OpenPositions#1: level of detail \"CLOSED_LOT\" is not read\n",
		},
		{
			name: "summary row without a level of detail",
			mode: ingest.ModeTransactions,
			rows: `<OpenPosition accountId="U0000000" currency="USD" assetCategory="STK" symbol="TSTA" isin="PLTEST000010" position="7" />`,
			want: "assert acc-ibkr 2026-03-13 holding:PLTEST000010/TSTA 7\n",
		},
		{
			name: "position without a symbol",
			mode: ingest.ModeTransactions,
			rows: `<OpenPosition accountId="U0000000" currency="USD" assetCategory="STK" isin="PLTEST000010" position="7" />`,
			want: "problem unsupported FlexStatement#1/OpenPositions#1: missing or unreadable attribute: symbol\n",
		},
		{
			name: "foreign element",
			mode: ingest.ModeTransactions,
			rows: `<PriorPeriodPosition accountId="U0000000" />`,
			want: "problem skipped FlexStatement#1/OpenPositions#1: <PriorPeriodPosition> is not an open position\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapRendered(t, tt.mode, "<OpenPositions>"+tt.rows+"</OpenPositions>")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mapping (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMapCashReportRows(t *testing.T) {
	t.Parallel()
	got := mapRendered(t, ingest.ModeOpening, `<CashReport><Other currency="USD" endingCash="1" /></CashReport>`)
	want := "problem skipped FlexStatement#1/CashReport#1[currency=USD]: <Other> is not a cash report row\n"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mapping (-want +got):\n%s", diff)
	}
}
