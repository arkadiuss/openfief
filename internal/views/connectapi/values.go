package connectapi

import (
	"context"
	"fmt"
	"math"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/views"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// firstYear and lastYear bound the year ListYearlyTaxes accepts: the
// calendar the dates of the record can hold.
const (
	firstYear = 1
	lastYear  = 9999
)

// GetSettings answers GET /v1/settings.
func (s *readService) GetSettings(ctx context.Context, _ *connect.Request[viewsv1.GetSettingsRequest]) (*connect.Response[viewsv1.GetSettingsResponse], error) {
	got, err := s.views.Settings(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.GetSettingsResponse{
		ReportingCurrency: got.ReportingCurrency.String(), NumberFormat: pbNumberFormat(got.NumberFormat),
		NumberFormatDefault: got.NumberFormatDefault, TaxWrappers: make([]string, 0, len(got.Wrappers)),
	}
	for _, w := range got.Wrappers {
		out.TaxWrappers = append(out.TaxWrappers, string(w))
	}
	return connect.NewResponse(out), nil
}

// ListYearlyTaxes answers GET /v1/yearly-taxes.
func (s *readService) ListYearlyTaxes(ctx context.Context, req *connect.Request[viewsv1.ListYearlyTaxesRequest]) (*connect.Response[viewsv1.ListYearlyTaxesResponse], error) {
	year := int(req.Msg.GetYear())
	if year < firstYear || year > lastYear {
		return nil, toConnectError(ctx, s.log, fmt.Errorf("%w: year %d", errBadParameter, year))
	}
	taxes, err := s.views.YearlyTaxes(ctx, year)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListYearlyTaxesResponse{Taxes: make([]*viewsv1.YearlyTax, 0, len(taxes))}
	for i := range taxes {
		out.Taxes = append(out.Taxes, pbYearlyTax(&taxes[i]))
	}
	return connect.NewResponse(out), nil
}

// pbNumberFormat converts a number format.
func pbNumberFormat(f jurisdiction.NumberFormat) *viewsv1.NumberFormat {
	return &viewsv1.NumberFormat{Key: f.String(), GroupSeparator: string(f.Group), DecimalSeparator: string(f.Decimal)}
}

// pbYearlyTax converts a yearly tax; one whose law isn't in force carries
// only its wrapper and year.
func pbYearlyTax(t *jurisdiction.YearlyTax) *viewsv1.YearlyTax {
	out := &viewsv1.YearlyTax{TaxWrapper: string(t.Wrapper), Year: clampInt32(t.Year), InForce: t.InForce}
	if !t.InForce {
		return out
	}
	out.Days, out.Average, out.Limit = clampInt32(t.Days), pbMoney(t.Average), pbMoney(t.Limit)
	out.Rate, out.Excess, out.Tax = t.Rate.String(), pbMoney(t.Excess), pbMoney(t.Tax)
	return out
}

// clampInt32 returns n as an int32, clamped to its range.
func clampInt32(n int) int32 {
	switch {
	case n > math.MaxInt32:
		return math.MaxInt32
	case n < math.MinInt32:
		return math.MinInt32
	}
	return int32(n)
}

// valueKind reads a request's value kind: unspecified is gross.
func valueKind(k viewsv1.ValueKind) (views.ValueKind, error) {
	switch k {
	case viewsv1.ValueKind_VALUE_KIND_UNSPECIFIED, viewsv1.ValueKind_VALUE_KIND_GROSS:
		return views.Gross, nil
	case viewsv1.ValueKind_VALUE_KIND_NET:
		return views.Net, nil
	case viewsv1.ValueKind_VALUE_KIND_LIQUIDATION:
		return views.Liquidation, nil
	}
	return 0, fmt.Errorf("%w: value kind %d", errBadParameter, k)
}
