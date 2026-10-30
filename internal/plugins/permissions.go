package plugins

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// Permissions is a set of plugin permissions: capabilities, egress hosts,
// Secret names, data tiers, browser paths (D-81) and in-cluster services
// (D-74). It is what a plugin requests (Plugin.Requested), what a Grant
// approves, and what NeedsApproval reports as still unapproved.
type Permissions struct {
	Capabilities    []Capability
	Egress          []EgressHost
	Secrets         []string
	DataTiers       []DataTier
	BrowserPaths    []string
	ClusterServices []ClusterService
}

// IsEmpty reports whether p holds no permission at all.
func (p *Permissions) IsEmpty() bool {
	return len(p.Capabilities) == 0 && len(p.Egress) == 0 && len(p.Secrets) == 0 && len(p.DataTiers) == 0 &&
		len(p.BrowserPaths) == 0 && len(p.ClusterServices) == 0
}

func (p *Permissions) validate() error {
	errs := make([]error, 0, len(p.Capabilities)+len(p.Egress)+len(p.Secrets)+len(p.DataTiers)+len(p.BrowserPaths))
	errs = append(errs, validateCapabilities(p.Capabilities)...)
	errs = append(errs, validateEgress(p.Egress)...)
	errs = append(errs, validateSecretNames("secrets", p.Secrets)...)
	errs = append(errs, validateDataTiers(p.DataTiers)...)
	errs = append(errs, validateBrowserPaths("browserPaths", p.BrowserPaths)...)
	errs = append(errs, validateClusterServices("clusterServices", p.ClusterServices)...)
	return errors.Join(errs...)
}

// sorted returns p with every list sorted, so results compare and render
// deterministically. Empty lists become nil.
func (p *Permissions) sorted() Permissions {
	return Permissions{
		Capabilities: sortedCopy(p.Capabilities, cmp.Compare[Capability]),
		Egress:       sortedCopy(p.Egress, compareEgressHosts),
		Secrets:      sortedCopy(p.Secrets, cmp.Compare[string]),
		DataTiers:    sortedCopy(p.DataTiers, cmp.Compare[DataTier]),
		BrowserPaths: sortedCopy(p.BrowserPaths, cmp.Compare[string]),
		ClusterServices: sortedCopy(p.ClusterServices, func(a, b ClusterService) int {
			return cmp.Compare(a.String(), b.String())
		}),
	}
}

// Plugin is what a plugin's vendor requests: its name, a digest-pinned image,
// its manifest and its config (D-13, D-16). Name is the Plugin resource's
// name; Image is parsed with ParseImageRef; Config is a JSON document.
// ClusterService is the in-cluster service the owner picked at install for a
// manifest that asks for one (D-74); nil when there is none.
type Plugin struct {
	Name           string
	Image          string
	Manifest       Manifest
	Config         []byte
	ClusterService *ClusterService
}

// Requested returns the permissions p asks for, sorted: its manifest's
// capabilities, egress hosts, data tiers and browser path, the Secret each
// of its manifest's roles names for this instance (D-89), the host of each
// config field its manifest names (D-75) and its in-cluster service (D-74).
// A config host that can't be read is left out; Validate reports it.
func (p *Plugin) Requested() Permissions {
	m := &p.Manifest
	names := make([]string, 0, len(m.Secrets))
	for _, s := range m.Secrets {
		names = append(names, SecretName(p.Name, s.Role))
	}
	r := Permissions{
		Capabilities: m.Capabilities,
		Egress:       slices.Clone(m.Egress),
		Secrets:      names,
		DataTiers:    m.DataTiers,
	}
	if hosts, err := ConfigHosts(m, p.Config); err == nil {
		for _, h := range hosts {
			if !slices.Contains(r.Egress, h.Host) {
				r.Egress = append(r.Egress, h.Host)
			}
		}
	}
	if m.BrowserPath != "" {
		r.BrowserPaths = []string{m.BrowserPath}
	}
	if p.ClusterService != nil {
		r.ClusterServices = []ClusterService{*p.ClusterService}
	}
	return r.sorted()
}

