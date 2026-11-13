package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// NewSyncs returns the SyncService's path and HTTP handler over the owner's
// sync retries (D-196), with callers in front of every call to admit view
// plugins only, logging to log.
func NewSyncs(r *app.Retries, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return ingestv1connect.NewSyncServiceHandler(&syncs{uc: r, log: log}, connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxMessageBytes))
}

// syncs implements ingestv1connect.SyncServiceHandler.
type syncs struct {
	uc  *app.Retries
	log *slog.Logger
}

// RetrySync starts one sync of the instance at once for the owner.
func (s *syncs) RetrySync(ctx context.Context, req *connect.Request[ingestv1.RetrySyncRequest]) (*connect.Response[ingestv1.RetrySyncResponse], error) {
	if err := s.uc.RetrySync(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetInstance()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&ingestv1.RetrySyncResponse{}), nil
}

// fail maps err to its Connect error: a refusal with its reason, anything
// else logged without its text's amounts and answered as internal.
func (s *syncs) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, kernel.ErrAuthzUnavailable):
		return connect.NewError(connect.CodeUnavailable, kernel.ErrAuthzUnavailable)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrNoInstance):
		return connect.NewError(connect.CodeNotFound, app.ErrNoInstance)
	case errors.Is(err, app.ErrSyncRunning):
		return connect.NewError(connect.CodeFailedPrecondition, app.ErrSyncRunning)
	case errors.Is(err, app.ErrRetryTooSoon):
		return connect.NewError(connect.CodeResourceExhausted, app.ErrRetryTooSoon)
	case errors.Is(err, app.ErrSyncsNotRunning):
		return connect.NewError(connect.CodeUnavailable, app.ErrSyncsNotRunning)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "sync retry failed", slog.String("error", app.LogSafe(err)))
	return connect.NewError(connect.CodeInternal, errInternal)
}
