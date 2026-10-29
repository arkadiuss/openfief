package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	identitykube "github.com/arkadiuss/openfief/internal/identity/kube"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

const (
	// schedulerTick is how often the scheduler looks for due syncs.
	schedulerTick = time.Minute
	// discoveryResync is how often the instance caches relist.
	discoveryResync = 10 * time.Minute
	// discoveryTimeout bounds connector sync's wait for the instance caches.
	discoveryTimeout = 30 * time.Second
	// batchIDBytes is the random bytes in an import batch ID.
	batchIDBytes = 8
)

// errNoCluster reports a command that needs the Kubernetes API outside a
// cluster.
var errNoCluster = errors.New("not in a Kubernetes cluster: no in-cluster configuration")

// cluster is the Kubernetes API, as the core's ServiceAccount, in the
// openfief namespace.
type cluster struct {
	client    kubernetes.Interface
	dynamic   dynamic.Interface
	namespace string
}

// inCluster returns the cluster the core runs in, or errNoCluster outside
// one.
func inCluster() (*cluster, error) {
	cfg, err := rest.InClusterConfig()
	if errors.Is(err, rest.ErrNotInCluster) {
		return nil, errNoCluster
	}
	if err != nil {
		return nil, fmt.Errorf("in-cluster configuration: %w", err)
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubernetes client: %w", err)
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("kubernetes client: %w", err)
	}
	return &cluster{client: client, dynamic: dyn, namespace: plugins.Namespace}, nil
}

// clusterOrNil returns the cluster the core runs in, or nil outside one or
// when its configuration is unusable, which it logs.
func clusterOrNil(log *slog.Logger) *cluster {
	cl, err := inCluster()
	if err != nil && !errors.Is(err, errNoCluster) {
		log.Warn("no Kubernetes access", "error", err)
	}
	return cl
}

// connectorDeps are what the connector scheduler is built from.
type connectorDeps struct {
	pool     *pgxpool.Pool
	ledger   *ledgerapp.Ledger
	recorder *valuationapp.InputRecorder
	// crossChecks records cross-check prices (D-22).
	crossChecks *valuationapp.CrossChecks
	today       func() kernel.Date
	now         func() time.Time
	log         *slog.Logger
	// entries is the client the owner's manual entries go through; the
	// scheduler sets it once it has found the instances.
	entries *entryCaller
}

// withConnectorDeps connects to dsn, migrates, and runs fn with what the
// connector commands need; their batches start on today's date in UTC.
func withConnectorDeps(ctx context.Context, dsn string, log *slog.Logger, fn func(d *connectorDeps) error) error {
	return withPool(ctx, dsn, func(p *deps) error {
		if err := migrate(ctx, p.pool); err != nil {
			return err
		}
		vstore, err := valuationpg.NewStore(p.pool)
		if err != nil {
			return fmt.Errorf("valuation store: %w", err)
		}
		l, recorder := ingestDeps(p.pool, vstore)
		checks, err := valuationapp.NewCrossChecks(vstore, kernel.Decimal{})
		if err != nil {
			return fmt.Errorf("cross checks: %w", err)
		}
		return fn(&connectorDeps{
			pool: p.pool, ledger: l, recorder: recorder, crossChecks: checks, now: time.Now, log: log,
			today: valuationpg.TodayIn(time.UTC, time.Now),
		})
	})
}

// newScheduler wires the connector scheduler (D-49) over d: the instances
// and their addresses come from discovery, the core's tokens for them from
// the TokenRequest API (D-59), and the sync state and bindings from the
// ingest schema.
func newScheduler(d *connectorDeps, c *cluster, discovery *ingestkube.Discovery, cfg *syncConfig) *ingestapp.Scheduler {
	tokens := identitykube.NewCoreTokens(c.client, c.namespace, cfg.ServiceAccount, d.now)
	store := ingestpg.New(d.pool)
	caller := connectrpc.NewClient(connectrpc.H2CClient(), discovery.BaseURL, tokens)
	inputs := valuationInputs{recorder: d.recorder}
	return &ingestapp.Scheduler{
		Connectors: &ingestapp.Connectors{
			Caller: caller, Bindings: store, Ledger: ledgerImporter{ledger: d.ledger},
			Prices: inputs, FXRates: inputs, Indices: inputs, Levels: inputs, Terms: inputs,
			CrossChecks: crossCheckInputs{checks: d.crossChecks}, Today: d.today,
		},
		Instances:  discovery,
		Store:      store,
		Policy:     cfg.Policy,
		Timeout:    cfg.Timeout,
		Now:        d.now,
		NewBatchID: newBatchID,
		Jitter:     jitter,
		Log:        d.log,
	}
}

