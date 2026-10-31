// Package connectapi serves the ledger context's TransferService
// (proto/openfief/ledger/v1/transfers.proto) to view plugins over Connect:
// the built-in UI hands the owner's decision on a suggested transfer to the
// core through it (D-105, D-79). Callers are admitted by the interceptor
// the caller passes, cmd/core's view-plugin check; the owner's session
// travels in the request and the use case checks its scope. Errors map to
// Connect codes: no valid session is unauthenticated, a session that is not
// the owner's is permission_denied, a pair that can't be a transfer is
// invalid_argument, an unknown leg not_found, a leg of another confirmed
// transfer failed_precondition, anything else internal with nothing of the
// error.
package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// errInternal is what a caller sees of an internal failure.
var errInternal = errors.New("internal error")

// NewTransfers returns the TransferService's path and HTTP handler over
// transfers, with callers in front of every call to admit view plugins
// only, logging to log.
func NewTransfers(transfers *app.Transfers, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ledgerv1connect.NewTransferServiceHandler(&transferService{transfers: transfers, log: log},
		connect.WithInterceptors(callers))
}

// transferService implements ledgerv1connect.TransferServiceHandler.
type transferService struct {
	transfers *app.Transfers
	log       *slog.Logger
}

// ConfirmTransfer records the owner's confirmation of a pair.
func (s *transferService) ConfirmTransfer(ctx context.Context, req *connect.Request[ledgerv1.ConfirmTransferRequest]) (*connect.Response[ledgerv1.ConfirmTransferResponse], error) {
	m := req.Msg
	l, err := s.transfers.Decide(ctx, m.GetOrigin(), m.GetSessionToken(), pairFrom(m.GetWithdrawal(), m.GetDeposit()), ledger.TransferConfirmed)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ledgerv1.ConfirmTransferResponse{Link: pbLink(&l)}), nil
}

// RejectTransfer records the owner's rejection of a pair.
func (s *transferService) RejectTransfer(ctx context.Context, req *connect.Request[ledgerv1.RejectTransferRequest]) (*connect.Response[ledgerv1.RejectTransferResponse], error) {
	m := req.Msg
	l, err := s.transfers.Decide(ctx, m.GetOrigin(), m.GetSessionToken(), pairFrom(m.GetWithdrawal(), m.GetDeposit()), ledger.TransferRejected)
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ledgerv1.RejectTransferResponse{Link: pbLink(&l)}), nil
}

// fail maps err to its Connect error. A refused pair carries its text,
// which names the legs and the rule, so the owner can see why.
func (s *transferService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, ledger.ErrInvalidTransfer):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrUnknownTransaction):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ledger.ErrTransferTaken):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "transfer call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// pairFrom reads a pair from its legs' keys.
func pairFrom(w, d *ledgerv1.TransactionKey) ledger.TransferPair {
	return ledger.TransferPair{
		Withdrawal: ledger.TransactionKey{Source: w.GetSource(), ExternalID: w.GetExternalId()},
		Deposit:    ledger.TransactionKey{Source: d.GetSource(), ExternalID: d.GetExternalId()},
	}
}

// pbLink converts a transfer link.
func pbLink(l *ledger.TransferLink) *ledgerv1.TransferLink {
	decision := ledgerv1.TransferDecision_TRANSFER_DECISION_UNSPECIFIED
	switch l.Decision {
	case ledger.TransferConfirmed:
		decision = ledgerv1.TransferDecision_TRANSFER_DECISION_CONFIRMED
	case ledger.TransferRejected:
		decision = ledgerv1.TransferDecision_TRANSFER_DECISION_REJECTED
	}
	return &ledgerv1.TransferLink{
		Withdrawal: &ledgerv1.TransactionKey{Source: l.Pair.Withdrawal.Source, ExternalId: l.Pair.Withdrawal.ExternalID},
		Deposit:    &ledgerv1.TransactionKey{Source: l.Pair.Deposit.Source, ExternalId: l.Pair.Deposit.ExternalID},
		Decision:   decision,
		Version:    int64(l.Version),
	}
}
