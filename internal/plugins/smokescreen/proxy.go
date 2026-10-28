// Package smokescreen is the egress proxy adapter: Stripe's Smokescreen,
// embedded as a library (D-32). A plugin opens an HTTP CONNECT tunnel and
// proves its plugin identity with its projected ServiceAccount token in
// "Proxy-Authorization: Bearer <token>" (D-27); the proxy checks the
// destination against that plugin's egress allow-list, resolves the hostname
// itself and refuses it if any resolved address is private or internal.
package smokescreen

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	ss "github.com/stripe/smokescreen/pkg/smokescreen"
	"github.com/stripe/smokescreen/pkg/smokescreen/acl/v1"

	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/egress"
)

// Errors behind a refused request; Smokescreen logs them, never the token.
var (
	ErrNoListener          = errors.New("egress proxy needs a listener")
	ErrProxyEnvironment    = errors.New("egress proxy must dial destinations itself: unset HTTPS_PROXY and https_proxy")
	ErrBadProxyAuth        = errors.New(`Proxy-Authorization is not "Bearer <token>"`)
	ErrNotConnect          = errors.New("only HTTP CONNECT is proxied")
	ErrUpstreamProxy       = errors.New("chaining to an upstream proxy is not allowed")
	ErrBadConnectAuthority = errors.New("CONNECT target is not host:port")
)

// Options configure the proxy. Only Listener is required.
type Options struct {
	// Listener is where the proxy accepts plugins' connections.
	Listener net.Listener
	// Resolver looks up destination hostnames; nil means the system resolver.
	Resolver ss.Resolver
	// Dial connects to a checked destination address; nil means a plain
	// net.Dialer. Tests use it to reach a local server.
	Dial func(ctx context.Context, network, address string) (net.Conn, error)
	// Healthcheck, when set, answers GET /healthcheck.
	Healthcheck http.Handler
	// Log receives Smokescreen's request log; nil means its default logger.
	Log *logrus.Logger
}

// NewConfig returns a Smokescreen configuration that serves on
// opts.Listener, identifies each plugin through egress.Identify and allows
// only destinations egress.Decide allows. Run it with
// smokescreen.StartWithConfig.
func NewConfig(e *app.Egress, opts Options) (*ss.Config, error) {
	if opts.Listener == nil {
		return nil, ErrNoListener
	}
	// Smokescreen's goproxy chains every CONNECT through HTTPS_PROXY when it
	// is set, which would bypass the resolver check below.
	if os.Getenv("HTTPS_PROXY") != "" || os.Getenv("https_proxy") != "" {
		return nil, ErrProxyEnvironment
	}
	cfg := ss.NewConfig()
	cfg.Listener = opts.Listener
	if ap, err := netip.ParseAddrPort(opts.Listener.Addr().String()); err == nil {
		cfg.Port = ap.Port() // lets Smokescreen refuse connections to itself
	}
	cfg.RoleFromRequest = func(req *http.Request) (string, error) { return identify(e, req) }
	cfg.EgressACL = decider{egress: e}
	if err := cfg.SetDenyRanges(forbiddenRanges()); err != nil {
		return nil, fmt.Errorf("egress proxy deny ranges: %w", err)
	}
	base := opts.Resolver
	if base == nil {
		base = cfg.Resolver
	}
	cfg.Resolver = checkingResolver{base: base}
	dial := opts.Dial
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}
	cfg.ProxyDialTimeout = func(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return dial(ctx, network, address)
	}
	cfg.Healthcheck = opts.Healthcheck
	if opts.Log != nil {
		cfg.Log = opts.Log
	}
	return cfg, nil
}

// identify is the role function: the plugin named by the request's token.
func identify(e *app.Egress, req *http.Request) (string, error) {
	token, err := bearerToken(req.Header.Get("Proxy-Authorization"))
	if err != nil {
		return "", err
	}
	plugin, err := e.Identify(req.Context(), token)
	if err != nil {
		return "", fmt.Errorf("identify plugin: %w", err)
	}
	return plugin, nil
}

// bearerToken reads "Bearer <token>"; the scheme is case-insensitive
// (RFC 9110). An absent header means no token.
func bearerToken(header string) (string, error) {
	if header == "" {
		return "", app.ErrNoToken
	}
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" || strings.ContainsAny(token, " \t") {
		return "", ErrBadProxyAuth
	}
	return token, nil
}

// decider is Smokescreen's EgressACL: the plugin's egress allow-list, for
// CONNECT tunnels only, with no upstream proxy.
type decider struct{ egress *app.Egress }

// Decide implements acl.Decider. args.Service is the plugin name from the role function.
func (d decider) Decide(args acl.DecideArgs) (acl.Decision, error) {
	deny := func(err error) (acl.Decision, error) {
		return acl.Decision{Result: acl.Deny, Reason: err.Error()}, nil
	}
	switch {
	case args.Req == nil || args.Req.Method != http.MethodConnect:
		return deny(ErrNotConnect)
	case args.ConnectProxyHost != "":
		return deny(ErrUpstreamProxy)
	}
	port, err := connectPort(args.Req.Host)
	if err != nil {
		return deny(err)
	}
	decision := d.egress.Decide(args.Service, args.Host, port)
	result := acl.Deny
	if decision.Allowed {
		result = acl.Allow
	}
	return acl.Decision{Result: result, Reason: decision.Reason}, nil
}

// connectPort reads the port of a CONNECT request's "host:port" target.
func connectPort(authority string) (uint16, error) {
	_, p, err := net.SplitHostPort(authority)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrBadConnectAuthority, err)
	}
	port, err := strconv.ParseUint(p, 10, 16)
	if err != nil || port == 0 {
		return 0, ErrBadConnectAuthority
	}
	return uint16(port), nil
}

// forbiddenRanges is egress.ForbiddenPrefixes as Smokescreen deny ranges.
func forbiddenRanges() []string {
	prefixes := egress.ForbiddenPrefixes()
	out := make([]string, len(prefixes))
	for i, p := range prefixes {
		out[i] = p.String()
	}
	return out
}

// checkingResolver refuses a hostname when any address it resolves to fails
// egress.CheckAddress, so a name that mixes public and private addresses is
// refused outright rather than dialed at its public one.
type checkingResolver struct{ base ss.Resolver }

// LookupPort implements smokescreen.Resolver.
func (r checkingResolver) LookupPort(ctx context.Context, network, service string) (int, error) {
	port, err := r.base.LookupPort(ctx, network, service)
	if err != nil {
		return 0, fmt.Errorf("look up port %q: %w", service, err)
	}
	return port, nil
}

// LookupIP implements smokescreen.Resolver.
func (r checkingResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	ips, err := r.base.LookupIP(ctx, network, host)
	if err != nil {
		return nil, fmt.Errorf("resolve %q: %w", host, err)
	}
	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			return nil, &forbiddenError{host: host, err: egress.ErrForbiddenAddress}
		}
		if err := egress.CheckAddress(addr.Unmap()); err != nil {
			return nil, &forbiddenError{host: host, err: err}
		}
	}
	return ips, nil
}

// forbiddenError is a resolved address the proxy must not dial. It is a
// net.Error, so Smokescreen answers 502 Bad Gateway rather than 500.
type forbiddenError struct {
	host string
	err  error
}

func (e *forbiddenError) Error() string   { return fmt.Sprintf("%q resolves to a %v", e.host, e.err) }
func (e *forbiddenError) Unwrap() error   { return e.err }
func (e *forbiddenError) Timeout() bool   { return false }
func (e *forbiddenError) Temporary() bool { return false }
