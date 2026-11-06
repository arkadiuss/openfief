// Command fixture is the end-to-end suite's institution fixture server,
// built only for the suite (scripts/e2e-images.sh) and never released. It
// stands in for the external services the first-party network connectors
// call, so an end-to-end test never calls a real broker or vendor
// (docs/dev/testing.md): IBKR's Flex Web Service, EODHD's API and NBP's Web
// API, each below its own path, which the connectors reach through their
// baseUrl config fields:
//
//	https://<fixture>/ibkr/AccountManagement/FlexWebService   ibkr-flex
//	https://<fixture>/eodhd                                   eodhd
//	https://<fixture>/nbp                                     nbp-fx
//
// It serves HTTPS only, with the certificate and key files -cert and -key
// name, which the suite signs with a CA it makes for each run and hands the
// connectors as their extra CA bundle (<instance>--ca-bundle). Every answer
// is synthetic: fake account IDs, synthetic ISINs and made-up prices and
// rates; nothing here is real financial data.
//
// -flex-token and -eodhd-token, when set, are the only credentials it
// accepts, so a test can tell that a connector read its Secret; both are
// synthetic values the suite makes up.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Server timeouts.
const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	os.Exit(serve())
}

// serve runs the fixture server until SIGTERM or an interrupt and returns
// its exit code.
func serve() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	addr := flag.String("addr", ":8443", "address to serve HTTPS on")
	cert := flag.String("cert", "/etc/fixture/tls/tls.crt", "PEM certificate chain file")
	key := flag.String("key", "/etc/fixture/tls/tls.key", "PEM private key file")
	flexToken := flag.String("flex-token", "", "the only Flex Web Service token accepted; empty accepts any")
	eodhdToken := flag.String("eodhd-token", "", "the only EODHD API token accepted; empty accepts any")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	h := newHandler(credentials{flexToken: *flexToken, eodhdToken: *eodhdToken}, log)
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", *addr)
	if err != nil {
		log.Error("fixture stopped", "error", kernel.LogSafe(err))
		return 1
	}
	log.Info("fixture serving", "addr", ln.Addr().String())
	if err := run(ctx, ln, *cert, *key, h); err != nil {
		log.Error("fixture stopped", "error", kernel.LogSafe(err))
		return 1
	}
	return 0
}

// run serves h over HTTPS on ln with the certificate and key files until
// ctx is done, and closes ln.
func run(ctx context.Context, ln net.Listener, certFile, keyFile string, h http.Handler) error {
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return errors.Join(fmt.Errorf("TLS key pair: %w", err), ln.Close())
	}
	srv := &http.Server{
		Handler: h, ReadHeaderTimeout: readHeaderTimeout,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12},
	}
	errs := make(chan error, 1)
	go func() { errs <- srv.ServeTLS(ln, "", "") }()
	select {
	case err := <-errs:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdown); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errs; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
