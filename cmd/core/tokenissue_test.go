package main

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identitypg "github.com/arkadiuss/openfief/internal/identity/postgres"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// flakyWrites is a SpiceDB that takes every write unless failWrites is set,
// recording what it took, as writeLog does.
type flakyWrites struct {
	writeLog

	failWrites atomic.Bool
}

func (f *flakyWrites) WriteRelationships(ctx context.Context, in *v1.WriteRelationshipsRequest, opts ...grpc.CallOption) (*v1.WriteRelationshipsResponse, error) {
	if f.failWrites.Load() {
		return nil, errNoAnswer
	}
	return f.writeLog.WriteRelationships(ctx, in, opts...)
}

var errNoAnswer = errors.New("connection refused")

// issuer returns the token issue of the fixture's database as under SpiceDB:
// grants written to perms, ping asking whether it answers, waiting at most
// wait between polls of a millisecond.
func (f *grantsFixture) issuer(perms spicedb.Permissions, ping func(context.Context) error, wait time.Duration) tokenAdmin {
	clock := func() time.Time { return f.now }
	store := identitypg.New(f.pool)
	toks := identityapp.NewTokens(store, clock, nil)
	return tokenAdmin{
		store: store, tokens: toks, grants: identityapp.NewGrants(store, toks, householdOnly{}, spicedb.NewSyncer(perms), clock),
		ping: ping, wait: wait, pollEvery: time.Millisecond,
	}
}

func up(context.Context) error   { return nil }
func down(context.Context) error { return errNoAnswer }

// TestTokenIssueWaitsForSpiceDBAndIssuesNothingWithoutIt checks that, with
// the household grant under SpiceDB, token issue first waits for SpiceDB to
// answer and, when it never does, stores nothing: no token, no grant, no audit
// entry, no write, no secret, and an error that says so.
func TestTokenIssueWaitsForSpiceDBAndIssuesNothingWithoutIt(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	perms := &flakyWrites{}
	var out bytes.Buffer
	cmd := &command{Name: cmdTokenIssue, TokenName: "nightly", TTL: time.Hour, GrantHousehold: true}
	err := tokenCommand(t.Context(), cmd, f.issuer(perms, down, 20*time.Millisecond), &out, f.now)
	if !errors.Is(err, errSpiceDBSilent) {
		t.Fatalf("token issue with SpiceDB down = %v, want errSpiceDBSilent", err)
	}
	if out.Len() != 0 {
		t.Errorf("token issue with SpiceDB down printed %q, want nothing", out.String())
	}
	toks, listErr := f.tokens.List(t.Context())
	if listErr != nil || len(toks) != 0 {
		t.Errorf("tokens after the failed issue = %v, %v, want none", toks, listErr)
	}
	if trail := f.trail(t); len(trail) != 0 {
		t.Errorf("audit log after the failed issue = %v, want empty", trail)
	}
	if len(perms.writes) != 0 {
		t.Errorf("written to SpiceDB = %v, want nothing", perms.writes)
	}
}

// TestTokenIssueWaitsUntilSpiceDBAnswers checks that a SpiceDB that answers
// only on the third ask is waited for, and the token is then issued.
func TestTokenIssueWaitsUntilSpiceDBAnswers(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	perms := &flakyWrites{}
	asked := 0
	ping := func(context.Context) error {
		asked++
		if asked < 3 {
			return errNoAnswer
		}
		return nil
	}
	tok, secret, err := f.issuer(perms, ping, time.Minute).Issue(t.Context(), "nightly", time.Hour, true)
	if err != nil || !strings.HasPrefix(secret, "ofs_") || asked != 3 {
		t.Fatalf("Issue = %v, %q, %v after %d asks, want the token after 3", tok, secret, err, asked)
	}
}

// TestTokenIssueWithoutAGrantNeedsNothingOfSpiceDB checks that a token with
// no household grant is issued although SpiceDB does not answer: there is
// nothing to write there.
func TestTokenIssueWithoutAGrantNeedsNothingOfSpiceDB(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	perms := &flakyWrites{}
	_, secret, err := f.issuer(perms, down, 0).Issue(t.Context(), "nightly", time.Hour, false)
	if err != nil || !strings.HasPrefix(secret, "ofs_") {
		t.Fatalf("Issue without a grant, SpiceDB down = %q, %v, want the token", secret, err)
	}
	if len(perms.writes) != 0 {
		t.Errorf("written to SpiceDB = %v, want nothing", perms.writes)
	}
}

