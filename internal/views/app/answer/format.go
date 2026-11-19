package answer

import (
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

const (
	// NarrowSpace is how a space group separator is written: a narrow
	// no-break space, so an amount never wraps.
	NarrowSpace = " "
	// groupDigits is how many digits a thousands group has.
	groupDigits = 3
	// sharePlaces is how many decimal places a share in percent shows.
	sharePlaces = 1
	// hundred turns a fraction into percent.
	hundred = 100
)

// NumberFormat is how amounts are written (D-113): only the separator
// between groups of three integer digits and the decimal separator. It
// changes nothing else: not the digits, the sign or the currency.
type NumberFormat struct {
	// Group separates thousands as written: a space format's separator is
	// the narrow no-break space; empty for none.
	Group string
	// Decimal is the decimal separator.
	Decimal string
}

// PlainFormat writes amounts as the read API sends them, grouped with a
// narrow space.
func PlainFormat() NumberFormat { return NumberFormat{Group: NarrowSpace, Decimal: "."} }

// FormatOf returns the number format the read API's settings name, a space
// written as the narrow no-break space; a format without a decimal separator
// is the plain format.
func FormatOf(f *viewsv1.NumberFormat) NumberFormat {
	return Separators(f.GetGroupSeparator(), f.GetDecimalSeparator())
}

// Separators returns the number format with group and decimal as the
// settings name them, a space group written as the narrow no-break space; no
// decimal separator is the plain format.
func Separators(group, decimal string) NumberFormat {
	if decimal == "" {
		return PlainFormat()
	}
	if group == " " {
		group = NarrowSpace
	}
	return NumberFormat{Group: group, Decimal: decimal}
}

// Amount writes an exact decimal amount in the format: its integer digits
// grouped in thousands, its sign and every decimal place kept, so with a
// space and a decimal comma "-1234567.80" becomes "-1 234 567,80".
// Anything that isn't a decimal is returned as is, so the page shows what
// the read API sent rather than nothing.
func (f NumberFormat) Amount(amount string) string {
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

// Money writes a Money as its amount in the format and its currency code;
// an unset Money is empty.
func (f NumberFormat) Money(m *viewsv1.Money) string {
	if m == nil {
		return ""
	}
	return f.Amount(m.GetAmount()) + " " + m.GetCurrency()
}

// Signed writes a change of money with its sign, "+8 214,90 PLN" or
// "-3,10 PLN"; an unset Money is empty.
func (f NumberFormat) Signed(m *viewsv1.Money) string {
	out := f.Money(m)
	if out != "" && !strings.HasPrefix(out, "-") {
		return "+" + out
	}
	return out
}

// Share returns part as a percentage of total, rounded half-even to one
// decimal place, such as "42.5%". It is empty when either is not a decimal
// or total is zero: a share of nothing is not shown.
func Share(part, total string) string {
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

// SumMoney returns the sum of ms, every one in the same currency; nil for
// none, a sum that isn't a decimal, or mixed currencies.
func SumMoney(ms []*viewsv1.Money) *viewsv1.Money {
	if len(ms) == 0 {
		return nil
	}
	var sum kernel.Decimal
	for _, m := range ms {
		d, err := kernel.ParseDecimal(m.GetAmount())
		if err != nil || m.GetCurrency() != ms[0].GetCurrency() {
			return nil
		}
		sum = sum.Add(d)
	}
	return &viewsv1.Money{Amount: sum.String(), Currency: ms[0].GetCurrency()}
}

// InstrumentLabel names an instrument by its display name, when it has
// one, before its venue symbol and ISIN (D-197).
func InstrumentLabel(i *viewsv1.Instrument) string {
	label := i.GetSymbol() + " (" + i.GetIsin() + ")"
	if name := i.GetName(); name != "" {
		return name + " · " + label
	}
	return label
}

// EnumLabel turns a protobuf enum value name into words: the prefix is
// dropped and underscores become spaces, so TRANSACTION_KIND_OPENING_LOT
// reads "opening lot".
func EnumLabel(name, prefix string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(name, prefix), "_", " "))
}

// WrapperLabel names a tax wrapper key of the jurisdiction pack (D-76): the
// Polish ones by their usual names, any other as its key.
func WrapperLabel(key string) string {
	labels := map[string]string{"taxable": "taxable account", "ike": "IKE", "ikze": "IKZE", "oki": "OKI"}
	if l, ok := labels[key]; ok {
		return l
	}
	if key == "" {
		return "no tax wrapper"
	}
	return key
}

// LatestValued returns the latest date any account is valued through,
// otherwise today.
func LatestValued(statuses []*viewsv1.ValuationStatus, today kernel.Date) kernel.Date {
	var latest kernel.Date
	for _, s := range statuses {
		d, err := kernel.ParseDate(s.GetValuedThrough())
		if err == nil && d.After(latest) {
			latest = d
		}
	}
	if latest.IsZero() {
		return today
	}
	return latest
}

// UTCToday is the clock's date in UTC.
func UTCToday() kernel.Date {
	y, m, d := time.Now().UTC().Date()
	t, err := kernel.NewDate(y, m, d)
	if err != nil { // a calendar date of the clock is always valid
		return kernel.Date{}
	}
	return t
}

// OptionalString returns d as YYYY-MM-DD, or "" for the zero date.
func OptionalString(d kernel.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.String()
}

// Dimension is one way to split the allocation (D-30): its query value,
// its label and the read API's enum.
type Dimension struct {
	Key   string
	Label string
	Enum  viewsv1.AllocationDimension
}

// The allocation's dimension keys, as its switch's links name them.
const (
	DimAssetClass  = "asset-class"
	DimCurrency    = "currency"
	DimInstitution = "institution"
	DimSleeve      = "sleeve"
)

// Dimensions are the allocation's dimensions in the order the switch shows
// them; the first is the default.
func Dimensions() []Dimension {
	return []Dimension{
		{Key: DimAssetClass, Label: "Asset class", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_ASSET_CLASS},
		{Key: DimCurrency, Label: "Currency", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_CURRENCY},
		{Key: DimInstitution, Label: "Institution", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_INSTITUTION},
		{Key: DimSleeve, Label: "Sleeve", Enum: viewsv1.AllocationDimension_ALLOCATION_DIMENSION_SLEEVE},
	}
}

// ParseDimension returns the dimension a query value names, and the default
// (asset class) for an empty or unknown one.
func ParseDimension(key string) Dimension {
	all := Dimensions()
	for _, d := range all {
		if d.Key == key {
			return d
		}
	}
	return all[0]
}
