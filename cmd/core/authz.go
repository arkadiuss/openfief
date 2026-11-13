package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/arkadiuss/openfief/authz"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
)

// The authorizers --authz chooses between (D-163, D-164).
const (
	// authzPolicy is identityapp.Policy: the owner acts on every account,
	// every other holder of read access views every account. The default.
	authzPolicy = "policy"
	// authzSpiceDB is SpiceDB answering over authz/schema.zed.
	authzSpiceDB = "spicedb"
)

const (
	// spicedbEndpointEnv names the environment variable with SpiceDB's gRPC
	// endpoint, host:port, which --spicedb-endpoint overrides.
	spicedbEndpointEnv = "OPENFIEF_SPICEDB_ENDPOINT"
	// spicedbKeyEnv names the environment variable with SpiceDB's preshared
	// key. In the cluster it comes from a Secret (D-29). It is never a flag,
	// so it never shows in a process listing, and never logged.
	spicedbKeyEnv = "OPENFIEF_SPICEDB_PRESHARED_KEY"
	// defaultSpiceDBKeyFile is where the core's Deployment mounts the key of
	// the Secret openfief-spicedb-key, as an optional volume: the file
	// appears when the credentials Job has created the Secret (D-216), and
	// the core, which starts before it, waits for it.
	defaultSpiceDBKeyFile = "/var/run/openfief/spicedb/key"
	// keyPollEvery is how often a core without the key looks for the file.
	keyPollEvery = 2 * time.Second
	// defaultSpiceDBTimeout bounds one call to SpiceDB.
	defaultSpiceDBTimeout = spicedb.DefaultTimeout
	// authzRetryEvery is how often the start retries a reconcile SpiceDB
	// could not take.
	authzRetryEvery = 10 * time.Second
)

var (
	// errBadAuthz reports an --authz that is neither policy nor spicedb.
	errBadAuthz = errors.New("--authz must be policy or spicedb")
	// errNoSpiceDBEndpoint reports --authz=spicedb without an endpoint.
	errNoSpiceDBEndpoint = errors.New("--authz=spicedb needs --spicedb-endpoint or " + spicedbEndpointEnv)
	// errBadSpiceDBTimeout reports a --spicedb-timeout of zero or less.
	errBadSpiceDBTimeout = errors.New("--spicedb-timeout must be positive")
	// errNoSpiceDBKey reports --authz=spicedb without a preshared key, from
	// the environment or a file, to a command that can't wait for one.
	errNoSpiceDBKey = errors.New("--authz=spicedb needs the preshared key in " + spicedbKeyEnv + " or in --spicedb-key-file")
	// errSpiceDBKeyPending is what every check answers until the core has
	// SpiceDB's key: refused, closed (D-164).
	errSpiceDBKeyPending = fmt.Errorf("%w: SpiceDB's preshared key is not there yet; every check is refused until it is", kernel.ErrAuthzUnavailable)
)

// secret is a credential that prints as nothing: no format verb, log
// attribute or dump of a config shows its value.
type secret string

// String returns a placeholder.
func (secret) String() string { return "[redacted]" }

// GoString returns a placeholder.
func (secret) GoString() string { return "[redacted]" }

// LogValue returns a placeholder for slog.
func (secret) LogValue() slog.Value { return slog.StringValue("[redacted]") }

// authzConfig is serve's authorization configuration.
type authzConfig struct {
	// Mode is authzPolicy, or authzSpiceDB; empty is authzPolicy.
	Mode string
	// Endpoint is SpiceDB's gRPC address, host:port.
	Endpoint string
	// Key is SpiceDB's preshared key, when the environment gave it; no
	// encoder writes it. Without it serve reads KeyFile, and waits for it.
	Key secret `json:"-"`
	// KeyFile is the file with the preshared key, mounted from the Secret
	// as an optional volume. The environment's key, when set, wins.
	KeyFile string
	// Insecure speaks plaintext gRPC, for the in-cluster Service behind its
	// network policy; otherwise the connection is TLS, verified against the
	// system's certificate authorities.
	Insecure bool
	// Timeout bounds one call to SpiceDB; a check that takes longer is
	// refused.
	Timeout time.Duration
}

