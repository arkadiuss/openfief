package views

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TWRPlaces is how many decimal places the time-weighted return keeps: each
// day's growth factor and their running product are rounded to it with
// TWRRounding. Each day adds at most 2 × 10^-12 of the growth in error, and
// half-even rounding mostly cancels out: a synthetic ten-year series drifts
// by a few billionths of an index point (TestPerformanceDrift).
const TWRPlaces int32 = 12

// TWRRounding is how the time-weighted return's factors are rounded:
// half-even, so rounding errors don't lean one way over thousands of days.
const TWRRounding = kernel.RoundHalfEven

// valueRounding rounds cash, money in transit and flows converted to the
// reporting currency to its minor unit, as the valuation context rounds
// holding values, so they add up alike.
const valueRounding = valuation.ValueRounding

// indexBase is the value of the performance index on the series' first day.
const indexBase = 100

// PerformanceGap says why a performance series ends before the last day
// asked for. The zero value means it doesn't.
type PerformanceGap int

const (
	// GapUnvalued is a day after an account's valued-through date, or one
	// on which a holding has no valuation, a valuation lacking an input, or
	// one of a quantity the transaction log no longer gives.
	GapUnvalued PerformanceGap = iota + 1
	// GapNoFXRate is a day on which cash, a cash flow or money in transit
	// is in a currency with no FX rate on or before it.
	GapNoFXRate
)

// String returns the gap's name: unvalued or no_fx_rate.
func (g PerformanceGap) String() string {
	switch g {
	case GapUnvalued:
		return "unvalued"
	case GapNoFXRate:
		return "no_fx_rate"
	}
	return fmt.Sprintf("PerformanceGap(%d)", int(g))
}

// PerformanceInput is what the portfolio's time-weighted return is computed
// from: every account's current transactions, their stored holding
// valuations and the FX rates to PLN.
type PerformanceInput struct {
	// From and To are the first and last day asked for, inclusive.
	From, To kernel.Date
	// ValuedThrough is the earliest valued-through date of the accounts:
	// no later day is valued for all of them.
	ValuedThrough kernel.Date
	// Transactions are the current transactions of every account, any
	// order, from the first on: cash and holdings are summed from them.
	Transactions []ledger.Transaction
	// Valuations are the holding valuations of every account dated From to
	// To.
	Valuations []valuation.HoldingValuation
	// Rates are FX rates to PLN, any order, of every currency the accounts'
	// cash is in.
	Rates     []valuation.FXRate
	Reporting kernel.Currency
	// Links are the owner's current transfer links (D-105): only the
	// confirmed ones leave cash flows out.
	Links []ledger.TransferLink
}

// PerformanceDay is the portfolio at the end of one day, in the reporting
// currency: its value, holdings plus cash plus money in transit, the net
// external cash flow of the day, and the time-weighted return since the
// first day as an index that starts at 100.
type PerformanceDay struct {
	Date      kernel.Date
	Value     kernel.Money
	Holdings  kernel.Money
	Cash      kernel.Money
	InTransit kernel.Money
	Flow      kernel.Money
	Index     kernel.Decimal
}

// PerformanceSeries is the portfolio's time-weighted return per day (D-61),
// with the confirmed transfers between its accounts, which were not counted
// as cash flows. Gap says why Days ends before the last day asked for.
type PerformanceSeries struct {
	Days      []PerformanceDay
	Gap       PerformanceGap
	Transfers []Transfer
}

