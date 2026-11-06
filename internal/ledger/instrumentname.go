package ledger

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MaxInstrumentNameRunes caps an instrument's display name, as a ticker's
// name is capped (D-116).
const MaxInstrumentNameRunes = 200

// ErrInvalidInstrumentName reports a display name that is empty, too long
// or holds a control character.
var ErrInvalidInstrumentName = errors.New("invalid instrument name")

// InstrumentName is one display name of an instrument (D-197), as it was
// recorded: one a connector delivered, such as XTB's display name, IBKR's
// description or the market-data lookup's, or one the owner gave it.
// Source names where a delivered name came from, such as the connector
// instance; it is empty for the owner's.
type InstrumentName struct {
	ISIN    ISIN
	Name    string
	ByOwner bool
	Source  string
}

// CleanInstrumentName returns s with its surrounding space trimmed and
// every run of white space inside it written as one space. It fails with
// ErrInvalidInstrumentName when nothing is left, when it is longer than
// MaxInstrumentNameRunes, or when it holds a control character other than
// white space.
func CleanInstrumentName(s string) (string, error) {
	out := strings.Join(strings.Fields(s), " ")
	switch {
	case out == "":
		return "", fmt.Errorf("%w: empty", ErrInvalidInstrumentName)
	case utf8.RuneCountInString(out) > MaxInstrumentNameRunes:
		return "", fmt.Errorf("%w: longer than %d characters", ErrInvalidInstrumentName, MaxInstrumentNameRunes)
	case !utf8.ValidString(out) || strings.ContainsFunc(out, unicode.IsControl):
		return "", fmt.Errorf("%w: control characters", ErrInvalidInstrumentName)
	}
	return out, nil
}

// CurrentInstrumentNames returns the display name each instrument shows
// (D-197), given every name recorded, in the order they were recorded: the
// owner's latest, or else the first a connector delivered. An instrument
// with no name recorded has none.
func CurrentInstrumentNames(records []InstrumentName) map[ISIN]string {
	out := map[ISIN]string{}
	owned := map[ISIN]bool{}
	for _, r := range records {
		switch {
		case r.ByOwner:
			out[r.ISIN], owned[r.ISIN] = r.Name, true
		case !owned[r.ISIN]:
			if _, named := out[r.ISIN]; !named {
				out[r.ISIN] = r.Name
			}
		}
	}
	return out
}

// FirstDeliveredNames returns, of names a connector delivered, the first
// valid one for each instrument, cleaned (CleanInstrumentName), in the order
// of names: what the ledger keeps when no name of the instrument is
// recorded yet. A name that isn't valid is left out rather than failing a
// delivery.
func FirstDeliveredNames(names []InstrumentName) []InstrumentName {
	seen := map[ISIN]bool{}
	var out []InstrumentName
	for _, n := range names {
		clean, err := CleanInstrumentName(n.Name)
		if err != nil || n.ISIN.IsZero() || seen[n.ISIN] {
			continue
		}
		seen[n.ISIN] = true
		out = append(out, InstrumentName{ISIN: n.ISIN, Name: clean, Source: n.Source})
	}
	return out
}
