// Package connectapi serves the attention context's AttentionService
// (proto/openfief/attention/v1/attention.proto) to view plugins over
// Connect: the built-in UI hands the owner's acknowledgement or resolution
// of an attention item to the core through it (D-130, D-79). Callers are
// admitted by the interceptor the caller passes, cmd/core's view-plugin
// check; the owner's session travels in the request and the use case
// checks its scope. Errors map to Connect codes: no valid session is
// unauthenticated, a session that is not the owner's permission_denied, an
// unknown item not_found, an action the item doesn't allow
// failed_precondition, anything else internal with nothing of the error.
package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
	attentionv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/attention/v1/attentionv1connect"
)

// errInternal is what a caller sees of an internal failure.
var errInternal = errors.New("internal error")

// NewAttention returns the AttentionService's path and HTTP handler over
// items, with callers in front of every call to admit view plugins only,
// logging to log.
func NewAttention(items *app.Items, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return attentionv1connect.NewAttentionServiceHandler(&attentionService{items: items, log: log},
		connect.WithInterceptors(callers))
}

// attentionService implements attentionv1connect.AttentionServiceHandler.
type attentionService struct {
	items *app.Items
	log   *slog.Logger
}

// AcknowledgeAttentionItem records the owner's acknowledgement of an item.
func (s *attentionService) AcknowledgeAttentionItem(ctx context.Context, req *connect.Request[attentionv1.AcknowledgeAttentionItemRequest]) (*connect.Response[attentionv1.AcknowledgeAttentionItemResponse], error) {
	m := req.Msg
	if _, err := s.items.Acknowledge(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetItemId()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&attentionv1.AcknowledgeAttentionItemResponse{}), nil
}

// ResolveAttentionItem records that the owner resolved an item.
func (s *attentionService) ResolveAttentionItem(ctx context.Context, req *connect.Request[attentionv1.ResolveAttentionItemRequest]) (*connect.Response[attentionv1.ResolveAttentionItemResponse], error) {
	m := req.Msg
	if _, err := s.items.Resolve(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetItemId()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&attentionv1.ResolveAttentionItemResponse{}), nil
}

// fail maps err to its Connect error. A refused action carries its text,
// which says why, so the owner can see it.
func (s *attentionService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrUnknownItem):
		return connect.NewError(connect.CodeNotFound, app.ErrUnknownItem)
	case errors.Is(err, attention.ErrInvalidTransition), errors.Is(err, attention.ErrNotOwnerResolvable):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "attention call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}
