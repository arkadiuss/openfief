package valuation

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownMissingReason reports a MissingReason outside the declared
// constants, or text that names none of them.
var ErrUnknownMissingReason = errors.New("unknown missing-input reason")

// ErrBondLot reports a bond lot that cannot be valued, such as one whose
// cost is not in PLN.
var ErrBondLot = errors.New("bond lot cannot be valued")

// StatutorySource is the source of a bond's valuation: the issue letter's
// formulas in internal/jurisdiction/pl, not a market price.
const StatutorySource = "statutory"

// ValueRounding rounds a holding's value in PLN to grosze. Nothing in
// docs/DESIGN.md fixes it; half up is what the bond tables use (D-37).
const ValueRounding = kernel.RoundHalfUp

// purchasePricePlaces is how many decimal places a bond lot's price per
// bond (its cost over its quantity) is taken to: grosze.
const purchasePricePlaces = 2

// MissingReason says which input a holding's valuation lacked. The zero
// value means none: the holding is valued.
type MissingReason int

const (
	// MissingPrice is no price for the instrument on or before the date.
	MissingPrice MissingReason = iota + 1
	// MissingFXRate is no FX rate to PLN for the price's currency on or
	// before the date.
	MissingFXRate
	// MissingIndexReading is no period index reading for an interest period
	// of a bond lot.
	MissingIndexReading
	// MissingMarginalRate is the owner's marginal income tax rate, which an
	// IKZE's liquidation value needs (D-78), not set.
	MissingMarginalRate
	// MissingLots is no open lots for a holding, such as a short one: its
	// cost basis, which its net and liquidation values need, is unknown.
	MissingLots
)

// String returns the reason's name: price, fx_rate, index_reading,
// marginal_rate or lots.
func (r MissingReason) String() string {
	switch r {
	case MissingPrice:
		return "price"
	case MissingFXRate:
		return "fx_rate"
	case MissingIndexReading:
		return "index_reading"
	case MissingMarginalRate:
		return "marginal_rate"
	case MissingLots:
		return "lots"
	}
	return "MissingReason(" + strconv.Itoa(int(r)) + ")"
}

