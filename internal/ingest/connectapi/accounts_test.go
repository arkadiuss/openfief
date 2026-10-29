package connectapi_test

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"

	gocmp "github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectapi"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// memRegistry is the ledger's accounts over the default sleeves, in
// memory, or failing with err.
type memRegistry struct {
	accounts []ledger.Account
	err      error
}

func (m *memRegistry) Accounts(context.Context) ([]ledger.Account, error) { return m.accounts, m.err }

func (m *memRegistry) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	return portfolio.DefaultSleeves(), m.err
}

func (m *memRegistry) SaveAccount(_ context.Context, a ledger.Account) error {
	m.accounts = append(slices.DeleteFunc(m.accounts, func(o ledger.Account) bool { return o.ID == a.ID }), a)
	return nil
}

// memLinks keeps connector bindings and the synthetic broker accounts
// U0000001, unmapped, and U0000002, mapped to acc-bonds, of ibkr-main.
type memLinks struct {
	bound   []string
	brokers []ingest.BrokerAccount
}

func (m *memLinks) Bind(_ context.Context, instance string, account ledger.AccountID, _ time.Time) error {
	m.bound = append(m.bound, instance+">"+string(account))
	return nil
}

func (m *memLinks) BoundAccounts(context.Context, string) ([]ledger.AccountID, error) {
	return nil, nil
}

func (m *memLinks) RecordBrokerAccounts(context.Context, string, ingest.InstitutionID, []ingest.BrokerAccountID, kernel.Date) error {
	return nil
}

func (m *memLinks) BrokerAccounts(context.Context, string) ([]ingest.BrokerAccount, error) {
	return m.brokers, nil
}

func (m *memLinks) MarkSeeded(context.Context, string, []ingest.BrokerAccountID) error { return nil }

func (m *memLinks) MapBrokerAccount(_ context.Context, _ string, id ingest.BrokerAccountID, account ledger.AccountID, _ time.Time) error {
	for i := range m.brokers {
		if m.brokers[i].ID == id {
			m.brokers[i].Account = account
		}
	}
	return nil
}

// accountsClient is the generated AccountService client.
type accountsClient struct {
	ingestv1connect.AccountServiceClient
}

// newAccountsClient serves the AccountService over the fakes and returns
// a client calling it with the caller token.
func newAccountsClient(t *testing.T, reg *memRegistry, links *memLinks, caller string) accountsClient {
	t.Helper()
	accounts := &app.Accounts{
		Owners: owners{}, Ledger: reg, Bindings: links, Brokers: links,
		Now: func() time.Time { return time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC) },
	}
	_, handler := connectapi.NewAccounts(accounts, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return accountsClient{ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

// fixture returns the fakes: acc-bonds stored, ibkr-main's broker accounts.
func fixture() (*memRegistry, *memLinks) {
	return &memRegistry{accounts: []ledger.Account{{ID: bondsAccount, Institution: "treasury-bonds", Name: "Bonds", Sleeve: "bonds"}}},
		&memLinks{brokers: []ingest.BrokerAccount{
			{Instance: "ibkr-main", ID: "U0000001", Institution: "ibkr"},
			{Instance: "ibkr-main", ID: "U0000002", Institution: "ibkr", Account: bondsAccount},
		}}
}

func TestAccountServiceRoundTrip(t *testing.T) {
	t.Parallel()
	reg, links := fixture()
	c := newAccountsClient(t, reg, links, viewCaller)
	ctx := t.Context()
	created, err := c.CreateAccount(ctx, connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: local, SessionToken: ownerSession, Name: "IBKR main", Sleeve: "satellite",
		ConnectorInstance: "ibkr-main", BrokerAccount: "U0000001",
	}))
	want := &ingestv1.Account{Id: "ibkr-main", Institution: "ibkr", Name: "IBKR main", Sleeve: "satellite"}
	if err != nil || !gocmp.Equal(want, created.Msg.GetAccount(), protocmp.Transform()) {
		t.Fatalf("CreateAccount = %v, %v; want %v", created, err, want)
	}
	updated, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: "ibkr-main", Name: "IBKR", Sleeve: "core",
	}))
	if err != nil || updated.Msg.GetAccount().GetSleeve() != "core" || updated.Msg.GetAccount().GetName() != "IBKR" {
		t.Errorf("UpdateAccount = %v, %v; want IBKR in core", updated, err)
	}
	if _, err := c.BindAccount(ctx, connect.NewRequest(&ingestv1.BindAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, ConnectorInstance: "bonds-main",
	})); err != nil {
		t.Errorf("BindAccount: %v", err)
	}
	if diff := gocmp.Diff([]string{"ibkr-main>ibkr-main", "bonds-main>acc-bonds"}, links.bound); diff != "" || links.brokers[0].Account != "ibkr-main" {
		t.Errorf("bindings (-want +got):\n%s\nU0000001 mapped to %q", diff, links.brokers[0].Account)
	}
}

func TestAccountServiceErrors(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	create := func(session, name, sleeve, instance, broker string) *connect.Request[ingestv1.CreateAccountRequest] {
		return connect.NewRequest(&ingestv1.CreateAccountRequest{
			Origin: local, SessionToken: session, Name: name, Institution: "x", Sleeve: sleeve, ConnectorInstance: instance, BrokerAccount: broker,
		})
	}
	tests := []struct {
		name   string
		caller string
		down   bool
		call   func(c accountsClient) connect.Code
		want   connect.Code
	}{
		{
			name: "not a view plugin", caller: "another-plugin", want: connect.CodePermissionDenied,
			call: func(c accountsClient) connect.Code {
				_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "", ""))
				return connect.CodeOf(err)
			},
		},
		{name: "no session", want: connect.CodeUnauthenticated, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create("nope", "Cash", "core", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "a read session", want: connect.CodePermissionDenied, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(readSession, "Cash", "core", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "a taken ID", want: connect.CodeAlreadyExists, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Acc bonds", "core", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "an unknown sleeve", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "gold", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "a malformed instance", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "IBKR", ""))
			return connect.CodeOf(err)
		}},
		{name: "a broker account mapped elsewhere", want: connect.CodeFailedPrecondition, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "ibkr-main", "U0000002"))
			return connect.CodeOf(err)
		}},
		{name: "a broker account never reported", want: connect.CodeFailedPrecondition, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "ibkr-main", "U0000009"))
			return connect.CodeOf(err)
		}},
		{name: "an unknown account", want: connect.CodeNotFound, call: func(c accountsClient) connect.Code {
			_, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{Origin: local, SessionToken: ownerSession, AccountId: "nope", Name: "X", Sleeve: "core"}))
			return connect.CodeOf(err)
		}},
		{name: "the ledger down", down: true, want: connect.CodeInternal, call: func(c accountsClient) connect.Code {
			_, err := c.BindAccount(ctx, connect.NewRequest(&ingestv1.BindAccountRequest{Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, ConnectorInstance: "x"}))
			return connect.CodeOf(err)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg, links := fixture()
			if tt.down {
				reg.err = errDown
			}
			caller := viewCaller
			if tt.caller != "" {
				caller = tt.caller
			}
			if got := tt.call(newAccountsClient(t, reg, links, caller)); got != tt.want {
				t.Errorf("code = %v, want %v", got, tt.want)
			}
			if len(links.bound) != 0 {
				t.Errorf("a refused call bound %v", links.bound)
			}
		})
	}
}
