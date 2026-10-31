package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ledger"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// NewTickers returns the TickerService's path and HTTP handler over the
// owner's ticker confirmations (D-116), with callers in front of every call
// to admit view plugins only, logging to log.
func NewTickers(t *app.Tickers, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ingestv1connect.NewTickerServiceHandler(&tickers{uc: t, log: log}, connect.WithInterceptors(callers))
}

// tickers implements ingestv1connect.TickerServiceHandler.
type tickers struct {
	uc  *app.Tickers
	log *slog.Logger
}

// ListPendingTickers returns the tickers waiting for the owner.
func (s *tickers) ListPendingTickers(ctx context.Context, req *connect.Request[ingestv1.ListPendingTickersRequest]) (*connect.Response[ingestv1.ListPendingTickersResponse], error) {
	pending, err := s.uc.Pending(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &ingestv1.ListPendingTickersResponse{Tickers: make([]*ingestv1.PendingTicker, 0, len(pending))}
	for i := range pending {
		p := &pending[i]
		t := &ingestv1.PendingTicker{
			Instance: p.Instance, Symbol: p.Ticker.Symbol, Country: p.Ticker.Country, Name: p.Ticker.Name,
			Waiting: count(p.Waiting),
		}
		if !p.Proposal.ISIN.IsZero() {
			t.ProposedIsin, t.ProposedName = p.Proposal.ISIN.String(), p.Proposal.Name
		}
		out.Tickers = append(out.Tickers, t)
	}
	return connect.NewResponse(out), nil
}

// ConfirmTicker records the owner's ISIN for a ticker and books what waits
// for it.
func (s *tickers) ConfirmTicker(ctx context.Context, req *connect.Request[ingestv1.ConfirmTickerRequest]) (*connect.Response[ingestv1.ConfirmTickerResponse], error) {
	isin, err := ledger.ParseISIN(req.Msg.GetIsin())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	c := ingest.TickerConfirmation{Instance: req.Msg.GetInstance(), Symbol: req.Msg.GetSymbol(), ISIN: isin}
	sum, err := s.uc.Confirm(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), c)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	booked := count(sum.New + sum.NewVersions + sum.Unchanged)
	return connect.NewResponse(&ingestv1.ConfirmTickerResponse{Booked: booked}), nil
}

// fail maps err to its Connect error.
func (s *tickers) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, ledger.ErrInvalidISIN):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrUnknownTicker):
		return connect.NewError(connect.CodeNotFound, app.ErrUnknownTicker)
	case errors.Is(err, app.ErrTickerConfirmed):
		return connect.NewError(connect.CodeFailedPrecondition, app.ErrTickerConfirmed)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "ticker confirmation failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}
