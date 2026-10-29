package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/views"
)

// ErrViewsUnknown reports that the view plugins that serve browsers can't
// be read yet, such as before the core has read its cluster.
var ErrViewsUnknown = errors.New("the view plugins that serve browsers are not known")

// BrowserViews reads which view plugin instances serve browsers, and where
// the core calls each one. internal/views/kube implements it from the
// operator's Plugin resources and the Services it creates.
type BrowserViews interface {
	// BrowserRoutes returns the routes of the view plugins that run now
	// with a granted browser path (D-81), in any order. It fails with an
	// error wrapping ErrViewsUnknown while it can't answer.
	BrowserRoutes(ctx context.Context) ([]views.BrowserRoute, error)
	// BaseURL returns the URL the core calls instance at.
	BaseURL(instance string) (string, error)
}

// PluginTokens issues the core's token for a plugin instance (D-59).
// internal/identity/kube's CoreTokens implements it.
type PluginTokens interface {
	TokenFor(ctx context.Context, instance string) (string, error)
}

// BrowserTarget is where the core sends one browser request: the view
// plugin instance, its base URL, and the core's token for it.
type BrowserTarget struct {
	Instance string
	BaseURL  string
	Token    string
}

// FrontDoor decides where a browser's request goes (D-81): the core is the
// only entry point for browsers, and it calls the view plugin that serves
// the path with its own token for that instance, as it calls any plugin
// (D-59).
type FrontDoor struct {
	Views  BrowserViews
	Tokens PluginTokens
}

// Target returns where the browser's request for path goes. It fails with
// an error wrapping views.ErrNoBrowserRoute or
// views.ErrAmbiguousBrowserRoute when no single view plugin serves path,
// ErrViewsUnknown while the routes can't be read, and otherwise when the
// instance has no address or no token can be issued for it.
func (f *FrontDoor) Target(ctx context.Context, path string) (BrowserTarget, error) {
	routes, err := f.Views.BrowserRoutes(ctx)
	if err != nil {
		return BrowserTarget{}, fmt.Errorf("browser routes: %w", err)
	}
	route, err := views.RouteBrowserPath(routes, path)
	if err != nil {
		return BrowserTarget{}, fmt.Errorf("routing: %w", err)
	}
	base, err := f.Views.BaseURL(route.Instance)
	if err != nil {
		return BrowserTarget{}, fmt.Errorf("view plugin %s: %w", route.Instance, err)
	}
	token, err := f.Tokens.TokenFor(ctx, route.Instance)
	if err != nil {
		return BrowserTarget{}, fmt.Errorf("token for view plugin %s: %w", route.Instance, err)
	}
	return BrowserTarget{Instance: route.Instance, BaseURL: base, Token: token}, nil
}
