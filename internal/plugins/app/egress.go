// Package app holds the plugins context's use cases and the ports they
// consume. The egress use case decides, for the egress proxy, which plugin
// is asking and whether it may reach a destination (D-27, D-32).
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// TokenReviewer verifies a projected ServiceAccount token and returns the
// Kubernetes username it authenticates. Implementations must accept only
// tokens issued for plugins.CoreAudience, and return an error wrapping
// ErrUnauthenticated for any token they do not accept.
type TokenReviewer interface {
	ReviewToken(ctx context.Context, token string) (username string, err error)
}

// AllowListSource returns a plugin's current egress allow-list. The operator
// may change the lists at any time; each call sees the latest one, and a
// plugin without a list (false) may reach nothing.
type AllowListSource interface {
	AllowList(plugin string) (plugins.EgressAllowList, bool)
}

// Identification errors. The egress proxy refuses the request on any of them.
var (
	// ErrNoToken reports a request that carries no plugin token.
	ErrNoToken = errors.New("no plugin token")
	// ErrUnauthenticated reports a token the TokenReviewer does not accept.
	ErrUnauthenticated = errors.New("token not authenticated")
	// ErrNotAPlugin reports an authenticated identity that is not plugin identity.
	ErrNotAPlugin = errors.New("identity is not a plugin")
)

// Reasons an EgressDecision gives.
const (
	ReasonListed      = "destination is on the plugin's egress allow-list"
	ReasonNotListed   = "destination is not on the plugin's egress allow-list"
	ReasonNoAllowList = "plugin has no egress allow-list"
)

// EgressDecision is the egress proxy's answer to one plugin's CONNECT
// request: the plugin, the destination, whether it is allowed and why.
type EgressDecision struct {
	Plugin  string
	Host    string
	Port    uint16
	Allowed bool
	Reason  string
}

// Egress is the egress proxy's use case: it identifies the plugin behind a
// projected token and checks destinations against that plugin's egress
// allow-list. It does not resolve or check addresses; the proxy adapter does
// that for every address it dials (egress.CheckAddress).
type Egress struct {
	tokens TokenReviewer
	lists  AllowListSource
}

// NewEgress returns the egress use case over its two ports.
func NewEgress(tokens TokenReviewer, lists AllowListSource) *Egress {
	return &Egress{tokens: tokens, lists: lists}
}

// Identify returns the name of the plugin whose projected token this is
// (plugin identity, D-27). The error never contains the token.
func (e *Egress) Identify(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrNoToken
	}
	username, err := e.tokens.ReviewToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("review plugin token: %w", err)
	}
	plugin, ok := plugins.PluginForServiceAccount(username)
	if !ok {
		return "", fmt.Errorf("%w: %q", ErrNotAPlugin, username)
	}
	return plugin, nil
}

// Decide checks one destination against the plugin's current egress
// allow-list. A plugin without a list may reach nothing.
func (e *Egress) Decide(plugin, host string, port uint16) EgressDecision {
	d := EgressDecision{Plugin: plugin, Host: host, Port: port}
	list, ok := e.lists.AllowList(plugin)
	switch {
	case !ok:
		d.Reason = ReasonNoAllowList
	case list.Allows(host, port):
		d.Allowed, d.Reason = true, ReasonListed
	default:
		d.Reason = ReasonNotListed
	}
	return d
}
