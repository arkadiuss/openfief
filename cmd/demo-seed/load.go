package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// errAlreadyLoaded explains a load that the ledger refused because the
// demo database holds another day's or another seed's dataset.
var errAlreadyLoaded = errors.New("the demo database holds a dataset of another day or seed: run with --reset-demo to replace it")

// run connects to the database and loads the demo dataset into it.
func run(ctx context.Context, cfg *config, log *slog.Logger) error {
	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	defer pool.Close()
	return load(ctx, pool, cfg, log)
}

// loader holds the wired use cases of one load.
type loader struct {
	ds       *demo.Dataset
	log      *slog.Logger
	ledger   *ledgerapp.Ledger
	store    *ledgerpg.Store
	vstore   *valuationpg.Store
	recorder *valuationapp.InputRecorder
	batches  int
}

// load checks the database (prepare), migrates it, and loads the dataset
// for cfg: the registry, the IBKR statements, the direct accounts and the
// valuation inputs, then waits until every account is valued.
func load(ctx context.Context, pool *pgxpool.Pool, cfg *config, log *slog.Logger) error {
	ds, err := demo.Generate(cfg.Seed, cfg.Today)
	if err != nil {
		return fmt.Errorf("generating the demo dataset: %w", err)
	}
	action, err := prepare(ctx, pool, cfg)
	if err != nil {
		return err
	}
	log.Info("demo database checked", "action", action.String(), "today", ds.Today.String(), "go_live", ds.GoLive.String())
	if migrateErr := migrate(ctx, pool, action); migrateErr != nil {
		return migrateErr
	}
	if done, topUpErr := topUpAnotherDataset(ctx, pool, cfg, action, log); done {
		return topUpErr
	}
	l, err := newLoader(pool, &ds, log)
	if err != nil {
		return err
	}
	rates, err := fxRates(ctx, cfg, &ds, log)
	if err != nil {
		return err
	}
	steps := []func(context.Context) error{l.registry, l.ibkr, l.direct, func(ctx context.Context) error { return l.inputs(ctx, rates) }}
	for _, step := range steps {
		if err := step(ctx); err != nil {
			if errors.Is(err, ledger.ErrAlreadySeeded) {
				return fmt.Errorf("%w: %w", errAlreadyLoaded, err)
			}
			return err
		}
	}
	return revalue(ctx, pool, l, cfg)
}

// migrate migrates the ledger and valuation schemas, after a reset with
// fresh connections: statements the pool's connections prepared on the
// dropped tables would fail on the new ones.
func migrate(ctx context.Context, pool *pgxpool.Pool, action demo.LoadAction) error {
	if action == demo.ResetAndLoad {
		pool.Reset()
	}
	if err := errors.Join(ledgerpg.Migrate(ctx, pool), valuationpg.Migrate(ctx, pool)); err != nil {
		return fmt.Errorf("migrating: %w", err)
	}
	return nil
}

// topUpAnotherDataset tops up (topUp) a database marked as holding another
// seed's or day's dataset than cfg's, and reports whether it did, with
// topUp's error; a load of cfg's own dataset goes on.
func topUpAnotherDataset(ctx context.Context, pool *pgxpool.Pool, cfg *config, action demo.LoadAction, log *slog.Logger) (bool, error) {
	if action != demo.LoadAgain {
		return false, nil
	}
	seed, day, err := loadedDataset(ctx, pool)
	if err != nil {
		return true, err
	}
	if seed == cfg.Seed && day.Equal(cfg.Today) {
		return false, nil
	}
	return true, topUp(ctx, pool, cfg, seed, day, log)
}