// authzFlags defines serve's authorization flags on fs, into a.
func authzFlags(fs *flag.FlagSet, a *authzConfig) {
	fs.StringVar(&a.Mode, "authz", authzPolicy, "who answers authorization (D-163, D-164): policy, the owner acts on every account and every other holder of read access views every account, or spicedb")
	fs.StringVar(&a.Endpoint, "spicedb-endpoint", "", "SpiceDB's gRPC address, host:port, for --authz=spicedb; "+spicedbEndpointEnv+" by default")
	fs.StringVar(&a.KeyFile, "spicedb-key-file", defaultSpiceDBKeyFile, "the file with SpiceDB's preshared key, for --authz=spicedb, which "+spicedbKeyEnv+" overrides. serve starts without the file and refuses every check, reporting not ready, until it has a key")
	fs.BoolVar(&a.Insecure, "spicedb-insecure", false, "speak plaintext gRPC to SpiceDB, for the in-cluster Service behind its network policy; TLS otherwise")
	fs.DurationVar(&a.Timeout, "spicedb-timeout", defaultSpiceDBTimeout, "how long one call to SpiceDB may take; a check that takes longer is refused")
}

// parseAuthz checks a and reads SpiceDB's endpoint, when the flag left it
// empty, and its preshared key, when the environment has it, through getenv.
// The key may still be missing: serve waits for the file, and a command asks
// requireKey. With --authz=policy nothing of SpiceDB is read or needed.
func parseAuthz(a *authzConfig, getenv func(string) string) error {
	switch a.Mode {
	case authzPolicy:
		return nil
	case authzSpiceDB:
	default:
		return fmt.Errorf("%w, not %q", errBadAuthz, a.Mode)
	}
	if a.Endpoint == "" {
		a.Endpoint = getenv(spicedbEndpointEnv)
	}
	if a.Endpoint == "" {
		return errNoSpiceDBEndpoint
	}
	a.Key = secret(strings.TrimSpace(getenv(spicedbKeyEnv)))
	if a.Key == "" && a.KeyFile == "" {
		return errNoSpiceDBKey
	}
	if a.Timeout <= 0 {
		return errBadSpiceDBTimeout
	}
	return nil
}

// readKeyFile returns the preshared key in path, trimmed, or "" when the file
// is not there yet, can't be read or holds only whitespace: in every such
// case the key is still to come. The value is returned to the caller alone.
func readKeyFile(path string) secret {
	// The file is a symbolic link into the directory the kubelet projects the Secret to, which a root keeps within it.
	dir, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return ""
	}
	defer dir.Close() //nolint:errcheck // a read-only directory
	b, err := dir.ReadFile(filepath.Base(path))
	if err != nil {
		return ""
	}
	return secret(strings.TrimSpace(string(b)))
}

// requireKey gives a, for a command that can't wait, the key of its file when
// the environment gave none, or errNoSpiceDBKey. It does nothing with the
// policy.
func (a *authzConfig) requireKey() error {
	if a.Mode != authzSpiceDB || a.Key != "" {
		return nil
	}
	if a.Key = readKeyFile(a.KeyFile); a.Key == "" {
		return errNoSpiceDBKey
	}
	return nil
}

// authorization is what --authz wires: the authorizer every read path and
// owner service asks (D-163), and, with SpiceDB, the syncer that keeps its
// relationships to the registry's accounts and sleeves.
type authorization struct {
	authorizer identityapp.Authorizer
	// syncer is nil with the policy.
	syncer accessSyncer
	// closer closes the connection to SpiceDB; nil with the policy.
	closer io.Closer
	// gate is non-nil while the core may still be waiting for SpiceDB's key
	// (D-164): serve without the key in the environment.
	gate *keyGate
	// ping asks SpiceDB whether it answers; nil with the policy.
	ping func(context.Context) error
	// started is set once SpiceDB has answered and holds the schema and the
	// relationships the core owns: the start's reconcile succeeded. Readiness
	// waits for it, and it is never unset (an outage later is the monitor's);
	// nil with the policy.
	started *atomic.Bool
}

// waiting says what the core still waits for before it is ready, or "" when
// nothing: with SpiceDB, its key, then an answer with the schema written.
// It only ever turns from waiting to "", so a SpiceDB outage after the start
// never makes the core not ready.
func (az authorization) waiting() string {
	if az.gate != nil && !az.gate.isConnected() {
		return "waiting for SpiceDB's preshared key"
	}
	if az.started != nil && !az.started.Load() {
		return "waiting for SpiceDB to answer and take its schema"
	}
	return ""
}

