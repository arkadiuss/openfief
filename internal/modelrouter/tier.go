package modelrouter

import (
	"errors"
	"fmt"
	"slices"
)

// Errors about data tiers.
var (
	// ErrUnknownTier reports a tier that is none of T0, TF, T1 and T2.
	ErrUnknownTier = errors.New("unknown data tier")
	// ErrUntagged reports a request with a piece of data that carries no
	// tier, or with no data at all: data nobody classified is not sent.
	ErrUntagged = errors.New("request data is not tagged with a data tier")
	// ErrInvalidTierModels reports a tier's list with too many models or
	// one twice.
	ErrInvalidTierModels = errors.New("invalid models for a tier")
)

// Tier is the data tier of a piece of data a role sends a model (D-206).
// Tiers are not ordered: each has its own list of models the owner allows.
type Tier string

// The data tiers (D-206).
const (
	// TierT0 is whole-portfolio holdings, amounts, transactions and tax
	// data.
	TierT0 Tier = "T0"
	// TierTF is one fief's own lease, cash, positions and working orders.
	TierTF Tier = "TF"
	// TierT1 is a thesis's text and instruments.
	TierT1 Tier = "T1"
	// TierT2 is public data such as prices, filings and news.
	TierT2 Tier = "T2"
)

// Tiers returns every tier, most sensitive first.
func Tiers() []Tier { return []Tier{TierT0, TierTF, TierT1, TierT2} }

// ParseTier parses a tier as its constant spells it.
func ParseTier(s string) (Tier, error) {
	t := Tier(s)
	if !slices.Contains(Tiers(), t) {
		return "", fmt.Errorf("%w: %q", ErrUnknownTier, s)
	}
	return t, nil
}

// TierSet is the distinct tiers of a request's data, in Tiers' order.
type TierSet []Tier

// NewTierSet returns the set of tags, or an error wrapping ErrUntagged when
// there are none or one is empty, or ErrUnknownTier when one is no tier.
func NewTierSet(tags []Tier) (TierSet, error) {
	if len(tags) == 0 {
		return nil, ErrUntagged
	}
	var set TierSet
	for _, t := range Tiers() {
		if slices.Contains(tags, t) {
			set = append(set, t)
		}
	}
	for _, tag := range tags {
		if tag == "" {
			return nil, ErrUntagged
		}
		if !slices.Contains(set, tag) {
			return nil, fmt.Errorf("%w: %q", ErrUnknownTier, string(tag))
		}
	}
	return set, nil
}

// TierPolicy lists, per tier, the models the owner allows to receive it
// (D-206). A tier with no list, or a model not on it, is denied: the policy
// denies by default.
type TierPolicy map[Tier][]Ref

// NewTierPolicy returns the policy over lists, or an error wrapping
// ErrUnknownTier, ErrInvalidRef or ErrInvalidTierModels.
func NewTierPolicy(lists map[Tier][]Ref) (TierPolicy, error) {
	p := make(TierPolicy, len(lists))
	for tier, refs := range lists {
		if !slices.Contains(Tiers(), tier) {
			return nil, fmt.Errorf("%w: %q", ErrUnknownTier, string(tier))
		}
		if err := validateTierModels(refs); err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			p[tier] = slices.Clone(refs)
		}
	}
	return p, nil
}

// validateTierModels checks one tier's list.
func validateTierModels(refs []Ref) error {
	if len(refs) > MaxTierModels {
		return fmt.Errorf("%w: at most %d", ErrInvalidTierModels, MaxTierModels)
	}
	for i, r := range refs {
		if !r.valid() {
			return fmt.Errorf("%w: %q", ErrInvalidRef, r.String())
		}
		if slices.Contains(refs[:i], r) {
			return fmt.Errorf("%w: %q twice", ErrInvalidTierModels, r.String())
		}
	}
	return nil
}

// Allows reports whether model may receive data of tier.
func (p TierPolicy) Allows(tier Tier, model Ref) bool {
	return slices.Contains(p[tier], model)
}

// Denied returns the tiers of tiers that model may not receive, in the
// order of tiers; none when model is allowed for every one of them.
func (p TierPolicy) Denied(model Ref, tiers TierSet) []Tier {
	var denied []Tier
	for _, t := range tiers {
		if !p.Allows(t, model) {
			denied = append(denied, t)
		}
	}
	return denied
}
