package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/arkadiuss/openfief/internal/identity/app"
)

var _ app.PluginGenerationStore = (*Store)(nil)

// LockPluginGenerations takes the table's lock until the transaction ctx
// carries ends, so one observer at a time reads what was entered and enters
// what is new. Reads of the table go on.
func (s *Store) LockPluginGenerations(ctx context.Context) error {
	if _, err := s.db(ctx).Exec(ctx, "LOCK TABLE identity.plugin_generations IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return fmt.Errorf("locking the plugin generations: %w", err)
	}
	return nil
}

// PluginGenerations returns the generation entered for each instance.
func (s *Store) PluginGenerations(ctx context.Context) (map[string]int64, error) {
	rows, err := s.db(ctx).Query(ctx, "SELECT instance, generation FROM identity.plugin_generations")
	if err != nil {
		return nil, fmt.Errorf("loading the plugin generations: %w", err)
	}
	type pair struct {
		instance   string
		generation int64
	}
	pairs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (pair, error) {
		var p pair
		if scanErr := row.Scan(&p.instance, &p.generation); scanErr != nil {
			return pair{}, fmt.Errorf("scanning a plugin generation: %w", scanErr)
		}
		return p, nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading the plugin generations: %w", err)
	}
	out := make(map[string]int64, len(pairs))
	for _, p := range pairs {
		out[p.instance] = p.generation
	}
	return out, nil
}

// SavePluginGeneration records that generation of instance was entered at at.
func (s *Store) SavePluginGeneration(ctx context.Context, instance string, generation int64, at time.Time) error {
	_, err := s.db(ctx).Exec(ctx, `
		INSERT INTO identity.plugin_generations (instance, generation, observed_at) VALUES ($1, $2, $3)
		ON CONFLICT (instance) DO UPDATE SET generation = EXCLUDED.generation, observed_at = EXCLUDED.observed_at`,
		instance, generation, at)
	if err != nil {
		return fmt.Errorf("saving the generation of %s: %w", instance, err)
	}
	return nil
}

// DeletePluginGeneration forgets instance.
func (s *Store) DeletePluginGeneration(ctx context.Context, instance string) error {
	if _, err := s.db(ctx).Exec(ctx, "DELETE FROM identity.plugin_generations WHERE instance = $1", instance); err != nil {
		return fmt.Errorf("forgetting %s: %w", instance, err)
	}
	return nil
}
