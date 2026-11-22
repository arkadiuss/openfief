package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// keyGate holds the core's side of SpiceDB while its preshared key may still
// be missing (D-164, D-216). On a fresh install the core starts before the
// credentials Job creates the Secret that holds the key, and a core that
// exited for it would crash-loop with a growing back-off. Instead it starts,
// serves its health endpoints, refuses every check, closed, reports not
// ready, and polls the key's file, which the kubelet projects once the
// Secret exists, until the key is there. Then it connects, and everything
// that waited on connected goes on.
type keyGate struct {
	cfg  authzConfig
	poll time.Duration
	log  *slog.Logger
	conn atomic.Pointer[spicedbConn]
	// misses counts the looks that found no key; only step reads it, to log
	// the wait at the first and again every keyRemindEvery.
	misses int
	// connected is closed when the key is read and the client made.
	connected chan struct{}
}

// keyRemindEvery is how often a gate still without the key logs that it
// waits, so a Secret that never comes shows in the log, not only at start.
const keyRemindEvery = time.Minute

// newKeyGate returns the gate for cfg, which has no key of its own. Call run
// to start it waiting.
func newKeyGate(cfg *authzConfig, poll time.Duration, log *slog.Logger) *keyGate {
	return &keyGate{cfg: *cfg, poll: poll, log: log, connected: make(chan struct{})}
}

// Close closes the connection to SpiceDB, when there is one.
func (g *keyGate) Close() error {
	if c := g.conn.Load(); c != nil {
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("closing the connection to SpiceDB: %w", err)
		}
	}
	return nil
}

