package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Errors of the grant use cases. Every error they return wraps one of these
// or ErrInvalidRequest and ErrNotFound, or is a store failure.
var (
	// ErrGrantNotProjected reports a grant change that is stored, and
	// recorded in the audit log, but that the authorization model could not
	// take: SpiceDB is ahead or behind the grants table until the change is
	// made again, which is idempotent, or the core's next start reconciles
	// them. An error says what failed, never what was granted.
	ErrGrantNotProjected = errors.New("stored, but not passed on to the authorization model")
	// ErrDirectoryUnknown reports a holder or target the core can't tell is
	// there now, such as a plugin outside a cluster or before its cache has
	// synced. It refuses a new grant, and never revokes one.
	ErrDirectoryUnknown = errors.New("can't tell whether it exists yet")
	// ErrTokenRevoked reports a token that was issued and then revoked
	// because its household grant could not reach the authorization model:
	// no secret is returned for it, and the token views nothing.
	ErrTokenRevoked = errors.New("issued and revoked again: its grant could not reach the authorization model")
)

// revokeTimeout bounds the revocation of a token whose grant could not be
// passed on, which runs even when the caller's own context has ended.
const revokeTimeout = 30 * time.Second

// GrantStore keeps grants (D-209): the source of truth for who may view
// what, which the authorization model only projects. A grant and its audit
// entry share a transaction (Atomic, Auditor). *identity/postgres.Store
// implements it.
type GrantStore interface {
	Atomic
	Auditor
	// SaveGrant stores g as in force. It stores nothing and returns false
	// when a grant of g's target to g's holder is in force already.
	SaveGrant(ctx context.Context, g *identity.Grant) (bool, error)
	// RevokeGrants ends, as of at, every grant in force that m selects and
	// returns them as ended. A zero m selects every grant.
	RevokeGrants(ctx context.Context, m identity.GrantMatch, at time.Time) ([]identity.Grant, error)
	// Grants returns the grants m selects, oldest first and in ID order:
	// those in force only, or the ended ones too.
	Grants(ctx context.Context, m identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error)
	// GrantCutoverDone reports whether the cutover to SpiceDB ran (D-209).
	GrantCutoverDone(ctx context.Context) (bool, error)
	// MarkGrantCutover records that the cutover ran at at, and returns
	// false when it was recorded already.
	MarkGrantCutover(ctx context.Context, at time.Time) (bool, error)
}

// GrantProjector keeps the authorization model's copy of the grants in force
// (D-164): the relationship each grant stands for. Both calls are
// idempotent. *identity/spicedb.Syncer implements it; with the policy
// there is none, and grants are stored with no effect.
type GrantProjector interface {
	// ProjectGrants writes the relationship of each grant.
	ProjectGrants(ctx context.Context, grants []identity.Grant) error
	// UnprojectGrants removes the relationship of each grant.
	UnprojectGrants(ctx context.Context, grants []identity.Grant) error
}

// Directory says what the core's other contexts hold, so a grant never
// names a record that isn't there. cmd/core implements it over the ledger,
// the portfolio and the installed plugins.
type Directory interface {
	// TargetExists reports whether the account or sleeve t names is stored.
	// The household always is.
	TargetExists(ctx context.Context, t identity.GrantTarget) (bool, error)
	// ViewPluginInstalled reports whether a view plugin instance with this
	// name is installed. It fails with an error wrapping ErrDirectoryUnknown
	// when the core can't tell.
	ViewPluginInstalled(ctx context.Context, instance string) (bool, error)
	// ViewPlugins returns the instance names of the installed view
	// plugins, those that hold or ask for the view capability, or an
	// error wrapping ErrDirectoryUnknown.
	ViewPlugins(ctx context.Context) ([]string, error)
	// AccessViewPlugins returns the instance names of the installed view
	// plugins that read with their own identity and so need grants (D-210):
	// those that hold or ask for the view capability, less the one that
	// serves browsers, which reads as the signed-in person, and any being
	// removed. It fails with an error wrapping ErrDirectoryUnknown when the
	// core can't tell.
	AccessViewPlugins(ctx context.Context) ([]string, error)
}

