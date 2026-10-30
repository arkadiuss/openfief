package views

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrUnknownValueKind reports a ValueKind outside the declared constants.
var ErrUnknownValueKind = errors.New("unknown value kind")

// ValueKind is which of the three values of D-77 a view shows.
type ValueKind int

// The value kinds. The zero ValueKind is not a kind.
const (
	// Gross is the value before any tax or fee, the one broker statements
	// show and the default.
	Gross ValueKind = iota + 1
	// Net is what the owner keeps at the planned exit.
	Net
	// Liquidation is the cash in hand if everything were sold today.
	Liquidation
)

// String returns the kind's name: gross, net or liquidation.
func (k ValueKind) String() string {
	switch k {
	case Gross:
		return "gross"
	case Net:
		return "net"
	case Liquidation:
		return "liquidation"
	}
	return fmt.Sprintf("ValueKind(%d)", int(k))
}

// HoldingValue returns v's value of kind k in PLN, and the reason it is
// missing when it has none: the row's missing input for any kind, or the
// net or liquidation value's own reason.
func HoldingValue(v *valuation.HoldingValuation, k ValueKind) (kernel.Money, valuation.MissingReason, error) {
	if !v.Valued() {
		return kernel.Money{}, v.Missing, nil
	}
	switch k {
	case Gross:
		return v.PLN, 0, nil
	case Net:
		if v.NetMissing != 0 {
			return kernel.Money{}, v.NetMissing, nil
		}
		return v.Net, 0, nil
	case Liquidation:
		if v.LiquidationMissing != 0 {
			return kernel.Money{}, v.LiquidationMissing, nil
		}
		return v.Liquidation, 0, nil
	}
	return kernel.Money{}, 0, fmt.Errorf("%w: %d", ErrUnknownValueKind, k)
}

// CashValue returns cash's value of kind k from its three values, and
// false when the kind is unknown for want of an owner setting.
func CashValue(v *jurisdiction.Values, k ValueKind) (kernel.Money, bool, error) {
	switch k {
	case Gross:
		return v.Gross, true, nil
	case Net:
		return v.Net, true, nil
	case Liquidation:
		return v.Liquidation, !v.LiquidationUnknown, nil
	}
	return kernel.Money{}, false, fmt.Errorf("%w: %d", ErrUnknownValueKind, k)
}

// DayCash is one account's cash in one currency at the end of a day, with
// its three values in the reporting currency; Unconverted when the
// currency had no FX rate on or before the day, so it has no values.
type DayCash struct {
	Date        kernel.Date
	Amount      kernel.Money
	Values      jurisdiction.Values
	Unconverted bool
}

// CashBalance is an account's cash in one currency at the end of a day.
type CashBalance struct {
	Date   kernel.Date
	Amount kernel.Money
}

// CashOnDays returns the non-zero cash per currency at the end of each day
// of the range from..to, both inclusive, of one account's current
// transactions txs, in date then currency code order: the sum of the cash
// postings dated on or before the day.
func CashOnDays(txs []ledger.Transaction, from, to kernel.Date) []CashBalance {
	sorted := slices.Clone(txs)
	slices.SortStableFunc(sorted, func(a, b ledger.Transaction) int { return a.Date.Compare(b.Date) })
	balances := map[kernel.Currency]kernel.Decimal{}
	var out []CashBalance
	next := 0
	for day := from; !day.After(to); day = day.AddDays(1) {
		for ; next < len(sorted) && !sorted[next].Date.After(day); next++ {
			addCash(balances, &sorted[next])
		}
		out = append(out, nonZeroCash(balances, day)...)
	}
	return out
}

// addCash adds t's cash postings to balances.
func addCash(balances map[kernel.Currency]kernel.Decimal, t *ledger.Transaction) {
	for _, p := range t.Postings {
		if p.Book.Kind == ledger.BookCash {
			balances[p.Amount.Currency()] = balances[p.Amount.Currency()].Add(p.Amount.Amount())
		}
	}
}

// nonZeroCash returns the non-zero balances on day, in currency code order.
func nonZeroCash(balances map[kernel.Currency]kernel.Decimal, day kernel.Date) []CashBalance {
	var out []CashBalance
	for c, d := range balances {
		if !d.IsZero() {
			out = append(out, CashBalance{Date: day, Amount: kernel.NewMoney(d, c)})
		}
	}
	slices.SortFunc(out, func(a, b CashBalance) int {
		return cmp.Compare(a.Amount.Currency().String(), b.Amount.Currency().String())
	})
	return out
}
