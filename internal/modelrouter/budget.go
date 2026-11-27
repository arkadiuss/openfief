package modelrouter

import (
	"errors"
	"fmt"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of the day's budget (D-208).
var (
	// ErrBudgetSpent reports that the day's budget for model calls is used
	// up: the router refuses a call before sending it.
	ErrBudgetSpent = errors.New("the day's model budget is spent")
	// ErrUnpriceable reports a call the budget can't count against the cap, so
	// the router does not send it: a model entry priced in another currency
	// than the budget's, or ErrNoPrice.
	ErrUnpriceable = errors.New("a model call can't be priced against the budget")
	// ErrNoPrice is an ErrUnpriceable: no price is known for the model, because
	// the catalog has no entry for it, as after a restart until its provider
	// lists its models, or because its entry carries no price although a third
	// party runs it.
	ErrNoPrice = fmt.Errorf("%w: no price is known for its model", ErrUnpriceable)
)

// microPer is the power of ten of the million tokens a model entry's price is quoted for.
const microPer = 6

// Budget is the day's spending limit on calls to models (D-208): a cap in one
// currency, counted over every call of every role, each priced from its model
// entry's price per million tokens. The zero Budget has no cap, which is what
// an installation that sets none gets.
type Budget struct {
	// DailyCap is the most the day's calls may cost; zero or less is no cap.
	DailyCap kernel.Decimal
	// Currency is the currency of the cap, and the only one entries are
	// priced in while a cap is set.
	Currency kernel.Currency
}

// NewBudget returns a budget of cap in currency, or an error for a cap that is
// negative or has no currency. A zero cap is no budget.
func NewBudget(capacity kernel.Decimal, currency kernel.Currency) (Budget, error) {
	if capacity.Sign() < 0 {
		return Budget{}, fmt.Errorf("%w: the cap is negative", ErrInvalidBudget)
	}
	if capacity.Sign() > 0 && currency.IsZero() {
		return Budget{}, fmt.Errorf("%w: the cap has no currency", ErrInvalidBudget)
	}
	if capacity.IsZero() {
		return Budget{}, nil
	}
	return Budget{DailyCap: capacity, Currency: currency}, nil
}

// ErrInvalidBudget reports a budget with a negative cap or none without a
// currency.
var ErrInvalidBudget = errors.New("invalid model budget")

// Enabled reports whether b caps the day's spending.
func (b Budget) Enabled() bool { return b.DailyCap.Sign() > 0 }

// Exceeded reports whether spent has reached the cap: a day whose calls cost
// exactly the cap is spent. Without a cap nothing is exceeded.
func (b Budget) Exceeded(spent kernel.Decimal) bool {
	return b.Enabled() && spent.Cmp(b.DailyCap) >= 0
}

// BudgetSpentError is ErrBudgetSpent with the moment the budget starts over:
// the end of the day it was counted for.
type BudgetSpentError struct {
	ResetsAt time.Time
}

// Error says the budget is spent and never more.
func (e *BudgetSpentError) Error() string { return ErrBudgetSpent.Error() }

// Unwrap makes errors.Is(err, ErrBudgetSpent) true.
func (e *BudgetSpentError) Unwrap() error { return ErrBudgetSpent }

// CallCost prices what a call used at its entry's price per million tokens,
// input and output each at its own price. An entry with no price, such as a
// model on the owner's hardware, costs nothing. A price in another currency
// than the budget's can't be counted: the error wraps ErrUnpriceable.
func CallCost(e *ModelEntry, u Usage, currency kernel.Currency) (kernel.Decimal, error) {
	if !e.HasPrice {
		return kernel.Decimal{}, nil
	}
	if !e.PriceCurrency.Equal(currency) {
		return kernel.Decimal{}, fmt.Errorf("%w: the entry is priced in %s", ErrUnpriceable, e.PriceCurrency)
	}
	perToken := kernel.NewDecimal(1, -microPer)
	in := kernel.NewDecimal(u.InputTokens, 0).Mul(e.InputPrice)
	out := kernel.NewDecimal(u.OutputTokens, 0).Mul(e.OutputPrice)
	return in.Add(out).Mul(perToken), nil
}

// BudgetDay returns the calendar date of t in loc, which is the day the
// budget counts t's calls in, and the moment that day ends, when the next
// one's budget begins.
func BudgetDay(t time.Time, loc *time.Location) (kernel.Date, time.Time) {
	local := t.In(loc)
	day, err := kernel.NewDate(local.Year(), local.Month(), local.Day())
	if err != nil {
		// A time.Time always has a valid calendar date.
		return kernel.Date{}, local
	}
	return day, time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, loc)
}
