package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Errors of the owner's account entries.
var (
	// ErrAccountExists reports an account whose ID, made from its name, is
	// taken.
	ErrAccountExists = errors.New("an account with this ID exists")
	// ErrNoAccount reports an account to change that is not stored.
	ErrNoAccount = errors.New("no such account")
	// ErrNoSleeve reports an account in a sleeve that is not stored.
	ErrNoSleeve = errors.New("no such sleeve")
	// ErrBrokerAccountNeedsInstance reports a broker account given without
	// the connector instance that reported it.
	ErrBrokerAccountNeedsInstance = errors.New("a broker account needs the connector instance that reported it")
)

// AccountLedger is what the owner's account entries read from and write
// to the ledger. *ledger/app.Ledger implements it.
type AccountLedger interface {
	// Accounts returns every account, in ID order.
	Accounts(ctx context.Context) ([]ledger.Account, error)
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// SaveAccount creates a or replaces the account with its ID.
	SaveAccount(ctx context.Context, a ledger.Account) error
}

// AccountTerms keep an account's tax wrapper and fee rule (D-76, D-78),
// which the valuation context values it with: recording either revalues
// the account. cmd/core joins valuation's InputRecorder and store into it.
type AccountTerms interface {
	// RecordTaxWrapper records account's tax wrapper.
	RecordTaxWrapper(ctx context.Context, account ledger.AccountID, wrapper jurisdiction.Wrapper) error
	// RecordFeeRule records account's fee rule.
	RecordFeeRule(ctx context.Context, account ledger.AccountID, rule *jurisdiction.FeeRule) error
	// TaxWrapper returns account's recorded tax wrapper, and false when
	// none was recorded.
	TaxWrapper(ctx context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error)
	// FeeRule returns account's fee rule, the zero rule unless recorded.
	FeeRule(ctx context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error)
}

// AccountEntry is an account with the terms it is valued under: its tax
// wrapper and fee rule.
type AccountEntry struct {
	ledger.Account

	Wrapper jurisdiction.Wrapper
	Fee     jurisdiction.FeeRule
}

// AccountChange is what the owner changes in an account: its name and
// sleeve, and its tax wrapper and fee rule when they are given.
type AccountChange struct {
	Name   string
	Sleeve portfolio.SleeveID
	// Wrapper is the new tax wrapper; empty keeps the account's.
	Wrapper jurisdiction.Wrapper
	// Fee is the new fee rule; nil keeps the account's.
	Fee *jurisdiction.FeeRule
}

// NewAccount is an account the owner creates (D-72): its name,
// institution and sleeve, its tax wrapper and fee rule (D-76, D-78), and
// the connector instance that feeds it with, for a broker connector, the
// broker account it is.
type NewAccount struct {
	Name        string
	Institution ingest.InstitutionID
	Sleeve      portfolio.SleeveID
	// Instance is the connector instance that feeds the account; empty for
	// none yet.
	Instance string
	// BrokerAccount is one of the broker accounts Instance reported;
	// empty for a connector that names OpenFief accounts itself.
	BrokerAccount ingest.BrokerAccountID
	// Wrapper is the account's tax wrapper; empty takes the pack's
	// default.
	Wrapper jurisdiction.Wrapper
	// Fee is the account's fee rule; nil for the zero rule.
	Fee *jurisdiction.FeeRule
}

// Accounts are the owner's account entries (D-72, D-73): creating and
// changing accounts, with their tax wrapper and fee rule (D-76, D-78), and
// binding the connector instances that feed them, each for an
// owner-scoped session only (D-79). The wrappers are Pack's. An account's ID is made
// from its name (portfolio.KeyFromName) and never changes. Each step is
// checked before anything is stored, and each store is idempotent, so a
// failure between them leaves nothing a retry can't finish.
type Accounts struct {
	Owners   Owners
	Ledger   AccountLedger
	Bindings BindingStore
	Brokers  BrokerAccountStore
	Terms    AccountTerms
	Pack     jurisdiction.Pack
	Now      func() time.Time
}

