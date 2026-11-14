package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
)

// The audit outbox (D-211): a change of state enqueues the entry of it as a
// River job in the transaction of the change, and a worker writes the entry to
// identity's log afterwards, once, keyed by the job's ID, with the change's own
// time. The change and its pending entry commit or roll back together; no
// transaction is held open on identity's tables, and a log that is down only
// delays the entries, which the worker keeps retrying.

// auditQueue is the River queue of the audit entries.
const auditQueue = "audit"

// txKey is the context key of the transaction a change runs in.
type txKey struct{}

// changeTx is the transaction a change runs in, as a context carries it.
type changeTx struct {
	pgx.Tx
}

// joinedPool is the core's database for the stores: a statement run with a
// context that carries a transaction (Join, Outbox.Within) runs in it, so every
// store a change uses, whichever context it belongs to, writes in the
// change's one transaction; with any other context it runs on the pool, as
// before. Begin inside a transaction opens a savepoint, so a step that fails
// and is handled rolls back alone, and one that is not handled rolls back the
// whole change.
type joinedPool struct {
	*pgxpool.Pool
}

// joined returns the pool whose statements join the transaction a context
// carries.
func joined(p *pgxpool.Pool) joinedPool { return joinedPool{Pool: p} }

// txOf returns the transaction ctx carries.
func txOf(ctx context.Context) (changeTx, bool) {
	tx, ok := ctx.Value(txKey{}).(changeTx)
	return tx, ok
}

// Join returns ctx carrying tx, which the statements run with it join.
func (joinedPool) Join(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, changeTx{Tx: tx})
}

// Exec runs sql in the transaction ctx carries, or on the pool.
func (j joinedPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	var (
		tag pgconn.CommandTag
		err error
	)
	if tx, ok := txOf(ctx); ok {
		tag, err = tx.Exec(ctx, sql, args...)
	} else {
		tag, err = j.Pool.Exec(ctx, sql, args...)
	}
	if err != nil {
		return tag, fmt.Errorf("executing a statement: %w", err)
	}
	return tag, nil
}

// Query runs sql in the transaction ctx carries, or on the pool. The returned
// rows are pgx's interface, as the stores' database port says.
func (j joinedPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if tx, ok := txOf(ctx); ok {
		rows, err = tx.Query(ctx, sql, args...)
	} else {
		rows, err = j.Pool.Query(ctx, sql, args...)
	}
	if err != nil {
		return nil, fmt.Errorf("running a query: %w", err)
	}
	return rows, nil
}

