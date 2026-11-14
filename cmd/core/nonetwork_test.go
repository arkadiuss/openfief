package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
)

var (
	errSyncerDown = errors.New("SpiceDB is down")
	errFollowUp   = errors.New("the follow-up failed")
)

// changeContext returns a context as a change's transaction runs with it, the
// transaction itself not needed to tell it apart, with the list of what is to
// run after the change.
func changeContext(t *testing.T) (context.Context, *afterCommit) {
	t.Helper()
	after := &afterCommit{}
	ctx := context.WithValue(t.Context(), afterKey{}, after)
	return joined(nil).Join(ctx, nil), after
}

// noStream is a gRPC stream the tests never read.
type noStream struct{ grpc.ClientStream }

// countingTransport answers every request and counts them.
type countingTransport struct{ calls int }

func (c *countingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
}

// TestHTTPIsRefusedInsideAChange pins the rule that no network call is made
// inside a change's transaction (D-211) for HTTP: a request made with the
// context of a change fails with ErrNetworkInChange before anything is sent,
// and the same request goes through otherwise.
func TestHTTPIsRefusedInsideAChange(t *testing.T) {
	t.Parallel()
	inside, _ := changeContext(t)
	next := &countingTransport{}
	rt := noChangeTransport{next: next}
	if err := sendThrough(inside, t, rt); !errors.Is(err, ErrNetworkInChange) || next.calls != 0 {
		t.Errorf("RoundTrip inside a change = %v with %d requests sent; want ErrNetworkInChange and none", err, next.calls)
	}
	if err := sendThrough(outsideChange(inside), t, rt); err != nil || next.calls != 1 {
		t.Errorf("RoundTrip outside a change = %v with %d requests sent; want it sent", err, next.calls)
	}
}

// sendThrough sends a request made with ctx through rt, closes the response's
// body, and returns the error of the round trip.
func sendThrough(ctx context.Context, t *testing.T, rt http.RoundTripper) error {
	t.Helper()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://connector.invalid/call", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("round trip: %w", err)
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		t.Errorf("closing the body: %v", closeErr)
	}
	return nil
}

// TestPluginClientIsRefusedInsideAChange checks that the client the core calls
// connector instances and model providers with is built over the refusal, and
// that a request it makes inside a change never reaches the server.
func TestPluginClientIsRefusedInsideAChange(t *testing.T) {
	t.Parallel()
	inside, _ := changeContext(t)
	if _, ok := pluginClient().Transport.(noChangeTransport); !ok {
		t.Errorf("the plugin client's transport is %T, want the refusal over it", pluginClient().Transport)
	}
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit = true }))
	t.Cleanup(srv.Close)
	req, err := http.NewRequestWithContext(inside, http.MethodGet, srv.URL, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := pluginClient().Do(req)
	if err == nil {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("closing the body: %v", closeErr)
		}
	}
	if !errors.Is(err, ErrNetworkInChange) || hit {
		t.Errorf("a call of the plugin client inside a change = %v (server reached %v), want ErrNetworkInChange and no request", err, hit)
	}
}

// TestGRPCIsRefusedInsideAChange is the same for gRPC, which SpiceDB's client
// is built over: a unary call and a stream.
func TestGRPCIsRefusedInsideAChange(t *testing.T) {
	t.Parallel()
	inside, _ := changeContext(t)
	outside := outsideChange(inside)
	calls := 0
	invoker := func(context.Context, string, any, any, *grpc.ClientConn, ...grpc.CallOption) error {
		calls++
		return nil
	}
	if err := refuseUnaryInChange(inside, "/m", nil, nil, nil, invoker); !errors.Is(err, ErrNetworkInChange) || calls != 0 {
		t.Errorf("a unary call inside a change = %v with %d invoked, want ErrNetworkInChange and none", err, calls)
	}
	if err := refuseUnaryInChange(outside, "/m", nil, nil, nil, invoker); err != nil || calls != 1 {
		t.Errorf("a unary call outside a change = %v with %d invoked, want it made", err, calls)
	}
	streamed := 0
	streamer := func(context.Context, *grpc.StreamDesc, *grpc.ClientConn, string, ...grpc.CallOption) (grpc.ClientStream, error) {
		streamed++
		return noStream{}, nil
	}
	if _, err := refuseStreamInChange()(inside, nil, nil, "/m", streamer); !errors.Is(err, ErrNetworkInChange) || streamed != 0 {
		t.Errorf("a stream inside a change = %v with %d opened, want ErrNetworkInChange and none", err, streamed)
	}
	if _, err := refuseStreamInChange()(outside, nil, nil, "/m", streamer); err != nil || streamed != 1 {
		t.Errorf("a stream outside a change = %v with %d opened, want it opened", err, streamed)
	}
}

