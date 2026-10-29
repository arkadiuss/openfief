package ui

import (
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

const (
	// groupSeparator separates thousands: a narrow no-break space, so an
	// amount never wraps and reads the same in any locale.
	groupSeparator = " "
	// groupDigits is how many digits a thousands group has.
	groupDigits = 3
	// sharePlaces is how many decimal places a share in percent shows.
	sharePlaces = 1
	// hundred turns a fraction into percent.
	hundred = 100
)

// formatAmount groups an exact decimal amount's integer digits in
// thousands, keeping its sign and every decimal place: "-1234567.80"
// becomes "-1 234 567.80". Anything that isn't a decimal is returned as is,
// so the page shows what the read API sent rather than nothing.
func formatAmount(amount string) string {
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
			b.WriteString(groupSeparator)
		}
		b.WriteRune(r)
	}
	if hasFrac {
		b.WriteString("." + frac)
	}
	return b.String()
}

// formatMoney formats a Money as its grouped amount and currency code; an
// unset Money is empty.
func formatMoney(m *viewsv1.Money) string {
	if m == nil {
		return ""
	}
	return formatAmount(m.GetAmount()) + " " + m.GetCurrency()
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

// valuationState describes an account's valuation status (D-36) in words:
// how far it is valued, and whether a revaluation is running or pending.
func valuationState(s *viewsv1.ValuationStatus) string {
	state := "not valued yet"
	if through := s.GetValuedThrough(); through != "" {
		state = "valued through " + through
	}
	switch {
	case s.GetRunning():
		state += "; revaluation running"
	case s.GetPending():
		state += "; revaluation pending"
	}
	return state
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
