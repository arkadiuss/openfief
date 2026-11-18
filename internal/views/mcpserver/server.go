package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// maxRequestBytes bounds one JSON-RPC request: the arguments of a tool are a
// few words, so a megabyte is far more than a client sends.
const maxRequestBytes = 1 << 20

// Config is what the MCP server needs. Every field but Log and Version is
// required.
type Config struct {
	// Gate says whether the MCP server answers at all (D-212).
	Gate Gate
	// Tokens checks a script token's secret.
	Tokens Tokens
	// Toolbox offers and runs the tools.
	Toolbox Toolbox
	// Auditor records each tool call as usage (D-211).
	Auditor Auditor
	// Log gets one line per failure of the server's own, with no token,
	// argument or result in it; nil discards.
	Log *slog.Logger
	// Version is the version the server reports to a client; empty is "dev".
	Version string
	// OnUnauthenticated, when set, is called for each request refused with 401,
	// to count it: such a request makes no audit entry of its own.
	OnUnauthenticated func(ctx context.Context)
}

// ErrMissingDependency reports a Config without its Gate, Tokens, Toolbox or
// Auditor.
var ErrMissingDependency = errors.New("mcpserver: config needs a Gate, Tokens, Toolbox and Auditor")

// handler is the MCP endpoint (D-213): the owner's switch and the MCP toggle
// checked first, then the script token, then the SDK's Streamable HTTP
// handler, stateless, over a server whose tools run the toolbox with that
// token.
type handler struct {
	cfg  Config
	sdk  *sdk.StreamableHTTPHandler
	tool []Tool
}

// callerKey carries the authenticated caller from the endpoint to the tools
// of the request's server.
type callerKey struct{}

// caller is the script token a request authenticated with: its ID, for the
// audit entry, and its secret, which the toolbox reads the read API with, so
// the token's grants limit what every tool sees (D-209).
type caller struct {
	id, secret string
}

// New returns the MCP endpoint, to be mounted at /mcp on the browser port.
// It answers 404 while the switch or the MCP toggle is off, 401 without a
// valid script token, 403 for a token the owner did not allow to use MCP and
// otherwise MCP over Streamable HTTP, without sessions: each request stands
// alone.
func New(cfg *Config) (http.Handler, error) {
	if cfg.Gate == nil || cfg.Tokens == nil || cfg.Toolbox == nil || cfg.Auditor == nil {
		return nil, ErrMissingDependency
	}
	if cfg.Log == nil {
		cfg.Log = slog.New(slog.DiscardHandler)
	}
	if cfg.Version == "" {
		cfg.Version = "dev"
	}
	h := &handler{cfg: *cfg, tool: cfg.Toolbox.Tools()}
	h.sdk = sdk.NewStreamableHTTPHandler(h.serverFor, &sdk.StreamableHTTPOptions{
		Stateless: true, JSONResponse: true, MaxRequestBodyBytes: maxRequestBytes,
	})
	return h, nil
}

// ServeHTTP checks the switch, then the token, then lets the SDK answer.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	on, err := h.cfg.Gate.Enabled(r.Context())
	if err != nil {
		h.cfg.Log.ErrorContext(r.Context(), "mcp: checking the switch")
		plain(w, http.StatusServiceUnavailable, "the MCP server cannot be checked now")
		return
	}
	if !on {
		// With the switch or the toggle off no MCP endpoint answers: this
		// is what a path nothing serves says (D-212).
		plain(w, http.StatusNotFound, "404 page not found")
		return
	}
	secret, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		h.unauthenticated(r.Context())
		w.Header().Set("WWW-Authenticate", "Bearer")
		plain(w, http.StatusUnauthorized, "a script token is required")
		return
	}
	id, err := h.cfg.Tokens.Authenticate(r.Context(), secret)
	switch {
	case err == nil:
	case errors.Is(err, ErrInvalidToken):
		h.unauthenticated(r.Context())
		w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		plain(w, http.StatusUnauthorized, "the script token is not valid")
		return
	case errors.Is(err, ErrNotAllowed):
		// A valid token the owner did not allow to use MCP: the refusal is an entry of its own.
		if recErr := h.cfg.Auditor.RecordCall(r.Context(), Call{Token: id, Outcome: OutcomeNotAllowed}); recErr != nil {
			h.cfg.Log.ErrorContext(r.Context(), "mcp: recording a refusal")
		}
		plain(w, http.StatusForbidden, "the script token is not allowed to use MCP")
		return
	default:
		h.cfg.Log.ErrorContext(r.Context(), "mcp: verifying a token")
		plain(w, http.StatusInternalServerError, "the script token cannot be checked now")
		return
	}
	h.sdk.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), callerKey{}, caller{id: id, secret: secret})))
}

// unauthenticated counts a request refused with 401.
func (h *handler) unauthenticated(ctx context.Context) {
	if h.cfg.OnUnauthenticated != nil {
		h.cfg.OnUnauthenticated(ctx)
	}
}

// serverFor returns the MCP server of one request: the toolbox's tools, each
// running as the request's token. A request that reaches it is authenticated.
func (h *handler) serverFor(r *http.Request) *sdk.Server {
	c, ok := r.Context().Value(callerKey{}).(caller)
	if !ok {
		return nil
	}
	srv := sdk.NewServer(&sdk.Implementation{Name: "openfief", Version: h.cfg.Version}, &sdk.ServerOptions{
		Instructions: "OpenFief's read-only portfolio tools. Every tool shows only what this token's grants cover, and none changes anything.",
		Logger:       h.cfg.Log,
	})
	for i := range h.tool {
		t := h.tool[i]
		srv.AddTool(&sdk.Tool{
			Name: t.Name, Description: t.Description, InputSchema: json.RawMessage(t.Schema),
			Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: new(false), OpenWorldHint: new(false)},
		}, h.call(c, t.Name))
	}
	return srv
}

// call returns the handler of tool name for caller c: it runs the tool with c's
// token, enters the call in the audit log as usage, and returns the result
// only once that entry is written, so a call the log can't record shows
// nothing (D-211, fail closed).
func (h *handler) call(c caller, name string) sdk.ToolHandler {
	return func(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		return h.run(ctx, c, name, string(req.Params.Arguments)), nil
	}
}

// run runs the tool and enters the call; its result is the tool's, or the
// error that the call could not be recorded.
func (h *handler) run(ctx context.Context, c caller, name, args string) *sdk.CallToolResult {
	res := h.cfg.Toolbox.Call(ctx, c.secret, name, args)
	if err := h.cfg.Auditor.RecordCall(ctx, Call{Token: c.id, Tool: name, Outcome: res.Outcome}); err != nil {
		h.cfg.Log.ErrorContext(ctx, "mcp: recording a call")
		return textResult("the call could not be recorded, so nothing was read", true)
	}
	return textResult(res.Text, res.Outcome != OutcomeRan)
}

// textResult returns text as a tool's result; failed marks one that holds no
// data.
func textResult(text string, failed bool) *sdk.CallToolResult {
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: text}}, IsError: failed}
}

// bearerPrefix starts an Authorization header that carries a bearer token.
const bearerPrefix = "bearer "

// bearerToken returns the token of an Authorization header value of the form
// "Bearer <token>", the scheme in any case.
func bearerToken(header string) (string, bool) {
	if len(header) <= len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	return token, token != ""
}

// plain answers status with a short text and no data.
func plain(w http.ResponseWriter, status int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, text); err != nil {
		return // the client went away
	}
}