// TestSpiceDBIsRefusedInsideAChange checks that the client the core dials
// SpiceDB with refuses a check and a lookup made inside a change, before it
// connects to anything.
func TestSpiceDBIsRefusedInsideAChange(t *testing.T) {
	t.Parallel()
	inside, _ := changeContext(t)
	conn, err := dialSpiceDB(&authzConfig{Mode: authzSpiceDB, Endpoint: "192.0.2.1:50051", Insecure: true, Key: "synthetic", Timeout: time.Second})
	if err != nil {
		t.Fatalf("dialSpiceDB: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := conn.client.Close(); closeErr != nil {
			t.Errorf("closing the client: %v", closeErr)
		}
	})
	caller := identity.Caller{Holder: identity.HolderOwner, ID: "0123456789abcdef"}
	if _, err = conn.authorizer.MayAct(inside, caller, identity.ActionChange, "acc-1"); !errors.Is(err, ErrNetworkInChange) {
		t.Errorf("a SpiceDB check inside a change = %v, want ErrNetworkInChange", err)
	}
	if _, err = conn.authorizer.ActableAccounts(inside, caller, identity.ActionChange); !errors.Is(err, ErrNetworkInChange) {
		t.Errorf("a SpiceDB lookup inside a change = %v, want ErrNetworkInChange", err)
	}
}

// recordingSyncer is the syncer of SpiceDB's relationships that notes the
// placements and removals it is told of, and whether the context it was told
// with carried a change.
type recordingSyncer struct {
	accessSyncer

	placed, unplaced []string
	insideChange     bool
	err              error
}

func (s *recordingSyncer) AccountPlaced(ctx context.Context, account, sleeve string) error {
	s.insideChange = s.insideChange || refuseInChange(ctx) != nil
	s.placed = append(s.placed, account+" in "+sleeve)
	return s.err
}

func (s *recordingSyncer) AccountUnplaced(ctx context.Context, account string) error {
	s.insideChange = s.insideChange || refuseInChange(ctx) != nil
	s.unplaced = append(s.unplaced, account)
	return s.err
}

// TestPlacementsReachSpiceDBAfterTheChangeCommits checks that what a change
// tells SpiceDB, the placement of an account in a sleeve and the removal of one,
// is held until the change has committed and then told with a context outside
// it (D-211).
func TestPlacementsReachSpiceDBAfterTheChangeCommits(t *testing.T) {
	t.Parallel()
	ctx, after := changeContext(t)
	syncer := &recordingSyncer{}
	observer := placementSync{syncer: syncer}
	if err := observer.AccountPlaced(ctx, ledger.AccountID("acc-1"), portfolio.SleeveID("core")); err != nil {
		t.Fatalf("AccountPlaced: %v", err)
	}
	if err := observer.AccountRemoved(ctx, ledger.AccountID("acc-2")); err != nil {
		t.Fatalf("AccountRemoved: %v", err)
	}
	if len(syncer.placed)+len(syncer.unplaced) != 0 {
		t.Fatalf("SpiceDB was told %v and %v inside the change", syncer.placed, syncer.unplaced)
	}
	if err := after.run(outsideChange(ctx)); err != nil {
		t.Fatalf("what followed the change: %v", err)
	}
	if len(syncer.placed) != 1 || syncer.placed[0] != "acc-1 in core" || len(syncer.unplaced) != 1 || syncer.unplaced[0] != "acc-2" || syncer.insideChange {
		t.Errorf("after the commit SpiceDB was told %v and %v (inside a change: %v), want each once, outside the change",
			syncer.placed, syncer.unplaced, syncer.insideChange)
	}
}

