package plugins

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errRepoLength    = errors.New("repository must be 1 to 255 characters")
	errRepoTag       = errors.New("tagged image: name it by digest only")
	errRepoComponent = errors.New("invalid repository path component")
	errRegistryHost  = errors.New("invalid registry host")
)

const (
	digestAlgorithm = "sha256:"
	digestHexLen    = 64
	maxRepoLen      = 255
)

// ImageRef is a container image pinned by digest: repository@sha256:<64 hex>
// (D-17). Its zero value is no image; ParseImageRef is the only way to make one.
// Signature verification with cosign is an adapter's job; a verified image is
// still named by an ImageRef.
type ImageRef struct {
	repository string
	digest     string
}

// ParseImageRef accepts "repository@sha256:<64 lower-case hex digits>", where
// the repository may start with a registry host (with an optional port).
// A tag, with or without a digest, is refused: a tag can move, and an image
// named by both is ambiguous about which one counts.
func ParseImageRef(s string) (ImageRef, error) {
	repo, digest, pinned := strings.Cut(s, "@")
	if !pinned {
		return ImageRef{}, fmt.Errorf("%w: %q", ErrImageNotPinned, s)
	}
	if !validDigest(digest) {
		return ImageRef{}, fmt.Errorf("%w: %q: digest must be sha256 with 64 lower-case hex digits", ErrInvalidImageRef, s)
	}
	if err := validateRepository(repo); err != nil {
		return ImageRef{}, fmt.Errorf("%w: %q: %w", ErrInvalidImageRef, s, err)
	}
	return ImageRef{repository: repo, digest: digest}, nil
}

// Repository returns the image's repository, including any registry host.
func (r ImageRef) Repository() string { return r.repository }

// Digest returns "sha256:<hex>".
func (r ImageRef) Digest() string { return r.digest }

// String returns "repository@sha256:<hex>", the form a Pod spec takes.
func (r ImageRef) String() string {
	if r.IsZero() {
		return ""
	}
	return r.repository + "@" + r.digest
}

// IsZero reports whether r is the zero value, no image.
func (r ImageRef) IsZero() bool { return r.repository == "" }

// Equal reports whether r and o name the same image.
func (r ImageRef) Equal(o ImageRef) bool { return r == o }

func validDigest(d string) bool {
	hex, ok := strings.CutPrefix(d, digestAlgorithm)
	if !ok || len(hex) != digestHexLen {
		return false
	}
	for _, c := range hex {
		if !isDigit(c) && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// validateRepository follows the OCI distribution name grammar: an optional
// registry host[:port], then lower-case path components separated by '/'.
func validateRepository(repo string) error {
	if repo == "" || len(repo) > maxRepoLen {
		return errRepoLength
	}
	components := strings.Split(repo, "/")
	if len(components) > 1 && isRegistryHost(components[0]) {
		if err := validateRegistryHost(components[0]); err != nil {
			return err
		}
		components = components[1:]
	}
	for _, c := range components {
		if strings.Contains(c, ":") {
			return errRepoTag
		}
		if !validPathComponent(c) {
			return fmt.Errorf("%w: %q", errRepoComponent, c)
		}
	}
	return nil
}

// isRegistryHost applies the distribution rule: the first component is a
// registry host when it has a '.' or a ':' or is "localhost".
func isRegistryHost(c string) bool {
	return strings.ContainsAny(c, ".:") || c == "localhost"
}

func validateRegistryHost(c string) error {
	host, port, hasPort := strings.Cut(c, ":")
	if hasPort && !isDigits(port) {
		return fmt.Errorf("%w: %q", errRegistryHost, c)
	}
	for label := range strings.SplitSeq(host, ".") {
		if validateLabel(label) != nil {
			return fmt.Errorf("%w: %q", errRegistryHost, c)
		}
	}
	return nil
}

// validPathComponent matches [a-z0-9]+(?:(?:[._]|__|[-]+)[a-z0-9]+)*.
func validPathComponent(c string) bool {
	if c == "" || !isLowerAlnum(rune(c[0])) || !isLowerAlnum(rune(c[len(c)-1])) {
		return false
	}
	run := ""
	for _, r := range c {
		if isLowerAlnum(r) {
			if run != "" && !validSeparator(run) {
				return false
			}
			run = ""
			continue
		}
		run += string(r)
	}
	return true
}

func validSeparator(s string) bool {
	return s == "." || s == "_" || s == "__" || strings.Trim(s, "-") == ""
}
