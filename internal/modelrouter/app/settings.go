package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// Settings are the owner's agents-and-models settings use cases (D-212):
// reading them, and changing the one switch, a feature's toggle, a role's
// route and a tier's allowed models, each change checked against the
// owner's session.
type Settings struct {
	store  SettingsStore
	owners Owners
}

// NewSettings returns the settings use cases over store, with the owner's
// sessions checked by owners.
func NewSettings(store SettingsStore, owners Owners) *Settings {
	return &Settings{store: store, owners: owners}
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
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
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
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if err := s.store.SaveEnabled(ctx, on); err != nil {
		return fmt.Errorf("setting the models switch: %w", err)
	}
	return nil
}

// SetFeature turns a feature's toggle on or off, for the owner. An unknown
// feature is an error wrapping ErrInvalidInput.
func (s *Settings) SetFeature(ctx context.Context, origin, sessionToken, feature string, on bool) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	f, err := modelrouter.ParseFeature(feature)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err := s.store.SaveToggle(ctx, f, on); err != nil {
		return fmt.Errorf("setting a feature toggle: %w", err)
	}
	return nil
}

// SetRoute sets a role's primary model and fallbacks from "provider/model"
// references, for the owner; no references clear the route. An unknown
// role, a bad reference or a route breaking modelrouter.NewRoute is an error
// wrapping ErrInvalidInput.
func (s *Settings) SetRoute(ctx context.Context, origin, sessionToken, role string, refs []string) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
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
	if err := s.store.SaveRoute(ctx, r, parsed); err != nil {
		return fmt.Errorf("setting a route: %w", err)
	}
	return nil
}

// SetTierModels sets the models allowed to receive a data tier from
// "provider/model" references, for the owner; none clear the list, which
// denies the tier to every model (D-206). An unknown tier, a bad reference
// or a list breaking modelrouter.NewTierPolicy is an error wrapping
// ErrInvalidInput.
func (s *Settings) SetTierModels(ctx context.Context, origin, sessionToken, tier string, refs []string) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	t, err := modelrouter.ParseTier(tier)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	parsed, err := modelrouter.ParseRefs(refs)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if _, err := modelrouter.NewTierPolicy(map[modelrouter.Tier][]modelrouter.Ref{t: parsed}); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err := s.store.SaveTierModels(ctx, t, parsed); err != nil {
		return fmt.Errorf("setting a tier's models: %w", err)
	}
	return nil
}

// checkOwner checks that the session on origin is the owner's.
func (s *Settings) checkOwner(ctx context.Context, origin, sessionToken string) error {
	if err := s.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("the models settings: %w", err)
	}
	return nil
}
