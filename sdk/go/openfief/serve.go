package openfief

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"

	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

// AudiencePrefix starts the audience of the token the core presents to a
// plugin instance; the instance's name follows it (D-59).
const AudiencePrefix = "openfief-plugin:"

const (
	// readHeaderTimeout bounds how long a client may take to send its headers.
	readHeaderTimeout = 10 * time.Second
	// readTimeout bounds how long a client may take to send a whole request,
	// such as a statement of 64 MiB in chunks.
	readTimeout = 5 * time.Minute
	// idleTimeout closes keep-alive connections idle this long.
	idleTimeout = 2 * time.Minute
)

// Limits of what NewConnectorServer reads from one request, before the
// caller is checked or a message decoded.
const (
	// MaxMessageBytes caps one message of a call: a sync request with the
	// accounts' holdings, bond lots, or one chunk of a file. An option
	// passed to NewConnectorServer may change it.
	MaxMessageBytes = 4 << 20
	// MaxBodyBytes caps a request's body: an ImportFile stream of a 64 MiB
	// statement, the largest the core sends, with its header and envelopes.
	MaxBodyBytes = 65 << 20
)

// Serving errors.
var (
	// ErrNotCore reports a call that does not carry a token of the core's.
	ErrNotCore = errors.New("caller is not the OpenFief core")
	// ErrNoFileHeader reports an ImportFile stream whose first message is
	// not a header, or that has a second one.
	ErrNoFileHeader = errors.New("import file stream must start with exactly one header")
	// ErrFileTooLarge reports a file larger than the limit the connector set.
	ErrFileTooLarge = errors.New("file too large")
)

// InstanceAudience returns the audience of the token the core presents to
// the plugin instance named instance (D-48, D-59).
func InstanceAudience(instance string) string { return AudiencePrefix + instance }

// CoreVerifier decides whether a bearer token is the core's token for this
// plugin instance: a ServiceAccount token of the core with the instance's
// audience (InstanceAudience), unexpired. It returns an error wrapping
// ErrNotCore for any other token and never puts the token in an error.
type CoreVerifier interface {
	VerifyCore(ctx context.Context, token string) error
}

// RequireCore returns h behind a check that the caller is the core (D-59): a
// request without an "Authorization: Bearer" token that v accepts gets 401
// Unauthorized and never reaches h.
func RequireCore(v CoreVerifier, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok || v.VerifyCore(r.Context(), token) != nil {
			http.Error(w, ErrNotCore.Error(), http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// bearerToken returns the token of an "Authorization: Bearer <token>" value.
func bearerToken(value string) (string, bool) {
	scheme, token, ok := strings.Cut(value, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" || strings.ContainsAny(token, " \t") {
		return "", false
	}
	return token, true
}

// NewConnectorServer returns a server on addr, such as ListenAddress gives,
// that serves svc's connector contract (D-49) to the core only (RequireCore).
// It speaks HTTP/1.1 and HTTP/2 without TLS, since ImportFile streams. It
// reads a body of at most MaxBodyBytes and a message of at most
// MaxMessageBytes, unless opts say otherwise, and bounds how long a request
// may take to arrive. Each call runs in a server span that continues the
// core's trace (telemetry.ServerHandler).
func NewConnectorServer(addr string, svc connectorv1connect.ConnectorServiceHandler, v CoreVerifier, opts ...connect.HandlerOption) *http.Server {
	opts = append([]connect.HandlerOption{connect.WithReadMaxBytes(MaxMessageBytes)}, opts...)
	path, h := connectorv1connect.NewConnectorServiceHandler(svc, opts...)
	return newCoreServer(addr, path, h, v)
}

// newCoreServer returns a server on addr that serves h at path to the core
// only (RequireCore), over HTTP/1.1 and unencrypted HTTP/2, with the
// limits above and a server span per call.
func newCoreServer(addr, path string, h http.Handler, v CoreVerifier) *http.Server {
	mux := http.NewServeMux()
	mux.Handle(path, RequireCore(v, h))
	var p http.Protocols
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	return &http.Server{
		Addr: addr, Handler: telemetry.ServerHandler(http.MaxBytesHandler(mux, MaxBodyBytes)), Protocols: &p,
		ReadHeaderTimeout: readHeaderTimeout, ReadTimeout: readTimeout, IdleTimeout: idleTimeout,
	}
}

// ReceiveFile reads an ImportFile stream: its header, then its chunks in
// order, refusing a file larger than limit bytes.
func ReceiveFile(stream *connect.ClientStream[connectorv1.ImportFileRequest], limit int) (*connectorv1.FileHeader, []byte, error) {
	var (
		header *connectorv1.FileHeader
		data   []byte
	)
	for stream.Receive() {
		msg := stream.Msg()
		switch {
		case header == nil && msg.GetHeader() != nil:
			header = msg.GetHeader()
		case header == nil || msg.GetHeader() != nil:
			return nil, nil, ErrNoFileHeader
		case len(data)+len(msg.GetChunk()) > limit:
			return nil, nil, fmt.Errorf("%w: over %d bytes", ErrFileTooLarge, limit)
		default:
			data = append(data, msg.GetChunk()...)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, nil, fmt.Errorf("receive file: %w", err)
	}
	if header == nil {
		return nil, nil, ErrNoFileHeader
	}
	return header, data, nil
}
