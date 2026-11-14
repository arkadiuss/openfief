package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
)

// ErrCorrupt reports a stored row the modelrouter context can't read back.
var ErrCorrupt = errors.New("corrupt models row")

var _ app.SettingsStore = (*Store)(nil)

// Store keeps the owner's agents-and-models settings in the modelrouter schema.
// It implements app.SettingsStore.
type Store struct {
	pool *pgxpool.Pool
}

// New returns the store over pool. Run Migrate first.
func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// Settings returns the stored settings: the switch and its toggles, every
// route and every tier's allowed models. With nothing ever saved it returns
// the zero Settings, everything off (D-212).
func (s *Store) Settings(ctx context.Context) (modelrouter.Settings, error) {
	var out modelrouter.Settings
	err := s.pool.QueryRow(ctx, "SELECT enabled FROM modelrouter.switch").Scan(&out.Switch.Enabled)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return modelrouter.Settings{}, fmt.Errorf("loading the models switch: %w", err)
	}
	if out.Switch.Toggles, err = s.toggles(ctx); err != nil {
		return modelrouter.Settings{}, err
	}
	if out.Routes, err = s.routes(ctx); err != nil {
		return modelrouter.Settings{}, err
	}
	if out.Policy, err = s.policy(ctx); err != nil {
		return modelrouter.Settings{}, err
	}
	return out, nil
}

