// Package connectapi serves the thesis context's ThesisService
// (openfief.thesis.v1) over Connect RPC to view plugins (docs/DESIGN.md D-229),
// and translates its messages to and from the domain. Every call is the
// owner's: no read API, script token or demo session reaches it.
package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/thesis"
	"github.com/arkadiuss/openfief/internal/thesis/app"
	thesisv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1/thesisv1connect"
)

// maxRequestBytes caps one request's body.
const maxRequestBytes = 1 << 20

// errInternal is what a caller sees of a failure the core logs.
var errInternal = errors.New("internal error")

// New returns the ThesisService's path and HTTP handler over theses, with
// callers in front of every call to admit view plugins only, logging to log.
func New(theses *app.Theses, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return thesisv1connect.NewThesisServiceHandler(&service{theses: theses, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// service implements thesisv1connect.ThesisServiceHandler.
type service struct {
	theses *app.Theses
	log    *slog.Logger
}

// SaveThesis stores a draft as a new thesis.
func (s *service) SaveThesis(ctx context.Context, req *connect.Request[thesisv1.SaveThesisRequest]) (*connect.Response[thesisv1.SaveThesisResponse], error) {
	got, err := s.theses.Save(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), draftFrom(req.Msg.GetDraft()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&thesisv1.SaveThesisResponse{Thesis: thesisTo(&got)}), nil
}

// EditThesis stores a draft as the next version of a thesis.
func (s *service) EditThesis(ctx context.Context, req *connect.Request[thesisv1.EditThesisRequest]) (*connect.Response[thesisv1.EditThesisResponse], error) {
	m := req.Msg
	got, err := s.theses.Edit(ctx, m.GetOrigin(), m.GetSessionToken(), m.GetThesisId(), int(m.GetBaseVersion()), draftFrom(m.GetDraft()))
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&thesisv1.EditThesisResponse{Thesis: thesisTo(&got)}), nil
}

// GetThesis returns a thesis with every version.
func (s *service) GetThesis(ctx context.Context, req *connect.Request[thesisv1.GetThesisRequest]) (*connect.Response[thesisv1.GetThesisResponse], error) {
	got, err := s.theses.Get(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetThesisId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&thesisv1.GetThesisResponse{Thesis: thesisTo(&got)}), nil
}

// ListTheses returns every thesis as its latest version.
func (s *service) ListTheses(ctx context.Context, req *connect.Request[thesisv1.ListThesesRequest]) (*connect.Response[thesisv1.ListThesesResponse], error) {
	got, err := s.theses.List(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &thesisv1.ListThesesResponse{Theses: make([]*thesisv1.ThesisSummary, 0, len(got))}
	for i := range got {
		sum := &got[i]
		out.Theses = append(out.Theses, &thesisv1.ThesisSummary{
			Id: sum.ID, Version: int32(sum.Version), CreatedAt: timestamppb.New(sum.CreatedAt), //nolint:gosec // a version count
			UpdatedAt: timestamppb.New(sum.UpdatedAt), Draft: draftTo(&sum.Draft), DataTier: thesis.DataTier,
		})
	}
	return connect.NewResponse(out), nil
}

// draftFrom translates a draft of the contract; a missing one is empty, which
// the domain refuses.
func draftFrom(m *thesisv1.ThesisDraft) *thesis.Draft {
	d := &thesis.Draft{
		Statement: m.GetStatement(), Instruments: m.GetInstruments(), Invalidation: m.GetInvalidation(), Sources: m.GetSources(),
	}
	switch m.GetHorizon() {
	case thesisv1.Horizon_HORIZON_LONG_TERM:
		d.Horizon = thesis.HorizonLongTerm
	case thesisv1.Horizon_HORIZON_SHORT_TERM:
		d.Horizon = thesis.HorizonShortTerm
	case thesisv1.Horizon_HORIZON_UNSPECIFIED:
	}
	return d
}

// draftTo translates a draft for the contract.
func draftTo(d *thesis.Draft) *thesisv1.ThesisDraft {
	out := &thesisv1.ThesisDraft{
		Statement: d.Statement, Instruments: d.Instruments, Invalidation: d.Invalidation, Sources: d.Sources,
	}
	switch d.Horizon {
	case thesis.HorizonLongTerm:
		out.Horizon = thesisv1.Horizon_HORIZON_LONG_TERM
	case thesis.HorizonShortTerm:
		out.Horizon = thesisv1.Horizon_HORIZON_SHORT_TERM
	}
	return out
}

// thesisTo translates a thesis with its versions.
func thesisTo(t *thesis.Thesis) *thesisv1.Thesis {
	out := &thesisv1.Thesis{Id: t.ID, DataTier: thesis.DataTier, Versions: make([]*thesisv1.ThesisVersion, 0, len(t.Versions))}
	for i := range t.Versions {
		v := &t.Versions[i]
		out.Versions = append(out.Versions, &thesisv1.ThesisVersion{
			Version: int32(v.Number), SavedAt: timestamppb.New(v.SavedAt), Draft: draftTo(&v.Draft), //nolint:gosec // a version count
		})
	}
	return out
}

// fail maps err to its Connect error.
func (s *service) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, kernel.ErrAuditUnavailable):
		return connect.NewError(connect.CodeUnavailable, kernel.ErrAuditUnavailable)
	case errors.Is(err, thesis.ErrInvalid):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, thesis.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, thesis.ErrNotFound)
	case errors.Is(err, thesis.ErrStale):
		return connect.NewError(connect.CodeFailedPrecondition, thesis.ErrStale)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "thesis call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}