// Performance computes the portfolio's time-weighted return (D-61) for each
// day from in.From to in.To: holdings at their stored valuations plus cash,
// and money in transit between the owner's accounts, converted to the
// reporting currency at each day's FX rate, chained from day to day across
// external cash flows so they don't count as performance. Flows are taken
// at the end of their day: a day's factor is (value - flow) / the previous
// day's value, rounded to TWRPlaces, and 1 when the previous value is not
// positive. External flows are deposits and withdrawals, but not the legs of
// a transfer the owner confirmed (ConfirmedTransfers, D-105), and opening
// lots, at their share of the holding's
// valuation. The series stops before the first day that lacks an input,
// naming the gap. Stored valuations are in PLN, so only PLN reporting adds
// up; any other is an error wrapping ErrNotReportingCurrency.
func Performance(in *PerformanceInput) (PerformanceSeries, error) {
	if err := ValidateRange(in.From, in.To); err != nil {
		return PerformanceSeries{}, fmt.Errorf("performance: %w", err)
	}
	if !in.Reporting.Equal(valuation.PLN()) {
		return PerformanceSeries{}, fmt.Errorf("performance: %w %s: valuations are in %s", ErrNotReportingCurrency, in.Reporting, valuation.PLN())
	}
	transfers, external := ConfirmedTransfers(ExternalFlows(in.Transactions), in.Links)
	b := newDayBuilder(in, external, transfers)
	out := PerformanceSeries{Transfers: transfersIn(transfers, in.From, in.To)}
	growth := kernel.NewDecimal(1, 0)
	for day := in.From; !day.After(in.To); day = day.AddDays(1) {
		if day.After(in.ValuedThrough) {
			out.Gap = GapUnvalued
			break
		}
		pd, gap, err := b.day(day)
		if err != nil {
			return PerformanceSeries{}, fmt.Errorf("performance on %s: %w", day, err)
		}
		if gap != 0 {
			out.Gap = gap
			break
		}
		if n := len(out.Days); n > 0 {
			if growth, err = chain(growth, out.Days[n-1].Value, pd.Value, pd.Flow); err != nil {
				return PerformanceSeries{}, fmt.Errorf("performance on %s: %w", day, err)
			}
		}
		pd.Index = kernel.NewDecimal(indexBase, 0).Mul(growth)
		out.Days = append(out.Days, pd)
	}
	return out, nil
}

// chain returns growth times one day's factor, (value - flow) / prev,
// rounded to TWRPlaces; growth unchanged when prev is not positive, since a
// day that starts with nothing has no return.
func chain(growth kernel.Decimal, prev, value, flow kernel.Money) (kernel.Decimal, error) {
	if prev.Sign() <= 0 {
		return growth, nil
	}
	gain, err := value.Sub(flow)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("day's gain: %w", err)
	}
	factor, err := gain.Amount().Quo(prev.Amount(), TWRPlaces, TWRRounding)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("day's factor: %w", err)
	}
	out, err := growth.Mul(factor).Round(TWRPlaces, TWRRounding)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("chaining: %w", err)
	}
	return out, nil
}

// transfersIn returns the transfers with a leg dated from to to.
func transfersIn(ts []Transfer, from, to kernel.Date) []Transfer {
	var out []Transfer
	for i := range ts {
		w, d := ts[i].Withdrawal.Date, ts[i].Deposit.Date
		if (!w.Before(from) && !w.After(to)) || (!d.Before(from) && !d.After(to)) {
			out = append(out, ts[i])
		}
	}
	return out
}

// cashKey names one account's cash in one currency.
type cashKey struct {
	account  ledger.AccountID
	currency kernel.Currency
}

// holdingKey names one account's holding of one instrument.
type holdingKey struct {
	account    ledger.AccountID
	instrument ledger.Instrument
}

// dayBuilder values the portfolio day by day, in date order: it applies the
// transactions up to each day to running cash balances and holding
// quantities.
type dayBuilder struct {
	reporting   kernel.Currency
	txs         []ledger.Transaction // in date order
	applied     int
	cash        map[cashKey]kernel.Decimal
	units       map[holdingKey]kernel.Quantity
	valuations  map[kernel.Date][]valuation.HoldingValuation
	rates       map[kernel.Currency][]valuation.FXRate
	flows       map[kernel.Date][]CashFlow
	openingLots map[kernel.Date][]ledger.Transaction
	transfers   []Transfer
}

