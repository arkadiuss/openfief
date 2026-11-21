package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// clientEventsKey is the valuation.settings key of the owner's switch for the
// UI's client events (D-227). The setting exists only while they are off, with
// the value clientEventsOffValue, so a store that never saw it has them on.
const (
	clientEventsKey      = "client_events"
	clientEventsOffValue = "off"
)

// ClientEventsOff reports whether the owner turned the UI's client events off
// (D-227); with no setting stored they are on.
func (s *Store) ClientEventsOff(ctx context.Context) (bool, error) {
	value, err := s.setting(ctx, clientEventsKey)
	if err != nil {
		return false, err
	}
	return value == clientEventsOffValue, nil
}

// SaveClientEventsOff stores that the client events are off, or removes the
// setting for false, which is the default, on.
func (s *Store) SaveClientEventsOff(ctx context.Context, off bool) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if off {
			return putSetting(ctx, tx, clientEventsKey, clientEventsOffValue)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM valuation.settings WHERE key = $1`, clientEventsKey); err != nil {
			return fmt.Errorf("clearing the setting %s: %w", clientEventsKey, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving the client events switch: %w", err)
	}
	return nil
}
