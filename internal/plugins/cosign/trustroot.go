package cosign

import (
	_ "embed"
	"fmt"

	"github.com/sigstore/sigstore-go/pkg/root"
)

// trustedRootJSON is the Sigstore public-good trusted root, fetched through
// TUF by `task trustroot:update` (D-56).
//
//go:embed trusted_root.json
var trustedRootJSON []byte

// TrustedRoot returns the bundled Sigstore public-good trust root: Fulcio's
// CAs, Rekor's logs, the CT logs and the timestamp authority.
func TrustedRoot() (*root.TrustedRoot, error) {
	tr, err := root.NewTrustedRootFromJSON(trustedRootJSON)
	if err != nil {
		return nil, fmt.Errorf("bundled trusted root: %w", err)
	}
	return tr, nil
}
