package ui

import (
	"cmp"
	"context"
	"net/http"

	"github.com/arkadiuss/openfief/internal/kernel"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// formatKey is the request context key of the number format a session
// read.
type formatKey struct{}

// withFormat returns r carrying f, the number format every template the
// request renders writes amounts in.
func withFormat(r *http.Request, f numberFormat) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), formatKey{}, f))
}

// requestFormat returns the number format r carries, or plainFormat() for a
// request without a session.
func requestFormat(r *http.Request) numberFormat {
	if f, ok := r.Context().Value(formatKey{}).(numberFormat); ok {
		return f
	}
	return plainFormat()
}

// valueKind is one of the three values of D-77 as the page's switches
// offer it: its query value, its label and the read API's enum.
type valueKind struct {
	Key   string
	Label string
	Enum  viewsv1.ValueKind
}

// valueKinds are the three values in the order the switches show them;
// the first, gross, is the default (D-77).
func valueKinds() []valueKind {
	return []valueKind{
		{Key: "gross", Label: "Gross", Enum: viewsv1.ValueKind_VALUE_KIND_GROSS},
		{Key: "net", Label: "Net", Enum: viewsv1.ValueKind_VALUE_KIND_NET},
		{Key: "liquidation", Label: "Liquidation", Enum: viewsv1.ValueKind_VALUE_KIND_LIQUIDATION},
	}
}

// parseValueKind returns the value kind a query value names, and gross for
// an empty or unknown one.
func parseValueKind(key string) valueKind {
	all := valueKinds()
	for _, k := range all {
		if k.Key == key {
			return k
		}
	}
	return all[0]
}

// totalOf returns a date total's value of kind k, or nil when the total
// lacks it.
func totalOf(t *viewsv1.DateTotal, k valueKind) *viewsv1.Money {
	switch k.Enum {
	case viewsv1.ValueKind_VALUE_KIND_NET:
		return t.GetNetReporting()
	case viewsv1.ValueKind_VALUE_KIND_LIQUIDATION:
		return t.GetLiquidationReporting()
	case viewsv1.ValueKind_VALUE_KIND_GROSS, viewsv1.ValueKind_VALUE_KIND_UNSPECIFIED:
	}
	return t.GetValueReporting()
}

// missingLabel says why a value is missing, from the read API's reason;
// an unknown reason is shown as sent.
func missingLabel(reason string) string {
	labels := map[string]string{
		"fx_rate":       "no FX rate yet",
		"lots":          "its lots are not known",
		"marginal_rate": "needs your marginal PIT rate",
		"price":         "no price yet",
		"index_reading": "no index reading yet",
	}
	if l, ok := labels[reason]; ok {
		return l
	}
	return reason
}

// wrapperLabel names a tax wrapper key of the jurisdiction pack (D-76): the
// Polish ones by their usual names, any other as its key.
func wrapperLabel(key string) string {
	labels := map[string]string{"taxable": "taxable account", "ike": "IKE", "ikze": "IKZE", "oki": "OKI"}
	if l, ok := labels[key]; ok {
		return l
	}
	if key == "" {
		return "no tax wrapper"
	}
	return key
}

// isZero reports whether a decimal string is empty or zero.
func isZero(s string) bool {
	d, err := kernel.ParseDecimal(s)
	return err != nil || d.IsZero()
}

// feeRule describes an account's sale fee rule (D-78) with its minimum in
// the format: "0.29%, at least 5,00 PLN, plus 0.5% to convert a foreign
// currency", or that it is zero, as it is until the owner sets it.
func (f numberFormat) feeRule(r *viewsv1.FeeRule) string {
	if isZero(r.GetRate()) && isZero(r.GetMinimum().GetAmount()) && isZero(r.GetConversionRate()) {
		return "none set, so sales cost nothing"
	}
	out := cmp.Or(percent(r.GetRate()), "0%")
	if !isZero(r.GetMinimum().GetAmount()) {
		out += ", at least " + f.money(r.GetMinimum())
	}
	if !isZero(r.GetConversionRate()) {
		out += ", plus " + percent(r.GetConversionRate()) + " to convert a foreign currency"
	}
	return out
}
