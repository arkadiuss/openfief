package main

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// The arithmetic checks grade what the calc tool is for (D-228): a model does
// its sums with the tool and gets them exactly. A question that asks how much to
// move to reach weights is graded against what code computes from the
// allocation, never against the model's own sums.
const (
	checkRebalance = "rebalance"
	checkExact     = "exact"
	checkOnly      = "onlyNumbers"
	checkFees      = "beforeFees"
)

// The faults of a case the rebalance check cannot grade.
var (
	errNoClasses   = errors.New("the case names no asset classes")
	errNoSuchClass = errors.New("the allocation has no asset class")
)

// Rebalance is what a rebalance question asks: the groups the person names, each
// made of asset classes the allocation lists, and the weights in percent they
// state. The expected amounts are computed from the allocation of the bench's
// day, by code.
type Rebalance struct {
	// Groups are the asset classes of the allocation that each named group is
	// made of, such as stocks of equity and etf.
	Groups map[string][]string `yaml:"groups"`
	// Weights are the percent each named group is to reach.
	Weights map[string]int `yaml:"weights"`
}

// rebalanceAmounts checks that the answer gives the amounts to move that code
// computes from the allocation: for each named group, its target (the weight
// of the total) less what it holds, as a number of the answer, within the
// rounding its writing allows and ignoring its sign. Two readings of the total
// are accepted, since the question names no cash: the portfolio with its cash,
// which the weights then leave at zero, and only what the named groups hold.
func rebalanceAmounts(d *RunData) string {
	r := d.Case.Expect.Rebalance
	if r == nil || len(r.Weights) == 0 {
		return ""
	}
	held, total, err := r.holdings(d.Classes)
	if err != nil {
		return err.Error()
	}
	var named kernel.Decimal
	for _, h := range held {
		named = named.Add(h)
	}
	answer := numbersIn(d.Run.Text)
	var wants []string
	for _, base := range []kernel.Decimal{total, named} {
		moves := r.moves(held, base)
		if allPresent(answer, moves) {
			return ""
		}
		wants = append(wants, joinDecimals(moves))
	}
	return "the amounts to move are not those computed from the allocation: " + strings.Join(wants, " or ")
}

// holdings returns what each named group holds, from the allocation by asset
// class, and the whole portfolio.
func (r *Rebalance) holdings(classes map[string]kernel.Decimal) (map[string]kernel.Decimal, kernel.Decimal, error) {
	held := map[string]kernel.Decimal{}
	var total kernel.Decimal
	for _, v := range classes {
		total = total.Add(v)
	}
	for group := range r.Weights {
		classesOf, ok := r.Groups[group]
		if !ok {
			return nil, total, fmt.Errorf("%w for %s", errNoClasses, group)
		}
		for _, c := range classesOf {
			v, known := classes[c]
			if !known {
				return nil, total, fmt.Errorf("%w %s", errNoSuchClass, c)
			}
			held[group] = held[group].Add(v)
		}
	}
	return held, total, nil
}

// moves returns, for each named group in name order, the size of the amount to
// move to reach its weight of base, to hundredths.
func (r *Rebalance) moves(held map[string]kernel.Decimal, base kernel.Decimal) []kernel.Decimal {
	var out []kernel.Decimal
	for _, group := range slices.Sorted(maps.Keys(r.Weights)) {
		target := base.Mul(kernel.NewDecimal(int64(r.Weights[group]), 0))
		target, err := target.Quo(kernel.NewDecimal(percentScale, 0), moneyPlaces, kernel.RoundHalfEven)
		if err != nil { // the divisor is a constant
			continue
		}
		out = append(out, absDecimal(target.Sub(held[group])))
	}
	return out
}

// allPresent reports whether each of wants is a number of the answer.
func allPresent(answer []foundNumber, wants []kernel.Decimal) bool {
	for _, w := range wants {
		found := false
		for i := range answer {
			if grounded(&answer[i], []kernel.Decimal{w}) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return len(wants) > 0
}

// joinDecimals writes decimals as a comma-separated list.
func joinDecimals(ds []kernel.Decimal) string {
	parts := make([]string, len(ds))
	for i, d := range ds {
		parts[i] = d.String()
	}
	return strings.Join(parts, ", ")
}

// answerHasExact checks that every number the case lists is in the answer
// exactly, not rounded: 0.3 for 0.1 plus 0.2, and every digit of a long
// decimal. A float's 0.30000000000000004 is not 0.3. The answer is read as
// text, since the numbers check's reading splits a long decimal at its groups
// of three digits.
func answerHasExact(d *RunData) string {
	var missing []string
	for _, want := range d.Case.Exact {
		if _, err := kernel.ParseDecimal(want); err != nil {
			return "the case's exact number " + want + " is no number"
		}
		if !hasExactNumber(d.Run.Text, want) {
			missing = append(missing, want)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	return "the answer lacks the exact number: " + strings.Join(missing, ", ")
}

// hasExactNumber reports whether text writes the decimal want, as a plain
// decimal text, in full: in any of the ways the owner's number format writes it
// (a point or a comma for the decimal, and the integer digits grouped by a
// comma, a space or none), with no digit before or after it.
func hasExactNumber(text, want string) bool {
	whole, frac, hasFrac := strings.Cut(strings.TrimPrefix(want, "-"), ".")
	for _, group := range []string{"", ",", " ", "\u00a0", "\u202f", "."} {
		for _, decimal := range []string{".", ","} {
			if group == decimal {
				continue
			}
			written := grouped(whole, group)
			if hasFrac {
				written += decimal + frac
			}
			if containsWhole(text, written) {
				return true
			}
		}
	}
	return false
}

// grouped writes the digits with sep between each group of three from the right.
func grouped(digits, sep string) string {
	if sep == "" {
		return digits
	}
	var b strings.Builder
	for i, c := range digits {
		if i > 0 && (len(digits)-i)%groupDigits == 0 {
			b.WriteString(sep)
		}
		b.WriteRune(c)
	}
	return b.String()
}

// groupDigits is how many digits a group of an integer part has.
const groupDigits = 3

// containsWhole reports whether text holds written with no digit right before
// or after it.
func containsWhole(text, written string) bool {
	for from := 0; ; {
		i := strings.Index(text[from:], written)
		if i < 0 {
			return false
		}
		start, end := from+i, from+i+len(written)
		before := start == 0 || !isDigit(text[start-1])
		after := end == len(text) || !isDigit(text[end])
		if before && after {
			return true
		}
		from = start + 1
	}
}

// isDigit reports whether b is an ASCII digit.
func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// answerHasOnly checks that every number of the answer is one the case lists,
// as the question's own figures: an answer that refuses to compute from weights
// that don't add up gives no amount of its own.
func answerHasOnly(d *RunData) string {
	var allowed []kernel.Decimal
	for _, a := range d.Case.Only {
		v, err := kernel.ParseDecimal(a)
		if err != nil {
			return "the case's number " + a + " is no number"
		}
		allowed = append(allowed, v)
	}
	var stray []string
	for _, n := range numbersIn(d.Run.Text) {
		if !grounded(&n, allowed) {
			stray = append(stray, n.text)
		}
	}
	if len(stray) == 0 {
		return ""
	}
	return "numbers the question and its sum don't give: " + strings.Join(stray, ", ")
}

// saysBeforeFees checks that the answer says plainly that the amounts are before
// taxes and fees (D-228): they are at gross, and a sale would cost both.
func saysBeforeFees(d *RunData) string {
	if hasPhrase(d.answer(), "before taxes and fees") {
		return ""
	}
	return `the answer does not say the amounts are "before taxes and fees"`
}