// Grants are the use cases of the owner's grants (D-209, D-211). Each change
// is stored with its audit entry in one transaction, so if the entry can't
// be written the grant isn't made or ended, and is then passed on to the
// authorization model.
type Grants struct {
	store  GrantStore
	tokens *Tokens
	dir    Directory
	proj   GrantProjector
	now    func() time.Time
	// ping asks the authorization model whether it answers; nil with the
	// policy, which has nothing to ask.
	ping func(context.Context) error
}

// NewGrants returns the grant use cases over store, checking script tokens
// with tokens and every other holder and target with dir, passing changes to
// proj (nil with the policy) and reading the time from now.
func NewGrants(store GrantStore, tokens *Tokens, dir Directory, proj GrantProjector, now func() time.Time) *Grants {
	return &Grants{store: store, tokens: tokens, dir: dir, proj: proj, now: now}
}

// Enforced reports whether the grants limit what holders view now: they do
// while the authorization model answers, and have no effect under the policy,
// which lets every holder of read access view every account.
func (g *Grants) Enforced() bool { return g.proj != nil }

// WithPing returns g asking ping, before it stores a token that is to hold
// grants, whether the authorization model answers: a model that doesn't makes
// the token's creation fail with an error wrapping ErrGrantNotProjected and
// store nothing. ping is nil with the policy.
func (g *Grants) WithPing(ping func(context.Context) error) *Grants {
	g.ping = ping
	return g
}

