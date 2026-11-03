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
	// clusterService is the granted in-cluster service (D-74); nil if none.
	clusterService *ClusterService
	// marketData: the plugin is a market-data connector (D-123).
	marketData bool
	// bondTerms and indexLevels: the connector may deliver treasury bond
	// series terms and market index levels (D-179).
	bondTerms, indexLevels bool
	// indexReadings: the connector may deliver the index readings bond
	// terms depend on, such as CPI and the NBP reference rate (D-186).
	indexReadings bool
	// institutions: those the connector books for, declared and approved
	// (D-185), sorted.
	institutions []string
}

// Effective returns the intersection of what p requests and what g approves.
// Secrets match by the name the instance mounts for each role (D-89) and
// keep the mount path the manifest asks for. It
// fails when p or g is invalid, when g is for another plugin, and with
// ErrNoCapabilityGranted when no requested capability is approved.
func Effective(p *Plugin, g *Grant) (*EffectivePermissions, error) {
	image, err := checkPair(p, g)
	if err != nil {
		return nil, err
	}
	req := p.Requested()
	e := &EffectivePermissions{
		plugin:       p.Name,
		image:        image,
		capabilities: both(req.Capabilities, g.Permissions.Capabilities),
		egress:       both(req.Egress, g.Permissions.Egress),
		dataTiers:    both(req.DataTiers, g.Permissions.DataTiers),
		config:       bytes.Clone(p.Config),
	}
	for _, s := range p.Manifest.Secrets {
		if slices.Contains(g.Permissions.Secrets, SecretName(p.Name, s.Role)) {
			e.secrets = append(e.secrets, s)
		}
	}
	slices.SortFunc(e.secrets, func(a, b SecretMount) int { return cmp.Compare(a.Role, b.Role) })
	if len(e.capabilities) == 0 {
		return nil, fmt.Errorf("%w: plugin %q", ErrNoCapabilityGranted, p.Name)
	}
	e.grantSingles(p, g, &req)
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

// Secrets returns the enforced Secret mounts, sorted by role.
func (e *EffectivePermissions) Secrets() []SecretMount { return slices.Clone(e.secrets) }

// SecretNames returns the names of the enforced Secrets this instance
// mounts, sorted (D-89).
func (e *EffectivePermissions) SecretNames() []string {
	out := make([]string, 0, len(e.secrets))
	for _, s := range e.secrets {
		out = append(out, SecretName(e.plugin, s.Role))
	}
	return out
}

// DataTiers returns the enforced data tiers, sorted.
func (e *EffectivePermissions) DataTiers() []DataTier { return slices.Clone(e.dataTiers) }

// BrowserPath returns the path prefix the core proxies to the plugin for
// browsers (D-81): the one its manifest declares, when the grant approves it
// and the plugin may act as a view; empty otherwise.
func (e *EffectivePermissions) BrowserPath() string { return e.browserPath }

// MarketData reports whether the plugin is a market-data connector: its
// manifest declares it, the grant approves it and it may act as a
// connector (D-123). Only such a plugin is told, on each sync, the
// instruments the owner holds and the model portfolio's targets (D-93).
func (e *EffectivePermissions) MarketData() bool { return e.marketData }

// BondTerms reports whether the plugin may deliver treasury bond series
// terms: its manifest declares them, the grant approves that and it may
// act as a connector (D-179).
func (e *EffectivePermissions) BondTerms() bool { return e.bondTerms }

// IndexLevels reports whether the plugin may deliver market index levels:
// its manifest declares them, the grant approves that and it may act as a
// connector (D-179).
func (e *EffectivePermissions) IndexLevels() bool { return e.indexLevels }

// IndexReadings reports whether the plugin may deliver index readings, such
// as CPI and the NBP reference rate: its manifest declares them, the grant
// approves that and it may act as a connector (D-186).
func (e *EffectivePermissions) IndexReadings() bool { return e.indexReadings }

// Institutions returns the institutions the plugin books for, sorted: those
// its manifest declares and its grant approves, when it may act as a
// connector (D-185). The core refuses a delivery naming any other.
func (e *EffectivePermissions) Institutions() []string { return slices.Clone(e.institutions) }

// Config returns the plugin's config document.
func (e *EffectivePermissions) Config() []byte { return bytes.Clone(e.config) }

// ClusterService returns the in-cluster service the plugin may reach
// directly, the one it requests and its grant approves (D-74), and whether
// there is one.
func (e *EffectivePermissions) ClusterService() (ClusterService, bool) {
	if e.clusterService == nil {
		return ClusterService{}, false
	}
	return *e.clusterService, true
}

// Equal reports whether e and o enforce the same permissions for the same
// plugin, image and config.
func (e *EffectivePermissions) Equal(o *EffectivePermissions) bool {
	if e == nil || o == nil {
		return e == o
	}
	return e.plugin == o.plugin && e.image.Equal(o.image) && bytes.Equal(e.config, o.config) && e.samePermissions(o)
}

// samePermissions reports whether e and o enforce the same permissions.
func (e *EffectivePermissions) samePermissions(o *EffectivePermissions) bool {
	return slices.Equal(e.capabilities, o.capabilities) && slices.Equal(e.egress, o.egress) &&
		slices.Equal(e.secrets, o.secrets) && slices.Equal(e.dataTiers, o.dataTiers) &&
		e.browserPath == o.browserPath && equalPtr(e.clusterService, o.clusterService) && e.sameDeliveries(o)
}

// sameDeliveries reports whether e and o enforce the same connector
// permissions: market data (D-123), bond terms, index levels (D-179),
// index readings (D-186) and institutions (D-185).
func (e *EffectivePermissions) sameDeliveries(o *EffectivePermissions) bool {
	return e.marketData == o.marketData && e.bondTerms == o.bondTerms && e.indexLevels == o.indexLevels &&
		e.indexReadings == o.indexReadings && slices.Equal(e.institutions, o.institutions)
}

// equalPtr reports whether a and b are both nil or point to equal values.
func equalPtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// grantSingles sets the permissions a plugin holds at most one of: the
// browser path, when the grant approves it and the plugin may act as a
// view (D-81); the in-cluster service (D-74); and market data, when the
// manifest declares it, the grant approves it and the plugin may act as a
// connector (D-123), and so bond terms and index levels (grantDeliveries).
func (e *EffectivePermissions) grantSingles(p *Plugin, g *Grant, req *Permissions) {
	if bp := p.Manifest.BrowserPath; bp != "" && slices.Contains(g.Permissions.BrowserPaths, bp) &&
		slices.Contains(e.capabilities, CapabilityView) {
		e.browserPath = bp
	}
	if cs := both(req.ClusterServices, g.Permissions.ClusterServices); len(cs) == 1 {
		e.clusterService = &cs[0]
	}
	e.marketData = req.MarketData && g.Permissions.MarketData && slices.Contains(e.capabilities, CapabilityConnector)
	e.grantDeliveries(g, req)
}

// grantDeliveries sets whether the plugin may deliver bond terms, index
// levels and index readings: each when the manifest declares it, the grant
// approves it and the plugin may act as a connector (D-179, D-186); and so
// the institutions it books for (D-185).
func (e *EffectivePermissions) grantDeliveries(g *Grant, req *Permissions) {
	connector := slices.Contains(e.capabilities, CapabilityConnector)
	e.bondTerms = req.BondTerms && g.Permissions.BondTerms && connector
	e.indexLevels = req.IndexLevels && g.Permissions.IndexLevels && connector
	e.indexReadings = req.IndexReadings && g.Permissions.IndexReadings && connector
	if connector {
		e.institutions = both(req.Institutions, g.Permissions.Institutions)
	}
}
