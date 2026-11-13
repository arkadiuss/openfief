package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// NewInstruments returns the InstrumentService's path and HTTP handler over
// names (D-197, D-79), with callers in front of every call to admit view
// plugins only, logging to log.
func NewInstruments(names *app.Names, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ledgerv1connect.NewInstrumentServiceHandler(&instrumentService{names: names, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// instrumentService implements ledgerv1connect.InstrumentServiceHandler.
type instrumentService struct {
	names *app.Names
	log   *slog.Logger
}

// RenameInstrument records the owner's display name of an instrument.
func (s *instrumentService) RenameInstrument(ctx context.Context, req *connect.Request[ledgerv1.RenameInstrumentRequest]) (*connect.Response[ledgerv1.RenameInstrumentResponse], error) {
	m := req.Msg
	isin, err := ledger.ParseISIN(m.GetIsin())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	name, err := s.names.Rename(ctx, m.GetOrigin(), m.GetSessionToken(), isin, m.GetName())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ledgerv1.RenameInstrumentResponse{Name: name}), nil
}

// fail maps err to its Connect error. A refused name carries only the
// rule it broke, never the name.
func (s *instrumentService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, kernel.ErrAuthzUnavailable):
		return connect.NewError(connect.CodeUnavailable, kernel.ErrAuthzUnavailable)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, ledger.ErrInvalidISIN):
		return connect.NewError(connect.CodeInvalidArgument, ledger.ErrInvalidISIN)
	case errors.Is(err, ledger.ErrInvalidInstrumentName):
		return connect.NewError(connect.CodeInvalidArgument, ledger.ErrInvalidInstrumentName)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "instrument call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}
