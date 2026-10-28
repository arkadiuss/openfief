package ibkrflex

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// errField reports an attribute that is missing or unreadable. Rows with one
// become ingest problems; the attribute name is kept, the value is not.
var errField = errors.New("missing or unreadable attribute")

// The lengths of IBKR's two date formats.
const (
	compactDateLen = len("20060102")
	isoDateLen     = len("2006-01-02")
	// maxDecimalDigits is far more than any amount or quantity needs.
	maxDecimalDigits = 40
)

// text returns attribute name of r, or "" when it is absent.
func (r *row) text(name string) string { return r.attrs[name] }

// decimal parses attribute name of r as an exact decimal written in plain
// notation, such as "-1234.5". Exponents are refused: "1e999999999" would
// otherwise be a tiny attribute that expands to a huge number.
func (r *row) decimal(name string) (kernel.Decimal, error) {
	s := r.text(name)
	if !plainDecimal(s) {
		return kernel.Decimal{}, fmt.Errorf("%w: %s", errField, name)
	}
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		return kernel.Decimal{}, fmt.Errorf("%w: %s", errField, name)
	}
	return d, nil
}

// plainDecimal reports whether s is an optional minus sign, at most
// maxDecimalDigits digits and at most one decimal point between digits.
func plainDecimal(s string) bool {
	s = strings.TrimPrefix(s, "-")
	whole, frac, hasPoint := strings.Cut(s, ".")
	if whole == "" || (hasPoint && frac == "") || len(whole)+len(frac) > maxDecimalDigits {
		return false
	}
	for _, part := range []string{whole, frac} {
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// currency parses attribute name of r as an ISO 4217 code.
func (r *row) currency(name string) (kernel.Currency, error) {
	c, err := kernel.ParseCurrency(r.text(name))
	if err != nil {
		return kernel.Currency{}, fmt.Errorf("%w: %s", errField, name)
	}
	return c, nil
}

// money parses attribute amount of r in the row's currency.
func (r *row) money(amount string) (kernel.Money, error) {
	c, err := r.currency("currency")
	if err != nil {
		return kernel.Money{}, err
	}
	d, err := r.decimal(amount)
	if err != nil {
		return kernel.Money{}, err
	}
	return kernel.NewMoney(d, c), nil
}

// date parses the first of names that r has as a calendar date.
func (r *row) date(names ...string) (kernel.Date, error) {
	for _, name := range names {
		if s := r.text(name); s != "" {
			d, err := parseDate(s)
			if err != nil {
				return kernel.Date{}, fmt.Errorf("%w: %s", errField, name)
			}
			return d, nil
		}
	}
	return kernel.Date{}, fmt.Errorf("%w: %s", errField, strings.Join(names, " or "))
}

// instrument returns r's instrument from its isin and symbol attributes.
func (r *row) instrument() (ledger.Instrument, error) {
	isin, err := ledger.ParseISIN(r.text("isin"))
	if err != nil {
		return ledger.Instrument{}, fmt.Errorf("%w: isin", errField)
	}
	in := ledger.Instrument{ISIN: isin, Symbol: r.text("symbol")}
	if in.Symbol == "" {
		return ledger.Instrument{}, fmt.Errorf("%w: symbol", errField)
	}
	return in, nil
}

// parseDate parses an IBKR date, yyyyMMdd or yyyy-MM-dd, optionally followed
// by a separator (";", ",", " " or "T") and a time of day, which is dropped:
// the calendar date is taken as IBKR reports it, with no time zone
// conversion.
func parseDate(s string) (kernel.Date, error) {
	n := compactDateLen
	if len(s) > compactDateLen && s[4] == '-' {
		n = isoDateLen
	}
	if len(s) < n {
		return kernel.Date{}, fmt.Errorf("%w: date %q", errField, s)
	}
	day, rest := s[:n], s[n:]
	if rest != "" && !strings.ContainsRune(";, T", rune(rest[0])) {
		return kernel.Date{}, fmt.Errorf("%w: date %q", errField, s)
	}
	if n == compactDateLen {
		day = day[:4] + "-" + day[4:6] + "-" + day[6:]
	}
	d, err := kernel.ParseDate(day)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("%w: date %q", errField, s)
	}
	return d, nil
}
