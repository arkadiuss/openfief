package main

import (
	"context"
	"errors"
	"fmt"
	"slices"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	attentionapp "github.com/arkadiuss/openfief/internal/attention/app"
	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/views"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
	viewskube "github.com/arkadiuss/openfief/internal/views/kube"
)

// readyTrue is the status of a condition that holds.
const readyTrue = "True"

// ungrantedDetail follows the instance in an ungranted view plugin's item.
const ungrantedDetail = " can't read anything until you grant it accounts"

// grantsInForce is what the check of ungranted view plugins reads of the
// owner's grants; *identityapp.Grants implements it.
type grantsInForce interface {
	InForce(ctx context.Context) ([]identity.Grant, error)
}

// ungrantedViewPluginsIn is ungrantedViewPlugins for the core's configuration
// and the plugins its cluster runs, nil outside a cluster.
func ungrantedViewPluginsIn(cfg *serveConfig, installed *viewskube.Plugins, grants grantsInForce) attentionapp.Check {
	var lister installedPlugins
	if installed != nil {
		lister = installed
	}
	return ungrantedViewPlugins(cfg.Authz.Mode == authzSpiceDB, lister, grants)
}

// ungrantedViewPlugins is the attention check of view plugins that can't read
// anything (D-196, D-209): while the core answers from SpiceDB, a view plugin
// that is installed, runs (its Ready condition is True) and holds the view
// capability, and that holds no grant in force, views nothing until the owner
// grants it accounts. Each is a condition of the kind view-plugin-ungranted,
// its subject the instance, which clears when a grant exists or the plugin is
// uninstalled. The plugin that serves browsers, the built-in UI, is exempt:
// it reads with the signed-in person's session, never its own identity. With
// the policy there is nothing to grant, since every holder of read access
// views every account, so the check finds nothing; nor does it outside a
// cluster, where installed is nil. A plugin cache that has not synced is an
// error, so no item is cleared on a guess.
func ungrantedViewPlugins(spicedbOn bool, installed installedPlugins, grants grantsInForce) attentionapp.Check {
	return attentionapp.Check{Kind: attentiondomain.KindViewPluginUngranted, Find: func(ctx context.Context) ([]attentiondomain.Condition, error) {
		if !spicedbOn || installed == nil {
			return nil, nil
		}
		records, err := installed.InstalledPlugins(ctx)
		if errors.Is(err, viewsapp.ErrPluginsUnknown) {
			return nil, fmt.Errorf("the installed plugins are not known yet: %w", err)
		}
		if err != nil {
			return nil, fmt.Errorf("listing the installed plugins: %w", err)
		}
		held, err := pluginsWithGrants(ctx, grants)
		if err != nil {
			return nil, err
		}
		return ungrantedConditions(records, held), nil
	}}
}

// pluginsWithGrants returns the view plugin instances that hold a grant in
// force, of any target.
func pluginsWithGrants(ctx context.Context, grants grantsInForce) (map[string]bool, error) {
	inForce, err := grants.InForce(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing the grants: %w", err)
	}
	held := make(map[string]bool, len(inForce))
	for i := range inForce {
		if inForce[i].Holder.Kind == identity.GrantHolderViewPlugin {
			held[inForce[i].Holder.ID] = true
		}
	}
	return held, nil
}

// ungrantedConditions returns the condition of each of records that needs a
// grant (needsGrant) and is not among the instances that hold one.
func ungrantedConditions(records []views.PluginRecord, held map[string]bool) []attentiondomain.Condition {
	var out []attentiondomain.Condition
	for i := range records {
		if needsGrant(&records[i]) && !held[records[i].Instance] {
			out = append(out, attentiondomain.Condition{
				Kind: attentiondomain.KindViewPluginUngranted, Subject: records[i].Instance, Detail: records[i].Instance + ungrantedDetail,
			})
		}
	}
	return out
}

// needsGrant reports whether r is a view plugin that runs and reads with its
// own identity: it is not being removed, is Ready, holds the view capability
// and does not serve browsers.
func needsGrant(r *views.PluginRecord) bool {
	return !r.Deleting && r.Ready.Status == readyTrue && slices.Contains(r.Effective, string(plugins.CapabilityView)) && r.BrowserPath == ""
}
