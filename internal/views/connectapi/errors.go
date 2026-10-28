package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/views"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// errBadParameter reports a request field the API does not accept, such as
// a malformed date or an unspecified enum.
var errBadParameter = errors.New("invalid parameter")

// What clients see of an error: fixed texts, because an error's own text can
// quote account IDs, amounts or what was asked for.
var (
	errNoSuchAccount = errors.New("no such account")
	errModelOff      = errors.New("no model portfolio")
	errBadRange      = errors.New("the date range ends before it starts or is too long")
	errBadPageToken  = errors.New("the page token is not one this API returned")
	errBadPageSize   = errors.New("the page size is out of range")
	errBadDimension  = errors.New("the allocation dimension is not known")
	errBadField      = errors.New("a request field is missing or malformed")
	errFailed        = errors.New("the request failed; its request ID names it in the server log")
	errTokenRequired = errors.New("a bearer token is required")
	errTokenNotValid = errors.New("the bearer token is not valid")
)

// invalidMessage returns what a client sees of an error that means the
// request asked for something invalid, or nil for any other error.
func invalidMessage(err error) error {
	switch {
	case errors.Is(err, views.ErrInvalidRange):
		return errBadRange
	case errors.Is(err, views.ErrInvalidCursor):
		return errBadPageToken
	case errors.Is(err, views.ErrInvalidPageSize):
		return errBadPageSize
	case errors.Is(err, views.ErrUnknownDimension):
		return errBadDimension
	case errors.Is(err, errBadParameter):
		return errBadField
	}
	return nil
}

// toConnectError returns the Connect error err stands for: not found for an
// unknown account, invalid argument for an invalid request, and internal,
// logged under the request ID only, for anything else. Vanguard maps these
// codes to HTTP statuses for REST (404, 400, 500).
func toConnectError(ctx context.Context, log *slog.Logger, err error) *connect.Error {
	if errors.Is(err, app.ErrUnknownAccount) {
		return connect.NewError(connect.CodeNotFound, errNoSuchAccount)
	}
	if errors.Is(err, app.ErrModelOff) {
		return connect.NewError(connect.CodeNotFound, errModelOff)
	}
	if msg := invalidMessage(err); msg != nil {
		return connect.NewError(connect.CodeInvalidArgument, msg)
	}
	logFailure(ctx, log, "answering")
	return internalError(ctx)
}

// internalError is what a client sees of an internal failure: the request ID
// to find it in the log, and nothing of the error itself.
func internalError(ctx context.Context) *connect.Error {
	e := connect.NewError(connect.CodeInternal, fmt.Errorf("%w: %s", errFailed, requestID(ctx)))
	e.Meta().Set("X-Request-ID", requestID(ctx))
	return e
}
