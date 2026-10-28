package openfief

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// mountDir is the directory OpenFief mounts its files for a plugin in.
const mountDir = "/var/run/secrets/openfief"

// DefaultTokenPath is where OpenFief mounts a plugin's projected
// ServiceAccount token.
const DefaultTokenPath = mountDir + "/token"

// ErrEmptyToken reports a token file that holds no token.
var ErrEmptyToken = errors.New("plugin token file is empty")

// TokenSource returns the plugin's current token. Implementations must return
// the current token on every call, since tokens rotate.
type TokenSource interface {
	Token() (string, error)
}

// TokenFile is a TokenSource that reads the token from a file on every call.
// The zero value reads the file OPENFIEF_TOKEN_FILE names, or
// DefaultTokenPath when it is not set.
type TokenFile struct {
	// Path is the token file; empty means OPENFIEF_TOKEN_FILE, then
	// DefaultTokenPath.
	Path string
}

// Token reads the token file now, so a rotated token is used as soon as the
// kubelet writes it. Errors never contain the token.
func (f TokenFile) Token() (string, error) {
	path := f.Path
	if path == "" {
		path = os.Getenv(EnvIdentityFile)
	}
	if path == "" {
		path = DefaultTokenPath
	}
	b, err := readInDir(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("read plugin token: %w", err)
	}
	token := strings.TrimSpace(string(b))
	if token == "" {
		return "", fmt.Errorf("%w: %s", ErrEmptyToken, path)
	}
	return token, nil
}

// readInDir reads the file at path through an os.Root of its directory, so
// the read cannot leave that directory.
func readInDir(path string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("open token directory: %w", err)
	}
	b, err := root.ReadFile(filepath.Base(path))
	return b, errors.Join(err, root.Close())
}
