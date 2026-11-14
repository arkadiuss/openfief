package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	viewskube "github.com/arkadiuss/openfief/internal/views/kube"
)

// pluginObserveEvery is how often the core looks at the Plugin resources to
// enter what the operator did to them in the audit log, so an install, an
// upgrade or a removal is entered within a minute of the core seeing it.
const pluginObserveEvery = time.Minute

// observedPlugins returns what the audit log keeps of records: the instance
// each Plugin resource names, its generation and whether it is being deleted.
func observedPlugins(records []views.PluginRecord) []identityapp.ObservedPlugin {
	out := make([]identityapp.ObservedPlugin, 0, len(records))
	for i := range records {
		out = append(out, identityapp.ObservedPlugin{Instance: records[i].Instance, Generation: records[i].Generation, Deleting: records[i].Deleting})
	}
	return out
}

// startPluginAudit starts, when the core runs in a cluster (installed is nil
// outside one), the loop that enters what the operator did to the plugins in
// the audit log, in the background: it can't fail the core.
func startPluginAudit(ctx context.Context, pool *pgxpool.Pool, installed *viewskube.Plugins, now func() time.Time, log *slog.Logger) {
	if installed == nil {
		return
	}
	go observePlugins(ctx, identityapp.NewPluginObservations(identitypg.New(pool), now), installed, pluginObserveEvery, log)
}

// observePlugins enters in the audit log, as the operator's, each install,
// upgrade and removal of a plugin that the core sees in installed (D-211): a
// Plugin resource's generation is what the operator's new spec makes, and
// the operator is another process that can't be made to wait for an entry, so
// the core records what it observes, once per instance and generation
// (identityapp.PluginObservations). It runs until ctx ends and tries again
// until an observation is written: a plugin cache that has not synced is asked
// again soon, and a failed write after ten seconds, with nothing entered.
// What it logs is the class of an error, never its text.
func observePlugins(ctx context.Context, observations *identityapp.PluginObservations, installed installedPlugins, every time.Duration, log *slog.Logger) {
	var wait time.Duration
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		wait = every
		records, err := installed.InstalledPlugins(ctx)
		if errors.Is(err, viewsapp.ErrPluginsUnknown) {
			wait = grantsRetryUnknown
			continue
		}
		if err != nil {
			log.Warn("the installed plugins could not be listed for the audit log; retrying", "error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded))
			wait = grantsRetryFailed
			continue
		}
		n, err := observations.Observe(ctx, observedPlugins(records))
		if err != nil {
			log.Warn("what the operator did to the plugins is not in the audit log yet; retrying", "error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded))
			wait = grantsRetryFailed
			continue
		}
		if n > 0 {
			log.Info("entered what the operator did to the plugins in the audit log", "entries", n)
		}
	}
}
