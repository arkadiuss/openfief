package main

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/kernel"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// bondsConnector maps bond lots as the treasury-bonds plugin does, into the
// demo's bond account, as of coreNow's date.
type bondsConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler
}

func (bondsConnector) ImportBondLots(_ context.Context, req *connect.Request[connectorv1.ImportBondLotsRequest]) (*connect.Response[connectorv1.ImportBondLotsResponse], error) {
	lots, err := delivery.BondLots(req.Msg.GetLots())
	if err != nil {
		return nil, refused(err)
	}
	today, err := kernel.ParseDate("2026-10-28")
	if err != nil {
		return nil, refused(err)
	}
	m, terms, err := treasurybonds.Mapper{Account: demo.BondsAccount}.Map(lots, today)
	if err != nil {
		return nil, refused(err)
	}
	d, err := delivery.FromMapping(req.Msg.GetBatchId(), string(treasurybonds.Institution), &m)
	if err != nil {
		return nil, refused(err)
	}
	d.BondTerms = delivery.BondTerms(terms)
	return connect.NewResponse(&connectorv1.ImportBondLotsResponse{Delivery: d}), nil
}

// refused is the connector's INVALID_ARGUMENT answer to err.
func refused(err error) *connect.Error { return connect.NewError(connect.CodeInvalidArgument, err) }

// staticTokens issues coreToken for every instance.
type staticTokens struct{}

func (staticTokens) TokenFor(context.Context, string) (string, error) { return coreToken, nil }

// bondLotAPI is a generated BondLotService client.
type bondLotAPI struct {
	ingestv1connect.BondLotServiceClient
}

// bondLotClient returns a BondLotService client of srv that presents token
// as the view plugin's.
func bondLotClient(srv *httptest.Server, token string) bondLotAPI {
	bearer := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	})
	return bondLotAPI{ingestv1connect.NewBondLotServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer))}
}

// edoRequest asks to add a synthetic lot of 40 EDO0536 bonds bought on
// 2026-05-04, at 6.80% then a 2.00% margin, with a 2.00 PLN fee, to account.
func edoRequest(session, account string) *connect.Request[ingestv1.AddBondLotRequest] {
	return connect.NewRequest(&ingestv1.AddBondLotRequest{
		Origin: localUI, SessionToken: session, AccountId: account,
		Lot: &ingestv1.BondLot{
			Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04",
			FirstPeriodRatePercent: "6.80", MarginPercent: "2.00", EarlyRedemptionFeePln: "2.00",
		},
	})
}

func TestBondLotService(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	out := enrollCmd(t, dsn)
	code := enrollCode(t, out)
	c, srv, _ := startCoreWith(t, dsn, func(cfg *serveConfig) { cfg.Auth.DevCallerToken = uiCaller })
	ac := authClient(srv.URL, srv.Client(), uiCaller)
	session := passkeyLogin(t, ac, passkeyEnroll(t, ac, code))
	saveDemoRegistry(t, dsn)
	if _, err := connectorCmd(t, dsn, nil, "bind", "--instance=bonds-main", "--account="+string(demo.BondsAccount)); err != nil {
		t.Fatalf("connector bind: %v", err)
	}
	lots := bondLotClient(srv, uiCaller)
	if _, err := lots.AddBondLot(t.Context(), edoRequest(session, string(demo.BondsAccount))); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("AddBondLot before the core found any connector = %v, want Unavailable", err)
	}
	port := serveConnector(t, bondsConnector{})
	c.deps.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), func(string) (string, error) {
		return fmt.Sprintf("http://127.0.0.1:%d", port), nil
	}, staticTokens{}))

	got, err := lots.AddBondLot(t.Context(), edoRequest(session, string(demo.BondsAccount)))
	if err != nil || got.Msg.GetNew() != 2 {
		t.Fatalf("AddBondLot = %v, %v; want a deposit and a buy, both new", got, err)
	}
	again, err := lots.AddBondLot(t.Context(), edoRequest(session, string(demo.BondsAccount)))
	if err != nil || again.Msg.GetUnchanged() != 2 || again.Msg.GetNew() != 0 {
		t.Errorf("entering the lot again = %v, %v; want both unchanged (D-18)", again, err)
	}
	checkBondLotRefusals(t, lots, session)
	checkBondValued(t, srv, session)
}

// enrollCode returns the code of the enrollment link enroll-link printed.
func enrollCode(t *testing.T, out string) string {
	t.Helper()
	var link string
	if _, err := fmt.Sscanln(out, &link); err != nil {
		t.Fatalf("enroll-link printed %q: %v", out, err)
	}
	const prefix = localUI + "/enroll#code="
	if len(link) <= len(prefix) || link[:len(prefix)] != prefix {
		t.Fatalf("enroll-link printed %q", out)
	}
	return link[len(prefix):]
}

// checkBondLotRefusals checks that the core refuses a lot from a caller that
// is not a view plugin, from an unknown session, and for an account no
// connector is bound to.
func checkBondLotRefusals(t *testing.T, lots bondLotAPI, session string) {
	t.Helper()
	ctx := t.Context()
	if _, err := lots.AddBondLot(ctx, edoRequest("ofsess_nope_nope", string(demo.BondsAccount))); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("an unknown session: %v, want Unauthenticated", err)
	}
	if _, err := lots.AddBondLot(ctx, edoRequest(session, string(demo.IBKRAccount))); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("an account without a bond connector: %v, want FailedPrecondition", err)
	}
}

// checkBondValued waits until the demo bond account is valued through
// 2026-10-28 and checks the lot's stored value there: the existing bond
// rules, net if redeemed in a taxable account. One bond is worth
// 100 × (1 + 0.068 × 177 / 365) = 103.30 PLN, so 4132.00 for the lot; the
// 2.00 PLN fee per bond is under its 3.30 PLN of interest, 80.00 in all;
// Belka tax is 19% × (132.00 − 80.00) = 9.88; net 4042.12 PLN.
func checkBondValued(t *testing.T, srv *httptest.Server, session string) {
	t.Helper()
	path := "/v1/accounts/" + string(demo.BondsAccount) + "/valuations?from=2026-10-28&to=2026-10-28"
	deadline := time.Now().Add(2 * time.Minute)
	for {
		var got viewsv1.ListAccountValuationsResponse
		getJSON(t, srv, path, session, &got)
		for _, v := range got.GetValuations() {
			if v.GetInstrument().GetSymbol() != "EDO0536" || v.GetValue() == nil {
				continue
			}
			if v.GetBasis() != viewsv1.ValuationBasis_VALUATION_BASIS_NET_IF_REDEEMED || v.GetValue().GetAmount() != "4042.12" {
				t.Errorf("the lot's valuation = %v, want net if redeemed 4042.12 PLN", v)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the bond lot was not valued by %s: %v", deadline, got.GetValuations())
		}
		time.Sleep(50 * time.Millisecond)
	}
}
