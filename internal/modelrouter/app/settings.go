package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// Settings are the owner's agents-and-models settings use cases (D-212):
// reading them, and changing the one switch, a feature's toggle, a role's
// route and a tier's allowed models, each change checked against the
// owner's session.
type Settings struct {
	store  SettingsStore
	owners Owners

	directory ProviderDirectory
	instances ProviderInstances
	audit     SettingsAuditor
	now       func() time.Time
}

// NewSettings returns the settings use cases over store, with the owner's
// sessions checked by owners.
func NewSettings(store SettingsStore, owners Owners) *Settings {
	return &Settings{store: store, owners: owners}
}

// WithProviders makes the plugins the settings can list and describe: the
// directory names the model-provider instances that run, and instances
// describes each and lists its models.
func (s *Settings) WithProviders(directory ProviderDirectory, instances ProviderInstances) *Settings {
	s.directory, s.instances = directory, instances
	return s
}

// WithAudit makes every change write its entry to auditor, stamped by now,
// inside the transaction that stores it.
func (s *Settings) WithAudit(auditor SettingsAuditor, now func() time.Time) *Settings {
	s.audit, s.now = auditor, now
	return s
}

// Switch returns the switch and its toggles, for the core's own use, such
// as the Ask box deciding whether it has a model: it holds no secret and
// needs no session.
func (s *Settings) Switch(ctx context.Context) (modelrouter.Switch, error) {
	got, err := s.store.Settings(ctx)
	if err != nil {
		return modelrouter.Switch{}, fmt.Errorf("loading the models switch: %w", err)
	}
	return got.Switch, nil
}

// Owner returns every setting to the owner: a session of owner scope on
// origin.
func (s *Settings) Owner(ctx context.Context, origin, sessionToken string) (modelrouter.Settings, error) {
	if _, err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return modelrouter.Settings{}, err
	}
	got, err := s.store.Settings(ctx)
	if err != nil {
		return modelrouter.Settings{}, fmt.Errorf("loading the models settings: %w", err)
	}
	return got, nil
}

// SetEnabled turns the one switch on or off, for the owner.
func (s *Settings) SetEnabled(ctx context.Context, origin, sessionToken string, on bool) error {
	session, err := s.checkOwner(ctx, origin, sessionToken)
	if err != nil {
		return err
	}
	if err := s.store.SaveEnabled(ctx, on, s.recording(session, "switch:"+onOff(on))); err != nil {
		return fmt.Errorf("setting the models switch: %w", err)
	}
	return nil
}

// SetFeature turns a feature's toggle on or off, for the owner. An unknown
// feature is an error wrapping ErrInvalidInput.
func (s *Settings) SetFeature(ctx context.Context, origin, sessionToken, feature string, on bool) error {
	session, err := s.checkOwner(ctx, origin, sessionToken)
	if err != nil {
		return err
	}
	f, err := modelrouter.ParseFeature(feature)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err := s.store.SaveToggle(ctx, f, on, s.recording(session, "feature:"+string(f)+":"+onOff(on))); err != nil {
		return fmt.Errorf("setting a feature toggle: %w", err)
	}
	return nil
}

// SetRoute sets a role's primary model and fallbacks from "provider/model"
// references, for the owner; no references clear the route. An unknown
// role, a bad reference or a route breaking modelrouter.NewRoute is an error
// wrapping ErrInvalidInput.
func (s *Settings) SetRoute(ctx context.Context, origin, sessionToken, role string, refs []string) error {
	session, err := s.checkOwner(ctx, origin, sessionToken)
	if err != nil {
		return err
	}
	r, err := modelrouter.ParseRole(role)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	parsed, err := modelrouter.ParseRefs(refs)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if len(parsed) > 0 {
		if _, err := modelrouter.NewRoute(r, parsed); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidInput, err)
		}
	}
	if err := s.store.SaveRoute(ctx, r, parsed, s.recording(session, "route:"+string(r))); err != nil {
		return fmt.Errorf("setting a route: %w", err)
	}
	return nil
}

