package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
	viewskube "github.com/arkadiuss/openfief/internal/views/kube"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	"github.com/arkadiuss/openfief/sdk/go/telemetry"
)

const (
	// readHeaderTimeout bounds how long a client may take to send headers.
	readHeaderTimeout = 10 * time.Second
	// readTimeout bounds how long a client may take to send a whole
	// request: the built-in UI gives a statement upload of 64 MiB five
	// minutes.
	readTimeout = 5 * time.Minute
	// idleTimeout closes keep-alive connections idle this long.
	idleTimeout = 2 * time.Minute
	// maxAPIBodyBytes caps the body of a request to the core's own APIs:
	// one message of at most 1 MiB, its envelope and room to spare (D-184).
	maxAPIBodyBytes = 2 << 20
	// maxUploadBodyBytes caps the body of a statement upload: on the
	// browser port the form the front door passes to the view plugin, 64
	// MiB of file plus its fields, and on the plugin port the stream of
	// 64 KiB chunks the view plugin sends on, with their envelopes.
	maxUploadBodyBytes = 65 << 20
	// pingTimeout bounds the readiness check's database ping.
	pingTimeout = 2 * time.Second
	// retryEvery is how often start retries a database that is not reachable
	// yet, such as one still starting next to the core.
	retryEvery = 5 * time.Second
	// portCount is how many ports the core listens on: the browser port and
	// the plugin port (D-115).
	portCount = 2
)

// server answers the health endpoints at once and its two ports' APIs from
// when the core is ready until it starts shutting down (D-115): the browser
// port, which the Service's http port targets and a route points at, and
// the plugin port, which the Service's connect port targets and only plugin
// pods may reach.
type server struct {
	pool *pgxpool.Pool
	// browser is what the browser port serves past the health endpoints:
	// the front door and the read API.
	browser atomic.Pointer[http.Handler]
	// plugin is what the plugin port serves: the read API, the AuthService
	// and the owner services.
	plugin   atomic.Pointer[http.Handler]
	draining atomic.Bool
	// authzReady reports whether the authorizer can answer, once the core is
	// wired: nil when it always can, with the policy, or with SpiceDB's key
	// from the environment. Without the key it is false until the key is read
	// (D-164), so the Service sends no read to a core that refuses them all.
	authzReady atomic.Pointer[func() bool]
}

// ready stores the handlers c serves on each port; s then reports ready.
func (s *server) ready(c *core) {
	if g := c.access.gate; g != nil {
		ready := g.isConnected
		s.authzReady.Store(&ready)
	}
	s.plugin.Store(&c.pluginPort)
	s.browser.Store(&c.browserPort)
}

// browserHandler returns the browser port's mux: GET /healthz and GET
// /readyz, unauthenticated and never returning data, and the front door
// with the read API at every other path.
func (s *server) browserHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.Handle("/", serveWhenReady(&s.browser))
	return boundBodies(mux, browserBodyLimit)
}

// pluginHandler returns the plugin port's handler: the read API and the
// services only plugins call, with no health endpoint and no front door.
func (s *server) pluginHandler() http.Handler {
	return boundBodies(serveWhenReady(&s.plugin), pluginBodyLimit)
}

// boundBodies refuses a compressed request body with 415 and caps every
// other at the bytes limit returns for its path, answering 413 at once when
// the declared length is over it, before next authenticates or decodes
// anything (D-184).
func boundBodies(next http.Handler, limit func(path string) int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if connectapi.Compressed(r.Header) {
			plain(w, http.StatusUnsupportedMediaType, "compressed requests are not accepted")
			return
		}
		n := limit(r.URL.Path)
		if r.ContentLength > n {
			plain(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, n)
		next.ServeHTTP(w, r)
	})
}

// browserBodyLimit is the browser port's cap for path: the core's own
// paths, the read API among them, take maxAPIBodyBytes, and the paths the
// front door proxies to the view plugin maxUploadBodyBytes, for the
// statement upload form.
func browserBodyLimit(path string) int64 {
	if plugins.CoreServesPath(path) {
		return maxAPIBodyBytes
	}
	return maxUploadBodyBytes
}

// pluginBodyLimit is the plugin port's cap for path: the statement upload
// stream takes maxUploadBodyBytes, every other call maxAPIBodyBytes.
func pluginBodyLimit(path string) int64 {
	if path == ingestv1connect.StatementServiceUploadStatementProcedure {
		return maxUploadBodyBytes
	}
	return maxAPIBodyBytes
}

// healthz reports that the process serves HTTP: liveness.
func (*server) healthz(w http.ResponseWriter, _ *http.Request) {
	plain(w, http.StatusOK, "ok")
}

