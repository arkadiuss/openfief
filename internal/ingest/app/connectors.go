package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

const (
	// SourceSync is the ingest path of a batch the core pulled from a
	// connector plugin with Sync.
	SourceSync = "sync"
	// SourceEntry is the ingest path of a batch the owner entered by hand,
	// such as a treasury bond lot (D-9), which the core handed to a
	// connector plugin.
	SourceEntry = "entry"
	// SourceUpload is the ingest path of a statement file the owner
	// uploaded, which the core handed to a connector plugin (D-49).
	SourceUpload = "upload"
)

// Delivery is what a connector plugin returns for one import batch: its
// mapping of what the institution reported, plus any prices, FX rates,
// index readings, market index levels, treasury bond series terms and
// cross-check prices, and for a sync the watermark the next one continues
// from.
type Delivery struct {
	Institution   ingest.InstitutionID
	Mapping       ingest.Mapping
	Prices        []valuation.Price
	FXRates       []valuation.FXRate
	IndexReadings []ingest.IndexReading
	IndexLevels   []valuation.IndexLevel
	BondTerms     []ingest.BondTerms
	// CrossCheckPrices are the institution's own prices, which only
	// cross-check Prices (D-22) and never value a holding.
	CrossCheckPrices []valuation.CrossCheckPrice
	Watermark        string
}

// Errors a ConnectorCaller wraps when the connector answers that it won't
// take what the owner entered or uploaded, so the owner learns why.
var (
	// ErrEntryRefused reports entries the connector refused, such as a bond
	// lot that breaks a rule or a file it can't read; the error carries the
	// connector's reason.
	ErrEntryRefused = errors.New("the connector refused the entry")
	// ErrNoEntries reports a connector that takes no entries of the kind,
	// such as one that takes no bond lots or no files.
	ErrNoEntries = errors.New("the connector takes no such entries")
	// ErrConnectorUnavailable reports a connector instance that can't be
	// reached now, such as one that isn't running; the same entry may be
	// sent again later.
	ErrConnectorUnavailable = errors.New("the connector can't be reached")
	// ErrInvalidDelivery reports a connector's answer that doesn't translate
	// into a delivery, such as a malformed amount or ISIN; the error names
	// the field.
	ErrInvalidDelivery = errors.New("the connector's answer doesn't translate")
)

// ConnectorCaller calls connector plugins' connector contract (D-49). The
// instance names the connector instance (D-48). Implementations
// authenticate as the core (D-59) and translate the contract into a
// Delivery; a delivery for another batch than id is an error.
type ConnectorCaller interface {
	Sync(ctx context.Context, instance string, call *SyncCall) (Delivery, error)
	// ImportFile streams a statement file to instance. A file the connector
	// refuses wraps ErrEntryRefused with its reason, a connector that takes
	// no files ErrNoEntries, one that can't be reached
	// ErrConnectorUnavailable, and a failure to read the file the read's
	// error.
	ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file NamedFile) (Delivery, error)
	// ImportBondLots hands treasury bond lots the owner entered to instance
	// as a transactions batch. A refusal wraps ErrEntryRefused, a connector
	// that takes no bond lots ErrNoEntries.
	ImportBondLots(ctx context.Context, instance string, id ingest.BatchID, e *ingest.BondLotEntries) (Delivery, error)
}

// SyncCall is one Sync call to a connector instance: the import batch, its
// mode, the watermark of the previous sync, and, for a market-data
// connector, what to price (D-93), and for one bound to accounts, what each
// of them holds with the bond terms of their treasury bonds (D-117).
type SyncCall struct {
	ID        ingest.BatchID
	Mode      ingest.Mode
	Watermark string
	Scope     ingest.MarketDataScope
	Held      []ingest.HeldAccount
	BondTerms []ingest.BondTerms
}

