// Command nbp-fx is OpenFief's first-party NBP FX connector plugin (D-20,
// D-49): the core calls its Sync for the National Bank of Poland's table A
// exchange rates to PLN since the previous sync. It reaches api.nbp.pl only,
// through the egress proxy (D-32), and has no ImportFile. The NBP adapter
// is internal/valuation/nbp; this binary only wires it to the SDK.
//
// Config (OPENFIEF_CONFIG, checked against the manifest's schema):
//
//	{"start": "2025-01-01"}
//
// start is the first day the first sync asks for; without it, the first
// sync covers the last 93 days.
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
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// requestTimeout bounds one request to NBP.
const requestTimeout = time.Minute

// config is the plugin's configuration.
type config struct {
	Start string `json:"start"`
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
		log.Error("nbp-fx stopped", "error", err.Error())
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
	start, err := cfg.startDate(today())
	if err != nil {
		return err
	}
	hc, err := openfief.NewEgressClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	hc.Timeout = requestTimeout
	log.Info("nbp-fx serving", "first_sync_from", start.String())
	if err := openfief.ServeConnector(ctx, newConnector(hc, start)); err != nil {
		return fmt.Errorf("nbp-fx: %w", err)
	}
	return nil
}

// newConnector is the connector reading NBP's public API with hc.
func newConnector(hc *http.Client, start kernel.Date) *connector {
	return &connector{rates: nbp.New(hc, nbp.DefaultBaseURL), today: today, start: start}
}

// startDate is the configured start, or the first of the last
// nbp.MaxRangeDays days up to today.
func (c config) startDate(today kernel.Date) (kernel.Date, error) {
	if c.Start == "" {
		return today.AddDays(1 - nbp.MaxRangeDays), nil
	}
	d, err := kernel.ParseDate(c.Start)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("config start: %w", err)
	}
	return d, nil
}

// today is the current date in UTC. NBP dates its tables in Polish time, at
// most two hours ahead, so a sync just after midnight there asks for one
// day less; the next sync gets it.
func today() kernel.Date {
	y, m, d := time.Now().UTC().Date()
	t, err := kernel.NewDate(y, m, d)
	if err != nil { // a calendar date of the clock is always valid
		return kernel.Date{}
	}
	return t
}