func newDayBuilder(in *PerformanceInput, external []CashFlow, transfers []Transfer) *dayBuilder {
	b := &dayBuilder{
		reporting: in.Reporting, txs: slices.Clone(in.Transactions),
		cash: map[cashKey]kernel.Decimal{}, units: map[holdingKey]kernel.Quantity{},
		valuations: map[kernel.Date][]valuation.HoldingValuation{}, rates: map[kernel.Currency][]valuation.FXRate{},
		flows: map[kernel.Date][]CashFlow{}, openingLots: map[kernel.Date][]ledger.Transaction{}, transfers: transfers,
	}
	slices.SortStableFunc(b.txs, func(x, y ledger.Transaction) int { return x.Date.Compare(y.Date) })
	for i := range in.Valuations {
		v := in.Valuations[i]
		b.valuations[v.Date] = append(b.valuations[v.Date], v)
	}
	for _, r := range in.Rates {
		b.rates[r.Currency] = append(b.rates[r.Currency], r)
	}
	for _, f := range external {
		b.flows[f.Date] = append(b.flows[f.Date], f)
	}
	for i := range b.txs {
		if b.txs[i].Kind == ledger.KindOpeningLot {
			b.openingLots[b.txs[i].Date] = append(b.openingLots[b.txs[i].Date], b.txs[i])
		}
	}
	return b
}

// day values the portfolio at the end of day, which must come after every
// day valued before, or names the gap that leaves it unvalued.
func (b *dayBuilder) day(day kernel.Date) (PerformanceDay, PerformanceGap, error) {
	b.applyThrough(day)
	pd := PerformanceDay{Date: day}
	rows, gap, err := b.holdings(day, &pd)
	if gap != 0 || err != nil {
		return PerformanceDay{}, gap, err
	}
	if gap, err = b.cashOn(day, &pd); gap != 0 || err != nil {
		return PerformanceDay{}, gap, err
	}
	if gap, err = b.inTransit(day, &pd); gap != 0 || err != nil {
		return PerformanceDay{}, gap, err
	}
	if gap, err = b.flowsOn(day, rows, &pd); gap != 0 || err != nil {
		return PerformanceDay{}, gap, err
	}
	if pd.Value, err = sumIn(b.reporting, pd.Holdings, pd.Cash, pd.InTransit); err != nil {
		return PerformanceDay{}, 0, err
	}
	return pd, 0, nil
}

// applyThrough adds the cash and holding postings of the transactions dated
// up to day, not yet applied, to the running balances.
func (b *dayBuilder) applyThrough(day kernel.Date) {
	for ; b.applied < len(b.txs) && !b.txs[b.applied].Date.After(day); b.applied++ {
		t := &b.txs[b.applied]
		for _, p := range t.Postings {
			switch p.Book.Kind {
			case ledger.BookCash:
				k := cashKey{account: t.Account, currency: p.Amount.Currency()}
				b.cash[k] = b.cash[k].Add(p.Amount.Amount())
			case ledger.BookHolding:
				k := holdingKey{account: t.Account, instrument: p.Book.Instrument}
				b.units[k] = b.units[k].Add(p.Quantity)
			case ledger.BookFee, ledger.BookWithholdingTax, ledger.BookFX, ledger.BookEquity:
			}
		}
	}
}

// holdings sums the valuations of day into pd.Holdings and returns them by
// holding. Every holding the transactions give must have exactly one valued
// valuation of its quantity, and no other valuation may be dated day.
func (b *dayBuilder) holdings(day kernel.Date, pd *PerformanceDay) (map[holdingKey]*valuation.HoldingValuation, PerformanceGap, error) {
	rows := map[holdingKey]*valuation.HoldingValuation{}
	vals := b.valuations[day]
	parts := make([]kernel.Money, 0, len(vals))
	for i := range vals {
		v := &vals[i]
		k := holdingKey{account: v.Account, instrument: v.Instrument}
		if _, dup := rows[k]; dup || !v.Valued() || !v.Quantity.Equal(b.units[k]) {
			return nil, GapUnvalued, nil
		}
		rows[k] = v
		parts = append(parts, v.PLN)
	}
	for k, q := range b.units {
		if _, ok := rows[k]; !ok && !q.IsZero() {
			return nil, GapUnvalued, nil
		}
	}
	var err error
	if pd.Holdings, err = sumIn(b.reporting, parts...); err != nil {
		return nil, 0, err
	}
	return rows, 0, nil
}