// TestPlacementsOutsideAChangeReachSpiceDBAtOnce checks that a placement made
// by no change, as the core's own start-up makes, is told at once.
func TestPlacementsOutsideAChangeReachSpiceDBAtOnce(t *testing.T) {
	t.Parallel()
	syncer := &recordingSyncer{}
	if err := (placementSync{syncer: syncer}).AccountPlaced(t.Context(), "acc-3", "growth"); err != nil || len(syncer.placed) != 1 {
		t.Errorf("AccountPlaced outside a change = %v with %v told, want it told at once", err, syncer.placed)
	}
}

// TestPlacementsThatSpiceDBRefusesAreTheObserversToRepair checks that a SpiceDB
// that is down after the change committed leaves an error that wraps the
// registry's ErrPlacementNotObserved and the syncer's own: the change itself is
// stored, and the observer's reconcile repairs the rest.
func TestPlacementsThatSpiceDBRefusesAreTheObserversToRepair(t *testing.T) {
	t.Parallel()
	down := &recordingSyncer{err: errSyncerDown}
	ctx, after := changeContext(t)
	if err := (placementSync{syncer: down}).AccountPlaced(ctx, "acc-4", "core"); err != nil {
		t.Fatalf("AccountPlaced with SpiceDB down, held: %v", err)
	}
	err := after.run(outsideChange(ctx))
	if !errors.Is(err, portfolioapp.ErrPlacementNotObserved) || !errors.Is(err, errSyncerDown) {
		t.Errorf("what followed the change with SpiceDB down = %v, want ErrPlacementNotObserved wrapping the syncer's error", err)
	}
}

// TestOutsideChangeLeavesNoTransaction checks that work a change starts that
// outlives it, such as the sync the owner retries, gets a context that is no
// change's, so its own writes and calls are not taken for the change's.
func TestOutsideChangeLeavesNoTransaction(t *testing.T) {
	t.Parallel()
	inside, _ := changeContext(t)
	if refuseInChange(inside) == nil {
		t.Fatal("the change's context is not a change's")
	}
	if err := refuseInChange(outsideChange(inside)); err != nil {
		t.Errorf("refuseInChange(outsideChange(ctx)) = %v, want none", err)
	}
	if _, ok := txOf(outsideChange(inside)); ok {
		t.Error("the context outside the change still carries its transaction")
	}
}

// followUps records, as a change's follow-ups, the name of each that ran with
// how many accounts the database held then.
type followUps struct {
	t     *testing.T
	f     *outboxFixture
	which []string
}

// seen returns the follow-up that notes name.
func (u *followUps) seen(name string) func(context.Context) error {
	return func(ctx context.Context) error {
		if refuseInChange(ctx) != nil {
			u.t.Errorf("%s ran with a change's context", name)
		}
		var n int
		if err := u.f.pool.QueryRow(ctx, `SELECT count(*) FROM ledger.accounts WHERE id LIKE 'after-%'`).Scan(&n); err != nil {
			return fmt.Errorf("counting the accounts the follow-up sees: %w", err)
		}
		u.which = append(u.which, name+":"+strconv.Itoa(n))
		return nil
	}
}

// TestAfterChangeRunsOnlyOnceTheOutermostChangeCommits checks over a real
// database (D-211) that what a change holds for after it runs once the
// outermost change has committed, with the change's writes visible and a context
// that is no change's, and not when only a nested change has.
func TestAfterChangeRunsOnlyOnceTheOutermostChangeCommits(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	u := &followUps{t: t, f: f}
	err := f.out.Within(t.Context(), f.entry(t, "after-1"), func(ctx context.Context) error {
		if err := insertAccount(ctx, f.pool, "after-1"); err != nil {
			return err
		}
		nested := f.out.Within(ctx, f.entry(t, "after-2"), func(ctx context.Context) error {
			return errors.Join(insertAccount(ctx, f.pool, "after-2"), afterChange(ctx, u.seen("nested")))
		})
		if nested != nil || len(u.which) != 0 {
			t.Errorf("the nested change = %v, with %v run before the outermost change committed", nested, u.which)
		}
		return afterChange(ctx, u.seen("outer"))
	})
	if err != nil {
		t.Fatalf("Within: %v", err)
	}
	if len(u.which) != 2 || u.which[0] != "nested:2" || u.which[1] != "outer:2" {
		t.Errorf("ran %v, want the nested then the outer follow-up, once the two accounts were committed", u.which)
	}
}