// SetDataTierModels sets the models allowed to receive a data tier from
// "provider/model" references, for the owner; none clear the list, which
// denies the tier to every model (D-206). An unknown tier, a bad reference
// or a list breaking modelrouter.NewDataTierPolicy is an error wrapping
// ErrInvalidInput.
func (s *Settings) SetDataTierModels(ctx context.Context, origin, sessionToken, tier string, refs []string) error {
	session, err := s.checkOwner(ctx, origin, sessionToken)
	if err != nil {
		return err
	}
	t, err := modelrouter.ParseDataTier(tier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	parsed, err := modelrouter.ParseRefs(refs)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if _, err := modelrouter.NewDataTierPolicy(map[modelrouter.DataTier][]modelrouter.Ref{t: parsed}); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err := s.store.SaveDataTierModels(ctx, t, parsed, s.recording(session, "tier:"+string(t))); err != nil {
		return fmt.Errorf("setting a tier's models: %w", err)
	}
	return nil
}

// Provider is a model-provider plugin instance as the owner sees it: where
// its models run, what it keeps, and its models, for choosing routes and tier
// lists (D-205). An instance that did not answer is not Available and has
// neither terms nor models.
type Provider struct {
	Instance  string
	Available bool
	Terms     modelrouter.ProviderInstance
	Models    []modelrouter.ModelEntry
}

// Providers lists the model-provider instances that run, in name order, each
// described and with its models, for the owner. An instance that is down is
// listed as unavailable. Before the core has found the plugins, or without a
// directory, the list is empty.
func (s *Settings) Providers(ctx context.Context, origin, sessionToken string) ([]Provider, error) {
	if _, err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return nil, err
	}
	if s.directory == nil || s.instances == nil {
		return nil, nil
	}
	names, err := s.directory.ModelProviders(ctx)
	if errors.Is(err, ErrPluginsNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("listing the model providers: %w", err)
	}
	out := make([]Provider, 0, len(names))
	for _, name := range slices.Sorted(slices.Values(names)) {
		out = append(out, s.describe(ctx, name))
	}
	return out, nil
}

// BrokenRoutes returns the entries of every route, in use or not, that name a
// provider instance no plugin has or one that is no model provider, for the
// settings page (D-196). Before the core has found the plugins it can't tell
// and returns none.
func (s *Settings) BrokenRoutes(ctx context.Context) ([]modelrouter.RouteProblem, error) {
	if s.instances == nil {
		return nil, nil
	}
	got, err := s.store.Settings(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading the models settings: %w", err)
	}
	out, err := routeProblems(ctx, &got, s.instances, false)
	if errors.Is(err, ErrPluginsNotFound) {
		return nil, nil
	}
	return out, err
}

// providerTimeout bounds what the settings page waits for one provider
// instance: a slow one is listed as unavailable.
const providerTimeout = 10 * time.Second

// describe asks the instance name for its terms and models; one that does not
// answer in time is unavailable.
func (s *Settings) describe(ctx context.Context, name string) Provider {
	ctx, cancel := context.WithTimeout(ctx, providerTimeout)
	defer cancel()
	p := Provider{Instance: name}
	terms, descErr := s.instances.Describe(ctx, name)
	models, listErr := s.instances.Models(ctx, name)
	if descErr == nil && listErr == nil {
		p.Available, p.Terms, p.Models = true, terms, models
	}
	return p
}

// recording returns the hook that writes the audit entry of a change of
// setting by session, inside the transaction that stores it; nil without an
// auditor.
func (s *Settings) recording(session, setting string) func(context.Context) error {
	if s.audit == nil {
		return nil
	}
	return func(ctx context.Context) error {
		c := &SettingChange{At: s.now().UTC(), Session: session, Setting: setting}
		if err := s.audit.RecordChange(ctx, c); err != nil {
			return fmt.Errorf("recording a change of the models settings: %w", err)
		}
		return nil
	}
}

// onOff spells a switch's state as the audit entry does.
func onOff(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

// checkOwner checks that the session on origin is the owner's and returns its
// ID.
func (s *Settings) checkOwner(ctx context.Context, origin, sessionToken string) (string, error) {
	id, err := s.owners.OwnerSession(ctx, origin, sessionToken)
	if err != nil {
		return "", fmt.Errorf("the models settings: %w", err)
	}
	return id, nil
}
