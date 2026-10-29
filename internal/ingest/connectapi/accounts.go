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
	"github.com/arkadiuss/openfief/internal/portfolio"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// NewAccounts returns the AccountService's path and HTTP handler over
// accounts (D-72, D-79), with callers in front of every call to admit view
// plugins only, logging to log.
func NewAccounts(accounts *app.Accounts, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ingestv1connect.NewAccountServiceHandler(&accountService{accounts: accounts, log: log}, connect.WithInterceptors(callers))
}

// accountService implements ingestv1connect.AccountServiceHandler.
type accountService struct {
	accounts *app.Accounts
	log      *slog.Logger
}

// CreateAccount creates the owner's account and connects its connector.
func (s *accountService) CreateAccount(ctx context.Context, req *connect.Request[ingestv1.CreateAccountRequest]) (*connect.Response[ingestv1.CreateAccountResponse], error) {
	m := req.Msg
	acc, err := s.accounts.Create(ctx, m.GetOrigin(), m.GetSessionToken(), &app.NewAccount{
		Name: m.GetName(), Institution: ingest.InstitutionID(m.GetInstitution()), Sleeve: portfolio.SleeveID(m.GetSleeve()),
		Instance: m.GetConnectorInstance(), BrokerAccount: ingest.BrokerAccountID(m.GetBrokerAccount()),
	})
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ingestv1.CreateAccountResponse{Account: pbAccount(&acc)}), nil
}

// UpdateAccount renames an account or moves it to another sleeve.
func (s *accountService) UpdateAccount(ctx context.Context, req *connect.Request[ingestv1.UpdateAccountRequest]) (*connect.Response[ingestv1.UpdateAccountResponse], error) {
	m := req.Msg
	acc, err := s.accounts.Update(ctx, m.GetOrigin(), m.GetSessionToken(), ledger.AccountID(m.GetAccountId()), m.GetName(), portfolio.SleeveID(m.GetSleeve()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ingestv1.UpdateAccountResponse{Account: pbAccount(&acc)}), nil
}

// BindAccount binds a connector instance to an account and maps its broker
// account.
func (s *accountService) BindAccount(ctx context.Context, req *connect.Request[ingestv1.BindAccountRequest]) (*connect.Response[ingestv1.BindAccountResponse], error) {
	m := req.Msg
	err := s.accounts.Bind(ctx, m.GetOrigin(), m.GetSessionToken(), ledger.AccountID(m.GetAccountId()),
		m.GetConnectorInstance(), ingest.BrokerAccountID(m.GetBrokerAccount()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ingestv1.BindAccountResponse{}), nil
}

// pbAccount converts an account.
func pbAccount(a *ledger.Account) *ingestv1.Account {
	return &ingestv1.Account{Id: string(a.ID), Institution: a.Institution, Name: a.Name, Sleeve: string(a.Sleeve)}
}

// fail maps err to its Connect error. A refusal carries its text, which
// names the account, sleeve or instance, so the owner can fix it; it never
// carries a broker account number.
func (s *accountService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, ledger.ErrInvalidAccount), errors.Is(err, app.ErrNoSleeve),
		errors.Is(err, ingest.ErrInvalidInstanceName), errors.Is(err, app.ErrBrokerAccountNeedsInstance):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, app.ErrAccountExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, app.ErrNoAccount):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, app.ErrUnknownBrokerAccount), errors.Is(err, app.ErrBrokerAccountMapped):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "account call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}
