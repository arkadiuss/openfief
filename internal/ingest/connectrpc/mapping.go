package connectrpc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// ErrInvalidMessage reports a connector's message that does not translate
// into the domain: a malformed decimal, date, currency, ISIN or enum value, or a
// missing required field. The error names the field.
var ErrInvalidMessage = errors.New("invalid message")

// invalid returns an ErrInvalidMessage for field, wrapping cause when given.
func invalid(field string, cause error) error {
	if cause == nil {
		return fmt.Errorf("%w: %s", ErrInvalidMessage, field)
	}
	return fmt.Errorf("%w: %s: %w", ErrInvalidMessage, field, cause)
}

// plainDecimal reports whether s is a decimal as the contract writes it: an
// optional minus sign, digits, and an optional point followed by digits.
// Exponents, grouping, a plus sign and surrounding space are refused.
func plainDecimal(s string) bool {
	whole, frac, point := strings.Cut(strings.TrimPrefix(s, "-"), ".")
	return allDigits(whole) && (!point || allDigits(frac))
}

// allDigits reports whether s is one or more ASCII digits.
func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// translate translates each contract message of a repeated field with from,
// naming each element field[i] in errors.
func translate[P, T any](field string, in []P, from func(string, P) (T, error)) ([]T, error) {
	out := make([]T, len(in))
	for i, p := range in {
		var err error
		if out[i], err = from(fmt.Sprintf("%s[%d]", field, i), p); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// decimalFrom parses a contract decimal.
func decimalFrom(field, s string) (kernel.Decimal, error) {
	if !plainDecimal(s) {
		return kernel.Decimal{}, invalid(fmt.Sprintf("%s %q is not a plain decimal", field, s), nil)
	}
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		return kernel.Decimal{}, invalid(field, err)
	}
	return d, nil
}

// dateFrom parses a contract date, YYYY-MM-DD.
func dateFrom(field, s string) (kernel.Date, error) {
	d, err := kernel.ParseDate(s)
	if err != nil {
		return kernel.Date{}, invalid(field, err)
	}
	return d, nil
}

// optionalDateFrom parses a contract date that may be empty: empty is no date.
func optionalDateFrom(field, s string) (kernel.Date, error) {
	if s == "" {
		return kernel.Date{}, nil
	}
	return dateFrom(field, s)
}

// currencyFrom parses an ISO 4217 code.
func currencyFrom(field, s string) (kernel.Currency, error) {
	c, err := kernel.ParseCurrency(s)
	if err != nil {
		return kernel.Currency{}, invalid(field, err)
	}
	return c, nil
}

// moneyFrom translates a contract amount; a missing one is an error.
func moneyFrom(field string, m *connectorv1.Money) (kernel.Money, error) {
	if m == nil {
		return kernel.Money{}, invalid(field+" is missing", nil)
	}
	c, err := currencyFrom(field+".currency", m.GetCurrency())
	if err != nil {
		return kernel.Money{}, err
	}
	d, err := decimalFrom(field+".amount", m.GetAmount())
	if err != nil {
		return kernel.Money{}, err
	}
	return kernel.NewMoney(d, c), nil
}

// instrumentFrom translates an instrument: a valid ISIN plus a venue symbol
// (D-35).
func instrumentFrom(field string, in *connectorv1.Instrument) (ledger.Instrument, error) {
	isin, err := ledger.ParseISIN(in.GetIsin())
	if err != nil {
		return ledger.Instrument{}, invalid(field+".isin", err)
	}
	out := ledger.Instrument{ISIN: isin, Symbol: in.GetSymbol()}
	if err := out.Validate(); err != nil {
		return ledger.Instrument{}, invalid(field, err)
	}
	return out, nil
}

// bookKindFrom translates a ledger account kind.
func bookKindFrom(field string, k connectorv1.BookKind) (ledger.BookKind, error) {
	switch k {
	case connectorv1.BookKind_BOOK_KIND_CASH:
		return ledger.BookCash, nil
	case connectorv1.BookKind_BOOK_KIND_HOLDING:
		return ledger.BookHolding, nil
	case connectorv1.BookKind_BOOK_KIND_FEE:
		return ledger.BookFee, nil
	case connectorv1.BookKind_BOOK_KIND_WITHHOLDING_TAX:
		return ledger.BookWithholdingTax, nil
	case connectorv1.BookKind_BOOK_KIND_FX:
		return ledger.BookFX, nil
	case connectorv1.BookKind_BOOK_KIND_EQUITY:
		return ledger.BookEquity, nil
	case connectorv1.BookKind_BOOK_KIND_UNSPECIFIED:
	}
	return 0, invalid(fmt.Sprintf("%s.kind %s", field, k), nil)
}

// bookFrom translates a ledger account: a holding names its instrument, any
// other kind its currency.
func bookFrom(field string, b *connectorv1.Book) (ledger.Book, error) {
	kind, err := bookKindFrom(field, b.GetKind())
	if err != nil {
		return ledger.Book{}, err
	}
	if kind == ledger.BookHolding {
		in, inErr := instrumentFrom(field+".instrument", b.GetInstrument())
		if inErr != nil {
			return ledger.Book{}, inErr
		}
		return ledger.HoldingBook(in), nil
	}
	c, err := currencyFrom(field+".currency", b.GetCurrency())
	if err != nil {
		return ledger.Book{}, err
	}
	return ledger.Book{Kind: kind, Currency: c}, nil
}

// postingFrom translates one posting; its quantity may be empty.
func postingFrom(field string, p *connectorv1.Posting) (ledger.Posting, error) {
	book, err := bookFrom(field+".book", p.GetBook())
	if err != nil {
		return ledger.Posting{}, err
	}
	amount, err := moneyFrom(field+".amount", p.GetAmount())
	if err != nil {
		return ledger.Posting{}, err
	}
	out := ledger.Posting{Book: book, Amount: amount}
	if p.GetQuantity() != "" {
		q, err := decimalFrom(field+".quantity", p.GetQuantity())
		if err != nil {
			return ledger.Posting{}, err
		}
		out.Quantity = kernel.NewQuantity(q)
	}
	return out, nil
}

// transactionKinds returns a map of every contract transaction kind to the ledger's.
func transactionKinds() map[connectorv1.TransactionKind]ledger.TransactionKind {
	return map[connectorv1.TransactionKind]ledger.TransactionKind{
		connectorv1.TransactionKind_TRANSACTION_KIND_BUY:           ledger.KindBuy,
		connectorv1.TransactionKind_TRANSACTION_KIND_SELL:          ledger.KindSell,
		connectorv1.TransactionKind_TRANSACTION_KIND_DIVIDEND:      ledger.KindDividend,
		connectorv1.TransactionKind_TRANSACTION_KIND_FEE:           ledger.KindFee,
		connectorv1.TransactionKind_TRANSACTION_KIND_FX_CONVERSION: ledger.KindFXConversion,
		connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT:       ledger.KindDeposit,
		connectorv1.TransactionKind_TRANSACTION_KIND_WITHDRAWAL:    ledger.KindWithdrawal,
		connectorv1.TransactionKind_TRANSACTION_KIND_OPENING_LOT:   ledger.KindOpeningLot,
		connectorv1.TransactionKind_TRANSACTION_KIND_INTEREST:      ledger.KindInterest,
		connectorv1.TransactionKind_TRANSACTION_KIND_DELETION:      ledger.KindDeletion,
	}
}

// transactionKindFrom translates a transaction kind.
func transactionKindFrom(field string, k connectorv1.TransactionKind) (ledger.TransactionKind, error) {
	if kind, ok := transactionKinds()[k]; ok {
		return kind, nil
	}
	return 0, invalid(fmt.Sprintf("%s.kind %s", field, k), nil)
}

// transactionFrom translates one transaction. The ledger validates it when
// it is imported; this only translates.
func transactionFrom(field string, t *connectorv1.Transaction) (ledger.Transaction, error) {
	kind, err := transactionKindFrom(field, t.GetKind())
	if err != nil {
		return ledger.Transaction{}, err
	}
	date, err := dateFrom(field+".date", t.GetDate())
	if err != nil {
		return ledger.Transaction{}, err
	}
	opened, err := optionalDateFrom(field+".lot_opened", t.GetLotOpened())
	if err != nil {
		return ledger.Transaction{}, err
	}
	postings, err := translate(field+".postings", t.GetPostings(), postingFrom)
	if err != nil {
		return ledger.Transaction{}, err
	}
	return ledger.Transaction{
		Key:     ledger.TransactionKey{Source: t.GetKey().GetSource(), ExternalID: t.GetKey().GetExternalId()},
		Account: ledger.AccountID(t.GetAccountId()), Kind: kind, Date: date, LotOpened: opened,
		Description: t.GetDescription(), Postings: postings,
	}, nil
}

// assertionFrom translates one balance assertion.
func assertionFrom(field string, a *connectorv1.BalanceAssertion) (ledger.BalanceAssertion, error) {
	date, err := dateFrom(field+".date", a.GetDate())
	if err != nil {
		return ledger.BalanceAssertion{}, err
	}
	book, err := bookFrom(field+".book", a.GetBook())
	if err != nil {
		return ledger.BalanceAssertion{}, err
	}
	balance, err := decimalFrom(field+".balance", a.GetBalance())
	if err != nil {
		return ledger.BalanceAssertion{}, err
	}
	return ledger.BalanceAssertion{Account: ledger.AccountID(a.GetAccountId()), Date: date, Book: book, Balance: balance}, nil
}

// problemFrom translates one ingest problem.
func problemFrom(field string, p *connectorv1.Problem) (ingest.Problem, error) {
	switch p.GetKind() {
	case connectorv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED:
		return ingest.Unsupported(p.GetRef(), p.GetReason()), nil
	case connectorv1.ProblemKind_PROBLEM_KIND_SKIPPED:
		return ingest.Skipped(p.GetRef(), p.GetReason()), nil
	case connectorv1.ProblemKind_PROBLEM_KIND_UNSPECIFIED:
	}
	return ingest.Problem{}, invalid(fmt.Sprintf("%s.kind %s", field, p.GetKind()), nil)
}

// modeTo translates an ingest mode into the contract.
func modeTo(m ingest.Mode) (connectorv1.Mode, error) {
	switch m {
	case ingest.ModeOpening:
		return connectorv1.Mode_MODE_OPENING, nil
	case ingest.ModeTransactions:
		return connectorv1.Mode_MODE_TRANSACTIONS, nil
	}
	return connectorv1.Mode_MODE_UNSPECIFIED, fmt.Errorf("%w: %s", ingest.ErrInvalidBatch, m)
}

// deliveryFrom translates a connector's delivery for batch id, with the
// watermark of a sync. A delivery that echoes another batch ID is refused.
func deliveryFrom(d *connectorv1.Delivery, id ingest.BatchID, watermark string) (app.Delivery, error) {
	if d == nil {
		return app.Delivery{}, invalid("delivery is missing", nil)
	}
	if d.GetBatchId() != string(id) {
		return app.Delivery{}, invalid(fmt.Sprintf("delivery.batch_id %q is not %q", d.GetBatchId(), id), nil)
	}
	m, err := mappingFrom(d)
	if err != nil {
		return app.Delivery{}, err
	}
	prices, err := translate("prices", d.GetPrices(), priceFrom)
	if err != nil {
		return app.Delivery{}, err
	}
	rates, err := translate("fx_rates", d.GetFxRates(), fxRateFrom)
	if err != nil {
		return app.Delivery{}, err
	}
	readings, err := translate("index_readings", d.GetIndexReadings(), indexReadingFrom)
	if err != nil {
		return app.Delivery{}, err
	}
	levels, err := translate("index_levels", d.GetIndexLevels(), indexLevelFrom)
	if err != nil {
		return app.Delivery{}, err
	}
	terms, err := translate("bond_terms", d.GetBondTerms(), bondTermsFrom)
	if err != nil {
		return app.Delivery{}, err
	}
	crossChecks, err := translate("cross_check_prices", d.GetCrossCheckPrices(), crossCheckPriceFrom)
	if err != nil {
		return app.Delivery{}, err
	}
	return app.Delivery{
		Institution: ingest.InstitutionID(d.GetInstitution()), Mapping: m,
		Prices: prices, FXRates: rates, IndexReadings: readings, IndexLevels: levels, BondTerms: terms,
		CrossCheckPrices: crossChecks, Watermark: watermark,
	}, nil
}

// crossCheckPriceFrom translates one cross-check price: the same message
// and rules as a price, kept apart as a cross-check price (D-22). Its
// source is set by the use case.
func crossCheckPriceFrom(field string, p *connectorv1.Price) (valuation.CrossCheckPrice, error) {
	price, err := priceFrom(field, p)
	if err != nil {
		return valuation.CrossCheckPrice{}, err
	}
	return valuation.CrossCheckPrice{Instrument: price.Instrument, Date: price.Date, Close: price.Close}, nil
}

// bondTermsFrom translates one treasury bond series' terms: the
// instrument's symbol must be a series code, the rates plain decimals and
// the fee a PLN amount of zero or more.
func bondTermsFrom(field string, b *connectorv1.BondTerms) (ingest.BondTerms, error) {
	in, err := instrumentFrom(field+".instrument", b.GetInstrument())
	if err != nil {
		return ingest.BondTerms{}, err
	}
	series, err := pl.ParseSeries(in.Symbol)
	if err != nil {
		return ingest.BondTerms{}, invalid(field+".instrument.symbol", err)
	}
	first, err := decimalFrom(field+".first_period_rate", b.GetFirstPeriodRate())
	if err != nil {
		return ingest.BondTerms{}, err
	}
	margin, err := decimalFrom(field+".margin", b.GetMargin())
	if err != nil {
		return ingest.BondTerms{}, err
	}
	fee, err := moneyFrom(field+".early_redemption_fee", b.GetEarlyRedemptionFee())
	if err != nil {
		return ingest.BondTerms{}, err
	}
	if !fee.Currency().Equal(pl.Nominal().Currency()) || fee.Sign() < 0 {
		return ingest.BondTerms{}, invalid(field+".early_redemption_fee is not a PLN amount of zero or more", nil)
	}
	return ingest.BondTerms{Instrument: in, Terms: pl.Terms{
		Series: series, FirstPeriodRate: first, Margin: margin, EarlyRedemptionFee: fee,
	}}, nil
}

// bondLotsTo writes treasury bond lots as contract messages, rates as
// fractions and every amount as an exact decimal string.
func bondLotsTo(lots []ingest.BondLot) []*connectorv1.BondLot {
	out := make([]*connectorv1.BondLot, 0, len(lots))
	for i := range lots {
		l := &lots[i]
		fee := l.Terms.EarlyRedemptionFee
		out = append(out, &connectorv1.BondLot{
			Series: l.Series().String(), Quantity: l.Quantity.String(), PurchaseDate: l.Purchased.String(),
			FirstPeriodRate: l.Terms.FirstPeriodRate.String(), Margin: l.Terms.Margin.String(),
			EarlyRedemptionFee: &connectorv1.Money{Amount: fee.Amount().String(), Currency: fee.Currency().String()},
		})
	}
	return out
}

// instrumentsTo writes instruments as contract messages; none is nil.
func instrumentsTo(ins []ledger.Instrument) []*connectorv1.Instrument {
	if len(ins) == 0 {
		return nil
	}
	out := make([]*connectorv1.Instrument, 0, len(ins))
	for _, in := range ins {
		out = append(out, &connectorv1.Instrument{Isin: in.ISIN.String(), Symbol: in.Symbol})
	}
	return out
}

// mappingFrom translates a delivery's transactions, balance assertions and
// ingest problems, with those named by broker account apart (D-72).
func mappingFrom(d *connectorv1.Delivery) (ingest.Mapping, error) {
	txs, err := translate("transactions", d.GetTransactions(), transactionFrom)
	if err != nil {
		return ingest.Mapping{}, err
	}
	assertions, err := translate("balance_assertions", d.GetBalanceAssertions(), assertionFrom)
	if err != nil {
		return ingest.Mapping{}, err
	}
	problems, err := translate("problems", d.GetProblems(), problemFrom)
	if err != nil {
		return ingest.Mapping{}, err
	}
	m := ingest.Mapping{Transactions: txs, Assertions: assertions, Problems: problems}
	if err := splitByBrokerAccount(d, &m); err != nil {
		return ingest.Mapping{}, err
	}
	return m, nil
}

// priceFrom translates one end-of-day price; the close must be positive.
// Its instrument is written "ISIN/symbol" and its source is set by the
// use case.
func priceFrom(field string, p *connectorv1.Price) (valuation.Price, error) {
	in, err := instrumentFrom(field+".instrument", p.GetInstrument())
	if err != nil {
		return valuation.Price{}, err
	}
	date, err := dateFrom(field+".date", p.GetDate())
	if err != nil {
		return valuation.Price{}, err
	}
	closing, err := moneyFrom(field+".close", p.GetClose())
	if err != nil {
		return valuation.Price{}, err
	}
	if closing.Sign() <= 0 {
		return valuation.Price{}, invalid(field+".close is not positive", nil)
	}
	return valuation.Price{Instrument: in.String(), Date: date, Close: closing}, nil
}

// fxRateFrom translates one FX rate to PLN; the rate must be positive.
func fxRateFrom(field string, r *connectorv1.FxRate) (valuation.FXRate, error) {
	c, err := currencyFrom(field+".currency", r.GetCurrency())
	if err != nil {
		return valuation.FXRate{}, err
	}
	date, err := dateFrom(field+".date", r.GetDate())
	if err != nil {
		return valuation.FXRate{}, err
	}
	rate, err := decimalFrom(field+".rate", r.GetRate())
	if err != nil {
		return valuation.FXRate{}, err
	}
	if rate.Sign() <= 0 {
		return valuation.FXRate{}, invalid(field+".rate is not positive", nil)
	}
	return valuation.FXRate{Currency: c, Date: date, Rate: rate}, nil
}

// indexFrom translates an index.
func indexFrom(field string, i connectorv1.Index) (ingest.Index, error) {
	switch i {
	case connectorv1.Index_INDEX_CPI:
		return ingest.IndexCPI, nil
	case connectorv1.Index_INDEX_REFERENCE_RATE:
		return ingest.IndexReferenceRate, nil
	case connectorv1.Index_INDEX_UNSPECIFIED:
	}
	return 0, invalid(fmt.Sprintf("%s.index %s", field, i), nil)
}

// indexReadingFrom translates one index reading; its source is set by the
// use case.
func indexReadingFrom(field string, r *connectorv1.IndexReading) (ingest.IndexReading, error) {
	index, err := indexFrom(field, r.GetIndex())
	if err != nil {
		return ingest.IndexReading{}, err
	}
	date, err := dateFrom(field+".date", r.GetDate())
	if err != nil {
		return ingest.IndexReading{}, err
	}
	value, err := decimalFrom(field+".value", r.GetValue())
	if err != nil {
		return ingest.IndexReading{}, err
	}
	return ingest.IndexReading{Index: index, Date: date, Value: value}, nil
}

// indexLevelFrom translates one market index level, keyed by the index's
// name as the connector sends it; the level must be positive and its source
// is set by the use case.
func indexLevelFrom(field string, l *connectorv1.IndexLevel) (valuation.IndexLevel, error) {
	date, err := dateFrom(field+".date", l.GetDate())
	if err != nil {
		return valuation.IndexLevel{}, err
	}
	level, err := decimalFrom(field+".level", l.GetLevel())
	if err != nil {
		return valuation.IndexLevel{}, err
	}
	out := valuation.IndexLevel{Index: l.GetIndex(), Date: date, Level: level}
	if err := out.Validate(); err != nil {
		return valuation.IndexLevel{}, invalid(field+" is not a valid index level", err)
	}
	return out, nil
}