// Validate returns every problem in p, each a *FieldError, joined; or nil.
func (p *Plugin) Validate() error {
	var errs []error
	if !validPluginName(p.Name) {
		errs = append(errs, fieldErr("name", ErrInvalidName, p.Name))
	}
	if _, err := ParseImageRef(p.Image); err != nil {
		errs = append(errs, &FieldError{Field: "image", Err: err})
	}
	if err := p.Manifest.Validate(); err != nil {
		errs = append(errs, prefixed("manifest.", err)...)
	}
	if len(p.Config) > 0 && !json.Valid(p.Config) {
		errs = append(errs, &FieldError{Field: "config", Err: ErrInvalidJSON})
	} else if _, err := ConfigHosts(&p.Manifest, p.Config); err != nil {
		errs = append(errs, prefixed("", err)...)
	}
	errs = append(errs, p.validateClusterService()...)
	return errors.Join(errs...)
}

// Grant is what the owner approved for the plugin named Plugin (D-16).
// A Grant may approve more than a plugin requests; only what both hold is
// enforced (Effective).
type Grant struct {
	Plugin      string
	Permissions Permissions
}

// Validate returns every problem in g, each a *FieldError, joined; or nil.
func (g *Grant) Validate() error {
	var errs []error
	if !validPluginName(g.Plugin) {
		errs = append(errs, fieldErr("plugin", ErrInvalidName, g.Plugin))
	}
	if err := g.Permissions.validate(); err != nil {
		errs = append(errs, prefixed("permissions.", err)...)
	}
	return errors.Join(errs...)
}

// validPluginName checks that the name can be a Kubernetes label value and
// part of resource names: at most 63 lower-case letters, digits and '-',
// starting with a letter and ending with a letter or digit, with no "--" in
// the third and fourth places (the same label rule as EgressHost).
func validPluginName(n string) bool {
	return validateLabel(n) == nil && !isDigit(rune(n[0]))
}

// prefixed returns the errors joined in err, with prefix put before the Field
// of each *FieldError.
func prefixed(prefix string, err error) []error {
	var errs []error
	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		errs = joined.Unwrap()
	} else {
		errs = []error{err}
	}
	out := make([]error, 0, len(errs))
	for _, e := range errs {
		if fe, ok := errors.AsType[*FieldError](e); ok {
			out = append(out, &FieldError{Field: prefix + fe.Field, Err: fe.Err})
			continue
		}
		out = append(out, fmt.Errorf("%s%w", prefix, e))
	}
	return out
}

// checkPair validates a plugin and its grant, checks they belong together,
// and returns the plugin's parsed image.
func checkPair(p *Plugin, g *Grant) (ImageRef, error) {
	if p == nil || g == nil {
		return ImageRef{}, ErrMissing
	}
	if err := p.Validate(); err != nil {
		return ImageRef{}, fmt.Errorf("plugin %q: %w", p.Name, err)
	}
	if err := g.Validate(); err != nil {
		return ImageRef{}, fmt.Errorf("grant for %q: %w", g.Plugin, err)
	}
	if g.Plugin != p.Name {
		return ImageRef{}, fmt.Errorf("%w: grant names %q, plugin is %q", ErrGrantMismatch, g.Plugin, p.Name)
	}
	image, err := ParseImageRef(p.Image)
	if err != nil {
		return ImageRef{}, fmt.Errorf("plugin %q: %w", p.Name, err)
	}
	return image, nil
}

// NeedsApproval returns exactly the permissions plugin p requests that grant
// g does not approve, sorted. An empty result means p may be deployed under
// g; anything else means the upgrade waits until the owner approves the
// result (D-16). Secrets are compared by the names the instance mounts.
func NeedsApproval(g *Grant, p *Plugin) (Permissions, error) {
	if _, err := checkPair(p, g); err != nil {
		return Permissions{}, err
	}
	req := p.Requested()
	granted := g.Permissions
	return Permissions{
		Capabilities:    missing(req.Capabilities, granted.Capabilities),
		Egress:          missing(req.Egress, granted.Egress),
		Secrets:         missing(req.Secrets, granted.Secrets),
		DataTiers:       missing(req.DataTiers, granted.DataTiers),
		BrowserPaths:    missing(req.BrowserPaths, granted.BrowserPaths),
		ClusterServices: missing(req.ClusterServices, granted.ClusterServices),
	}, nil
}

// missing returns the elements of want that are not in have, in want's order,
// or nil when there are none.
func missing[T comparable](want, have []T) []T {
	var out []T
	for _, w := range want {
		if !slices.Contains(have, w) {
			out = append(out, w)
		}
	}
	return out
}
