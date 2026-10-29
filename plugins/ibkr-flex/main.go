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
// Config (OPENFIEF_CONFIG, checked against the manifest's schema) maps each
// IBKR account the Flex Query reports to the OpenFief account it feeds; a
// statement of any other account is skipped:
//
//	{"accounts": {"U0000000": "acc-ibkr"}}
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
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// requestTimeout bounds one request to the Flex Web Service.
const requestTimeout = time.Minute

// config is the plugin's configuration.
type config struct {
	Accounts map[string]string `json:"accounts"`
}

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
	var cfg config
	if err := openfief.LoadConfig(&cfg); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	hc, err := openfief.NewEgressClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	hc.Timeout = requestTimeout
	c := &connector{
		mapper:  ibkrflex.Mapper{Accounts: cfg.ledgerAccounts()},
		secrets: flexMount,
		fetch: func(ctx context.Context, token, queryID string) ([]byte, error) {
			client := ibkrflex.Client{HTTP: hc, Token: token, QueryID: queryID}
			return client.FetchStatement(ctx)
		},
	}
	log.Info("ibkr-flex serving", "accounts", len(cfg.Accounts))
	if err := openfief.ServeConnector(ctx, c); err != nil {
		return fmt.Errorf("ibkr-flex: %w", err)
	}
	return nil
}

// ledgerAccounts is the account map as the Flex mapper takes it.
func (c config) ledgerAccounts() map[string]ledger.AccountID {
	out := make(map[string]ledger.AccountID, len(c.Accounts))
	for ibkr, account := range c.Accounts {
		out[ibkr] = ledger.AccountID(account)
	}
	return out
}
