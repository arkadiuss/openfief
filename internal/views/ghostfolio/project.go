package ghostfolio

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views"
)

// Source is what the view projects: the record's accounts as the read API
// gives them, plus what the instance decides.
type Source struct {
	// Instance is the view instance's name; owned accounts' UUIDs and the
	// MANUAL assets' symbols derive from it (AccountUUID, BondSymbol,
	// FeeSymbol).
	Instance string
	// User is the Ghostfolio user's ID; the MANUAL assets' symbols derive
	// from it too.
	User string
	// Currency is the currency of the accounts the view creates: the
	// Ghostfolio user's base currency.
	Currency kernel.Currency
	// Accounts are the record's accounts.
	Accounts []SourceAccount
	// Listings say under which Ghostfolio data source and symbol each
	// listed instrument is known, by ISIN; the owner configures them.
	Listings map[ledger.ISIN]Listing
}

// SourceAccount is one account of the record: its transactions and the
// stored valuations of its bonds.
type SourceAccount struct {
	ID           ledger.AccountID
	Name         string
	Transactions []ledger.Transaction
	Valuations   []Valuation
}

// Valuation is a holding's stored value on a date, from the read API.
type Valuation struct {
	Date       kernel.Date
	Instrument ledger.Instrument
	Value      kernel.Money
}

// Listing is how Ghostfolio knows a listed instrument: a data source such
// as YAHOO and that source's symbol.
type Listing struct {
	DataSource string
	Symbol     string
}

// Desired is what the owned accounts should hold after a sync.
type Desired struct {
	// Accounts are the owned accounts, one per account of the record, in
	// ID order.
	Accounts []Account
	// Activities are in account, then comment order.
	Activities []Activity
	// Assets are the MANUAL assets with their full price history, in
	// symbol order.
	Assets []AssetProfile
	// Balances are the owned accounts' cash balances, in account, then
	// date order.
	Balances []Balance
}

// ProblemReason says why a transaction has no activity.
type ProblemReason int

// The reasons a transaction is left out.
const (
	// ProblemUnlisted is a listed instrument the config gives no listing.
	ProblemUnlisted ProblemReason = iota + 1
	// ProblemTooLarge is a number with more than 15 integer digits.
	ProblemTooLarge
	// ProblemCommentTooLong is a transaction key longer than a Ghostfolio
	// comment may be.
	ProblemCommentTooLong
	// ProblemMalformed is a transaction whose postings don't have the
	// shape its kind has.
	ProblemMalformed
	// ProblemCashCurrency is an account's cash in a currency other than
	// its owned account's: a Ghostfolio account has one balance, in its
	// own currency.
	ProblemCashCurrency
)

// String names the reason.
func (r ProblemReason) String() string {
	switch r {
	case ProblemUnlisted:
		return "unlisted instrument"
	case ProblemTooLarge:
		return "number too large"
	case ProblemCommentTooLong:
		return "transaction key too long"
	case ProblemMalformed:
		return "malformed transaction"
	case ProblemCashCurrency:
		return "cash in another currency than the owned account's"
	}
	return fmt.Sprintf("ProblemReason(%d)", int(r))
}

// Problem is something the view leaves out of Ghostfolio, and why: a
// transaction, named by its key, or an account's cash balance, named by the
// account and, for cash in another currency, that currency.
type Problem struct {
	Transaction ledger.TransactionKey
	Account     ledger.AccountID
	Currency    kernel.Currency
	Reason      ProblemReason
}

// Ghostfolio's asset classes for a bond, and the most characters a comment
// may have.
const (
	assetClassFixedIncome = "FIXED_INCOME"
	assetSubClassBond     = "BOND"
	maxCommentLen         = 2000
	// quoPlaces is how many decimal places a unit price is computed to
	// before toWire rounds it to 15 significant digits.
	quoPlaces = 16
)

// errUnlisted and errMalformed are the internal causes of a problem.
var (
	errUnlisted  = errors.New("unlisted instrument")
	errMalformed = errors.New("malformed transaction")
)