// cashOn converts each account's cash balance in each currency at the end of
// day and sums them into pd.Cash.
func (b *dayBuilder) cashOn(day kernel.Date, pd *PerformanceDay) (PerformanceGap, error) {
	keys := make([]cashKey, 0, len(b.cash))
	for k, amount := range b.cash {
		if !amount.IsZero() {
			keys = append(keys, k)
		}
	}
	slices.SortFunc(keys, func(x, y cashKey) int {
		return cmp.Or(cmp.Compare(x.account, y.account), cmp.Compare(x.currency.String(), y.currency.String()))
	})
	balances := make([]kernel.Money, 0, len(keys))
	for _, k := range keys {
		balances = append(balances, kernel.NewMoney(b.cash[k], k.currency))
	}
	sum, gap, err := b.convertAll(day, balances)
	pd.Cash = sum
	return gap, err
}

// inTransit sums the money in transit between accounts at the end of day
// into pd.InTransit.
func (b *dayBuilder) inTransit(day kernel.Date, pd *PerformanceDay) (PerformanceGap, error) {
	var moving []kernel.Money
	for i := range b.transfers {
		if m, ok := b.transfers[i].InTransitOn(day); ok {
			moving = append(moving, m)
		}
	}
	sum, gap, err := b.convertAll(day, moving)
	pd.InTransit = sum
	return gap, err
}

// flowsOn sums day's external cash flows into pd.Flow: deposits and
// withdrawals converted at the day's FX rate, and each opening lot at its
// share of its holding's valuation on day, from rows.
func (b *dayBuilder) flowsOn(day kernel.Date, rows map[holdingKey]*valuation.HoldingValuation, pd *PerformanceDay) (PerformanceGap, error) {
	amounts := make([]kernel.Money, 0, len(b.flows[day]))
	for _, f := range b.flows[day] {
		amounts = append(amounts, f.Amount)
	}
	sum, gap, err := b.convertAll(day, amounts)
	if gap != 0 || err != nil {
		return gap, err
	}
	parts := []kernel.Money{sum}
	for i := range b.openingLots[day] {
		lot := &b.openingLots[day][i]
		for _, p := range lot.Postings {
			if p.Book.Kind != ledger.BookHolding {
				continue
			}
			row, ok := rows[holdingKey{account: lot.Account, instrument: p.Book.Instrument}]
			if !ok {
				return GapUnvalued, nil
			}
			share, shareErr := row.PLN.Amount().Mul(p.Quantity.Decimal()).Quo(row.Quantity.Decimal(), plnPlaces, valueRounding)
			if shareErr != nil {
				return 0, fmt.Errorf("opening lot %s: %w", lot.Key, shareErr)
			}
			parts = append(parts, kernel.NewMoney(share, row.PLN.Currency()))
		}
	}
	pd.Flow, err = sumIn(b.reporting, parts...)
	return 0, err
}

// plnPlaces is the number of decimal places of a PLN amount (grosze).
const plnPlaces = 2

// convertAll converts each of ms to the reporting currency at the FX rate
// for day and sums them, or names the gap when a currency has no rate.
func (b *dayBuilder) convertAll(day kernel.Date, ms []kernel.Money) (kernel.Money, PerformanceGap, error) {
	converted := make([]kernel.Money, 0, len(ms))
	for _, m := range ms {
		c, err := valuation.ConvertToPLN(m, b.rates[m.Currency()], day, valueRounding)
		if errors.Is(err, valuation.ErrNoRate) {
			return kernel.Money{}, GapNoFXRate, nil
		}
		if err != nil {
			return kernel.Money{}, 0, fmt.Errorf("converting %s: %w", m.Currency(), err)
		}
		converted = append(converted, c)
	}
	sum, err := sumIn(b.reporting, converted...)
	return sum, 0, err
}

// sumIn sums ms, each in currency c; no amounts sum to zero in c.
func sumIn(c kernel.Currency, ms ...kernel.Money) (kernel.Money, error) {
	sum := kernel.NewMoney(kernel.Decimal{}, c)
	for _, m := range ms {
		var err error
		if sum, err = addIn(c, sum, m); err != nil {
			return kernel.Money{}, err
		}
	}
	return sum, nil
}
