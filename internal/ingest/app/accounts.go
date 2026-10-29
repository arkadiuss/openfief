package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
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

// NewAccount is an account the owner creates (D-72): its name,
// institution and sleeve, and the connector instance that feeds it with,
// for a broker connector, the broker account it is.
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
}

// Accounts are the owner's account entries (D-72, D-73): creating and
// changing accounts and binding the connector instances that feed them,
// each for an owner-scoped session only (D-79). An account's ID is made
// from its name (portfolio.KeyFromName) and never changes. Each step is
// checked before anything is stored, and each store is idempotent, so a
// failure between them leaves nothing a retry can't finish.
type Accounts struct {
	Owners   Owners
	Ledger   AccountLedger
	Bindings BindingStore
	Brokers  BrokerAccountStore
	Now      func() time.Time
}

// Create creates the account n names, binds its instance to it and maps its
// broker account to it. An empty institution takes the broker account's.
// It returns the account as stored.
func (a *Accounts) Create(ctx context.Context, origin, sessionToken string, n *NewAccount) (ledger.Account, error) {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return ledger.Account{}, fmt.Errorf("creating an account: %w", err)
	}
	id, err := portfolio.KeyFromName(n.Name)
	if err != nil {
		return ledger.Account{}, fmt.Errorf("creating an account: %w: %w", ledger.ErrInvalidAccount, err)
	}
	accounts, err := a.Ledger.Accounts(ctx)
	if err != nil {
		return ledger.Account{}, fmt.Errorf("creating an account: %w", err)
	}
	if hasAccount(accounts, ledger.AccountID(id)) {
		return ledger.Account{}, fmt.Errorf("creating account %s: %w", id, ErrAccountExists)
	}
	broker, err := a.brokerAccount(ctx, n.Instance, n.BrokerAccount, ledger.AccountID(id))
	if err != nil {
		return ledger.Account{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	institution := n.Institution
	if institution == "" {
		institution = broker.Institution
	}
	acc := ledger.Account{ID: ledger.AccountID(id), Institution: string(institution), Name: n.Name, Sleeve: n.Sleeve}
	if err := a.save(ctx, acc); err != nil {
		return ledger.Account{}, fmt.Errorf("creating account %s: %w", id, err)
	}
	if n.Instance != "" {
		if err := a.connect(ctx, acc.ID, n.Instance, n.BrokerAccount); err != nil {
			return ledger.Account{}, fmt.Errorf("creating account %s: %w", id, err)
		}
	}
	return acc, nil
}

// Update renames the stored account id and moves it to sleeve; its ID and
// institution stay. It returns the account as stored.
func (a *Accounts) Update(ctx context.Context, origin, sessionToken string, id ledger.AccountID, name string, sleeve portfolio.SleeveID) (ledger.Account, error) {
	if err := a.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return ledger.Account{}, fmt.Errorf("changing an account: %w", err)
	}
	acc, err := a.stored(ctx, id)
	if err != nil {
		return ledger.Account{}, err
	}
	acc.Name, acc.Sleeve = name, sleeve
	if err := a.save(ctx, acc); err != nil {
		return ledger.Account{}, fmt.Errorf("changing account %s: %w", id, err)
	}
	return acc, nil
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