// NamedFile is a statement file to import: its contents, the name the owner
// gave it (a hint, never a path) and its media type when known, with the
// cash balance the owner entered once with an opening upload of a
// statement that holds none, such as an xStation export's open positions
// (D-136): nil when none was entered.
type NamedFile struct {
	Name        string
	MediaType   string
	Contents    io.Reader
	OpeningCash *kernel.Money
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

// SourceStore records the market-data sources a delivery's values come
// from: the instance, with the institution it named, for each kind it
// delivered (D-94). The institution places a source in its default rank.
type SourceStore interface {
	StoreSources(ctx context.Context, batch ingest.BatchID, sources []valuation.Source) error
}

// CrossCheckPriceStore records cross-check prices (D-22) apart from
// prices: they value nothing. Recording one again for the same instrument,
// date and source replaces it, so a retried batch has no further effect.
type CrossCheckPriceStore interface {
	StoreCrossCheckPrices(ctx context.Context, batch ingest.BatchID, prices []valuation.CrossCheckPrice) error
}

// BondTermsStore records treasury bond series terms (D-9), which make the
// core value holdings of their instruments as bonds. Recording a series'
// terms again replaces them, so a retried batch has no further effect.
type BondTermsStore interface {
	StoreBondTerms(ctx context.Context, batch ingest.BatchID, terms []ingest.BondTerms) error
}

// AccountSeeding tells whether an account is seeded, and its go-live date
// (D-23): the date of its opening lots and opening cash. The ledger
// (ledger/app.Ledger.GoLive) answers it.
type AccountSeeding interface {
	GoLive(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error)
}

// AccountStarts keeps the go-live date an opening upload recorded for each
// account it started (D-160), even when nothing could be booked yet.
// *ingest/postgres.Store implements it.
type AccountStarts interface {
	// RecordStarts records each account's go-live; an account started
	// before keeps its date.
	RecordStarts(ctx context.Context, starts map[ledger.AccountID]kernel.Date) error
	// Start returns account's recorded go-live, and false when none was.
	Start(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error)
}

// ConnectorDeclarations tells what the running version of each connector
// instance declares in its signed manifest (D-112, D-109).
// ingest/kube.Discovery implements it from the Plugin resources.
type ConnectorDeclarations interface {
	Declarations(ctx context.Context, instance string) (ingest.Declarations, error)
}

// AttentionRaiser raises and clears attention items (D-130), such as a
// connector instance's undeclared cross-check prices (D-141). The ingest
// context declares it with plain strings, so it doesn't import the
// attention context; cmd/core implements it over the attention items.
type AttentionRaiser interface {
	// Raise raises the item of kind and subject with detail, or opens it
	// again when its detail changed.
	Raise(ctx context.Context, kind, subject, detail string) error
	// Clear resolves the open or acknowledged item of kind and subject.
	Clear(ctx context.Context, kind, subject string) error
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
	Terms    BondTermsStore
	// CrossChecks stores cross-check prices; a delivery that carries some
	// fails without it.
	CrossChecks CrossCheckPriceStore
	// Declarations tells which instances declare cross-check prices in their
	// manifest: only theirs are stored (D-112); and which are market-data
	// connectors (D-123), and which may deliver bond terms and index levels
	// (D-179). Without it, none declares.
	Declarations ConnectorDeclarations
	// Attention raises an attention item for an instance that delivered
	// bond terms or index levels it may not deliver, and clears it on a
	// delivery that carries none (D-179); nil raises nothing.
	Attention AttentionRaiser
	// Brokers keeps the broker accounts instances report and the OpenFief
	// accounts they are mapped to (D-72); a delivery that names broker
	// accounts fails without it.
	Brokers BrokerAccountStore
	// Sources records which instances delivered prices, FX rates and index
	// readings, for the owner's ranking of them (D-94); nil records none.
	Sources SourceStore
	// Seeding says which accounts are seeded and from when (D-23, D-95).
	// A sync never books again what an account's seeding holds: its
	// opening batch leaves out the seeding of accounts already seeded, and
	// its transactions batch the transactions dated on or before an
	// account's go-live (ingest.KeepFromGoLive). Without it, no account
	// counts as seeded.
	Seeding AccountSeeding
	// Starts keeps the go-live an opening upload recorded for an account
	// it started before anything could be booked (D-160): until the ledger
	// knows the account's go-live, it counts as seeded from that date. Nil
	// records none.
	Starts AccountStarts
	// Tickers keeps the tickers instances name instruments by without an
	// ISIN and their entries until the owner confirms each ticker's ISIN
	// (D-116); a delivery with ticker entries fails without it.
	Tickers TickerStore
	// Lookup asks the market-data connectors, the running Instances whose
	// Declarations make them one (D-123), to propose the ISIN of a ticker
	// first reported; nil, or nil Instances, proposes none.
	Lookup    InstrumentLookup
	Instances InstanceLister
	// Today returns the date a batch starts on.
	Today func() kernel.Date
}

// Sync pulls instance's data since the call's watermark as its import
// batch and imports it.
func (c *Connectors) Sync(ctx context.Context, instance string, call *SyncCall) (SyncResult, error) {
	d, err := c.Caller.Sync(ctx, instance, call)
	if err != nil {
		return SyncResult{}, fmt.Errorf("sync %s batch %s: %w", instance, call.ID, err)
	}
	batch, err := c.deliver(ctx, instance, call.ID, SourceSync, call.Mode, &d, nil)
	if err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Batch: batch, Watermark: d.Watermark}, nil
}

