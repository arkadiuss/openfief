package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// loaderBuild names this build of the loader: the SHA-256 of its own
// executable. CI builds a new demo-seed image only when the code it
// loads with changed, and the demo marker records the build that loaded
// the data, so a deploy of a new build reseeds the demo instead of
// migrating what an older one stored (D-161, demo.DecideLoad).
func loaderBuild() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding the loader's executable: %w", err)
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("reading the loader's executable: %w", err)
	}
	return buildDigest(data), nil
}

// buildDigest returns the hex SHA-256 of an executable's contents.
func buildDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