// topUp records again the valuation inputs of the dataset the database
// holds, that of seed and day, and then refuses to load cfg's with
// errAlreadyLoaded. Every input is an upsert by its key, so this changes
// nothing a load of that day wrote, and adds what a newer loader brings,
// such as the cross-check prices or the FX rates of older lots' tax dates,
// to a demo loaded by an older one. The ledger is left alone, and the
// revaluations the inputs enqueue are the running core's to work, through
// its own today.
func topUp(ctx context.Context, pool *pgxpool.Pool, cfg *config, seed uint64, day kernel.Date, log *slog.Logger) error {
	ds, err := demo.Generate(seed, day)
	if err != nil {
		return fmt.Errorf("generating the loaded demo dataset: %w", err)
	}
	l, err := newLoader(pool, &ds, log)
	if err != nil {
		return err
	}
	rates, err := fxRates(ctx, cfg, &ds, log)
	if err != nil {
		return err
	}
	if err := l.inputs(ctx, rates); err != nil {
		return err
	}
	log.Info("recorded the loaded dataset's valuation inputs again", "loaded_on", day.String(), "today", cfg.Today.String())
	return fmt.Errorf("%w: it holds the dataset of %s", errAlreadyLoaded, day)
}

// newLoader wires the ledger, with the valuation store's import hook that
// enqueues revaluations (D-36), and valuation's input recorder.
func newLoader(pool *pgxpool.Pool, ds *demo.Dataset, log *slog.Logger) (*loader, error) {
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store, vstore.ImportHook())
	recorder := valuationapp.NewInputRecorder(valuationapp.NewPlanner(l, store), vstore, pl.NewPack())
	return &loader{ds: ds, log: log, ledger: l, store: store, vstore: vstore, recorder: recorder}, nil
}

// registry saves the demo's sleeves and accounts.
func (l *loader) registry(ctx context.Context) error {
	for _, s := range l.ds.Sleeves {
		if err := l.ledger.SaveSleeve(ctx, s); err != nil {
			return fmt.Errorf("demo registry: %w", err)
		}
	}
	for _, a := range l.ds.Accounts {
		if err := l.ledger.SaveAccount(ctx, a); err != nil {
			return fmt.Errorf("demo registry: %w", err)
		}
	}
	return nil
}

// ibkr ingests the IBKR statements through the ibkrflex connector: the
// opening statement seeds the account (D-23), each activity statement is
// one import batch.
func (l *loader) ibkr(ctx context.Context) error {
	svc := &ingestapp.Service{
		Ledger:     importer{l.ledger},
		Mappers:    map[ingest.InstitutionID]ingestapp.Mapper{demo.IBKRInstitution: ibkrflex.Mapper{Accounts: map[string]ledger.AccountID{demo.IBKRFlexAccount: demo.IBKRAccount}}},
		Today:      func() kernel.Date { return l.ds.Today },
		NewBatchID: l.nextBatch,
	}
	if err := l.ingest(ctx, svc, ingest.ModeOpening, l.ds.IBKROpening); err != nil {
		return err
	}
	for _, st := range l.ds.IBKRActivity {
		if err := l.ingest(ctx, svc, ingest.ModeTransactions, st); err != nil {
			return err
		}
	}
	return nil
}

// ingest runs one import batch of a statement and logs its outcome.
func (l *loader) ingest(ctx context.Context, svc *ingestapp.Service, mode ingest.Mode, statement []byte) error {
	b, err := svc.IngestFile(ctx, demo.IBKRInstitution, mode, bytes.NewReader(statement))
	if err != nil {
		return fmt.Errorf("demo IBKR statement: %w", err)
	}
	l.log.Info("import batch", "batch", string(b.ID), "mode", b.Mode.String(), "transactions", b.Transactions,
		"new", b.Imported.New, "unchanged", b.Imported.Unchanged, "flagged_assertions", b.Imported.FailedAssertions,
		"problems", len(b.Problems))
	return nil
}

// nextBatch returns the next import batch ID of this load.
func (l *loader) nextBatch() ingest.BatchID {
	l.batches++
	return ingest.BatchID("demo-" + l.ds.Today.String() + "-" + strconv.Itoa(l.batches))
}