// GoLiveDates returns the go-live date of each of accounts that is seeded,
// as Seeding answers, or else started by an opening upload, as Starts
// answers (D-160); none without either.
func (c *Connectors) GoLiveDates(ctx context.Context, accounts []ledger.AccountID) (map[ledger.AccountID]kernel.Date, error) {
	out := map[ledger.AccountID]kernel.Date{}
	asked := map[ledger.AccountID]bool{}
	for _, a := range accounts {
		if asked[a] {
			continue
		}
		asked[a] = true
		day, live, err := c.goLive(ctx, a)
		if err != nil {
			return nil, fmt.Errorf("go-live of %s: %w", a, err)
		}
		if live {
			out[a] = day
		}
	}
	return out, nil
}

// ImportFile hands file, a statement the owner uploaded, to instance as
// import batch id and imports what it maps, with source SourceUpload (D-49:
// uploads go through the core to the right connector). A statement of
// several accounts, such as an IBKR Flex statement of linked accounts,
// books every mapped account bound to instance, and the batch lists the
// broker accounts that aren't mapped (D-109); a transaction or balance
// assertion in an account not bound to instance refuses it whole.
func (c *Connectors) ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file NamedFile) (ingest.Batch, error) {
	d, err := c.Caller.ImportFile(ctx, instance, id, mode, file)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("import file into %s batch %s: %w", instance, id, err)
	}
	b, err := c.deliver(ctx, instance, id, SourceUpload, mode, &d, nil)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("import file into %s: %w", instance, err)
	}
	return b, nil
}

// ImportBondLots hands treasury bond lots the owner entered for account, and
// those the owner deletes, to instance as transactions batch id, and
// imports what it maps (D-9, D-49: manual bond-lot entries go through the
// core to the right connector). It sends the account with the lots (D-118),
// so one instance serves every bond account. Each lot is booked on its
// real purchase date, as a deposit of its cost and a buy, so the account's
// history starts at its first purchase (D-98); only an account seeded with
// opening lots before that rule gets its go-live date, as Seeding answers,
// so those lots keep their keys. Any opening lots delivered are added to
// the account one at a time (LedgerImporter.AddOpeningLots). Every
// transaction and balance assertion delivered must be in account, or
// nothing is imported.
func (c *Connectors) ImportBondLots(ctx context.Context, instance string, id ingest.BatchID, account ledger.AccountID, entries *ingest.BondLotEntries) (ingest.Batch, error) {
	e := *entries
	e.Account = account
	goLive, err := c.GoLiveDates(ctx, []ledger.AccountID{account})
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("import bond lots into %s batch %s: %w", instance, id, err)
	}
	e.GoLive = goLive[account]
	d, err := c.Caller.ImportBondLots(ctx, instance, id, &e)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("import bond lots into %s batch %s: %w", instance, id, err)
	}
	b, err := c.deliver(ctx, instance, id, SourceEntry, ingest.ModeTransactions, &d, []ledger.AccountID{account})
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("import bond lots into %s: %w", instance, err)
	}
	return b, nil
}

