package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
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
	// SaveAccount creates a or replaces the account with its ID.
	SaveAccount(ctx context.Context, a ledger.Account) error
}

// AccountSleeves is what the owner's account entries read from and write
// to the portfolio context, which keeps which account is in which sleeve.
// *portfolio/app.Registry implements it.
type AccountSleeves interface {
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// AccountSleeves returns the sleeve of every account in one.
	AccountSleeves(ctx context.Context) (map[ledger.AccountID]portfolio.SleeveID, error)
	// PlaceAccount puts the stored account in sleeve, replacing the sleeve
	// it was in.
	PlaceAccount(ctx context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error
	// RemoveAccount takes a deleted account out of its sleeve.
	RemoveAccount(ctx context.Context, account ledger.AccountID) error
}

// PayoutSettings keeps each account's payout setting (D-135) with the
// settings in force before it (D-162), in the ingest schema: an account
// setting the treasury-bonds connector reads. ingest/postgres.Store
// implements it.
type PayoutSettings interface {
	PayoutSettingReader
	// SavePayoutSetting stores account's payout setting, replacing the
	// stored one.
	SavePayoutSetting(ctx context.Context, account ledger.AccountID, p ingest.PayoutSetting) error
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

// AccountEntry is an account with the sleeve the portfolio context puts it
// in, its payout setting, and the terms it is valued under: its tax
// wrapper and fee rule.
type AccountEntry struct {
	ledger.Account

	// Sleeve is the sleeve the account is in; empty for none.
	Sleeve portfolio.SleeveID
	// Payout is the account's payout setting (D-135, D-162); the zero
	// setting keeps its payouts as cash.
	Payout  ingest.PayoutSetting
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
	// Payout is the new payout setting (D-135); nil keeps the account's.
	Payout *ingest.Payout
	// PayoutFrom is the day Payout takes effect from (D-162): today or an
	// earlier day; zero is today.
	PayoutFrom kernel.Date
	// OpenedOnSeedingDay marks or unmarks the account as opened on its
	// seeding day (D-168); nil keeps the account's mark.
	OpenedOnSeedingDay *bool
}

// PayoutLedger is what a change of an account's payout setting reads from
// and writes to the ledger (D-162): the account's current transactions,
// and the import of the deletions that supersede the payout withdrawals the
// new setting no longer makes. cmd/core implements it over ledger/app.Ledger.
type PayoutLedger interface {
	CurrentTransactions(ctx context.Context, account ledger.AccountID) ([]ledger.Transaction, error)
	Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error)
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
	// Payout is the account's payout setting (D-135); nil keeps its
	// payouts as cash.
	Payout *ingest.Payout
	// OpenedOnSeedingDay marks the account as opened on its seeding day
	// (D-168), so its seeding doesn't move the portfolio's charts' default
	// start (D-138).
	OpenedOnSeedingDay bool
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
	// Deletions, Connections and Forget delete an account and remap a broker
	// account (D-120) and list the broker accounts for the owner (D-121).
	Deletions   AccountDeletions
	Connections AccountConnections
	Forget      TermsForgetter
	// Sleeves keeps which sleeve each account is in, in the portfolio
	// context.
	Sleeves AccountSleeves
	// PayoutSettings keeps each account's payout setting (D-135, D-162).
	PayoutSettings PayoutSettings
	// Payouts supersedes the payout withdrawals a changed payout setting no
	// longer makes (D-162); nil supersedes none.
	Payouts PayoutLedger
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
	entry.Account = ledger.Account{
		ID: ledger.AccountID(id), Institution: string(institution), Name: n.Name, OpenedOnSeedingDay: n.OpenedOnSeedingDay,
	}
	entry.Sleeve = n.Sleeve
	if entry.Payout, err = a.firstPayout(institution, n.Payout); err != nil {
		return AccountEntry{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	if err := a.storeNew(ctx, &entry, n); err != nil {
		return AccountEntry{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	return entry, nil
}

// Update renames the stored account id and moves it to a sleeve, marks or
// unmarks it as opened on its seeding day when c says (D-168), and
// records its tax wrapper and fee rule when c changes them, and its payout
// setting (D-135), in force from the day c gives, today or earlier, today
// when it gives none (D-162). The payout withdrawals booked from that day on
// that the new setting no longer makes are then superseded
// (ingest.StalePayouts), and the next sync books the coupons and
// maturities from that day anew. Its ID and institution stay. It returns
// the account as stored.
func (a *Accounts) Update(ctx context.Context, origin, sessionToken string, id ledger.AccountID, c *AccountChange) (AccountEntry, error) {
	if err := mayChange(ctx, a.Owners, origin, sessionToken, id); err != nil {
		return AccountEntry{}, fmt.Errorf("changing an account: %w", err)
	}
	acc, err := a.stored(ctx, id)
	if err != nil {
		return AccountEntry{}, err
	}
	entry, from, err := a.changed(ctx, acc, c)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	if err := a.save(ctx, &entry, c.Payout != nil); err != nil {
		return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	if err := a.recordTerms(ctx, id, c.Wrapper, c.Fee); err != nil {
		return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	if c.Payout != nil {
		if err := a.supersedePayouts(ctx, &entry, from); err != nil {
			return AccountEntry{}, fmt.Errorf("changing account %s: %w", id, err)
		}
	}
	return entry, nil
}

// Bind binds instance to the stored account id, as `core connector bind`
// does, and maps broker, one of the broker accounts instance reported, to
// it when broker is not empty (D-72). Binding again changes nothing.
func (a *Accounts) Bind(ctx context.Context, origin, sessionToken string, id ledger.AccountID, instance string, broker ingest.BrokerAccountID) error {
	if err := mayChange(ctx, a.Owners, origin, sessionToken, id); err != nil {
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
	if err := a.save(ctx, entry, n.Payout != nil); err != nil {
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

// entry returns account acc with its sleeve, its payout setting, its tax
// wrapper, the pack's default unless one was recorded, and its fee rule.
func (a *Accounts) entry(ctx context.Context, acc ledger.Account) (AccountEntry, error) {
	sleeves, err := a.Sleeves.AccountSleeves(ctx)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("loading the sleeve of %s: %w", acc.ID, err)
	}
	payout, err := a.PayoutSettings.PayoutSetting(ctx, acc.ID)
	if err != nil {
		return AccountEntry{}, fmt.Errorf("loading the payout setting of %s: %w", acc.ID, err)
	}
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
	return AccountEntry{Account: acc, Sleeve: sleeves[acc.ID], Payout: payout, Wrapper: w, Fee: fee}, nil
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

// firstPayout returns the payout setting p of a new account at
// institution, which covers its whole history (D-162), or the default, as
// cash, for nil p, checked as payout checks it.
func (a *Accounts) firstPayout(institution ingest.InstitutionID, p *ingest.Payout) (ingest.PayoutSetting, error) {
	if p == nil {
		return ingest.PayoutSetting{}, nil
	}
	if err := ingest.CheckPayout(institution, *p); err != nil {
		return ingest.PayoutSetting{}, fmt.Errorf("%w: %w", ledger.ErrInvalidAccount, err)
	}
	return ingest.PayoutSetting{Payout: *p}, nil
}

// payout returns entry's payout setting with p in force from the day from
// (D-162), or its own for nil p, checked: only a hand-entered bond
// account may pay out (ingest.CheckPayout), as an invalid account
// otherwise.
func (a *Accounts) payout(entry *AccountEntry, p *ingest.Payout, from kernel.Date) (ingest.PayoutSetting, error) {
	if p == nil {
		return entry.Payout, nil
	}
	if err := ingest.CheckPayout(ingest.InstitutionID(entry.Institution), *p); err != nil {
		return ingest.PayoutSetting{}, fmt.Errorf("%w: %w", ledger.ErrInvalidAccount, err)
	}
	return entry.Payout.Changed(*p, from), nil
}

// payoutFrom returns the day a payout change takes effect from: from, or
// today when it is zero, and never after today (ingest.CheckPayoutFrom),
// as an invalid account otherwise.
func (a *Accounts) payoutFrom(from kernel.Date) (kernel.Date, error) {
	now := a.Now().UTC()
	today, err := kernel.NewDate(now.Year(), now.Month(), now.Day())
	if err != nil {
		return kernel.Date{}, fmt.Errorf("today: %w", err)
	}
	if from.IsZero() {
		return today, nil
	}
	if err := ingest.CheckPayoutFrom(from, today); err != nil {
		return kernel.Date{}, fmt.Errorf("%w: %w", ledger.ErrInvalidAccount, err)
	}
	return from, nil
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

// save checks entry's account, its sleeve and its payout setting, then
// stores each with the context that owns it: the account in the ledger,
// its sleeve in the portfolio context, and its payout setting here when
// withPayout says it was given. An account is in one sleeve, so an empty
// sleeve is an invalid account and one not stored fails with ErrNoSleeve.
func (a *Accounts) save(ctx context.Context, entry *AccountEntry, withPayout bool) error {
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("checking the account: %w", err)
	}
	if entry.Sleeve == "" {
		return fmt.Errorf("checking the account: %w %s: no sleeve", ledger.ErrInvalidAccount, entry.ID)
	}
	if err := entry.Payout.Validate(); err != nil {
		return fmt.Errorf("checking the account: %w %s: %w", ledger.ErrInvalidAccount, entry.ID, err)
	}
	sleeves, err := a.Sleeves.Sleeves(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeves: %w", err)
	}
	if !slices.ContainsFunc(sleeves, func(s portfolio.Sleeve) bool { return s.ID == entry.Sleeve }) {
		return fmt.Errorf("sleeve %q: %w", entry.Sleeve, ErrNoSleeve)
	}
	if err := a.Ledger.SaveAccount(ctx, entry.Account); err != nil {
		return fmt.Errorf("saving the account: %w", err)
	}
	if err := a.Sleeves.PlaceAccount(ctx, entry.ID, entry.Sleeve); err != nil {
		return fmt.Errorf("placing the account in its sleeve: %w", err)
	}
	if !withPayout {
		return nil
	}
	if err := a.PayoutSettings.SavePayoutSetting(ctx, entry.ID, entry.Payout); err != nil {
		return fmt.Errorf("saving the payout setting: %w", err)
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

// supersedePayouts imports the deletions of the payout withdrawals of acc dated
// on or after from that its payout setting no longer makes
// (ingest.StalePayouts); without Payouts, none. Running it again finds
// none left.
func (a *Accounts) supersedePayouts(ctx context.Context, acc *AccountEntry, from kernel.Date) error {
	if a.Payouts == nil {
		return nil
	}
	current, err := a.Payouts.CurrentTransactions(ctx, acc.ID)
	if err != nil {
		return fmt.Errorf("loading transactions: %w", err)
	}
	stale, err := ingest.StalePayouts(current, acc.Payout, from)
	if err != nil {
		return fmt.Errorf("superseding payout withdrawals: %w", err)
	}
	if len(stale) == 0 {
		return nil
	}
	if _, err := a.Payouts.Import(ctx, stale, nil); err != nil {
		return fmt.Errorf("superseding payout withdrawals: %w", err)
	}
	return nil
}

// changed returns account acc as c changes it, checked, with the day a
// payout change takes effect from: its name and sleeve, its tax wrapper
// and fee rule when c gives them, its mark of being opened on its seeding
// day when c gives one (D-168), and its payout setting (payout).
func (a *Accounts) changed(ctx context.Context, acc ledger.Account, c *AccountChange) (AccountEntry, kernel.Date, error) {
	entry, entryErr := a.entry(ctx, acc)
	if entryErr != nil {
		return AccountEntry{}, kernel.Date{}, entryErr
	}
	if c.Wrapper != "" {
		entry.Wrapper = c.Wrapper
	}
	if c.Fee != nil {
		entry.Fee = *c.Fee
	}
	if err := a.checkTerms(entry.Wrapper, &entry.Fee); err != nil {
		return AccountEntry{}, kernel.Date{}, err
	}
	entry.Name, entry.Sleeve = c.Name, c.Sleeve
	if c.OpenedOnSeedingDay != nil {
		entry.OpenedOnSeedingDay = *c.OpenedOnSeedingDay
	}
	from, err := a.payoutFrom(c.PayoutFrom)
	if err != nil {
		return AccountEntry{}, kernel.Date{}, err
	}
	if entry.Payout, err = a.payout(&entry, c.Payout, from); err != nil {
		return AccountEntry{}, kernel.Date{}, err
	}
	return entry, from, nil
}
