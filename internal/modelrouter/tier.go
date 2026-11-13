package modelrouter

import (
	"errors"
	"fmt"
	"slices"
)

// Errors about data tiers.
var (
	// ErrUnknownDataTier reports a tier that is none of T0, TF, T1 and T2.
	ErrUnknownDataTier = errors.New("unknown data tier")
	// ErrUntagged reports a request with a piece of data that carries no
	// tier, or with no data at all: data nobody classified is not sent.
	ErrUntagged = errors.New("request data is not tagged with a data tier")
	// ErrInvalidDataTierModels reports a tier's list with too many models or
	// one twice.
	ErrInvalidDataTierModels = errors.New("invalid models for a tier")
)

// DataTier is the data tier of a piece of data a role sends a model (D-206).
// DataTiers are not ordered: each has its own list of models the owner allows.
type DataTier string

// The data tiers (D-206).
const (
	// DataTierT0 is whole-portfolio holdings, amounts, transactions and tax
	// data.
	DataTierT0 DataTier = "T0"
	// DataTierTF is one fief's own lease, cash, positions and working orders.
	DataTierTF DataTier = "TF"
	// DataTierT1 is a thesis's text and instruments.
	DataTierT1 DataTier = "T1"
	// DataTierT2 is public data such as prices, filings and news.
	DataTierT2 DataTier = "T2"
)

// DataTiers returns every tier, most sensitive first.
func DataTiers() []DataTier { return []DataTier{DataTierT0, DataTierTF, DataTierT1, DataTierT2} }

// ParseDataTier parses a tier as its constant spells it.
func ParseDataTier(s string) (DataTier, error) {
	t := DataTier(s)
	if !slices.Contains(DataTiers(), t) {
		return "", fmt.Errorf("%w: %q", ErrUnknownDataTier, s)
	}
	return t, nil
}

// DataTierSet is the distinct tiers of a request's data, in DataTiers' order.
type DataTierSet []DataTier

// NewDataTierSet returns the set of tags, or an error wrapping ErrUntagged when
// there are none or one is empty, or ErrUnknownDataTier when one is no tier.
func NewDataTierSet(tags []DataTier) (DataTierSet, error) {
	if len(tags) == 0 {
		return nil, ErrUntagged
	}
	var set DataTierSet
	for _, t := range DataTiers() {
		if slices.Contains(tags, t) {
			set = append(set, t)
		}
	}
	for _, tag := range tags {
		if tag == "" {
			return nil, ErrUntagged
		}
		if !slices.Contains(set, tag) {
			return nil, fmt.Errorf("%w: %q", ErrUnknownDataTier, string(tag))
		}
	}
	return set, nil
}

// DataTierPolicy lists, per tier, the models the owner allows to receive it
// (D-206). A tier with no list, or a model not on it, is denied: the policy
// denies by default.
type DataTierPolicy map[DataTier][]Ref

// NewDataTierPolicy returns the policy over lists, or an error wrapping
// ErrUnknownDataTier, ErrInvalidRef or ErrInvalidDataTierModels.
func NewDataTierPolicy(lists map[DataTier][]Ref) (DataTierPolicy, error) {
	p := make(DataTierPolicy, len(lists))
	for tier, refs := range lists {
		if !slices.Contains(DataTiers(), tier) {
			return nil, fmt.Errorf("%w: %q", ErrUnknownDataTier, string(tier))
		}
		if err := validateDataTierModels(refs); err != nil {
			return nil, err
		}
		if len(refs) > 0 {
			p[tier] = slices.Clone(refs)
		}
	}
	return p, nil
}

// validateDataTierModels checks one tier's list.
func validateDataTierModels(refs []Ref) error {
	if len(refs) > MaxDataTierModels {
		return fmt.Errorf("%w: at most %d", ErrInvalidDataTierModels, MaxDataTierModels)
	}
	for i, r := range refs {
		if !r.valid() {
			return fmt.Errorf("%w: %q", ErrInvalidRef, r.String())
		}
		if slices.Contains(refs[:i], r) {
			return fmt.Errorf("%w: %q twice", ErrInvalidDataTierModels, r.String())
		}
	}
	return nil
}

// Allows reports whether model may receive data of tier.
func (p DataTierPolicy) Allows(tier DataTier, model Ref) bool {
	return slices.Contains(p[tier], model)
}

// Denied returns the tiers of tiers that model may not receive, in the
// order of tiers; none when model is allowed for every one of them.
func (p DataTierPolicy) Denied(model Ref, tiers DataTierSet) []DataTier {
	var denied []DataTier
	for _, t := range tiers {
		if !p.Allows(t, model) {
			denied = append(denied, t)
		}
	}
	return denied
}