// runConnectors starts discovery and runs the scheduler until ctx ends.
// Once discovery runs, the owner's manual entries can reach the instances
// too (d.entries). Neither ever fails the core: a cluster that can't be
// reached is logged and leaves the read API serving.
func runConnectors(ctx context.Context, d *connectorDeps, c *cluster, cfg *syncConfig) {
	discovery := ingestkube.NewDiscovery(c.dynamic, c.client, c.namespace, discoveryResync)
	if err := discovery.Start(ctx); err != nil {
		d.log.Warn("connector discovery did not start; no syncs", "error", err)
		return
	}
	if d.entries != nil {
		tokens := identitykube.NewCoreTokens(c.client, c.namespace, cfg.ServiceAccount, d.now)
		d.entries.set(connectrpc.NewClient(connectrpc.H2CClient(), discovery.BaseURL, tokens))
	}
	ticker := time.NewTicker(schedulerTick)
	defer ticker.Stop()
	d.log.Info("connector scheduler running", "every", cfg.Policy.Interval.String())
	newScheduler(d, c, discovery, cfg).Run(ctx, ticker.C)
}

// connectorCommand runs connector bind or connector sync over the
// database, writing what it did to out.
func connectorCommand(ctx context.Context, cmd *command, d *connectorDeps, c *cluster, out io.Writer) error {
	args := &cmd.Connector
	switch cmd.Name {
	case cmdBind:
		at := d.now()
		if err := ingestapp.Bind(ctx, d.ledger, ingestpg.New(d.pool), args.Instance, args.Account, at); err != nil {
			return fmt.Errorf("connector bind: %w", err)
		}
		_, err := fmt.Fprintf(out, "bound connector instance %s to account %s\n", args.Instance, args.Account)
		return writeErr(err)
	case cmdSync:
		if c == nil {
			return fmt.Errorf("connector sync: %w", errNoCluster)
		}
		res, err := syncNow(ctx, d, c, args)
		if err != nil {
			return fmt.Errorf("connector sync: %w", err)
		}
		b := &res.Batch
		_, err = fmt.Fprintf(out, "batch %s (%s): %d transactions, %d new, %d new versions, %d unchanged, %d flagged assertions, %d problems\n",
			b.ID, b.Mode, b.Transactions, b.Imported.New, b.Imported.NewVersions, b.Imported.Unchanged, b.Imported.FailedAssertions, len(b.Problems))
		return writeErr(err)
	}
	return fmt.Errorf("%w: %q", errUsage, cmd.Name)
}

// syncNow syncs one instance at once, with the default sync settings.
func syncNow(ctx context.Context, d *connectorDeps, c *cluster, args *connectorArgs) (ingestapp.SyncResult, error) {
	dctx, cancel := context.WithTimeout(ctx, discoveryTimeout)
	defer cancel()
	discovery := ingestkube.NewDiscovery(c.dynamic, c.client, c.namespace, 0)
	if err := discovery.Start(dctx); err != nil {
		return ingestapp.SyncResult{}, fmt.Errorf("discovering instances: %w", err)
	}
	cfg := syncConfig{
		Policy:         ingest.SyncPolicy{Interval: defaultSyncEvery, Jitter: defaultSyncJitter, RetryFirst: defaultRetryFirst, RetryMax: defaultRetryMax},
		Timeout:        defaultSyncTimeout,
		ServiceAccount: defaultServiceAccount,
	}
	res, err := newScheduler(d, c, discovery, &cfg).SyncNow(ctx, args.Instance, args.Mode)
	if err != nil {
		return ingestapp.SyncResult{}, fmt.Errorf("syncing %s: %w", args.Instance, err)
	}
	return res, nil
}

// newBatchID returns a random import batch ID, such as "b-1a2b3c4d5e6f7a8b".
func newBatchID() ingest.BatchID {
	b := make([]byte, batchIDBytes)
	_, _ = rand.Read(b) // crypto/rand.Read never fails
	return ingest.BatchID("b-" + hex.EncodeToString(b))
}

// jitter returns a random duration in [0, limit).
func jitter(limit time.Duration) time.Duration {
	if limit <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		return 0
	}
	return time.Duration(n.Int64())
}

// ledgerImporter is the ledger as ingest's LedgerImporter: it turns the
// ledger's import summary into the ingest context's.
type ledgerImporter struct {
	ledger *ledgerapp.Ledger
}

