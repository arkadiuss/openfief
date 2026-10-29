package views

import (
	"errors"
	"fmt"
	"strings"
)

// Browser routing errors.
var (
	// ErrNoBrowserRoute reports a path no view plugin serves to browsers.
	ErrNoBrowserRoute = errors.New("no view plugin serves this path")
	// ErrAmbiguousBrowserRoute reports a path whose longest matching
	// prefix more than one view plugin serves; the core serves neither
	// rather than pick one.
	ErrAmbiguousBrowserRoute = errors.New("more than one view plugin serves this path")
)

// BrowserRoute is a view plugin instance and the path prefix it serves to
// browsers through the core (D-81): what its manifest declares and its
// grant approves. Prefix is "/" or a path without a trailing slash.
type BrowserRoute struct {
	Instance string
	Prefix   string
}

// matches reports whether the route's prefix covers path: "/" covers every
// path, any other prefix itself and the paths below it.
func (r BrowserRoute) matches(path string) bool {
	return r.Prefix == "/" || path == r.Prefix || strings.HasPrefix(path, r.Prefix+"/")
}

// RouteBrowserPath returns the route that serves path to a browser: the one
// with the longest prefix that covers it. It fails with ErrNoBrowserRoute
// when no route covers path, and with ErrAmbiguousBrowserRoute when two
// instances declare that same longest prefix.
func RouteBrowserPath(routes []BrowserRoute, path string) (BrowserRoute, error) {
	best := make([]BrowserRoute, 0, len(routes))
	for _, r := range routes {
		if !r.matches(path) {
			continue
		}
		switch {
		case len(best) == 0 || len(r.Prefix) > len(best[0].Prefix):
			best = append(best[:0], r)
		case len(r.Prefix) == len(best[0].Prefix):
			best = append(best, r)
		}
	}
	switch len(best) {
	case 0:
		return BrowserRoute{}, fmt.Errorf("%w: %s", ErrNoBrowserRoute, path)
	case 1:
		return best[0], nil
	}
	return BrowserRoute{}, fmt.Errorf("%w: %s and %s both serve %s", ErrAmbiguousBrowserRoute, best[0].Instance, best[1].Instance, best[0].Prefix)
}
