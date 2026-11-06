package xtb

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// The Cash Operations types the mapper books (R-24).
const (
	typeDeposit     = "Deposit"
	typeWithdrawal  = "Withdrawal"
	typePurchase    = "Stock purchase"
	typeSell        = "Stock sell"
	typeSale        = "Stock sale"
	typeDividend    = "Dividend"
	typeWithholding = "Withholding tax"
	typeInterest    = "Free funds interest"
	typeInterestTax = "Free funds interest tax"
)

// The Cash Operations columns the mapper reads. The comment is read only for
// a purchase or sale; a deposit's names the sending bank and a
// withdrawal's the XTB account, and neither is ever read.
const (
	colType     = "Type"
	colTicker   = "Ticker"
	colName     = "Instrument"
	colCategory = "Category"
	colTime     = "Time"
	colAmount   = "Amount"
	colID       = "ID"
	colComment  = "Comment"
)

// tradeComment is a purchase's or sale's comment, such as
// "OPEN BUY 10/10 @ 12.345": the side, the volume filled, of the order or
// position, and the price (R-24).
var tradeComment = regexp.MustCompile(`^(OPEN|CLOSE) (BUY|SELL) ([0-9]+(?:\.[0-9]+)?)/([0-9]+(?:\.[0-9]+)?) @ [0-9]+(?:\.[0-9]+)?$`)

// errLedger reports a row whose transaction the ledger rejects.
var errLedger = errors.New("not a valid ledger transaction")

// problemError is a row's failure that becomes an ingest problem.
type problemError struct {
	kind   ingest.ProblemKind
	reason string
}

// Error returns the reason.
func (e *problemError) Error() string { return e.reason }

// unsupported returns the error of a row OpenFief cannot book yet.
func unsupported(format string, args ...any) error {
	return &problemError{kind: ingest.ProblemUnsupported, reason: fmt.Sprintf(format, args...)}
}

// skipped returns the error of a row left out on purpose.
func skipped(format string, args ...any) error {
	return &problemError{kind: ingest.ProblemSkipped, reason: fmt.Sprintf(format, args...)}
}

// problemOf turns a row's error into an ingest problem of ref. An error
// from the ledger is reduced to its kind, since it can carry amounts.
func problemOf(ref string, err error) ingest.Problem {
	if pe, ok := errors.AsType[*problemError](err); ok {
		return ingest.Problem{Kind: pe.kind, Ref: ref, Reason: pe.reason}
	}
	if errors.Is(err, errCellValue) {
		return ingest.Unsupported(ref, err.Error())
	}
	return ingest.Unsupported(ref, errLedger.Error())
}

// cashRow is one Cash Operations row read for mapping.
type cashRow struct {
	rec    record
	ref    string
	id     string
	kind   string
	on     kernel.Date
	amount kernel.Money
	symbol string
}

// cashOperations maps Cash Operations, oldest first: the sheet lists them
// newest first (R-24). Dividends and free-funds interest are booked with
// the tax withheld from them, which has rows of its own.
func (m *mapping) cashOperations() {
	rows := m.readCashRows()
	withheld := m.pairWithholding(rows.byType[typeDividend], rows.byType[typeWithholding])
	maps.Copy(withheld, m.pairInterestTax(rows.byType[typeInterest], rows.byType[typeInterestTax]))
	for _, r := range rows.all {
		if r.kind == typeWithholding || r.kind == typeInterestTax {
			continue
		}
		if err := m.cashOperation(r, withheld[r]); err != nil {
			m.problem(problemOf(r.ref, err))
		}
	}
}

// cashRows are the Cash Operations rows read, oldest first, and by type.
type cashRows struct {
	all    []*cashRow
	byType map[string][]*cashRow
}

