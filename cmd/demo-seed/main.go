// Command demo-seed loads the demo dataset (internal/demo) into an OpenFief
// database, through the same paths real data takes: the IBKR history through
// the ibkrflex connector and the ingest service, the other accounts through
// the ledger's API, prices, FX rates, CPI and bond terms through valuation's
// input recorder, and the stored valuations through the River revaluation
// queue (D-36). It exits when every account is valued.
//
// It is for demo databases only. It refuses a database that holds ledger or
// valuation data without the demo marker it writes itself, so it cannot
// overwrite a real portfolio by accident; --reset-demo empties a marked demo
// database first. It logs internal IDs and counts only, never amounts.
//
// Usage:
//
//	OPENFIEF_DEMO_DSN=postgres://... demo-seed [--reset-demo] [--today 2026-10-28] [--seed N] [--offline]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // the distroless image has no zoneinfo for Europe/Warsaw

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// dsnEnv names the environment variable with the database's connection
// string. In the cluster it comes from CloudNativePG's app secret.
const dsnEnv = "OPENFIEF_DEMO_DSN"

const (
	// defaultSeed seeds the demo dataset unless --seed says otherwise.
	defaultSeed = 20261028
	// defaultWait bounds the revaluations unless --wait says otherwise.
	defaultWait = 10 * time.Minute
	// exitUsage is the exit status for bad arguments.
	exitUsage = 2
)

// errNoDSN reports a missing connection string.
var errNoDSN = errors.New(dsnEnv + " is not set")

// config is what one run of the loader needs.
type config struct {
	DSN   string
	Seed  uint64
	Today kernel.Date
	Reset bool
	// Offline skips NBP and uses the committed table A rates.
	Offline bool
	NBPURL  string
	// Wait bounds how long the revaluations may take.
	Wait time.Duration
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cfg, err := parseConfig(os.Args[1:], os.Getenv(dsnEnv), time.Now())
	if err != nil {
		log.Error("demo-seed: bad arguments", "error", err)
		os.Exit(exitUsage)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	err = run(ctx, &cfg, log)
	stop()
	if err != nil {
		log.Error("demo-seed failed", "error", err)
		os.Exit(1)
	}
}

// parseConfig reads the flags in args and the connection string dsn. Today
// defaults to now's date in Warsaw.
func parseConfig(args []string, dsn string, now time.Time) (config, error) {
	fs := flag.NewFlagSet("demo-seed", flag.ContinueOnError)
	reset := fs.Bool("reset-demo", false, "empty a database marked as a demo database before loading")
	today := fs.String("today", "", "the dataset's today, YYYY-MM-DD (default: today in Europe/Warsaw)")
	seed := fs.Uint64("seed", defaultSeed, "seed of the generated history and prices")
	offline := fs.Bool("offline", false, "use the committed NBP rates instead of fetching them")
	nbpURL := fs.String("nbp-url", nbp.DefaultBaseURL, "NBP Web API base URL")
	wait := fs.Duration("wait", defaultWait, "how long to wait for the revaluations")
	if err := fs.Parse(args); err != nil {
		return config{}, fmt.Errorf("parsing flags: %w", err)
	}
	if dsn == "" {
		return config{}, errNoDSN
	}
	cfg := config{DSN: dsn, Seed: *seed, Reset: *reset, Offline: *offline, NBPURL: *nbpURL, Wait: *wait}
	if *today == "" {
		loc, err := time.LoadLocation("Europe/Warsaw")
		if err != nil {
			return config{}, fmt.Errorf("time zone: %w", err)
		}
		t := now.In(loc)
		*today = t.Format(time.DateOnly)
	}
	d, err := kernel.ParseDate(*today)
	if err != nil {
		return config{}, fmt.Errorf("--today: %w", err)
	}
	cfg.Today = d
	return cfg, nil
}
