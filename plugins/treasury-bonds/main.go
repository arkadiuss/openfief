// Command treasury-bonds is OpenFief's first-party Polish retail treasury
// bond connector plugin (D-9, D-49). The core calls its ImportBondLots with
// the lots the owner enters by hand in the built-in UI, and it maps each to
// ledger transactions and the series' terms: neither the registry nor the
// bank that sells the bonds offers an API (R-5). The core calls its Sync on
// its schedule for the period index readings that set the bonds' rates
// after the first interest period: year-on-year CPI from GUS's public SDP
// API (api-sdp.stat.gov.pl) and the NBP reference rate from NBP's published
// archive (static.nbp.pl), both reached anonymously through the egress
// proxy (D-32). The mapping and the index sync are
// internal/ingest/treasurybonds, the adapters internal/valuation/gus and
// internal/valuation/nbp; this binary only wires them to the SDK.
//
// Config (OPENFIEF_CONFIG, checked against the manifest's schema) names the
// OpenFief account its lots go into; the owner binds the instance to the
// same account (core connector bind). One instance serves one account
// (D-48). start, optional, is the first month the first sync delivers index
// readings for; without it, January 13 years before today's year:
//
//	{"account": "acc-bonds", "start": "2014-01-01"}
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
	_ "time/tzdata" // the distroless image has no zoneinfo

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/gus"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// warsaw is the time zone the bonds are sold in, which dates "today".
const warsaw = "Europe/Warsaw"

// requestTimeout bounds one request to GUS or NBP.
const requestTimeout = time.Minute

// errStartTooEarly reports a configured start before the holiday calendar
// the reference rate's day needs.
var errStartTooEarly = errors.New("start is before the Polish holiday calendar")

// config is the plugin's configuration.
type config struct {
	Account string `json:"account"`
	Start   string `json:"start"`
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
	today := func() kernel.Date { return dateIn(time.Now(), loc) }
	start, err := cfg.startDate(today())
	if err != nil {
		return err
	}
	hc, err := openfief.NewEgressClient()
	if err != nil {
		return fmt.Errorf("egress: %w", err)
	}
	hc.Timeout = requestTimeout
	c := newConnector(hc, ledger.AccountID(cfg.Account), start, today)
	log.Info("treasury-bonds serving", "account", cfg.Account, "first_sync_from", start.String())
	if err := openfief.ServeConnector(ctx, c); err != nil {
		return fmt.Errorf("treasury-bonds: %w", err)
	}
	return nil
}

// newConnector is the connector for account, reading GUS and NBP with hc
// from start on.
func newConnector(hc *http.Client, account ledger.AccountID, start kernel.Date, today func() kernel.Date) *connector {
	return &connector{
		mapper: treasurybonds.Mapper{Account: account},
		indices: &treasurybonds.IndexSync{
			CPI: gus.New(hc, gus.DefaultBaseURL), Rates: nbp.NewArchive(hc, nbp.DefaultArchiveURL), Start: start,
		},
		today: today,
	}
}

// startDate is the configured start, or treasurybonds.DefaultIndexStart.
// It must be in a year after pl.FirstCalendarYear, so that every month's
// reference-rate day is in the holiday calendar.
func (c config) startDate(today kernel.Date) (kernel.Date, error) {
	if c.Start == "" {
		return treasurybonds.DefaultIndexStart(today), nil
	}
	d, err := kernel.ParseDate(c.Start)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("config start: %w", err)
	}
	if d.Year() <= pl.FirstCalendarYear {
		return kernel.Date{}, fmt.Errorf("config start %s: %w", d, errStartTooEarly)
	}
	return d, nil
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
