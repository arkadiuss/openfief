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