// accessSyncer is what the core asks of the syncer that keeps SpiceDB's
// relationships: *spicedb.Syncer, or the gate's, which refuses until the key
// is there.
type accessSyncer interface {
	identityapp.GrantProjector
	AccountPlaced(ctx context.Context, account, sleeve string) error
	AccountUnplaced(ctx context.Context, account string) error
	Reconcile(ctx context.Context, load func(context.Context) (spicedb.Desired, error)) error
}

// newAuthorization returns the authorizer cfg chooses, for a command, which
// has its key. With SpiceDB it dials lazily: an unreachable SpiceDB fails
// each check, closed, not the start.
func newAuthorization(cfg *authzConfig) (authorization, error) {
	if cfg.Mode != authzSpiceDB {
		return authorization{authorizer: identityapp.Policy{}}, nil
	}
	c, err := dialSpiceDB(cfg)
	if err != nil {
		return authorization{}, err
	}
	return authorization{authorizer: c.authorizer, syncer: c.syncer, closer: c.client, ping: c.authorizer.Ping, started: new(atomic.Bool)}, nil
}

// serveAuthorization returns the authorization serve runs with. With SpiceDB
// and no key in the environment it is the gate's, which waits for the key's
// file in the background (D-164, D-216) while every check is refused; with the
// key in the environment, or with the policy, it is newAuthorization's. With
// SpiceDB the authorizer is watched: a change of its availability is logged
// once and counted (authzMonitor).
func serveAuthorization(ctx context.Context, cfg *authzConfig, log *slog.Logger) (authorization, error) {
	if cfg.Mode != authzSpiceDB {
		return newAuthorization(cfg)
	}
	var az authorization
	if cfg.Key != "" {
		var err error
		if az, err = newAuthorization(cfg); err != nil {
			return authorization{}, err
		}
	} else {
		g := newKeyGate(cfg, keyPollEvery, log)
		go g.run(ctx)
		az = gatedAuthorization(g)
	}
	m, err := newAuthzMonitor(ctx, log, otel.Meter(coreService), az.ping, authzProbeEvery)
	if err != nil {
		return authorization{}, err
	}
	return m.watch(az), nil
}

// spicedbConn is a client of SpiceDB and what is built on it.
type spicedbConn struct {
	client     *authzed.Client
	authorizer *spicedb.Authorizer
	syncer     *spicedb.Syncer
}

// dialSpiceDB connects, lazily, to cfg's SpiceDB with cfg's key.
func dialSpiceDB(cfg *authzConfig) (spicedbConn, error) {
	var opts []grpc.DialOption
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()), grpcutil.WithInsecureBearerToken(string(cfg.Key)))
	} else {
		tls, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
		if err != nil {
			return spicedbConn{}, fmt.Errorf("SpiceDB certificates: %w", err)
		}
		opts = append(opts, tls, grpcutil.WithBearerToken(string(cfg.Key)))
	}
	client, err := authzed.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return spicedbConn{}, fmt.Errorf("SpiceDB client for %s: %w", cfg.Endpoint, err)
	}
	return spicedbConn{
		client:     client,
		authorizer: spicedb.NewAuthorizer(client.PermissionsServiceClient, spicedb.WithTimeout(cfg.Timeout)),
		// The syncer also writes the model this build was written against at every start, so a new
		// SpiceDB, or an upgraded core, needs no step of its own.
		syncer: spicedb.NewSyncer(client.PermissionsServiceClient, spicedb.WithTimeout(cfg.Timeout),
			spicedb.WithSchema(client.SchemaServiceClient, authz.Schema)),
	}, nil
}

// placementSync is portfolioapp.PlacementObserver over the syncer: it
// translates the registry's IDs into the syncer's.
type placementSync struct{ syncer accessSyncer }

var _ portfolioapp.PlacementObserver = placementSync{}

