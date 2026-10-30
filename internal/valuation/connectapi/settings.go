package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// errUnknownFormat reports a number format key that is not one of the
// formats the owner picks from.
var errUnknownFormat = errors.New("not one of the number formats offered")

// NewSettings returns the SettingsService's path and HTTP handler over
// settings (D-78, D-79, D-113), with callers in front of every call to
// admit view plugins only, logging to log.
func NewSettings(settings *app.Settings, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return valuationv1connect.NewSettingsServiceHandler(&settingsService{settings: settings, log: log},
		connect.WithInterceptors(callers))
}

// settingsService implements valuationv1connect.SettingsServiceHandler.
type settingsService struct {
	settings *app.Settings
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
	return connect.NewResponse(out), nil
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

// fail maps err to its Connect error, as the MarketDataService does.
func (s *settingsService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "settings call failed", slog.String("error", err.Error()))
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
