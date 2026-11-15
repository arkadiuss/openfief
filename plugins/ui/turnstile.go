package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// The instance's optional Secret ui--turnstile-secret (role turnstile-secret) is
// mounted by the manifest at turnstileMountDir, with the widget's secret key in
// the file turnstileKeyFile. It is created out of band (task demo:turnstile-secret)
// and is never in config, git or a log.
const (
	turnstileMountDir = "/etc/openfief/turnstile"
	turnstileKeyFile  = "secret-key"
)

// errNoTurnstileKey reports a Turnstile key that is not mounted, can't be read or
// is empty. It names the directory and never the key.
var errNoTurnstileKey = errors.New("no Turnstile secret key in the mounted Secret")

// keyFromMount returns the function that reads the Turnstile secret key from the
// Secret mounted at dir on every call, through an os.Root so the read can't leave
// dir, so a rotated Secret works at once. A missing or empty key is an error.
func keyFromMount(dir string) func() (string, error) {
	return func() (string, error) {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return "", fmt.Errorf("%w: %s", errNoTurnstileKey, dir)
		}
		raw, readErr := root.ReadFile(turnstileKeyFile)
		closeErr := root.Close()
		key := strings.TrimSpace(string(raw))
		if errors.Is(readErr, fs.ErrNotExist) || readErr != nil || closeErr != nil || key == "" {
			return "", fmt.Errorf("%w: %s", errNoTurnstileKey, dir)
		}
		return key, nil
	}
}

// turnstile returns the Turnstile config the plugin's config asks for, or nil when
// it asks for none: the site key, the siteverify URL, the secret key read from the
// instance's Secret and the egress client, which reaches only the host the grant
// approved.
func (r *rawConfig) turnstile() (*ui.TurnstileConfig, error) {
	on, err := r.turnstileOn()
	if err != nil || !on {
		return nil, err
	}
	ep, err := openfief.LoadEndpoint(r.TurnstileVerifyURL, "")
	if err != nil {
		return nil, fmt.Errorf("config turnstileVerifyUrl: %w", err)
	}
	hc, err := ep.NewClient()
	if err != nil {
		return nil, fmt.Errorf("egress: %w", err)
	}
	return &ui.TurnstileConfig{SiteKey: r.TurnstileSiteKey, VerifyURL: ep.BaseURL, Secret: keyFromMount(turnstileMountDir), HTTP: hc}, nil
}
