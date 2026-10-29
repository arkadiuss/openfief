package plugins

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"
)

// EffectivePermissions is what OpenFief enforces for one plugin: only what
// the plugin requests and its grant approves (D-16), with the plugin's image,
// Secret mount paths and config. Its zero value grants nothing; Effective is
// the only way to make one, so a non-nil value is always valid.
type EffectivePermissions struct {
	plugin       string
	image        ImageRef
	capabilities []Capability
	egress       []EgressHost
	secrets      []SecretMount
	dataTiers    []DataTier
	browserPath  string
	config       []byte
}

// Effective returns the intersection of what p requests and what g approves.
// Secrets match by name and keep the mount path the manifest asks for. It
// fails when p or g is invalid, when g is for another plugin, and with
// ErrNoCapabilityGranted when no requested capability is approved.
func Effective(p *Plugin, g *Grant) (*EffectivePermissions, error) {
	image, err := checkPair(p, g)
	if err != nil {
		return nil, err
	}
	req := p.Manifest.Requested()
	e := &EffectivePermissions{
		plugin:       p.Name,
		image:        image,
		capabilities: both(req.Capabilities, g.Permissions.Capabilities),
		egress:       both(req.Egress, g.Permissions.Egress),
		dataTiers:    both(req.DataTiers, g.Permissions.DataTiers),
		config:       bytes.Clone(p.Config),
	}
	for _, s := range p.Manifest.Secrets {
		if slices.Contains(g.Permissions.Secrets, s.Name) {
			e.secrets = append(e.secrets, s)
		}
	}
	slices.SortFunc(e.secrets, func(a, b SecretMount) int { return cmp.Compare(a.Name, b.Name) })
	if len(e.capabilities) == 0 {
		return nil, fmt.Errorf("%w: plugin %q", ErrNoCapabilityGranted, p.Name)
	}
	if bp := p.Manifest.BrowserPath; bp != "" && slices.Contains(g.Permissions.BrowserPaths, bp) &&
		slices.Contains(e.capabilities, CapabilityView) {
		e.browserPath = bp
	}
	return e, nil
}

// both returns the elements of want that are also in have, in want's order,
// or nil when there are none.
func both[T comparable](want, have []T) []T {
	var out []T
	for _, w := range want {
		if slices.Contains(have, w) {
			out = append(out, w)
		}
	}
	return out
}

// Plugin returns the plugin's name.
func (e *EffectivePermissions) Plugin() string { return e.plugin }

// Image returns the plugin's digest-pinned image.
func (e *EffectivePermissions) Image() ImageRef { return e.image }

// Capabilities returns the enforced capabilities, sorted.
func (e *EffectivePermissions) Capabilities() []Capability { return slices.Clone(e.capabilities) }

// Egress returns the enforced egress hosts, sorted.
func (e *EffectivePermissions) Egress() []EgressHost { return slices.Clone(e.egress) }

// Secrets returns the enforced Secret mounts, sorted by Secret name.
func (e *EffectivePermissions) Secrets() []SecretMount { return slices.Clone(e.secrets) }

// DataTiers returns the enforced data tiers, sorted.
func (e *EffectivePermissions) DataTiers() []DataTier { return slices.Clone(e.dataTiers) }

// BrowserPath returns the path prefix the core proxies to the plugin for
// browsers (D-81): the one its manifest declares, when the grant approves it
// and the plugin may act as a view; empty otherwise.
func (e *EffectivePermissions) BrowserPath() string { return e.browserPath }

// Config returns the plugin's config document.
func (e *EffectivePermissions) Config() []byte { return bytes.Clone(e.config) }

// Equal reports whether e and o enforce the same permissions for the same
// plugin, image and config.
func (e *EffectivePermissions) Equal(o *EffectivePermissions) bool {
	if e == nil || o == nil {
		return e == o
	}
	return e.plugin == o.plugin && e.image.Equal(o.image) &&
		slices.Equal(e.capabilities, o.capabilities) && slices.Equal(e.egress, o.egress) &&
		slices.Equal(e.secrets, o.secrets) && slices.Equal(e.dataTiers, o.dataTiers) &&
		e.browserPath == o.browserPath && bytes.Equal(e.config, o.config)
}