// deliver imports a delivery as one import batch: entries named by broker
// account resolved into the accounts they are mapped to (D-72), in a sync,
// or an upload in transactions mode, what accounts' seeding already holds
// left out (D-95, D-136), ticker entries booked with their confirmed ISIN
// or held for the owner (D-116), the transactions
// scoped to the instance into the ledger, seeded in an
// opening batch (D-23), then prices, FX rates, index readings, index
// levels and cross-check prices with the instance as their source, the
// instance recorded as a source of each kind of market data, then
// treasury bond series terms. Every index reading must be valid and every
// transaction and balance assertion must be in an account the instance is
// bound to, and with only in one of only, or nothing is imported.
func (c *Connectors) deliver(ctx context.Context, instance string, id ingest.BatchID, source string, mode ingest.Mode, d *Delivery, only []ledger.AccountID) (ingest.Batch, error) {
	batch, err := ingest.NewBatch(id, d.Institution, source, mode, c.Today())
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s: %w", instance, err)
	}
	declared, err := c.declarations(ctx, instance, d.Institution)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	readings, err := sourcedReadings(instance, d.IndexReadings)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	r, starts, err := c.admit(ctx, instance, source, mode, d, only)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	m := r.Mapping
	if d, err = c.keepDeclared(ctx, instance, &declared, d, &m); err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	m.Transactions = ingest.ScopeToConnector(instance, m.Transactions)
	imported, err := importMapping(ctx, c.Ledger, mode, source, &m)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	if err := c.storeMarketData(ctx, instance, id, d, readings); err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	if err := c.markSeeded(ctx, instance, r.Seeding); err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	if err := c.recordStarts(ctx, source, mode, starts); err != nil {
		return ingest.Batch{}, fmt.Errorf("deliver from %s batch %s: %w", instance, id, err)
	}
	batch.Complete(&m, imported)
	batch.Unmapped = r.Unmapped
	return batch, nil
}

// declarations returns what instance's manifest declares (D-112, D-182),
// nothing without Declarations, after refusing a delivery naming an
// institution it doesn't declare (D-182).
func (c *Connectors) declarations(ctx context.Context, instance string, institution ingest.InstitutionID) (ingest.Declarations, error) {
	var none ingest.Declarations
	if c.Declarations == nil {
		if err := none.CheckInstitution(institution); err != nil {
			return none, fmt.Errorf("refused: %w", err)
		}
		return none, nil
	}
	declared, err := c.Declarations.Declarations(ctx, instance)
	if err != nil {
		return ingest.Declarations{}, fmt.Errorf("manifest declarations: %w", err)
	}
	if err := declared.CheckInstitution(institution); err != nil {
		return ingest.Declarations{}, fmt.Errorf("refused: %w", err)
	}
	return declared, nil
}

// declaredCrossChecks returns d as it is when it carries no cross-check
// prices or its connector's manifest declares them, and otherwise without
// them, adding to m the ingest problem that says none was stored (D-112).
func declaredCrossChecks(declared *ingest.Declarations, d *Delivery, m *ingest.Mapping) *Delivery {
	if len(d.CrossCheckPrices) == 0 || declared.CrossCheckPrices {
		return d
	}
	out := *d
	out.CrossCheckPrices = nil
	m.Problems = append(slices.Clone(m.Problems), ingest.UndeclaredCrossCheckPrices(len(d.CrossCheckPrices)))
	return &out
}

// keepDeclared returns d with only what instance's manifest declares, as
// declared says: without undeclared cross-check prices (D-112, D-141) and
// without bond terms and index levels it may not deliver (D-179).
func (c *Connectors) keepDeclared(ctx context.Context, instance string, declared *ingest.Declarations, d *Delivery, m *ingest.Mapping) (*Delivery, error) {
	return c.declaredDeliveries(ctx, instance, declared, declaredCrossChecks(declared, d, m), m)
}

