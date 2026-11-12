package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
)

var errObserver = errors.New("authorization model unreachable")

// placements records what a PlacementObserver was told, and fails with err
// when set.
type placements struct {
	log []string
	err error
}

func (p *placements) AccountPlaced(_ context.Context, a ledger.AccountID, s portfolio.SleeveID) error {
	p.log = append(p.log, fmt.Sprintf("placed %s in %s", a, s))
	return p.err
}

func (p *placements) AccountRemoved(_ context.Context, a ledger.AccountID) error {
	p.log = append(p.log, "removed "+string(a))
	return p.err
}

// observedRegistry returns a registry over the built-in kinds and the sleeve
// core, with the ledger holding acc-1, that tells obs of each placement.
func observedRegistry(t *testing.T, obs *placements) *app.Registry {
	t.Helper()
	r := app.NewRegistry(newMemRegistry(), ledgerAccounts{acc1}, app.WithPlacementObserver(obs))
	for _, k := range portfolio.DefaultKinds() {
		if err := r.SaveKind(t.Context(), k); err != nil {
			t.Fatalf("SaveKind: %v", err)
		}
	}
	if err := r.SaveSleeve(t.Context(), coreSleeve()); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	return r
}

// TestRegistryTellsThePlacementObserver checks that the authorization
// model's copy hears of every stored placement and removal (D-164), and of
// no change that was refused.
func TestRegistryTellsThePlacementObserver(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	obs := &placements{}
	r := observedRegistry(t, obs)
	if err := r.PlaceAccount(ctx, "acc-9", "core"); err == nil {
		t.Fatal("PlaceAccount of an account the ledger lacks succeeded")
	}
	if err := r.PlaceAccount(ctx, acc1, "gold"); err == nil {
		t.Fatal("PlaceAccount in a sleeve not stored succeeded")
	}
	if len(obs.log) != 0 {
		t.Fatalf("the observer heard of changes that were refused: %v", obs.log)
	}
	if err := r.PlaceAccount(ctx, acc1, "core"); err != nil {
		t.Fatalf("PlaceAccount: %v", err)
	}
	if err := r.RemoveAccount(ctx, acc1); err != nil {
		t.Fatalf("RemoveAccount: %v", err)
	}
	if diff := cmp.Diff([]string{"placed acc-1 in core", "removed acc-1"}, obs.log); diff != "" {
		t.Errorf("observer log (-want +got):\n%s", diff)
	}
}

// TestRegistryReportsAnObserverThatFails checks that a failure of the
// observer leaves the change stored and says so.
func TestRegistryReportsAnObserverThatFails(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	obs := &placements{err: errObserver}
	r := observedRegistry(t, obs)
	err := r.PlaceAccount(ctx, acc1, "core")
	if !errors.Is(err, app.ErrPlacementNotObserved) || !errors.Is(err, errObserver) {
		t.Errorf("PlaceAccount with a failing observer = %v, want ErrPlacementNotObserved wrapping its error", err)
	}
	in, aerr := r.AccountSleeves(ctx)
	checkResult(t, "AccountSleeves", map[ledger.AccountID]portfolio.SleeveID{acc1: "core"}, in, aerr)
	if err := r.RemoveAccount(ctx, acc1); !errors.Is(err, app.ErrPlacementNotObserved) {
		t.Errorf("RemoveAccount with a failing observer = %v, want ErrPlacementNotObserved", err)
	}
	in, aerr = r.AccountSleeves(ctx)
	checkResult(t, "AccountSleeves after the removal", map[ledger.AccountID]portfolio.SleeveID{}, in, aerr)
}
