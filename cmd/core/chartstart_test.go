package main

import (
	"fmt"
	"testing"

	"connectrpc.com/connect"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestChartStartSetting runs the owner's charts' start setting (D-167) and
// the mark of an account opened on its seeding day (D-168) through the
// real core: the portfolio's and an account's start are stored apart, the
// read API's performance names each as chosen_from, going back to the
// default clears it, a malformed start or an unknown account is refused,
// and the account's mark is stored and read back.
func TestChartStartSetting(t *testing.T) {
	t.Parallel()
	oc := ownerSession(t)
	ctx := t.Context()
	accounts := ingestv1connect.NewAccountServiceClient(oc.srv.Client(), oc.srv.URL, connect.WithInterceptors(viewBearer(uiCaller)))
	created, err := accounts.CreateAccount(ctx, connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: localUI, SessionToken: oc.session, Name: "Synthetic broker", Institution: "xtb", Sleeve: "core",
		OpenedOnSeedingDay: true,
	}))
	if err != nil || !created.Msg.GetAccount().GetOpenedOnSeedingDay() {
		t.Fatalf("CreateAccount = %v, %v; want it marked as opened on its seeding day", created, err)
	}
	var acc viewsv1.GetAccountResponse
	getJSON(t, oc.srv, "/v1/accounts/synthetic-broker", oc.session, &acc)
	if !acc.GetAccount().GetOpenedOnSeedingDay() {
		t.Errorf("the read API's account = %v, want it marked", acc.GetAccount())
	}

	for _, s := range []struct{ account, start string }{{"", "2026-03-02"}, {"synthetic-broker", "2026-04-01"}} {
		if setErr := setChartStart(t, oc, oc.session, s.account, s.start); setErr != nil {
			t.Fatalf("SetChartStart(%q, %s): %v", s.account, s.start, setErr)
		}
	}
	checkChosenFrom(t, oc, "/v1/performance", "2026-03-02")
	checkChosenFrom(t, oc, "/v1/performance?account_id=synthetic-broker", "2026-04-01")
	if clearErr := setChartStart(t, oc, oc.session, "", ""); clearErr != nil {
		t.Fatalf("SetChartStart back to the default: %v", clearErr)
	}
	checkChosenFrom(t, oc, "/v1/performance", "")
	checkChosenFrom(t, oc, "/v1/performance?account_id=synthetic-broker", "2026-04-01")
	checkChartStartRefusals(t, oc)
}

// setChartStart calls SetChartStart on oc's core with session.
func setChartStart(t *testing.T, oc ownerCore, session, account, start string) error {
	t.Helper()
	_, err := settingsClient(oc.srv, uiCaller).SetChartStart(t.Context(), connect.NewRequest(&valuationv1.SetChartStartRequest{
		Origin: localUI, SessionToken: session, AccountId: account, Start: start,
	}))
	if err != nil {
		return fmt.Errorf("SetChartStart: %w", err)
	}
	return nil
}

// checkChartStartRefusals checks that the core refuses a malformed start,
// an account that is not stored and an unknown session.
func checkChartStartRefusals(t *testing.T, oc ownerCore) {
	t.Helper()
	refusals := []struct {
		name, session, account, start string
		want                          connect.Code
	}{
		{name: "a malformed start", session: oc.session, start: "02.03.2026", want: connect.CodeInvalidArgument},
		{name: "an account that is not stored", session: oc.session, account: "acc-none", start: "2026-03-02", want: connect.CodeNotFound},
		{name: "an unknown session", session: "ofsess_nope_nope", start: "2026-03-02", want: connect.CodeUnauthenticated},
	}
	for _, r := range refusals {
		if err := setChartStart(t, oc, r.session, r.account, r.start); connect.CodeOf(err) != r.want {
			t.Errorf("%s: SetChartStart = %v, want %s", r.name, err, r.want)
		}
	}
}

// checkChosenFrom checks that GET path names want as the stored charts'
// start.
func checkChosenFrom(t *testing.T, oc ownerCore, path, want string) {
	t.Helper()
	var perf viewsv1.GetPerformanceResponse
	getJSON(t, oc.srv, path, oc.session, &perf)
	if perf.GetChosenFrom() != want {
		t.Errorf("GET %s: chosen_from = %q, want %q", path, perf.GetChosenFrom(), want)
	}
}