// Create grants target to holder on behalf of actor, which records it in the
// audit log in the same transaction as the grant, and passes it on to the
// authorization model. It returns the grant in force: granting what is
// granted already changes and records nothing, and only repeats the
// idempotent projection, which repairs a SpiceDB that missed it. A holder or
// target that doesn't exist fails with an error wrapping ErrNotFound, a
// malformed one, or a kind with no holder yet, with ErrInvalidRequest.
func (g *Grants) Create(ctx context.Context, actor identity.Actor, holder identity.GrantHolder, target identity.GrantTarget) (identity.Grant, error) {
	grant, err := identity.NewGrant(holder, target, g.now().UTC().Truncate(time.Microsecond))
	if err != nil {
		return identity.Grant{}, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if holder.Kind == identity.GrantHolderPerson {
		return identity.Grant{}, fmt.Errorf("%w: viewers don't exist yet", ErrInvalidRequest)
	}
	if requireErr := g.require(ctx, holder, target); requireErr != nil {
		return identity.Grant{}, requireErr
	}
	created, err := g.saveGrant(ctx, actor, &grant)
	if err != nil {
		return identity.Grant{}, err
	}
	if !created {
		if grant, err = g.inForceOf(ctx, holder, target); err != nil {
			return identity.Grant{}, err
		}
	}
	if projectErr := g.project(ctx, []identity.Grant{grant}); projectErr != nil {
		return grant, projectErr
	}
	return grant, nil
}

// Revoke ends the grant of target to holder on behalf of actor: the
// relationship is removed from the authorization model first, so a failure
// there changes nothing and the holder keeps no access the grants table
// has ended, and then the table and the audit entry commit together. Revoking a
// grant that ended already removes its relationship again and records
// nothing; one that never existed fails with an error wrapping ErrNotFound.
func (g *Grants) Revoke(ctx context.Context, actor identity.Actor, holder identity.GrantHolder, target identity.GrantTarget) error {
	m := identity.GrantMatch{Holder: holder, Target: target}
	if err := m.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if holder == (identity.GrantHolder{}) || target == (identity.GrantTarget{}) {
		return fmt.Errorf("%w: a grant has a holder and a target", ErrInvalidRequest)
	}
	n, err := g.revoke(ctx, actor, m)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	ended, err := g.store.Grants(ctx, m, false)
	if err != nil {
		return fmt.Errorf("loading the grant: %w", err)
	}
	if len(ended) == 0 {
		return fmt.Errorf("%w: no such grant", ErrNotFound)
	}
	return g.unproject(ctx, ended)
}

// RevokeHolder ends every grant in force of holder, each with its own audit
// entry by actor, and returns how many it ended: the grants of a script
// token that was revoked, or of a view plugin that was uninstalled (D-209).
func (g *Grants) RevokeHolder(ctx context.Context, actor identity.Actor, holder identity.GrantHolder) (int, error) {
	if err := holder.Validate(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	return g.revoke(ctx, actor, identity.GrantMatch{Holder: holder})
}

// RevokeTarget ends every grant in force of target, each with its own audit
// entry by actor, and returns how many it ended: the grants of a sleeve that
// was deleted (D-120), so a sleeve created later under the same ID doesn't
// inherit them.
func (g *Grants) RevokeTarget(ctx context.Context, actor identity.Actor, target identity.GrantTarget) (int, error) {
	if err := target.Validate(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	return g.revoke(ctx, actor, identity.GrantMatch{Target: target})
}

// List returns the grants m selects, oldest first: those in force only, or
// the ended ones, the history, too.
func (g *Grants) List(ctx context.Context, m identity.GrantMatch, inForceOnly bool) ([]identity.Grant, error) {
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	grants, err := g.store.Grants(ctx, m, inForceOnly)
	if err != nil {
		return nil, fmt.Errorf("listing grants: %w", err)
	}
	return grants, nil
}

// InForce returns every grant in force, which the authorization model's
// reconcile writes as relationships.
func (g *Grants) InForce(ctx context.Context) ([]identity.Grant, error) {
	return g.List(ctx, identity.GrantMatch{}, true)
}

// RevokeToken revokes the script token id as Tokens.Revoke does, ends its
// grants and records both in the audit log, all in one transaction (D-209,
// D-211), and then removes the grants from the authorization model: the
// token no longer authenticates, so a failure there leaves nothing viewable
// and fails with an error wrapping ErrGrantNotProjected, which the next
// start's reconcile repairs. It is RevokeToken with grants.
func (g *Grants) RevokeToken(ctx context.Context, actor identity.Actor, id identity.ScriptTokenID) error {
	var ended []identity.Grant
	err := within(ctx, g.store, func(ctx context.Context) error {
		if revokeErr := RevokeToken(ctx, g.store, g.tokens, actor, id); revokeErr != nil {
			return revokeErr
		}
		var endErr error
		ended, endErr = g.revokeStored(ctx, actor, identity.GrantMatch{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(id)}})
		return endErr
	})
	if err != nil {
		return err
	}
	if len(ended) == 0 || g.proj == nil {
		return nil
	}
	if unprojectErr := g.proj.UnprojectGrants(ctx, ended); unprojectErr != nil {
		return fmt.Errorf("%w: %w", ErrGrantNotProjected, unprojectErr)
	}
	return nil
}

// NewToken is what the owner chooses when creating a script token (D-210):
// its name, how long it lasts, whether it may use MCP (D-213), and the grants
// it starts with.
type NewToken struct {
	Name string
	// TTL is the token's lifetime, at most identity.MaxTTL.
	TTL time.Duration
	// MCPAllowed is stored with the token and changes nothing until MCP is
	// served.
	MCPAllowed bool
	// Targets are the grants the token is given: accounts, sleeves or the
	// household. None leaves a token that views nothing under SpiceDB.
	Targets []identity.GrantTarget
}

// IssueToken creates a read-only token for actor as IssueToken does and,
// when household is set, its household grant (D-209): see IssueTokenWith.
func (g *Grants) IssueToken(ctx context.Context, actor identity.Actor, name string, ttl time.Duration, household bool) (identity.ScriptToken, string, error) {
	spec := NewToken{Name: name, TTL: ttl}
	if household {
		spec.Targets = []identity.GrantTarget{identity.Household()}
	}
	return g.IssueTokenWith(ctx, actor, &spec)
}

// IssueTokenWith creates a read-only token for actor and the grants spec
// names (D-209, D-210): the token, its grants and the audit entries of all of
// them in one transaction (D-211), and then the grants are written to the
// authorization model. If that write fails, the token is revoked at once,
// with its own audit entry, no secret is returned, and the error wraps
// ErrTokenRevoked and ErrGrantNotProjected; if the revocation fails too, the
// error says the token must be revoked by hand. Before storing anything it
// checks that every target exists, as an error wrapping ErrNotFound, and,
// when the token is to hold grants, that the model answers: a model that
// doesn't fails with an error wrapping ErrGrantNotProjected and stores
// nothing. A token without grants needs nothing of the model.
func (g *Grants) IssueTokenWith(ctx context.Context, actor identity.Actor, spec *NewToken) (identity.ScriptToken, string, error) {
	targets, err := g.checkTargets(ctx, spec.Targets)
	if err != nil {
		return identity.ScriptToken{}, "", err
	}
	if len(targets) > 0 && g.ping != nil {
		if pingErr := g.ping(ctx); pingErr != nil {
			return identity.ScriptToken{}, "", fmt.Errorf("%w: the authorization model does not answer: %w", ErrGrantNotProjected, pingErr)
		}
	}
	tok, secret, grants, err := g.storeToken(ctx, actor, spec, targets)
	if err != nil {
		return identity.ScriptToken{}, "", err
	}
	if len(grants) == 0 {
		return tok, secret, nil
	}
	projectErr := g.project(ctx, grants)
	if projectErr == nil {
		return tok, secret, nil
	}
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), revokeTimeout)
	defer cancel()
	// The grants never reached the model, so the removal RevokeToken then asks of it may fail too:
	// the token is revoked in the table either way, and the model holds nothing of it.
	if revokeErr := g.RevokeToken(rctx, actor, tok.ID); revokeErr != nil && !errors.Is(revokeErr, ErrGrantNotProjected) {
		return identity.ScriptToken{ID: tok.ID}, "", fmt.Errorf("token %s was issued but its grants could not reach the authorization model, and revoking it failed, so revoke it by hand: %w", tok.ID, errors.Join(projectErr, revokeErr))
	}
	return identity.ScriptToken{ID: tok.ID}, "", fmt.Errorf("token %s: %w: %w", tok.ID, ErrTokenRevoked, projectErr)
}

