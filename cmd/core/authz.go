package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
	// errNoSpiceDBKey reports --authz=spicedb without a preshared key.
	errNoSpiceDBKey = errors.New("--authz=spicedb needs the preshared key in " + spicedbKeyEnv)
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
	// Key is SpiceDB's preshared key; no encoder writes it.
	Key secret `json:"-"`
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
	fs.StringVar(&a.Endpoint, "spicedb-endpoint", "", "SpiceDB's gRPC address, host:port, for --authz=spicedb; "+spicedbEndpointEnv+" by default. The preshared key is read from "+spicedbKeyEnv)
	fs.BoolVar(&a.Insecure, "spicedb-insecure", false, "speak plaintext gRPC to SpiceDB, for the in-cluster Service behind its network policy; TLS otherwise")
	fs.DurationVar(&a.Timeout, "spicedb-timeout", defaultSpiceDBTimeout, "how long one call to SpiceDB may take; a check that takes longer is refused")
}

// parseAuthz checks a and reads SpiceDB's endpoint, when the flag left it
// empty, and its preshared key through getenv. With --authz=policy nothing
// of SpiceDB is read or needed.
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
	if a.Key = secret(getenv(spicedbKeyEnv)); a.Key == "" {
		return errNoSpiceDBKey
	}
	if a.Timeout <= 0 {
		return errBadSpiceDBTimeout
	}
	return nil
}

// authorization is what --authz wires: the authorizer every read path and
// owner service asks (D-163), and, with SpiceDB, the syncer that keeps its
// relationships to the registry's accounts and sleeves.
type authorization struct {
	authorizer identityapp.Authorizer
	// syncer is nil with the policy.
	syncer *spicedb.Syncer
	// closer closes the connection to SpiceDB; nil with the policy.
	closer io.Closer
}

// newAuthorization returns the authorizer cfg chooses. With SpiceDB it
// dials lazily: an unreachable SpiceDB fails each check, closed, not the
// start.
func newAuthorization(cfg *authzConfig) (authorization, error) {
	if cfg.Mode != authzSpiceDB {
		return authorization{authorizer: identityapp.Policy{}}, nil
	}
	var opts []grpc.DialOption
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()), grpcutil.WithInsecureBearerToken(string(cfg.Key)))
	} else {
		tls, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
		if err != nil {
			return authorization{}, fmt.Errorf("SpiceDB certificates: %w", err)
		}
		opts = append(opts, tls, grpcutil.WithBearerToken(string(cfg.Key)))
	}
	client, err := authzed.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return authorization{}, fmt.Errorf("SpiceDB client for %s: %w", cfg.Endpoint, err)
	}
	return authorization{
		authorizer: spicedb.NewAuthorizer(client.PermissionsServiceClient, spicedb.WithTimeout(cfg.Timeout)),
		syncer:     spicedb.NewSyncer(client.PermissionsServiceClient, spicedb.WithTimeout(cfg.Timeout)),
		closer:     client,
	}, nil
}

// placementSync is portfolioapp.PlacementObserver over the syncer: it
// translates the registry's IDs into the syncer's.
type placementSync struct{ syncer *spicedb.Syncer }

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

// registryOptions returns the options of the registry az needs: with
// SpiceDB, the placement observer.
func (az authorization) registryOptions() []portfolioapp.RegistryOption {
	if az.syncer == nil {
		return nil
	}
	return []portfolioapp.RegistryOption{portfolioapp.WithPlacementObserver(placementSync{az.syncer})}
}

// desiredAccess returns the loader of what SpiceDB's relationships the core
// owns must be: every account the ledger holds, the sleeve each is in, and
// whether demo sessions exist (spicedb.Syncer.Reconcile).
func desiredAccess(l *ledgerapp.Ledger, sleeves *portfolioapp.Registry, demo bool) func(context.Context) (spicedb.Desired, error) {
	return func(ctx context.Context) (spicedb.Desired, error) {
		accounts, err := l.Accounts(ctx)
		if err != nil {
			return spicedb.Desired{}, fmt.Errorf("loading accounts: %w", err)
		}
		in, err := sleeves.AccountSleeves(ctx)
		if err != nil {
			return spicedb.Desired{}, fmt.Errorf("loading the accounts' sleeves: %w", err)
		}
		d := spicedb.Desired{Sleeves: make(map[string]string, len(in)), DemoSessions: demo}
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
// without any credential.
func reconcileAccess(ctx context.Context, syncer *spicedb.Syncer, load func(context.Context) (spicedb.Desired, error), log *slog.Logger) {
	if syncer == nil {
		return
	}
	err := syncer.Reconcile(ctx, load)
	if err == nil {
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
			log.Info("authorization reconciled with SpiceDB")
			return
		}
	}()
}