func (p placementSync) AccountPlaced(ctx context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error {
	if err := p.syncer.AccountPlaced(ctx, string(account), string(sleeve)); err != nil {
		return fmt.Errorf("passing on a placement: %w", err)
	}
	return nil
}

func (p placementSync) AccountRemoved(ctx context.Context, account ledger.AccountID) error {
	if err := p.syncer.AccountUnplaced(ctx, string(account)); err != nil {
		return fmt.Errorf("passing on a removal: %w", err)
	}
	return nil
}

// SleeveRemoved does nothing: only an empty sleeve is deleted, so no account
// has a relationship to it, and its grants are the grants' to end.
func (placementSync) SleeveRemoved(context.Context, portfolio.SleeveID) error { return nil }

// registryOptions returns the options of the registry az needs: with
// SpiceDB, the placement observer.
func (az authorization) registryOptions() []portfolioapp.RegistryOption {
	if az.syncer == nil {
		return nil
	}
	return []portfolioapp.RegistryOption{portfolioapp.WithPlacementObserver(placementSync{az.syncer})}
}

// startReconcile starts the reconcile of SpiceDB with what load says the core
// owns, and returns the function that runs it once more, for the grants'
// start. With the key to come, nothing is changed until it is there: the
// reconcile runs then, in the background.
func (az authorization) startReconcile(ctx context.Context, load func(context.Context) (spicedb.Desired, error), log *slog.Logger) func(context.Context) error {
	if az.gate == nil {
		reconcileAccess(ctx, az.syncer, load, log, az.markStarted)
	} else {
		go func() {
			if az.gate.wait(ctx) {
				reconcileAccess(ctx, az.syncer, load, log, az.markStarted)
			}
		}()
	}
	return func(ctx context.Context) error {
		if err := az.syncer.Reconcile(ctx, load); err != nil {
			return fmt.Errorf("reconciling: %w", err)
		}
		return nil
	}
}

// markStarted records that the start's reconcile succeeded.
func (az authorization) markStarted() {
	if az.started != nil {
		az.started.Store(true)
	}
}

// desiredAccess returns the loader of what SpiceDB's relationships the core
// owns must be: every account the ledger holds, the sleeve each is in,
// whether demo sessions exist and the grants in force
// (spicedb.Syncer.Reconcile).
func desiredAccess(l *ledgerapp.Ledger, sleeves *portfolioapp.Registry, grants *identityapp.Grants, demo bool) func(context.Context) (spicedb.Desired, error) {
	return func(ctx context.Context) (spicedb.Desired, error) {
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return spicedb.Desired{}, fmt.Errorf("loading accounts: %w", err)
		}
		in, err := sleeves.AccountSleeves(ctx)
		if err != nil {
			return spicedb.Desired{}, fmt.Errorf("loading the accounts' sleeves: %w", err)
		}
		inForce, err := grants.InForce(ctx)
		if err != nil {
			return spicedb.Desired{}, fmt.Errorf("loading the grants: %w", err)
		}
		d := spicedb.Desired{Sleeves: make(map[string]string, len(in)), DemoSessions: demo, Grants: inForce}
		for i := range accounts {
			d.Accounts = append(d.Accounts, string(accounts[i].ID))
		}
		for a, s := range in {
			d.Sleeves[string(a)] = string(s)
		}
		return d, nil
	}
}

// reconcileAccess makes SpiceDB hold the relationships the core owns, once
// at start and, when SpiceDB can't take it yet, again every authzRetryEvery
// in the background until it does or ctx ends. Until then every check is
// refused, closed, since no account has a household. A failure is logged,
// without any credential. done, when not nil, is called once the reconcile
// succeeded: SpiceDB has answered and holds the schema.
func reconcileAccess(ctx context.Context, syncer accessSyncer, load func(context.Context) (spicedb.Desired, error), log *slog.Logger, done func()) {
	if syncer == nil {
		return
	}
	if done == nil {
		done = func() {}
	}
	err := syncer.Reconcile(ctx, load)
	if err == nil {
		done()
		log.Info("authorization reconciled with SpiceDB")
		return
	}
	log.Warn("SpiceDB is not reconciled yet; every check is refused until it is, retrying", "error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded, spicedb.ErrInvalidID))
	go func() {
		t := time.NewTicker(authzRetryEvery)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			if err := syncer.Reconcile(ctx, load); err != nil {
				log.Warn("SpiceDB is not reconciled yet; retrying", "error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded, spicedb.ErrInvalidID))
				continue
			}
			done()
			log.Info("authorization reconciled with SpiceDB")
			return
		}
	}()
}
