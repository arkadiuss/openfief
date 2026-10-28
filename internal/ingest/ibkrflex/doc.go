// Package ibkrflex is the Interactive Brokers Flex Query connector's adapter
// and anti-corruption layer: it parses a Flex Query XML statement
// (FlexQueryResponse > FlexStatements > FlexStatement) and translates IBKR's
// sections into ledger transactions, balance assertions and ingest problems.
// It also fetches statements from the Flex Web Service, the daily-pull
// ingest path. IBKR vocabulary (position, trade, conid, Flex section and
// attribute names) stays in this package.
//
// The schema follows IBKR's public Flex Query reference; OQ-22 has not yet
// verified it against a real account. What is read:
//
//   - OpenPositions: LOT rows become opening lots in opening mode (D-23);
//     the SUMMARY row of an instrument, or else the sum of its LOT rows,
//     becomes a holding balance assertion.
//   - CashReport: endingCash per currency becomes a cash balance assertion
//     on the statement's last day, and in opening mode the opening cash.
//   - Trades: EXECUTION rows of stocks and ETFs (STK) become buys and sells
//     keyed by tradeID; CASH rows (FX trades such as EUR.USD) become FX
//     conversions.
//   - CashTransactions: DETAIL rows of Dividends (paired with their
//     Withholding Tax rows), Deposits/Withdrawals and Other Fees, keyed by
//     transactionID.
//
// Everything else is an ingest problem, not an error. ConversionRates are
// IBKR's own rates to the base currency and are ignored: FX rates come from
// NBP (D-20).
//
// Dates: IBKR writes dates as yyyyMMdd or yyyy-MM-dd and date-times with a
// ";HHmmss" (or other separator) suffix. Only the calendar date is kept, as
// IBKR reports it: it is not converted between time zones (IBKR reports
// trade times in US Eastern time).
//
// Hostile statements (DESIGN §3.5): input is capped in size, nesting depth,
// element count, attributes per element and attribute length; DOCTYPE and
// other directives are rejected, so no DTD or external entity is ever read
// (encoding/xml doesn't resolve external entities anyway, and Strict mode
// rejects unknown entity references); only UTF-8 input is accepted. Amounts
// are parsed from attribute text as exact decimals, never through float64.
package ibkrflex
