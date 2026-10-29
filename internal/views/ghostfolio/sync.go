package ghostfolio

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrRestrictedView reports that Ghostfolio's restricted view setting is
// on. It blanks comments, fees and quantities in every response, so every
// activity would look stale and the view refuses to sync (D-64).
var ErrRestrictedView = errors.New("the restricted view setting is on in Ghostfolio: turn it off to sync")

// Ghostfolio is what Syncer needs from Ghostfolio's API; *Client
// implements it.
type Ghostfolio interface {
	Login(ctx context.Context) error
	User(ctx context.Context) (User, error)
	Accounts(ctx context.Context) ([]string, error)
	CreateAccount(ctx context.Context, a Account) error
	Activities(ctx context.Context, accounts []string) ([]Activity, error)
	Import(ctx context.Context, b *Batch, dryRun bool) ([]ImportedActivity, error)
	DeleteActivity(ctx context.Context, id string) error
	DeleteActivitiesInAccounts(ctx context.Context, accounts []string) error
}

// BatchSize is the most activities one import carries, which keeps each
// request well under Ghostfolio's 10 MB body limit (R-15).
const BatchSize = 500

// Syncer runs the reconcile loop's one pass: it reads the record and makes
// the Ghostfolio accounts it owns hold exactly its projection (D-4, D-64).
type Syncer struct {
	record     Record
	ghostfolio Ghostfolio
	instance   string
	listings   map[ledger.ISIN]Listing
}

// NewSyncer returns the syncer of the view instance named instance, which
// reads record and writes to g, with the owner's listings.
func NewSyncer(record Record, g Ghostfolio, instance string, listings map[ledger.ISIN]Listing) *Syncer {
	return &Syncer{record: record, ghostfolio: g, instance: instance, listings: listings}
}

// Report counts what one sync changed, and lists the transactions it left
// out. It holds no amount.
type Report struct {
	AccountsCreated int
	Imported        int
	Deleted         int
	AccountsCleared int
	Problems        []Problem
}

// Sync runs one pass. Every step is safe to repeat, so a pass that fails
// halfway is finished by the next one:
//
//  1. log in, and refuse while the restricted view is on;
//  2. read the record and project it (Project);
//  3. read the owned accounts' activities and plan (MakePlan);
//  4. create the missing owned accounts with their derived UUIDs;
//  5. delete the stale activities an import would skip as duplicates;
//  6. import the missing activities with every asset's price history, each
//     batch checked first with a dry run (CheckPreview);
//  7. delete the remaining stale activities, one by one or with a bulk
//     delete filtered to an owned account.
func (s *Syncer) Sync(ctx context.Context) (Report, error) {
	g := s.ghostfolio
	if err := g.Login(ctx); err != nil {
		return Report{}, fmt.Errorf("logging in to Ghostfolio: %w", err)
	}
	user, err := g.User(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("reading the Ghostfolio user: %w", err)
	}
	if user.RestrictedView {
		return Report{}, ErrRestrictedView
	}
	accounts, err := ReadRecord(ctx, s.record)
	if err != nil {
		return Report{}, err
	}
	desired, problems := Project(&Source{Instance: s.instance, Currency: user.BaseCurrency, Accounts: accounts, Listings: s.listings})
	actual, err := s.actual(ctx, &desired)
	if err != nil {
		return Report{}, err
	}
	plan := MakePlan(&desired, &actual)
	rep := Report{Problems: problems}
	err = s.apply(ctx, &plan, &rep)
	return rep, err
}

// actual reads which accounts exist and the activities in the owned ones.
func (s *Syncer) actual(ctx context.Context, desired *Desired) (Actual, error) {
	ids, err := s.ghostfolio.Accounts(ctx)
	if err != nil {
		return Actual{}, fmt.Errorf("listing Ghostfolio accounts: %w", err)
	}
	var owned []string
	for _, a := range desired.Accounts {
		if slices.Contains(ids, a.ID) {
			owned = append(owned, a.ID)
		}
	}
	act := Actual{Accounts: ids}
	if len(owned) == 0 {
		return act, nil
	}
	if act.Activities, err = s.ghostfolio.Activities(ctx, owned); err != nil {
		return Actual{}, fmt.Errorf("listing the owned accounts' activities: %w", err)
	}
	return act, nil
}

// apply carries out p in Sync's order, counting into rep.
func (s *Syncer) apply(ctx context.Context, p *Plan, rep *Report) error {
	for _, a := range p.CreateAccounts {
		if err := s.ghostfolio.CreateAccount(ctx, a); err != nil {
			return fmt.Errorf("creating an owned account: %w", err)
		}
		rep.AccountsCreated++
	}
	if err := s.delete(ctx, p.PreDelete, rep); err != nil {
		return err
	}
	for _, b := range Batches(p.Import, p.Assets, BatchSize) {
		if err := s.importBatch(ctx, &b); err != nil {
			return err
		}
		rep.Imported += len(b.Activities)
	}
	if err := s.delete(ctx, p.Delete, rep); err != nil {
		return err
	}
	if len(p.Clear) > 0 {
		if err := s.ghostfolio.DeleteActivitiesInAccounts(ctx, p.Clear); err != nil {
			return fmt.Errorf("emptying owned accounts: %w", err)
		}
		rep.AccountsCleared += len(p.Clear)
	}
	return nil
}

// delete deletes acts one by one.
func (s *Syncer) delete(ctx context.Context, acts []Activity, rep *Report) error {
	for i := range acts {
		if err := s.ghostfolio.DeleteActivity(ctx, acts[i].ID); err != nil {
			return fmt.Errorf("deleting a stale activity: %w", err)
		}
		rep.Deleted++
	}
	return nil
}

// importBatch imports b once its dry run shows every activity stored as
// sent (D-64). An empty batch is skipped.
func (s *Syncer) importBatch(ctx context.Context, b *Batch) error {
	if len(b.Activities) == 0 && len(b.Assets) == 0 {
		return nil
	}
	preview, err := s.ghostfolio.Import(ctx, b, true)
	if err != nil {
		return fmt.Errorf("dry run of an import: %w", err)
	}
	if err := CheckPreview(b.Activities, preview); err != nil {
		return err
	}
	if _, err := s.ghostfolio.Import(ctx, b, false); err != nil {
		return fmt.Errorf("importing: %w", err)
	}
	return nil
}