// UpdateGrants changes what holder may view on behalf of actor: it ends the
// grants of each target in remove, then makes those of each target in add,
// every change an audit entry (D-211) passed on to the authorization model as
// Create and Revoke do. Everything is checked before anything changes: the
// holder must exist, as must each target to add. Removing a grant holder
// doesn't hold, and adding one it holds already, change nothing. It returns
// the grants in force afterwards. A failure part way, which only a store or
// model outage causes, leaves the changes made before it, each of them
// audited, and the error says what failed.
func (g *Grants) UpdateGrants(ctx context.Context, actor identity.Actor, holder identity.GrantHolder, add, remove []identity.GrantTarget) ([]identity.Grant, error) {
	add, err := g.checkUpdate(ctx, holder, add, remove)
	if err != nil {
		return nil, err
	}
	for _, t := range remove {
		if revokeErr := g.Revoke(ctx, actor, holder, t); revokeErr != nil && !errors.Is(revokeErr, ErrNotFound) {
			return nil, revokeErr
		}
	}
	for _, t := range add {
		if _, createErr := g.Create(ctx, actor, holder, t); createErr != nil {
			return nil, createErr
		}
	}
	return g.List(ctx, identity.GrantMatch{Holder: holder}, true)
}

// ViewPluginAccess returns the installed view plugins that read with their
// own identity, in name order, each with its grants in force (D-210). An
// error wrapping ErrDirectoryUnknown says the core can't read the plugins
// now.
func (g *Grants) ViewPluginAccess(ctx context.Context) ([]ViewPluginAccess, error) {
	names, err := g.dir.AccessViewPlugins(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing the view plugins: %w", err)
	}
	inForce, err := g.List(ctx, identity.GrantMatch{}, true)
	if err != nil {
		return nil, err
	}
	slices.Sort(names)
	out := make([]ViewPluginAccess, 0, len(names))
	for _, name := range names {
		out = append(out, ViewPluginAccess{Instance: name, Grants: grantsOf(inForce, identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: name})})
	}
	return out, nil
}

