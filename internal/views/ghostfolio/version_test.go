package ghostfolio_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// TestCheckVersion checks the supported range of Ghostfolio versions
// (D-188): 3.93.x is supported, anything else, a development build's "dev"
// or a version that isn't MAJOR.MINOR.PATCH is refused.
func TestCheckVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version string
		ok      bool
	}{
		{"3.93.0", true},
		{"3.93.12", true},
		{"3.92.9", false},
		{"3.94.0", false},
		{"4.0.0", false},
		{"2.93.0", false},
		{"dev", false},
		{"", false},
		{"3.93", false},
		{"3.93.0-beta.1", false},
		{"3.93.-1", false},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()
			err := ghostfolio.CheckVersion(tt.version)
			if (err == nil) != tt.ok || (err != nil && !errors.Is(err, ghostfolio.ErrUnsupportedVersion)) {
				t.Errorf("CheckVersion(%q) = %v, want ok %v", tt.version, err, tt.ok)
			}
		})
	}
}

// TestClientVersion checks that the client reads Ghostfolio's version from
// its export's meta, asking for an activity ID no activity has, so the
// export carries none of the owner's activities (D-188).
func TestClientVersion(t *testing.T) {
	t.Parallel()
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/export" {
			http.NotFound(w, r)
			return
		}
		query = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, `{"meta":{"date":"2026-11-03T00:00:00.000Z","version":"3.93.0"},"activities":[]}`); err != nil {
			t.Errorf("writing the export: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	got, err := ghostfolio.NewClient(srv.Client(), srv.URL, nil).Version(t.Context())
	if err != nil || got != "3.93.0" {
		t.Errorf("Version = %q, %v; want 3.93.0", got, err)
	}
	if query != "activityIds=00000000-0000-0000-0000-000000000000" {
		t.Errorf("export query %q, want only the activity ID no activity has", query)
	}
}

// TestClientFollowsNoRedirect checks that the client never follows a
// redirect (D-188): a 307 or 308 on the login would send its body, with
// the security token, again to wherever it points. The request fails with
// ErrRedirect and the redirect's target is never called; the caller's own
// HTTP client keeps its redirect policy.
func TestClientFollowsNoRedirect(t *testing.T) {
	t.Parallel()
	var followed atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { followed.Add(1) }))
	t.Cleanup(target.Close)
	for _, status := range []int{http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Location", target.URL+"/api/v1/auth/anonymous")
			w.WriteHeader(status)
		}))
		t.Cleanup(srv.Close)
		hc := srv.Client()
		c := ghostfolio.NewClient(hc, srv.URL, func() (string, error) { return "synthetic-token", nil })
		if err := c.Login(t.Context()); !errors.Is(err, ghostfolio.ErrRedirect) {
			t.Errorf("Login answered with %d = %v, want ErrRedirect", status, err)
		}
		if err := c.DeleteActivitiesInAccounts(t.Context(), []string{"acc"}); !errors.Is(err, ghostfolio.ErrRedirect) {
			t.Errorf("bulk delete answered with %d = %v, want ErrRedirect", status, err)
		}
		if hc.CheckRedirect != nil {
			t.Error("NewClient changed the caller's HTTP client")
		}
	}
	if n := followed.Load(); n != 0 {
		t.Errorf("the redirect's target was called %d times", n)
	}
}

// TestSyncRefusesUnsupportedVersion checks that the view changes and
// deletes nothing in a Ghostfolio outside SupportedVersions, nor in one
// whose version it can't read (D-188): it logs in, reads the version and
// stops.
func TestSyncRefusesUnsupportedVersion(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"3.94.0", "dev", "x.y.z"} {
		g := fakeGhostfolio{ghostfolioModel: newModel(), version: version}
		_, err := ghostfolio.NewSyncer(syntheticRecord(), &g, instance, map[ledger.ISIN]ghostfolio.Listing{}).Sync(t.Context())
		if !errors.Is(err, ghostfolio.ErrUnsupportedVersion) {
			t.Errorf("version %q: Sync = %v, want ErrUnsupportedVersion", version, err)
		}
		if calls := strings.Join(g.calls, " "); calls != "login version" {
			t.Errorf("version %q: calls %q, want only login and version", version, calls)
		}
	}
}
