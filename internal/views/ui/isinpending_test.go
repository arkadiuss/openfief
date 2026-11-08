package ui

import (
	"net/http"
	"strings"
	"testing"
)

// pendingPill is the pill next to a bond held under its provisional ISIN.
const pendingPill = `<span class="pill pill-neutral" title="A provisional ISIN: it moves to the registry's ISIN when the Ministry of Finance's data lists the series">pending`

// TestISINPendingPill checks that a treasury bond series held under its
// provisional ISIN carries a "pending" pill, once, wherever its ISIN shows:
// the holdings behind an allocation bar, the account's holdings and its open
// lots, where an entered lot appears, and the instrument page; and that an
// instrument under its registry ISIN carries none (D-214).
func TestISINPendingPill(t *testing.T) {
	t.Parallel()
	h := bondHarness(t)
	for _, tt := range []struct {
		name, path string
		want       int
	}{
		{"holdings behind a bar", "/holdings?by=currency&key=PLN&date=2026-03-31", 1},
		{"account holdings", "/accounts/" + accountID + "?tab=holdings", 1},
		{"account open lots", "/accounts/" + accountID + "?tab=lots", 1},
		{"instrument page", "/instruments/" + pendingBond, 1},
		{"a bond under its registry ISIN", "/instruments/" + assumedBond, 0},
		{"a listed instrument", "/instruments/XX0000000001", 0},
	} {
		rec := h.serve(t, withSession(request(http.MethodGet, tt.path, ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		if got := strings.Count(rec.Body.String(), pendingPill); got != tt.want {
			t.Errorf("%s: %d pending pills, want %d; body:\n%s", tt.name, got, tt.want, rec.Body.String())
		}
	}
	wantBody(t, h.serve(t, withSession(request(http.MethodGet, "/instruments/"+pendingBond, ""), readSession)),
		`ISIN <span class="mono">`+pendingBond+`</span> `+pendingPill)
	other := newHarness(t)
	rec := other.serve(t, withSession(request(http.MethodGet, "/accounts/"+accountID+"?tab=lots", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, pendingPill)
}
