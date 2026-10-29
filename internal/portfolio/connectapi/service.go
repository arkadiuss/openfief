// Package connectapi serves the portfolio context's ModelPortfolioService
// (proto/openfief/portfolio/v1/model.proto) to view plugins over Connect:
// the built-in UI hands the owner's model portfolio to the core through it
// (D-52, D-49). Callers are admitted by the interceptor the caller passes,
// cmd/core's view-plugin check; the owner's session travels in the request
// and the use cases check its scope. Errors map to Connect codes: no valid
// session is unauthenticated, a session that is not the owner's is
// permission_denied, a model breaking a rule is invalid_argument with the
// rule it broke, anything else internal with nothing of the error.
package connectapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

var (
	// errInternal is what a caller sees of an internal failure.
	errInternal = errors.New("internal error")
	// errUnknownKind reports a target of no known kind.
	errUnknownKind = errors.New("unknown target kind")
)

// New returns the ModelPortfolioService's path and HTTP handler over
// models, with callers in front of every call to admit view plugins only,
// logging to log.
func New(models *app.Models, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return portfoliov1connect.NewModelPortfolioServiceHandler(&service{models: models, log: log},
		connect.WithInterceptors(callers))
}

// service implements portfoliov1connect.ModelPortfolioServiceHandler.
type service struct {
	models *app.Models
	log    *slog.Logger
}

// GetModelPortfolio returns the stored model portfolio to the owner.
func (s *service) GetModelPortfolio(ctx context.Context, req *connect.Request[portfoliov1.GetModelPortfolioRequest]) (*connect.Response[portfoliov1.GetModelPortfolioResponse], error) {
	m, err := s.models.Settings(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.GetModelPortfolioResponse{Model: pbModel(&m)}), nil
}

// SaveModelPortfolio replaces the model portfolio for the owner.
func (s *service) SaveModelPortfolio(ctx context.Context, req *connect.Request[portfoliov1.SaveModelPortfolioRequest]) (*connect.Response[portfoliov1.SaveModelPortfolioResponse], error) {
	m, err := modelFrom(req.Msg.GetModel())
	if err == nil {
		m, err = s.models.Save(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), m)
	}
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&portfoliov1.SaveModelPortfolioResponse{Model: pbModel(&m)}), nil
}

// fail maps err to its Connect error. Only a broken rule carries its text,
// which names the rule and the target, so the owner can fix it.
func (s *service) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, portfolio.ErrInvalidModel):
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "model portfolio call failed", slog.String("error", err.Error()))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// pbModel converts a model portfolio to its message.
func pbModel(m *portfolio.Model) *portfoliov1.ModelPortfolio {
	out := &portfoliov1.ModelPortfolio{Enabled: m.Enabled, Targets: make([]*portfoliov1.Target, 0, len(m.Targets))}
	for i := range m.Targets {
		t := &m.Targets[i]
		pb := &portfoliov1.Target{WeightPercent: t.Weight.String()}
		switch t.Kind {
		case portfolio.TargetInstrument:
			pb.Kind = portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT
			if in, err := app.ParseInstrument(t.Name); err == nil {
				pb.Isin, pb.Symbol = in.ISIN.String(), in.Symbol
			}
		case portfolio.TargetIndex:
			pb.Kind, pb.Index, pb.Currency = portfoliov1.TargetKind_TARGET_KIND_INDEX, t.Name, t.Currency.String()
		}
		out.Targets = append(out.Targets, pb)
	}
	return out
}

// modelFrom reads a model portfolio message. A malformed field is an error
// wrapping portfolio.ErrInvalidModel; the rules are checked by the use case.
func modelFrom(pb *portfoliov1.ModelPortfolio) (portfolio.Model, error) {
	m := portfolio.Model{Enabled: pb.GetEnabled()}
	for i, t := range pb.GetTargets() {
		target, err := targetFrom(t)
		if err != nil {
			return portfolio.Model{}, fmt.Errorf("%w: target %d: %w", portfolio.ErrInvalidModel, i+1, err)
		}
		m.Targets = append(m.Targets, target)
	}
	return m, nil
}

// targetFrom reads one target message.
func targetFrom(pb *portfoliov1.Target) (portfolio.Target, error) {
	w, err := kernel.ParseDecimal(pb.GetWeightPercent())
	if err != nil {
		return portfolio.Target{}, fmt.Errorf("weight %q: %w", pb.GetWeightPercent(), err)
	}
	switch pb.GetKind() {
	case portfoliov1.TargetKind_TARGET_KIND_INSTRUMENT:
		return portfolio.Target{Kind: portfolio.TargetInstrument, Name: pb.GetIsin() + "/" + pb.GetSymbol(), Weight: w}, nil
	case portfoliov1.TargetKind_TARGET_KIND_INDEX:
		c, err := kernel.ParseCurrency(pb.GetCurrency())
		if err != nil {
			return portfolio.Target{}, fmt.Errorf("currency %q: %w", pb.GetCurrency(), err)
		}
		return portfolio.Target{Kind: portfolio.TargetIndex, Name: pb.GetIndex(), Currency: c, Weight: w}, nil
	case portfoliov1.TargetKind_TARGET_KIND_UNSPECIFIED:
	}
	return portfolio.Target{}, fmt.Errorf("%w: %s", errUnknownKind, pb.GetKind())
}
