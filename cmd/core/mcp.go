package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	modelsdomain "github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/views/mcpserver"
	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
)

// mcpPath is where the core serves MCP on its browser port (D-213). The front
// door treats it as the core's own path (plugins.CoreServesPath), so no view
// plugin's browser path can take it.
const mcpPath = "/mcp"

// withMCP returns api, the browser port's read API, with the MCP endpoint
// mounted at mcpPath in front of it. Every other path is the read API's.
func withMCP(api, mcp http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(mcpPath, mcp)
	mux.Handle("/", api)
	return mux
}

// mcpSettings reads the models switch: *modelsapp.Settings implements it.
type mcpSettings interface {
	Switch(ctx context.Context) (modelsdomain.Switch, error)
}

// mcpGate is the MCP server's Gate over the owner's switch for agents and
// models (D-212): the switch and the MCP toggle, read at every request, so
// turning either off stops the answers at once.
type mcpGate struct {
	settings mcpSettings
	// authorizer is the one every read path asks: MCP is never served where it
	// ignores a script token's grants (D-209).
	authorizer identityapp.Authorizer
}

var _ mcpserver.Gate = mcpGate{}

// errGrantsNotEnforced reports an authorizer that lets a script token read what
// its grants do not cover, so the MCP server does not answer (503).
var errGrantsNotEnforced = errors.New("the authorizer does not enforce script token grants")

// Enabled reports whether the switch is on and the MCP toggle under it is on;
// with them on it fails, answering 503, unless the authorizer limits script
// tokens to their grants.
func (g mcpGate) Enabled(ctx context.Context) (bool, error) {
	s, err := g.settings.Switch(ctx)
	if err != nil {
		return false, fmt.Errorf("reading the models switch: %w", err)
	}
	if !s.Allows(modelsdomain.FeatureMCP) {
		return false, nil
	}
	if !identityapp.EnforcesTokenGrants(g.authorizer) {
		return false, errGrantsNotEnforced
	}
	return true, nil
}

// mcpTokens is the MCP server's Tokens over the script tokens (D-213): a
// token that is valid, read-scoped and that the owner allowed to use MCP.
type mcpTokens struct{ tokens *identityapp.Tokens }

var _ mcpserver.Tokens = mcpTokens{}

// Authenticate returns the ID of the token secret stands for, or an error
// wrapping mcpserver.ErrInvalidToken or mcpserver.ErrNotAllowed.
func (t mcpTokens) Authenticate(ctx context.Context, secret string) (string, error) {
	tok, err := t.tokens.VerifyMCP(ctx, secret)
	switch {
	case err == nil:
		return string(tok.ID), nil
	case errors.Is(err, identityapp.ErrMCPNotAllowed):
		return string(tok.ID), fmt.Errorf("%w: %w", mcpserver.ErrNotAllowed, err)
	case errors.Is(err, identityapp.ErrInvalidToken):
		return "", fmt.Errorf("%w: %w", mcpserver.ErrInvalidToken, err)
	}
	return "", fmt.Errorf("verifying a token for MCP: %w", err)
}

// uiToolbox is the UI's toolbox, the Ask box's read-only tools: *ui.Toolbox
// implements it.
type uiToolbox interface {
	Tools() []ui.ToolInfo
	Call(ctx context.Context, token, name, args string) ui.ToolResult
}

// mcpTools is the MCP server's Toolbox over the Ask box's tools (D-213,
// D-203): the same table, the same answerers and the same read API, with the
// caller's own token.
type mcpTools struct{ box uiToolbox }

var _ mcpserver.Toolbox = mcpTools{}

// Tools returns the Ask box's tools.
func (m mcpTools) Tools() []mcpserver.Tool {
	infos := m.box.Tools()
	out := make([]mcpserver.Tool, len(infos))
	for i, t := range infos {
		out[i] = mcpserver.Tool{Name: t.Name, Description: t.Description, Schema: t.Schema}
	}
	return out
}

// Call runs the tool with the caller's own token.
func (m mcpTools) Call(ctx context.Context, secret, name, args string) mcpserver.Result {
	res := m.box.Call(ctx, secret, name, args)
	outcome := mcpserver.OutcomeReadFailed
	switch res.Outcome {
	case ui.ToolRan:
		outcome = mcpserver.OutcomeRan
	case ui.ToolUnknown, ui.ToolBadArguments:
		outcome = mcpserver.OutcomeBadArguments
	case ui.ToolReadFailed, ui.ToolUnauthenticated:
	}
	return mcpserver.Result{Outcome: outcome, Text: res.Text}
}

// mcpAudit is the MCP server's Auditor over the identity context's audit log
// (D-211): each call is one entry of the action mcp_call whose actor is the
// script token, by its ID, and that names the tool and how the call ended.
// There is nowhere in it for an argument, a result or an amount.
type mcpAudit struct {
	log auditAppender
	now func() time.Time
}

var _ mcpserver.Auditor = mcpAudit{}

