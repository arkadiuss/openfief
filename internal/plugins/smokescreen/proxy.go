// Package smokescreen is the egress proxy adapter: Stripe's Smokescreen,
// embedded as a library through BuildProxy inside OpenFief's own HTTP server
// (D-32, D-63). A plugin opens an HTTP CONNECT tunnel and proves its plugin
// identity with its projected ServiceAccount token, sent as the password of
// the proxy URL, which HTTP clients send as a Basic Proxy-Authorization
// header (D-63). The proxy checks the destination against that plugin's
// egress allow-list, resolves the hostname itself and refuses it if any
// resolved address is private or internal.
package smokescreen

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	ss "github.com/stripe/smokescreen/pkg/smokescreen"
	"github.com/stripe/smokescreen/pkg/smokescreen/acl/v1"
	"github.com/stripe/smokescreen/pkg/smokescreen/conntrack"

	"github.com/arkadiuss/openfief/internal/plugins/app"
	"github.com/arkadiuss/openfief/internal/plugins/egress"
)

const (
	// DefaultDrainTimeout is how long Serve lets open tunnels finish after
	// its context ends, inside Kubernetes' default 30 s termination grace.
	DefaultDrainTimeout = 25 * time.Second
	// tunnelIdleTimeout closes a tunnel with no traffic in either direction.
	// Smokescreen bumps the connection's deadlines on every read and write
	// only when an idle timeout is set; without one, the HTTP server's write
	// timeout would cut every tunnel after a fixed time.
	tunnelIdleTimeout = 5 * time.Minute
)

// Errors behind a refused request or a proxy that will not start;
// Smokescreen logs them, never the token.
var (
	ErrNoListener          = errors.New("egress proxy needs a listener")
	ErrNoEgress            = errors.New("egress proxy needs the egress use case and its allow-lists")
	ErrProxyEnvironment    = errors.New("egress proxy must dial destinations itself: unset HTTPS_PROXY and https_proxy")
	ErrBadProxyAuth        = errors.New(`Proxy-Authorization is not "Basic" with the token as the password`)
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
	// DrainTimeout bounds how long Serve waits for open tunnels on shutdown;
	// zero means DefaultDrainTimeout.
	DrainTimeout time.Duration
}

// Proxy is the egress proxy: Smokescreen's handler from BuildProxy, served
// by an http.Server this package owns. Unlike Smokescreen's
// StartWithConfig, it never exits the process and installs no signal
// handlers; Serve stops when its context ends (D-63).
type Proxy struct {
	server   *http.Server
	listener net.Listener
	tracker  *conntrack.Tracker
	drain    time.Duration
	log      *logrus.Logger
}

// New returns a proxy that serves on opts.Listener, identifies each plugin
// through e.Identify and allows only destinations e.Decide allows. It always
// installs that check as Smokescreen's egress ACL: Smokescreen allows every
// destination when it has none, so a proxy without one is never built, and
// a plugin without an allow-list may reach nothing (D-63).
func New(e *app.Egress, opts Options) (*Proxy, error) {
	switch {
	case e == nil:
		return nil, ErrNoEgress
	case opts.Listener == nil:
		return nil, ErrNoListener
	}
	cfg, tracker, err := newConfig(e, opts)
	if err != nil {
		return nil, err
	}
	var handler http.Handler = ss.BuildProxy(cfg)
	if opts.Healthcheck != nil {
		handler = ss.HealthcheckMiddleware{Proxy: handler, Healthcheck: opts.Healthcheck}
	}
	drain := opts.DrainTimeout
	if drain <= 0 {
		drain = DefaultDrainTimeout
	}
	return &Proxy{
		server: &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
		listener: opts.Listener,
		tracker:  tracker,
		drain:    drain,
		log:      cfg.Log,
	}, nil
}