// TestTokenIssueRevokesWhenItsGrantCantReachSpiceDB checks the race: SpiceDB
// answered, but the write of the grant fails. The token, its grant and their
// entries were stored, so the token is revoked and its grant ended, each with
// its entry, no secret is printed, and the error says it was issued and
// revoked because its grant could not reach SpiceDB.
func TestTokenIssueRevokesWhenItsGrantCantReachSpiceDB(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	perms := &flakyWrites{}
	perms.failWrites.Store(true)
	var out bytes.Buffer
	cmd := &command{Name: cmdTokenIssue, TokenName: "nightly", TTL: time.Hour, GrantHousehold: true}
	err := tokenCommand(t.Context(), cmd, f.issuer(perms, up, time.Minute), &out, f.now)
	if !errors.Is(err, identityapp.ErrTokenRevoked) || !errors.Is(err, identityapp.ErrGrantNotProjected) {
		t.Fatalf("token issue = %v, want it issued and revoked because the grant was not passed on", err)
	}
	if out.Len() != 0 {
		t.Errorf("token issue printed %q, want no secret and nothing else", out.String())
	}
	toks, listErr := f.tokens.List(t.Context())
	if listErr != nil || len(toks) != 1 || toks[0].Status(f.now) != "revoked" {
		t.Fatalf("tokens = %v, %v, want the one, revoked", toks, listErr)
	}
	if !strings.Contains(err.Error(), string(toks[0].ID)) {
		t.Errorf("the error %q does not name the token %s", err, toks[0].ID)
	}
	f.requireNoGrantAndTrail(t, []string{"admin script_token_created ok", "admin grant_created ok", "admin script_token_revoked ok", "admin grant_revoked ok"})
}

// requireNoGrantAndTrail fails t unless no grant is in force and the audit log
// is exactly want.
func (f *grantsFixture) requireNoGrantAndTrail(t *testing.T, want []string) {
	t.Helper()
	if left, err := f.grants.InForce(t.Context()); err != nil || len(left) != 0 {
		t.Errorf("grants in force = %v, %v, want none", left, err)
	}
	if diff := cmp.Diff(want, f.trail(t)); diff != "" {
		t.Errorf("audit log (-want +got):\n%s", diff)
	}
}

// TestTokenIssueWithSpiceDBUp checks the happy path: the token and its grant
// are stored with their entries, the grant is written to SpiceDB, and the
// secret is printed first.
func TestTokenIssueWithSpiceDBUp(t *testing.T) {
	t.Parallel()
	f := newGrantsFixture(t)
	perms := &flakyWrites{}
	var out bytes.Buffer
	cmd := &command{Name: cmdTokenIssue, TokenName: "nightly", TTL: time.Hour, GrantHousehold: true}
	if err := tokenCommand(t.Context(), cmd, f.issuer(perms, up, time.Minute), &out, f.now); err != nil {
		t.Fatalf("token issue: %v", err)
	}
	secret, _, _ := strings.Cut(out.String(), "\n")
	toks, err := f.tokens.List(t.Context())
	if err != nil || len(toks) != 1 || toks[0].Status(f.now) != "active" || !strings.HasPrefix(secret, "ofs_") {
		t.Fatalf("tokens = %v, %v after printing %q, want one active token and its secret first", toks, err, out.String())
	}
	if want := []string{"OPERATION_TOUCH household:home#reader@script_token:" + string(toks[0].ID)}; !slices.Equal(perms.writes, want) {
		t.Errorf("written to SpiceDB = %v, want %v", perms.writes, want)
	}
	want := []string{"admin script_token_created ok", "admin grant_created ok"}
	if diff := cmp.Diff(want, f.trail(t)); diff != "" {
		t.Errorf("audit log (-want +got):\n%s", diff)
	}
}

// TestWaitForSpiceDB checks the wait: it returns at once when SpiceDB
// answers, asks once with a wait of zero, gives up at the wait with the last
// answer, and stops when the command is interrupted.
func TestWaitForSpiceDB(t *testing.T) {
	t.Parallel()
	if err := waitForSpiceDB(t.Context(), up, time.Minute, time.Hour); err != nil {
		t.Errorf("an answer at once = %v", err)
	}
	asked := 0
	counting := func(context.Context) error { asked++; return errNoAnswer }
	if err := waitForSpiceDB(t.Context(), counting, 0, time.Hour); !errors.Is(err, errSpiceDBSilent) || !errors.Is(err, errNoAnswer) || asked != 1 {
		t.Errorf("a wait of zero = %v after %d asks, want errSpiceDBSilent with the last answer after 1", err, asked)
	}
	if err := waitForSpiceDB(t.Context(), down, 10*time.Millisecond, time.Millisecond); !errors.Is(err, errSpiceDBSilent) {
		t.Errorf("a SpiceDB that never answers = %v, want errSpiceDBSilent", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := waitForSpiceDB(ctx, down, time.Hour, time.Hour); !errors.Is(err, errSpiceDBSilent) || !errors.Is(err, context.Canceled) {
		t.Errorf("an interrupted wait = %v, want errSpiceDBSilent and context.Canceled", err)
	}
}

// TestGateRefusesAsUnavailable checks that a core still without SpiceDB's
// key refuses as unavailable, so the requests that need a check get 503, and
// that its ping, which token issue waits on, says the same.
func TestGateRefusesAsUnavailable(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	az := gatedAuthorization(gateFor(t, &out))
	if err := az.ping(t.Context()); !errors.Is(err, kernel.ErrAuthzUnavailable) {
		t.Errorf("ping without the key = %v, want unavailable", err)
	}
	owner := identity.Caller{Holder: identity.HolderOwner, ID: "s"}
	if _, err := az.authorizer.MayAct(t.Context(), owner, identity.ActionView, "bonds"); !errors.Is(err, kernel.ErrAuthzUnavailable) {
		t.Errorf("MayAct without the key = %v, want unavailable", err)
	}
}
