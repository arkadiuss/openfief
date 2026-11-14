package main

import (
	"math/big"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// One-step arithmetic. A model that is asked why a value changed may write the
// change itself, "an increase of 1,966.85 PLN", a number no card holds but that
// two cards give. A number passes the numbers check when it is the difference or
// the sum of two numbers that stand in the cards shown, or, followed by a
// percent sign, the percentage change of one to the other, and nothing more: no
// chain of operations and no number from outside the cards.
//
// The rule is exact. The result is computed from the cards' own numbers and
// rounded half-up to the places the answer writes it with, at most two (grosze
// for an amount, 0.01 of a percent), and the answer must equal it: 1,966.85 and
// 1,967 pass for a difference of 1,966.85, and 1,966.00, 1,966 and 2,000 do not.

// maxDerivedPlaces is the most decimal places a derived number is held to: an
// amount is exact to the grosz and a percent to 0.01.
const maxDerivedPlaces = 2

// derivedFrom reports whether n is the one-step result of two of the card
// numbers, under one of its readings. A number written with a scale word ("k",
// "mln") is not: its precision is that of the scale.
func derivedFrom(n *foundNumber, card []kernel.Decimal) bool {
	values := distinctRats(card)
	for i := range n.readings {
		r := &n.readings[i]
		if r.scaled {
			continue
		}
		want, ok := rationalOf(r.value)
		if !ok {
			continue
		}
		places := min(r.places, maxDerivedPlaces)
		if matchesPair(values, want, places, n.percent) {
			return true
		}
	}
	return false
}

// matchesPair reports whether want is, rounded half-up to places, the
// difference or the sum of two of values (a percentage change for a percent).
func matchesPair(values []*big.Rat, want *big.Rat, places int, percent bool) bool {
	for i, a := range values {
		for j, b := range values {
			if i == j {
				continue
			}
			for _, got := range oneStep(a, b, percent) {
				if roundHalfUp(got, places).Cmp(want) == 0 {
					return true
				}
			}
		}
	}
	return false
}

// oneStep returns the results of one operation on a and b that a number of the
// kind can be: for an amount their difference (a minus b, whose absolute value
// is the other order's) and their sum; for a percent the change from a to b,
// as a percentage of a. A sign is ignored, as everywhere in the numbers check.
func oneStep(a, b *big.Rat, percent bool) []*big.Rat {
	if percent {
		if a.Sign() == 0 {
			return nil
		}
		change := new(big.Rat).Sub(b, a)
		change.Quo(change, a)
		return []*big.Rat{new(big.Rat).Abs(change.Mul(change, big.NewRat(percentScale, 1)))}
	}
	diff := new(big.Rat).Abs(new(big.Rat).Sub(a, b))
	return []*big.Rat{diff, new(big.Rat).Add(a, b)}
}

// percentScale turns a ratio into percent; decimalBase and halfDenominator are
// the base of the places and the denominator of one half.
const (
	percentScale    = 100
	decimalBase     = 10
	halfDenominator = 2
)

// roundHalfUp returns the non-negative x rounded half-up to places decimal
// places, exactly.
func roundHalfUp(x *big.Rat, places int) *big.Rat {
	scale := new(big.Int).Exp(big.NewInt(decimalBase), big.NewInt(int64(places)), nil)
	scaled := new(big.Rat).Mul(x, new(big.Rat).SetInt(scale))
	// Half-up: add one half, then take the whole part.
	scaled.Add(scaled, big.NewRat(1, halfDenominator))
	whole := new(big.Int).Quo(scaled.Num(), scaled.Denom())
	return new(big.Rat).SetFrac(whole, scale)
}

// rationalOf returns d as an exact rational, ok false when it does not read.
func rationalOf(d kernel.Decimal) (*big.Rat, bool) {
	return new(big.Rat).SetString(d.String())
}

// distinctRats returns the distinct values of card, exactly.
func distinctRats(card []kernel.Decimal) []*big.Rat {
	seen := map[string]bool{}
	var out []*big.Rat
	for _, d := range card {
		r, ok := rationalOf(d)
		if !ok || seen[r.RatString()] {
			continue
		}
		seen[r.RatString()] = true
		out = append(out, r)
	}
	return out
}
