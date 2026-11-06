package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// errMalformed reports a request whose date, currency or amount doesn't
// parse.
var errMalformed = errors.New("malformed cash checkpoint")

// NewCashCheckpoints returns the CashCheckpointService's path and HTTP
// handler over checkpoints (D-158, D-79), with callers in front of every
// call to admit view plugins only, logging to log.
func NewCashCheckpoints(checkpoints *app.CashCheckpoints, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ledgerv1connect.NewCashCheckpointServiceHandler(&checkpointService{checkpoints: checkpoints, log: log},
		connect.WithInterceptors(callers))
}

// checkpointService implements ledgerv1connect.CashCheckpointServiceHandler.
type checkpointService struct {
	checkpoints *app.CashCheckpoints
	log         *slog.Logger
}

// RecordCashCheckpoint records the cash an institution showed.
func (s *checkpointService) RecordCashCheckpoint(ctx context.Context, req *connect.Request[ledgerv1.RecordCashCheckpointRequest]) (*connect.Response[ledgerv1.RecordCashCheckpointResponse], error) {
	m := req.Msg
	on, c, err := dateAndCurrency(m.GetDate(), m.GetCurrency())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	cash, err := kernel.ParseMoney(m.GetAmount(), c)
	if err != nil {
		return nil, s.fail(ctx, fmt.Errorf("%w: amount %q: %w", errMalformed, m.GetAmount(), err))
	}
	check, err := s.checkpoints.Record(ctx, m.GetOrigin(), m.GetSessionToken(), ledger.AccountID(m.GetAccountId()), on, cash)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	a := &check.Assertion
	return connect.NewResponse(&ledgerv1.RecordCashCheckpointResponse{Checkpoint: &ledgerv1.CashCheckpoint{
		AccountId: string(a.Account), Date: a.Date.String(), Currency: a.Book.Currency.String(),
		Shown: a.Balance.String(), Derived: check.Derived.String(), Matched: !check.Flagged(),
	}}), nil
}

// BookCashDifference books the difference to a checkpoint.
func (s *checkpointService) BookCashDifference(ctx context.Context, req *connect.Request[ledgerv1.BookCashDifferenceRequest]) (*connect.Response[ledgerv1.BookCashDifferenceResponse], error) {
	m := req.Msg
	on, c, err := dateAndCurrency(m.GetDate(), m.GetCurrency())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	t, err := s.checkpoints.BookDifference(ctx, m.GetOrigin(), m.GetSessionToken(), ledger.AccountID(m.GetAccountId()), on, c)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	cash := t.Postings[0].Amount
	if cash.Sign() < 0 {
		cash = cash.Neg()
	}
	return connect.NewResponse(&ledgerv1.BookCashDifferenceResponse{
		Key:  &ledgerv1.TransactionKey{Source: t.Key.Source, ExternalId: t.Key.ExternalID},
		Kind: t.Kind.String(), Amount: cash.Amount().String(), Currency: cash.Currency().String(),
	}), nil
}

// dateAndCurrency parses a request's date and currency.
func dateAndCurrency(date, currency string) (kernel.Date, kernel.Currency, error) {
	on, err := kernel.ParseDate(date)
	if err != nil {
		return kernel.Date{}, kernel.Currency{}, fmt.Errorf("%w: date %q: %w", errMalformed, date, err)
	}
	c, err := kernel.ParseCurrency(currency)
	if err != nil {
		return kernel.Date{}, kernel.Currency{}, fmt.Errorf("%w: currency %q: %w", errMalformed, currency, err)
	}
	return on, c, nil
}

// fail maps err to its Connect error. A refusal carries its text, which
// names the account, date and currency, so the owner can see why.
func (s *checkpointService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, errMalformed), errors.Is(err, ledger.ErrInvalidAssertion):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrUnknownAccount), errors.Is(err, app.ErrUnknownCheckpoint):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ledger.ErrNoCashDifference):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "cash checkpoint call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}
