package demo

import (
	"sort"
	"strconv"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// flexDoc writes a synthetic Flex Query response: the format IBKR's Flex
// Web Service returns, so that the demo's IBKR history enters the ledger
// through the real ibkrflex connector. Only the attributes the connector
// reads, plus a few for realism, are written.
type flexDoc struct {
	b strings.Builder
}

// The Flex attributes every row of the demo's statements has.
const (
	attrAccount  = "accountId"
	attrLevel    = "levelOfDetail"
	attrCurrency = "currency"
)

// escapeAttr escapes an attribute value.
func escapeAttr(v string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;").Replace(v)
}

// elem writes an element with attrs, given as name, value pairs; closed
// makes it self-closing.
func (x *flexDoc) elem(name string, closed bool, attrs ...string) {
	x.b.WriteString("<" + name)
	for i := 0; i+1 < len(attrs); i += 2 {
		x.b.WriteString(" " + attrs[i] + `="` + escapeAttr(attrs[i+1]) + `"`)
	}
	if closed {
		x.b.WriteString(" />\n")
		return
	}
	x.b.WriteString(">\n")
}

// end closes element name.
func (x *flexDoc) end(name string) { x.b.WriteString("</" + name + ">\n") }

// begin writes the response's header and opens its one statement for
// [from, to].
func (x *flexDoc) begin(from, to kernel.Date) {
	x.b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	x.b.WriteString("<!-- OpenFief demo dataset: a synthetic Flex Query statement. Fake account, fake IDs, " +
		"synthetic ISINs (" + SyntheticMarker + "), generated prices. No real data. -->\n")
	x.elem("FlexQueryResponse", false, "queryName", "openfief-demo", "type", "AF")
	x.elem("FlexStatements", false, "count", "1")
	x.elem("FlexStatement", false, attrAccount, IBKRFlexAccount, "fromDate", from.String(), "toDate", to.String(),
		"period", "Custom", "whenGenerated", to.String()+";235900")
}

// finish closes the statement and the response and returns the document.
func (x *flexDoc) finish() []byte {
	x.end("FlexStatement")
	x.end("FlexStatements")
	x.end("FlexQueryResponse")
	return []byte(x.b.String())
}

// stamp returns d as a Flex date and time.
func stamp(d kernel.Date) string { return d.String() + ";120000" }

// amount returns n hundredths as a decimal string.
func amount(n int64) string { return cents(n).String() }

// openingFlex returns the statement an account is seeded from at go-live
// (D-23): its lot-level open positions, with a summary row per instrument,
// and its cash.
func openingFlex(h *ibkrHistory) []byte {
	const daysCovered = 4
	var x flexDoc
	x.begin(h.goLive.AddDays(-daysCovered), h.goLive)
	x.elem("OpenPositions", false)
	total := map[*Instrument]int64{}
	for i := range h.lots {
		total[h.lots[i].in] += h.lots[i].units
	}
	for _, in := range sortedInstruments(total) {
		x.position(in, total[in], h.goLive, "SUMMARY")
		for i := range h.lots {
			if l := &h.lots[i]; l.in == in {
				x.elem("OpenPosition", true, append(positionAttrs(in, l.units, h.goLive, "LOT"),
					"costBasisMoney", amount(l.costCents), "openDateTime", stamp(l.opened), "originatingTransactionID", l.id)...)
			}
		}
	}
	x.end("OpenPositions")
	x.cashReport(h.openingCash, h.goLive)
	return x.finish()
}

// activityFlex returns one period's activity statement: its trades and cash
// transactions, the open positions at its end and its cash report.
func activityFlex(st *statement) []byte {
	var x flexDoc
	x.begin(st.from, st.to)
	x.elem("OpenPositions", false)
	for _, in := range sortedInstruments(st.end.units) {
		x.position(in, st.end.units[in], st.to, "SUMMARY")
	}
	x.end("OpenPositions")
	x.cashReport(st.end.cash, st.to)
	x.elem("Trades", false)
	for i := range st.ops {
		x.trade(&st.ops[i])
	}
	x.end("Trades")
	x.elem("CashTransactions", false)
	for i := range st.ops {
		x.cashTransaction(&st.ops[i])
	}
	x.end("CashTransactions")
	return x.finish()
}

// positionAttrs returns an OpenPosition row's attributes.
func positionAttrs(in *Instrument, units int64, on kernel.Date, level string) []string {
	return []string{
		attrAccount, IBKRFlexAccount, attrCurrency, in.Currency.String(), "assetCategory", "STK", "symbol", in.Instrument.Symbol,
		"description", in.Name, "isin", in.Instrument.ISIN.String(), "multiplier", "1", "reportDate", on.String(),
		"position", strconv.FormatInt(units, 10), attrLevel, level, "side", "Long",
	}
}

// position writes an OpenPosition row.
func (x *flexDoc) position(in *Instrument, units int64, on kernel.Date, level string) {
	x.elem("OpenPosition", true, positionAttrs(in, units, on, level)...)
}

// cashReport writes the CashReport section: the base-currency summary row
// IBKR always adds, then the ending cash of each currency.
func (x *flexDoc) cashReport(cash map[kernel.Currency]int64, on kernel.Date) {
	x.elem("CashReport", false)
	x.elem("CashReportCurrency", true, attrAccount, IBKRFlexAccount, attrCurrency, "BASE_SUMMARY",
		attrLevel, "BaseCurrency", "toDate", on.String(), "endingCash", "0")
	for _, c := range sortedCurrencies(cash) {
		x.elem("CashReportCurrency", true, attrAccount, IBKRFlexAccount, attrCurrency, c.String(),
			attrLevel, "Currency", "toDate", on.String(), "endingCash", amount(cash[c]))
	}
	x.end("CashReport")
}

// trade writes the Trade row of a buy, a sale or an FX conversion; other
// operations have none.
func (x *flexDoc) trade(o *op) {
	common := []string{
		attrAccount, IBKRFlexAccount, attrCurrency, o.currency.String(), "tradeID", o.id, "tradeDate", o.date.String(),
		"dateTime", o.date.String() + ";101500", "transactionType", "ExchTrade",
		"ibCommission", amount(-o.fee), "ibCommissionCurrency", o.currency.String(), attrLevel, "EXECUTION",
	}
	switch o.kind {
	case opBuy, opSell:
		qty, proceeds, side := o.units, -o.cents, "BUY"
		if o.kind == opSell {
			qty, proceeds, side = -o.units, o.cents, "SELL"
		}
		x.elem("Trade", true, append(common, "assetCategory", "STK", "symbol", o.in.Instrument.Symbol,
			"isin", o.in.Instrument.ISIN.String(), "multiplier", "1", "buySell", side, "quantity", strconv.FormatInt(qty, 10),
			"tradePrice", amount(o.cents/o.units), "proceeds", amount(proceeds), "taxes", "0")...)
	case opFX:
		x.elem("Trade", true, append(common, "assetCategory", "CASH", "symbol", o.base.String()+"."+o.currency.String(),
			"isin", "", "buySell", "BUY", "quantity", strconv.FormatInt(o.baseUnits, 10), "proceeds", amount(-o.cents))...)
	case opDeposit, opDividend, opFee:
	}
}

// cashTransaction writes the CashTransaction rows of a deposit, a fee or a
// dividend. A dividend gets three: the dividend, the tax withheld from it
// under the same actionID, and the SUMMARY row that IBKR adds when a Flex
// Query asks for summaries too. The connector skips the summary row on
// purpose: it is the demo's one intended ingest problem.
func (x *flexDoc) cashTransaction(o *op) {
	row := func(typ, id, level string, n int64, desc string, extra ...string) {
		attrs := append([]string{
			attrAccount, IBKRFlexAccount, attrCurrency, o.currency.String(), "type", typ, "description", desc,
			"amount", amount(n), "dateTime", stamp(o.date), "transactionID", id, attrLevel, level,
		}, extra...)
		x.elem("CashTransaction", true, attrs...)
	}
	switch o.kind {
	case opDeposit:
		row("Deposits/Withdrawals", o.id, "DETAIL", o.cents, "CASH RECEIPTS / ELECTRONIC FUND TRANSFERS")
	case opFee:
		row("Other Fees", o.id, "DETAIL", -o.cents, o.desc)
	case opDividend:
		sym, isin, seq := o.in.Instrument.Symbol, o.in.Instrument.ISIN.String(), o.id[2:]
		instrument := []string{"assetCategory", "STK", "symbol", sym, "isin", isin, "actionID", "73" + seq}
		row("Dividends", o.id, "DETAIL", o.cents, sym+" CASH DIVIDEND (DEMO)", instrument...)
		row("Withholding Tax", "92"+seq, "DETAIL", -o.withheld, sym+" CASH DIVIDEND - TAX (DEMO)", instrument...)
		row("Dividends", "", "SUMMARY", o.cents, sym+" DIVIDEND TOTAL (DEMO)", "assetCategory", "STK", "symbol", sym, "isin", isin)
	case opBuy, opSell, opFX:
	}
}

// sortedInstruments returns the keys of units in ISIN order.
func sortedInstruments(units map[*Instrument]int64) []*Instrument {
	out := make([]*Instrument, 0, len(units))
	for in := range units {
		out = append(out, in)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Instrument.ISIN.String() < out[j].Instrument.ISIN.String() })
	return out
}

// sortedCurrencies returns the keys of cash in code order.
func sortedCurrencies(cash map[kernel.Currency]int64) []kernel.Currency {
	out := make([]kernel.Currency, 0, len(cash))
	for c := range cash {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
