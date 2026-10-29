package app

import (
	"context"
	"fmt"
	"io"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// SourceSync is the ingest path of a batch the core pulled from a connector
// plugin with Sync.
const SourceSync = "sync"

// Delivery is what a connector plugin returns for one import batch: its
// mapping of what the institution reported, plus any prices, FX rates,
// index readings and market index levels, and for a sync the watermark the
// next one continues from.
type Delivery struct {
	Institution   ingest.InstitutionID
	Mapping       ingest.Mapping
	Prices        []valuation.Price
	FXRates       []valuation.FXRate
	IndexReadings []ingest.IndexReading
	IndexLevels   []valuation.IndexLevel
	Watermark     string
}

// ConnectorCaller calls connector plugins' connector contract (D-49). The
// instance names the connector instance (D-48). Implementations
// authenticate as the core (D-59) and translate the contract into a
// Delivery; a delivery for another batch than id is an error.
type ConnectorCaller interface {
	Sync(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, watermark string) (Delivery, error)
	ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file NamedFile) (Delivery, error)
}

// NamedFile is a statement file to import: its contents, the name the owner
// gave it (a hint, never a path) and its media type when known.
type NamedFile struct {
	Name      string
	MediaType string
	Contents  io.Reader
}

// CoreTokens issues the token the core presents to a connector instance: a
// short-lived ServiceAccount token whose audience is that instance (D-59).
// The identity context implements it with the TokenRequest API.
type CoreTokens interface {
	TokenFor(ctx context.Context, instance string) (string, error)
}

// PriceStore records end-of-day prices (D-21). Recording a price again for
// the same instrument, date and source replaces it, so a retried batch has
// no further effect.
type PriceStore interface {
	StorePrices(ctx context.Context, batch ingest.BatchID, prices []valuation.Price) error
}

// FXRateStore records FX rates (D-20), idempotently like PriceStore, keyed
// by currency, date and source.
type FXRateStore interface {
	StoreFXRates(ctx context.Context, batch ingest.BatchID, rates []valuation.FXRate) error
}

// IndexReadingStore records index readings, idempotently like PriceStore,
// keyed by index, date and source.
type IndexReadingStore interface {
	StoreIndexReadings(ctx context.Context, batch ingest.BatchID, readings []ingest.IndexReading) error
}

// IndexLevelStore records market index levels (D-51), idempotently like
// PriceStore, keyed by index and date.
type IndexLevelStore interface {
	StoreIndexLevels(ctx context.Context, batch ingest.BatchID, levels []valuation.IndexLevel) error
}

// ConnectorBindings knows which accounts each connector instance may write
// to: its connector bindings.
type ConnectorBindings interface {
	BoundAccounts(ctx context.Context, instance string) ([]ledger.AccountID, error)
}

// SyncResult is a completed sync: the import batch and the watermark the
// next sync continues from.
type SyncResult struct {
	Batch     ingest.Batch
	Watermark string
}

// Connectors runs connector plugins' ingest paths (D-7, D-49): the core
// calls a connector instance to sync or to import a file, and imports what
// it delivers. The caller of Sync and ImportFile owns the schedule and the
// retries, and retries a failed batch with the same batch ID.
//
// A delivery is imported in two steps that are separate database
// transactions: the ledger first, then the valuation inputs. A failure
// between them leaves the transactions imported and the inputs not, and the
// batch fails; retrying it with the same batch ID imports the transactions
// again with no further effect (D-18) and stores the inputs, which replace
// any stored for the same key. The caller stores a sync's watermark only
// after both steps succeed.
type Connectors struct {
	Caller ConnectorCaller
	// Bindings says which accounts each instance may write to. A delivery
	// that names any other account is refused whole.
	Bindings ConnectorBindings
	Ledger   LedgerImporter
	Prices   PriceStore
	FXRates  FXRateStore
	Indices  IndexReadingStore
	Levels   IndexLevelStore
	// Today returns the date a batch starts on.
	Today func() kernel.Date
}

// Sync pulls instance's data since watermark as import batch id and
// imports it.
func (c *Connectors) Sync(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, watermark string) (SyncResult, error) {
	d, err := c.Caller.Sync(ctx, instance, id, mode, watermark)
	if err != nil {
		return SyncResult{}, fmt.Errorf("sync %s batch %s: %w", instance, id, err)
	}
	batch, err := c.deliver(ctx, instance, id, SourceSync, mode, &d)
	if err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Batch: batch, Watermark: d.Watermark}, nil
}

