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
	// ErrSeveralBrowserViews reports more than one view plugin instance
	// holding a browser path. At most one may (D-114), so the core serves
	// none of them rather than pick one.
	ErrSeveralBrowserViews = errors.New("more than one view plugin holds a browser path")
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
// view plugin instance holding a browser path (D-114), when its prefix
// covers path by whole segments. It fails with ErrNoBrowserRoute when no
// instance holds one or its prefix doesn't cover path, and with
// ErrSeveralBrowserViews when more than one instance holds one.
func RouteBrowserPath(routes []BrowserRoute, path string) (BrowserRoute, error) {
	switch len(routes) {
	case 0:
		return BrowserRoute{}, fmt.Errorf("%w: %s", ErrNoBrowserRoute, path)
	case 1:
	default:
		return BrowserRoute{}, fmt.Errorf("%w: %s and %s", ErrSeveralBrowserViews, routes[0].Instance, routes[1].Instance)
	}
	if !routes[0].matches(path) {
		return BrowserRoute{}, fmt.Errorf("%w: %s", ErrNoBrowserRoute, path)
	}
	return routes[0], nil
}
