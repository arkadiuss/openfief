package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// ErrUnknownSleeve reports an account whose sleeve is not stored.
var ErrUnknownSleeve = errors.New("unknown sleeve")

// ErrUnknownAccount reports a transaction, balance assertion or query for an
// account that is not stored.
var ErrUnknownAccount = errors.New("unknown account")

// Ledger runs the ledger context's use cases over a Store.
type Ledger struct {
	store Store
}

// New returns the use cases over store.
func New(store Store) *Ledger {
	return &Ledger{store: store}
}

// SaveSleeve validates s and stores it, replacing a sleeve with its ID.
func (l *Ledger) SaveSleeve(ctx context.Context, s portfolio.Sleeve) error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("saving sleeve: %w", err)
	}
	if err := l.store.SaveSleeve(ctx, s); err != nil {
		return fmt.Errorf("saving sleeve %s: %w", s.ID, err)
	}
	return nil
}

// SaveAccount validates a, checks that its sleeve is stored, and stores it,
// replacing an account with its ID.
func (l *Ledger) SaveAccount(ctx context.Context, a ledger.Account) error {
	if err := a.Validate(); err != nil {
		return fmt.Errorf("saving account: %w", err)
	}
	sleeves, err := l.store.Sleeves(ctx)
	if err != nil {
		return fmt.Errorf("loading sleeves: %w", err)
	}
	if !hasSleeve(sleeves, a.Sleeve) {
		return fmt.Errorf("%w %s for account %s", ErrUnknownSleeve, a.Sleeve, a.ID)
	}
	if err := l.store.SaveAccount(ctx, a); err != nil {
		return fmt.Errorf("saving account %s: %w", a.ID, err)
	}
	return nil
}

// hasSleeve reports whether sleeves include id.
func hasSleeve(sleeves []portfolio.Sleeve, id portfolio.SleeveID) bool {
	for i := range sleeves {
		if sleeves[i].ID == id {
			return true
		}
	}
	return false
}

// account returns the stored account id, or an error wrapping
// ErrUnknownAccount.
func (l *Ledger) account(ctx context.Context, id ledger.AccountID) (ledger.Account, error) {
	a, err := l.store.Account(ctx, id)
	if errors.Is(err, ErrNotFound) {
		return ledger.Account{}, fmt.Errorf("%w: %s", ErrUnknownAccount, id)
	}
	if err != nil {
		return ledger.Account{}, fmt.Errorf("loading account %s: %w", id, err)
	}
	return a, nil
}
