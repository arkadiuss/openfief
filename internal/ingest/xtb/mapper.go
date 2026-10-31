package xtb

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of a file the mapper refuses whole.
var (
	// ErrUnknownMode reports an ingest mode the mapper doesn't know.
	ErrUnknownMode = errors.New("unknown ingest mode")
	// ErrOpeningMode reports an opening batch: the export holds neither the
	// account's cash balance nor what its positions cost in the account's
	// currency, so it can't seed an account (D-23).
	ErrOpeningMode = errors.New("an xStation export can't seed an account: upload it in transactions mode with the account's whole history")
	// ErrCurrency reports an account currency the file doesn't name, or
	// names in two ways, or that the mapper doesn't know.
	ErrCurrency = errors.New("no account currency")
)

// BrokerSource is the source of every transaction key the mapper writes; the
// core appends ":" and the OpenFief account's ID, never the XTB account
// number (D-72).
const BrokerSource = "xtb"

// minorUnit returns the decimal places of the minor unit of an account
// currency the mapper takes (ISO 4217); XTB's account currencies all have
// two.
func minorUnit(code string) (int32, bool) {
	switch code {
	case "PLN", "EUR", "USD", "GBP", "CZK", "HUF", "RON", "CHF":
		return centPlaces, true
	}
	return 0, false
}

// centPlaces are the decimal places of a currency whose minor unit is a
// hundredth.
const centPlaces = 2

// Mapper maps an xStation "Historia operacji" export (R-24), one account per
// file, into the connector contract's mapping: deposits and withdrawals,
// and the buys, sells and dividends of stocks and ETFs as ticker entries,
// which the core books once the owner confirms each ticker's ISIN, since the
// export names no ISIN (D-116). Everything is named by the XTB account in
// ByBrokerAccount, which the core maps to the OpenFief account the owner
// chose (D-72). Only stocks and ETFs are booked; CFD positions are ingest
// problems (D-24). The comments of deposits and withdrawals, which name the
// sending bank and the XTB account, are never read.
type Mapper struct{}

// export is one file's account, currency and tables.
type export struct {
	account  ingest.BrokerAccountID
	currency kernel.Currency
	places   int32
	date1904 bool
	cash     *table
	closed   *table
	open     *table
}

// Map maps the export data, uploaded under fileName, in mode. It fails when
// the file is not a readable export or mode is opening; rows it can't book
// become ingest problems.
func (Mapper) Map(data []byte, fileName string, mode ingest.Mode) (ingest.Mapping, error) {
	switch mode {
	case ingest.ModeTransactions:
	case ingest.ModeOpening:
		return ingest.Mapping{}, ErrOpeningMode
	default:
		return ingest.Mapping{}, fmt.Errorf("%w: %s", ErrUnknownMode, mode)
	}
	wb, err := readWorkbook(data)
	if err != nil {
		return ingest.Mapping{}, err
	}
	x, err := readExport(wb, fileName)
	if err != nil {
		return ingest.Mapping{}, err
	}
	run := mapping{x: x}
	run.cashOperations()
	run.closedPositions()
	run.openPositions()
	return run.result(), nil
}

// readExport reads the export's account, currency and tables. Cash
// Operations is required; Closed Positions and Open Positions are read
// when present.
func readExport(wb *workbook, fileName string) (*export, error) {
	x := &export{date1904: wb.date1904}
	cash, ok := wb.sheets[sheetCash]
	if !ok {
		return nil, fmt.Errorf("%w: no sheet %q", ErrLayout, sheetCash)
	}
	var err error
	if x.cash, err = findTable(cash, "Type", "Ticker", "Time", "Amount", "ID"); err != nil {
		return nil, err
	}
	if s, ok := wb.sheets[sheetClosed]; ok {
		if x.closed, err = findTable(s, "Ticker", "Category", "Position ID"); err != nil {
			return nil, err
		}
	}
	if s, ok := wb.sheets[sheetOpen]; ok {
		if x.open, err = findTable(s, "Ticker", "Category"); err != nil {
			return nil, err
		}
	}
	if x.account, err = exportAccount(x, fileName); err != nil {
		return nil, err
	}
	if x.currency, x.places, err = exportCurrency(x, fileName); err != nil {
		return nil, err
	}
	return x, nil
}

// tables returns the export's tables that are present.
func (x *export) tables() []*table {
	var out []*table
	for _, t := range []*table{x.cash, x.closed, x.open} {
		if t != nil {
			out = append(out, t)
		}
	}
	return out
}

// exportAccount returns the XTB account every sheet's row 1 names, which
// the file name, when it is xStation's, must name too.
func exportAccount(x *export, fileName string) (ingest.BrokerAccountID, error) {
	var account string
	for _, t := range x.tables() {
		n, err := accountNumber(t.sheet)
		if err != nil {
			return "", err
		}
		if account != "" && n != account {
			return "", fmt.Errorf("%w: the sheets name different accounts", ErrLayout)
		}
		account = n
	}
	if n, ok := nameParts(fileName); ok && n.account != account {
		return "", fmt.Errorf("%w: the file name names another account than its sheets", ErrLayout)
	}
	id := ingest.BrokerAccountID(account)
	if err := id.Validate(); err != nil {
		return "", fmt.Errorf("%w: %w", ErrLayout, err)
	}
	return id, nil
}