// direct books the XTB and bond accounts through the ledger's API: their
// seeding, then their history.
func (l *loader) direct(ctx context.Context) error {
	for _, acc := range []struct {
		id                ledger.AccountID
		opening, activity *demo.Batch
	}{
		{demo.XTBAccount, &l.ds.XTBOpening, &l.ds.XTBActivity},
		{demo.BondsAccount, &l.ds.BondsOpening, &l.ds.BondsActivity},
	} {
		seeded, err := l.ledger.Seed(ctx, acc.opening.Transactions, acc.opening.Assertions)
		if err != nil {
			return fmt.Errorf("demo account %s: %w", acc.id, err)
		}
		imported, err := l.ledger.Import(ctx, acc.activity.Transactions, acc.activity.Assertions)
		if err != nil {
			return fmt.Errorf("demo account %s: %w", acc.id, err)
		}
		l.log.Info("account booked", "account", string(acc.id), "seeded_new", seeded.New,
			"new", imported.New, "unchanged", imported.Unchanged+seeded.Unchanged,
			"flagged_assertions", len(seeded.Flagged())+len(imported.Flagged()))
	}
	return nil
}

// inputs records the valuation inputs (D-36): the bond account's tax
// wrapper, the bond terms, CPI, prices, FX rates and the version of the
// jurisdiction pack's rules. Each enqueues the revaluations it causes.
func (l *loader) inputs(ctx context.Context, rates []valuation.FXRate) error {
	if err := l.recorder.RecordTaxWrapper(ctx, demo.BondsAccount, jurisdiction.Wrapper(l.ds.TaxWrapper.String())); err != nil {
		return fmt.Errorf("demo inputs: %w", err)
	}
	for i := range l.ds.Bonds {
		b := &l.ds.Bonds[i]
		if err := l.recorder.RecordBondTerms(ctx, b.Instrument.ISIN, &b.Terms); err != nil {
			return fmt.Errorf("demo inputs: %w", err)
		}
	}
	err := errors.Join(
		l.recorder.RecordIndexReadings(ctx, l.ds.CPI),
		l.recorder.RecordPrices(ctx, l.ds.Prices),
		l.recorder.RecordFXRates(ctx, rates),
		valuationapp.NewCrossChecks(l.vstore, nil).RecordCrossCheckPrices(ctx, l.ds.CrossCheckPrices),
	)
	if err != nil {
		return fmt.Errorf("demo inputs: %w", err)
	}
	if _, err := l.recorder.ApplyRules(ctx); err != nil {
		return fmt.Errorf("demo inputs: %w", err)
	}
	l.log.Info("valuation inputs recorded", "prices", len(l.ds.Prices), "fx_rates", len(rates),
		"cpi_readings", len(l.ds.CPI), "bond_series", len(l.ds.Bonds), "cross_check_prices", len(l.ds.CrossCheckPrices))
	return nil
}

// importer is the ledger as the ingest service's LedgerImporter: it turns
// the ledger's import summary into the ingest context's.
type importer struct {
	ledger *ledgerapp.Ledger
}

// Import imports through the ledger.
func (i importer) Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	sum, err := i.ledger.Import(ctx, txs, assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("ledger: %w", err)
	}
	return summary(&sum), nil
}

// AddOpeningLots adds opening lots one at a time through the ledger.
func (i importer) AddOpeningLots(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	sum, err := i.ledger.AddOpeningLots(ctx, txs, assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("ledger: %w", err)
	}
	return summary(&sum), nil
}

// Seed seeds through the ledger.
func (i importer) Seed(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	sum, err := i.ledger.Seed(ctx, txs, assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("ledger: %w", err)
	}
	return summary(&sum), nil
}

// summary returns the ingest context's view of a ledger import summary.
func summary(s *ledgerapp.ImportSummary) ingest.ImportSummary {
	return ingest.ImportSummary{
		New: s.New, NewVersions: s.NewVersions, Unchanged: s.Unchanged,
		FailedAssertions: len(s.Flagged()), AlreadySeeded: s.AlreadySeeded,
	}
}