// Cutover is the step D-209 describes for the day SpiceDB takes over: every
// script token that may still be used and every installed view plugin that
// has no grant yet is given the household grant, so none loses or gains
// access silently, each with its own audit entry by actor, in one
// transaction with the record that the cutover ran. It runs once: a later
// call, a later start, changes nothing, so a token created afterwards, and
// given no grant, is not given one. A token or plugin that has a grant, or
// had one the owner revoked, is left as it is. It returns how many grants it
// made, and an error wrapping ErrDirectoryUnknown, with nothing done, when
// the installed plugins can't be listed yet.
func (g *Grants) Cutover(ctx context.Context, actor identity.Actor) (int, error) {
	done, err := g.store.GrantCutoverDone(ctx)
	if err != nil {
		return 0, fmt.Errorf("cutover: %w", err)
	}
	if done {
		return 0, nil
	}
	plugins, err := g.dir.ViewPlugins(ctx)
	if err != nil {
		return 0, fmt.Errorf("cutover: listing the view plugins: %w", err)
	}
	tokens, err := g.tokens.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("cutover: %w", err)
	}
	var made []identity.Grant
	err = within(ctx, g.store, func(ctx context.Context) error {
		first, markErr := g.store.MarkGrantCutover(ctx, g.now().UTC())
		if markErr != nil {
			return fmt.Errorf("recording the cutover: %w", markErr)
		}
		if !first {
			return nil
		}
		var grantErr error
		made, grantErr = g.grantHousehold(ctx, actor, g.holdersToGrant(tokens, plugins))
		return grantErr
	})
	if err != nil {
		return 0, fmt.Errorf("cutover: %w", err)
	}
	if projectErr := g.project(ctx, made); projectErr != nil {
		return len(made), projectErr
	}
	return len(made), nil
}

