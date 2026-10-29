// Package connectapi serves the identity context's AuthService
// (proto/openfief/identity/v1/auth.proto) to view plugins over Connect: the
// built-in UI calls it as the backend of the owner's browser (author
// decision, auth split). Every call must carry a view plugin's projected
// ServiceAccount token (audience openfief-core, D-27) as its bearer token,
// checked by a Callers port; the owner's session token travels in the
// request message. Errors map to Connect codes: a refused caller or a
// forbidden request is permission_denied, an invalid session or a failed
// login is unauthenticated. Neither a token nor a WebAuthn payload is ever
// logged.
package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// ErrCallerRefused reports a caller that is not a view plugin, or whose
// token does not verify. Callers implementations return it wrapped.
var ErrCallerRefused = errors.New("caller refused")

// errNoCallerToken reports a call without a bearer token.
var errNoCallerToken = errors.New("a view plugin's ServiceAccount token is required")

// Callers verifies the bearer token of an AuthService call: a view plugin's
// projected ServiceAccount token with audience openfief-core (D-27). It
// returns the plugin's name, or an error wrapping ErrCallerRefused.
type Callers interface {
	VerifyCaller(ctx context.Context, token string) (plugin string, err error)
}

// New returns the AuthService's path and HTTP handler over auth, admitting
// only the callers the Callers port accepts and logging to log.
func New(auth *app.Auth, callers Callers, log *slog.Logger) (string, http.Handler) {
	return identityv1connect.NewAuthServiceHandler(&authService{auth: auth, log: log},
		connect.WithInterceptors(NewCallerInterceptor(callers, log)))
}

// CallerInterceptor admits only verified view plugins.
type CallerInterceptor struct {
	callers Callers
	log     *slog.Logger
}

// NewCallerInterceptor returns the interceptor New puts in front of the
// AuthService: it admits only the view plugins callers accepts, refusing
// any other call with permission_denied. cmd/core puts it in front of the
// other services it serves to view plugins only, such as the model
// portfolio's (D-52), so they admit callers the same way.
func NewCallerInterceptor(callers Callers, log *slog.Logger) *CallerInterceptor {
	return &CallerInterceptor{callers: callers, log: log}
}

// WrapUnary refuses a unary call from a caller the port refuses.
func (c *CallerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		plugin, refused := c.verify(ctx, req.Header().Get("Authorization"))
		if refused != nil {
			c.log.LogAttrs(ctx, slog.LevelWarn, "auth call refused", slog.String("procedure", req.Spec().Procedure))
			return nil, refused
		}
		resp, err := next(ctx, req)
		c.log.LogAttrs(ctx, slog.LevelInfo, "auth call", slog.String("procedure", req.Spec().Procedure),
			slog.String("plugin", plugin), slog.String("code", codeOf(err)))
		return resp, err
	}
}

// WrapStreamingClient leaves client streams alone: the service has none.
func (*CallerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler refuses every stream: the service has none.
func (*CallerInterceptor) WrapStreamingHandler(connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(context.Context, connect.StreamingHandlerConn) error {
		return connect.NewError(connect.CodeUnimplemented, errNoStreams)
	}
}

// errNoStreams reports a streaming call.
var errNoStreams = errors.New("AuthService has no streaming methods")

// verify returns the plugin the bearer token of header proves, or a
// permission_denied (refused) or internal (not checkable) error.
func (c *CallerInterceptor) verify(ctx context.Context, header string) (string, *connect.Error) {
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", connect.NewError(connect.CodePermissionDenied, errNoCallerToken)
	}
	plugin, err := c.callers.VerifyCaller(ctx, strings.TrimSpace(header[len(prefix):]))
	switch {
	case err == nil:
		return plugin, nil
	case errors.Is(err, ErrCallerRefused):
		return "", connect.NewError(connect.CodePermissionDenied, ErrCallerRefused)
	}
	c.log.LogAttrs(ctx, slog.LevelError, "verifying an auth caller failed", slog.String("error", err.Error()))
	return "", connect.NewError(connect.CodeInternal, errInternal)
}

// errInternal is what a caller sees of an internal failure.
var errInternal = errors.New("internal error")

// codeOf names err's Connect code, "ok" for nil.
func codeOf(err error) string {
	if err == nil {
		return "ok"
	}
	return connect.CodeOf(err).String()
}

// toConnectError maps an app error to its Connect code. The message is the
// error class only; the details stay in the log.
func toConnectError(ctx context.Context, log *slog.Logger, err error) *connect.Error {
	classes := []struct {
		class error
		code  connect.Code
	}{
		{app.ErrInvalidSession, connect.CodeUnauthenticated},
		{app.ErrAuthenticationFailed, connect.CodeUnauthenticated},
		{app.ErrForbidden, connect.CodePermissionDenied},
		{app.ErrRegistrationFailed, connect.CodeInvalidArgument},
		{app.ErrInvalidRequest, connect.CodeInvalidArgument},
		{app.ErrCeremony, connect.CodeFailedPrecondition},
		{app.ErrLastPasskey, connect.CodeFailedPrecondition},
		{app.ErrNotFound, connect.CodeNotFound},
	}
	for _, c := range classes {
		if errors.Is(err, c.class) {
			log.LogAttrs(ctx, slog.LevelInfo, "auth call denied", slog.String("reason", err.Error()))
			return connect.NewError(c.code, c.class)
		}
	}
	log.LogAttrs(ctx, slog.LevelError, "auth call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}