// readCashRows reads Cash Operations, oldest first, and returns the rows
// it could read; a row it can't read is an ingest problem.
func (m *mapping) readCashRows() cashRows {
	recs := m.x.cash.records()
	slices.Reverse(recs)
	out := cashRows{byType: map[string][]*cashRow{}}
	for _, rec := range recs {
		r, err := m.readCash(rec)
		if err != nil {
			m.problem(problemOf(m.cashRef(rec, ""), err))
			continue
		}
		out.byType[r.kind] = append(out.byType[r.kind], r)
		out.all = append(out.all, r)
	}
	return out
}

// cashRef names a Cash Operations row in a problem, with its ID when known.
func (m *mapping) cashRef(rec record, id string) string {
	ref := fmt.Sprintf("%s#%d", sheetCash, rec.n)
	if id != "" {
		ref += "[ID=" + shown(id) + "]"
	}
	return ref
}

// readCash reads the fields every Cash Operations row has: its ID, type,
// UTC date, amount in the account's currency and ticker.
func (m *mapping) readCash(rec record) (*cashRow, error) {
	id, err := rec.id(colID)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, unsupported("cash operation without an ID")
	}
	r := &cashRow{rec: rec, ref: m.cashRef(rec, id), id: id, kind: rec.text(colType), symbol: rec.text(colTicker)}
	if r.on, err = rec.date(colTime, m.x.date1904); err != nil {
		return nil, err
	}
	if r.amount, err = m.money(rec, colAmount); err != nil {
		return nil, err
	}
	return r, nil
}

// money returns the amount in column label in the account's currency,
// rounded to its minor unit: amounts are binary floating-point cells, whose
// 13th to 15th decimals can carry artifacts (R-24). An amount that differs
// from its rounding by more than such an artifact is refused, not rounded.
func (m *mapping) money(rec record, label string) (kernel.Money, error) {
	d, err := rec.decimal(label)
	if err != nil {
		return kernel.Money{}, err
	}
	rounded, err := d.Round(m.x.places, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Money{}, rec.errorf(label)
	}
	if !artifact(d.Sub(rounded)) {
		return kernel.Money{}, unsupported("an amount with more decimals than the currency's minor unit is not supported")
	}
	return kernel.NewMoney(rounded, m.x.currency), nil
}

// artifactExponent makes 1e-9 the largest difference from the rounded
// amount taken as a floating-point artifact.
const artifactExponent = -9

// artifact reports whether diff is within 1e-9 of zero.
func artifact(diff kernel.Decimal) bool {
	bound := kernel.NewDecimal(1, artifactExponent)
	return diff.Cmp(bound) <= 0 && diff.Neg().Cmp(bound) <= 0
}

// standIn is the account a transaction is built in before the core knows
// the OpenFief account; it never leaves the mapper.
const standIn = ledger.AccountID("xtb-account")

// header returns the transaction header of row r: its key is the
// operation's ID, under BrokerSource.
func (r *cashRow) header() ledger.TransactionHeader {
	return ledger.TransactionHeader{Key: ledger.TransactionKey{Source: BrokerSource, ExternalID: r.id}, Date: r.on}
}

// cashOperation books one Cash Operations row other than withholding tax;
// withheld is the tax paired with a dividend.
func (m *mapping) cashOperation(r *cashRow, withheld kernel.Money) error {
	if r.amount.IsZero() {
		return skipped("a cash operation of no amount")
	}
	switch r.kind {
	case typeDeposit:
		return m.transfer(r, true)
	case typeWithdrawal:
		return m.transfer(r, false)
	case typePurchase, typeSell, typeSale:
		return m.trade(r)
	case typeDividend:
		return m.dividend(r, withheld)
	case typeInterest:
		return m.interest(r, withheld)
	case typeInterestTax:
	}
	return unsupported("cash operation type %q is not supported", shown(r.kind))
}

