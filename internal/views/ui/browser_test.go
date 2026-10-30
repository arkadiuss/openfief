package ui

import (
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
)

const (
	// serveMargin is how long before the test's deadline the server stops.
	serveMargin = 5 * time.Second
	// browserSignIn is where the browser checks get the owner's session.
	browserSignIn = "/test/sign-in"
)

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
	mux := http.NewServeMux()
	mux.Handle("/", h.handler)
	// The browser signs in here instead of with a passkey: the cookie comes
	// from a Set-Cookie header, as the UI's own sign-in sends it.
	mux.HandleFunc("GET "+browserSignIn, func(w http.ResponseWriter, r *http.Request) {
		setSession(w, ownerSession, issued(identityv1.SessionScope_SESSION_SCOPE_OWNER))
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
