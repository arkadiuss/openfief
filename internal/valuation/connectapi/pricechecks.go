package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
)

// errNoInstrument reports a request without an ISIN or a symbol.
var errNoInstrument = errors.New("an instrument needs an ISIN and a symbol")

// NewPriceChecks returns the PriceCheckService's path and HTTP handler over
// checks (D-110, D-112, D-79), with callers in front of every call to admit
// view plugins only, logging to log.
func NewPriceChecks(checks *app.CrossChecks, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return valuationv1connect.NewPriceCheckServiceHandler(&priceChecks{checks: checks, log: log},
		connect.WithInterceptors(callers))
}

// priceChecks implements valuationv1connect.PriceCheckServiceHandler.
type priceChecks struct {
	checks *app.CrossChecks
	log    *slog.Logger
}

// SetPriceCheckThreshold sets or clears one instrument's threshold.
func (s *priceChecks) SetPriceCheckThreshold(ctx context.Context, req *connect.Request[valuationv1.SetPriceCheckThresholdRequest]) (*connect.Response[valuationv1.SetPriceCheckThresholdResponse], error) {
	instrument, err := instrumentOf(req.Msg.GetIsin(), req.Msg.GetSymbol())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	var threshold *kernel.Decimal
	if t := req.Msg.GetThreshold(); t != "" {
		d, parseErr := kernel.ParseDecimal(t)
		if parseErr != nil {
			return nil, s.fail(ctx, fmt.Errorf("%w: %w: %w", app.ErrInvalidInput, valuation.ErrInvalidPriceCheckThreshold, parseErr))
		}
		threshold = &d
	}
	if err := s.checks.SetThreshold(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), instrument, threshold); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.SetPriceCheckThresholdResponse{}), nil
}

// AcknowledgePriceCheck records the owner's acknowledgement of a flagged
// check.
func (s *priceChecks) AcknowledgePriceCheck(ctx context.Context, req *connect.Request[valuationv1.AcknowledgePriceCheckRequest]) (*connect.Response[valuationv1.AcknowledgePriceCheckResponse], error) {
	instrument, err := instrumentOf(req.Msg.GetIsin(), req.Msg.GetSymbol())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	day, err := kernel.ParseDate(req.Msg.GetDate())
	if err != nil {
		return nil, s.fail(ctx, fmt.Errorf("%w: date: %w", app.ErrInvalidInput, err))
	}
	if _, err := s.checks.Acknowledge(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), instrument, day, req.Msg.GetSource()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&valuationv1.AcknowledgePriceCheckResponse{}), nil
}

// instrumentOf writes an instrument as valuation keys it, ISIN/symbol;
// either missing is ErrInvalidInput.
func instrumentOf(isin, symbol string) (string, error) {
	if isin == "" || symbol == "" {
		return "", fmt.Errorf("%w: %w", app.ErrInvalidInput, errNoInstrument)
	}
	return isin + "/" + symbol, nil
}

// fail maps err to its Connect error: no session unauthenticated, not the
// owner permission_denied, a broken rule invalid_argument with its text, an
// unknown check not_found, anything else internal with nothing of it.
func (s *priceChecks) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrUnknownPriceCheck):
		return connect.NewError(connect.CodeNotFound, app.ErrUnknownPriceCheck)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "price check call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}