// SaveEnabled turns the one switch on or off, running within before the
// commit.
func (s *Store) SaveEnabled(ctx context.Context, on bool, within func(context.Context) error) error {
	err := s.inTx(ctx, within, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO modelrouter.switch (enabled) VALUES ($1)
			ON CONFLICT (singleton) DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = now()`, on)
		if err != nil {
			return fmt.Errorf("writing the switch: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving the models switch: %w", err)
	}
	return nil
}

// SaveToggle turns one feature's toggle on or off, running within before the
// commit.
func (s *Store) SaveToggle(ctx context.Context, f modelrouter.Feature, on bool, within func(context.Context) error) error {
	err := s.inTx(ctx, within, func(tx pgx.Tx) error {
		var err error
		if on {
			_, err = tx.Exec(ctx, `
				INSERT INTO modelrouter.feature_toggles (feature) VALUES ($1)
				ON CONFLICT (feature) DO UPDATE SET updated_at = now()`, string(f))
		} else {
			_, err = tx.Exec(ctx, "DELETE FROM modelrouter.feature_toggles WHERE feature = $1", string(f))
		}
		if err != nil {
			return fmt.Errorf("writing the toggle: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving a feature toggle: %w", err)
	}
	return nil
}

// SaveRoute replaces role's route with route in one transaction; no models
// clear it. within runs before the commit.
func (s *Store) SaveRoute(ctx context.Context, role modelrouter.Role, route []modelrouter.Ref, within func(context.Context) error) error {
	err := s.inTx(ctx, within, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "DELETE FROM modelrouter.routes WHERE role = $1", string(role)); err != nil {
			return fmt.Errorf("clearing the route: %w", err)
		}
		for i, m := range route {
			_, err := tx.Exec(ctx, "INSERT INTO modelrouter.routes (role, position, provider, model) VALUES ($1, $2, $3, $4)",
				string(role), i, m.Instance, m.Model)
			if err != nil {
				return fmt.Errorf("writing the route: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving a route: %w", err)
	}
	return nil
}

// SaveDataTierModels replaces tier's allowed models with allowed in one
// transaction; none clear the list. within runs before the commit.
func (s *Store) SaveDataTierModels(ctx context.Context, tier modelrouter.DataTier, allowed []modelrouter.Ref, within func(context.Context) error) error {
	err := s.inTx(ctx, within, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "DELETE FROM modelrouter.tier_models WHERE tier = $1", string(tier)); err != nil {
			return fmt.Errorf("clearing the list: %w", err)
		}
		for i, m := range allowed {
			_, err := tx.Exec(ctx, "INSERT INTO modelrouter.tier_models (tier, position, provider, model) VALUES ($1, $2, $3, $4)",
				string(tier), i, m.Instance, m.Model)
			if err != nil {
				return fmt.Errorf("writing the list: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("saving a tier's models: %w", err)
	}
	return nil
}

// inTx runs write in a transaction, then within when it is not nil, and
// commits; a failure of either stores nothing.
func (s *Store) inTx(ctx context.Context, within func(context.Context) error, write func(tx pgx.Tx) error) error {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := write(tx); err != nil {
			return err
		}
		if within == nil {
			return nil
		}
		return within(ctx)
	})
	if err != nil {
		return fmt.Errorf("storing a change in one transaction: %w", err)
	}
	return nil
}

// toggles returns the features whose toggle is on.
func (s *Store) toggles(ctx context.Context) (map[modelrouter.Feature]bool, error) {
	rows, err := s.pool.Query(ctx, "SELECT feature FROM modelrouter.feature_toggles")
	if err != nil {
		return nil, fmt.Errorf("loading the feature toggles: %w", err)
	}
	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("loading the feature toggles: %w", err)
	}
	var out map[modelrouter.Feature]bool
	for _, n := range names {
		f, err := modelrouter.ParseFeature(n)
		if err != nil {
			return nil, fmt.Errorf("%w: a feature toggle: %w", ErrCorrupt, err)
		}
		if out == nil {
			out = map[modelrouter.Feature]bool{}
		}
		out[f] = true
	}
	return out, nil
}

// refRow is a stored reference with the list it belongs to.
type refRow struct {
	list, provider, model string
}

// refRows returns the rows of query, each a list name, a provider and a
// model, in the order of the query.
func (s *Store) refRows(ctx context.Context, query string) ([]refRow, error) {
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("loading references: %w", err)
	}
	out, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (refRow, error) {
		var x refRow
		if scanErr := r.Scan(&x.list, &x.provider, &x.model); scanErr != nil {
			return refRow{}, fmt.Errorf("scanning a reference: %w", scanErr)
		}
		return x, nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading references: %w", err)
	}
	return out, nil
}

// routes returns every role's route.
func (s *Store) routes(ctx context.Context) (map[modelrouter.Role]modelrouter.Route, error) {
	rows, err := s.refRows(ctx, "SELECT role, provider, model FROM modelrouter.routes ORDER BY role, position")
	if err != nil {
		return nil, fmt.Errorf("loading the routes: %w", err)
	}
	byRole := map[modelrouter.Role][]modelrouter.Ref{}
	for _, r := range rows {
		role, parseErr := modelrouter.ParseRole(r.list)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: a route: %w", ErrCorrupt, parseErr)
		}
		byRole[role] = append(byRole[role], modelrouter.Ref{Instance: r.provider, Model: r.model})
	}
	var out map[modelrouter.Role]modelrouter.Route
	for role, refs := range byRole {
		route, routeErr := modelrouter.NewRoute(role, refs)
		if routeErr != nil {
			return nil, fmt.Errorf("%w: a route: %w", ErrCorrupt, routeErr)
		}
		if out == nil {
			out = map[modelrouter.Role]modelrouter.Route{}
		}
		out[role] = route
	}
	return out, nil
}

// policy returns the models allowed per tier.
func (s *Store) policy(ctx context.Context) (modelrouter.DataTierPolicy, error) {
	rows, err := s.refRows(ctx, "SELECT tier, provider, model FROM modelrouter.tier_models ORDER BY tier, position")
	if err != nil {
		return nil, fmt.Errorf("loading the tiers' models: %w", err)
	}
	lists := map[modelrouter.DataTier][]modelrouter.Ref{}
	for _, r := range rows {
		tier, parseErr := modelrouter.ParseDataTier(r.list)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: a tier's models: %w", ErrCorrupt, parseErr)
		}
		lists[tier] = append(lists[tier], modelrouter.Ref{Instance: r.provider, Model: r.model})
	}
	if len(lists) == 0 {
		return modelrouter.DataTierPolicy{}, nil
	}
	p, err := modelrouter.NewDataTierPolicy(lists)
	if err != nil {
		return nil, fmt.Errorf("%w: a tier's models: %w", ErrCorrupt, err)
	}
	return p, nil
}
