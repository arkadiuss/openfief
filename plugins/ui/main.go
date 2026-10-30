// Command ui is OpenFief's built-in UI (D-3), a view plugin: it serves the
// owner's browser on the plugin port, through the core, which proxies the
// UI's browser path "/" to it (D-81), and talks to the core only. Every
// request must carry the core's token for this instance (D-59), which the
// SDK's verifier checks; only GET /healthz answers without one. It calls
// the core's AuthService with its own projected ServiceAccount token
// (D-27) and the public read API with the owner's session token (D-6). The
// pages, handlers and static files are internal/views/ui; this binary only
// wires them to the plugin's environment.
//
// The manifest asks for the view capability and no egress hosts, and takes
// no config. There is no way to run it without a projected token: local
// development runs it in a local cluster such as kind or k3d (D-84).
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
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/valuation/v1/valuationv1connect"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

const (
	// coreTimeout bounds one call to the core.
	coreTimeout = 30 * time.Second
	// uploadTimeout bounds one statement upload, the file streamed to the
	// core and the connector's import together: up to 64 MiB, on a slow
	// link from the owner's browser.
	uploadTimeout = 5 * time.Minute
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
// core's address, and the token the AuthService, ModelPortfolioService,
// BondLotService and StatementService clients present (authClient,
// uploadClient).
func newHandler(log *slog.Logger) (http.Handler, error) {
	coreURL, err := openfief.CoreURL()
	if err != nil {
		return nil, fmt.Errorf("core address: %w", err)
	}
	h, err := ui.New(&ui.Config{
		Auth:        identityv1connect.NewAuthServiceClient(authClient(openfief.TokenFile{}), coreURL),
		Model:       portfoliov1connect.NewModelPortfolioServiceClient(authClient(openfief.TokenFile{}), coreURL),
		BondLots:    ingestv1connect.NewBondLotServiceClient(authClient(openfief.TokenFile{}), coreURL),
		Statements:  ingestv1connect.NewStatementServiceClient(uploadClient(openfief.TokenFile{}), coreURL),
		Sleeves:     portfoliov1connect.NewSleeveServiceClient(authClient(openfief.TokenFile{}), coreURL),
		Accounts:    ingestv1connect.NewAccountServiceClient(authClient(openfief.TokenFile{}), coreURL),
		MarketData:  valuationv1connect.NewMarketDataServiceClient(authClient(openfief.TokenFile{}), coreURL),
		PriceChecks: valuationv1connect.NewPriceCheckServiceClient(authClient(openfief.TokenFile{}), coreURL),
		Read:        viewsv1connect.NewReadServiceClient(readClient(), coreURL),
		Log:         log,
	})
	if err != nil {
		return nil, fmt.Errorf("ui: %w", err)
	}
	verifier, err := openfief.NewCoreVerifier()
	if err != nil {
		return nil, fmt.Errorf("core verifier: %w", err)
	}
	return requireCore(verifier, h), nil
}

// requireCore admits only requests that carry the core's token for this
// instance (D-59): the core is the only entry point for browsers and sends
// them here with that token (D-81). GET /healthz, which shows nothing,
// stays open for probes.
func requireCore(v openfief.CoreVerifier, h http.Handler) http.Handler {
	guarded := openfief.RequireCore(v, h)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/healthz" {
			h.ServeHTTP(w, r)
			return
		}
		guarded.ServeHTTP(w, r)
	})
}

// authClient is the HTTP client for the AuthService and the owner's entry
// services. It presents the plugin's projected token from tokens, read on
// every call since it rotates.
func authClient(tokens openfief.TokenSource) *http.Client {
	c := openfief.CoreHTTPClient(tokens)
	c.Timeout = coreTimeout
	return c
}

// uploadClient is authClient for statement uploads, which may take up to
// uploadTimeout.
func uploadClient(tokens openfief.TokenSource) *http.Client {
	c := authClient(tokens)
	c.Timeout = uploadTimeout
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
