// Command ui is OpenFief's built-in UI (D-3), a view plugin: it serves the
// owner's browser on the plugin port and talks to the core only. It calls
// the core's AuthService with its own projected ServiceAccount token
// (D-27) and the public read API with the owner's session token (D-6). The
// pages, handlers and static files are internal/views/ui; this binary only
// wires them to the plugin's environment.
//
// The manifest asks for the view capability and no egress hosts, and takes
// no config. Outside a cluster, against a core run locally, the UI presents
// OPENFIEF_AUTH_CALLER_TOKEN to the AuthService instead of a projected
// token.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/views/ui"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

const (
	// coreTimeout bounds one call to the core.
	coreTimeout = 30 * time.Second
	// readHeaderTimeout bounds how long a browser may take to send headers.
	readHeaderTimeout = 10 * time.Second
	// idleTimeout closes idle keep-alive connections.
	idleTimeout = 2 * time.Minute
	// shutdownTimeout bounds how long requests in flight may run on after
	// SIGTERM.
	shutdownTimeout = 20 * time.Second
)

func main() {
	os.Exit(serve())
}

// serve runs the UI until SIGTERM or an interrupt and returns its exit code.
func serve() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := run(ctx, log); err != nil {
		log.Error("ui stopped", "error", err.Error())
		return 1
	}
	return 0
}

// run serves the UI until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	addr, err := openfief.ListenAddress()
	if err != nil {
		return fmt.Errorf("listen address: %w", err)
	}
	handler, err := newHandler(log)
	if err != nil {
		return err
	}
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: readHeaderTimeout, IdleTimeout: idleTimeout}
	log.Info("ui serving", "address", addr)
	return serveUntilDone(ctx, srv)
}

// newHandler builds the UI's handler from the plugin's environment: the
// core's address, and the token the AuthService, ModelPortfolioService and
// BondLotService clients present (authClient).
func newHandler(log *slog.Logger) (http.Handler, error) {
	coreURL, err := openfief.CoreURL()
	if err != nil {
		return nil, fmt.Errorf("core address: %w", err)
	}
	h, err := ui.New(&ui.Config{
		Auth:     identityv1connect.NewAuthServiceClient(authClient(os.Getenv), coreURL),
		Model:    portfoliov1connect.NewModelPortfolioServiceClient(authClient(os.Getenv), coreURL),
		BondLots: ingestv1connect.NewBondLotServiceClient(authClient(os.Getenv), coreURL),
		Read:     viewsv1connect.NewReadServiceClient(readClient(), coreURL),
		Log:      log,
	})
	if err != nil {
		return nil, fmt.Errorf("ui: %w", err)
	}
	return h, nil
}

// authCallerEnv names the token a UI running outside a cluster presents to
// the AuthService instead of a projected token: the core started with the
// same OPENFIEF_AUTH_CALLER_TOKEN accepts it, and only outside a cluster.
const authCallerEnv = "OPENFIEF_AUTH_CALLER_TOKEN"

// callerToken is a fixed AuthService caller token.
type callerToken string

// Token returns the fixed token.
func (c callerToken) Token() (string, error) { return string(c), nil }

// authClient is the HTTP client for the AuthService. It presents the token
// in OPENFIEF_AUTH_CALLER_TOKEN when that is set (outside a cluster), and
// otherwise the projected token, read from its file on every call since it
// rotates.
func authClient(getenv func(string) string) *http.Client {
	var c *http.Client
	if token := getenv(authCallerEnv); token != "" {
		c = openfief.CoreHTTPClient(callerToken(token))
	} else {
		c = openfief.CoreHTTPClient(openfief.TokenFile{})
	}
	c.Timeout = coreTimeout
	return c
}

// readClient is the HTTP client for the read API. It sets no Authorization
// of its own: each call carries the owner's session token. Like every call
// to the core it never uses a proxy.
func readClient() *http.Client {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	t := base.Clone()
	t.Proxy = nil
	return &http.Client{Transport: t, Timeout: coreTimeout}
}

// serveUntilDone runs srv until it fails or ctx is done, then shuts it down,
// giving requests in flight up to shutdownTimeout.
func serveUntilDone(ctx context.Context, srv *http.Server) error {
	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe() }()
	select {
	case err := <-done:
		return fmt.Errorf("serve ui: %w", err)
	case <-ctx.Done():
	}
	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(stopCtx); err != nil {
		return fmt.Errorf("shut down ui: %w", err)
	}
	if err := <-done; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve ui: %w", err)
	}
	return nil
}
