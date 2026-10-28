package openfief_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// authRecorder records the Authorization header of every request.
type authRecorder struct {
	mu    sync.Mutex
	auths []string
}

func (a *authRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.auths = append(a.auths, r.Header.Get("Authorization"))
	w.WriteHeader(http.StatusOK)
}

func (a *authRecorder) seen() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return slices.Clone(a.auths)
}

// get sends a GET to url with client and returns its error; a response other
// than 200 fails the test.
func get(t *testing.T, client *http.Client, url string) error {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	return nil
}

func TestCoreHTTPClientSendsCurrentToken(t *testing.T) {
	t.Parallel()
	core := &authRecorder{}
	srv := httptest.NewServer(core)
	t.Cleanup(srv.Close)
	path := filepath.Join(t.TempDir(), "token")
	client := openfief.CoreHTTPClient(openfief.TokenFile{Path: path})
	for _, token := range []string{"synthetic-token-1", "synthetic-token-2"} {
		writeToken(t, path, token)
		if err := get(t, client, srv.URL); err != nil {
			t.Fatalf("GET: %v", err)
		}
	}
	want := []string{"Bearer synthetic-token-1", "Bearer synthetic-token-2"}
	if got := core.seen(); !slices.Equal(got, want) {
		t.Errorf("Authorization headers = %q, want %q", got, want)
	}
}

func TestCoreHTTPClientWithoutToken(t *testing.T) {
	t.Parallel()
	core := &authRecorder{}
	srv := httptest.NewServer(core)
	t.Cleanup(srv.Close)
	client := openfief.CoreHTTPClient(openfief.TokenFile{Path: filepath.Join(t.TempDir(), "missing")})
	if err := get(t, client, srv.URL); err == nil {
		t.Fatal("GET without a token succeeded")
	}
	if got := core.seen(); len(got) != 0 {
		t.Errorf("the core got %d requests without a token, want none", len(got))
	}
}
