package valuation

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"

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
)

// String returns the reason's name: price, fx_rate or index_reading.
func (r MissingReason) String() string {
	switch r {
	case MissingPrice:
		return "price"
	case MissingFXRate:
		return "fx_rate"
	case MissingIndexReading:
		return "index_reading"
	}
	return "MissingReason(" + strconv.Itoa(int(r)) + ")"
}

// ParseMissingReason parses a reason name as String writes it.
func ParseMissingReason(s string) (MissingReason, error) {
	for _, r := range []MissingReason{MissingPrice, MissingFXRate, MissingIndexReading} {
		if r.String() == s {
			return r, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownMissingReason, s)
}

// HoldingValuation is the stored valuation of one holding on one day
// (D-36): its value in the instrument's own currency and in PLN, on a
// basis, with the inputs' source and dates. When an input is missing, Missing
// says which and Detail why, and Value and PLN are zero: the row records the
// gap instead of failing the revaluation.
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
}

// Valued reports whether no input was missing.
func (v *HoldingValuation) Valued() bool { return v.Missing == 0 }

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

// ValueMarketHolding values h on h.Date at the latest price of its
// instrument on or before that date (D-21), converted to PLN at the latest
// FX rate on or before it (D-20) and rounded with ValueRounding. A missing
// price or rate gives a row with Missing set, not an error.
func ValueMarketHolding(h *ledger.Holding, prices []Price, rates []FXRate) (HoldingValuation, error) {
	row := HoldingValuation{
		Account: h.Account, Instrument: h.Instrument, Date: h.Date, Basis: Market, Quantity: h.Quantity,
	}
	p, err := PriceOnOrBefore(prices, h.Instrument.String(), h.Date)
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
		row.PLN, err = roundPLN(v.Value.Amount(), ValueRounding)
		return row, err
	}
	rate, err := RateOnOrBefore(rates, v.Value.Currency(), h.Date)
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
	return row, nil
}

// BondHolding is what valuing a holding of retail treasury bonds needs
// besides the holding: its open lots, its series terms, the period index
// readings, the account's tax wrapper and the configured bond basis.
type BondHolding struct {
	Lots     []ledger.Lot
	Terms    *pl.Terms
	Readings []IndexReading
	Wrapper  pl.TaxWrapper
	Basis    Basis
}

// ValueBondHolding values h on h.Date as the sum of its lots' statutory
// values (pl.ValueLot) on the configured bond basis (D-25): net if redeemed
// unless gross_accrued is configured. Each lot's purchase date is its open
// date and its price per bond is its cost over its quantity, which must be in
// PLN. A missing period index reading gives a row with Missing set, not an
// error.
func ValueBondHolding(h *ledger.Holding, b *BondHolding) (HoldingValuation, error) {
	basis, err := BondBasis(b.Basis)
	if err != nil {
		return HoldingValuation{}, err
	}
	row := HoldingValuation{
		Account: h.Account, Instrument: h.Instrument, Date: h.Date, Basis: basis, Quantity: h.Quantity,
		Source: StatutorySource,
	}
	total := kernel.NewMoney(kernel.Decimal{}, PLN())
	for i := range b.Lots {
		v, err := valueBondLot(h, &b.Lots[i], b, basis)
		if missingIndex, ok := errors.AsType[*pl.MissingIndexError](err); ok {
			return missing(&row, MissingIndexReading, missingIndex), nil
		}
		if err != nil {
			return HoldingValuation{}, err
		}
		if total, err = total.Add(v.Value); err != nil {
			return HoldingValuation{}, fmt.Errorf("valuing %s: %w", h.Instrument, err)
		}
	}
	row.Value, row.PLN = total, total
	return row, nil
}

// valueBondLot values one lot of h with pl.ValueLot on basis.
func valueBondLot(h *ledger.Holding, lot *ledger.Lot, b *BondHolding, basis Basis) (Valuation, error) {
	if !lot.Cost.Currency().Equal(PLN()) || lot.Quantity.Sign() <= 0 {
		return Valuation{}, fmt.Errorf("%w: %s lot %s: cost %s for %s bonds", ErrBondLot, h.Instrument, lot.Origin, lot.Cost, lot.Quantity)
	}
	price, err := lot.Cost.Amount().Quo(lot.Quantity.Decimal(), purchasePricePlaces, kernel.RoundHalfEven)
	if err != nil {
		return Valuation{}, fmt.Errorf("%w: %s lot %s: %w", ErrBondLot, h.Instrument, lot.Origin, err)
	}
	idx, err := LotIndices(b.Terms.Series, lot.Opened, h.Date, b.Readings)
	if err != nil {
		return Valuation{}, err
	}
	purchase := pl.Purchase{Date: lot.Opened, Price: kernel.NewMoney(price, PLN())}
	v, err := pl.ValueLot(b.Terms, idx, lot.Quantity, purchase, b.Wrapper, h.Date)
	if err != nil {
		return Valuation{}, fmt.Errorf("valuing %s lot %s: %w", h.Instrument, lot.Origin, err)
	}
	return BondValuation(h.Instrument.String(), h.Date, &v, basis, StatutorySource)
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