// Project returns what the owned accounts should hold for src (D-4, D-64,
// D-108), and what it leaves out:
//
//   - each account of the record gets an owned account (AccountUUID), in
//     the Ghostfolio user's base currency, whose cash balance follows the
//     account's cash in that currency: one balance at the end of each date
//     it changes. Cash in another currency is left out as a problem;
//   - a buy, a sell and an opening lot become a BUY or SELL at the unit price
//     the holding's posting implies, with the fees in its currency, and an
//     opening lot is dated when it was first acquired;
//   - a dividend becomes a DIVIDEND, and interest, such as a bond's coupon
//     (D-101), an INTEREST, of quantity 1 at the gross amount, with the tax
//     withheld in its currency as its fee;
//   - a fee becomes a FEE against the MANUAL fee asset of its currency, and
//     fees or tax in another currency than their activity's become FEEs of
//     their own;
//   - FX conversions, deposits and withdrawals have no activity;
//   - a bond is a MANUAL asset (BondSymbol) named by its series code, whose
//     price history is its stored valuations divided by the quantity held;
//   - a listed instrument takes its data source and symbol from Listings.
//
// Each activity's comment is CommentPrefix and the transaction's key, plus
// "#fee:" and a currency for a fee split off its activity. A transaction
// left out still moves the cash balance.
func Project(src *Source) (Desired, []Problem) {
	p := projector{src: src, bonds: map[string]*AssetProfile{}, fees: map[string]kernel.Currency{}}
	accounts := slices.SortedFunc(slices.Values(src.Accounts), func(a, b SourceAccount) int { return cmp.Compare(a.ID, b.ID) })
	var d Desired
	for i := range accounts {
		a := &accounts[i]
		id := AccountUUID(src.Instance, a.ID)
		d.Accounts = append(d.Accounts, Account{ID: id, Name: a.Name, Currency: src.Currency})
		d.Balances = append(d.Balances, p.balances(id, a)...)
		for j := range a.Transactions {
			acts, problem := p.transaction(id, &a.Transactions[j])
			if problem != nil {
				p.problems = append(p.problems, *problem)
				continue
			}
			d.Activities = append(d.Activities, acts...)
		}
	}
	for i := range accounts {
		p.prices(&accounts[i])
	}
	slices.SortFunc(d.Activities, func(a, b Activity) int {
		return cmp.Or(cmp.Compare(a.AccountID, b.AccountID), cmp.Compare(a.Comment, b.Comment))
	})
	d.Assets = p.assets()
	return d, p.problems
}

// projector holds what Project collects across accounts: the bond assets
// with their prices, the fee assets' currencies by symbol, and the problems.
type projector struct {
	src      *Source
	bonds    map[string]*AssetProfile
	fees     map[string]kernel.Currency
	problems []Problem
}

// transaction returns t's activities in the owned account accountID, or the
// problem that leaves t out.
func (p *projector) transaction(accountID string, t *ledger.Transaction) ([]Activity, *Problem) {
	comment := strings.TrimSpace(CommentPrefix + t.Key.String())
	if utf8.RuneCountInString(comment)+len("#fee:XXX") > maxCommentLen {
		return nil, &Problem{Transaction: t.Key, Reason: ProblemCommentTooLong}
	}
	acts, err := p.activities(accountID, comment, t)
	switch {
	case errors.Is(err, errUnlisted):
		return nil, &Problem{Transaction: t.Key, Reason: ProblemUnlisted}
	case errors.Is(err, ErrNumberTooLarge):
		return nil, &Problem{Transaction: t.Key, Reason: ProblemTooLarge}
	case err != nil:
		return nil, &Problem{Transaction: t.Key, Reason: ProblemMalformed}
	}
	return acts, nil
}