// Create creates the account n names, records its tax wrapper, the pack's
// default when n gives none, and its fee rule, the zero rule when n gives
// none, binds its instance to it and maps its broker account to it. An
// empty institution takes the broker account's. It returns the account as
// stored.
func (a *Accounts) Create(ctx context.Context, origin, sessionToken string, n *NewAccount) (AccountEntry, error) {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return AccountEntry{}, fmt.Errorf("creating an account: %w", err)
	}
	id, err := portfolio.KeyFromName(n.Name)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("creating an account: %w: %w", ledger.ErrInvalidAccount, err)
	}
	entry, err := a.newEntry(n)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	accounts, err := a.Ledger.Accounts(ctx)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("creating an account: %w", err)
	}
	if hasAccount(accounts, ledger.AccountID(id)) {
		return AccountEntry{}, fmt.Errorf("creating account %s: %w", id, ErrAccountExists)
	}
	broker, err := a.brokerAccount(ctx, n.Instance, n.BrokerAccount, ledger.AccountID(id))
	if err != nil {
		return AccountEntry{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	institution := n.Institution
	if institution == "" {
		institution = broker.Institution
	}
	entry.Account = ledger.Account{ID: ledger.AccountID(id), Institution: string(institution), Name: n.Name, Sleeve: n.Sleeve}
	if err := a.storeNew(ctx, &entry, n); err != nil {
		return AccountEntry{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	return entry, nil
}

// Update renames the stored account id and moves it to a sleeve, and
// records its tax wrapper and fee rule when c changes them; its ID and
// institution stay. It returns the account as stored.
func (a *Accounts) Update(ctx context.Context, origin, sessionToken string, id ledger.AccountID, c *AccountChange) (AccountEntry, error) {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return AccountEntry{}, fmt.Errorf("changing an account: %w", err)
	}
	acc, err := a.stored(ctx, id)
	if err != nil {
		return AccountEntry{}, err
	}
	entry, err := a.entry(ctx, acc)
	if err != nil {
		return AccountEntry{}, err
	}
	if c.Wrapper != "" {
		entry.Wrapper = c.Wrapper
	}
	if c.Fee != nil {
		entry.Fee = *c.Fee
	}
	if err := a.checkTerms(entry.Wrapper, &entry.Fee); err != nil {
		return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	entry.Name, entry.Sleeve = c.Name, c.Sleeve
	if err := a.save(ctx, entry.Account); err != nil {
		return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	if err := a.recordTerms(ctx, id, c.Wrapper, c.Fee); err != nil {
		return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	return entry, nil
}

// Bind binds instance to the stored account id, as `core connector bind`
// does, and maps broker, one of the broker accounts instance reported, to
// it when broker is not empty (D-72). Binding again changes nothing.
func (a *Accounts) Bind(ctx context.Context, origin, sessionToken string, id ledger.AccountID, instance string, broker ingest.BrokerAccountID) error {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("binding an account: %w", err)
	}
	if _, err := a.stored(ctx, id); err != nil {
		return err
	}
	if instance == "" {
		return fmt.Errorf("binding account %s: %w", id, ingest.ErrInvalidInstanceName)
	}
	if _, err := a.brokerAccount(ctx, instance, broker, id); err != nil {
		return fmt.Errorf("binding account %s: %w", id, err)
	}
	if err := a.connect(ctx, id, instance, broker); err != nil {
		return fmt.Errorf("binding account %s: %w", id, err)
	}
	return nil
}

// storeNew saves the new account entry, records its terms, and binds n's
// instance to it with its broker account.
func (a *Accounts) storeNew(ctx context.Context, entry *AccountEntry, n *NewAccount) error {
	if err := a.save(ctx, entry.Account); err != nil {
		return err
	}
	if err := a.recordTerms(ctx, entry.ID, entry.Wrapper, n.Fee); err != nil {
		return err
	}
	if n.Instance == "" {
		return nil
	}
	return a.connect(ctx, entry.ID, n.Instance, n.BrokerAccount)
}

// entry returns account acc with its tax wrapper, the pack's
// default unless one was recorded, and its fee rule.
func (a *Accounts) entry(ctx context.Context, acc ledger.Account) (AccountEntry, error) {
	w, recorded, err := a.Terms.TaxWrapper(ctx, acc.ID)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("loading the tax wrapper of %s: %w", acc.ID, err)
	}
	if !recorded {
		w = jurisdiction.DefaultWrapper(a.Pack)
	}
	fee, err := a.Terms.FeeRule(ctx, acc.ID)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("loading the fee rule of %s: %w", acc.ID, err)
	}
	return AccountEntry{Account: acc, Wrapper: w, Fee: fee}, nil
}

// newEntry returns the terms n gives a new account, checked: its wrapper,
// the pack's default when n gives none, and its fee rule.
func (a *Accounts) newEntry(n *NewAccount) (AccountEntry, error) {
	entry := AccountEntry{Wrapper: n.Wrapper}
	if entry.Wrapper == "" {
		entry.Wrapper = jurisdiction.DefaultWrapper(a.Pack)
	}
	if n.Fee != nil {
		entry.Fee = *n.Fee
	}
	if err := a.checkTerms(entry.Wrapper, &entry.Fee); err != nil {
		return AccountEntry{}, err
	}
	return entry, nil
}

// checkTerms checks that w is one of the pack's wrappers and fee a valid
// rule, as invalid accounts otherwise.
func (a *Accounts) checkTerms(w jurisdiction.Wrapper, fee *jurisdiction.FeeRule) error {
	if !jurisdiction.HasWrapper(a.Pack, w) {
		return fmt.Errorf("%w: %w: %q", ledger.ErrInvalidAccount, jurisdiction.ErrUnknownWrapper, w)
	}
	if err := fee.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ledger.ErrInvalidAccount, err)
	}
	return nil
}

