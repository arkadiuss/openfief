package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// The fixed texts a caller sees of a refusal. None quotes a note, a reason or
// a figure (GR-12).
var (
	errMaintainerOff     = errors.New("the maintainer is off: turn on the models switch and the maintainer, and the model portfolio")
	errUnknownProposal   = errors.New("unknown proposal")
	errProposalDecided   = errors.New("the proposal is already decided")
	errProposalStale     = errors.New("the model portfolio changed since the proposal was made: dismiss it")
	errProposalInvalid   = errors.New("the proposal no longer gives a valid model portfolio: dismiss it")
	errNotWholePortfolio = errors.New("this session may not read the whole portfolio")
)

// NewMaintainer returns the MaintainerService's path and HTTP handler over
// maintainer, with callers in front of every call to admit view plugins only,
// logging to log.
func NewMaintainer(maintainer *app.Maintainer, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return portfoliov1connect.NewMaintainerServiceHandler(&maintainerService{m: maintainer, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// maintainerService implements portfoliov1connect.MaintainerServiceHandler.
type maintainerService struct {
	m   *app.Maintainer
	log *slog.Logger
}

// GetMaintainer returns the latest note and where the maintainer stands.
func (s *maintainerService) GetMaintainer(ctx context.Context, req *connect.Request[portfoliov1.GetMaintainerRequest]) (*connect.Response[portfoliov1.GetMaintainerResponse], error) {
	ov, err := s.m.Overview(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &portfoliov1.GetMaintainerResponse{Available: ov.Available, Running: ov.Running, PendingProposals: int64(ov.Pending)}
	if ov.Note != nil {
		out.Note = &portfoliov1.Note{
			RunId: ov.Note.RunID, Text: ov.Note.Text, CreatedAt: ov.Note.At.UTC().Format(time.RFC3339),
			Model: ov.Note.Model, Trigger: pbTrigger(ov.Note.Trigger),
		}
	}
	if ov.LastFailure != nil {
		out.LastFailure = &portfoliov1.RunFailure{
			At: ov.LastFailure.At.UTC().Format(time.RFC3339), ErrorClass: ov.LastFailure.Class, Trigger: pbTrigger(ov.LastFailure.Trigger),
		}
	}
	return connect.NewResponse(out), nil
}

// AskMaintainer queues a run for the owner.
func (s *maintainerService) AskMaintainer(ctx context.Context, req *connect.Request[portfoliov1.AskMaintainerRequest]) (*connect.Response[portfoliov1.AskMaintainerResponse], error) {
	started, err := s.m.Ask(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.AskMaintainerResponse{Started: started}), nil
}

// GetProposal returns a proposal with its figures to the owner.
func (s *maintainerService) GetProposal(ctx context.Context, req *connect.Request[portfoliov1.GetProposalRequest]) (*connect.Response[portfoliov1.GetProposalResponse], error) {
	p, err := s.m.Proposal(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.GetProposalResponse{Proposal: pbProposal(&p)}), nil
}

// ApplyProposal applies a proposal for the owner.
func (s *maintainerService) ApplyProposal(ctx context.Context, req *connect.Request[portfoliov1.ApplyProposalRequest]) (*connect.Response[portfoliov1.ApplyProposalResponse], error) {
	p, err := s.m.Apply(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.ApplyProposalResponse{Proposal: pbProposal(&p)}), nil
}

// DismissProposal dismisses a proposal for the owner.
func (s *maintainerService) DismissProposal(ctx context.Context, req *connect.Request[portfoliov1.DismissProposalRequest]) (*connect.Response[portfoliov1.DismissProposalResponse], error) {
	p, err := s.m.Dismiss(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetId())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.DismissProposalResponse{Proposal: pbProposal(&p)}), nil
}

// fail maps err to its Connect error by its kind, with a fixed text and never
// the error's own, which could hold a note or a figure.
func (s *maintainerService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrNotReader):
		return connect.NewError(connect.CodePermissionDenied, errNotWholePortfolio)
	case errors.Is(err, app.ErrMaintainerOff):
		return connect.NewError(connect.CodeFailedPrecondition, errMaintainerOff)
	case errors.Is(err, app.ErrUnknownProposal):
		return connect.NewError(connect.CodeNotFound, errUnknownProposal)
	case errors.Is(err, app.ErrProposalDecided):
		return connect.NewError(connect.CodeFailedPrecondition, errProposalDecided)
	case errors.Is(err, portfolio.ErrStaleProposal):
		return connect.NewError(connect.CodeFailedPrecondition, errProposalStale)
	case errors.Is(err, portfolio.ErrInvalidProposal):
		return connect.NewError(connect.CodeFailedPrecondition, errProposalInvalid)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "maintainer call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// pbTrigger converts a run's trigger.
func pbTrigger(t app.RunTrigger) portfoliov1.RunTrigger {
	switch t {
	case app.TriggerWeekly:
		return portfoliov1.RunTrigger_RUN_TRIGGER_WEEKLY
	case app.TriggerOwner:
		return portfoliov1.RunTrigger_RUN_TRIGGER_OWNER
	}
	return portfoliov1.RunTrigger_RUN_TRIGGER_UNSPECIFIED
}

// pbStatus converts a proposal's status.
func pbStatus(s app.ProposalStatus) portfoliov1.ProposalStatus {
	switch s {
	case app.ProposalPending:
		return portfoliov1.ProposalStatus_PROPOSAL_STATUS_PENDING
	case app.ProposalApplied:
		return portfoliov1.ProposalStatus_PROPOSAL_STATUS_APPLIED
	case app.ProposalDismissed:
		return portfoliov1.ProposalStatus_PROPOSAL_STATUS_DISMISSED
	case app.ProposalSuperseded:
		return portfoliov1.ProposalStatus_PROPOSAL_STATUS_SUPERSEDED
	}
	return portfoliov1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
}

// pbProposal converts a proposal with its changes.
func pbProposal(p *app.StoredProposal) *portfoliov1.Proposal {
	out := &portfoliov1.Proposal{
		Id: p.ID, Status: pbStatus(p.Status), CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339), Model: p.Model, Note: p.Note,
		Changes: make([]*portfoliov1.TargetChange, 0, len(p.Proposal.Changes)),
	}
	if !p.DecidedAt.IsZero() {
		out.DecidedAt = p.DecidedAt.UTC().Format(time.RFC3339)
	}
	for i := range p.Proposal.Changes {
		out.Changes = append(out.Changes, pbChange(&p.Proposal.Changes[i]))
	}
	return out
}

// pbChange converts one change; an index has no current weight or drift.
func pbChange(c *portfolio.Change) *portfoliov1.TargetChange {
	out := &portfoliov1.TargetChange{
		Target: c.Target, BeforePercent: c.Before.String(), AfterPercent: c.After.String(), Held: c.Held, Reason: c.Reason,
	}
	if c.Held {
		out.CurrentPercent, out.DriftPoints = c.Current.String(), c.Drift.String()
	}
	return out
}