// Sweep ends, on behalf of actor, the grants in force that outlived the
// holder or the target they name: a script token that is unknown or was
// revoked, a view plugin that is no longer installed, an account or a sleeve
// that is no longer stored. The core removes them with the record as it goes,
// and this repairs what a crash or an outage in between left (D-209): an
// account or a sleeve created later under the same ID would otherwise inherit
// them. What the core can't tell is there now (ErrDirectoryUnknown) is left
// alone. It returns how many grants it ended.
func (g *Grants) Sweep(ctx context.Context, actor identity.Actor) (int, error) {
	inForce, err := g.InForce(ctx)
	if err != nil {
		return 0, err
	}
	gone, err := g.goneSides(ctx, inForce)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, side := range gone {
		n, err := g.revoke(ctx, actor, side)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// goneSides returns, in the order grants name them, each holder and each
// target of grants that is certainly gone, as a match that selects its
// grants.
func (g *Grants) goneSides(ctx context.Context, grants []identity.Grant) ([]identity.GrantMatch, error) {
	var gone []identity.GrantMatch
	seen := map[identity.GrantMatch]bool{}
	for i := range grants {
		for _, side := range []identity.GrantMatch{{Holder: grants[i].Holder}, {Target: grants[i].Target}} {
			if seen[side] {
				continue
			}
			seen[side] = true
			missing, err := g.missing(ctx, side)
			if err != nil {
				return nil, err
			}
			if missing {
				gone = append(gone, side)
			}
		}
	}
	return gone, nil
}

// revoke ends the grants in force that m selects: first in the authorization
// model, then, in one transaction, in the table with an audit entry each.
func (g *Grants) revoke(ctx context.Context, actor identity.Actor, m identity.GrantMatch) (int, error) {
	inForce, err := g.store.Grants(ctx, m, true)
	if err != nil {
		return 0, fmt.Errorf("loading the grants to revoke: %w", err)
	}
	if len(inForce) == 0 {
		return 0, nil
	}
	if unprojectErr := g.unproject(ctx, inForce); unprojectErr != nil {
		return 0, unprojectErr
	}
	var ended []identity.Grant
	err = within(ctx, g.store, func(ctx context.Context) error {
		var revokeErr error
		ended, revokeErr = g.revokeStored(ctx, actor, m)
		return revokeErr
	})
	if err != nil {
		return 0, fmt.Errorf("revoking: %w", err)
	}
	return len(ended), nil
}

// revokeStored ends the grants in force that m selects, with an audit entry
// each, inside the transaction ctx carries, and returns them as ended.
func (g *Grants) revokeStored(ctx context.Context, actor identity.Actor, m identity.GrantMatch) ([]identity.Grant, error) {
	ended, err := g.store.RevokeGrants(ctx, m, g.now().UTC())
	if err != nil {
		return nil, fmt.Errorf("ending the grants: %w", err)
	}
	for i := range ended {
		if recordErr := g.record(ctx, actor, identity.ActionGrantRevoked, &ended[i]); recordErr != nil {
			return nil, recordErr
		}
	}
	return ended, nil
}

// require checks that holder and target exist, as an error wrapping
// ErrNotFound when one is not there.
func (g *Grants) require(ctx context.Context, holder identity.GrantHolder, target identity.GrantTarget) error {
	ok, err := g.holderExists(ctx, holder)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: no such %s", ErrNotFound, holder.Kind)
	}
	if ok, err = g.dir.TargetExists(ctx, target); err != nil {
		return fmt.Errorf("looking for the %s: %w", target.Kind, err)
	}
	if !ok {
		return fmt.Errorf("%w: no such %s", ErrNotFound, target.Kind)
	}
	return nil
}

// holderExists reports whether holder is a script token that may still be
// used, or an installed view plugin.
func (g *Grants) holderExists(ctx context.Context, holder identity.GrantHolder) (bool, error) {
	switch holder.Kind {
	case identity.GrantHolderScriptToken:
		tok, err := g.tokens.Get(ctx, identity.ScriptTokenID(holder.ID))
		if errors.Is(err, ErrUnknownToken) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("looking for the script token: %w", err)
		}
		return tok.Status(g.now()) == "active", nil
	case identity.GrantHolderViewPlugin:
		ok, err := g.dir.ViewPluginInstalled(ctx, holder.ID)
		if err != nil {
			return false, fmt.Errorf("looking for the view plugin: %w", err)
		}
		return ok, nil
	case identity.GrantHolderPerson:
	}
	return false, nil
}

// record appends the entry that actor did action to grant, naming its holder
// and its target by ID, inside the transaction ctx carries.
func (g *Grants) record(ctx context.Context, actor identity.Actor, action identity.AuditAction, grant *identity.Grant) error {
	e, err := identity.NewAuditEntry(g.now(), actor, action, identity.OutcomeOK, grant.Holder.Ref(), grant.Target.Ref())
	if err != nil {
		return fmt.Errorf("building an audit entry: %w", err)
	}
	if appendErr := g.store.AppendAudit(ctx, &e); appendErr != nil {
		return fmt.Errorf("appending an audit entry: %w", appendErr)
	}
	return nil
}

// project writes grants to the authorization model, if there is one.
func (g *Grants) project(ctx context.Context, grants []identity.Grant) error {
	if g.proj == nil || len(grants) == 0 {
		return nil
	}
	if err := g.proj.ProjectGrants(ctx, grants); err != nil {
		return fmt.Errorf("%w: %w", ErrGrantNotProjected, err)
	}
	return nil
}