// ImportFile hands file to instance as import batch id and imports what it
// maps (D-49: uploads go through the core to the right connector).
func (c *Connectors) ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file NamedFile) (ingest.Batch, error) {
	d, err := c.Caller.ImportFile(ctx, instance, id, mode, file)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("import file into %s batch %s: %w", instance, id, err)
	}
	return c.deliver(ctx, instance, id, SourceFile, mode, &d)
}

// deliver imports a delivery as one import batch: transactions scoped to the
// instance into the ledger, seeded in an opening batch (D-23), then prices,
// FX rates, index readings and index levels with the instance as their
// source. Every
// index reading must be valid and every transaction and balance assertion
// must be in an account the instance is bound to, or nothing is imported.
func (c *Connectors) deliver(ctx context.Context, instance string, id ingest.BatchID, source string, mode ingest.Mode, d *Delivery) (ingest.Batch, error) {
	batch, err := ingest.NewBatch(id, d.Institution, source, mode, c.Today())
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s: %w", instance, err)
	}
	readings, err := sourcedReadings(instance, d.IndexReadings)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	m := d.Mapping
	if err = c.checkBound(ctx, instance, &m); err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	m.Transactions = ingest.ScopeToConnector(instance, m.Transactions)
	imported, err := importMapping(ctx, c.Ledger, mode, &m)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	if err := c.storeInputs(ctx, instance, id, d, readings); err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	batch.Complete(&m, imported)
	return batch, nil
}

// checkBound refuses m unless every account it names is bound to instance.
func (c *Connectors) checkBound(ctx context.Context, instance string, m *ingest.Mapping) error {
	if len(m.Transactions) == 0 && len(m.Assertions) == 0 {
		return nil
	}
	bound, err := c.Bindings.BoundAccounts(ctx, instance)
	if err != nil {
		return fmt.Errorf("connector bindings: %w", err)
	}
	if err := ingest.CheckBound(instance, bound, m); err != nil {
		return fmt.Errorf("refused: %w", err)
	}
	return nil
}

// storeInputs stores a delivery's prices, FX rates and the validated index
// readings, each with the instance as its source; an empty kind is skipped.
func (c *Connectors) storeInputs(ctx context.Context, instance string, id ingest.BatchID, d *Delivery, readings []ingest.IndexReading) error {
	if len(d.Prices) > 0 {
		prices := make([]valuation.Price, len(d.Prices))
		for i, p := range d.Prices {
			p.Source = instance
			prices[i] = p
		}
		if err := c.Prices.StorePrices(ctx, id, prices); err != nil {
			return fmt.Errorf("store prices: %w", err)
		}
	}
	if len(d.FXRates) > 0 {
		rates := make([]valuation.FXRate, len(d.FXRates))
		for i, r := range d.FXRates {
			r.Source = instance
			rates[i] = r
		}
		if err := c.FXRates.StoreFXRates(ctx, id, rates); err != nil {
			return fmt.Errorf("store FX rates: %w", err)
		}
	}
	if len(readings) > 0 {
		if err := c.Indices.StoreIndexReadings(ctx, id, readings); err != nil {
			return fmt.Errorf("store index readings: %w", err)
		}
	}
	return c.storeLevels(ctx, instance, id, d.IndexLevels)
}

// storeLevels stores a delivery's market index levels with the instance as
// their source; none is skipped.
func (c *Connectors) storeLevels(ctx context.Context, instance string, id ingest.BatchID, levels []valuation.IndexLevel) error {
	if len(levels) == 0 {
		return nil
	}
	out := make([]valuation.IndexLevel, len(levels))
	for i, l := range levels {
		l.Source = instance
		out[i] = l
	}
	if err := c.Levels.StoreIndexLevels(ctx, id, out); err != nil {
		return fmt.Errorf("store index levels: %w", err)
	}
	return nil
}

// sourcedReadings returns copies of readings with instance as their source,
// or the first reading's validation error.
func sourcedReadings(instance string, readings []ingest.IndexReading) ([]ingest.IndexReading, error) {
	out := make([]ingest.IndexReading, len(readings))
	for i, r := range readings {
		r.Source = instance
		if err := r.Validate(); err != nil {
			return nil, fmt.Errorf("index reading %d: %w", i, err)
		}
		out[i] = r
	}
	return out, nil
}