// transfer books a deposit (in) or a withdrawal. Its description is fixed:
// the comment, which names the sending bank or the XTB account, is never
// read.
func (m *mapping) transfer(r *cashRow, in bool) error {
	var (
		t   ledger.Transaction
		err error
	)
	// The ledger's rules need an account: a stand-in, removed at once, as
	// the core names the account (D-72).
	h := r.header()
	h.Account = standIn
	switch {
	case in && r.amount.Sign() > 0:
		t, err = ledger.NewDeposit(h, r.amount)
		t.Description = "deposit"
	case !in && r.amount.Sign() < 0:
		t, err = ledger.NewWithdrawal(h, r.amount.Neg())
		t.Description = "withdrawal"
	default:
		return unsupported("a %s of that sign is not supported", shown(r.kind))
	}
	if err != nil {
		return fmt.Errorf("%w: %w", errLedger, err)
	}
	t.Account = ""
	m.entries.Transactions = append(m.entries.Transactions, t)
	return nil
}

// interest books free-funds interest, XTB's interest on uninvested cash,
// as cash income (D-152), with the tax withheld from it, which has a row
// of its own: income gives the interest, withholding tax takes the tax and
// cash receives the rest. Its description is fixed; the comment, which
// names the month, isn't read.
func (m *mapping) interest(r *cashRow, withheld kernel.Money) error {
	if r.amount.Sign() < 0 {
		return unsupported("negative free-funds interest is not supported")
	}
	h := r.header()
	h.Account = standIn
	t, err := ledger.NewCashIncome(h, r.amount, withheld)
	if err != nil {
		return fmt.Errorf("%w: %w", errLedger, err)
	}
	t.Account, t.Description = "", "free-funds interest"
	m.entries.Transactions = append(m.entries.Transactions, t)
	return nil
}

// trade books a stock purchase or sale as a ticker entry: its units come
// from the comment, its cost or proceeds are the amount, in the account's
// currency.
func (m *mapping) trade(r *cashRow) error {
	if err := m.needStock(r); err != nil {
		return err
	}
	side, qty, err := tradeVolume(r)
	if err != nil {
		return err
	}
	e := ingest.TickerEntry{Symbol: r.symbol, Header: r.header(), Quantity: qty, Ref: r.ref}
	buy := r.kind == typePurchase
	switch {
	case buy && side == "OPEN BUY" && r.amount.Sign() < 0:
		e.Kind, e.Amount, e.Description = ledger.KindBuy, r.amount.Neg(), "buy "+r.symbol
	case !buy && side == "CLOSE BUY" && r.amount.Sign() > 0:
		e.Kind, e.Amount, e.Description = ledger.KindSell, r.amount, "sell "+r.symbol
	default:
		return unsupported("a %s of that side or sign is not supported", shown(r.kind))
	}
	return m.tickerEntry(r, &e)
}

// tradeVolume returns a purchase's or sale's side, such as "OPEN BUY", and
// the volume filled, from its comment.
func tradeVolume(r *cashRow) (string, kernel.Quantity, error) {
	parts := tradeComment.FindStringSubmatch(r.rec.text(colComment))
	if parts == nil {
		return "", kernel.Quantity{}, unsupported("a %s without a comment naming its volume is not supported", shown(r.kind))
	}
	qty, err := kernel.ParseQuantity(parts[3])
	if err != nil || qty.Sign() <= 0 {
		return "", kernel.Quantity{}, unsupported("a %s without a positive volume is not supported", shown(r.kind))
	}
	return parts[1] + " " + parts[2], qty, nil
}

// dividend books a dividend as a ticker entry, with the tax withheld from
// it.
func (m *mapping) dividend(r *cashRow, withheld kernel.Money) error {
	if err := m.needStock(r); err != nil {
		return err
	}
	if r.amount.Sign() < 0 {
		return unsupported("a dividend reversal is not supported")
	}
	e := ingest.TickerEntry{
		Symbol: r.symbol, Header: r.header(), Kind: ledger.KindDividend,
		Amount: r.amount, Charge: withheld, Description: "dividend " + r.symbol, Ref: r.ref,
	}
	return m.tickerEntry(r, &e)
}