// unproject removes grants from the authorization model, if there is one.
func (g *Grants) unproject(ctx context.Context, grants []identity.Grant) error {
	if g.proj == nil || len(grants) == 0 {
		return nil
	}
	if err := g.proj.UnprojectGrants(ctx, grants); err != nil {
		return fmt.Errorf("removing from the authorization model: %w", err)
	}
	return nil
}

// holdersToGrant returns the script tokens that may still be used and the
// view plugins among instances, as holders.
func (g *Grants) holdersToGrant(tokens []identity.ScriptToken, instances []string) []identity.GrantHolder {
	now := g.now()
	var out []identity.GrantHolder
	for i := range tokens {
		if tokens[i].Status(now) == "active" {
			out = append(out, identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(tokens[i].ID)})
		}
	}
	for _, name := range instances {
		out = append(out, identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: name})
	}
	return out
}

// grantHousehold gives the household grant, with an audit entry, to each of
// holders that has no grant at all, in force or ended, inside the transaction
// ctx carries, and returns the grants it made.
func (g *Grants) grantHousehold(ctx context.Context, actor identity.Actor, holders []identity.GrantHolder) ([]identity.Grant, error) {
	have, err := g.store.Grants(ctx, identity.GrantMatch{}, false)
	if err != nil {
		return nil, fmt.Errorf("loading the grants: %w", err)
	}
	granted := make(map[identity.GrantHolder]bool, len(have))
	for i := range have {
		granted[have[i].Holder] = true
	}
	var made []identity.Grant
	for _, h := range holders {
		if granted[h] {
			continue
		}
		grant, err := identity.NewGrant(h, identity.Household(), g.now().UTC().Truncate(time.Microsecond))
		if err != nil {
			return nil, fmt.Errorf("building a grant: %w", err)
		}
		created, err := g.saveGrant(ctx, actor, &grant)
		if err != nil {
			return nil, err
		}
		if created {
			made = append(made, grant)
		}
		granted[h] = true
	}
	return made, nil
}

// missing reports whether the one holder or target m names is certainly gone.
func (g *Grants) missing(ctx context.Context, m identity.GrantMatch) (bool, error) {
	if m.Holder != (identity.GrantHolder{}) {
		return g.holderGone(ctx, m.Holder)
	}
	ok, err := g.dir.TargetExists(ctx, m.Target)
	if errors.Is(err, ErrDirectoryUnknown) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("looking for the %s: %w", m.Target.Kind, err)
	}
	return !ok, nil
}

// holderGone reports whether a holder is certainly gone: a script token
// unknown or revoked (an expired one is left, as it is no event), a view
// plugin not installed.
func (g *Grants) holderGone(ctx context.Context, h identity.GrantHolder) (bool, error) {
	switch h.Kind {
	case identity.GrantHolderScriptToken:
		tok, err := g.tokens.Get(ctx, identity.ScriptTokenID(h.ID))
		if errors.Is(err, ErrUnknownToken) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("looking for the script token: %w", err)
		}
		return tok.Status(g.now()) == "revoked", nil
	case identity.GrantHolderViewPlugin:
		ok, err := g.dir.ViewPluginInstalled(ctx, h.ID)
		if errors.Is(err, ErrDirectoryUnknown) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("looking for the view plugin: %w", err)
		}
		return !ok, nil
	case identity.GrantHolderPerson:
	}
	return false, nil
}

// saveGrant stores grant and, when it is new, the audit entry that actor made
// it, in one transaction, and reports whether it was new.
func (g *Grants) saveGrant(ctx context.Context, actor identity.Actor, grant *identity.Grant) (bool, error) {
	var created bool
	err := within(ctx, g.store, func(ctx context.Context) error {
		var saveErr error
		if created, saveErr = g.store.SaveGrant(ctx, grant); saveErr != nil {
			return fmt.Errorf("saving the grant: %w", saveErr)
		}
		if !created {
			return nil
		}
		return g.record(ctx, actor, identity.ActionGrantCreated, grant)
	})
	if err != nil {
		return false, fmt.Errorf("granting: %w", err)
	}
	return created, nil
}

