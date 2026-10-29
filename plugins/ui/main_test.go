package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// TestManifest checks that manifest.json, which the image carries as its
// signed label (D-42), asks for exactly what the UI needs: the view
// capability, the browser path "/" the core proxies to it (D-81), no egress
// host, no Secret, and a schema that takes no config.
func TestManifest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	got, err := plugins.ParseManifestLabel(string(data))
	if err != nil {
		t.Fatalf("ParseManifestLabel: %v", err)
	}
	want := plugins.Manifest{Capabilities: []plugins.Capability{plugins.CapabilityView}, BrowserPath: "/"}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty(), cmpopts.IgnoreFields(plugins.Manifest{}, "ConfigSchema")); diff != "" {
		t.Errorf("manifest (-want +got):\n%s", diff)
	}
	if len(got.ConfigSchema) == 0 {
		t.Error("the manifest has no config schema")
	}
}

// tokenFile writes a synthetic projected token to a file and returns it as
// the UI reads it.
func tokenFile(t *testing.T) openfief.TokenFile {
	t.Helper()
	path := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(path, []byte("synthetic-projected-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return openfief.TokenFile{Path: path}
}

// TestAuthClient checks what the UI presents to the AuthService: its
// projected token, read from the token file.
func TestAuthClient(t *testing.T) {
	t.Parallel()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
	}))
	t.Cleanup(srv.Close)
	c := authClient(tokenFile(t))
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if got != "Bearer synthetic-projected-token" || c.Timeout != coreTimeout {
		t.Errorf("Authorization = %q, timeout %v; want the projected token and %v", got, c.Timeout, coreTimeout)
	}
}

// TestUploadClient checks that statement uploads present the same token
// as the AuthService client, with the longer upload timeout.
func TestUploadClient(t *testing.T) {
	t.Parallel()
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
	}))
	t.Cleanup(srv.Close)
	c := uploadClient(tokenFile(t))
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if got != "Bearer synthetic-projected-token" || c.Timeout != uploadTimeout {
		t.Errorf("Authorization = %q, timeout %v; want the projected token and %v", got, c.Timeout, uploadTimeout)
	}
}

// TestReadClient checks that the read API client never goes through a
// proxy, sets no Authorization of its own and has a timeout.
func TestReadClient(t *testing.T) {
	t.Parallel()
	c := readClient()
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("transport is %T", c.Transport)
	}
	if tr.Proxy != nil {
		t.Error("the read API client uses a proxy")
	}
	if c.Timeout != coreTimeout {
		t.Errorf("timeout = %v, want %v", c.Timeout, coreTimeout)
	}
}
