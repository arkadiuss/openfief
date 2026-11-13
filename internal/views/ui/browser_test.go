package ui

import (
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

const (
	// serveMargin is how long before the test's deadline the server stops.
	serveMargin = 5 * time.Second
	// browserSignIn is where the browser checks get the owner's session.
	browserSignIn = "/test/sign-in"
	// browserEndedSession is where they get a session the core rejects.
	browserEndedSession = "/test/ended-session"
	// browserAskModel is where they set up the Ask box's model (D-203): the
	// tiers query parameter lists the data tiers the model may receive, and
	// "off" takes the model away again.
	browserAskModel = "/test/ask-model"
)

// browserInboxItems adds an item of each kind the Inbox settles in place
// or shows evidence for (D-196): tickers to confirm, a flagged price check,
// a sync that keeps failing, an unmapped broker account and a flagged
// balance assertion, so the browser checks see each at phone width.
func browserInboxItems(h *harness) {
	tickerItems(h)
	priceCheckItem(h)
	syncItem(h, 40, "xtb-main")
	h.attention.add(&viewsv1.AttentionItem{
		Id: 60, Kind: kindBrokerUnmapped, Subject: "ibkr-main", Detail: "last seen 2026-03-31",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DATA_COMING_IN,
		ConnectorInstance: "ibkr-main",
	}, &viewsv1.AttentionItem{
		Id: 50, Kind: kindAssertion, Subject: accountID + "/2026-03-15/cash:EUR", Detail: "reported 100.00, derived 90.00",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_NUMBERS_TO_CHECK,
		AccountId: accountID, AccountIds: []string{accountID}, OwnerResolves: true,
	})
}

// TestServeForBrowser serves the UI over the in-memory fakes on the
// address in OPENFIEF_UI_SERVE (such as localhost:8099) until the test's
// deadline, for the browser checks in scripts/ui-browser.mjs (task
// ui:browser): screenshots at phone and desktop width, no sideways scroll,
// an installable manifest. Without the variable it skips, so go test never
// listens on a port.
func TestServeForBrowser(t *testing.T) {
	t.Parallel()
	addr := os.Getenv("OPENFIEF_UI_SERVE")
	if addr == "" {
		t.Skip("set OPENFIEF_UI_SERVE to serve the UI over the fakes for a browser")
	}
	h := newHarness(t)
	h.read.locked(func() { h.read.modelOn = true })
	// The actions D-183 protects ask for a passkey check first, which the
	// browser checks run with a virtual authenticator.
	h.auth.locked(func() { h.auth.stale = true })
	browserInboxItems(h)
	mux := http.NewServeMux()
	mux.Handle("/", h.handler)
	// The browser signs in here instead of with a passkey: the cookie comes
	// from a Set-Cookie header, as the UI's own sign-in sends it.
	mux.HandleFunc("GET "+browserSignIn, func(w http.ResponseWriter, r *http.Request) {
		setSession(w, ownerSession, issued(identityv1.SessionScope_SESSION_SCOPE_OWNER))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	// Here the browser gets a session the core no longer accepts, as one
	// that expired or was revoked (D-151's wipe when the session ends).
	mux.HandleFunc("GET "+browserEndedSession, func(w http.ResponseWriter, r *http.Request) {
		setSession(w, "session-ended", issued(identityv1.SessionScope_SESSION_SCOPE_OWNER))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	mux.HandleFunc("GET "+browserAskModel, func(w http.ResponseWriter, r *http.Request) {
		browserSetUpAskModel(h.askModel, r.URL.Query().Get("tiers"))
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: time.Second}
	stop := 10 * time.Minute
	if deadline, ok := t.Deadline(); ok {
		stop = time.Until(deadline) - serveMargin
	}
	timer := time.AfterFunc(stop, func() {
		if err := srv.Close(); err != nil {
			t.Log(err)
		}
	})
	defer timer.Stop()
	t.Logf("serving the UI over the fakes on http://%s", addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}
