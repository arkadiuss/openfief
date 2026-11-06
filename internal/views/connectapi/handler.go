package connectapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"

	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

const (
	// requestIDBytes is the number of random bytes in a request ID.
	requestIDBytes = 8
	// MaxRequestBytes caps one request message: the read API's requests
	// carry only a few IDs, dates and a page token (D-184).
	MaxRequestBytes = 1 << 20
	// maxBodyBytes caps a request's body: one message and its envelope.
	maxBodyBytes = MaxRequestBytes + 1<<10
)

// Option configures New.
type Option func(*config)

// config is what Options set.
type config struct {
	log        *slog.Logger
	requestIDs func() string
}

// WithLogger logs to l instead of slog.Default().
func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.log = l }
}

// WithRequestIDs names requests with ids instead of random hex strings.
func WithRequestIDs(ids func() string) Option {
	return func(c *config) { c.requestIDs = ids }
}

// New returns the read API's HTTP handler over v: the Connect, gRPC and
// gRPC-Web protocols at /openfief.views.v1.ReadService/ and, transcoded by
// Vanguard from the same handlers, REST under /v1/ (D-50).
// Every call needs a bearer token that verifier accepts. Before any token is
// checked or any message decoded, a compressed request is refused with 415
// and a body or message over MaxRequestBytes with RESOURCE_EXHAUSTED or 413
// (D-184): the token is checked only once the message is read.
func New(v *app.Views, verifier app.TokenVerifier, opts ...Option) (http.Handler, error) {
	c := config{log: slog.Default(), requestIDs: randomRequestID}
	for _, o := range opts {
		o(&c)
	}
	path, handler := viewsv1connect.NewReadServiceHandler(&readService{views: v, log: c.log},
		connect.WithInterceptors(&authInterceptor{verifier: verifier, log: c.log}, &namesInterceptor{views: v, log: c.log}),
		connect.WithReadMaxBytes(MaxRequestBytes))
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{vanguard.NewService(path, handler)})
	if err != nil {
		return nil, fmt.Errorf("transcoding the read API: %w", err)
	}
	return withRequestLog(c, withSecurityHeaders(withBoundedBody(transcoder))), nil
}

// withBoundedBody refuses a request whose body is compressed, with 415, and
// caps the rest at maxBodyBytes, before next decodes anything (D-184).
// Vanguard would decompress a REST body without a limit, and a small gzip
// body can inflate to gigabytes.
func withBoundedBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Compressed(r.Header) {
			http.Error(w, "compressed requests are not accepted", http.StatusUnsupportedMediaType)
			return
		}
		if r.ContentLength > maxBodyBytes {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		next.ServeHTTP(w, r)
	})
}

// Compressed reports whether h declares a compressed request body in any
// protocol the core serves: Content-Encoding (HTTP and Connect unary),
// Connect-Content-Encoding (Connect streams) or Grpc-Encoding (gRPC and
// gRPC-Web). "identity" is not compression.
func Compressed(h http.Header) bool {
	for _, name := range []string{"Content-Encoding", "Connect-Content-Encoding", "Grpc-Encoding"} {
		if v := h.Get(name); v != "" && !strings.EqualFold(v, "identity") {
			return true
		}
	}
	return false
}

// randomRequestID returns a new random request ID.
func randomRequestID() string {
	b := make([]byte, requestIDBytes)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// requestIDKey is the context key of the request ID.
type requestIDKey struct{}

// requestID returns the ID withRequestLog gave the request of ctx.
func requestID(ctx context.Context) string {
	id, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

// statusRecorder remembers the status a handler wrote.
type statusRecorder struct {
	http.ResponseWriter

	status int
}

// WriteHeader records status and writes it.
func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// Flush flushes the response if the underlying writer can, which streaming
// protocols need.
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the underlying writer, for http.ResponseController.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// withRequestLog names each request with a request ID, returns it in the
// X-Request-ID header and logs the method and status under it. The path is
// not logged: it can name an account.
func withRequestLog(c config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := c.requestIDs()
		w.Header().Set("X-Request-ID", id)
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
		c.log.LogAttrs(r.Context(), slog.LevelInfo, "read API request",
			slog.String("request_id", id), slog.String("method", r.Method), slog.Int("status", rec.status))
	})
}

// withSecurityHeaders sets security headers on every response: the API
// serves financial data to scripts and plugins, never pages, so nothing may
// cache, frame, sniff or embed it.
func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// logFailure logs that the call of ctx failed while doing what, under its
// request ID only: error texts can carry account IDs or amounts.
func logFailure(ctx context.Context, log *slog.Logger, what string) {
	log.LogAttrs(ctx, slog.LevelError, "read API call failed",
		slog.String("request_id", requestID(ctx)), slog.String("while", what))
}
