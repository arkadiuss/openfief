package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// errMarkerDown is the fixed demo marker's failure.
var errMarkerDown = errors.New("marker table unreachable")

// fixedDemo is a demoLoads with a fixed answer.
type fixedDemo struct {
	loading bool
	err     error
}

func (d fixedDemo) DemoLoading(context.Context) (bool, error) { return d.loading, d.err }

// acceptingAccounts is an AccountService that accepts every account and
// lists no broker account.
type acceptingAccounts struct {
	ingestv1connect.UnimplementedAccountServiceHandler
}

func (acceptingAccounts) CreateAccount(context.Context, *connect.Request[ingestv1.CreateAccountRequest]) (*connect.Response[ingestv1.CreateAccountResponse], error) {
	return connect.NewResponse(&ingestv1.CreateAccountResponse{}), nil
}

func (acceptingAccounts) ListBrokerAccounts(context.Context, *connect.Request[ingestv1.ListBrokerAccountsRequest]) (*connect.Response[ingestv1.ListBrokerAccountsResponse], error) {
	return connect.NewResponse(&ingestv1.ListBrokerAccountsResponse{}), nil
}

// acceptingSleeves is a SleeveService that accepts every sleeve.
type acceptingSleeves struct {
	portfoliov1connect.UnimplementedSleeveServiceHandler
}

func (acceptingSleeves) AddSleeve(context.Context, *connect.Request[portfoliov1.AddSleeveRequest]) (*connect.Response[portfoliov1.AddSleeveResponse], error) {
	return connect.NewResponse(&portfoliov1.AddSleeveResponse{}), nil
}

// TestRefuseWhileDemoLoads checks that, while a demo deployment's data is
// loading, the core refuses what the first-run wizard would write, sleeves
// and accounts, with FailedPrecondition, and lets reads through; that a
// ready demo and an install without demo sessions accept the writes; and
// that a demo marker it can't read refuses as internal.
func TestRefuseWhileDemoLoads(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		demo      fixedDemo
		wantWrite connect.Code
		wantRead  connect.Code
	}{
		{name: "demo data loading: writes refused, reads served", demo: fixedDemo{loading: true}, wantWrite: connect.CodeFailedPrecondition},
		{name: "demo ready or no demo: writes accepted", demo: fixedDemo{}},
		{name: "demo marker unreadable: writes refused as internal", demo: fixedDemo{err: errMarkerDown}, wantWrite: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			accountPath, accountHandler := ingestv1connect.NewAccountServiceHandler(acceptingAccounts{})
			sleevePath, sleeveHandler := portfoliov1connect.NewSleeveServiceHandler(acceptingSleeves{})
			mux := http.NewServeMux()
			mux.Handle(accountPath, refuseWhileDemoLoads(accountHandler, tt.demo, quiet()))
			mux.Handle(sleevePath, refuseWhileDemoLoads(sleeveHandler, tt.demo, quiet()))
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)
			accounts := ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL)
			sleeves := portfoliov1connect.NewSleeveServiceClient(srv.Client(), srv.URL)
			_, err := accounts.CreateAccount(t.Context(), connect.NewRequest(&ingestv1.CreateAccountRequest{Name: "Synthetic"}))
			wantDemoAnswer(t, "CreateAccount", err, tt.wantWrite)
			_, err = sleeves.AddSleeve(t.Context(), connect.NewRequest(&portfoliov1.AddSleeveRequest{}))
			wantDemoAnswer(t, "AddSleeve", err, tt.wantWrite)
			_, err = accounts.ListBrokerAccounts(t.Context(), connect.NewRequest(&ingestv1.ListBrokerAccountsRequest{}))
			wantDemoAnswer(t, "ListBrokerAccounts", err, tt.wantRead)
		})
	}
}

// wantDemoAnswer checks that the call to method succeeded when want is 0,
// and otherwise failed with want.
func wantDemoAnswer(t *testing.T, method string, err error, want connect.Code) {
	t.Helper()
	if want == 0 {
		if err != nil {
			t.Errorf("%s = %v, want it served", method, err)
		}
		return
	}
	if connect.CodeOf(err) != want {
		t.Errorf("%s = %v, want %v", method, err, want)
	}
}