// activities builds t's activities, as Project describes.
func (p *projector) activities(accountID, comment string, t *ledger.Transaction) ([]Activity, error) {
	switch t.Kind {
	case ledger.KindBuy, ledger.KindSell, ledger.KindOpeningLot:
		return p.trade(accountID, comment, t)
	case ledger.KindDividend:
		return p.income(accountID, comment, TypeDividend, t)
	case ledger.KindInterest:
		return p.income(accountID, comment, TypeInterest, t)
	case ledger.KindFee:
		return p.feesOnly(accountID, comment, t)
	case ledger.KindFXConversion, ledger.KindDeposit, ledger.KindWithdrawal, ledger.KindDeletion:
		return nil, nil
	}
	return nil, fmt.Errorf("%w: kind %s", errMalformed, t.Kind)
}

// trade is a buy, a sell or an opening lot: one BUY or SELL, plus FEEs in
// other currencies.
func (p *projector) trade(accountID, comment string, t *ledger.Transaction) ([]Activity, error) {
	h, err := holdingPosting(t)
	if err != nil {
		return nil, err
	}
	typ, date := TypeBuy, t.Date
	switch {
	case t.Kind == ledger.KindOpeningLot:
		date = t.LotOpened
	case h.Quantity.Sign() < 0:
		typ = TypeSell
	}
	if h.Quantity.IsZero() {
		return nil, fmt.Errorf("%w: no quantity", errMalformed)
	}
	qty := abs(h.Quantity.Decimal())
	unit, err := abs(h.Amount.Amount()).Quo(qty, quoPlaces, kernel.RoundHalfEven)
	if err != nil {
		return nil, fmt.Errorf("unit price: %w", err)
	}
	a := Activity{AccountID: accountID, Comment: comment, Type: typ, Date: date, Currency: h.Amount.Currency(), Quantity: qty, UnitPrice: unit}
	return p.withFees(&a, h, t, ledger.BookFee)
}

// income is a dividend or interest: one activity of typ, DIVIDEND or
// INTEREST, of quantity 1 at the gross amount, its tax withheld as the fee,
// plus FEEs in other currencies.
func (p *projector) income(accountID, comment string, typ ActivityType, t *ledger.Transaction) ([]Activity, error) {
	h, err := holdingPosting(t)
	if err != nil {
		return nil, err
	}
	a := Activity{
		AccountID: accountID, Comment: comment, Type: typ, Date: t.Date,
		Currency: h.Amount.Currency(), Quantity: kernel.NewDecimal(1, 0), UnitPrice: abs(h.Amount.Amount()),
	}
	return p.withFees(&a, h, t, ledger.BookWithholdingTax)
}

