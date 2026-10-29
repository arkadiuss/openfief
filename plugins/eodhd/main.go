// Command eodhd is OpenFief's first-party market-data connector plugin
// (D-22, D-51, D-49): the core calls its Sync for the end-of-day closes of
// the instruments the owner holds and the model portfolio targets, which
// each sync request names (D-93), and the levels of the model portfolio's
// market indices, from EODHD on the owner's own API key. It reaches eodhd.com only, through
// the egress proxy (D-32), and has no ImportFile. The EODHD adapter is
// internal/valuation/eodhd; this binary only wires it to the SDK.
//
// The API key comes from the Secret eodhd-credentials (file api-token),
// which the owner creates (D-29) and the manifest mounts at
// /etc/openfief/eodhd. It is read on every sync and never logged.
//
// Config (OPENFIEF_CONFIG, checked against the manifest's schema):
//
//	{"start": "2020-01-02",
//	 "instruments": [{"isin": "XX00000SYNT3", "exchange": "XETRA", "symbol": "SYNT"}],
//	 "indices": [{"name": "FAKE20", "ticker": "FAKE20.INDX"}]}
//
// Every part is optional. instruments adds instruments to watch beyond
// those the core asks for (D-93), and picks the exchange of one the core
// asks for that is listed on several. Each is found by its ISIN on the
// Warsaw Stock Exchange, the LSE, Xetra or Euronext; exchange (an EODHD
// exchange code) picks one when it is listed on several. symbol is its
// venue symbol as the ledger knows it, by default EODHD's ticker code; an
// instrument the core asks for is also delivered under the symbol it gave.
// indices names a market index's EODHD ticker when it isn't the index's
// own name plus .INDX; an index the core asks for by name is fetched as
// that ticker otherwise (D-96). start is the first day the first sync of
// each asks for; without it, EODHD's whole history.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// requestTimeout bounds one request to EODHD.
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
		log.Error("eodhd stopped", "error", err.Error())
		return 1
	}
	return 0
}

// run serves the connector until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	var raw rawConfig
	if err := openfief.LoadConfig(&raw); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	cfg, err := raw.parse()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	hc, err := openfief.NewEgressClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	hc.Timeout = requestTimeout
	log.Info("eodhd serving", "instruments", len(cfg.instruments), "indices", len(cfg.indices))
	if err := openfief.ServeConnector(ctx, newConnector(hc, cfg)); err != nil {
		return fmt.Errorf("eodhd: %w", err)
	}
	return nil
}

// newConnector is the connector reading EODHD's API with hc.
func newConnector(hc *http.Client, cfg config) *connector {
	return &connector{
		cfg: cfg, secrets: keyMount, today: today,
		source: func(token string) marketData { return eodhd.New(hc, eodhd.DefaultBaseURL, token) },
	}
}

// today is the current date in UTC. A close EODHD hasn't published yet is
// fetched by the next sync, which continues after the latest delivered one.
func today() kernel.Date {
	y, m, d := time.Now().UTC().Date()
	t, err := kernel.NewDate(y, m, d)
	if err != nil { // a calendar date of the clock is always valid
		return kernel.Date{}
	}
	return t
}