// inForceOf returns the one grant of target to holder in force, or an error
// wrapping ErrNotFound when it ended meanwhile.
func (g *Grants) inForceOf(ctx context.Context, holder identity.GrantHolder, target identity.GrantTarget) (identity.Grant, error) {
	inForce, err := g.store.Grants(ctx, identity.GrantMatch{Holder: holder, Target: target}, true)
	if err != nil {
		return identity.Grant{}, fmt.Errorf("loading the grant in force: %w", err)
	}
	if len(inForce) != 1 {
		return identity.Grant{}, fmt.Errorf("%w: the grant ended meanwhile", ErrNotFound)
	}
	return inForce[0], nil
}

// checkTargets validates each of targets, drops a repeated one, and checks
// that it exists: an invalid one wraps ErrInvalidRequest, a missing one
// ErrNotFound.
func (g *Grants) checkTargets(ctx context.Context, targets []identity.GrantTarget) ([]identity.GrantTarget, error) {
	var out []identity.GrantTarget
	for _, t := range targets {
		if err := t.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
		}
		if slices.Contains(out, t) {
			continue
		}
		ok, err := g.dir.TargetExists(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("looking for the %s: %w", t.Kind, err)
		}
		if !ok {
			return nil, fmt.Errorf("%w: no such %s", ErrNotFound, t.Kind)
		}
		out = append(out, t)
	}
	return out, nil
}

// checkUpdate checks what UpdateGrants is asked: the holder is valid, may hold
// grants and exists, every target to remove is valid and every target to add
// exists. It returns the targets to add without repeats.
func (g *Grants) checkUpdate(ctx context.Context, holder identity.GrantHolder, add, remove []identity.GrantTarget) ([]identity.GrantTarget, error) {
	if err := holder.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	if holder.Kind == identity.GrantHolderPerson {
		return nil, fmt.Errorf("%w: viewers don't exist yet", ErrInvalidRequest)
	}
	ok, err := g.holderExists(ctx, holder)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%w: no such %s", ErrNotFound, holder.Kind)
	}
	for _, t := range remove {
		if validErr := t.Validate(); validErr != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, validErr)
		}
	}
	return g.checkTargets(ctx, add)
}

// storeToken stores the token spec describes, a grant of each of targets and
// the audit entry of each, in one transaction, and returns them with the
// token's secret.
func (g *Grants) storeToken(ctx context.Context, actor identity.Actor, spec *NewToken, targets []identity.GrantTarget) (identity.ScriptToken, string, []identity.Grant, error) {
	var (
		tok    identity.ScriptToken
		secret string
		grants []identity.Grant
	)
	err := within(ctx, g.store, func(ctx context.Context) error {
		var issueErr error
		if tok, secret, issueErr = IssueTokenWith(ctx, g.store, g.tokens, actor, spec.Name, spec.TTL, spec.MCPAllowed); issueErr != nil {
			return issueErr
		}
		holder := identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: string(tok.ID)}
		for _, target := range targets {
			grant, grantErr := identity.NewGrant(holder, target, g.now().UTC().Truncate(time.Microsecond))
			if grantErr != nil {
				return fmt.Errorf("building a grant: %w", grantErr)
			}
			if _, saveErr := g.store.SaveGrant(ctx, &grant); saveErr != nil {
				return fmt.Errorf("saving the grant: %w", saveErr)
			}
			if recordErr := g.record(ctx, actor, identity.ActionGrantCreated, &grant); recordErr != nil {
				return recordErr
			}
			grants = append(grants, grant)
		}
		return nil
	})
	if err != nil {
		return identity.ScriptToken{}, "", nil, err
	}
	return tok, secret, grants, nil
}