// newConfig builds and validates the Smokescreen configuration and its
// connection tracker, doing the setup StartWithConfig would otherwise do.
func newConfig(e *app.Egress, opts Options) (*ss.Config, *conntrack.Tracker, error) {
	// Smokescreen's goproxy chains every CONNECT through HTTPS_PROXY when it
	// is set, which would bypass the resolver check below.
	if os.Getenv("HTTPS_PROXY") != "" || os.Getenv("https_proxy") != "" {
		return nil, nil, ErrProxyEnvironment
	}
	cfg := ss.NewConfig()
	if opts.Log != nil {
		cfg.Log = opts.Log
	}
	cfg.Listener = opts.Listener
	if ap, err := netip.ParseAddrPort(opts.Listener.Addr().String()); err == nil {
		cfg.Port = ap.Port()
	}
	cfg.RoleFromRequest = func(req *http.Request) (string, error) { return identify(e, req) }
	cfg.EgressACL = decider{egress: e}
	if err := cfg.SetDenyRanges(forbiddenRanges()); err != nil {
		return nil, nil, fmt.Errorf("egress proxy deny ranges: %w", err)
	}
	base := opts.Resolver
	if base == nil {
		base = cfg.Resolver
	}
	cfg.Resolver = checkingResolver{base: base}
	cfg.ProxyDialTimeout = dialer(opts.Dial)
	cfg.IdleTimeout = tunnelIdleTimeout
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("egress proxy config: %w", err)
	}
	// Refuse destinations on the proxy's own addresses, on every port.
	if err := cfg.InitializeSelfConnectionDetection(); err != nil {
		return nil, nil, fmt.Errorf("egress proxy self-connection guard: %w", err)
	}
	tracker := conntrack.NewTracker(cfg.IdleTimeout, cfg.MetricsClient, cfg.Log, atomic.Value{}, nil)
	tracker.ShuttingDown.Store(false)
	cfg.ConnTracker = tracker
	return cfg, tracker, nil
}

// dialer adapts a dial function, or a plain net.Dialer when nil, to
// Smokescreen's dial-with-timeout hook.
func dialer(dial func(ctx context.Context, network, address string) (net.Conn, error)) func(context.Context, string, string, time.Duration) (net.Conn, error) {
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}
	return func(ctx context.Context, network, address string, timeout time.Duration) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return dial(ctx, network, address)
	}
}

// Serve accepts plugins' connections until ctx ends, then stops accepting,
// lets open tunnels finish for at most the drain timeout and closes the
// rest. It returns nil after a shutdown that ctx asked for.
func (p *Proxy) Serve(ctx context.Context) error {
	served := make(chan error, 1)
	go func() { served <- p.server.Serve(p.listener) }()
	select {
	case err := <-served:
		p.closeTunnels()
		return fmt.Errorf("egress proxy serve: %w", err)
	case <-ctx.Done():
	}
	p.tracker.ShuttingDown.Store(true)
	drainCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), p.drain)
	defer cancel()
	err := p.server.Shutdown(drainCtx)
	p.waitForTunnels(drainCtx)
	p.closeTunnels()
	if serveErr := <-served; !errors.Is(serveErr, http.ErrServerClosed) {
		err = errors.Join(err, serveErr)
	}
	if err != nil {
		return fmt.Errorf("egress proxy shutdown: %w", err)
	}
	return nil
}

// waitForTunnels waits until every tunnel has closed or ctx ends.
func (p *Proxy) waitForTunnels(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		p.tracker.Wg().Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

// closeTunnels closes every tunnel still open. Server.Shutdown does not:
// Smokescreen hijacks the connections behind CONNECT tunnels.
func (p *Proxy) closeTunnels() {
	p.tracker.Range(func(k, _ any) bool {
		if c, ok := k.(*conntrack.InstrumentedConn); ok {
			if err := c.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				p.log.WithError(err).Warn("closing a tunnel at shutdown")
			}
		}
		return true
	})
}

// identify is the role function: the plugin named by the request's token.
func identify(e *app.Egress, req *http.Request) (string, error) {
	token, err := proxyToken(req.Header.Get("Proxy-Authorization"))
	if err != nil {
		return "", err
	}
	plugin, err := e.Identify(req.Context(), token)
	if err != nil {
		return "", fmt.Errorf("identify plugin: %w", err)
	}
	return plugin, nil
}

// proxyToken reads the token from "Basic base64(user:token)", the header
// HTTP clients send for a proxy URL with credentials (D-63). The user name
// is ignored; the scheme is case-insensitive (RFC 9110). An absent header
// means no token.
func proxyToken(header string) (string, error) {
	if header == "" {
		return "", app.ErrNoToken
	}
	scheme, encoded, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Basic") {
		return "", ErrBadProxyAuth
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrBadProxyAuth
	}
	_, token, ok := strings.Cut(string(decoded), ":")
	if !ok || token == "" || strings.ContainsAny(token, " \t\r\n") {
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
