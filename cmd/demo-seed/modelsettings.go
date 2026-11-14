package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	modelspg "github.com/arkadiuss/openfief/internal/modelrouter/postgres"
)

// errBadModelSettings reports a model settings file the loader refuses: not
// JSON of the two members it knows, or a route or a tier list the model
// router would refuse. It names the member, never the file's text.
var errBadModelSettings = errors.New("invalid model settings")

// modelSettingsFile is the file --model-settings names: the demo's routes and
// tier lists, as "provider/model" references (docs/concepts/models.md). It has
// no member for the switch or a toggle, and an unknown member is refused, so
// seeding never turns a model on: that is the owner's act in Settings (D-212).
type modelSettingsFile struct {
	// Routes holds each role's models, primary first, by the role's name.
	Routes map[string][]string `json:"routes"`
	// Tiers holds the models allowed to receive each data tier, by the
	// tier's name (D-206).
	Tiers map[string][]string `json:"tiers"`
}

// modelSettings is a checked modelSettingsFile.
type modelSettings struct {
	routes map[modelrouter.Role][]modelrouter.Ref
	tiers  map[modelrouter.DataTier][]modelrouter.Ref
}

// parseModelSettings checks data as the router would check the owner's own
// routes and tier lists: known roles and tiers, references of the form
// provider/model, a route of one to MaxRouteModels distinct models.
func parseModelSettings(data []byte) (modelSettings, error) {
	var f modelSettingsFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return modelSettings{}, fmt.Errorf("%w: %w", errBadModelSettings, err)
	}
	out := modelSettings{
		routes: make(map[modelrouter.Role][]modelrouter.Ref, len(f.Routes)),
		tiers:  make(map[modelrouter.DataTier][]modelrouter.Ref, len(f.Tiers)),
	}
	for name, refs := range f.Routes {
		role, err := modelrouter.ParseRole(name)
		if err != nil {
			return modelSettings{}, fmt.Errorf("%w: routes: %w", errBadModelSettings, err)
		}
		parsed, err := modelrouter.ParseRefs(refs)
		if err != nil {
			return modelSettings{}, fmt.Errorf("%w: route %s: %w", errBadModelSettings, name, err)
		}
		if _, err := modelrouter.NewRoute(role, parsed); err != nil {
			return modelSettings{}, fmt.Errorf("%w: route %s: %w", errBadModelSettings, name, err)
		}
		out.routes[role] = parsed
	}
	for name, refs := range f.Tiers {
		tier, err := modelrouter.ParseDataTier(name)
		if err != nil {
			return modelSettings{}, fmt.Errorf("%w: tiers: %w", errBadModelSettings, err)
		}
		parsed, err := modelrouter.ParseRefs(refs)
		if err != nil {
			return modelSettings{}, fmt.Errorf("%w: tier %s: %w", errBadModelSettings, name, err)
		}
		if _, err := modelrouter.NewDataTierPolicy(map[modelrouter.DataTier][]modelrouter.Ref{tier: parsed}); err != nil {
			return modelSettings{}, fmt.Errorf("%w: tier %s: %w", errBadModelSettings, name, err)
		}
		out.tiers[tier] = parsed
	}
	return out, nil
}

// modelSettingsStore is the part of the models settings store the seed uses;
// *modelspg.Store implements it.
type modelSettingsStore interface {
	Settings(ctx context.Context) (modelrouter.Settings, error)
	SaveRoute(ctx context.Context, role modelrouter.Role, route []modelrouter.Ref, within func(context.Context) error) error
	SaveDataTierModels(ctx context.Context, tier modelrouter.DataTier, allowed []modelrouter.Ref, within func(context.Context) error) error
}

// seedModelSettings stores the routes and tier lists of ms when the store
// holds none of either, and reports whether it did. What the owner set, or an
// earlier seed stored, is left as it is: a Job that runs again on every deploy
// never overwrites a choice made in Settings. It never touches the switch or a
// toggle, which stay as they are (off, until the owner turns them on).
func seedModelSettings(ctx context.Context, store modelSettingsStore, ms *modelSettings) (bool, error) {
	cur, err := store.Settings(ctx)
	if err != nil {
		return false, fmt.Errorf("reading the models settings: %w", err)
	}
	if len(cur.Routes) > 0 || len(cur.Policy) > 0 {
		return false, nil
	}
	for _, role := range modelrouter.Roles() {
		if refs, ok := ms.routes[role]; ok {
			if err := store.SaveRoute(ctx, role, refs, nil); err != nil {
				return false, fmt.Errorf("storing the %s route: %w", role, err)
			}
		}
	}
	for _, tier := range modelrouter.DataTiers() {
		if refs, ok := ms.tiers[tier]; ok {
			if err := store.SaveDataTierModels(ctx, tier, refs, nil); err != nil {
				return false, fmt.Errorf("storing the models allowed for %s: %w", tier, err)
			}
		}
	}
	return true, nil
}

// applyModelSettings reads the file at path, migrates the modelrouter schema
// (the core migrates it too, under the same advisory lock) and seeds it.
func applyModelSettings(ctx context.Context, pool *pgxpool.Pool, path string, log *slog.Logger) error {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("%w: opening the directory of --model-settings", errBadModelSettings)
	}
	data, readErr := root.ReadFile(filepath.Base(path))
	if closeErr := root.Close(); readErr != nil || closeErr != nil {
		return fmt.Errorf("%w: reading --model-settings", errBadModelSettings)
	}
	ms, err := parseModelSettings(data)
	if err != nil {
		return err
	}
	if migrateErr := modelspg.Migrate(ctx, pool); migrateErr != nil {
		return fmt.Errorf("migrating the models settings: %w", migrateErr)
	}
	applied, err := seedModelSettings(ctx, modelspg.New(pool), &ms)
	if err != nil {
		return err
	}
	log.Info("demo model settings", "applied", applied, "routes", len(ms.routes), "tiers", len(ms.tiers))
	return nil
}