// readyz reports whether the core is ready: migrated and wired, not shutting
// down, with SpiceDB's key there when it is to be waited for, and with the
// database reachable now.
func (s *server) readyz(w http.ResponseWriter, r *http.Request) {
	switch {
	case s.draining.Load():
		plain(w, http.StatusServiceUnavailable, "shutting down")
		return
	case s.browser.Load() == nil:
		plain(w, http.StatusServiceUnavailable, "starting")
		return
	}
	if ready := s.authzReady.Load(); ready != nil && !(*ready)() {
		plain(w, http.StatusServiceUnavailable, "waiting for SpiceDB's preshared key")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		plain(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}
	plain(w, http.StatusOK, "ok")
}

// serveWhenReady hands each request to the handler h holds, or answers 503
// until the core stored it.
func serveWhenReady(h *atomic.Pointer[http.Handler]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api := h.Load()
		if api == nil {
			w.Header().Set("Retry-After", "5")
			plain(w, http.StatusServiceUnavailable, "starting")
			return
		}
		(*api).ServeHTTP(w, r)
	})
}

// plain writes a short plain-text answer that nothing may cache.
func plain(w http.ResponseWriter, status int, body string) {
	h := w.Header()
	h.Set("Content-Type", "text/plain; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, body); err != nil {
		return // the client went away
	}
}

// serve runs the core until ctx ends: it listens on both ports at once, so
// the health endpoints answer while it starts, then migrates, wires and
// starts the revaluation workers, and serves both ports' APIs. On shutdown
// it stops reporting ready, drains requests and running revaluations for up
// to the shutdown timeout, and cancels what is left.
func serve(ctx context.Context, cmd *command, log *slog.Logger, now func() time.Time) error {
	cfg := &cmd.Serve
	pool, err := newPool(ctx, cmd.DSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	s := &server{pool: pool}
	served := make(chan error, portCount)
	servers, err := listen(ctx, s, cfg, served, log)
	if err != nil {
		return err
	}
	c, startErr := start(ctx, pool, cfg, now, log, clusterOrNil(log))
	if startErr == nil {
		s.ready(c)
		log.Info("ready")
	}
	var serveErr error
	if startErr == nil {
		select {
		case <-ctx.Done():
		case serveErr = <-served:
		}
	}
	log.Info("shutting down")
	stopErr := shutdown(ctx, s, c, cfg.ShutdownTimeout, servers...)
	if errors.Is(startErr, context.Canceled) && ctx.Err() != nil {
		startErr = nil // stopped while starting
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(startErr, serveErr, stopErr)
}

// listen starts serving s's browser port on cfg.Listen and its plugin port
// on cfg.PluginListen (D-115), sending what each server's Serve returns to
// served. Every request but the health probes runs in a server span: on the
// browser port a new trace, whatever the caller sent; on the plugin port the
// trace the plugin propagated. When a port can't listen, it closes what it
// started.
func listen(ctx context.Context, s *server, cfg *serveConfig, served chan<- error, log *slog.Logger) ([]*http.Server, error) {
	ports := []struct {
		name, address string
		handler       http.Handler
	}{
		{"browser", cfg.Listen, telemetry.PublicHandler(s.browserHandler())},
		{"plugin", cfg.PluginListen, telemetry.ServerHandler(s.pluginHandler())},
	}
	servers := make([]*http.Server, 0, len(ports))
	for _, port := range ports {
		ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", port.address)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("listening on %s: %w", port.address, err), closeAll(servers))
		}
		httpSrv := &http.Server{
			Handler: port.handler, ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout: readTimeout, IdleTimeout: idleTimeout,
		}
		servers = append(servers, httpSrv)
		go func() { served <- httpSrv.Serve(ln) }()
		log.Info("listening", "port", port.name, "address", ln.Addr().String())
	}
	return servers, nil
}

// closeAll closes servers at once, for a start that failed.
func closeAll(servers []*http.Server) error {
	var errs []error
	for _, srv := range servers {
		if err := srv.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing an HTTP server: %w", err))
		}
	}
	return errors.Join(errs...)
}

