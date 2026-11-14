package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/spicedb"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/portfolio"
	portfolioapp "github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/views"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

const (
	// grantsRetryUnknown is how often the start retries a step that waits
	// for the installed plugins to be read.
	grantsRetryUnknown = 2 * time.Second
	// grantsRetryFailed is how often the start retries a step that failed
	// for another reason, such as a SpiceDB that isn't up yet.
	grantsRetryFailed = 10 * time.Second
	// grantsCallTimeout bounds one grant change the core makes on its own,
	// when a plugin is uninstalled.
	grantsCallTimeout = 30 * time.Second
)

// installedPlugins is what the grants' directory and the cutover read of the
// plugins the operator runs; *viewskube.Plugins implements it.
type installedPlugins interface {
	InstalledPlugins(ctx context.Context) ([]views.PluginRecord, error)
}

// accountLookup finds an account in the ledger; *ledgerapp.Ledger implements
// it.
type accountLookup interface {
	Account(ctx context.Context, id ledger.AccountID) (ledger.Account, error)
}

// sleeveLister lists the sleeves; *portfolioapp.Registry implements it.
type sleeveLister interface {
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
}

// grantDirectory is identityapp.Directory over the other contexts: the
// ledger's accounts, the portfolio's sleeves and the installed plugins. Its
// fields are set once the contexts are wired, which they need the grants for.
type grantDirectory struct {
	accounts accountLookup
	sleeves  sleeveLister
	// plugins is nil outside a cluster, where no plugin is installed.
	plugins installedPlugins
}

var _ identityapp.Directory = (*grantDirectory)(nil)

// TargetExists implements identityapp.Directory.
func (d *grantDirectory) TargetExists(ctx context.Context, t identity.GrantTarget) (bool, error) {
	switch t.Kind {
	case identity.GrantTargetHousehold:
		return true, nil
	case identity.GrantTargetAccount:
		_, err := d.accounts.Account(ctx, ledger.AccountID(t.ID))
		if errors.Is(err, ledgerapp.ErrUnknownAccount) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("looking for an account: %w", err)
		}
		return true, nil
	case identity.GrantTargetSleeve:
		sleeves, err := d.sleeves.Sleeves(ctx)
		if err != nil {
			return false, fmt.Errorf("looking for a sleeve: %w", err)
		}
		return slices.ContainsFunc(sleeves, func(s portfolio.Sleeve) bool { return string(s.ID) == t.ID }), nil
	}
	return false, nil
}

// ViewPluginInstalled implements identityapp.Directory. Outside a cluster it
// can't tell, so a grant of a view plugin is neither made nor swept there.
func (d *grantDirectory) ViewPluginInstalled(ctx context.Context, instance string) (bool, error) {
	records, err := d.records(ctx)
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(records, func(r views.PluginRecord) bool { return r.Instance == instance && !r.Deleting && isViewRecord(&r) }), nil
}

// ViewPlugins implements identityapp.Directory: the plugins that hold the
// view capability or ask for it, so one the operator has not approved yet is
// given the household grant too, as it would view every account under the
// policy once approved. Outside a cluster there is none.
func (d *grantDirectory) ViewPlugins(ctx context.Context) ([]string, error) {
	if d.plugins == nil {
		return nil, nil
	}
	records, err := d.records(ctx)
	if err != nil {
		return nil, err
	}
	var out []string
	for i := range records {
		if !records[i].Deleting && isViewRecord(&records[i]) {
			out = append(out, records[i].Instance)
		}
	}
	slices.Sort(out)
	return out, nil
}

// AccessViewPlugins implements identityapp.Directory: the view plugins that
// read with their own identity, so the owner's Access page lists them
// (D-210). It leaves out the plugin that serves browsers, which reads with the
// signed-in person's session, and a plugin being removed. Outside a cluster
// there is none.
func (d *grantDirectory) AccessViewPlugins(ctx context.Context) ([]string, error) {
	if d.plugins == nil {
		return nil, nil
	}
	records, err := d.records(ctx)
	if err != nil {
		return nil, err
	}
	var out []string
	for i := range records {
		if !records[i].Deleting && isViewRecord(&records[i]) && records[i].BrowserPath == "" {
			out = append(out, records[i].Instance)
		}
	}
	slices.Sort(out)
	return out, nil
}

