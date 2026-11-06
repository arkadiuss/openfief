package browserproxy_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/browserproxy"
)

// TestProxyLogsNoQuery checks that a request the front door can't route to
// one view, or whose view doesn't answer, is logged by its path alone: a query, such
// as a question asked in the UI's Ask box (D-194), never reaches the
// core's log (GR-12).
func TestProxyLogsNoQuery(t *testing.T) {
	t.Parallel()
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	ui := make([]views.BrowserRoute, 0, 2)
	ui = append(ui, views.BrowserRoute{Instance: "ui", Prefix: "/"})
	doors := map[string]*app.FrontDoor{
		"two views hold a browser path": {Views: fakeViews{routes: append(ui, views.BrowserRoute{Instance: "gallery", Prefix: "/gallery"})}, Tokens: fakeTokens{}},
		"the plugin does not answer":    {Views: fakeViews{routes: ui, urls: map[string]string{"ui": closed.URL}}, Tokens: fakeTokens{}},
	}
	for name, door := range doors {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var log bytes.Buffer
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://localhost:8080/ask?q=synthetic-secret-question", http.NoBody)
			browserproxy.New(door, slog.New(slog.NewTextHandler(&log, nil))).ServeHTTP(httptest.NewRecorder(), req)
			if !strings.Contains(log.String(), "path=/ask") || strings.Contains(log.String(), "synthetic-secret-question") {
				t.Errorf("log = %q, want the path without the query", log.String())
			}
		})
	}
}