// declaredDeliveries returns d without the bond terms and index levels
// instance may not deliver: those its signed manifest doesn't declare or
// its grant doesn't approve (D-179). Each kind dropped adds an ingest
// problem to m and raises the instance's one attention item, which a
// delivery that drops nothing clears. The rest of d is imported either
// way.
func (c *Connectors) declaredDeliveries(ctx context.Context, instance string, declared *ingest.Declarations, d *Delivery, m *ingest.Mapping) (*Delivery, error) {
	drop := ingest.UndeclaredDeliveries(*declared, len(d.BondTerms), len(d.IndexLevels))
	problems, detail := drop.Problems, drop.Detail
	if c.Attention != nil {
		var err error
		if detail == "" {
			err = c.Attention.Clear(ctx, ingest.KindUndeclaredDeliveries, instance)
		} else {
			err = c.Attention.Raise(ctx, ingest.KindUndeclaredDeliveries, instance, detail)
		}
		if err != nil {
			return nil, fmt.Errorf("attention item for undeclared deliveries: %w", err)
		}
	}
	if detail == "" {
		return d, nil
	}
	out := *d
	if !declared.BondTerms {
		out.BondTerms = nil
	}
	if !declared.IndexLevels {
		out.IndexLevels = nil
	}
	m.Problems = append(slices.Clone(m.Problems), problems...)
	return &out, nil
}

// keepFromGoLive returns m without what its accounts' seeding already holds
// (ingest.KeepFromGoLive), when it comes from source in mode: a sync, or a
// statement uploaded in transactions mode, which may reach back before the
// account's go-live (D-136). An opening upload is kept whole: the ledger
// decides whether it seeds the account again (D-23).
func (c *Connectors) keepFromGoLive(ctx context.Context, source string, mode ingest.Mode, m *ingest.Mapping) (ingest.Mapping, error) {
	if source != SourceSync && (source != SourceUpload || mode != ingest.ModeTransactions) {
		return *m, nil
	}
	accounts := make([]ledger.AccountID, 0, len(m.Transactions)+len(m.TickerEntries))
	for i := range m.Transactions {
		accounts = append(accounts, m.Transactions[i].Account)
	}
	for i := range m.TickerEntries {
		accounts = append(accounts, m.TickerEntries[i].Header.Account)
	}
	goLive, err := c.GoLiveDates(ctx, accounts)
	if err != nil {
		return ingest.Mapping{}, err
	}
	return ingest.KeepFromGoLive(m, mode, goLive), nil
}

// admit returns the mapping of d, from source, to import, with its broker
// accounts resolved, the broker accounts left unmapped and those an
// opening batch of it seeds, what its accounts' go-live keeps out left out
// (keepFromGoLive), and its ticker entries booked or held (bookTickers),
// after checking that every account it names is in only, when only is
// set, and bound to instance. It also returns the mapping to import before
// its ticker entries were booked or held, from which an opening upload
// records the accounts it starts (recordStarts).
func (c *Connectors) admit(ctx context.Context, instance, source string, mode ingest.Mode, d *Delivery, only []ledger.AccountID) (ingest.Resolution, *ingest.Mapping, error) {
	r, err := c.resolveBrokerAccounts(ctx, instance, mode, d)
	if err != nil {
		return ingest.Resolution{}, nil, err
	}
	if only != nil {
		if boundErr := ingest.CheckBound(instance, only, &r.Mapping); boundErr != nil {
			return ingest.Resolution{}, nil, fmt.Errorf("refused: %w", boundErr)
		}
	}
	if boundErr := c.checkBound(ctx, instance, &r.Mapping); boundErr != nil {
		return ingest.Resolution{}, nil, boundErr
	}
	kept, err := c.keepFromGoLive(ctx, source, mode, &r.Mapping)
	if err != nil {
		return ingest.Resolution{}, nil, err
	}
	booked, err := c.bookTickers(ctx, instance, &kept)
	if err != nil {
		return ingest.Resolution{}, nil, err
	}
	r.Mapping = booked
	return r, &kept, nil
}

