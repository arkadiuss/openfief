package ibkrflex

import (
	"strings"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// IBKR's trade sides, and the level of detail of one fill.
const (
	sideBuy        = "BUY"
	sideSell       = "SELL"
	levelExecution = "EXECUTION"
)

// bookedTradeType reports whether trades of IBKR transaction type tt move
// units and cash as they stand. Cancellations and corrections are not
// supported yet.
func bookedTradeType(tt string) bool {
	switch tt {
	case "", "ExchTrade", "BookTrade", "FracShare":
		return true
	}
	return false
}

// trades maps the Trades section: each EXECUTION row of a stock or ETF
// becomes a buy or a sell, and each EXECUTION row of a currency pair an FX
// conversion, keyed by IBKR's tradeID. In opening mode every row is skipped:
// the opening lots and opening cash already hold its effect (D-23).
func (s *stmt) trades(rows []row) {
	for i := range rows {
		r := &rows[i]
		ref := s.ref("Trades", r, "tradeID")
		t, err := s.trade(r)
		s.add(ref, &t, err)
	}
}

// trade translates one Trades row.
func (s *stmt) trade(r *row) (ledger.Transaction, error) {
	if err := s.needActivityRow(r, "Trade", levelExecution); err != nil {
		return ledger.Transaction{}, err
	}
	if tt := r.text("transactionType"); !bookedTradeType(tt) {
		return ledger.Transaction{}, unsupported("trade type %q is not supported", tt)
	}
	id := r.text("tradeID")
	if id == "" {
		return ledger.Transaction{}, unsupported("trade without a tradeID")
	}
	on, err := r.date("tradeDate", "dateTime")
	if err != nil {
		return ledger.Transaction{}, err
	}
	fee, err := tradeFee(r)
	if err != nil {
		return ledger.Transaction{}, err
	}
	h := s.header(id, on)
	switch c := r.text("assetCategory"); c {
	case categoryStock:
		return stockTrade(h, r, fee)
	case categoryCash:
		return fxTrade(h, r, fee)
	default:
		return ledger.Transaction{}, unsupported("asset category %q is not supported", c)
	}
}

// needActivityRow fails unless r is an elem row at level (or with no level),
// in transactions mode.
func (s *stmt) needActivityRow(r *row, elem, level string) error {
	switch {
	case r.elem != elem:
		return skipped("<%s> is not a <%s> row", r.elem, elem)
	case r.text("levelOfDetail") != "" && r.text("levelOfDetail") != level:
		return skipped("level of detail %q is not read", r.text("levelOfDetail"))
	case s.mode == ingest.ModeOpening:
		return skipped("opening mode: the opening lots and opening cash hold its effect")
	}
	return nil
}

// tradeSide reports whether r buys (true) or sells (false): a BUY with a
// positive quantity or a SELL with a negative one.
func tradeSide(r *row, qty kernel.Decimal) (bool, error) {
	switch side := r.text("buySell"); {
	case side == sideBuy && qty.Sign() > 0:
		return true, nil
	case side == sideSell && qty.Sign() < 0:
		return false, nil
	default:
		return false, unsupported("trade side %q with a quantity of that sign is not supported", side)
	}
}

// stockTrade returns the buy or sell of a stock or ETF trade. Proceeds are
// negative for a buy and positive for a sell, as IBKR reports them.
func stockTrade(h ledger.TransactionHeader, r *row, fee kernel.Money) (ledger.Transaction, error) {
	if err := needStock(r); err != nil {
		return ledger.Transaction{}, err
	}
	in, err := r.instrument()
	if err != nil {
		return ledger.Transaction{}, err
	}
	qty, proceeds, buy, err := tradeAmounts(r)
	if err != nil {
		return ledger.Transaction{}, err
	}
	var t ledger.Transaction
	if buy {
		t, err = ledger.NewBuy(h, in, kernel.NewQuantity(qty), proceeds.Neg(), fee)
		return booked(&t, err, "buy "+in.Symbol)
	}
	t, err = ledger.NewSell(h, in, kernel.NewQuantity(qty.Neg()), proceeds, fee)
	return booked(&t, err, "sell "+in.Symbol)
}

// tradeAmounts reads a trade's signed quantity and proceeds, and its side.
func tradeAmounts(r *row) (kernel.Decimal, kernel.Money, bool, error) {
	qty, err := r.decimal("quantity")
	if err != nil {
		return kernel.Decimal{}, kernel.Money{}, false, err
	}
	proceeds, err := r.money("proceeds")
	if err != nil {
		return kernel.Decimal{}, kernel.Money{}, false, err
	}
	buy, err := tradeSide(r, qty)
	return qty, proceeds, buy, err
}

// fxTrade returns the FX conversion of a currency pair trade, such as
// EUR.USD: quantity is in the base currency (EUR), proceeds in the quote
// currency (USD), which is the row's currency. Buying the pair buys the base
// currency and sells the quote currency.
func fxTrade(h ledger.TransactionHeader, r *row, fee kernel.Money) (ledger.Transaction, error) {
	pair := r.text("symbol")
	base, quote, ok := strings.Cut(pair, ".")
	baseCur, err := kernel.ParseCurrency(base)
	if !ok || err != nil || quote != r.text("currency") {
		return ledger.Transaction{}, unsupported("currency pair %q is not BASE.QUOTE in the row's currency", pair)
	}
	qty, proceeds, buy, err := tradeAmounts(r)
	if err != nil {
		return ledger.Transaction{}, err
	}
	baseAmount := kernel.NewMoney(qty, baseCur)
	sold, bought := proceeds.Neg(), baseAmount
	if !buy {
		sold, bought = baseAmount.Neg(), proceeds
	}
	t, err := ledger.NewFXConversion(h, sold, bought, fee)
	return booked(&t, err, "fx "+pair)
}

// tradeFee returns what a trade cost in commission and transaction taxes,
// as a positive amount, or a zero Money for none. IBKR reports both as
// negative numbers; the commission may be in another currency than the
// trade, and both must then be in the same one.
func tradeFee(r *row) (kernel.Money, error) {
	commission, err := charge(r, "ibCommission", "ibCommissionCurrency")
	if err != nil {
		return kernel.Money{}, err
	}
	taxes, err := charge(r, "taxes", "currency")
	switch {
	case err != nil:
		return kernel.Money{}, err
	case taxes.IsZero():
		return commission, nil
	case commission.IsZero():
		return taxes, nil
	}
	fee, err := commission.Add(taxes)
	if err != nil {
		return kernel.Money{}, unsupported("commission and taxes in different currencies are not supported")
	}
	return fee, nil
}

// charge reads a charge that IBKR reports as a negative amount and returns
// it as a positive one, or a zero Money when the row has none.
func charge(r *row, amount, cur string) (kernel.Money, error) {
	if r.text(amount) == "" {
		return kernel.Money{}, nil
	}
	d, err := r.decimal(amount)
	switch {
	case err != nil:
		return kernel.Money{}, err
	case d.IsZero():
		return kernel.Money{}, nil
	case d.Sign() > 0:
		return kernel.Money{}, unsupported("%s is a credit, which is not supported", amount)
	}
	c, err := r.currency(cur)
	if err != nil {
		return kernel.Money{}, err
	}
	return kernel.NewMoney(d.Neg(), c), nil
}