// ParseMissingReason parses a reason name as String writes it.
func ParseMissingReason(s string) (MissingReason, error) {
	for _, r := range []MissingReason{MissingPrice, MissingFXRate, MissingIndexReading, MissingMarginalRate, MissingLots} {
		if r.String() == s {
			return r, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownMissingReason, s)
}

// HoldingValuation is the stored valuation of one holding on one day
// (D-36): its gross value in the instrument's own currency and in PLN, on a
// basis, with the inputs' source and dates, and its net and liquidation
// values in PLN (D-77). When an input is missing, Missing says which and
// Detail why, and no value is set: the row records the gap instead of
// failing the revaluation. A row with its gross value can still lack its
// net or liquidation value, which NetMissing and LiquidationMissing name.
type HoldingValuation struct {
	Account    ledger.AccountID
	Instrument ledger.Instrument
	Date       kernel.Date
	Basis      Basis
	Quantity   kernel.Quantity
	Value      kernel.Money
	PLN        kernel.Money
	Source     string
	// PriceDate and RateDate are the dates of the price and FX rate used,
	// on or before Date; zero when none was used.
	PriceDate kernel.Date
	RateDate  kernel.Date
	Missing   MissingReason
	Detail    string
	// Net and Liquidation are the holding's net and liquidation values in
	// PLN, set when the gross value is and nothing they need is missing.
	Net                kernel.Money
	Liquidation        kernel.Money
	NetMissing         MissingReason
	LiquidationMissing MissingReason
}

// Valued reports whether no input was missing: the gross value is set.
func (v *HoldingValuation) Valued() bool { return v.Missing == 0 }

// NetValued reports whether the net value is set.
func (v *HoldingValuation) NetValued() bool { return v.Valued() && v.NetMissing == 0 }

// LiquidationValued reports whether the liquidation value is set.
func (v *HoldingValuation) LiquidationValued() bool { return v.Valued() && v.LiquidationMissing == 0 }

// PriceOnOrBefore returns the latest price of instrument dated on or before
// cutoff, the first of them if several share that date, or an error wrapping
// ErrNoPrice.
func PriceOnOrBefore(prices []Price, instrument string, cutoff kernel.Date) (Price, error) {
	var best Price
	found := false
	for _, p := range prices {
		if p.Instrument != instrument || p.Date.After(cutoff) {
			continue
		}
		if !found || p.Date.After(best.Date) {
			best, found = p, true
		}
	}
	if !found {
		return Price{}, fmt.Errorf("%w: %s on or before %s", ErrNoPrice, instrument, cutoff)
	}
	return best, nil
}

// MarketHolding is what valuing a holding priced at market needs besides
// the holding: the prices of its instrument, the FX rates of every currency
// its price, its lots' costs and its account's fee rule are in, its open
// lots, and the account's terms.
type MarketHolding struct {
	Prices []Price
	Rates  []FXRate
	// Lots are the holding's open lots on its date. LotsUnknown reports
	// that they could not be derived, such as for a short holding: its net
	// and liquidation values then lack them.
	Lots        []ledger.Lot
	LotsUnknown bool
	Terms       *AccountTerms
}

// ValueMarketHolding values h on h.Date at the latest price of its
// instrument on or before that date (D-21), converted to PLN at the latest
// FX rate on or before it (D-20) and rounded with ValueRounding: its gross
// value. Its net and liquidation values (D-77) follow from the account's
// terms (AccountTerms.listed). A missing price or rate gives a row with
// Missing set, not an error.
func ValueMarketHolding(h *ledger.Holding, m *MarketHolding) (HoldingValuation, error) {
	row := HoldingValuation{
		Account: h.Account, Instrument: h.Instrument, Date: h.Date, Basis: Market, Quantity: h.Quantity,
	}
	p, err := PriceOnOrBefore(m.Prices, h.Instrument.String(), h.Date)
	if errors.Is(err, ErrNoPrice) {
		return missing(&row, MissingPrice, err), nil
	}
	v, err := MarketValuation(h.Instrument.String(), h.Quantity, &p)
	if err != nil {
		return HoldingValuation{}, err
	}
	row.Source, row.PriceDate = p.Source, p.Date
	if v.Value.Currency().Equal(PLN()) {
		row.Value = v.Value
		if row.PLN, err = roundPLN(v.Value.Amount(), ValueRounding); err != nil {
			return HoldingValuation{}, err
		}
		err = m.Terms.listed(&row, m)
		return row, err
	}
	rate, err := RateOnOrBefore(m.Rates, v.Value.Currency(), h.Date)
	if errors.Is(err, ErrNoRate) {
		return missing(&row, MissingFXRate, err), nil
	}
	if err != nil {
		return HoldingValuation{}, err
	}
	pln, err := ToPLN(v.Value, rate, ValueRounding)
	if err != nil {
		return HoldingValuation{}, err
	}
	row.Value, row.PLN, row.RateDate = v.Value, pln, rate.Date
	err = m.Terms.listed(&row, m)
	return row, err
}

// BondHolding is what valuing a holding of retail treasury bonds needs
// besides the holding: its open lots, its series terms, the period index
// readings and the account's terms.
type BondHolding struct {
	Lots     []ledger.Lot
	Terms    *pl.Terms
	Readings []IndexReading
	Account  *AccountTerms
}

// ValueBondHolding values h on h.Date as the sum of its lots' statutory
// values (pl.ValueLot) in the account's tax wrapper: its gross accrued
// value, and the net and liquidation values the pack computes from each
// lot's interest, purchase discount and early-redemption fee (D-77). Each
// lot's purchase date is its open date and its price per bond is its cost
// over its quantity, which must be in PLN. A missing period index reading
// gives a row with Missing set, not an error.
func ValueBondHolding(h *ledger.Holding, b *BondHolding) (HoldingValuation, error) {
	wrapper, err := pl.ParseTaxWrapper(string(b.Account.Wrapper))
	if err != nil {
		return HoldingValuation{}, fmt.Errorf("valuing %s: %w", h.Instrument, err)
	}
	row := HoldingValuation{
		Account: h.Account, Instrument: h.Instrument, Date: h.Date, Basis: GrossAccrued, Quantity: h.Quantity,
		Source: StatutorySource,
	}
	var sum jurisdiction.Values
	for i := range b.Lots {
		v, err := valueBondLot(h, &b.Lots[i], b, wrapper)
		if missingIndex, ok := errors.AsType[*pl.MissingIndexError](err); ok {
			return missing(&row, MissingIndexReading, missingIndex), nil
		}
		if err != nil {
			return HoldingValuation{}, err
		}
		if sum, err = addValues(&sum, &v); err != nil {
			return HoldingValuation{}, fmt.Errorf("valuing %s: %w", h.Instrument, err)
		}
	}
	zero := kernel.NewMoney(kernel.Decimal{}, PLN())
	row.Value, row.PLN, row.Net, row.Liquidation = zero, zero, zero, zero
	if len(b.Lots) > 0 {
		row.Value, row.PLN, row.Net, row.Liquidation = sum.Gross, sum.Gross, sum.Net, sum.Liquidation
	}
	if sum.LiquidationUnknown {
		row.Liquidation, row.LiquidationMissing = kernel.Money{}, MissingMarginalRate
	}
	return row, nil
}

// valueBondLot values one lot of h with pl.ValueLot in wrapper, and its
// three values with the account's pack.
func valueBondLot(h *ledger.Holding, lot *ledger.Lot, b *BondHolding, wrapper pl.TaxWrapper) (jurisdiction.Values, error) {
	if !lot.Cost.Currency().Equal(PLN()) || lot.Quantity.Sign() <= 0 {
		return jurisdiction.Values{}, fmt.Errorf("%w: %s lot %s: cost %s for %s bonds", ErrBondLot, h.Instrument, lot.Origin, lot.Cost, lot.Quantity)
	}
	price, err := lot.Cost.Amount().Quo(lot.Quantity.Decimal(), purchasePricePlaces, kernel.RoundHalfEven)
	if err != nil {
		return jurisdiction.Values{}, fmt.Errorf("%w: %s lot %s: %w", ErrBondLot, h.Instrument, lot.Origin, err)
	}
	idx, err := LotIndices(b.Terms.Series, lot.Opened, h.Date, b.Readings)
	if err != nil {
		return jurisdiction.Values{}, err
	}
	purchase := pl.Purchase{Date: lot.Opened, Price: kernel.NewMoney(price, PLN())}
	v, err := pl.ValueLot(b.Terms, idx, lot.Quantity, purchase, wrapper, h.Date)
	if err != nil {
		return jurisdiction.Values{}, fmt.Errorf("valuing %s lot %s: %w", h.Instrument, lot.Origin, err)
	}
	income, err := v.Interest.Add(v.Discount)
	if err != nil {
		return jurisdiction.Values{}, fmt.Errorf("valuing %s lot %s: %w", h.Instrument, lot.Origin, err)
	}
	a := b.Account
	values, err := a.Pack.ValueAccrued(&jurisdiction.Accrued{Wrapper: a.Wrapper, Gross: v.Gross, Income: income, EarlyFee: v.Fee}, &a.Settings)
	if err != nil {
		return jurisdiction.Values{}, fmt.Errorf("valuing %s lot %s: %w", h.Instrument, lot.Origin, err)
	}
	return values, nil
}

// missing returns row marked as lacking an input for reason, with err's text
// as the detail. The text names instruments, currencies, series and dates
// only, never amounts.
func missing(row *HoldingValuation, reason MissingReason, err error) HoldingValuation {
	row.Missing, row.Detail = reason, err.Error()
	return *row
}

// ValuedThrough returns the date an account is valued through after a
// revaluation produced rows for the days from its start date to today: the day before the first
// row with a missing input, or today when there is none. An earlier stored
// date (prior, when hasPrior) stays if it is before the day before the start
// date: the gap it stopped at came earlier, so this revaluation did not
// touch it.
func ValuedThrough(prior kernel.Date, hasPrior bool, from, today kernel.Date, rows []HoldingValuation) kernel.Date {
	through := today
	for i := range rows {
		if day := rows[i].Date.AddDays(-1); !rows[i].Valued() && day.Before(through) {
			through = day
		}
	}
	if hasPrior && prior.Before(from.AddDays(-1)) && prior.Before(through) {
		through = prior
	}
	return through
}

// Revaluation is a pending revaluation of one account from a date on (D-36).
type Revaluation struct {
	Account ledger.AccountID
	From    kernel.Date
}

// Coalesce merges revaluations of one account into one from the earliest
// date (D-36), in account order.
func Coalesce(rs []Revaluation) []Revaluation {
	earliest := map[ledger.AccountID]kernel.Date{}
	for _, r := range rs {
		if from, ok := earliest[r.Account]; !ok || r.From.Before(from) {
			earliest[r.Account] = r.From
		}
	}
	out := make([]Revaluation, 0, len(earliest))
	for account, from := range earliest {
		out = append(out, Revaluation{Account: account, From: from})
	}
	slices.SortFunc(out, func(a, b Revaluation) int { return cmp.Compare(a.Account, b.Account) })
	return out
}
