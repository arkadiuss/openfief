package jurisdiction

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidNumberFormat reports a number format with an unknown separator,
// no decimal separator, or the same separator for both jobs.
var ErrInvalidNumberFormat = errors.New("invalid number format")

// Separator is a character a number format writes between digits.
type Separator string

// The separators a number format may use.
const (
	// SeparatorNone writes nothing: thousands are not grouped.
	SeparatorNone Separator = ""
	// SeparatorSpace groups thousands with a space; the UI writes a narrow
	// no-break space, so an amount never wraps.
	SeparatorSpace Separator = " "
	// SeparatorComma is a comma.
	SeparatorComma Separator = ","
	// SeparatorDot is a full stop.
	SeparatorDot Separator = "."
	// SeparatorApostrophe is an apostrophe, as Switzerland groups thousands.
	SeparatorApostrophe Separator = "'"
)

// separatorNames names each separator for a number format's key.
func separatorNames() map[Separator]string {
	return map[Separator]string{
		SeparatorNone: "none", SeparatorSpace: "space", SeparatorComma: "comma",
		SeparatorDot: "dot", SeparatorApostrophe: "apostrophe",
	}
}

// NumberFormat is how amounts are written (D-113): the separator between
// groups of three integer digits and the decimal separator. It changes only
// the separators, never the digits or the currency.
type NumberFormat struct {
	Group   Separator
	Decimal Separator
}

// Validate reports whether f groups with a known separator or none, has a
// comma or a full stop as its decimal separator, and uses different ones.
func (f NumberFormat) Validate() error {
	names := separatorNames()
	if _, ok := names[f.Group]; !ok {
		return fmt.Errorf("%w: group separator %q", ErrInvalidNumberFormat, f.Group)
	}
	if f.Decimal != SeparatorComma && f.Decimal != SeparatorDot {
		return fmt.Errorf("%w: decimal separator %q", ErrInvalidNumberFormat, f.Decimal)
	}
	if f.Group == f.Decimal {
		return fmt.Errorf("%w: %q both groups and separates decimals", ErrInvalidNumberFormat, f.Decimal)
	}
	return nil
}

// String returns f's key, its separators' names joined by a hyphen, such
// as "space-comma".
func (f NumberFormat) String() string {
	names := separatorNames()
	return names[f.Group] + "-" + names[f.Decimal]
}

// ParseNumberFormat parses a key as String writes it, and checks it.
func ParseNumberFormat(key string) (NumberFormat, error) {
	group, decimal, ok := strings.Cut(key, "-")
	if !ok {
		return NumberFormat{}, fmt.Errorf("%w: %q", ErrInvalidNumberFormat, key)
	}
	var f NumberFormat
	foundGroup, foundDecimal := false, false
	for sep, name := range separatorNames() {
		if name == group {
			f.Group, foundGroup = sep, true
		}
		if name == decimal {
			f.Decimal, foundDecimal = sep, true
		}
	}
	if !foundGroup || !foundDecimal {
		return NumberFormat{}, fmt.Errorf("%w: %q", ErrInvalidNumberFormat, key)
	}
	if err := f.Validate(); err != nil {
		return NumberFormat{}, err
	}
	return f, nil
}

// NumberFormats returns the number formats an owner picks from, in the
// order a form lists them.
func NumberFormats() []NumberFormat {
	return []NumberFormat{
		{Group: SeparatorSpace, Decimal: SeparatorComma},
		{Group: SeparatorSpace, Decimal: SeparatorDot},
		{Group: SeparatorComma, Decimal: SeparatorDot},
		{Group: SeparatorDot, Decimal: SeparatorComma},
		{Group: SeparatorApostrophe, Decimal: SeparatorDot},
		{Group: SeparatorNone, Decimal: SeparatorComma},
		{Group: SeparatorNone, Decimal: SeparatorDot},
	}
}
