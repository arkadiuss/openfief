package ui

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

var (
	errSeriesMismatch = errors.New("invalid bond lot: EDO0636 bought on 2026-05-04 matures on 2036-05-04")
	errBondsDown      = errors.New("the connector can't be reached")
)

// fakeBondLots is an in-memory BondLotService. Like the core, it wants the
// plugin's token and an owner-scoped session; it refuses series EDO0636 as
// a mismatched series and answers series COI0530 as if the connector were
// down.
type fakeBondLots struct {
	ingestv1connect.UnimplementedBondLotServiceHandler

	mu      sync.Mutex
	bearers []string
	added   []*ingestv1.AddBondLotRequest
}

func (f *fakeBondLots) AddBondLot(_ context.Context, r *connect.Request[ingestv1.AddBondLotRequest]) (*connect.Response[ingestv1.AddBondLotResponse], error) {
	f.locked(func() { f.bearers = append(f.bearers, r.Header().Get("Authorization")) })
	if r.Header().Get("Authorization") != "Bearer "+pluginToken {
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	switch r.Msg.GetLot().GetSeries() {
	case "EDO0636":
		return nil, codeErr(connect.CodeInvalidArgument, errSeriesMismatch)
	case "COI0530":
		return nil, codeErr(connect.CodeUnavailable, errBondsDown)
	}
	f.locked(func() { f.added = append(f.added, r.Msg) })
	return connect.NewResponse(&ingestv1.AddBondLotResponse{BatchId: "b-1", New: 2}), nil
}

// locked runs fn with the fake's lock held.
func (f *fakeBondLots) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// bondHarness is the harness with the fake account at the treasury bond
// registry.
func bondHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	h.read.locked(func() { h.read.institution = bondInstitution })
	return h
}

// lotForm is a posted lot of 40 bonds of series, bought on 2026-05-04, at
// 6.80% then a 2.00% margin, with a 2.00 PLN fee, typed with spaces and in
// lower case.
func lotForm(series string) *http.Request {
	return form("/accounts/"+accountID+"/bond-lots",
		"series=+"+series+"+&quantity=40&purchase_date=2026-05-04&first_rate=6.80&margin=2.00&fee=2.00")
}

// TestAccountBondLotForm checks that only the owner sees the "Add bond
// lot" form, and only on a bond account's page.
func TestAccountBondLotForm(t *testing.T) {
	t.Parallel()
	h := bondHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<h2 id="bond-lot-title">Add bond lot</h2>`, `action="/accounts/`+accountID+`/bond-lots"`,
		`name="series"`, `name="purchase_date"`, `name="first_rate"`, `name="fee"`, "Open lots")
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Add bond lot")
	other := newHarness(t)
	rec = other.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Add bond lot")
}

// TestAddBondLot checks that a posted lot goes to the core with the
// plugin's token and the owner's session, and that the core's refusal and
// an unreachable connector show the page again with the reason and the
// lot as posted.
func TestAddBondLot(t *testing.T) {
	t.Parallel()
	h := bondHarness(t)
	rec := h.serve(t, withSession(lotForm("edo0536"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/accounts/"+accountID+"#lots-title" {
		t.Errorf("Location = %q, want the account's open lots", loc)
	}
	var added []*ingestv1.AddBondLotRequest
	h.bondLots.locked(func() { added = append(added, h.bondLots.added...) })
	want := []*ingestv1.AddBondLotRequest{{
		Origin: "https://ui.example.com", SessionToken: ownerSession, AccountId: accountID,
		Lot: &ingestv1.BondLot{
			Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04",
			FirstPeriodRatePercent: "6.80", MarginPercent: "2.00", EarlyRedemptionFeePln: "2.00",
		},
	}}
	if diff := cmp.Diff(want, added, protocmp.Transform()); diff != "" {
		t.Errorf("lots handed to the core (-want +got):\n%s", diff)
	}
	if bearers := seen(h.bondLots, func() []string { return h.bondLots.bearers }); len(bearers) == 0 || bearers[0] != "Bearer "+pluginToken {
		t.Errorf("BondLotService callers = %v, want the plugin's token", bearers)
	}

	rec = h.serve(t, withSession(lotForm("EDO0636"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The bond lot was not added", "matures on 2036-05-04", `value="EDO0636"`, `value="2026-05-04"`)
	rec = h.serve(t, withSession(lotForm("COI0530"), ownerSession))
	wantStatus(t, rec, http.StatusServiceUnavailable)
	wantBody(t, rec, "its connector can&#39;t be reached now", `value="COI0530"`)
}

// TestAddBondLotRefusals checks that a read session can't add a lot and
// that a malformed form never reaches the core.
func TestAddBondLotRefusals(t *testing.T) {
	t.Parallel()
	h := bondHarness(t)
	rec := h.serve(t, withSession(lotForm("EDO0536"), readSession))
	wantStatus(t, rec, http.StatusForbidden)
	rec = h.serve(t, withSession(form("/accounts/"+accountID+"/bond-lots", "series=EDO0536&quantity=40"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The bond lot form is malformed.")
	if bearers := seen(h.bondLots, func() []string { return h.bondLots.bearers }); len(bearers) != 0 {
		t.Errorf("refused forms reached the core %d times", len(bearers))
	}
}
