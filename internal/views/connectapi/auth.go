package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/views/app"
)

// bearerPrefix starts an Authorization header that carries a bearer token.
const bearerPrefix = "bearer "

// authInterceptor checks every call, over any protocol, before its handler
// runs (see authorize).
type authInterceptor struct {
	verifier app.TokenVerifier
	log      *slog.Logger
}

// WrapUnary refuses a unary call that authorize refuses.
func (a *authInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := a.authorize(ctx, req.Header().Get("Authorization")); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient leaves client streams alone: the read API serves none.
func (*authInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// WrapStreamingHandler refuses a streaming call that authorize refuses; the
// read API has none, so this only keeps a future one from skipping the check.
func (a *authInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if err := a.authorize(ctx, conn.RequestHeader().Get("Authorization")); err != nil {
			return err
		}
		return next(ctx, conn)
	}
}

// authorize returns nil when header carries a bearer token the verifier
// accepts; otherwise unauthenticated with a Bearer challenge, or internal
// when the token cannot be checked.
func (a *authInterceptor) authorize(ctx context.Context, header string) *connect.Error {
	token, ok := bearerToken(header)
	if !ok {
		e := connect.NewError(connect.CodeUnauthenticated, errTokenRequired)
		e.Meta().Set("WWW-Authenticate", "Bearer")
		return e
	}
	err := a.verifier.VerifyToken(ctx, token)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, app.ErrInvalidToken):
		e := connect.NewError(connect.CodeUnauthenticated, errTokenNotValid)
		e.Meta().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
		return e
	}
	logFailure(ctx, a.log, "verifying a token")
	return internalError(ctx)
}

// bearerToken returns the token of an Authorization header value of the
// form "Bearer <token>", the scheme in any case.
func bearerToken(header string) (string, bool) {
	if len(header) <= len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	return token, token != ""
}
