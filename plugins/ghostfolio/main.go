// Command ghostfolio is OpenFief's Ghostfolio view plugin (D-3, D-4, D-64):
// every interval it reads the record through the core's read API, with its
// own projected ServiceAccount token (D-27), and makes the Ghostfolio
// accounts it owns hold exactly the record's projection. It reaches an
// external Ghostfolio, the host of its config's url, through the egress
// proxy (D-32, D-75), or a Ghostfolio in the same cluster directly at the
// in-cluster service the owner picked at install (D-74). The reconcile loop
// is internal/views/ghostfolio; this binary only wires it to the plugin's
// environment and runs it on a timer.
//
// Config (OPENFIEF_CONFIG, checked against the manifest's schema):
//
//	{"url": "https://ghostfolio.example.com", "interval": "1h",
//	 "listings": [{"isin": "IETEST000026", "dataSource": "YAHOO", "symbol": "TEST.L"}]}
//
// url is the owner's Ghostfolio, unless it runs in the cluster and the
// install picked it as the in-cluster service, whose address the operator
// passes in OPENFIEF_CLUSTER_SERVICE_ADDRESS; interval is the time between syncs, an
// hour by default and at least five minutes; listings say under which
// Ghostfolio data source and symbol each listed instrument is known. The
// instance's Secret <instance>-security-token (D-89) holds the Ghostfolio user's security token
// in the file security-token.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1/viewsv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

const (
	// mountDir is where the manifest mounts the instance's security-token
	// Secret.
	mountDir = "/etc/openfief/ghostfolio"
	// keyFile is the Secret's file with the security token.
	keyFile = "security-token"
	// coreTimeout bounds one call to the core.
	coreTimeout = 30 * time.Second
	// ghostfolioTimeout bounds one call to Ghostfolio; an import of a full
	// batch with years of bond prices takes a while.
	ghostfolioTimeout = 2 * time.Minute
)

// Errors of the plugin's wiring.
var (
	// errNoToken reports a missing or empty security token file.
	errNoToken = errors.New("no Ghostfolio security token")
	// errNoGhostfolio reports neither a url nor an in-cluster service.
	errNoGhostfolio = errors.New("no Ghostfolio: set url or install with an in-cluster service")
	// errTwoGhostfolios reports both a url and an in-cluster service.
	errTwoGhostfolios = errors.New("both a url and an in-cluster service: give one")
)

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
		log.Error("ghostfolio stopped", "error", err.Error())
		return 1
	}
	return 0
}

// run syncs every interval until ctx is done.
func run(ctx context.Context, log *slog.Logger) error {
	var raw rawConfig
	if err := openfief.LoadConfig(&raw); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	cfg, err := raw.parse()
	if err != nil {
		return err
	}
	s, err := newSyncer(&cfg, os.Getenv)
	if err != nil {
		return err
	}
	log.Info("ghostfolio syncing", "interval", cfg.interval.String(), "supported_versions", ghostfolio.SupportedVersions)
	loop(ctx, s, cfg.interval, log)
	return nil
}

// newSyncer wires the reconcile loop from the plugin's environment: the
// core's read API with the plugin's projected token, the instance's name,
// and Ghostfolio with the security token from the Secret: through the
// egress proxy at the config's url, or, when the operator passed an
// in-cluster service's address, directly at that address over plain HTTP,
// which the install warned about (D-74).
func newSyncer(cfg *config, getenv func(string) string) (*ghostfolio.Syncer, error) {
	coreURL, err := openfief.CoreURL()
	if err != nil {
		return nil, fmt.Errorf("core address: %w", err)
	}
	instance := getenv(openfief.EnvInstance)
	if instance == "" {
		return nil, fmt.Errorf("instance: %w", openfief.ErrNoInstance)
	}
	core := openfief.CoreHTTPClient(openfief.TokenFile{})
	core.Timeout = coreTimeout
	record := viewsv1connect.NewReadServiceClient(core, coreURL)
	inCluster := getenv(plugins.EnvClusterServiceAddress)
	switch {
	case inCluster != "" && cfg.url != "":
		return nil, errTwoGhostfolios
	case inCluster != "":
		direct := *cfg
		direct.url = "http://" + inCluster
		// No proxy: the in-cluster service is reached at its ClusterIP.
		hc := &http.Client{Timeout: ghostfolioTimeout, Transport: &http.Transport{Proxy: nil}}
		return newSyncerWith(&direct, instance, record, hc, mountDir), nil
	case cfg.url == "":
		return nil, errNoGhostfolio
	}
	egress, err := openfief.NewEgressClient()
	if err != nil {
		return nil, fmt.Errorf("egress: %w", err)
	}
	egress.Timeout = ghostfolioTimeout
	return newSyncerWith(cfg, instance, record, egress, mountDir), nil
}

// newSyncerWith is the syncer of instance over record and Ghostfolio
// through hc, with the security token read from dir on every login.
func newSyncerWith(cfg *config, instance string, record ghostfolio.Record, hc *http.Client, dir string) *ghostfolio.Syncer {
	client := ghostfolio.NewClient(hc, cfg.url, func() (string, error) { return readSecretFile(dir, keyFile) })
	return ghostfolio.NewSyncer(record, client, instance, cfg.listings)
}

// syncer is one pass of the reconcile loop; *ghostfolio.Syncer is one.
type syncer interface {
	Sync(ctx context.Context) (ghostfolio.Report, error)
}

// loop syncs at once, then every interval, until ctx is done. A failed
// sync is logged and retried at the next tick: every step is safe to
// repeat (D-64). Only counts are logged, never an amount.
func loop(ctx context.Context, s syncer, interval time.Duration, log *slog.Logger) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		rep, err := s.Sync(ctx)
		if err != nil {
			log.Error("sync failed", "error", err.Error())
		} else {
			log.Info("synced", "accounts_created", rep.AccountsCreated, "imported", rep.Imported,
				"deleted", rep.Deleted, "accounts_cleared", rep.AccountsCleared, "balances_set", rep.BalancesSet,
				"balances_deleted", rep.BalancesDeleted, "left_out", len(rep.Problems))
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// readSecretFile reads one file of the Secret in dir, through an os.Root so
// the read can't leave dir. An error names the file, never its content.
func readSecretFile(dir, name string) (string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errNoToken, dir)
	}
	b, readErr := root.ReadFile(name)
	closeErr := root.Close()
	v := strings.TrimSpace(string(b))
	if readErr != nil || closeErr != nil || v == "" {
		return "", fmt.Errorf("%w: %s", errNoToken, filepath.Join(dir, name))
	}
	return v, nil
}
