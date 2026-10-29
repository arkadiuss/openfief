package ghostfolio_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// TestClientNeverDeletesUnfiltered checks that the client refuses, without
// sending anything, a bulk delete or a listing with no account to filter
// by: Ghostfolio reads either as every activity the user has (D-64).
func TestClientNeverDeletesUnfiltered(t *testing.T) {
	t.Parallel()
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	t.Cleanup(srv.Close)
	c := ghostfolio.NewClient(srv.Client(), srv.URL, nil)
	if err := c.DeleteActivitiesInAccounts(t.Context(), nil); !errors.Is(err, ghostfolio.ErrNoAccountFilter) {
		t.Errorf("DeleteActivitiesInAccounts(nil) = %v, want ErrNoAccountFilter", err)
	}
	if _, err := c.Activities(t.Context(), []string{}); !errors.Is(err, ghostfolio.ErrNoAccountFilter) {
		t.Errorf("Activities(none) = %v, want ErrNoAccountFilter", err)
	}
	if requests != 0 {
		t.Errorf("%d requests sent", requests)
	}
}

// TestClientErrors checks how the client reports Ghostfolio's answers: a
// deleted activity that is already gone counts as deleted, any other error
// status names the call and the status but never the response body, which
// can quote what was sent, and the security token is read for each login.
func TestClientErrors(t *testing.T) {
	t.Parallel()
	const secret = "synthetic-body-detail"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/gone"):
			http.Error(w, "Not Found", http.StatusNotFound)
		default:
			http.Error(w, secret, http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)
	reads := 0
	c := ghostfolio.NewClient(srv.Client(), srv.URL+"/", func() (string, error) {
		reads++
		return "synthetic-token", nil
	})
	if err := c.DeleteActivity(t.Context(), "gone"); err != nil {
		t.Errorf("DeleteActivity of a gone activity = %v, want nil", err)
	}
	checkHTTPError(t, c.DeleteActivity(t.Context(), "other"), secret)
	if err := c.Login(t.Context()); !errors.Is(err, ghostfolio.ErrHTTP) || reads != 1 {
		t.Errorf("Login = %v after %d token reads, want ErrHTTP after 1", err, reads)
	}
}

// checkHTTPError checks that err is an HTTPError for status 400 that wraps
// ErrHTTP and doesn't quote body.
func checkHTTPError(t *testing.T, err error, body string) {
	t.Helper()
	he, ok := errors.AsType[*ghostfolio.HTTPError](err)
	if !ok || he.Status != http.StatusBadRequest || !errors.Is(err, ghostfolio.ErrHTTP) || strings.Contains(err.Error(), body) {
		t.Errorf("error = %v, want an HTTPError 400 without the body", err)
	}
}
