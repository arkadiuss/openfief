package ui

import (
	"context"
	"net/http"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	valuationv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1"
)

// SetChartStart saves the owner's charts' start where the fake read API's
// GetPerformance answers it from (D-167); an empty start clears it.
func (f *fakeSettings) SetChartStart(_ context.Context, r *connect.Request[valuationv1.SetChartStartRequest]) (*connect.Response[valuationv1.SetChartStartResponse], error) {
	if r.Msg.GetSessionToken() != ownerSession {
		return nil, codeErr(connect.CodePermissionDenied, errNotOwnerFake)
	}
	if start := r.Msg.GetStart(); start != "" {
		if _, err := kernel.ParseDate(start); err != nil {
			return nil, codeErr(connect.CodeInvalidArgument, err)
		}
	}
	f.read.locked(func() {
		if f.read.chartStarts == nil {
			f.read.chartStarts = map[string]string{}
		}
		if r.Msg.GetStart() == "" {
			delete(f.read.chartStarts, r.Msg.GetAccountId())
			return
		}
		f.read.chartStarts[r.Msg.GetAccountId()] = r.Msg.GetStart()
	})
	return connect.NewResponse(&valuationv1.SetChartStartResponse{}), nil
}

// savedStart returns the charts' start the fake stored for account.
func (h *harness) savedStart(account string) string {
	var got string
	h.read.locked(func() { got = h.read.chartStarts[account] })
	return got
}

// TestChartStartSaved checks the owner's charts' start as a setting
// (D-167): the overview's date field saves it, the page then starts its
// charts on it without naming it in the URL, going back to the default
// clears it, an account's is saved apart from the portfolio's, and a
// start in the URL still previews without saving.
func TestChartStartSaved(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	page := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, page, http.StatusOK)
	wantBody(t, page, `method="post" action="/chart-start"`, `<button type="submit">Save</button>`, "until you go back to the default")
	wantNoBody(t, page, `name="clear"`)

	rec := h.serve(t, withSession(form("/chart-start", "account=&date=2026-03-31&start=2026-03-31"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/?date=2026-03-31" {
		t.Errorf("saving redirects to %q, want /?date=2026-03-31", loc)
	}
	if got := h.savedStart(""); got != "2026-03-31" {
		t.Errorf("saved start = %q, want 2026-03-31", got)
	}
	page = h.serve(t, withSession(request(http.MethodGet, "/?date=2026-03-31", ""), ownerSession))
	wantBody(t, page, `name="start" type="date" value="2026-03-31"`, `name="clear" value="1"`,
		`Back to the default, <span class="mono">2026-03-30</span>`, "time-weighted return as an index, 100 on 2026-03-31")
	h.read.locked(func() {
		if n := len(h.read.performance); n == 0 || h.read.performance[n-1].GetFrom() != "" {
			t.Errorf("GetPerformance requests = %v, want the saved start left to the core", h.read.performance)
		}
	})

	wantStatus(t, h.serve(t, withSession(form("/chart-start", "account=&date=2026-03-31&clear=1"), ownerSession)), http.StatusSeeOther)
	if got := h.savedStart(""); got != "" {
		t.Errorf("saved start after going back to the default = %q, want none", got)
	}

	rec = h.serve(t, withSession(form("/chart-start", "account="+accountID+"&date=2026-03-31&start=2026-03-31"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if loc := rec.Header().Get("Location"); loc != "/accounts/"+accountID+"?date=2026-03-31" {
		t.Errorf("saving an account's start redirects to %q", loc)
	}
	if got, portfolio := h.savedStart(accountID), h.savedStart(""); got != "2026-03-31" || portfolio != "" {
		t.Errorf("saved starts: account %q, portfolio %q; want only the account's", got, portfolio)
	}
	account := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), ownerSession))
	wantBody(t, account, `name="account" value="`+accountID+`"`, `name="clear" value="1"`)

	preview := h.serve(t, withSession(request(http.MethodGet, "/?start=2026-03-31&date=2026-03-31", ""), ownerSession))
	wantBody(t, preview, `href="/?date=2026-03-31"`, "shows a start without saving it", "100 on 2026-03-31")
	if got := h.savedStart(""); got != "" {
		t.Errorf("a preview saved %q", got)
	}
}

// TestChartStartSaveRefused checks that only the owner saves a charts'
// start, and that a malformed one is refused before the core is asked.
func TestChartStartSaveRefused(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	wantStatus(t, h.serve(t, withSession(form("/chart-start", "date=2026-03-31&start=2026-03-31"), readSession)), http.StatusForbidden)
	wantStatus(t, h.serve(t, withSession(form("/chart-start", "date=2026-03-31&start=31.03.2026"), ownerSession)), http.StatusBadRequest)
	if got := h.savedStart(""); got != "" {
		t.Errorf("a refused save stored %q", got)
	}
	page := h.serve(t, withSession(request(http.MethodGet, "/", ""), readSession))
	wantBody(t, page, `method="get" action="/"`, `<button type="submit">Show</button>`)
	wantNoBody(t, page, `action="/chart-start"`)
}

// TestAccountOpenedOnSeedingDayForm checks the account form's checkbox
// for an account opened on its seeding day (D-168): a new account takes
// it, a change sends it checked or unchecked, and a form without it keeps
// the account's mark.
func TestAccountOpenedOnSeedingDayForm(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	page := h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), ownerSession))
	wantBody(t, page, `name="opened_on_seeding_day" value="1"`, "Opened on its seeding day")
	wantStatus(t, h.serve(t, withSession(form("/accounts", "name=Opened&sleeve=core&opened_on_seeding_day_shown=1&opened_on_seeding_day=1"), ownerSession)), http.StatusSeeOther)
	if created := h.accounts.lastCreate(); !created.GetOpenedOnSeedingDay() {
		t.Errorf("CreateAccount = %v, want it marked", created)
	}
	edits := []struct {
		name, body string
		// want is the mark sent: "" for none, which keeps the account's.
		want string
	}{
		{name: "checked", body: "name=Renamed&sleeve=core&opened_on_seeding_day_shown=1&opened_on_seeding_day=1", want: "true"},
		{name: "unchecked", body: "name=Renamed&sleeve=core&opened_on_seeding_day_shown=1", want: "false"},
		{name: "not on the form", body: "name=Renamed&sleeve=core"},
	}
	for _, e := range edits {
		wantStatus(t, h.serve(t, withSession(form("/accounts/"+accountID+"/edit", e.body), ownerSession)), http.StatusSeeOther)
		updated := h.accounts.lastUpdate()
		got := ""
		if updated.OpenedOnSeedingDay != nil {
			got = map[bool]string{true: "true", false: "false"}[updated.GetOpenedOnSeedingDay()]
		}
		if got != e.want {
			t.Errorf("%s: UpdateAccount sent the mark %q, want %q", e.name, got, e.want)
		}
	}
}
