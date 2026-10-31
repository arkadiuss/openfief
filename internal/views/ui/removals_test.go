package ui

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
)

// ListBrokerAccounts lists, to the owner only (D-121), the synthetic broker
// accounts ibkr-main reported: U0000001 unmapped, U0000002 mapped to the
// synthetic account.
func (f *fakeAccounts) ListBrokerAccounts(_ context.Context, r *connect.Request[ingestv1.ListBrokerAccountsRequest]) (*connect.Response[ingestv1.ListBrokerAccountsResponse], error) {
	if r.Header().Get("Authorization") != "Bearer "+pluginToken {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&ingestv1.ListBrokerAccountsResponse{BrokerAccounts: []*ingestv1.BrokerAccount{
		{ConnectorInstance: "ibkr-main", Id: "U0000001", Institution: "ibkr", LastSeen: "2026-03-31", InLastDelivery: true},
		{ConnectorInstance: "ibkr-main", Id: "U0000002", Institution: "ibkr", AccountId: accountID, LastSeen: "2026-03-31"},
	}}), nil
}

// errHasTransactions is the core's refusal to delete an account that holds
// transactions, or to remap a broker account booked into one (D-120).
var errHasTransactions = errors.New("the account holds transactions")

// DeleteAccount deletes any account but the synthetic one, which holds
// transactions.
func (f *fakeAccounts) DeleteAccount(_ context.Context, r *connect.Request[ingestv1.DeleteAccountRequest]) (*connect.Response[ingestv1.DeleteAccountResponse], error) {
	m := r.Msg
	if err := f.admit(r.Header(), m.GetSessionToken(), "delete "+m.GetAccountId()); err != nil {
		return nil, err
	}
	if m.GetAccountId() == accountID {
		return nil, codeErr(connect.CodeFailedPrecondition, errHasTransactions)
	}
	return connect.NewResponse(&ingestv1.DeleteAccountResponse{}), nil
}

// RemapBrokerAccount remaps any broker account but U0000002, whose account
// holds transactions from it.
func (f *fakeAccounts) RemapBrokerAccount(_ context.Context, r *connect.Request[ingestv1.RemapBrokerAccountRequest]) (*connect.Response[ingestv1.RemapBrokerAccountResponse], error) {
	m := r.Msg
	if err := f.admit(r.Header(), m.GetSessionToken(), "remap "+m.GetConnectorInstance()+"|"+m.GetBrokerAccount()+"|"+m.GetAccountId()); err != nil {
		return nil, err
	}
	if m.GetBrokerAccount() == "U0000002" && m.GetAccountId() != accountID {
		return nil, codeErr(connect.CodeFailedPrecondition, errHasTransactions)
	}
	return connect.NewResponse(&ingestv1.RemapBrokerAccountResponse{}), nil
}

// errInUse is the core's refusal to delete a sleeve with accounts or a kind
// a sleeve has (D-120).
var errInUse = errors.New("in use")

// DeleteSleeve deletes any sleeve but core, which has accounts.
func (f *fakeSleeves) DeleteSleeve(_ context.Context, r *connect.Request[portfoliov1.DeleteSleeveRequest]) (*connect.Response[portfoliov1.DeleteSleeveResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken(), "delete sleeve "+r.Msg.GetId()); err != nil {
		return nil, err
	}
	if r.Msg.GetId() == "core" {
		return nil, codeErr(connect.CodeFailedPrecondition, errInUse)
	}
	return connect.NewResponse(&portfoliov1.DeleteSleeveResponse{}), nil
}

// DeleteSleeveKind deletes any kind but core, which a sleeve has.
func (f *fakeSleeves) DeleteSleeveKind(_ context.Context, r *connect.Request[portfoliov1.DeleteSleeveKindRequest]) (*connect.Response[portfoliov1.DeleteSleeveKindResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken(), "delete kind "+r.Msg.GetKey()); err != nil {
		return nil, err
	}
	if r.Msg.GetKey() == "core" {
		return nil, codeErr(connect.CodeFailedPrecondition, errInUse)
	}
	return connect.NewResponse(&portfoliov1.DeleteSleeveKindResponse{}), nil
}