// tickerEntry adds e, checked, and its ticker.
func (m *mapping) tickerEntry(r *cashRow, e *ingest.TickerEntry) error {
	return m.addEntry(r.rec.text(colName), e)
}

// addEntry adds e, checked, and its ticker, named name.
func (m *mapping) addEntry(name string, e *ingest.TickerEntry) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("%w: %w", errLedger, err)
	}
	m.ticker(e.Symbol, name)
	m.entries.TickerEntries = append(m.entries.TickerEntries, *e)
	return nil
}

// needStock fails unless r names a ticker of a stock or ETF (D-24).
func (m *mapping) needStock(r *cashRow) error {
	if r.symbol == "" {
		return unsupported("a %s without a ticker is not supported", shown(r.kind))
	}
	if c := r.rec.text(colCategory); !stockCategory(c) {
		return unsupported("category %q is not a stock or ETF (D-24)", shown(c))
	}
	if err := (&ingest.Ticker{Symbol: r.symbol}).Validate(); err != nil {
		return unsupported("the ticker is not valid")
	}
	return nil
}

// pairInterestTax returns the tax withheld from each free-funds interest
// row: each tax row belongs to the one interest row of its date, and
// several rows of one interest add up. A tax row with no such interest,
// more than one, or a positive amount (a refund) is an ingest problem.
func (m *mapping) pairInterestTax(interest, taxes []*cashRow) map[*cashRow]kernel.Money {
	return m.pairTax(interest, taxes, taxNames{"free-funds interest tax", "free-funds interest", "free-funds interest rows"},
		func(_, _ *cashRow) bool { return true })
}

// pairWithholding returns the tax withheld from each dividend: each
// withholding tax row belongs to the one dividend of its ticker on its
// date, and several rows of one dividend add up. A row with no such
// dividend, more than one, or a positive amount (a refund) is an ingest
// problem.
func (m *mapping) pairWithholding(dividends, withholdings []*cashRow) map[*cashRow]kernel.Money {
	return m.pairTax(dividends, withholdings, taxNames{"withholding tax", "dividend", "dividends"},
		func(d, w *cashRow) bool { return d.symbol == w.symbol })
}

// taxNames name a tax row and its income, once and as several, in
// problems.
type taxNames struct {
	tax, income, incomes string
}

// pairTax returns the tax each of incomes was withheld: each of taxes,
// named by names, belongs to the one income dated on its date that same
// accepts, and several taxes of one income add up. A tax with no such
// income, more than one, or a positive amount (a refund) is an ingest
// problem.
func (m *mapping) pairTax(incomes, taxes []*cashRow, names taxNames, same func(income, tax *cashRow) bool) map[*cashRow]kernel.Money {
	withheld := map[*cashRow]kernel.Money{}
	for _, w := range taxes {
		var found []*cashRow
		for _, d := range incomes {
			if same(d, w) && d.on.Equal(w.on) {
				found = append(found, d)
			}
		}
		switch {
		case w.amount.Sign() >= 0:
			m.problem(ingest.Unsupported(w.ref, "a "+names.tax+" refund is not supported"))
			continue
		case len(found) == 0:
			m.problem(ingest.Unsupported(w.ref, names.tax+" without its "+names.income+" in this file"))
			continue
		case len(found) > 1:
			m.problem(ingest.Unsupported(w.ref, fmt.Sprintf("%s matches %d %s", names.tax, len(found), names.incomes)))
			continue
		}
		d := found[0]
		sum, ok := withheld[d]
		if !ok {
			withheld[d] = w.amount.Neg()
			continue
		}
		total, err := sum.Add(w.amount.Neg())
		if err != nil {
			m.problem(ingest.Unsupported(w.ref, names.tax+" in another currency than its "+names.income))
			continue
		}
		withheld[d] = total
	}
	return withheld
}