// xstationName is what an export's file name says: the account's currency
// and number.
type xstationName struct {
	currency string
	account  string
}

// nameParts returns the currency and account of an xStation export's file
// name, <currency>_<account>_<from>_<to>.xlsx (R-24).
func nameParts(fileName string) (xstationName, bool) {
	base, ok := strings.CutSuffix(fileName, ".xlsx")
	if !ok {
		return xstationName{}, false
	}
	parts := strings.Split(base, "_")
	const nameFields = 4
	if len(parts) != nameFields || len(parts[0]) != len("PLN") || parts[1] == "" {
		return xstationName{}, false
	}
	return xstationName{currency: parts[0], account: parts[1]}, true
}

// exportCurrency returns the account's currency and its minor unit's
// places: from the file name, or from Open Positions' summary, which must
// agree when both name it.
func exportCurrency(x *export, fileName string) (kernel.Currency, int32, error) {
	var named []string
	if n, ok := nameParts(fileName); ok {
		named = append(named, n.currency)
	}
	if x.open != nil {
		for _, c := range summaryCurrencies(x.open.sheet, x.open.header) {
			if !slices.Contains(named, c) {
				named = append(named, c)
			}
		}
	}
	switch len(named) {
	case 0:
		return kernel.Currency{}, 0, fmt.Errorf("%w: name the file as xStation does, <currency>_<account>_<from>_<to>.xlsx", ErrCurrency)
	case 1:
	default:
		return kernel.Currency{}, 0, fmt.Errorf("%w: the file names more than one", ErrCurrency)
	}
	places, ok := minorUnit(named[0])
	if !ok {
		return kernel.Currency{}, 0, fmt.Errorf("%w: %s is not an account currency the mapper knows", ErrCurrency, shown(named[0]))
	}
	c, err := kernel.ParseCurrency(named[0])
	if err != nil {
		return kernel.Currency{}, 0, fmt.Errorf("%w: %w", ErrCurrency, err)
	}
	return c, places, nil
}

// mapping collects what one export maps to.
type mapping struct {
	x        *export
	entries  ingest.BrokerEntries
	tickers  []ingest.Ticker
	problems []ingest.Problem
}

// result returns the mapping, everything named by the export's account.
func (m *mapping) result() ingest.Mapping {
	m.entries.Account = m.x.account
	out := ingest.Mapping{
		Problems:       m.problems,
		BrokerAccounts: []ingest.BrokerAccountID{m.x.account},
		Tickers:        m.tickers,
	}
	if len(m.entries.Transactions) > 0 || len(m.entries.TickerEntries) > 0 {
		out.ByBrokerAccount = []ingest.BrokerEntries{m.entries}
	}
	return out
}

// problem records an ingest problem of the row ref.
func (m *mapping) problem(p ingest.Problem) { m.problems = append(m.problems, p) }

// ticker records the instrument of symbol named name, once.
func (m *mapping) ticker(symbol, name string) {
	if slices.ContainsFunc(m.tickers, func(t ingest.Ticker) bool { return t.Symbol == symbol }) {
		return
	}
	t := splitSymbol(symbol)
	t.Name = shownName(name)
	m.tickers = append(m.tickers, t)
}

// maxNameLen bounds an instrument's display name.
const maxNameLen = 200

// shownName returns an instrument's display name, at most 200 characters.
func shownName(s string) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > maxNameLen {
		r = r[:maxNameLen]
	}
	return string(r)
}

// marketCountry returns the ISO 3166-1 country of an xStation market
// suffix (R-24 saw .PL, .UK and .NL): the suffix itself, but GB for UK; an
// unknown suffix names none.
func marketCountry(suffix string) string {
	switch suffix {
	case "UK":
		return "GB"
	case "PL", "NL", "DE", "FR", "BE", "PT", "IE", "IT", "ES", "US", "DK", "SE", "NO", "FI", "CH", "CZ":
		return suffix
	}
	return ""
}

// splitSymbol returns the ticker of an xStation symbol with its code and
// market country, such as "TESTA" and "PL" for "TESTA.PL".
func splitSymbol(symbol string) ingest.Ticker {
	i := strings.LastIndex(symbol, ".")
	if i <= 0 {
		return ingest.Ticker{Symbol: symbol, Code: symbol}
	}
	return ingest.Ticker{Symbol: symbol, Code: symbol[:i], Country: marketCountry(symbol[i+1:])}
}

// stockCategory reports whether a row's category is a stock or ETF, the only
// instruments booked (D-24). A row without a category counts as one.
func stockCategory(c string) bool {
	switch strings.ToUpper(c) {
	case "", "STOCK", "STOCKS", "ETF":
		return true
	}
	return false
}
