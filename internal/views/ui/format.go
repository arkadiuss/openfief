package ui

import (
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

const (
	// narrowSpace is how a space group separator is written: a narrow
	// no-break space, so an amount never wraps.
	narrowSpace = "\u202f"
	// groupDigits is how many digits a thousands group has.
	groupDigits = 3
	// sharePlaces is how many decimal places a share in percent shows.
	sharePlaces = 1
	// hundred turns a fraction into percent.
	hundred = 100
)

// numberFormat is how the page writes amounts (D-113): only the separator
// between groups of three integer digits and the decimal separator. It
// changes nothing else: not the digits, the sign or the currency.
type numberFormat struct {
	// Group separates thousands as written: a space format's separator is
	// the narrow no-break space; empty for none.
	Group string
	// Decimal is the decimal separator.
	Decimal string
}

// plainFormat writes amounts as the read API sends them, grouped with a
// narrow space. Only pages without a session, which show no amount, use it.
func plainFormat() numberFormat { return numberFormat{Group: narrowSpace, Decimal: "."} }

// formatOf returns the number format the read API's settings name, a
// space written as the narrow no-break space; a format without a decimal
// separator is the plain format.
func formatOf(f *viewsv1.NumberFormat) numberFormat {
	return separators(f.GetGroupSeparator(), f.GetDecimalSeparator())
}

// separators returns the number format with group and decimal as the
// settings name them, a space group written as the narrow no-break space;
// no decimal separator is the plain format.
func separators(group, decimal string) numberFormat {
	if decimal == "" {
		return plainFormat()
	}
	if group == " " {
		group = narrowSpace
	}
	return numberFormat{Group: group, Decimal: decimal}
}

// amount writes an exact decimal amount in the format: its integer digits
// grouped in thousands, its sign and every decimal place kept, so with a
// space and a decimal comma "-1234567.80" becomes "-1 234 567,80".
// Anything that isn't a decimal is returned as is, so the page shows what
// the read API sent rather than nothing.
func (f numberFormat) amount(amount string) string {
	if _, err := kernel.ParseDecimal(amount); err != nil {
		return amount
	}
	sign, digits := "", amount
	if strings.HasPrefix(digits, "-") || strings.HasPrefix(digits, "+") {
		sign, digits = digits[:1], digits[1:]
	}
	if sign == "+" {
		sign = ""
	}
	whole, frac, hasFrac := strings.Cut(digits, ".")
	var b strings.Builder
	b.WriteString(sign)
	for i, r := range whole {
		if i > 0 && (len(whole)-i)%groupDigits == 0 {
			b.WriteString(f.Group)
		}
		b.WriteRune(r)
	}
	if hasFrac {
		b.WriteString(f.Decimal + frac)
	}
	return b.String()
}

// money writes a Money as its amount in the format and its currency code;
// an unset Money is empty.
func (f numberFormat) money(m *viewsv1.Money) string {
	if m == nil {
		return ""
	}
	return f.amount(m.GetAmount()) + " " + m.GetCurrency()
}

// balance writes a balance of a ledger account, such as a balance
// assertion's reported or derived balance, with its unit: a holding's is a
// quantity of its instrument, any other's an amount in the ledger
// account's currency, which is always shown (D-20).
func (f numberFormat) balance(b *viewsv1.Book, balance string) string {
	if b.GetKind() == viewsv1.BookKind_BOOK_KIND_HOLDING {
		return f.amount(balance)
	}
	return f.amount(balance) + " " + b.GetCurrency()
}

// share returns part as a percentage of total, rounded half-even to one
// decimal place, such as "42.5%". It is empty when either is not a decimal
// or total is zero: a share of nothing is not shown.
func share(part, total string) string {
	p, err := kernel.ParseDecimal(part)
	if err != nil {
		return ""
	}
	t, err := kernel.ParseDecimal(total)
	if err != nil || t.IsZero() {
		return ""
	}
	q, err := p.Mul(kernel.NewDecimal(hundred, 0)).Quo(t, sharePlaces, kernel.RoundHalfEven)
	if err != nil {
		return ""
	}
	return q.String() + "%"
}

// dimension is one way to split the allocation (D-30): its query value,
// its label and the read API's enum.
type dimension struct {
	Key   string
	Label string
	Enum  viewsv1.AllocationDimension
}

// dimensions are the allocation's dimensions in the order the switch shows
// them; the first is the default.
func dimensions() []dimension {
	return []dimension{
		{Key: "asset-class", Label: "Asset class", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS},
		{Key: "currency", Label: "Currency", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY},
		{Key: "institution", Label: "Institution", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_INSTITUTION},
		{Key: "sleeve", Label: "Sleeve", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_SLEEVE},
	}
}

// parseDimension returns the dimension a query value names, and the default
// (asset class) for an empty or unknown one.
func parseDimension(key string) dimension {
	all := dimensions()
	for _, d := range all {
		if d.Key == key {
			return d
		}
	}
	return all[0]
}

// bookLabel names a ledger account: a holding by its instrument, any other
// kind by its kind and currency.
func bookLabel(b *viewsv1.Book) string {
	if b.GetKind() == viewsv1.BookKind_BOOK_KIND_HOLDING {
		return instrumentLabel(b.GetInstrument())
	}
	return enumLabel(b.GetKind().String(), "BOOK_KIND_") + " " + b.GetCurrency()
}

// instrumentLabel names an instrument by its venue symbol and ISIN.
func instrumentLabel(i *viewsv1.Instrument) string {
	return i.GetSymbol() + " (" + i.GetIsin() + ")"
}

// enumLabel turns a protobuf enum value name into words: the prefix is
// dropped and underscores become spaces, so TRANSACTION_KIND_OPENING_LOT
// reads "opening lot".
func enumLabel(name, prefix string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(name, prefix), "_", " "))
}

// scopeLabel names a session scope for the page header.
func scopeLabel(s identityv1.SessionScope) string {
	return enumLabel(s.String(), "SESSION_SCOPE_")
}

// percent writes a decimal fraction, such as the read API's relative
// difference "0.028571" or threshold "0.01", as a percentage: "2.8571%",
// "1%". It is exact, with the fraction's own digits; text that isn't a
// decimal, or none, gives "".
func percent(fraction string) string {
	d, err := kernel.ParseDecimal(fraction)
	if err != nil {
		return ""
	}
	return d.Mul(kernel.NewDecimal(hundred, 0)).String() + "%"
}
