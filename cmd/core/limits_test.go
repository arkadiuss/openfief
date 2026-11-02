package main

import (
	"errors"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// lazyBody is a body of zero bytes, made as it is read and never held, that
// counts what was read of it.
type lazyBody struct {
	left int64
	read atomic.Int64
}

func (b *lazyBody) Read(p []byte) (int, error) {
	if b.left <= 0 {
		return 0, io.EOF
	}
	n := min(int64(len(p)), b.left)
	clear(p[:n])
	b.left -= n
	b.read.Add(n)
	return int(n), nil
}

func (*lazyBody) Close() error { return nil }

// drain reads the whole body, answering 200, or 413 once a cap stops it.
func drain(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(io.Discard, r.Body); err != nil {
		if _, capped := errors.AsType[*http.MaxBytesError](err); capped {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TestPortsBoundBodies: each port refuses a compressed body with 415 and
// caps a body by its path before the API behind it reads it (D-184): 2 MiB
// for the core's own APIs, 65 MiB for the statement upload, the form the
// browser port proxies to the view plugin and the stream the plugin port
// takes. A declared length over the cap is refused without reading.
func TestPortsBoundBodies(t *testing.T) {
	t.Parallel()
	const over = maxAPIBodyBytes + 1<<20
	tests := []struct {
		name     string
		plugin   bool
		path     string
		header   http.Header
		size     int64
		declared int64
		want     int
		maxRead  int64
	}{
		{name: "read API over the cap", path: "/v1/accounts", size: over, declared: -1, want: http.StatusRequestEntityTooLarge, maxRead: maxAPIBodyBytes + 64<<10},
		{name: "read API declared over the cap", path: "/v1/accounts", size: over, declared: over, want: http.StatusRequestEntityTooLarge},
		{name: "upload form to the view plugin", path: "/accounts/acc-1/statements", size: over, declared: -1, want: http.StatusOK, maxRead: over},
		{name: "proxied body declared over the cap", path: "/accounts/acc-1/statements", size: maxUploadBodyBytes + 1, declared: maxUploadBodyBytes + 1, want: http.StatusRequestEntityTooLarge},
		{name: "compressed on the browser port", path: "/v1/accounts", header: http.Header{"Content-Encoding": {"gzip"}}, size: 1 << 10, declared: -1, want: http.StatusUnsupportedMediaType},
		{name: "statement upload stream", plugin: true, path: "/openfief.ingest.v1.StatementService/UploadStatement", size: over, declared: -1, want: http.StatusOK, maxRead: over},
		{name: "owner service over the cap", plugin: true, path: "/openfief.identity.v1.AuthService/WhoAmI", size: over, declared: -1, want: http.StatusRequestEntityTooLarge, maxRead: maxAPIBodyBytes + 64<<10},
		{name: "compressed stream on the plugin port", plugin: true, path: "/openfief.ingest.v1.StatementService/UploadStatement", header: http.Header{"Connect-Content-Encoding": {"gzip"}}, size: 1 << 10, declared: -1, want: http.StatusUnsupportedMediaType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &server{}
			var api http.Handler = http.HandlerFunc(drain)
			s.browser.Store(&api)
			s.plugin.Store(&api)
			h := s.browserHandler()
			if tt.plugin {
				h = s.pluginHandler()
			}
			body := &lazyBody{left: tt.size}
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, tt.path, body)
			req.ContentLength = tt.declared
			maps.Copy(req.Header, tt.header)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Errorf("POST %s = %d, want %d", tt.path, rec.Code, tt.want)
			}
			if got := body.read.Load(); got > tt.maxRead {
				t.Errorf("read %d bytes, want at most %d", got, tt.maxRead)
			}
		})
	}
}

// TestListenTimeouts: both ports bound how long a request may take to
// arrive and how long a connection may idle (D-184).
func TestListenTimeouts(t *testing.T) {
	t.Parallel()
	served := make(chan error, portCount)
	cfg := &serveConfig{Listen: "127.0.0.1:0", PluginListen: "127.0.0.1:0"}
	servers, err := listen(t.Context(), &server{}, cfg, served, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() {
		if err := closeAll(servers); err != nil {
			t.Errorf("closeAll: %v", err)
		}
	})
	for _, srv := range servers {
		if srv.ReadHeaderTimeout != readHeaderTimeout || srv.ReadTimeout != readTimeout || srv.IdleTimeout != idleTimeout {
			t.Errorf("server timeouts = %v, %v, %v; want %v, %v, %v",
				srv.ReadHeaderTimeout, srv.ReadTimeout, srv.IdleTimeout, readHeaderTimeout, readTimeout, idleTimeout)
		}
	}
}
