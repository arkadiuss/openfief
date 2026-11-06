package ui

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestAccountFrame checks an account's page (D-195, D-196): the breadcrumb,
// the name with its sleeve, wrapper and valuation pills, the four figure
// cards with the time-weighted return since the charts' start and the day
// it is valued through, and the six tabs, each a link that works without
// JavaScript and an htmx swap, the holdings' open by default and an
// unknown one falling back to them.
func TestAccountFrame(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID, ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`<a href="/">Overview</a>`, `<a href="/accounts">Accounts</a>`, `<span aria-current="page">Synthetic brokerage</span>`,
		"<h1>Synthetic brokerage</h1>", `<span class="pill pill-neutral">Sleeve Core</span>`,
		"Gross, with cash", "Net", "Liquidation", "TWR since 2026-03-30", "&#43;1,5%", `valued through <span class="mono">2026-03-31</span>`,
		`aria-current="page">Holdings</a>`, `id="holdings-title"`,
	)
	for _, tab := range accountTabs() {
		wantBody(t, rec, `href="/accounts/acc-synthetic-1?date=2026-03-31&amp;tab=`+tab.Key+`"`,
			`hx-get="/accounts/acc-synthetic-1/tab?date=2026-03-31&amp;tab=`+tab.Key+`"`, ">"+tab.Label+"</a>")
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID+"?tab=nonsense", ""), readSession))
	wantBody(t, rec, `aria-current="page">Holdings</a>`)
}

// TestAccountTabFragment checks that htmx gets only the tabs showing the
// one asked for, the account's settings among them with its connector
// instances, and that a browser without htmx gets the whole page on it.
func TestAccountTabFragment(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/accounts/"+accountID+"/tab?tab=settings&date=2026-03-31", ""), readSession)))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `<section id="account-tabs"`, `aria-current="page">Account settings</a>`, `id="settings-title"`,
		`<span class="mono">ibkr-main</span>`, "synthetic-broker", "no tax wrapper", `<a href="/accounts">Accounts</a> page`)
	wantNoBody(t, rec, "<html", "Gross, with cash")
	rec = h.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID+"/tab?tab=settings", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<html", "Gross, with cash", `aria-current="page">Account settings</a>`)
}

// TestSyncPill checks the account header's sync pill: a connector instance
// bound to it that keeps failing names itself, otherwise the latest
// success's day, and none without one.
func TestSyncPill(t *testing.T) {
	t.Parallel()
	statuses := []*viewsv1.SyncStatus{
		{ConnectorInstance: "ibkr-main", LastSuccessAt: "2026-03-30T18:00:00Z"},
		{ConnectorInstance: "nbp-fx", LastSuccessAt: "2026-03-31T18:00:00Z"},
		{ConnectorInstance: "xtb-main", LastSuccessAt: "2026-03-01T07:00:00Z", KeepsFailing: true},
	}
	tests := []struct {
		name      string
		instances []string
		want      *statusPill
	}{
		{"synced", []string{"ibkr-main"}, &statusPill{"Synced 2026-03-30", "pill-ok"}},
		{"one keeps failing", []string{"ibkr-main", "xtb-main"}, &statusPill{"Sync failing · xtb-main", "pill-danger"}},
		{"never synced", []string{"flex-main"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, syncPill(tt.instances, statuses)); diff != "" {
				t.Errorf("syncPill (-want +got):\n%s", diff)
			}
		})
	}
}

// TestHoldingsBehindBar checks the holdings behind an allocation bar of a
// currency or an asset class (D-195): each valued holding with its account
// and a link to its instrument, the cash of the group, and a refusal of
// any other dimension.
func TestHoldingsBehindBar(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/holdings?by=currency&key=EUR&date=2026-03-31", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "<h1>Holdings · EUR</h1>", `href="/instruments/XX0000000001?account=acc-synthetic-1"`,
		"250,00 EUR", "1\u202f000,00 PLN", `<a href="/accounts/acc-synthetic-1?tab=cash">Synthetic brokerage</a>`, "12\u202f345,67 EUR",
		`<a href="/?by=currency&amp;date=2026-03-31&amp;over=allocation#over-time">Allocation by currency</a>`)
	rec = h.serve(t, withSession(request(http.MethodGet, "/holdings?by=asset-class&key=listed&date=2026-03-31", ""), readSession))
	wantBody(t, rec, "<h1>Holdings · listed</h1>", `href="/instruments/XX0000000001?account=acc-synthetic-1"`)
	wantNoBody(t, rec, "12\u202f345,67 EUR")
	rec = h.serve(t, withSession(request(http.MethodGet, "/holdings?by=asset-class&key=cash&date=2026-03-31", ""), readSession))
	wantBody(t, rec, "No valued holding on this date.", "12\u202f345,67 EUR")
	for _, q := range []string{"by=sleeve&key=core", "by=currency", "by=currency&key=EUR&date=yesterday"} {
		rec = h.serve(t, withSession(request(http.MethodGet, "/holdings?"+q, ""), readSession))
		wantStatus(t, rec, http.StatusBadRequest)
	}
}
