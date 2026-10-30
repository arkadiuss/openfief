package plugins

import (
	"encoding/json"
	"errors"
	"path"
	"slices"
	"strings"
)

// Capability is what a plugin may act as (D-11).
type Capability string

// The M0 capabilities.
const (
	// CapabilityConnector brings data in through an ingest path.
	CapabilityConnector Capability = "connector"
	// CapabilityView presents the canonical record elsewhere, read-only.
	CapabilityView Capability = "view"
)

func (c Capability) known() bool {
	switch c {
	case CapabilityConnector, CapabilityView:
		return true
	}
	return false
}

// DataTier names a class of data a plugin asks to see. Tiers are declared and
// granted like any other permission; what each tier contains is not decided
// for M0 (proposal P-33), so a tier is only a name here.
type DataTier string

const maxDataTierLen = 32

func (d DataTier) valid() bool {
	if d == "" || len(d) > maxDataTierLen {
		return false
	}
	for _, r := range d {
		if !isLowerAlnum(r) && ('A' > r || r > 'Z') && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

// SecretMount is a Kubernetes Secret, in the openfief namespace, that a plugin
// asks to have mounted read-only at MountPath. The manifest names it by its
// Role, such as "flex-token"; each instance mounts its own Secret,
// SecretName(instance, role), so two instances of one image never share one
// (D-89).
type SecretMount struct {
	Role      string
	MountPath string
}

// SecretName is the name of the Secret an instance mounts for a role:
// "<instance>-<role>" (D-89).
func SecretName(instance, role string) string { return instance + "-" + role }

// ReservedMountPath is where Kubernetes and OpenFief put service account
// tokens; no Secret may be mounted at, above or below it.
const ReservedMountPath = "/var/run/secrets"

// Manifest is what a plugin declares: capabilities, egress hosts, secrets,
// data tiers, config schema (D-11) and, for a view plugin that serves
// browsers, the path prefix the core proxies to it (D-81). The config schema
// is a JSON Schema document kept as bytes; checking a config against it is
// an adapter's job.
type Manifest struct {
	Capabilities []Capability
	Egress       []EgressHost
	Secrets      []SecretMount
	DataTiers    []DataTier
	ConfigSchema []byte
	// BrowserPath is the path prefix the plugin serves to browsers through
	// the core, such as "/"; empty when it serves no browser.
	BrowserPath string
	// EgressFromConfig names top-level config fields whose value, an https
	// URL, gives a host the plugin reaches through the egress proxy instead
	// of a fixed one (D-75).
	EgressFromConfig []string
	// ClusterService, when not nil, asks for one in-cluster service that the
	// owner picks at install and the plugin reaches directly (D-74).
	ClusterService *ClusterServiceRequest
	// CrossCheckPrices declares that a connector sends cross-check prices:
	// the core stores them only from a connector that declares it (D-112).
	CrossCheckPrices bool
	// StatementFiles declares that a connector takes statement files the
	// owner uploads, so the UI offers the upload form on the accounts bound
	// to it (D-109).
	StatementFiles bool
}

// Validate returns every problem in m, each a *FieldError, joined; or nil.
func (m *Manifest) Validate() error {
	var errs []error
	if len(m.Capabilities) == 0 {
		errs = append(errs, &FieldError{Field: "capabilities", Err: ErrNoCapability})
	}
	errs = append(errs, validateCapabilities(m.Capabilities)...)
	errs = append(errs, validateEgress(m.Egress)...)
	errs = append(errs, validateSecretMounts(m.Secrets)...)
	errs = append(errs, validateDataTiers(m.DataTiers)...)
	errs = append(errs, validateManifestBrowserPath(m)...)
	errs = append(errs, validateConfigFields(m.EgressFromConfig)...)
	errs = append(errs, m.ClusterService.validate()...)
	if len(m.ConfigSchema) > 0 && !json.Valid(m.ConfigSchema) {
		errs = append(errs, &FieldError{Field: "configSchema", Err: ErrInvalidJSON})
	}
	return errors.Join(errs...)
}

func validateCapabilities(cs []Capability) []error {
	var errs []error
	for i, c := range cs {
		if !c.known() {
			errs = append(errs, fieldErr(indexed("capabilities", i), ErrUnknownCapability, string(c)))
		}
	}
	return append(errs, duplicates("capabilities", cs, func(c Capability) string { return string(c) })...)
}

func validateEgress(hs []EgressHost) []error {
	var errs []error
	for i, h := range hs {
		if err := h.validate(); err != nil {
			errs = append(errs, &FieldError{Field: indexed("egress", i), Err: err})
		}
	}
	return append(errs, duplicates("egress", hs, EgressHost.String)...)
}

func validateDataTiers(ds []DataTier) []error {
	var errs []error
	for i, d := range ds {
		if !d.valid() {
			errs = append(errs, fieldErr(indexed("dataTiers", i), ErrInvalidDataTier, string(d)))
		}
	}
	return append(errs, duplicates("dataTiers", ds, func(d DataTier) string { return string(d) })...)
}

func validateSecretNames(field string, names []string) []error {
	var errs []error
	for i, n := range names {
		if !validSecretName(n) {
			errs = append(errs, fieldErr(indexed(field, i), ErrInvalidName, n))
		}
	}
	return append(errs, duplicates(field, names, func(s string) string { return s })...)
}

func validateSecretMounts(ms []SecretMount) []error {
	roles := make([]string, 0, len(ms))
	paths := make([]string, 0, len(ms))
	for _, m := range ms {
		roles = append(roles, m.Role)
		paths = append(paths, m.MountPath)
	}
	var errs []error
	for i, r := range roles {
		if validateLabel(r) != nil {
			errs = append(errs, fieldErr(indexed("secrets", i)+".role", ErrInvalidName, r))
		}
	}
	errs = append(errs, duplicates("secrets", roles, func(s string) string { return s })...)
	for i, p := range paths {
		if !validMountPath(p) {
			errs = append(errs, fieldErr(indexed("secrets", i)+".mountPath", ErrInvalidMountPath, p))
		}
	}
	return append(errs, overlappingPaths("secrets", paths)...)
}

// validSecretName checks the Kubernetes Secret name rule (a DNS-1123
// subdomain): at most 253 lower-case letters, digits, '-' and '.', starting
// and ending with a letter or digit.
func validSecretName(n string) bool {
	if n == "" || len(n) > maxHostnameLen {
		return false
	}
	for label := range strings.SplitSeq(n, ".") {
		if label == "" || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if !isLowerAlnum(r) && r != '-' {
				return false
			}
		}
	}
	return true
}

// validMountPath accepts a clean absolute path that is not "/" and does not
// touch ReservedMountPath.
func validMountPath(p string) bool {
	return path.IsAbs(p) && path.Clean(p) == p && p != "/" && !strings.Contains(p, ":") &&
		!pathsOverlap(p, ReservedMountPath)
}

// pathsOverlap reports whether a and b are equal or one contains the other.
func pathsOverlap(a, b string) bool {
	return a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}

func overlappingPaths(field string, paths []string) []error {
	var errs []error
	for i, p := range paths {
		for _, q := range paths[:i] {
			if pathsOverlap(p, q) {
				errs = append(errs, fieldErr(indexed(field, i)+".mountPath", ErrOverlappingMountPaths, p))
				break
			}
		}
	}
	return errs
}

// duplicates reports each element of xs whose key appeared earlier.
func duplicates[T any](field string, xs []T, key func(T) string) []error {
	var errs []error
	seen := make(map[string]bool, len(xs))
	for i, x := range xs {
		k := key(x)
		if seen[k] {
			errs = append(errs, fieldErr(indexed(field, i), ErrDuplicate, k))
		}
		seen[k] = true
	}
	return errs
}

// sortedCopy returns a sorted copy of xs, or nil when xs is empty.
func sortedCopy[T any](xs []T, compare func(a, b T) int) []T {
	if len(xs) == 0 {
		return nil
	}
	out := slices.Clone(xs)
	slices.SortFunc(out, compare)
	return out
}
