// Package connectapi serves the modelrouter context's ModelSettingsService
// (openfief.modelrouter.v1) over Connect RPC to view plugins (docs/DESIGN.md
// D-212, D-205, D-206), and translates its messages to and from the domain.
package connectapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// maxRequestBytes caps one request's body.
const maxRequestBytes = 1 << 20

// errInternal is what a caller sees of a failure the core logs.
var errInternal = errors.New("internal error")

// NewSettings returns the ModelSettingsService's path and HTTP handler over
// settings, with callers in front of every call to admit view plugins only,
// logging to log.
func NewSettings(settings *app.Settings, callers connect.Interceptor, log *slog.Logger) (string, http.Handler) {
	return modelsv1connect.NewModelSettingsServiceHandler(&settingsService{settings: settings, log: log},
		connect.WithInterceptors(callers), connect.WithReadMaxBytes(maxRequestBytes))
}

// settingsService implements modelsv1connect.ModelSettingsServiceHandler.
type settingsService struct {
	settings *app.Settings
	log      *slog.Logger
}

// GetModelSettings returns the switch, the toggles, the routes and the
// tiers' allowed models to the owner.
func (s *settingsService) GetModelSettings(ctx context.Context, req *connect.Request[modelsv1.GetModelSettingsRequest]) (*connect.Response[modelsv1.GetModelSettingsResponse], error) {
	got, err := s.settings.Owner(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	out := &modelsv1.GetModelSettingsResponse{AgentsEnabled: got.Switch.Enabled}
	for _, f := range modelrouter.Features() {
		out.Features = append(out.Features, &modelsv1.FeatureToggle{Feature: string(f), Enabled: got.Switch.Toggles[f]})
	}
	for _, role := range modelrouter.Roles() {
		out.Routes = append(out.Routes, &modelsv1.Route{Role: string(role), Models: refStrings(got.Routes[role].Models)})
	}
	for _, tier := range modelrouter.Tiers() {
		out.Tiers = append(out.Tiers, &modelsv1.TierModels{Tier: string(tier), Models: refStrings(got.Policy[tier])})
	}
	return connect.NewResponse(out), nil
}

// SetAgentsEnabled turns the one switch on or off.
func (s *settingsService) SetAgentsEnabled(ctx context.Context, req *connect.Request[modelsv1.SetAgentsEnabledRequest]) (*connect.Response[modelsv1.SetAgentsEnabledResponse], error) {
	if err := s.settings.SetEnabled(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetEnabled()); err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&modelsv1.SetAgentsEnabledResponse{}), nil
}

// SetFeatureEnabled turns one feature's toggle on or off.
func (s *settingsService) SetFeatureEnabled(ctx context.Context, req *connect.Request[modelsv1.SetFeatureEnabledRequest]) (*connect.Response[modelsv1.SetFeatureEnabledResponse], error) {
	err := s.settings.SetFeature(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetFeature(), req.Msg.GetEnabled())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&modelsv1.SetFeatureEnabledResponse{}), nil
}

// SetRoute sets a role's primary model and fallbacks.
func (s *settingsService) SetRoute(ctx context.Context, req *connect.Request[modelsv1.SetRouteRequest]) (*connect.Response[modelsv1.SetRouteResponse], error) {
	err := s.settings.SetRoute(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetRole(), req.Msg.GetModels())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&modelsv1.SetRouteResponse{}), nil
}

// SetTierModels sets the models allowed to receive a data tier.
func (s *settingsService) SetTierModels(ctx context.Context, req *connect.Request[modelsv1.SetTierModelsRequest]) (*connect.Response[modelsv1.SetTierModelsResponse], error) {
	err := s.settings.SetTierModels(ctx, req.Msg.GetOrigin(), req.Msg.GetSessionToken(), req.Msg.GetTier(), req.Msg.GetModels())
	if err != nil {
		return nil, s.fail(ctx, err)
	}
	return connect.NewResponse(&modelsv1.SetTierModelsResponse{}), nil
}

// fail maps err to its Connect error.
func (s *settingsService) fail(ctx context.Context, err error) *connect.Error {
	switch {
	case errors.Is(err, app.ErrNoSession):
		return connect.NewError(connect.CodeUnauthenticated, app.ErrNoSession)
	case errors.Is(err, app.ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, app.ErrNotOwner)
	case errors.Is(err, app.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	s.log.LogAttrs(ctx, slog.LevelError, "models settings call failed", slog.String("error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded)))
	return connect.NewError(connect.CodeInternal, errInternal)
}

// refStrings spells refs as "provider/model" references.
func refStrings(refs []modelrouter.Ref) []string {
	return slices.Collect(func(yield func(string) bool) {
		for _, r := range refs {
			if !yield(r.String()) {
				return
			}
		}
	})
}
