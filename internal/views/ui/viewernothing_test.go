package ui

import (
	"net/http"
	"testing"
)

// TestViewerNothingSharedKeepsPlugins checks D-86 and D-165 for a viewer who was
// granted nothing: the Plugins page lists the installed plugins with their
// digests and signers exactly as the owner sees them, the Sleeves page says
// nothing has been shared and still offers the Plugins tab, and the Inbox says
// the same, with the empty state of the other pages and no empty list.
func TestViewerNothingSharedKeepsPlugins(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, session := range []string{emptyViewerSession, viewerSession, ownerSession} {
		rec := h.serve(t, withSession(request(http.MethodGet, "/plugins", ""), session))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec,
			"Installed plugins", "ibkr-main", "sha256:a1a1", "openfief-ci", "ui-local", "sha256:b2b2", "prices-main", "sha256:c3c3",
		)
		wantNoBody(t, rec, "Nothing has been shared with you yet")
	}
	rec := h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), emptyViewerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Nothing has been shared with you yet", `href="/plugins"`)

	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), emptyViewerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Nothing has been shared with you yet", "Viewer access", "Bob")
	wantNoBody(t, rec, `href="/settings"`, `href="/access"`, "Every ticker is confirmed")
	rec = h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), viewerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "Nothing has been shared with you yet")
}
