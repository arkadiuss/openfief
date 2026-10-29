package plugins

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Errors of resolving and planning an install with the openfief CLI (D-46).
var (
	// ErrNotInCatalog reports an install argument that is neither an image
	// reference nor the name of a catalog entry.
	ErrNotInCatalog = errors.New("no catalog entry of that name")
	// ErrAmbiguousCatalogName reports a name that several catalog entries
	// share, one per reviewed version (D-44); the owner picks one by image.
	ErrAmbiguousCatalogName = errors.New("several catalog entries share that name")
)

// maxInstanceNameLen is the longest instance name the Plugin CRD accepts, so
// that "plugin-<name>" still fits Kubernetes' 63-character names.
const maxInstanceNameLen = 56

// InstallSource is what an install starts from: the image, and the catalog
// entry that pins its digest, when there is one (D-43, D-44).
type InstallSource struct {
	Image ImageRef
	// Entry is the catalog entry whose digest is Image's; nil for an image
	// the catalog does not pin, which installs as unverified.
	Entry *CatalogEntry
}

// ResolveInstallSource reads the argument of `openfief plugin install`: a
// digest-pinned image reference, or the name of a catalog entry. An argument
// with a '/', ':' or '@' is an image reference and must be pinned by digest
// (D-17); a tag is refused. A catalog name that several entries share is
// refused with ErrAmbiguousCatalogName, listing their versions and images
// instead of picking one (D-90).
func ResolveInstallSource(c *Catalog, arg string) (InstallSource, error) {
	if strings.ContainsAny(arg, "/:@") {
		img, err := ParseImageRef(arg)
		if err != nil {
			return InstallSource{}, err
		}
		src := InstallSource{Image: img}
		for i := range c.Entries {
			if c.Entries[i].Image.Digest() == img.Digest() {
				e := c.Entries[i]
				src.Entry = &e
				break
			}
		}
		return src, nil
	}
	var found []CatalogEntry
	for i := range c.Entries {
		if c.Entries[i].Name == arg {
			found = append(found, c.Entries[i])
		}
	}
	switch len(found) {
	case 0:
		return InstallSource{}, fmt.Errorf("%w: %q", ErrNotInCatalog, arg)
	case 1:
		return InstallSource{Image: found[0].Image, Entry: &found[0]}, nil
	}
	return InstallSource{}, fmt.Errorf("%w: %q: install one of %s by image", ErrAmbiguousCatalogName, arg, listEntries(found))
}

// DefaultInstanceName is the instance name an install uses when the owner
// gives none (D-48): the catalog entry's name, or else the last path
// component of the image's repository. It is empty when neither is a valid
// instance name, and the owner must name the instance.
func DefaultInstanceName(src *InstallSource) string {
	candidates := []string{}
	if src.Entry != nil {
		candidates = append(candidates, src.Entry.Name)
	}
	repo := src.Image.Repository()
	candidates = append(candidates, repo[strings.LastIndex(repo, "/")+1:])
	for _, n := range candidates {
		if validInstanceName(n) {
			return n
		}
	}
	return ""
}

// validInstanceName reports whether n is a plugin name the Plugin CRD
// accepts.
func validInstanceName(n string) bool {
	return n != "" && len(n) <= maxInstanceNameLen && validPluginName(n)
}

// PlanInstall builds the Plugin and the Grant that installing a signed image
// as the instance name applies (D-46, D-48). The Plugin requests exactly the
// image's signed manifest label (D-42) with config and, when the manifest
// asks for one, the in-cluster service cs the owner picked (D-74); the Grant
// approves exactly what it requests, which is what the owner confirmed: the
// hosts of the config fields the manifest names (D-75) and that one service
// included. Both are validated; every problem is returned, joined.
func PlanInstall(name string, img *SignedImage, config []byte, cs *ClusterService) (Plugin, Grant, error) {
	p := Plugin{Name: name, Image: img.Image.String(), Manifest: img.Manifest, Config: config, ClusterService: cs}
	g := Grant{Plugin: name, Permissions: p.Requested()}
	var errs []error
	if len(name) > maxInstanceNameLen {
		errs = append(errs, fieldErr("name", ErrInvalidName, name))
	}
	errs = append(errs, p.Validate(), g.Validate())
	if err := errors.Join(errs...); err != nil {
		return Plugin{}, Grant{}, fmt.Errorf("install %q: %w", name, err)
	}
	return p, g, nil
}

// UninstallSecrets is what an uninstall does with the Secrets an instance
// named: Delete them, or Keep them because another instance still names
// them. Both lists are sorted and hold each name once.
type UninstallSecrets struct {
	Delete []string
	Keep   []string
}

// SecretsToDelete splits the Secrets an uninstalled instance named, in its
// manifest or its grant, into those the uninstall deletes (D-58) and those it
// keeps because another installed instance still names them. Each instance
// mounts its own <instance>-<role> Secrets (D-89), so this keeps only a
// Secret another instance's grant names, or one two instance-role pairs
// spell alike, such as instance "a-b" with role "c" and "a" with "b-c".
func SecretsToDelete(instance, others []string) UninstallSecrets {
	var out UninstallSecrets
	for _, s := range uniqueSorted(instance) {
		if slices.Contains(others, s) {
			out.Keep = append(out.Keep, s)
			continue
		}
		out.Delete = append(out.Delete, s)
	}
	return out
}

// uniqueSorted returns xs sorted, each value once.
func uniqueSorted(xs []string) []string {
	return slices.Compact(sortedCopy(xs, cmp.Compare[string]))
}