// records returns the installed plugins, or an error wrapping
// identityapp.ErrDirectoryUnknown when there is no cluster or the cache of
// them has not synced.
func (d *grantDirectory) records(ctx context.Context) ([]views.PluginRecord, error) {
	if d.plugins == nil {
		return nil, fmt.Errorf("%w: not in a Kubernetes cluster", identityapp.ErrDirectoryUnknown)
	}
	records, err := d.plugins.InstalledPlugins(ctx)
	if errors.Is(err, viewsapp.ErrPluginsUnknown) {
		return nil, fmt.Errorf("%w: %w", identityapp.ErrDirectoryUnknown, err)
	}
	if err != nil {
		return nil, fmt.Errorf("listing the installed plugins: %w", err)
	}
	return records, nil
}

// isViewRecord reports whether a plugin holds or asks for the view
// capability.
func isViewRecord(r *views.PluginRecord) bool {
	view := string(plugins.CapabilityView)
	return slices.Contains(r.Effective, view) || slices.Contains(r.Requested, view)
}

// grantCleanup is portfolioapp.PlacementObserver ending the grants of an
// account or a sleeve that was deleted (D-209), as the core: a sleeve created
// later under the same ID would otherwise inherit them. The deletion itself
// is the owner's; its own audit entry comes with the rest of D-211's list.
type grantCleanup struct{ grants *identityapp.Grants }

var _ portfolioapp.PlacementObserver = grantCleanup{}

// AccountPlaced does nothing: a placement grants nothing.
func (grantCleanup) AccountPlaced(context.Context, ledger.AccountID, portfolio.SleeveID) error {
	return nil
}

// AccountRemoved ends the grants of the deleted account, once the change that
// deleted it has committed (afterChange): ending a grant calls SpiceDB, which
// no change's transaction does.
func (c grantCleanup) AccountRemoved(ctx context.Context, account ledger.AccountID) error {
	return afterChange(ctx, func(ctx context.Context) error {
		_, err := c.grants.RevokeTarget(ctx, systemActor(), identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: string(account)})
		if err != nil {
			return fmt.Errorf("%w: ending the grants of an account: %w", portfolioapp.ErrPlacementNotObserved, err)
		}
		return nil
	})
}

// SleeveRemoved ends the grants of the deleted sleeve, once the change that
// deleted it has committed.
func (c grantCleanup) SleeveRemoved(ctx context.Context, sleeve portfolio.SleeveID) error {
	return afterChange(ctx, func(ctx context.Context) error {
		_, err := c.grants.RevokeTarget(ctx, systemActor(), identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: string(sleeve)})
		if err != nil {
			return fmt.Errorf("%w: ending the grants of a sleeve: %w", portfolioapp.ErrPlacementNotObserved, err)
		}
		return nil
	})
}

// grantsOnUninstall returns what the core does when a plugin instance is
// uninstalled (D-209): end its grants, as the core, each with an audit entry.
// A failure is logged without anything but the instance's name; the next
// start's sweep ends what is left.
func grantsOnUninstall(ctx context.Context, grants *identityapp.Grants, log *slog.Logger) func(instance string) {
	return func(instance string) {
		holder := identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: instance}
		if holder.Validate() != nil {
			return
		}
		call, cancel := context.WithTimeout(context.WithoutCancel(ctx), grantsCallTimeout)
		defer cancel()
		n, err := grants.RevokeHolder(call, systemActor(), holder)
		if err != nil {
			log.Warn("the grants of an uninstalled plugin could not be ended; the next start sweeps them",
				"plugin", instance, "error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded))
			return
		}
		if n > 0 {
			log.Info("ended the grants of an uninstalled plugin", "plugin", instance, "grants", n)
		}
	}
}