// start migrates the database, retrying until it is reachable or ctx ends,
// wires the core and starts its revaluation workers and, in the
// background, the attention checks and outbox relay (D-130). In cluster cl it puts
// the front door before the browser port's read API, which proxies browsers
// to the view plugin that serves them (D-81, D-114), and, as cfg asks, starts the connector
// scheduler in the background, which can't fail the start, so the read API
// serves whatever the cluster does.
func start(ctx context.Context, pool *pgxpool.Pool, cfg *serveConfig, now func() time.Time, log *slog.Logger, cl *cluster) (*core, error) {
	for {
		err := migrate(ctx, pool)
		if err == nil {
			break
		}
		log.Warn("database not ready; retrying", "error", err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("starting: %w", ctx.Err())
		case <-time.After(retryEvery):
		}
	}
	callers, err := newCallers(cl, now, log)
	if err != nil {
		return nil, err
	}
	var installed *viewskube.Plugins
	if cl != nil {
		installed = viewskube.NewPlugins(cl.dynamic, cl.namespace, discoveryResync)
	}
	c, err := newCore(ctx, pool, cfg, now, log, callers, installed)
	if err != nil {
		return nil, err
	}
	if cl != nil {
		var proxy http.Handler
		c.browser, proxy = newFrontDoor(cl, cfg.Sync.ServiceAccount, now, log)
		c.browserPort = withFrontDoor(c.browserPort, proxy)
	}
	// The workers outlive ctx: shutdown stops them gracefully.
	if err := c.river.Start(context.WithoutCancel(ctx)); err != nil {
		return nil, fmt.Errorf("starting the revaluation workers: %w", err)
	}
	go c.attention.run(ctx, attentionEvery, log)
	// The cutover waits for the installed plugins to be read, so it runs in the background: until it has,
	// a holder that is to be given the household grant views nothing under SpiceDB.
	go c.settleGrants(ctx, cfg.Authz.Mode == authzSpiceDB, log)
	startCluster(ctx, c, cfg, cl, log)
	return c, nil
}

// settleGrants runs grantsStartup, once SpiceDB has its key when that is still
// to come: until then every change is refused, and the cutover would only
// retry.
func (c *core) settleGrants(ctx context.Context, spicedbOn bool, log *slog.Logger) {
	if g := c.access.gate; g != nil && !g.wait(ctx) {
		return // stopped while waiting for SpiceDB's key
	}
	grantsStartup(ctx, c.access.grants, spicedbOn, c.access.reconcile, log)
}

// startCluster starts, when the core runs in cluster cl, the watch of the
// installed plugins the read API lists (D-46), the watch of the view
// plugins that serve browsers (D-81), and the connector scheduler as cfg
// asks. None can fail the core.
func startCluster(ctx context.Context, c *core, cfg *serveConfig, cl *cluster, log *slog.Logger) {
	if c.browser != nil {
		go func() {
			if err := c.browser.Start(ctx); err != nil {
				log.Warn("the view plugins could not be read; browsers get 503", "error", err)
			}
		}()
	}
	if cl != nil && c.plugins != nil {
		go func() {
			if err := c.plugins.Start(ctx); err != nil {
				log.Warn("the installed plugins could not be read; the read API can't list them", "error", err)
			}
		}()
	}
	if cl == nil {
		if cfg.Connectors {
			log.Info("not in a Kubernetes cluster: no connector syncs")
		}
		return
	}
	if cfg.Connectors {
		c.connectors = make(chan struct{})
		go func() {
			defer close(c.connectors)
			runConnectors(ctx, c.deps, cl, &cfg.Sync)
		}()
	}
}

// closeAuthz closes the connection to SpiceDB, when there is one.
func (c *core) closeAuthz() error {
	if c.authz == nil {
		return nil
	}
	if err := c.authz.Close(); err != nil {
		return fmt.Errorf("closing the connection to SpiceDB: %w", err)
	}
	return nil
}

// shutdown marks s as draining, then stops the HTTP servers and c's
// workers, each letting in-flight work finish until timeout; workers still
// running then are cancelled, and River retries their jobs on the next start.
func shutdown(ctx context.Context, s *server, c *core, timeout time.Duration, servers ...*http.Server) error {
	s.draining.Store(true)
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	var errs []error
	for _, httpSrv := range servers {
		if err := httpSrv.Shutdown(sctx); err != nil {
			errs = append(errs, fmt.Errorf("stopping an HTTP server: %w", err))
		}
	}
	if c == nil {
		return errors.Join(errs...)
	}
	if c.connectors != nil {
		select { // the scheduler stops with ctx and records its attempts
		case <-c.connectors:
		case <-sctx.Done():
		}
	}
	if err := c.river.Stop(sctx); err != nil {
		if cancelErr := c.river.StopAndCancel(context.WithoutCancel(ctx)); cancelErr != nil {
			errs = append(errs, fmt.Errorf("stopping the revaluation workers: %w", cancelErr))
		}
	}
	errs = append(errs, c.closeAuthz())
	return errors.Join(errs...)
}
