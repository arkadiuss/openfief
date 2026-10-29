// Command treasury-bonds is OpenFief's first-party Polish retail treasury
// bond connector plugin (D-9, D-49): the core calls its ImportBondLots with
// the lots the owner enters by hand in the built-in UI, and it maps each to
// ledger transactions and the series' terms. Neither the registry nor the
// bank that sells the bonds offers an API, so it reaches no host at all.
// The mapping is internal/ingest/treasurybonds; this binary only wires it
// to the SDK.
//
// Config (OPENFIEF_CONFIG, checked against the manifest's schema) names the
// OpenFief account its lots go into; the owner binds the instance to the
// same account (core connector bind). One instance serves one account
// (D-48):
//
//	{"account": "acc-bonds"}
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // the distroless image has no zoneinfo

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// warsaw is the time zone the bonds are sold in, which dates "today".
const warsaw = "Europe/Warsaw"

// config is the plugin's configuration.
type config struct {
	Account string `json:"account"`
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
		log.Error("treasury-bonds stopped", "error", err.Error())
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
	loc, err := time.LoadLocation(warsaw)
	if err != nil {
		return fmt.Errorf("time zone: %w", err)
	}
	c := &connector{
		mapper: treasurybonds.Mapper{Account: ledger.AccountID(cfg.Account)},
		today:  func() kernel.Date { return dateIn(time.Now(), loc) },
	}
	log.Info("treasury-bonds serving", "account", cfg.Account)
	if err := openfief.ServeConnector(ctx, c); err != nil {
		return fmt.Errorf("treasury-bonds: %w", err)
	}
	return nil
}

// dateIn is the calendar date of t in loc.
func dateIn(t time.Time, loc *time.Location) kernel.Date {
	y, m, d := t.In(loc).Date()
	day, err := kernel.NewDate(y, m, d)
	if err != nil { // a calendar date of the clock is always valid
		return kernel.Date{}
	}
	return day
}