// RecordCall appends the entry of c, failing when it can't be written.
func (a mcpAudit) RecordCall(ctx context.Context, c mcpserver.Call) error {
	outcome := identity.OutcomeFailed
	var refs []identity.Ref
	if c.Outcome != mcpserver.OutcomeNotAllowed {
		refs = []identity.Ref{{Kind: identity.RefMCPTool, ID: c.Tool}}
	}
	switch c.Outcome {
	case mcpserver.OutcomeRan:
		outcome = identity.OutcomeOK
	case mcpserver.OutcomeBadArguments, mcpserver.OutcomeNotAllowed:
		outcome = identity.OutcomeRefused
	case mcpserver.OutcomeReadFailed:
	}
	e, err := identity.NewAuditEntry(a.now(), identity.Actor{Kind: identity.ActorScriptToken, ID: c.Token},
		identity.ActionMCPCall, outcome, refs...)
	if err != nil {
		return fmt.Errorf("building the audit entry of an MCP call: %w", err)
	}
	if err := a.log.AppendAudit(ctx, &e); err != nil {
		return fmt.Errorf("writing the audit entry of an MCP call: %w", err)
	}
	return nil
}

// mountBrowserPort puts the MCP endpoint and then the front door before the
// browser port's read API: the endpoint first, so the front door sends /mcp to
// the core and never to a plugin.
func (c *core) mountBrowserPort(pool *pgxpool.Pool, cl *cluster, cfg *serveConfig, now func() time.Time, logger *slog.Logger) error {
	if err := c.mountMCP(pool, now, logger); err != nil {
		return err
	}
	return c.putFrontDoor(cl, cfg, now, logger)
}

// mountMCP puts the MCP endpoint (mcpEndpoint) at mcpPath on the browser port,
// in front of its read API, which the tools read in process. It runs before the
// front door, so /mcp is the core's and never a plugin's.
func (c *core) mountMCP(pool *pgxpool.Pool, now func() time.Time, logger *slog.Logger) error {
	store := identitypg.New(pool)
	mcp, err := mcpEndpoint(c.browserPort, &mcpDeps{
		settings: c.agents.settings, authorizer: c.access.authorizer, tokens: identityapp.NewTokens(store, now, nil), log: store, now: now, logger: logger,
	})
	if err != nil {
		return err
	}
	c.browserPort = withMCP(c.browserPort, mcp)
	return nil
}

// mcpDeps is what the MCP endpoint is wired from: the switch, the authorizer
// every read path asks, the script tokens, the audit log and the clock.
type mcpDeps struct {
	settings   mcpSettings
	authorizer identityapp.Authorizer
	tokens     *identityapp.Tokens
	log        auditAppender
	now        func() time.Time
	logger     *slog.Logger
}

// mcpEndpoint returns the MCP endpoint (D-213) over the read API api, which it
// calls in process with each caller's own token, so the read API's grants
// limit what every tool sees. A request refused with 401 is counted by the
// metric openfief.mcp.unauthenticated, as it makes no audit entry.
func mcpEndpoint(api http.Handler, d *mcpDeps) (http.Handler, error) {
	unauthenticated, err := otel.Meter(coreService).Int64Counter("openfief.mcp.unauthenticated",
		metric.WithDescription("Requests to the MCP endpoint refused with 401: no token, or one that is not valid"))
	if err != nil {
		return nil, fmt.Errorf("MCP metric: %w", err)
	}
	box, err := ui.NewToolbox(&ui.ToolboxConfig{Read: viewsv1connect.NewReadServiceClient(inProcess(api), "http://core.invalid")})
	if err != nil {
		return nil, fmt.Errorf("MCP tools: %w", err)
	}
	h, err := mcpserver.New(&mcpserver.Config{
		Gate: mcpGate{settings: d.settings, authorizer: d.authorizer}, Tokens: mcpTokens{tokens: d.tokens}, Toolbox: mcpTools{box: box},
		Auditor: mcpAudit{log: d.log, now: d.now}, Log: d.logger,
		OnUnauthenticated: func(ctx context.Context) { unauthenticated.Add(ctx, 1) },
	})
	if err != nil {
		return nil, fmt.Errorf("MCP server: %w", err)
	}
	return h, nil
}

// inProcess returns an HTTP client whose requests are answered by h in this
// process, not over the network: the MCP tools read the core's own read API
// the way the built-in UI does, with the caller's token in the Authorization
// header, through the same handler, interceptors and authorizer.
func inProcess(h http.Handler) *http.Client {
	return &http.Client{Transport: handlerTransport{h: h}}
}

// handlerTransport is an http.RoundTripper that serves each request with h.
type handlerTransport struct{ h http.Handler }

// RoundTrip serves req with the handler and returns what it wrote.
func (t handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	w := &bufferedResponse{header: http.Header{}, status: http.StatusOK}
	r := req.Clone(req.Context())
	if r.Body == nil {
		r.Body = http.NoBody
	}
	r.RequestURI = req.URL.RequestURI()
	t.h.ServeHTTP(w, r)
	return &http.Response{
		Status: fmt.Sprintf("%d %s", w.status, http.StatusText(w.status)), StatusCode: w.status,
		Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
		Header: w.header, Body: io.NopCloser(bytes.NewReader(w.body.Bytes())), ContentLength: int64(w.body.Len()), Request: req,
	}, nil
}

// bufferedResponse is an http.ResponseWriter that keeps what is written.
type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (b *bufferedResponse) Header() http.Header { return b.header }

func (b *bufferedResponse) WriteHeader(status int) { b.status = status }

func (b *bufferedResponse) Write(p []byte) (int, error) {
	n, err := b.body.Write(p)
	if err != nil { // a bytes.Buffer only fails by running out of memory
		return n, fmt.Errorf("buffering the response: %w", err)
	}
	return n, nil
}