// QueryRow runs sql in the transaction ctx carries, or on the pool.
func (j joinedPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := txOf(ctx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return j.Pool.QueryRow(ctx, sql, args...)
}

// Begin opens a savepoint in the transaction ctx carries, or a transaction on
// the pool.
func (j joinedPool) Begin(ctx context.Context) (pgx.Tx, error) {
	var (
		tx  pgx.Tx
		err error
	)
	if outer, ok := txOf(ctx); ok {
		tx, err = outer.Begin(ctx)
	} else {
		tx, err = j.Pool.Begin(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("beginning a transaction: %w", err)
	}
	return tx, nil
}

// joinedStore returns the valuation store over pool whose statements join the
// transaction a context carries.
func joinedStore(pool *pgxpool.Pool) (*valuationpg.Store, error) {
	s, err := valuationpg.NewStore(pool, jobTracing()...)
	if err != nil {
		return nil, fmt.Errorf("valuation store: %w", err)
	}
	return s.On(joined(pool)), nil
}

// auditRef is a reference of an audit entry, as the job carries it.
type auditRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// auditEntryArgs is the River job that writes one audit entry to the log: who,
// what, which records by ID, the outcome and the time of the change, never an
// amount, a name or a secret (GR-12).
type auditEntryArgs struct {
	ActorKind string     `json:"actorKind"`
	ActorID   string     `json:"actorId"`
	Action    string     `json:"action"`
	Outcome   string     `json:"outcome"`
	Refs      []auditRef `json:"refs"`
	At        time.Time  `json:"at"`
}

// Kind names the job kind.
func (auditEntryArgs) Kind() string { return "identity.audit_entry" }

// InsertOpts puts the job on the audit queue and has River keep retrying it:
// an entry the worker can't write now, such as with the log down, is never
// dropped, and River's default backoff keeps the attempts from piling up.
func (auditEntryArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: auditQueue, MaxAttempts: math.MaxInt16}
}

// auditArgsOf returns the job that writes e.
func auditArgsOf(e *identity.AuditEntry) auditEntryArgs {
	refs := make([]auditRef, 0, len(e.Refs))
	for _, r := range e.Refs {
		refs = append(refs, auditRef{Kind: string(r.Kind), ID: r.ID})
	}
	return auditEntryArgs{
		ActorKind: string(e.Actor.Kind), ActorID: e.Actor.ID, Action: string(e.Action), Outcome: string(e.Outcome),
		Refs: refs, At: e.At,
	}
}

// entryOf returns the audit entry the job a carries, validated.
func entryOf(a *auditEntryArgs) (identity.AuditEntry, error) {
	action, err := identity.ParseAuditAction(a.Action)
	if err != nil {
		return identity.AuditEntry{}, fmt.Errorf("the audit entry of a job: %w", err)
	}
	outcome, err := identity.ParseAuditOutcome(a.Outcome)
	if err != nil {
		return identity.AuditEntry{}, fmt.Errorf("the audit entry of a job: %w", err)
	}
	refs := make([]identity.Ref, 0, len(a.Refs))
	for _, r := range a.Refs {
		kind, kindErr := identity.ParseRefKind(r.Kind)
		if kindErr != nil {
			return identity.AuditEntry{}, fmt.Errorf("the audit entry of a job: %w", kindErr)
		}
		refs = append(refs, identity.Ref{Kind: kind, ID: r.ID})
	}
	actor := identity.Actor{Kind: identity.ActorKind(a.ActorKind), ID: a.ActorID}
	e, err := identity.NewAuditEntry(a.At, actor, action, outcome, refs...)
	if err != nil {
		return identity.AuditEntry{}, fmt.Errorf("the audit entry of a job: %w", err)
	}
	return e, nil
}

// auditLog is what the worker writes entries to: *identitypg.Store.
type auditLog interface {
	AppendAuditOnce(ctx context.Context, key int64, e *identity.AuditEntry) (bool, error)
}

var _ auditLog = (*identitypg.Store)(nil)

// auditWorker writes the audit entries the outbox holds to the log (D-211).
type auditWorker struct {
	river.WorkerDefaults[auditEntryArgs]

	log auditLog
}

// Work writes the entry of job, with the change's time, once: a job that runs
// again, after a crash or a commit that was not acknowledged, writes nothing.
// Any error is returned, so River tries again, with its default backoff: the
// worker never drops an entry.
func (w *auditWorker) Work(ctx context.Context, job *river.Job[auditEntryArgs]) error {
	return w.write(ctx, job.ID, &job.Args)
}

// write is Work for the job with ID id and args.
func (w *auditWorker) write(ctx context.Context, id int64, args *auditEntryArgs) error {
	e, err := entryOf(args)
	if err != nil {
		return err
	}
	if _, err := w.log.AppendAuditOnce(ctx, id, &e); err != nil {
		return fmt.Errorf("writing the audit entry of job %d: %w", id, err)
	}
	return nil
}

// auditJob is the River job of the audit entries, to run on the core's River
// client: the queue and the worker that writes to log.
func auditJob(log auditLog) valuationpg.Job {
	return valuationpg.Job{
		Queue: auditQueue, MaxWorkers: 1,
		Register: func(w *river.Workers) error { return river.AddWorkerSafely(w, &auditWorker{log: log}) },
	}
}

// ErrNoChangeTx reports an entry enqueued outside a transaction of a change.
var ErrNoChangeTx = errors.New("no transaction of a change to enqueue the audit entry in")

// maxSyncAccountsNamed is the most accounts a sync's entry names (see
// identity.ActionConnectorSynced); a sync that wrote in more names none.
const maxSyncAccountsNamed = 24

// outbox enqueues audit entries in the transaction of the changes they
// record, as River jobs (D-211).
type outbox struct {
	pool   *pgxpool.Pool
	insert *river.Client[pgx.Tx]
	now    func() time.Time
	// deliveryTimeout bounds each statement of a sync's delivery
	// transaction (AuditedSync); zero leaves the connection's own timeout.
	deliveryTimeout time.Duration
}

// defaultDeliveryTimeout is how long each statement of a sync's delivery
// transaction may run, unless --sync-delivery-timeout says otherwise.
const defaultDeliveryTimeout = 5 * time.Minute

// newOutbox returns the outbox over pool, whose River tables the valuation
// context's migration made. Its client only enqueues.
func newOutbox(pool *pgxpool.Pool, now func() time.Time) (*outbox, error) {
	c, err := river.NewClient(riverpgxv5.New(pool), &river.Config{Schema: valuationpg.RiverSchema})
	if err != nil {
		return nil, fmt.Errorf("audit outbox: %w", err)
	}
	return &outbox{pool: pool, insert: c, now: now, deliveryTimeout: defaultDeliveryTimeout}, nil
}

// Within runs change in one transaction of the core's, with the context's
// stores joined to it, and enqueues e in the same transaction: both commit or
// neither does. Inside a transaction already, it opens a savepoint instead. It
// implements identityapp.ChangeOutbox.
func (o *outbox) Within(ctx context.Context, e *identity.AuditEntry, change func(context.Context) error) error {
	return o.run(ctx, 0, func(ctx context.Context) (*identity.AuditEntry, error) {
		return e, change(ctx)
	})
}

// Enqueue enqueues e in the transaction ctx carries, which a change's store
// handed over: a model settings change enqueues its entry in the transaction
// that stores it.
func (o *outbox) Enqueue(ctx context.Context, e *identity.AuditEntry) error {
	tx, ok := txOf(ctx)
	if !ok {
		return fmt.Errorf("%w: %w", kernel.ErrAuditUnavailable, ErrNoChangeTx)
	}
	return o.enqueue(ctx, tx.Tx, e)
}

// AuditedSync runs write, a sync's writes, in one transaction and, when it
// reports a change, enqueues the entry of the sync as the system in it
// (ingestapp.SyncAuditor). A sync that changed nothing leaves no entry.
//
// The delivery is all or nothing (D-18), so it is never split to fit a time
// limit: each statement of its transaction has the delivery timeout, and a
// statement that runs past it fails the delivery, which rolls back whole. The
// error wraps ingestapp.ErrDeliveryTimeout, and the sync's attempt is recorded
// as failed with the class timeout (D-198).
func (o *outbox) AuditedSync(ctx context.Context, instance string, write func(context.Context) (ingestapp.SyncChange, error)) error {
	err := o.run(ctx, o.deliveryTimeout, func(ctx context.Context) (*identity.AuditEntry, error) {
		change, err := write(ctx)
		if err != nil || change.Empty() {
			return nil, err
		}
		e, err := syncEntry(o.now(), instance, &change)
		if err != nil {
			return nil, err
		}
		return &e, nil
	})
	if isStatementTimeout(err) {
		return fmt.Errorf("%w: a statement of the delivery ran past %s, so nothing of it was written: %w", ingestapp.ErrDeliveryTimeout, o.deliveryTimeout, err)
	}
	return err
}

// statementTimeoutCode is the SQLSTATE of a statement the database cancelled:
// query_canceled, which a statement timeout raises.
const statementTimeoutCode = "57014"

// isStatementTimeout reports whether err is the database cancelling a
// statement for running past statement_timeout.
func isStatementTimeout(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == statementTimeoutCode && strings.Contains(pgErr.Message, "statement timeout")
}

// run is Within for a change that says which entry it made: none when it
// returns nil. With a positive statementTimeout, each statement of the
// change's own transaction has that timeout (SET LOCAL, so it ends with the
// transaction). What the change handed to afterChange runs once the outermost
// change has committed, outside any transaction.
func (o *outbox) run(ctx context.Context, statementTimeout time.Duration, change func(context.Context) (*identity.AuditEntry, error)) error {
	var begin interface {
		Begin(ctx context.Context) (pgx.Tx, error)
	} = o.pool
	// What the change holds for after it is scoped to its transaction: a
	// savepoint that rolls back discards it, one that commits hands it to the
	// change it is in, and the outermost commit runs it.
	var parent *afterCommit
	if outer, ok := txOf(ctx); ok {
		begin = outer
		if held, isHeld := ctx.Value(afterKey{}).(*afterCommit); isHeld {
			parent = held
		}
	}
	after := &afterCommit{}
	ctx = context.WithValue(ctx, afterKey{}, after)
	err := pgx.BeginFunc(ctx, begin, func(tx pgx.Tx) error {
		if statementTimeout > 0 {
			if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = "+strconv.FormatInt(statementTimeout.Milliseconds(), 10)); err != nil {
				return fmt.Errorf("setting the statement timeout of the transaction: %w", err)
			}
		}
		jctx := joinedPool{}.Join(ctx, tx)
		e, err := change(jctx)
		if err != nil || e == nil {
			return err
		}
		return o.enqueue(jctx, tx, e)
	})
	if err != nil {
		return fmt.Errorf("in one transaction with its audit entry: %w", err)
	}
	if parent != nil {
		parent.adopt(after)
		return nil
	}
	return after.run(ctx)
}

