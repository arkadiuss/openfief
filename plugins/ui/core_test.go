package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// errNotTheCore is the fake verifier's refusal.
var errNotTheCore = errors.New("synthetic: not the core")

// coreVerifier accepts only "core-token-for-ui".
type coreVerifier struct{}

func (coreVerifier) VerifyCore(_ context.Context, presented string) error {
	if presented != "core-token-for-ui" {
		return errors.Join(openfief.ErrNotCore, errNotTheCore)
	}
	return nil
}

// TestRequireCore checks that the UI serves only the core, which proxies
// browsers to it with its token (D-59, D-81), and probes on GET /healthz.
func TestRequireCore(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, method, path, auth string
		want                     int
	}{
		{name: "the core's token", method: http.MethodGet, path: "/login", auth: "Bearer core-token-for-ui", want: http.StatusOK},
		{name: "no token", method: http.MethodGet, path: "/login", want: http.StatusUnauthorized},
		{name: "another token", method: http.MethodPost, path: "/login/passkey/begin", auth: "Bearer ofsess_x_y", want: http.StatusUnauthorized},
		{name: "a probe", method: http.MethodGet, path: "/healthz", want: http.StatusOK},
		{name: "not a probe", method: http.MethodPost, path: "/healthz", want: http.StatusUnauthorized},
	}
	h := requireCore(coreVerifier{}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequestWithContext(t.Context(), tt.method, "http://localhost:8080"+tt.path, http.NoBody)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("%s %s = %d, want %d", tt.method, tt.path, rec.Code, tt.want)
			}
		})
	}
}
