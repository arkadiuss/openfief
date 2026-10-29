// Package catalog is OpenFief's plugin catalog (D-43): plugins.json, a file
// in this repository released with the product and embedded in the operator.
// It lists the first-party and verified plugins; each entry pins one
// reviewed image digest and the identity of the publisher who signed it
// (D-44), so each new version needs a new entry. The release workflow adds
// each first-party plugin with the digest it signed and the release's
// version (D-88, scripts/release-catalog.sh); the author adds the others.
//
// The format is plugins.ParseCatalog's:
//
//	{"plugins": [
//	  {"name": "sample-connector", "version": "2026.11.0",
//	   "image": "ghcr.io/example/sample-connector@sha256:<64 hex digits>",
//	   "publisher": {"name": "example",
//	                 "issuer": "https://token.actions.githubusercontent.com",
//	                 "subject": "https://github.com/example/sample-connector/.github/workflows/release.yml@refs/heads/main"}}]}
//
// A publisher gives an issuer and exact subject (a keyless signature) or a
// PEM "publicKey". The catalog holds only what the operator's trust check
// needs; the CLI's install flow (D-46) may add fields later.
package catalog

import (
	_ "embed"
	"fmt"

	"github.com/arkadiuss/openfief/internal/plugins"
)

//go:embed plugins.json
var catalogJSON []byte

// Load parses the embedded catalog.
func Load() (plugins.Catalog, error) {
	c, err := plugins.ParseCatalog(catalogJSON)
	if err != nil {
		return plugins.Catalog{}, fmt.Errorf("embedded catalog: %w", err)
	}
	return c, nil
}