// grantsStartup is what the core does about grants at start, after the
// accounts and sleeves are wired: with SpiceDB, the cutover (D-209), once;
// then the sweep of grants that outlived the holder or the target they name;
// and, when a grant could not be passed on to SpiceDB, a reconcile. It
// retries what fails until it succeeds or ctx ends: the installed plugins
// are read in the background, and SpiceDB may not be up yet. A failure is
// logged without a credential or anything granted.
func grantsStartup(ctx context.Context, grants *identityapp.Grants, spicedbOn bool, reconcile func(context.Context) error, log *slog.Logger) {
	reconcileNeeded := false
	for {
		err := grantsStep(ctx, grants, spicedbOn, reconcile, &reconcileNeeded, log)
		if err == nil {
			return
		}
		wait := grantsRetryFailed
		if errors.Is(err, identityapp.ErrDirectoryUnknown) {
			wait = grantsRetryUnknown
		} else {
			log.Warn("the grants are not settled yet; retrying", "error", kernel.LogSafe(err, context.Canceled, context.DeadlineExceeded, identity.ErrInvalidGrant))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

// grantsStep runs the start's steps once: the cutover when SpiceDB answers,
// the sweep, and the reconcile when a change could not be passed on.
func grantsStep(ctx context.Context, grants *identityapp.Grants, spicedbOn bool, reconcile func(context.Context) error, reconcileNeeded *bool, log *slog.Logger) error {
	if spicedbOn {
		n, err := grants.Cutover(ctx, systemActor())
		if errors.Is(err, identityapp.ErrGrantNotProjected) {
			*reconcileNeeded = true
			err = nil
		}
		if err != nil {
			return fmt.Errorf("the cutover: %w", err)
		}
		if n > 0 {
			log.Info("gave the household grant to every script token and view plugin that held read access", "grants", n)
		}
	}
	swept, err := grants.Sweep(ctx, systemActor())
	if err != nil {
		return fmt.Errorf("the sweep: %w", err)
	}
	if swept > 0 {
		log.Info("ended the grants of holders and targets that are gone", "grants", swept)
	}
	if *reconcileNeeded && reconcile != nil {
		if err := reconcile(ctx); err != nil {
			return fmt.Errorf("the reconcile: %w", err)
		}
		*reconcileNeeded = false
	}
	return nil
}

// newGrants returns the grant use cases over store, passing each change to
// SpiceDB through az's syncer, and to nothing with the policy.
func newGrants(store identityapp.GrantStore, tokens *identityapp.Tokens, dir identityapp.Directory, az authorization, now func() time.Time) *identityapp.Grants {
	var g *identityapp.Grants
	if az.syncer == nil {
		g = identityapp.NewGrants(store, tokens, dir, nil, now)
	} else {
		g = identityapp.NewGrants(store, tokens, dir, az.syncer, now).WithPing(az.ping)
	}
	// A grant names a viewer only while they are active (D-65): the store that keeps the grants keeps the people.
	if people, ok := store.(identityapp.PersonReader); ok {
		g.WithPeople(people)
	}
	return g
}

// accountLister lists every account; *ledgerapp.Ledger implements it.
type accountLister interface {
	Accounts(ctx context.Context) ([]ledger.Account, error)
}

// sleevePlacements says which sleeve each account is in; *portfolioapp.Registry
// implements it.
type sleevePlacements interface {
	AccountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error)
}

// targetCover is identityapp.TargetCover over the ledger's accounts and the
// portfolio's placements: what a viewer's grants cover under the policy.
type targetCover struct {
	accounts accountLister
	sleeves  sleevePlacements
}

var _ identityapp.TargetCover = targetCover{}

// TargetAccounts implements identityapp.TargetCover.
func (c targetCover) TargetAccounts(ctx context.Context, t identity.GrantTarget) ([]string, error) {
	switch t.Kind {
	case identity.GrantTargetAccount:
		return []string{t.ID}, nil
	case identity.GrantTargetHousehold:
		all, err := c.accounts.Accounts(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing the accounts: %w", err)
		}
		out := make([]string, 0, len(all))
		for i := range all {
			out = append(out, string(all[i].ID))
		}
		return out, nil
	case identity.GrantTargetSleeve:
		placed, err := c.sleeves.AccountSleeves(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing the accounts of a sleeve: %w", err)
		}
		var out []string
		for account, sleeve := range placed {
			if string(sleeve) == t.ID {
				out = append(out, string(account))
			}
		}
		slices.Sort(out)
		return out, nil
	}
	return nil, nil
}

var _ accessSyncer = (*spicedb.Syncer)(nil)

// householdOnly is the identityapp.Directory of the core's own commands,
// which know no accounts or plugins: they grant the household alone.
type householdOnly struct{}

var _ identityapp.Directory = householdOnly{}

func (householdOnly) TargetExists(_ context.Context, t identity.GrantTarget) (bool, error) {
	return t.Kind == identity.GrantTargetHousehold, nil
}

func (householdOnly) ViewPluginInstalled(context.Context, string) (bool, error) {
	return false, fmt.Errorf("%w: a command of the core reads no plugins", identityapp.ErrDirectoryUnknown)
}

func (householdOnly) ViewPlugins(context.Context) ([]string, error) {
	return nil, fmt.Errorf("%w: a command of the core reads no plugins", identityapp.ErrDirectoryUnknown)
}

func (householdOnly) AccessViewPlugins(context.Context) ([]string, error) {
	return nil, fmt.Errorf("%w: a command of the core reads no plugins", identityapp.ErrDirectoryUnknown)
}
