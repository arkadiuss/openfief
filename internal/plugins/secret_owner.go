package plugins

import (
	"errors"
	"slices"
	"strings"
)

// OwnerInstanceLabel is the label the owner puts on each Secret a plugin
// instance mounts, naming that instance (D-182). The openfief CLI deletes a
// Secret on uninstall only when this label names the instance being
// uninstalled.
const OwnerInstanceLabel = "openfief.io/instance"

// reservedName is the name every object of the platform itself starts
// with: in the openfief namespace, the CloudNativePG cluster openfief-db and
// the Secrets it creates (openfief-db-app, openfief-db-ca,
// openfief-db-server, openfief-db-replication, openfief-db-superuser), the
// core, operator and egress proxy and their ServiceAccounts, and the
// ConfigMaps openfief-egress-allowlist, openfief-sa-jwks and
// openfief-trusted-signers (deploy/), and the demo's openfief-demo-tls.
const reservedName = "openfief"

// ErrReservedName reports an instance or Secret name the platform keeps for
// its own objects (D-182): "openfief" or one starting with "openfief-".
var ErrReservedName = errors.New("name reserved for OpenFief's own objects")

// ReservedName reports whether n is "openfief" or starts with "openfief-",
// the names the platform's own objects use. No instance may have such a
// name, so no instance's Secret, "<instance>--<role>" (D-89), can be one of
// the platform's, and no grant may name such a Secret (D-182).
func ReservedName(n string) bool {
	return n == reservedName || strings.HasPrefix(n, reservedName+"-")
}

// validateUnreserved returns ErrReservedName for field when n is reserved.
func validateUnreserved(field, n string) []error {
	if ReservedName(n) {
		return []error{fieldErr(field, ErrReservedName, n)}
	}
	return nil
}

// UninstallSecrets is what an uninstall does with the Secrets an instance
// named: Delete them; Keep them because another instance still names them;
// or leave alone those NotOwned, which exist but don't carry the instance's
// OwnerInstanceLabel (D-182). Each list is sorted and holds each name once; a
// named Secret that doesn't exist is in none of them.
type UninstallSecrets struct {
	Delete   []string
	Keep     []string
	NotOwned []string
}

// SecretsToDelete splits the Secrets instance named, in its manifest or its
// grant, into those the uninstall deletes (D-58), those it keeps because
// another installed instance still names them, and those it leaves alone
// because they aren't the instance's (D-182). owners maps each named Secret
// that exists to the value of its OwnerInstanceLabel, empty when it has none.
// Each instance mounts its own <instance>--<role> Secrets, and no two
// instance-role pairs spell one alike (D-89), so a Secret is kept only when
// another instance's grant names it.
func SecretsToDelete(instance string, named, others []string, owners map[string]string) UninstallSecrets {
	var out UninstallSecrets
	for _, s := range uniqueSorted(named) {
		owner, exists := owners[s]
		switch {
		case !exists:
		case owner != instance:
			out.NotOwned = append(out.NotOwned, s)
		case slices.Contains(others, s):
			out.Keep = append(out.Keep, s)
		default:
			out.Delete = append(out.Delete, s)
		}
	}
	return out
}