// checkBound refuses m unless every account it names is bound to instance.
func (c *Connectors) checkBound(ctx context.Context, instance string, m *ingest.Mapping) error {
	if len(m.Transactions) == 0 && len(m.Assertions) == 0 && len(m.TickerEntries) == 0 {
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

// storeMarketData records the instance as a source of each kind of market
// data d carries, then stores its valuation inputs and its treasury bond
// series terms.
func (c *Connectors) storeMarketData(ctx context.Context, instance string, id ingest.BatchID, d *Delivery, readings []ingest.IndexReading) error {
	if err := c.storeSources(ctx, instance, id, d); err != nil {
		return err
	}
	if err := c.storeInputs(ctx, instance, id, d, readings); err != nil {
		return err
	}
	return c.storeTerms(ctx, instance, id, d.BondTerms)
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
	if err := c.storeCrossChecks(ctx, instance, id, d.CrossCheckPrices); err != nil {
		return err
	}
	return c.storeLevels(ctx, instance, id, d.IndexLevels)
}

// storeSources records instance, with the delivery's institution, as a
// source of each kind of market data d carries (D-94). Cross-check prices
// are no ranked source: they never win (D-22).
func (c *Connectors) storeSources(ctx context.Context, instance string, id ingest.BatchID, d *Delivery) error {
	if c.Sources == nil {
		return nil
	}
	var sources []valuation.Source
	for kind, n := range map[valuation.MarketDataKind]int{
		valuation.KindPrices: len(d.Prices), valuation.KindFXRates: len(d.FXRates), valuation.KindIndexReadings: len(d.IndexReadings),
	} {
		if n > 0 {
			sources = append(sources, valuation.Source{Kind: kind, Name: instance, Institution: string(d.Institution)})
		}
	}
	if len(sources) == 0 {
		return nil
	}
	slices.SortFunc(sources, func(a, b valuation.Source) int { return int(a.Kind) - int(b.Kind) })
	if err := c.Sources.StoreSources(ctx, id, sources); err != nil {
		return fmt.Errorf("store market-data sources: %w", err)
	}
	return nil
}

// ErrNoCrossCheckStore reports a delivery with cross-check prices to a
// Connectors that has nowhere to store them.
var ErrNoCrossCheckStore = errors.New("no store for cross-check prices")

// storeCrossChecks stores a delivery's cross-check prices with the
// instance as their source, apart from its prices (D-22); none is skipped.
func (c *Connectors) storeCrossChecks(ctx context.Context, instance string, id ingest.BatchID, prices []valuation.CrossCheckPrice) error {
	if len(prices) == 0 {
		return nil
	}
	if c.CrossChecks == nil {
		return ErrNoCrossCheckStore
	}
	out := make([]valuation.CrossCheckPrice, len(prices))
	for i, p := range prices {
		p.Source = instance
		out[i] = p
	}
	if err := c.CrossChecks.StoreCrossCheckPrices(ctx, id, out); err != nil {
		return fmt.Errorf("store cross-check prices: %w", err)
	}
	return nil
}

// storeTerms stores a delivery's treasury bond series terms with the
// instance as their source (D-179); none is skipped.
func (c *Connectors) storeTerms(ctx context.Context, instance string, id ingest.BatchID, terms []ingest.BondTerms) error {
	if len(terms) == 0 {
		return nil
	}
	out := slices.Clone(terms)
	for i := range out {
		out[i].Source = instance
	}
	if err := c.Terms.StoreBondTerms(ctx, id, out); err != nil {
		return fmt.Errorf("store bond terms: %w", err)
	}
	return nil
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

// goLive returns account's go-live: the ledger's when it is seeded, else
// the one an opening upload recorded (D-160), and whether it has either.
func (c *Connectors) goLive(ctx context.Context, account ledger.AccountID) (kernel.Date, bool, error) {
	if c.Seeding != nil {
		day, seeded, err := c.Seeding.GoLive(ctx, account)
		if err != nil {
			return kernel.Date{}, false, fmt.Errorf("seeding: %w", err)
		}
		if seeded {
			return day, true, nil
		}
	}
	if c.Starts == nil {
		return kernel.Date{}, false, nil
	}
	day, started, err := c.Starts.Start(ctx, account)
	if err != nil {
		return kernel.Date{}, false, fmt.Errorf("recorded start: %w", err)
	}
	return day, started, nil
}

// recordStarts records the go-live of each account an opening upload's
// mapping m starts (ingest.OpeningStarts, D-160); any other batch, or no
// Starts, records none.
func (c *Connectors) recordStarts(ctx context.Context, source string, mode ingest.Mode, m *ingest.Mapping) error {
	if c.Starts == nil || source != SourceUpload || mode != ingest.ModeOpening {
		return nil
	}
	if err := c.Starts.RecordStarts(ctx, ingest.OpeningStarts(m)); err != nil {
		return fmt.Errorf("recording account starts: %w", err)
	}
	return nil
}
