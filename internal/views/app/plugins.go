package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/views"
)

// ErrPluginsUnknown reports that the installed plugins can't be read: the
// core runs outside a cluster, or has not read the cluster yet.
var ErrPluginsUnknown = errors.New("the installed plugins are not known")

// Plugins reads the installed plugin instances from the operator's Plugin
// resources (D-46). It only reads: the core can't change a plugin or a
// grant. internal/views/kube implements it; while it can't answer it fails
// with an error wrapping ErrPluginsUnknown.
type Plugins interface {
	// InstalledPlugins returns every Plugin resource, in any order.
	InstalledPlugins(ctx context.Context) ([]views.PluginRecord, error)
}

// WithPlugins reads the installed plugins from plugins. Without it they are
// unknown: Plugins fails with ErrPluginsUnknown and no transaction is
// marked as coming from a removed instance.
func WithPlugins(plugins Plugins) Option {
	return func(v *Views) { v.plugins = plugins }
}

// Plugins returns the installed plugin instances in name order, with their
// trust (D-56), or an error wrapping ErrPluginsUnknown when they can't be
// read.
func (v *Views) Plugins(ctx context.Context) ([]views.Plugin, error) {
	if v.plugins == nil {
		return nil, ErrPluginsUnknown
	}
	records, err := v.plugins.InstalledPlugins(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading installed plugins: %w", err)
	}
	out := make([]views.Plugin, 0, len(records))
	for i := range records {
		out = append(out, views.PluginOf(&records[i]))
	}
	slices.SortFunc(out, func(a, b views.Plugin) int { return cmp.Compare(a.Instance, b.Instance) })
	return out, nil
}

// installed returns the installed instances, or nothing known when they
// can't be read: a transaction listing never fails for want of them, it
// only marks nothing as removed.
func (v *Views) installed(ctx context.Context) views.Installed {
	plugins, err := v.Plugins(ctx)
	if err != nil {
		return views.Installed{}
	}
	return views.InstalledOf(plugins)
}