// recordTerms records account's wrapper unless it is empty and its fee
// rule unless it is nil, each revaluing the account.
func (a *Accounts) recordTerms(ctx context.Context, account ledger.AccountID, w jurisdiction.Wrapper, fee *jurisdiction.FeeRule) error {
	if w != "" {
		if err := a.Terms.RecordTaxWrapper(ctx, account, w); err != nil {
			return fmt.Errorf("recording the tax wrapper: %w", err)
		}
	}
	if fee != nil {
		if err := a.Terms.RecordFeeRule(ctx, account, fee); err != nil {
			return fmt.Errorf("recording the fee rule: %w", err)
		}
	}
	return nil
}

// stored returns the stored account id, or an error wrapping ErrNoAccount.
func (a *Accounts) stored(ctx context.Context, id ledger.AccountID) (ledger.Account, error) {
	accounts, err := a.Ledger.Accounts(ctx)
	if err != nil {
		return ledger.Account{}, fmt.Errorf("loading accounts: %w", err)
	}
	i := slices.IndexFunc(accounts, func(acc ledger.Account) bool { return acc.ID == id })
	if i < 0 {
		return ledger.Account{}, fmt.Errorf("account %q: %w", id, ErrNoAccount)
	}
	return accounts[i], nil
}

// save checks acc and its sleeve, then stores it.
func (a *Accounts) save(ctx context.Context, acc ledger.Account) error {
	if err := acc.Validate(); err != nil {
		return fmt.Errorf("checking the account: %w", err)
	}
	sleeves, err := a.Ledger.Sleeves(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeves: %w", err)
	}
	if !slices.ContainsFunc(sleeves, func(s portfolio.Sleeve) bool { return s.ID == acc.Sleeve }) {
		return fmt.Errorf("sleeve %q: %w", acc.Sleeve, ErrNoSleeve)
	}
	if err := a.Ledger.SaveAccount(ctx, acc); err != nil {
		return fmt.Errorf("saving the account: %w", err)
	}
	return nil
}

// brokerAccount checks instance's name and, when broker is not empty, that
// instance reported broker and it is unmapped or mapped to account, and
// returns it. Without broker it returns the zero BrokerAccount.
func (a *Accounts) brokerAccount(ctx context.Context, instance string, broker ingest.BrokerAccountID, account ledger.AccountID) (ingest.BrokerAccount, error) {
	if instance != "" {
		if err := ingest.ValidateInstanceName(instance); err != nil {
			return ingest.BrokerAccount{}, fmt.Errorf("connector instance: %w", err)
		}
	}
	if broker == "" {
		return ingest.BrokerAccount{}, nil
	}
	if instance == "" {
		return ingest.BrokerAccount{}, ErrBrokerAccountNeedsInstance
	}
	known, err := a.Brokers.BrokerAccounts(ctx, instance)
	if err != nil {
		return ingest.BrokerAccount{}, fmt.Errorf("loading the broker accounts of %s: %w", instance, err)
	}
	i := slices.IndexFunc(known, func(b ingest.BrokerAccount) bool { return b.ID == broker })
	switch {
	case i < 0:
		return ingest.BrokerAccount{}, fmt.Errorf("connector instance %s: %w", instance, ErrUnknownBrokerAccount)
	case known[i].Mapped() && known[i].Account != account:
		return ingest.BrokerAccount{}, fmt.Errorf("connector instance %s: %w", instance, ErrBrokerAccountMapped)
	}
	return known[i], nil
}

// connect binds instance to account and maps broker to it, when given.
func (a *Accounts) connect(ctx context.Context, account ledger.AccountID, instance string, broker ingest.BrokerAccountID) error {
	at := a.Now()
	if err := a.Bindings.Bind(ctx, instance, account, at); err != nil {
		return fmt.Errorf("binding %s: %w", instance, err)
	}
	if broker == "" {
		return nil
	}
	if err := a.Brokers.MapBrokerAccount(ctx, instance, broker, account, at); err != nil {
		return fmt.Errorf("mapping a broker account of %s: %w", instance, err)
	}
	return nil
}

// hasAccount reports whether accounts hold id.
func hasAccount(accounts []ledger.Account, id ledger.AccountID) bool {
	return slices.ContainsFunc(accounts, func(acc ledger.Account) bool { return acc.ID == id })
}