// ErrNetworkInChange reports a network call made with the context of a change's
// transaction. A change holds its transaction open, and its locks, for as long
// as it runs, so no call to a connector, a model provider, SpiceDB or the
// cluster's API is made inside one: the change's caller makes it first and
// hands the answer in, or the change defers it until it has committed
// (afterChange). The clients the core calls over the network refuse such a
// call with this error, so a change that makes one fails in its tests, not
// with a database that is held up (D-211).
var ErrNetworkInChange = errors.New("a network call inside a change's transaction")

// refuseInChange returns ErrNetworkInChange when ctx carries a change's
// transaction, and nil otherwise.
func refuseInChange(ctx context.Context) error {
	if _, ok := txOf(ctx); ok {
		return ErrNetworkInChange
	}
	return nil
}

// noChangeTransport is an http.RoundTripper that refuses a request made with
// the context of a change's transaction (ErrNetworkInChange): the HTTP and
// Connect clients of the core are built over it.
type noChangeTransport struct{ next http.RoundTripper }

// RoundTrip refuses req inside a change, and otherwise sends it.
func (t noChangeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := refuseInChange(req.Context()); err != nil {
		return nil, fmt.Errorf("%s: %w", req.URL.Host, err)
	}
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}
	return resp, nil
}

// refuseUnaryInChange is a gRPC unary client interceptor with the same
// refusal: SpiceDB's client is built with it.
func refuseUnaryInChange(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if err := refuseInChange(ctx); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	return nil
}

