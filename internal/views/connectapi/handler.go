package connectapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"

	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// requestIDBytes is the number of random bytes in a request ID.
const requestIDBytes = 8

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
// Every call needs a bearer token that verifier accepts.
func New(v *app.Views, verifier app.TokenVerifier, opts ...Option) (http.Handler, error) {
	c := config{log: slog.Default(), requestIDs: randomRequestID}
	for _, o := range opts {
		o(&c)
	}
	path, handler := viewsv1connect.NewReadServiceHandler(&readService{views: v, log: c.log},
		connect.WithInterceptors(&authInterceptor{verifier: verifier, log: c.log}))
	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{vanguard.NewService(path, handler)})
	if err != nil {
		return nil, fmt.Errorf("transcoding the read API: %w", err)
	}
	return withRequestLog(c, withSecurityHeaders(transcoder)), nil
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
