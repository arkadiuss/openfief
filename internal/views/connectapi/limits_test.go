package connectapi_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/arkadiuss/openfief/internal/views/connectapi"
)

// listAccounts is the Connect path of ReadService.ListAccounts.
const listAccounts = "/openfief.views.v1.ReadService/ListAccounts"

// countingBody is a request body of size zero bytes, made as it is read and
// never held, that counts what the handler read of it.
type countingBody struct {
	left int64
	read atomic.Int64
}

func (b *countingBody) Read(p []byte) (int, error) {
	if b.left <= 0 {
		return 0, io.EOF
	}
	n := min(int64(len(p)), b.left)
	clear(p[:n])
	b.left -= n
	b.read.Add(n)
	return int(n), nil
}

func (*countingBody) Close() error { return nil }

// count returns how many bytes were read, none of a nil body.
func (b *countingBody) count() int64 {
	if b == nil {
		return 0
	}
	return b.read.Load()
}

// gzipBomb returns n zero bytes gzipped: about a thousandth of n.
func gzipBomb(t *testing.T, n int64) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := io.CopyN(zw, &countingBody{left: n}, n); err != nil {
		t.Fatalf("gzip: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip: %v", err)
	}
	return buf.Bytes()
}

// checkRefused fails t unless rec has status want (any for 0) and a body
// saying wantBody, and never says the request was unauthenticated: a
// refusal must come before the token is checked.
func checkRefused(t *testing.T, rec *httptest.ResponseRecorder, want int, wantBody string) {
	t.Helper()
	if want != 0 && rec.Code != want {
		t.Errorf("status = %d, want %d: %s", rec.Code, want, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), wantBody) {
		t.Errorf("body = %s, want it to say %s", rec.Body, wantBody)
	}
	if rec.Code == http.StatusUnauthorized || strings.Contains(rec.Body.String(), "unauthenticated") {
		t.Errorf("the request reached the authentication: %d %s", rec.Code, rec.Body)
	}
}

// TestReadAPIBoundsBodies: before any token is checked or message decoded
// (D-184), the read API refuses a compressed request, such as a small gzip
// body that inflates to megabytes, with 415 without reading it; refuses a
// declared length over its cap with 413 without reading it; and stops
// reading a body of unknown length just past MaxRequestBytes, answering
// RESOURCE_EXHAUSTED. No request carries a token, so none gets as far as
// the authentication.
func TestReadAPIBoundsBodies(t *testing.T) {
	t.Parallel()
	const big = 8 * connectapi.MaxRequestBytes
	bomb := gzipBomb(t, big)
	tests := []struct {
		name, method, target string
		header               http.Header
		body                 *countingBody
		declared             int64
		raw                  []byte
		wantStatus           int
		wantBody             string
		maxRead              int64
	}{
		{
			name: "gzip bomb, Connect unary", method: http.MethodPost, target: listAccounts,
			header: http.Header{"Content-Type": {"application/proto"}, "Content-Encoding": {"gzip"}},
			raw:    bomb, wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "gzip bomb, REST", method: http.MethodGet, target: "/v1/accounts",
			header: http.Header{"Content-Encoding": {"gzip"}},
			raw:    bomb, wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "compressed Connect stream", method: http.MethodPost, target: listAccounts,
			header: http.Header{"Content-Type": {"application/connect+proto"}, "Connect-Content-Encoding": {"gzip"}},
			body:   &countingBody{left: big}, wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "compressed gRPC-Web", method: http.MethodPost, target: listAccounts,
			header: http.Header{"Content-Type": {"application/grpc-web+proto"}, "Grpc-Encoding": {"gzip"}},
			body:   &countingBody{left: big}, wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "declared length over the cap", method: http.MethodPost, target: listAccounts,
			header: http.Header{"Content-Type": {"application/proto"}},
			body:   &countingBody{left: big}, declared: big, wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name: "unknown length over the cap", method: http.MethodPost, target: listAccounts,
			header: http.Header{"Content-Type": {"application/proto"}, "Connect-Protocol-Version": {"1"}},
			body:   &countingBody{left: big}, declared: -1,
			wantBody: "resource_exhausted", maxRead: 2 * connectapi.MaxRequestBytes,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var body io.ReadCloser = tt.body
			length := tt.declared
			if tt.raw != nil {
				body, length = io.NopCloser(bytes.NewReader(tt.raw)), int64(len(tt.raw))
			}
			req := httptest.NewRequestWithContext(context.Background(), tt.method, tt.target, body)
			req.Header, req.ContentLength = tt.header, length
			rec := httptest.NewRecorder()
			newAPI(t, newLedger(t)).handler.ServeHTTP(rec, req)
			checkRefused(t, rec, tt.wantStatus, tt.wantBody)
			if read := tt.body.count(); read > tt.maxRead {
				t.Errorf("read %d bytes of the body, want at most %d", read, tt.maxRead)
			}
		})
	}
}
