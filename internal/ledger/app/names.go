package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// InstrumentNameStore keeps instruments' display names (D-197). It only
// ever appends: a rename is a new record, never a change of an old one.
type InstrumentNameStore interface {
	// AppendInstrumentNames stores names in their order: the owner's
	// always, a connector's only while no connector's name of its
	// instrument is stored yet, so the first delivered name stays.
	AppendInstrumentNames(ctx context.Context, names []ledger.InstrumentName) error
	// InstrumentNames returns every stored name, in the order stored.
	InstrumentNames(ctx context.Context) ([]ledger.InstrumentName, error)
}

// InstrumentPostings reads which accounts post to an instrument;
// *Ledger implements it.
type InstrumentPostings interface {
	// Accounts returns every account, in ID order.
	Accounts(ctx context.Context) ([]ledger.Account, error)
	// CurrentTransactions returns the current version of each of
	// account's transactions.
	CurrentTransactions(ctx context.Context, account ledger.AccountID) ([]ledger.Transaction, error)
}

// Names are instruments' display names (D-197): the first name a
// connector delivers for each, and the owner's renames, which only an
// owner-scoped session may make (D-79), asking the authorizer for every
// account that posts to the instrument (D-163).
type Names struct {
	Store    InstrumentNameStore
	Owners   Owners
	Postings InstrumentPostings
}

// Deliver records the names a connector delivered, from source, such as its
// instance: the first valid one of each instrument (ledger.FirstDeliveredNames),
// which the store keeps only while the instrument has no delivered name
// yet. A name that isn't valid is left out.
func (n *Names) Deliver(ctx context.Context, source string, names []ledger.InstrumentName) error {
	sourced := make([]ledger.InstrumentName, 0, len(names))
	for _, nm := range names {
		nm.Source, nm.ByOwner = source, false
		sourced = append(sourced, nm)
	}
	first := ledger.FirstDeliveredNames(sourced)
	if len(first) == 0 {
		return nil
	}
	if err := n.Store.AppendInstrumentNames(ctx, first); err != nil {
		return fmt.Errorf("recording instrument names: %w", err)
	}
	return nil
}

// Current returns the display name each named instrument shows
// (ledger.CurrentInstrumentNames).
func (n *Names) Current(ctx context.Context) (map[ledger.ISIN]string, error) {
	records, err := n.Store.InstrumentNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading instrument names: %w", err)
	}
	return ledger.CurrentInstrumentNames(records), nil
}

// Rename records name, cleaned (ledger.CleanInstrumentName), as the
// owner's display name of isin, which then shows before its ticker and ISIN
// on every page (D-197). The caller must be the owner, allowed to change
// every account that posts to the instrument (D-163); a name that isn't
// valid fails with ledger.ErrInvalidInstrumentName.
func (n *Names) Rename(ctx context.Context, origin, sessionToken string, isin ledger.ISIN, name string) (string, error) {
	if err := n.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return "", fmt.Errorf("renaming an instrument: %w", err)
	}
	clean, err := ledger.CleanInstrumentName(name)
	if err != nil {
		return "", fmt.Errorf("renaming %s: %w", isin, err)
	}
	accounts, err := n.postingAccounts(ctx, isin)
	if err != nil {
		return "", fmt.Errorf("renaming %s: %w", isin, err)
	}
	if err := mayChange(ctx, n.Owners, origin, sessionToken, accounts); err != nil {
		return "", fmt.Errorf("renaming %s: %w", isin, err)
	}
	if err := n.Store.AppendInstrumentNames(ctx, []ledger.InstrumentName{{ISIN: isin, Name: clean, ByOwner: true}}); err != nil {
		return "", fmt.Errorf("renaming %s: %w", isin, err)
	}
	return clean, nil
}

// postingAccounts returns the accounts whose current transactions post to
// isin, in ID order.
func (n *Names) postingAccounts(ctx context.Context, isin ledger.ISIN) ([]ledger.AccountID, error) {
	accounts, err := n.Postings.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	var out []ledger.AccountID
	for i := range accounts {
		txs, txErr := n.Postings.CurrentTransactions(ctx, accounts[i].ID)
		if txErr != nil {
			return nil, fmt.Errorf("loading transactions of %s: %w", accounts[i].ID, txErr)
		}
		if slices.ContainsFunc(txs, func(t ledger.Transaction) bool { return postsTo(&t, isin) }) {
			out = append(out, accounts[i].ID)
		}
	}
	return out, nil
}

// postsTo reports whether t posts to a holding of isin.
func postsTo(t *ledger.Transaction, isin ledger.ISIN) bool {
	return slices.ContainsFunc(t.Postings, func(p ledger.Posting) bool {
		return p.Book.Kind == ledger.BookHolding && p.Book.Instrument.ISIN.Equal(isin)
	})
}
