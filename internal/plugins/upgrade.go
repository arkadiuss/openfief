package plugins

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNoUpgradeVersion reports an upgrade to a catalog version that no
// catalog entry of the instance's plugin has.
var ErrNoUpgradeVersion = errors.New("no catalog entry of that version for this plugin")

// ResolveUpgradeSource reads the version argument of `openfief plugin
// upgrade` (D-90) for an instance that runs current: a digest-pinned image
// reference, read as ResolveInstallSource reads one, or a catalog version.
// A version is looked up among the catalog entries of the instance's plugin:
// those named like the entry that pins current's digest or, when none does,
// like the entries of current's repository. No such entry of that version
// fails with ErrNoUpgradeVersion; several fail with ErrAmbiguousCatalogName
// and list them.
func ResolveUpgradeSource(c *Catalog, current ImageRef, arg string) (InstallSource, error) {
	if strings.ContainsAny(arg, "/:@") {
		return ResolveInstallSource(c, arg)
	}
	names := catalogNamesOf(c, current)
	var found []CatalogEntry
	for i := range c.Entries {
		e := c.Entries[i]
		if e.Version == arg && names[e.Name] {
			found = append(found, e)
		}
	}
	switch len(found) {
	case 0:
		return InstallSource{}, fmt.Errorf("%w: %q for %s", ErrNoUpgradeVersion, arg, current)
	case 1:
		return InstallSource{Image: found[0].Image, Entry: &found[0]}, nil
	}
	return InstallSource{}, fmt.Errorf("%w: version %q: upgrade to one of %s by image", ErrAmbiguousCatalogName, arg, listEntries(found))
}

// catalogNamesOf returns the catalog names of the entries that pin image's
// digest or, when none does, of the entries of its repository.
func catalogNamesOf(c *Catalog, image ImageRef) map[string]bool {
	names := map[string]bool{}
	for i := range c.Entries {
		if c.Entries[i].Image.Digest() == image.Digest() {
			names[c.Entries[i].Name] = true
		}
	}
	if len(names) > 0 {
		return names
	}
	for i := range c.Entries {
		if c.Entries[i].Image.Repository() == image.Repository() {
			names[c.Entries[i].Name] = true
		}
	}
	return names
}

// listEntries names catalog entries for a message: each one's name,
// version, when it has one, and image.
func listEntries(es []CatalogEntry) string {
	out := make([]string, 0, len(es))
	for i := range es {
		s := es[i].Name
		if es[i].Version != "" {
			s += " " + es[i].Version
		}
		out = append(out, s+" ("+es[i].Image.String()+")")
	}
	return strings.Join(out, ", ")
}

// PlanUpgrade builds the Plugin and Grant that upgrading the installed
// current to a signed image applies (D-90): the same instance with the new
// image and its signed manifest label (D-42), the given config or, when
// config is nil, the current one, and the current in-cluster service while
// the new manifest still asks for one (D-74). The Grant approves exactly
// what the new version requests, which is what the owner confirmed.
func PlanUpgrade(current *Plugin, img *SignedImage, config []byte) (Plugin, Grant, error) {
	if config == nil {
		config = current.Config
	}
	var cs *ClusterService
	if img.Manifest.ClusterService != nil {
		cs = current.ClusterService
	}
	p, g, err := PlanInstall(current.Name, img, config, cs)
	if err != nil {
		return Plugin{}, Grant{}, fmt.Errorf("upgrade: %w", err)
	}
	return p, g, nil
}

// PermissionsDiff is how the permissions one plugin version requests differ
// from another's: what is Added and what is Removed, each sorted.
type PermissionsDiff struct {
	Added   Permissions
	Removed Permissions
}

// IsEmpty reports whether both versions request the same permissions.
func (d *PermissionsDiff) IsEmpty() bool { return d.Added.IsEmpty() && d.Removed.IsEmpty() }

// DiffPermissions compares what the version an upgrade replaces requests,
// from, with what the new version requests, to (D-90).
func DiffPermissions(from, to *Permissions) PermissionsDiff {
	a, b := from.sorted(), to.sorted()
	return PermissionsDiff{Added: subtract(&b, &a), Removed: subtract(&a, &b)}
}

// subtract returns what p holds and q does not.
func subtract(p, q *Permissions) Permissions {
	return Permissions{
		Capabilities:    missing(p.Capabilities, q.Capabilities),
		Egress:          missing(p.Egress, q.Egress),
		Secrets:         missing(p.Secrets, q.Secrets),
		DataTiers:       missing(p.DataTiers, q.DataTiers),
		BrowserPaths:    missing(p.BrowserPaths, q.BrowserPaths),
		ClusterServices: missing(p.ClusterServices, q.ClusterServices),
	}
}
