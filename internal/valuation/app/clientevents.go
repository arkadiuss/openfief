package app

import (
	"context"
	"errors"
	"fmt"
)

// errNoClientEvents reports settings built without WithClientEvents.
var errNoClientEvents = errors.New("no store of the client events switch")

// ClientEventsStore keeps the owner's switch for the built-in UI's client
// events (D-227); *valuation/postgres.Store implements it.
type ClientEventsStore interface {
	// ClientEventsOff reports whether the owner turned the client events off.
	// Nothing stored is the default: on.
	ClientEventsOff(ctx context.Context) (bool, error)
	// SaveClientEventsOff stores that the client events are off, or, for false,
	// removes the setting so they are on again.
	SaveClientEventsOff(ctx context.Context, off bool) error
}

// WithClientEvents has s keep the owner's switch for the UI's client events
// (D-227) in events, and returns s.
func (s *Settings) WithClientEvents(events ClientEventsStore) *Settings {
	s.events = events
	return s
}

// ClientEventsEnabled reports whether the UI's pages report client events
// about themselves (D-227): true until the owner turned them off. It is the one
// setting read without a session, because the login page reports too; it says
// nothing but that.
func (s *Settings) ClientEventsEnabled(ctx context.Context) (bool, error) {
	if s.events == nil {
		return false, errNoClientEvents
	}
	off, err := s.events.ClientEventsOff(ctx)
	if err != nil {
		return false, fmt.Errorf("loading the client events switch: %w", err)
	}
	return !off, nil
}

// SetClientEventsEnabled turns the UI's client events on or off, for the owner
// (D-227), entering the change in the audit log (D-211). It revalues nothing.
func (s *Settings) SetClientEventsEnabled(ctx context.Context, origin, sessionToken string, enabled bool) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if s.events == nil {
		return errNoClientEvents
	}
	err := auditedSetting(ctx, s.audit, origin, sessionToken, settingClientEvents, func(ctx context.Context) error {
		return s.events.SaveClientEventsOff(ctx, !enabled)
	})
	if err != nil {
		return fmt.Errorf("setting the client events: %w", err)
	}
	return nil
}