// TestAfterChangeNeverRunsForAChangeThatRollsBack checks that a change that
// fails after it held a follow-up never runs it, and leaves no entry.
func TestAfterChangeNeverRunsForAChangeThatRollsBack(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	u := &followUps{t: t, f: f}
	err := f.out.Within(t.Context(), f.entry(t, "gone-1"), func(ctx context.Context) error {
		if err := errors.Join(insertAccount(ctx, f.pool, "gone-1"), afterChange(ctx, u.seen("rolled back"))); err != nil {
			return err
		}
		return errLastStepBroke
	})
	if !errors.Is(err, errLastStepBroke) || len(u.which) != 0 || f.accounts(t, "gone-") != 0 {
		t.Errorf("Within of a change that rolls back = %v, ran %v with %d accounts; want the error, no follow-up and no account", err, u.which, f.accounts(t, "gone-"))
	}
}

// TestAfterChangeThatFailsLeavesTheChangeCommitted checks that an error of a
// follow-up is returned with the change, and its entry, committed.
func TestAfterChangeThatFailsLeavesTheChangeCommitted(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	err := f.out.Within(t.Context(), f.entry(t, "kept-1"), func(ctx context.Context) error {
		if insertErr := insertAccount(ctx, f.pool, "kept-1"); insertErr != nil {
			return insertErr
		}
		return afterChange(ctx, func(context.Context) error { return errFollowUp })
	})
	if !errors.Is(err, errFollowUp) || f.accounts(t, "kept-") != 1 || len(f.jobs(t)) != 1 {
		t.Errorf("Within with a follow-up that fails = %v with %d accounts and %d entries; want its error, the change and its entry committed", err, f.accounts(t, "kept-"), len(f.jobs(t)))
	}
}

// TestAfterChangeFollowsItsSavepoint checks that what a nested change holds
// for after it goes where its savepoint goes: a savepoint that rolls back
// discards it, one that commits hands it to the outer change, which runs it
// after its own commit.
func TestAfterChangeFollowsItsSavepoint(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	u := &followUps{t: t, f: f}
	err := f.out.Within(t.Context(), f.entry(t, "after-1"), func(ctx context.Context) error {
		undone := f.out.Within(ctx, f.entry(t, "after-2"), func(ctx context.Context) error {
			return errors.Join(afterChange(ctx, u.seen("rolled back")), errLastStepBroke)
		})
		if !errors.Is(undone, errLastStepBroke) {
			t.Errorf("the nested change = %v, want its error", undone)
		}
		return f.out.Within(ctx, f.entry(t, "after-3"), func(ctx context.Context) error {
			return afterChange(ctx, u.seen("committed"))
		})
	})
	if err != nil {
		t.Fatalf("Within: %v", err)
	}
	if len(u.which) != 1 || u.which[0] != "committed:0" {
		t.Errorf("ran %v, want only the committed savepoint's follow-up, after the outermost commit", u.which)
	}
}

// TestAfterChangeOfACommittedSavepointNeverRunsWhenTheOutermostRollsBack
// checks that a follow-up handed up by a savepoint that committed is dropped
// with the outermost transaction when that rolls back.
func TestAfterChangeOfACommittedSavepointNeverRunsWhenTheOutermostRollsBack(t *testing.T) {
	t.Parallel()
	f := newOutboxFixture(t)
	u := &followUps{t: t, f: f}
	err := f.out.Within(t.Context(), f.entry(t, "gone-1"), func(ctx context.Context) error {
		nested := f.out.Within(ctx, f.entry(t, "gone-2"), func(ctx context.Context) error {
			return afterChange(ctx, u.seen("committed savepoint"))
		})
		return errors.Join(nested, errLastStepBroke)
	})
	if !errors.Is(err, errLastStepBroke) || len(u.which) != 0 {
		t.Errorf("Within = %v, ran %v; want the error and nothing run", err, u.which)
	}
}
