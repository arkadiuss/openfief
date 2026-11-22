package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// What a core under SpiceDB says while it is not ready (D-164): an owner who
// upgraded without SpiceDB's database objects (install step 5), or whose
// Secret is missing, reads in /readyz and in the log exactly what is missing,
// by name, never a bare "not ready".
const (
	// initJobLabel is the label of the install's Job that creates
	// SpiceDB's Secrets (deploy/spicedb/secrets.yaml, D-216).
	initJobLabel = "app.kubernetes.io/name=openfief-init-credentials"
	// spicedbDatabaseName is SpiceDB's own database in the openfief-db
	// cluster, which its role owns (deploy/database/spicedb-database.yaml).
	spicedbDatabaseName = "spicedb"
	// schemaPending is what a core waits for once it has the key: SpiceDB
	// answering, with its schema written.
	schemaPending = "waiting for SpiceDB to answer and take its schema"
	// storeCheckTimeout bounds the database query a diagnosis makes.
	storeCheckTimeout = 2 * time.Second
	// demoPollEvery is how often a demo core reads its accounts to see
	// whether the demo's loader changed them (followDemo).
	demoPollEvery = 15 * time.Second
	// demoReconcileEvery is how often a demo core reconciles SpiceDB again
	// whatever it read, as a backstop (followDemo).
	demoReconcileEvery = 5 * time.Minute
)

// keyPendingReason says that the core waits for SpiceDB's preshared key, in
// file, naming the Secret it comes from and what creates it.
func keyPendingReason(file string) string {
	return fmt.Sprintf("waiting for SpiceDB's preshared key: the Secret %s is not in %s yet; "+
		"the install's credentials Job (label %s) creates it, and every authorization check is refused until it is there",
		identity.PresharedKeyName, file, initJobLabel)
}

// schemaPendingReason says that the core waits for SpiceDB, with why's
// reason when an attempt has failed.
func schemaPendingReason(why *atomic.Pointer[string]) string {
	if why != nil {
		if r := why.Load(); r != nil && *r != "" {
			return schemaPending + ": " + *r
		}
	}
	return schemaPending
}

// spicedbStore is what the core's database says of SpiceDB's own role and
// database in the same cluster.
type spicedbStore struct {
	// role and database are whether the role spicedb and the database
	// spicedb exist.
	role, database bool
}

