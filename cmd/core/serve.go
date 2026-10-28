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
)

const (
	// readHeaderTimeout bounds how long a client may take to send headers.
	readHeaderTimeout = 10 * time.Second
	// pingTimeout bounds the readiness check's database ping.
	pingTimeout = 2 * time.Second
	// retryEvery is how often start retries a database that is not reachable
	// yet, such as one still starting next to the core.
	retryEvery = 5 * time.Second
)

// server answers the health endpoints at once and the read API from when
// the core is ready until it starts shutting down.
type server struct {
	pool     *pgxpool.Pool
	api      atomic.Pointer[http.Handler]
	draining atomic.Bool
}

// handler returns the mux: GET /healthz and GET /readyz, unauthenticated
// and never returning data, and the read API at every other path.
func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthz)
	mux.HandleFunc("GET /readyz", s.readyz)
	mux.HandleFunc("/", s.serveAPI)
	return mux
}

// healthz reports that the process serves HTTP: liveness.
func (*server) healthz(w http.ResponseWriter, _ *http.Request) {
	plain(w, http.StatusOK, "ok")
}

// readyz reports whether the core is ready: migrated and wired, not shutting
// down, and with the database reachable now.
func (s *server) readyz(w http.ResponseWriter, r *http.Request) {
	switch {
	case s.draining.Load():
		plain(w, http.StatusServiceUnavailable, "shutting down")
		return
	case s.api.Load() == nil:
		plain(w, http.StatusServiceUnavailable, "starting")
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

// serveAPI hands r to the read API, or answers 503 until it is ready.
func (s *server) serveAPI(w http.ResponseWriter, r *http.Request) {
	api := s.api.Load()
	if api == nil {
		w.Header().Set("Retry-After", "5")
		plain(w, http.StatusServiceUnavailable, "starting")
		return
	}
	(*api).ServeHTTP(w, r)
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

// serve runs the core until ctx ends: it listens at once, so the health
// endpoints answer while it starts, then migrates, wires and starts the
// revaluation workers, and serves the read API. On shutdown it stops
// reporting ready, drains requests and running revaluations for up to the
// shutdown timeout, and cancels what is left.
func serve(ctx context.Context, cmd *command, log *slog.Logger, now func() time.Time) error {
	cfg := &cmd.Serve
	pool, err := pgxpool.New(ctx, cmd.DSN)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer pool.Close()
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Listen)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", cfg.Listen, err)
	}
	s := &server{pool: pool}
	httpSrv := &http.Server{Handler: s.handler(), ReadHeaderTimeout: readHeaderTimeout}
	served := make(chan error, 1)
	go func() { served <- httpSrv.Serve(ln) }()
	log.Info("listening", "address", ln.Addr().String())

	c, startErr := start(ctx, pool, cfg, now, log, clusterOrNil(log))
	if startErr == nil {
		s.api.Store(&c.api)
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
	stopErr := shutdown(ctx, s, httpSrv, c, cfg.ShutdownTimeout)
	if errors.Is(startErr, context.Canceled) && ctx.Err() != nil {
		startErr = nil // stopped while starting
	}
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(startErr, serveErr, stopErr)
}

// start migrates the database, retrying until it is reachable or ctx ends,
// wires the core and starts its revaluation workers. In cluster cl, and as
// cfg asks, it starts the connector scheduler in the background, which
// can't fail the start, so the read API serves whatever the cluster does.
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
	c, err := newCore(ctx, pool, cfg, now, log)
	if err != nil {
		return nil, err
	}
	// The workers outlive ctx: shutdown stops them gracefully.
	if err := c.river.Start(context.WithoutCancel(ctx)); err != nil {
		return nil, fmt.Errorf("starting the revaluation workers: %w", err)
	}
	startCluster(ctx, c, cfg, cl, log)
	return c, nil
}

// startCluster starts the connector scheduler, as cfg asks, when the core
// runs in cluster cl.
func startCluster(ctx context.Context, c *core, cfg *serveConfig, cl *cluster, log *slog.Logger) {
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

// shutdown marks s as draining, then stops the HTTP server and c's workers,
// each letting in-flight work finish until timeout; workers still running
// then are cancelled, and River retries their jobs on the next start.
func shutdown(ctx context.Context, s *server, httpSrv *http.Server, c *core, timeout time.Duration) error {
	s.draining.Store(true)
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	var errs []error
	if err := httpSrv.Shutdown(sctx); err != nil {
		errs = append(errs, fmt.Errorf("stopping the HTTP server: %w", err))
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
	return errors.Join(errs...)
}