// refuseStreamInChange returns the gRPC stream client interceptor with the
// same refusal as refuseUnaryInChange: SpiceDB's lookups are streams.
func refuseStreamInChange() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if err := refuseInChange(ctx); err != nil {
			return nil, fmt.Errorf("%s: %w", method, err)
		}
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", method, err)
		}
		return stream, nil
	}
}

// outsideChange returns ctx as it is, without the transaction it carries: for
// work a change starts that outlives it, such as a sync the owner retries,
// which runs in its own transactions after the change has committed.
func outsideChange(ctx context.Context) context.Context {
	return context.WithValue(ctx, txKey{}, nil)
}

// afterKey is the context key of what runs after the outermost change
// commits.
type afterKey struct{}

// afterCommit holds what the stores and observers a change calls asked to run
// once the change has committed.
type afterCommit struct {
	mu  sync.Mutex
	fns []func(context.Context) error
}

// run runs what was added, in order, with ctx, and joins the errors of those
// that failed. A change whose follow-ups fail is committed all the same: the
// error says so.
func (a *afterCommit) run(ctx context.Context) error {
	a.mu.Lock()
	fns := a.fns
	a.fns = nil
	a.mu.Unlock()
	var errs []error
	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("the change is committed, but what follows it failed: %w", err)
	}
	return nil
}