// Import imports through the ledger.
func (i ledgerImporter) Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	sum, err := i.ledger.Import(ctx, txs, assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("ledger: %w", err)
	}
	return importSummary(&sum), nil
}

// Seed seeds through the ledger.
func (i ledgerImporter) Seed(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	sum, err := i.ledger.Seed(ctx, txs, assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("ledger: %w", err)
	}
	return importSummary(&sum), nil
}

// importSummary returns the ingest context's view of a ledger import
// summary.
func importSummary(s *ledgerapp.ImportSummary) ingest.ImportSummary {
	return ingest.ImportSummary{
		New: s.New, NewVersions: s.NewVersions, Unchanged: s.Unchanged,
		FailedAssertions: len(s.Flagged()), AlreadySeeded: s.AlreadySeeded,
	}
}

// valuationInputs records a delivery's prices, FX rates, index readings,
// index levels and bond terms through valuation's input use cases, each with the revaluations it causes
// (D-36). Recording one again replaces it, so a retried batch has no
// further effect.
type valuationInputs struct {
	recorder *valuationapp.InputRecorder
}

// StorePrices records prices.
func (v valuationInputs) StorePrices(ctx context.Context, _ ingest.BatchID, prices []valuation.Price) error {
	if err := v.recorder.RecordPrices(ctx, prices); err != nil {
		return fmt.Errorf("valuation: %w", err)
	}
	return nil
}

// StoreFXRates records FX rates.
func (v valuationInputs) StoreFXRates(ctx context.Context, _ ingest.BatchID, rates []valuation.FXRate) error {
	if err := v.recorder.RecordFXRates(ctx, rates); err != nil {
		return fmt.Errorf("valuation: %w", err)
	}
	return nil
}

// StoreIndexReadings records index readings as valuation's monthly
// readings: each for the month of its date.
func (v valuationInputs) StoreIndexReadings(ctx context.Context, _ ingest.BatchID, readings []ingest.IndexReading) error {
	out := make([]valuation.IndexReading, 0, len(readings))
	for i := range readings {
		r, err := monthlyReading(&readings[i])
		if err != nil {
			return err
		}
		out = append(out, r)
	}
	if err := v.recorder.RecordIndexReadings(ctx, out); err != nil {
		return fmt.Errorf("valuation: %w", err)
	}
	return nil
}

// crossCheckInputs records a delivery's cross-check prices through
// valuation's cross-check use cases, apart from its prices (D-22).
type crossCheckInputs struct {
	checks *valuationapp.CrossChecks
}

// StoreCrossCheckPrices records cross-check prices.
func (c crossCheckInputs) StoreCrossCheckPrices(ctx context.Context, _ ingest.BatchID, prices []valuation.CrossCheckPrice) error {
	if err := c.checks.RecordCrossCheckPrices(ctx, prices); err != nil {
		return fmt.Errorf("valuation: %w", err)
	}
	return nil
}

// StoreIndexLevels records market index levels.
func (v valuationInputs) StoreIndexLevels(ctx context.Context, _ ingest.BatchID, levels []valuation.IndexLevel) error {
	if err := v.recorder.RecordIndexLevels(ctx, levels); err != nil {
		return fmt.Errorf("valuation: %w", err)
	}
	return nil
}

// StoreBondTerms records treasury bond series terms (D-9), each for its
// instrument's ISIN; holdings of it are then valued as bonds.
func (v valuationInputs) StoreBondTerms(ctx context.Context, _ ingest.BatchID, terms []ingest.BondTerms) error {
	for i := range terms {
		if err := v.recorder.RecordBondTerms(ctx, terms[i].Instrument.ISIN, &terms[i].Terms); err != nil {
			return fmt.Errorf("valuation: %w", err)
		}
	}
	return nil
}

// errUnknownIndex reports an index reading of an index valuation doesn't
// know.
var errUnknownIndex = errors.New("unknown index")

// monthlyReading returns valuation's reading of r: the same index and value
// for the month of r's date. A reference rate that takes effect during a
// month becomes that month's reading.
func monthlyReading(r *ingest.IndexReading) (valuation.IndexReading, error) {
	var kind valuation.IndexKind
	switch r.Index {
	case ingest.IndexCPI:
		kind = valuation.CPI
	case ingest.IndexReferenceRate:
		kind = valuation.ReferenceRate
	default:
		return valuation.IndexReading{}, fmt.Errorf("%w: %s", errUnknownIndex, r.Index)
	}
	return valuation.IndexReading{Kind: kind, Year: r.Date.Year(), Month: r.Date.Month(), Value: r.Value, Source: r.Source}, nil
}
