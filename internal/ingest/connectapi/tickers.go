package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// NewTickers returns the TickerService's path and HTTP handler over the
// owner's ticker confirmations (D-116), with callers in front of every call
// to admit view plugins only, logging to log.
func NewTickers(t *app.Tickers, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ingestv1connect.NewTickerServiceHandler(&tickers{uc: t, log: log}, connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxMessageBytes))
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
			Waiting: count(p.Waiting), CurrencyChoices: p.CurrencyChoices(),
		}
		if !p.Proposal.ISIN.IsZero() {
			t.ProposedIsin, t.ProposedName, t.ProposedCurrency = p.Proposal.ISIN.String(), p.Proposal.Name, p.Proposal.Currency
		}
		out.Tickers = append(out.Tickers, t)
	}
	return connect.NewResponse(out), nil
}

// ConfirmTicker records the owner's ISIN for a ticker, with the trading
// currency the owner picked when its listing names none and its exchange
// quotes in more than one (D-170), and books what waits for it.
func (s *tickers) ConfirmTicker(ctx context.Context, req *connect.Request[ingestv1.ConfirmTickerRequest]) (*connect.Response[ingestv1.ConfirmTickerResponse], error) {
	isin, err := ledger.ParseISIN(req.Msg.GetIsin())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	c := ingest.TickerConfirmation{
		Instance: req.Msg.GetInstance(), Symbol: req.Msg.GetSymbol(), ISIN: isin, Currency: req.Msg.GetCurrency(),
	}
	sum, err := s.uc.Confirm(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), c)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	booked := count(sum.New + sum.NewVersions + sum.Unchanged)
	return connect.NewResponse(&ingestv1.ConfirmTickerResponse{Booked: booked}), nil
}

// ListConfirmedTickers returns the tickers whose ISIN is confirmed, once
// per institution (D-153).
func (s *tickers) ListConfirmedTickers(ctx context.Context, req *connect.Request[ingestv1.ListConfirmedTickersRequest]) (*connect.Response[ingestv1.ListConfirmedTickersResponse], error) {
	confirmed, err := s.uc.Confirmed(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &ingestv1.ListConfirmedTickersResponse{Tickers: make([]*ingestv1.ConfirmedTicker, 0, len(confirmed))}
	for i := range confirmed {
		c := &confirmed[i]
		out.Tickers = append(out.Tickers, &ingestv1.ConfirmedTicker{
			Institution: c.Institution, Instance: c.Instance, Symbol: c.Ticker.Symbol, Country: c.Ticker.Country,
			Name: c.Ticker.Name, Isin: c.ISIN.String(), TradingCurrency: c.TradingCurrency,
			CurrencyChoices: c.CurrencyChoices(), Booked: count(c.Booked),
		})
	}
	return connect.NewResponse(out), nil
}

// CorrectTicker records the owner's corrected ISIN for a confirmed ticker,
// with the trading currency the owner picked when its listing names none
// (D-170), and books every row of the ticker anew (D-153).
func (s *tickers) CorrectTicker(ctx context.Context, req *connect.Request[ingestv1.CorrectTickerRequest]) (*connect.Response[ingestv1.CorrectTickerResponse], error) {
	isin, err := ledger.ParseISIN(req.Msg.GetIsin())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	c := ingest.TickerConfirmation{
		Instance: req.Msg.GetInstance(), Symbol: req.Msg.GetSymbol(), ISIN: isin, Currency: req.Msg.GetCurrency(),
	}
	sum, err := s.uc.Correct(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), c)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ingestv1.CorrectTickerResponse{Booked: count(sum.New + sum.NewVersions + sum.Unchanged)}), nil
}

// fail maps err to its Connect error.
func (s *tickers) fail(ctx context.Context, err error) *connect.Error {
	if choice, ok := errors.AsType[*ingest.TradingCurrencyChoiceError](err); ok {
		return connect.NewError(connect.CodeInvalidArgument, choice)
	}
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, kernel.ErrAuthzUnavailable):
		return connect.NewError(connect.CodeUnavailable, kernel.ErrAuthzUnavailable)
	case errors.Is(err, kernel.ErrAuditUnavailable):
		return connect.NewError(connect.CodeUnavailable, kernel.ErrAuditUnavailable)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, ledger.ErrInvalidISIN):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrUnknownTicker):
		return connect.NewError(connect.CodeNotFound, app.ErrUnknownTicker)
	case errors.Is(err, app.ErrTickerConfirmed):
		return connect.NewError(connect.CodeFailedPrecondition, app.ErrTickerConfirmed)
	case errors.Is(err, ingest.ErrTickerNotConfirmed):
		return connect.NewError(connect.CodeFailedPrecondition, ingest.ErrTickerNotConfirmed)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "ticker confirmation failed", slog.String("error", app.LogSafe(err)))
	return connect.NewError(connect.CodeInternal, errInternal)
}