// feesOnly is a fee on its own: one FEE per currency, the first currency
// in code order under the transaction's own comment.
func (p *projector) feesOnly(accountID, comment string, t *ledger.Transaction) ([]Activity, error) {
	byCurrency := sums(t, ledger.BookFee)
	if len(byCurrency) == 0 {
		return nil, fmt.Errorf("%w: a fee without a fee posting", errMalformed)
	}
	var out []Activity
	for i, m := range byCurrency {
		c := comment
		if i > 0 {
			c = feeComment(comment, m.Currency())
		}
		a, err := p.fee(accountID, c, t.Date, m)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// withFees completes a, the activity of holding posting h: its symbol, its
// fee from t's postings to kind in its currency, and a FEE for each other
// currency. Every number goes through toWire.
func (p *projector) withFees(a *Activity, h *ledger.Posting, t *ledger.Transaction, kind ledger.BookKind) ([]Activity, error) {
	if err := p.symbol(a, h); err != nil {
		return nil, err
	}
	var extra []Activity
	for _, m := range sums(t, kind) {
		if m.Currency().Equal(a.Currency) {
			a.Fee = m.Amount()
			continue
		}
		f, err := p.fee(a.AccountID, feeComment(a.Comment, m.Currency()), a.Date, m)
		if err != nil {
			return nil, err
		}
		extra = append(extra, f)
	}
	if err := wireNumbers(a); err != nil {
		return nil, err
	}
	return append([]Activity{*a}, extra...), nil
}

// fee is a FEE of m against the fee asset of its currency.
func (p *projector) fee(accountID, comment string, date kernel.Date, m kernel.Money) (Activity, error) {
	sym := FeeSymbol(p.src.Instance, p.src.User, m.Currency())
	p.fees[sym] = m.Currency()
	a := Activity{
		AccountID: accountID, Comment: comment, Type: TypeFee, Date: date, Currency: m.Currency(),
		DataSource: DataSourceManual, Symbol: sym, Fee: m.Amount(),
	}
	if err := wireNumbers(&a); err != nil {
		return Activity{}, err
	}
	return a, nil
}

// symbol sets a's data source and symbol from h's instrument: a bond's
// MANUAL asset, which it records with h's currency, or the configured
// listing.
func (p *projector) symbol(a *Activity, h *ledger.Posting) error {
	in := h.Book.Instrument
	if views.AssetClassOf(in) == views.AssetTreasuryBond {
		sym := BondSymbol(p.src.Instance, p.src.User, in.Symbol)
		if _, ok := p.bonds[sym]; !ok {
			p.bonds[sym] = &AssetProfile{
				Symbol: sym, Name: in.Symbol, ISIN: in.ISIN.String(), Currency: h.Amount.Currency(),
				AssetClass: assetClassFixedIncome, AssetSubClass: assetSubClassBond,
			}
		}
		a.DataSource, a.Symbol = DataSourceManual, sym
		return nil
	}
	l, ok := p.src.Listings[in.ISIN]
	if !ok {
		return fmt.Errorf("%w: %s", errUnlisted, in.ISIN)
	}
	a.DataSource, a.Symbol = l.DataSource, l.Symbol
	return nil
}

// prices adds a's bond prices to the bond assets: each valuation divided
// by the quantity held at the end of its date. A date another account
// already priced keeps that price.
func (p *projector) prices(a *SourceAccount) {
	vals := slices.SortedFunc(slices.Values(a.Valuations), func(x, y Valuation) int { return x.Date.Compare(y.Date) })
	for _, v := range vals {
		asset, ok := p.bonds[BondSymbol(p.src.Instance, p.src.User, v.Instrument.Symbol)]
		if !ok || views.AssetClassOf(v.Instrument) != views.AssetTreasuryBond {
			continue
		}
		qty := heldOn(a.Transactions, v.Instrument, v.Date)
		if qty.Sign() <= 0 || slices.ContainsFunc(asset.MarketData, func(pr Price) bool { return pr.Date == v.Date }) {
			continue
		}
		price, err := v.Value.Amount().Quo(qty, quoPlaces, kernel.RoundHalfEven)
		if err != nil {
			continue
		}
		if price, err = toWire(price); err != nil {
			continue
		}
		asset.MarketData = append(asset.MarketData, Price{Date: v.Date, Price: price})
	}
}

// assets returns the bond assets, with their prices in date order, and the
// fee assets, in symbol order.
func (p *projector) assets() []AssetProfile {
	out := make([]AssetProfile, 0, len(p.bonds)+len(p.fees))
	for _, a := range p.bonds {
		slices.SortFunc(a.MarketData, func(x, y Price) int { return x.Date.Compare(y.Date) })
		out = append(out, *a)
	}
	for sym, c := range p.fees {
		out = append(out, AssetProfile{Symbol: sym, Name: "OpenFief fees (" + c.String() + ")", Currency: c})
	}
	slices.SortFunc(out, func(a, b AssetProfile) int { return cmp.Compare(a.Symbol, b.Symbol) })
	return out
}

// balances returns the cash balances of the owned account id for a (D-108):
// a's cash in the source's currency at the end of each date on which it
// changes, from zero before a's first transaction. Cash in other
// currencies is left out with one problem per currency, and a balance too
// large for Ghostfolio leaves out all of a's balances with one problem.
func (p *projector) balances(id string, a *SourceAccount) []Balance {
	moves := cashChanges(a.Transactions, p.src.Currency)
	byDate := moves.byDate
	for _, code := range slices.Sorted(maps.Keys(moves.others)) {
		p.problems = append(p.problems, Problem{Account: a.ID, Currency: moves.others[code], Reason: ProblemCashCurrency})
	}
	var out []Balance
	var running kernel.Decimal
	for _, d := range slices.SortedFunc(maps.Keys(byDate), kernel.Date.Compare) {
		next := running.Add(byDate[d])
		if next.Equal(running) {
			continue
		}
		running = next
		v, err := toWire(running)
		if err != nil {
			p.problems = append(p.problems, Problem{Account: a.ID, Reason: ProblemTooLarge})
			return nil
		}
		out = append(out, Balance{AccountID: id, Date: d, Value: v})
	}
	return out
}

// cashMoves is how transactions move an account's cash: in one currency
// by date, and the other currencies they move, by code.
type cashMoves struct {
	byDate map[kernel.Date]kernel.Decimal
	others map[string]kernel.Currency
}

// cashChanges returns how ts move the cash in currency c and which other
// currencies their cash postings are in.
func cashChanges(ts []ledger.Transaction, c kernel.Currency) cashMoves {
	m := cashMoves{byDate: map[kernel.Date]kernel.Decimal{}, others: map[string]kernel.Currency{}}
	for i := range ts {
		t := &ts[i]
		for _, po := range t.Postings {
			switch {
			case po.Book.Kind != ledger.BookCash || po.Amount.IsZero():
			case po.Amount.Currency().Equal(c):
				m.byDate[t.Date] = m.byDate[t.Date].Add(po.Amount.Amount())
			default:
				m.others[po.Amount.Currency().String()] = po.Amount.Currency()
			}
		}
	}
	return m
}

// heldOn returns the quantity of in that ts hold at the end of date.
func heldOn(ts []ledger.Transaction, in ledger.Instrument, date kernel.Date) kernel.Decimal {
	var q kernel.Decimal
	for i := range ts {
		if ts[i].Date.After(date) {
			continue
		}
		for _, po := range ts[i].Postings {
			if po.Book.Kind == ledger.BookHolding && po.Book.Instrument == in {
				q = q.Add(po.Quantity.Decimal())
			}
		}
	}
	return q
}

// holdingPosting returns t's only posting to a holding.
func holdingPosting(t *ledger.Transaction) (*ledger.Posting, error) {
	var h *ledger.Posting
	for i := range t.Postings {
		if t.Postings[i].Book.Kind != ledger.BookHolding {
			continue
		}
		if h != nil {
			return nil, fmt.Errorf("%w: more than one holding posting", errMalformed)
		}
		h = &t.Postings[i]
	}
	if h == nil {
		return nil, fmt.Errorf("%w: no holding posting", errMalformed)
	}
	return h, nil
}

// sums returns the sum of t's postings to ledger accounts of kind, one per
// currency, in currency code order, leaving out zero sums.
func sums(t *ledger.Transaction, kind ledger.BookKind) []kernel.Money {
	by := map[string]kernel.Money{}
	for _, po := range t.Postings {
		if po.Book.Kind != kind {
			continue
		}
		code := po.Amount.Currency().String()
		prev, ok := by[code]
		if !ok {
			by[code] = po.Amount
			continue
		}
		by[code] = kernel.NewMoney(prev.Amount().Add(po.Amount.Amount()), po.Amount.Currency())
	}
	out := make([]kernel.Money, 0, len(by))
	for _, code := range slices.Sorted(maps.Keys(by)) {
		if m := by[code]; !m.IsZero() {
			out = append(out, kernel.NewMoney(abs(m.Amount()), m.Currency()))
		}
	}
	return out
}

// feeComment is the comment of a fee split off the activity commented c.
func feeComment(c string, cur kernel.Currency) string {
	return c + "#fee:" + cur.String()
}

// wireNumbers rounds a's numbers for Ghostfolio (toWire).
func wireNumbers(a *Activity) error {
	for _, d := range []*kernel.Decimal{&a.Quantity, &a.UnitPrice, &a.Fee} {
		w, err := toWire(*d)
		if err != nil {
			return err
		}
		*d = w
	}
	return nil
}

// abs returns |d|.
func abs(d kernel.Decimal) kernel.Decimal {
	if d.Sign() < 0 {
		return d.Neg()
	}
	return d
}
