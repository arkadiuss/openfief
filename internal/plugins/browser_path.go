package plugins

import (
	"path"
	"strings"
)

// maxBrowserPathLen bounds a browser path prefix.
const maxBrowserPathLen = 256

// coreOwnPaths returns the paths the core serves itself, which no view
// plugin may serve to browsers: the health endpoints and the read API's REST
// paths. Every Connect service's path, /openfief.<context>.v1.<Service>/, is
// the core's too (CoreServesPath).
func coreOwnPaths() []string { return []string{"/healthz", "/readyz", "/v1"} }

// coreServicePrefix starts the first path segment of every Connect service
// the core serves, such as /openfief.views.v1.ReadService/.
const coreServicePrefix = "openfief."

// CoreServesPath reports whether the core serves the URL path p itself, so
// that no view plugin's browser path may claim it (D-81): /healthz, /readyz,
// /v1 and everything under it, and every path whose first segment starts
// with "openfief.", the core's Connect services.
func CoreServesPath(p string) bool {
	for _, own := range coreOwnPaths() {
		if p == own || strings.HasPrefix(p, own+"/") {
			return true
		}
	}
	first, _, _ := strings.Cut(strings.TrimPrefix(p, "/"), "/")
	return strings.HasPrefix(first, coreServicePrefix)
}

// validBrowserPath accepts the path prefix a view plugin serves to browsers
// through the core (D-81): "/" or a clean absolute path of lower-case
// letters, digits, '-', '_' and '.' segments, without a trailing slash, at
// most 256 bytes, that the core does not serve itself. "/" means every path
// the core doesn't serve.
func validBrowserPath(p string) bool {
	if p == "/" {
		return true
	}
	if p == "" || len(p) > maxBrowserPathLen || !path.IsAbs(p) || path.Clean(p) != p || CoreServesPath(p) {
		return false
	}
	return !strings.ContainsFunc(p, func(r rune) bool {
		return !isLowerAlnum(r) && !strings.ContainsRune("/-_.", r)
	})
}

// validateBrowserPaths checks each browser path and reports duplicates.
func validateBrowserPaths(field string, ps []string) []error {
	var errs []error
	for i, p := range ps {
		if !validBrowserPath(p) {
			errs = append(errs, fieldErr(indexed(field, i), ErrInvalidBrowserPath, p))
		}
	}
	return append(errs, duplicates(field, ps, func(s string) string { return s })...)
}

// validateManifestBrowserPath checks a manifest's browser path: valid, and
// declared only by a plugin that may act as a view.
func validateManifestBrowserPath(m *Manifest) []error {
	if m.BrowserPath == "" {
		return nil
	}
	var errs []error
	if !validBrowserPath(m.BrowserPath) {
		errs = append(errs, fieldErr("browserPath", ErrInvalidBrowserPath, m.BrowserPath))
	}
	view := false
	for _, c := range m.Capabilities {
		view = view || c == CapabilityView
	}
	if !view {
		errs = append(errs, fieldErr("browserPath", ErrBrowserPathNeedsView, m.BrowserPath))
	}
	return errs
}
