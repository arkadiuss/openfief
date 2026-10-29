// Command ibkr-flex is OpenFief's first-party Interactive Brokers Flex
// Query connector plugin (D-49): the core calls its Sync to pull the
// statement of the owner's Flex Query from IBKR's Flex Web Service, and its
// ImportFile to map an uploaded Flex statement, in the opening or
// transactions mode the call names (D-23). It reaches
// ndcdyn.interactivebrokers.com only, through the egress proxy (D-32). The
// Flex parser and mapping are internal/ingest/ibkrflex; this binary only
// wires them to the SDK.
//
// The Flex Web Service token and the Flex Query's ID come from the Secret
// ibkr-flex-credentials (files token and query-id), which the owner
// creates (D-29) and the manifest mounts at /etc/openfief/ibkr-flex. They
// are read on every sync and never logged.
//
// The plugin takes no config. It names every statement's entries by the
// IBKR account the statement is of, and reports every IBKR account it saw,
// so the core maps them to the OpenFief accounts the owner picked in the
// built-in UI and books nothing for an account the owner hasn't mapped
// (D-72). The owner never types an IBKR account number, and the plugin
// never learns an OpenFief account ID.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// requestTimeout bounds one request to the Flex Web Service.
const requestTimeout = time.Minute

func main() {
	os.Exit(serve())
}

// serve runs the plugin until SIGTERM or an interrupt and returns its exit
// code.
func serve() int {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := run(ctx, log); err != nil {
		log.Error("ibkr-flex stopped", "error", err.Error())
		return 1
	}
	return 0
}

// run serves the connector until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	hc, err := openfief.NewEgressClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	hc.Timeout = requestTimeout
	c := &connector{
		mapper:  ibkrflex.Mapper{ByBrokerAccount: true},
		secrets: flexMount,
		fetch: func(ctx context.Context, token, queryID string) ([]byte, error) {
			client := ibkrflex.Client{HTTP: hc, Token: token, QueryID: queryID}
			return client.FetchStatement(ctx)
		},
	}
	log.Info("ibkr-flex serving")
	if err := openfief.ServeConnector(ctx, c); err != nil {
		return fmt.Errorf("ibkr-flex: %w", err)
	}
	return nil
}