// adopt takes over what child holds, to run with a.
func (a *afterCommit) adopt(child *afterCommit) {
	child.mu.Lock()
	fns := child.fns
	child.fns = nil
	child.mu.Unlock()
	a.mu.Lock()
	a.fns = append(a.fns, fns...)
	a.mu.Unlock()
}

// afterChange runs fn once the change ctx belongs to has committed, or at once
// when ctx belongs to none. It is for what a change does over the network,
// such as telling SpiceDB of a placement: not inside the change's transaction,
// which it would hold open (ErrNetworkInChange). A change that rolls back
// never runs fn.
func afterChange(ctx context.Context, fn func(context.Context) error) error {
	if _, ok := txOf(ctx); ok {
		if a, ok := ctx.Value(afterKey{}).(*afterCommit); ok {
			a.mu.Lock()
			defer a.mu.Unlock()
			a.fns = append(a.fns, fn)
			return nil
		}
	}
	return fn(ctx)
}

// enqueue inserts the job that writes e, in tx. A failure is an audit log
// that is unavailable: the change that was to share the transaction is not
// made.
func (o *outbox) enqueue(ctx context.Context, tx pgx.Tx, e *identity.AuditEntry) error {
	if err := e.Validate(); err != nil {
		return fmt.Errorf("enqueueing the audit entry of %s: %w", e.Action, err)
	}
	if _, err := o.insert.InsertTx(ctx, tx, auditArgsOf(e), nil); err != nil {
		return fmt.Errorf("%w: enqueueing the audit entry of %s: %w", kernel.ErrAuditUnavailable, e.Action, err)
	}
	return nil
}

// syncEntry returns the entry of a sync that changed state, at at: its
// instance, its batch, the accounts it wrote transactions in and how many
// transaction versions, prices, FX rates, index readings, index levels and
// bond terms it wrote.
func syncEntry(at time.Time, instance string, c *ingestapp.SyncChange) (identity.AuditEntry, error) {
	refs := []identity.Ref{
		{Kind: identity.RefPlugin, ID: instance},
		{Kind: identity.RefImportBatch, ID: string(c.Batch)},
	}
	if len(c.Accounts) <= maxSyncAccountsNamed {
		for _, a := range slices.Clone(c.Accounts) {
			refs = append(refs, identity.Ref{Kind: identity.RefAccount, ID: string(a)})
		}
	}
	for _, n := range []struct {
		kind  string
		count int
	}{
		{"transactions", c.Transactions},
		{"prices", c.Prices},
		{"fx_rates", c.FXRates},
		{"index_readings", c.IndexReadings},
		{"index_levels", c.IndexLevels},
		{"bond_terms", c.BondTerms},
	} {
		if n.count > 0 {
			refs = append(refs, identity.Ref{Kind: identity.RefSyncCount, ID: fmt.Sprintf("%s:%d", n.kind, n.count)})
		}
	}
	system := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	e, err := identity.NewAuditEntry(at, system, identity.ActionConnectorSynced, identity.OutcomeOK, refs...)
	if err != nil {
		return identity.AuditEntry{}, fmt.Errorf("building the audit entry of a sync: %w", err)
	}
	return e, nil
}

// auditStuckAfter is how many failed attempts to write its entry make an audit
// job stuck: the log is behind (D-211, D-196). River keeps retrying it, with
// no limit, whatever this says; it only decides when the owner is told.
const auditStuckAfter = 5

// jobCounter counts jobs of a kind that failed some attempts and are still
// waiting; *valuationpg.Store, which owns River's tables, implements it.
type jobCounter interface {
	JobsFailedAtLeast(ctx context.Context, kind string, failures int) (int, error)
}

// auditBacklog is attentionapp.AuditBacklog over the outbox's job queue: the
// audit entries that have failed auditStuckAfter attempts and still wait.
type auditBacklog struct{ jobs jobCounter }

// StuckEntries returns how many audit entries are stuck.
func (b auditBacklog) StuckEntries(ctx context.Context) (int, error) {
	n, err := b.jobs.JobsFailedAtLeast(ctx, auditEntryArgs{}.Kind(), auditStuckAfter)
	if err != nil {
		return 0, fmt.Errorf("the audit entries that wait: %w", err)
	}
	return n, nil
}
