package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// errUnknownFormat reports a number format key that is not one of the
// formats the owner picks from.
var errUnknownFormat = errors.New("not one of the number formats offered")

// NewSettings returns the SettingsService's path and HTTP handler over
// settings and the charts' start setting starts (D-78, D-79, D-113,
// D-167), with callers in front of every call to admit view plugins only,
// logging to log.
func NewSettings(settings *app.Settings, starts *app.ChartStarts, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return valuationv1connect.NewSettingsServiceHandler(&settingsService{settings: settings, starts: starts, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// settingsService implements valuationv1connect.SettingsServiceHandler.
type settingsService struct {
	settings *app.Settings
	starts   *app.ChartStarts
	log      *slog.Logger
}

// GetOwnerSettings returns every setting to the owner.
func (s *settingsService) GetOwnerSettings(ctx context.Context, req *connect.Request[valuationv1.GetOwnerSettingsRequest]) (*connect.Response[valuationv1.GetOwnerSettingsResponse], error) {
	got, err := s.settings.Owner(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &valuationv1.GetOwnerSettingsResponse{
		ReportingCurrency: got.ReportingCurrency.String(), NumberFormat: pbNumberFormat(got.NumberFormat),
		NumberFormatDefault: got.NumberFormatDefault,
	}
	for _, f := range jurisdiction.NumberFormats() {
		out.NumberFormats = append(out.NumberFormats, pbNumberFormat(f))
	}
	for _, w := range got.Wrappers {
		out.TaxWrappers = append(out.TaxWrappers, string(w))
	}
	if got.Values.HasMarginalRate {
		out.MarginalRate = got.Values.MarginalRate.String()
	}
	for _, c := range got.ExemptionClasses {
		out.ExemptionClasses = append(out.ExemptionClasses, string(c))
	}
	for _, isin := range slices.Sorted(maps.Keys(got.InstrumentClasses)) {
		out.InstrumentClasses = append(out.InstrumentClasses,
			&valuationv1.InstrumentClass{Isin: isin, ExemptionClass: string(got.InstrumentClasses[isin])})
	}
	return connect.NewResponse(out), nil
}

// SetEarlyRedemptionFee confirms or overrides a treasury bond series'
// early-redemption fee (D-97).
func (s *settingsService) SetEarlyRedemptionFee(ctx context.Context, req *connect.Request[valuationv1.SetEarlyRedemptionFeeRequest]) (*connect.Response[valuationv1.SetEarlyRedemptionFeeResponse], error) {
	err := s.settings.SetEarlyRedemptionFee(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetIsin(), req.Msg.GetFee())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetEarlyRedemptionFeeResponse{}), nil
}

// SetExemptionClass classifies an instrument for a yearly tax (D-132).
func (s *settingsService) SetExemptionClass(ctx context.Context, req *connect.Request[valuationv1.SetExemptionClassRequest]) (*connect.Response[valuationv1.SetExemptionClassResponse], error) {
	err := s.settings.SetExemptionClass(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetIsin(),
		jurisdiction.ExemptionClass(req.Msg.GetExemptionClass()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetExemptionClassResponse{}), nil
}

// SetNumberFormat makes one of the offered number formats the owner's.
func (s *settingsService) SetNumberFormat(ctx context.Context, req *connect.Request[valuationv1.SetNumberFormatRequest]) (*connect.Response[valuationv1.SetNumberFormatResponse], error) {
	f, err := offeredFormat(req.Msg.GetKey())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	if err := s.settings.SetNumberFormat(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), f); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetNumberFormatResponse{NumberFormat: pbNumberFormat(f)}), nil
}

// SetMarginalRate sets the owner's marginal income tax rate.
func (s *settingsService) SetMarginalRate(ctx context.Context, req *connect.Request[valuationv1.SetMarginalRateRequest]) (*connect.Response[valuationv1.SetMarginalRateResponse], error) {
	rate, err := kernel.ParseDecimal(req.Msg.GetRate())
	if err != nil {
		return nil, s.fail(ctx, fmt.Errorf("%w: marginal rate %q: %w", app.ErrInvalidInput, req.Msg.GetRate(), err))
	}
	if err := s.settings.SetMarginalRate(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), rate); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetMarginalRateResponse{}), nil
}

// SetChartStart stores the charts' start the owner chose, the portfolio's
// or an account's, or clears it for an empty start (D-167).
func (s *settingsService) SetChartStart(ctx context.Context, req *connect.Request[valuationv1.SetChartStartRequest]) (*connect.Response[valuationv1.SetChartStartResponse], error) {
	var start kernel.Date
	if text := req.Msg.GetStart(); text != "" {
		var err error
		if start, err = kernel.ParseDate(text); err != nil {
			return nil, s.fail(ctx, fmt.Errorf("%w: charts' start %q: %w", app.ErrInvalidInput, text, err))
		}
	}
	err := s.starts.Set(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), ledger.AccountID(req.Msg.GetAccountId()), start)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetChartStartResponse{}), nil
}

// fail maps err to its Connect error, as the MarketDataService does.
func (s *settingsService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrUnknownAccount):
		return connect.NewError(connect.CodeNotFound, app.ErrUnknownAccount)
	case errors.Is(err, app.ErrUnknownBond):
		return connect.NewError(connect.CodeNotFound, app.ErrUnknownBond)
	case errors.Is(err, pl.ErrFeeNotAssumed):
		return connect.NewError(connect.CodeFailedPrecondition, pl.ErrFeeNotAssumed)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "settings call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// offeredFormat returns the offered number format with key, or an error
// wrapping app.ErrInvalidInput.
func offeredFormat(key string) (jurisdiction.NumberFormat, error) {
	for _, f := range jurisdiction.NumberFormats() {
		if f.String() == key {
			return f, nil
		}
	}
	return jurisdiction.NumberFormat{}, fmt.Errorf("%w: %w: %q", app.ErrInvalidInput, errUnknownFormat, key)
}

// pbNumberFormat converts a number format to its message.
func pbNumberFormat(f jurisdiction.NumberFormat) *valuationv1.NumberFormat {
	return &valuationv1.NumberFormat{Key: f.String(), GroupSeparator: string(f.Group), DecimalSeparator: string(f.Decimal)}
}
