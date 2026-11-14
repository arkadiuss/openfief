package main

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// A number the model wrote, read in every way its separators can be read: with
// a comma as the decimal separator ("1 234,56", Polish) or a point ("1,234.56",
// English). Each reading carries how far a rounding could have moved it.
type reading struct {
	value kernel.Decimal
	// tol is the width of one unit of the last digit written, times the scale
	// ("k", "mln") it was written with: a rounded or cut number is within it.
	tol kernel.Decimal
}

// foundNumber is a number of a text and its readings.
type foundNumber struct {
	text     string
	readings []reading
}

// numberPattern finds a number: digits grouped by spaces (a space, a no-break
// space or a narrow no-break space), then digits grouped by points or commas,
// then a plain one, each with an optional decimal part.
func numberPattern() *regexp.Regexp {
	return regexp.MustCompile(`\d{1,3}(?:[ \x{00a0}\x{202f}]\d{3})+(?:[.,]\d+)?|\d+(?:[.,]\d{3})+(?:[.,]\d+)?|\d+(?:[.,]\d+)?`)
}

// isoDatePattern matches a whole ISO date.
func isoDatePattern() *regexp.Regexp { return regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`) }

// dateTexts are the dates a text can hold, which are not amounts: 2026-10-28
// and 28.10.2026.
func dateTexts() *regexp.Regexp {
	return regexp.MustCompile(`\d{4}-\d{2}-\d{2}|\d{1,2}[./]\d{1,2}[./]\d{4}`)
}

// The powers of ten a scale word gives.
const (
	thousands int32 = 3
	millions  int32 = 6
	billions  int32 = 9
	// maxDayDigits is how many digits a day of the month or a count has, at most,
	// to be used without a card; yearDigits is a year's.
	maxDayDigits = 2
	yearDigits   = 4
)

// scaleWords are the words that scale a number, as the model writes them after
// it: "1.2 mln", "123k", "3 million".
func scaleWords() map[string]int32 {
	return map[string]int32{
		"k": thousands, "tys": thousands, "thousand": thousands, "tysiące": thousands, "tysięcy": thousands,
		"m": millions, "mln": millions, "million": millions, "millions": millions, "miliony": millions,
		"bn": billions, "billion": billions, "mld": billions, "miliard": billions, "miliardy": billions, "miliardów": billions,
	}
}

// numbersIn returns the numbers of text that an amount, a percentage or a
// count can be: not the digits of a date, of an ISIN or a ticker (a number
// touching a letter), nor a whole number of a day of the month or a year.
func numbersIn(text string) []foundNumber {
	text = dateTexts().ReplaceAllString(text, " ")
	re := numberPattern()
	var out []foundNumber
	for _, loc := range re.FindAllStringIndex(text, -1) {
		start, end := loc[0], loc[1]
		if touchesLetter(text, start, end) {
			continue
		}
		raw := text[start:end]
		scale := scaleAfter(text[end:])
		plain := !strings.ContainsAny(raw, ".,") && scale == 0 && !percentAfter(text[end:])
		if plain && exemptWhole(raw) {
			continue
		}
		out = append(out, foundNumber{text: raw, readings: readingsOf(raw, scale)})
	}
	return out
}

// touchesLetter reports whether the match text[start:end] is glued to a letter
// or a digit on either side, as in "IE00B3RBWM25" or "DWLD2".
func touchesLetter(text string, start, end int) bool {
	if r, _ := utf8.DecodeLastRuneInString(text[:start]); start > 0 && (unicode.IsLetter(r) || r == '_') {
		return true
	}
	r, _ := utf8.DecodeRuneInString(text[end:])
	return end < len(text) && unicode.IsLetter(r) && !isScaleStart(text[end:])
}

// isScaleStart reports whether rest begins with a scale word, such as "k" in
// "123k" or "mln" in "1,2mln".
func isScaleStart(rest string) bool {
	return scaleAfter(rest) != 0
}

// scaleAfter returns the power of ten the scale word at the start of rest
// (after optional spaces) gives, or 0.
func scaleAfter(rest string) int32 {
	rest = strings.TrimLeft(rest, "   ")
	word := rest
	if i := strings.IndexFunc(rest, func(r rune) bool { return !unicode.IsLetter(r) && r != '.' }); i >= 0 {
		word = rest[:i]
	}
	return scaleWords()[strings.ToLower(strings.TrimRight(word, "."))]
}

// percentAfter reports whether a percent sign follows, after optional spaces.
func percentAfter(rest string) bool {
	return strings.HasPrefix(strings.TrimLeft(rest, "   "), "%")
}

// exemptWhole reports whether a plain whole number is one the model may use
// without a card: a day of the month, a small count, or a year.
func exemptWhole(raw string) bool {
	if len(raw) <= maxDayDigits {
		return true
	}
	return len(raw) == yearDigits && (strings.HasPrefix(raw, "19") || strings.HasPrefix(raw, "20"))
}

// readingsOf reads raw in each way its separators allow, scaled by 10^scale.
func readingsOf(raw string, scale int32) []reading {
	var out []reading
	factor := kernel.NewDecimal(1, scale)
	for _, norm := range separatorReadings(raw) {
		d, err := kernel.ParseDecimal(norm)
		if err != nil {
			continue
		}
		unit := kernel.NewDecimal(1, 0)
		if _, frac, ok := strings.Cut(norm, "."); ok && frac != "" {
			// One unit of the last digit written: 0.01 for two places.
			if u, uerr := kernel.ParseDecimal("0." + strings.Repeat("0", len(frac)-1) + "1"); uerr == nil {
				unit = u
			}
		}
		out = append(out, reading{value: d.Mul(factor), tol: unit.Mul(factor)})
	}
	return out
}

// separatorReadings returns raw as plain decimal texts, one for each way its
// separators can be read: a lone comma or point followed by three digits is
// either a group separator or a decimal one.
func separatorReadings(raw string) []string {
	clean := strings.NewReplacer(" ", "", " ", "", " ", "").Replace(raw)
	commas, points := strings.Count(clean, ","), strings.Count(clean, ".")
	switch {
	case commas > 0 && points > 0:
		last := max(strings.LastIndex(clean, ","), strings.LastIndex(clean, "."))
		return []string{strings.NewReplacer(",", "", ".", "").Replace(clean[:last]) + "." + clean[last+1:]}
	case commas+points == 0:
		return []string{clean}
	}
	sep := ","
	if points > 0 {
		sep = "."
	}
	if commas+points > 1 {
		return []string{strings.ReplaceAll(clean, sep, "")}
	}
	whole, frac, _ := strings.Cut(clean, sep)
	decimal := whole + "." + frac
	if len(frac) == 3 && len(whole) <= 3 {
		return []string{decimal, whole + frac}
	}
	return []string{decimal}
}

// grounded reports whether n is one of the card numbers within rounding, by
// size: a sign is ignored, since "down 5.2%" is a card's "-5.2%".
func grounded(n *foundNumber, card []kernel.Decimal) bool {
	for _, r := range n.readings {
		for _, c := range card {
			diff := absDecimal(c.Sub(r.value))
			if diff.Cmp(r.tol) <= 0 {
				return true
			}
		}
	}
	return false
}

// absDecimal returns |d|.
func absDecimal(d kernel.Decimal) kernel.Decimal {
	if d.Sign() < 0 {
		return d.Neg()
	}
	return d
}

// cardNumbers returns, by size, every reading of every number of the texts.
func cardNumbers(texts []string) []kernel.Decimal {
	var out []kernel.Decimal
	for _, t := range texts {
		// A card's numbers are exact, so they are read without exempting the
		// small ones: a model may quote a "5" a card holds.
		for _, loc := range numberPattern().FindAllString(dateTexts().ReplaceAllString(t, " "), -1) {
			for _, norm := range separatorReadings(loc) {
				if d, err := kernel.ParseDecimal(norm); err == nil {
					out = append(out, absDecimal(d))
				}
			}
		}
	}
	return out
}