// rowQuerier is the query the diagnosis runs: *pgxpool.Pool has it.
type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// readSpiceDBStore asks the database whether SpiceDB's role and database
// exist. Every role may read pg_roles and pg_database, so the core's own
// role asks without a right of its own, and reads nothing of SpiceDB's.
func readSpiceDBStore(ctx context.Context, q rowQuerier) (spicedbStore, error) {
	ctx, cancel := context.WithTimeout(ctx, storeCheckTimeout)
	defer cancel()
	var s spicedbStore
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = $1),
		EXISTS (SELECT FROM pg_catalog.pg_database WHERE datname = $2)`,
		identity.RoleUser, spicedbDatabaseName).Scan(&s.role, &s.database)
	if err != nil {
		return spicedbStore{}, fmt.Errorf("asking for SpiceDB's role and database: %w", err)
	}
	return s, nil
}

// spicedbWaitReason says why a reconcile of SpiceDB at endpoint failed with
// a call that ended with code, given what the database said of SpiceDB's
// role and database (storeErr when it could not be asked): what is missing,
// by name, and where it comes from. It never quotes the error.
func spicedbWaitReason(endpoint string, code codes.Code, store spicedbStore, storeErr error) string {
	if storeErr == nil && (!store.role || !store.database) {
		var missing []string
		if !store.role {
			missing = append(missing, "no role "+identity.RoleUser)
		}
		if !store.database {
			missing = append(missing, "no database "+spicedbDatabaseName)
		}
		return fmt.Sprintf("the openfief-db cluster has %s yet: add the managed role %s and the Database %s of the release's "+
			"database file to yours and apply it (install step 5); CloudNativePG creates the role once the Secret %s exists",
			strings.Join(missing, " and "), identity.RoleUser, spicedbDatabaseName, identity.RoleAuthName)
	}
	if code == codes.Unauthenticated || code == codes.PermissionDenied {
		return fmt.Sprintf("SpiceDB at %s refuses the core's preshared key: SpiceDB and the core read different keys of the Secret %s; "+
			"restart SpiceDB, then the core", endpoint, identity.PresharedKeyName)
	}
	if storeErr != nil {
		return fmt.Sprintf("SpiceDB at %s has not answered, and the database could not be asked whether it has SpiceDB's role %s and database %s",
			endpoint, identity.RoleUser, spicedbDatabaseName)
	}
	return fmt.Sprintf("SpiceDB at %s has not answered; the openfief-db cluster has its role and database %s, "+
		"so check that the Deployment openfief-spicedb is ready and that the Secret %s exists",
		endpoint, spicedbDatabaseName, identity.RoleAuthName)
}

// spicedbDiagnosis returns the authorization's diagnose for SpiceDB at
// endpoint over the core's database q: it asks the database for SpiceDB's
// role and database, and reads the gRPC code of the failed call.
func spicedbDiagnosis(q rowQuerier, endpoint string) func(context.Context, error) string {
	return func(ctx context.Context, err error) string {
		store, storeErr := readSpiceDBStore(ctx, q)
		if errors.Is(err, context.Canceled) && ctx.Err() != nil {
			return "the core is shutting down"
		}
		return spicedbWaitReason(endpoint, status.Code(err), store, storeErr)
	}
}

// followDemo keeps a demo core's SpiceDB in step with the accounts the
// demo's loader (cmd/demo-seed) writes. The loader drops and loads the ledger
// and the portfolio again next to a running core, which learns of no account
// it writes, so a start's reconcile that ran in between left SpiceDB without
// them and every demo session viewing nothing. Once the start's reconcile has
// succeeded, followDemo reads what SpiceDB must hold every poll and
// reconciles when the accounts or their sleeves differ from those it last
// reconciled, skipping an empty ledger, which on a demo is only the moment
// between the loader's drop and its accounts; and it reconciles every sweep
// whatever it read, as a backstop. A failure is logged as a warning without
// any credential; it stops when ctx ends.
func (az authorization) followDemo(ctx context.Context, load func(context.Context) (spicedb.Desired, error),
	reconcile func(context.Context) error, poll, sweep time.Duration, log *slog.Logger,
) {
	pollT, sweepT := time.NewTicker(poll), time.NewTicker(sweep)
	defer pollT.Stop()
	defer sweepT.Stop()
	var last string
	for {
		force := false
		select {
		case <-ctx.Done():
			return
		case <-pollT.C:
		case <-sweepT.C:
			force = true
		}
		last = az.demoStep(ctx, load, reconcile, last, force, log)
	}
}

// demoStep is one look of followDemo: it returns the placements last
// reconciled, which are what it read when it reconciled now.
func (az authorization) demoStep(ctx context.Context, load func(context.Context) (spicedb.Desired, error),
	reconcile func(context.Context) error, last string, force bool, log *slog.Logger,
) string {
	if az.start != nil && !az.start.done.Load() {
		return last // the start's reconcile is still retrying on its own
	}
	d, err := load(ctx)
	if err != nil || len(d.Accounts) == 0 {
		return last // the loader is between its drop and its accounts; the next poll sees them
	}
	placed := placements(&d)
	if placed == last && !force {
		return last
	}
	if err := reconcile(ctx); err != nil {
		log.Warn("SpiceDB's reconcile of the demo's accounts failed; retrying at the next poll",
			"error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded, spicedb.ErrInvalidID))
		return last
	}
	return placed
}

// placements returns d's accounts and the sleeve each is in, in one string
// that is equal for equal sets, whatever their order.
func placements(d *spicedb.Desired) string {
	lines := make([]string, 0, len(d.Accounts))
	for _, a := range d.Accounts {
		lines = append(lines, a+"\x00"+d.Sleeves[a])
	}
	slices.Sort(lines)
	return strings.Join(lines, "\n")
}
