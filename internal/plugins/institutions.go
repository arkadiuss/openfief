package plugins

import (
	"cmp"
	"slices"
)

// InstitutionHolder is an installed instance and the institutions it books
// for (D-185).
type InstitutionHolder struct {
	Instance     string
	Institutions []string
}

// SharedInstitution is an institution that another installed instance
// already books for (D-185).
type SharedInstitution struct {
	Institution string
	Instance    string
}

// SharedInstitutions returns each of institutions that an instance among
// holders other than instance already books for, with that instance,
// sorted by institution, then instance. The install and an upgrade that
// changes the institutions warn about each (D-185); two instances may book
// for one institution, such as two accounts at the same broker.
func SharedInstitutions(instance string, institutions []string, holders []InstitutionHolder) []SharedInstitution {
	var out []SharedInstitution
	for _, h := range holders {
		if h.Instance == instance {
			continue
		}
		for _, in := range institutions {
			if slices.Contains(h.Institutions, in) {
				out = append(out, SharedInstitution{Institution: in, Instance: h.Instance})
			}
		}
	}
	slices.SortFunc(out, func(a, b SharedInstitution) int {
		return cmp.Or(cmp.Compare(a.Institution, b.Institution), cmp.Compare(a.Instance, b.Instance))
	})
	return out
}

// ApprovedInstitutions returns the institutions a PluginGrant approves.
// approved is the list the grant names, nil when it names no list at all,
// as every grant made before D-185 does: such a grant approves the
// institutions its plugin already declares, those of deployed, the version
// that runs under it, or none when nothing runs. Since a version that
// declares another institution runs only once a grant approves it, this
// never approves an institution the owner didn't. A grant that names an
// empty list approves none.
func ApprovedInstitutions(approved *[]string, deployed *Plugin) []string {
	if approved != nil {
		return slices.Clone(*approved)
	}
	if deployed == nil {
		return nil
	}
	return slices.Clone(deployed.Manifest.Institutions)
}