// run polls the key's file every g.poll until it holds a key or ctx ends,
// then connects.
func (g *keyGate) run(ctx context.Context) {
	t := time.NewTicker(g.poll)
	defer t.Stop()
	for !g.step() {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// step looks at the key's file once and, when it holds a key, connects and
// returns true. It logs that it waits, naming the Secret and the file, at the
// first look and then every keyRemindEvery, as a warning, and never the key.
func (g *keyGate) step() bool {
	key := readKeyFile(g.cfg.KeyFile)
	if key == "" {
		if g.misses%remindAfter(g.poll) == 0 {
			g.log.Warn(keyPendingReason(g.cfg.KeyFile), "secret", identity.PresharedKeyName, "file", g.cfg.KeyFile)
		}
		g.misses++
		return false
	}
	cfg := g.cfg
	cfg.Key = key
	c, err := dialSpiceDB(&cfg)
	if err != nil {
		g.log.Warn("SpiceDB's client could not be made; retrying", "error", kernel.LogSafe(err))
		return false
	}
	g.conn.Store(&c)
	close(g.connected)
	g.log.Info("SpiceDB's preshared key is there; connected")
	return true
}

// remindAfter returns how many looks, poll apart, make keyRemindEvery: at
// least one.
func remindAfter(poll time.Duration) int {
	if poll <= 0 || poll >= keyRemindEvery {
		return 1
	}
	return int(keyRemindEvery / poll)
}

// wait returns true once the gate is connected, or false when ctx ends first.
func (g *keyGate) wait(ctx context.Context) bool {
	select {
	case <-g.connected:
		return true
	case <-ctx.Done():
		return false
	}
}

// isConnected reports whether the key is there and the client made.
func (g *keyGate) isConnected() bool { return g.conn.Load() != nil }

// gatedAuthorization returns the authorization of a serve whose key may be
// missing: the gate's authorizer and syncer, which refuse until the gate has
// connected, then answer as SpiceDB's.
func gatedAuthorization(g *keyGate) authorization {
	return authorization{
		authorizer: gatedAuthorizer{g}, syncer: gatedSyncer{g}, closer: g, gate: g, ping: gatedAuthorizer{g}.Ping,
		start: new(startState),
	}
}

// gatedAuthorizer is identityapp.Authorizer and AccountLister over the gate.
type gatedAuthorizer struct{ g *keyGate }

var (
	_ identityapp.Authorizer    = gatedAuthorizer{}
	_ identityapp.AccountLister = gatedAuthorizer{}
)

func (a gatedAuthorizer) MayAct(ctx context.Context, c identity.Caller, act identity.Action, account string) (bool, error) {
	conn := a.g.conn.Load()
	if conn == nil {
		return false, errSpiceDBKeyPending
	}
	ok, err := conn.authorizer.MayAct(ctx, c, act, account)
	if err != nil {
		return false, fmt.Errorf("SpiceDB's check: %w", err)
	}
	return ok, nil
}

// Ping asks SpiceDB whether it answers: not before the key is there.
func (a gatedAuthorizer) Ping(ctx context.Context) error {
	conn := a.g.conn.Load()
	if conn == nil {
		return errSpiceDBKeyPending
	}
	if err := conn.authorizer.Ping(ctx); err != nil {
		return fmt.Errorf("SpiceDB's ping: %w", err)
	}
	return nil
}

func (a gatedAuthorizer) ActableAccounts(ctx context.Context, c identity.Caller, act identity.Action) (map[string]struct{}, error) {
	conn := a.g.conn.Load()
	if conn == nil {
		return nil, errSpiceDBKeyPending
	}
	accounts, err := conn.authorizer.ActableAccounts(ctx, c, act)
	if err != nil {
		return nil, fmt.Errorf("SpiceDB's lookup: %w", err)
	}
	return accounts, nil
}

// gatedSyncer is accessSyncer over the gate: until it has connected every
// change is refused with errSpiceDBKeyPending, and the reconcile that follows
// the connection makes SpiceDB hold all the core owns.
type gatedSyncer struct{ g *keyGate }

var _ accessSyncer = gatedSyncer{}

func (s gatedSyncer) AccountPlaced(ctx context.Context, account, sleeve string) error {
	sy, err := s.syncer()
	if err != nil {
		return err
	}
	if err := sy.AccountPlaced(ctx, account, sleeve); err != nil {
		return fmt.Errorf("SpiceDB's syncer: %w", err)
	}
	return nil
}

func (s gatedSyncer) AccountUnplaced(ctx context.Context, account string) error {
	sy, err := s.syncer()
	if err != nil {
		return err
	}
	if err := sy.AccountUnplaced(ctx, account); err != nil {
		return fmt.Errorf("SpiceDB's syncer: %w", err)
	}
	return nil
}

func (s gatedSyncer) ProjectGrants(ctx context.Context, grants []identity.Grant) error {
	sy, err := s.syncer()
	if err != nil {
		return err
	}
	if err := sy.ProjectGrants(ctx, grants); err != nil {
		return fmt.Errorf("SpiceDB's syncer: %w", err)
	}
	return nil
}

func (s gatedSyncer) UnprojectGrants(ctx context.Context, grants []identity.Grant) error {
	sy, err := s.syncer()
	if err != nil {
		return err
	}
	if err := sy.UnprojectGrants(ctx, grants); err != nil {
		return fmt.Errorf("SpiceDB's syncer: %w", err)
	}
	return nil
}

func (s gatedSyncer) Reconcile(ctx context.Context, load func(context.Context) (spicedb.Desired, error)) error {
	sy, err := s.syncer()
	if err != nil {
		return err
	}
	if err := sy.Reconcile(ctx, load); err != nil {
		return fmt.Errorf("SpiceDB's syncer: %w", err)
	}
	return nil
}

// syncer returns the connected syncer, or errSpiceDBKeyPending.
func (s gatedSyncer) syncer() (*spicedb.Syncer, error) {
	conn := s.g.conn.Load()
	if conn == nil {
		return nil, errSpiceDBKeyPending
	}
	return conn.syncer, nil
}

var _ io.Closer = (*keyGate)(nil)
