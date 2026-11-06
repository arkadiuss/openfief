package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/ledger"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerconnect "github.com/arkadiuss/openfief/internal/ledger/connectapi"
	"github.com/arkadiuss/openfief/internal/views"
)

// ledgerOwners is ownerSessions answering with the ledger context's errors,
// asking authz which accounts the owner may change.
func ledgerOwners(auth *identityapp.Auth, origins identity.Origins, authz identityapp.Authorizer) *ownerSessions {
	return &ownerSessions{auth: auth, origins: origins, noSession: ledgerapp.ErrNoSession, notOwner: ledgerapp.ErrNotOwner, authz: authz}
}

// newTransfers returns the owner's transfer decisions (D-105, D-79) over
// the ledger, with the owner's sessions checked by owners.
func newTransfers(l *ledgerapp.Ledger, owners ledgerapp.Owners) *ledgerapp.Transfers {
	return &ledgerapp.Transfers{Owners: owners, Ledger: l}
}

// withTransferService returns h with the TransferService mounted next to
// it (D-105, D-79): its calls go to the transfers' handler, admitted like
// the AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withTransferService(h http.Handler, transfers *ledgerapp.Transfers, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := ledgerconnect.NewTransfers(transfers, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}

// transferSuggestions are the transfers between the owner's own accounts
// that the read side suggests and the owner hasn't decided yet (D-105),
// over every account's current transactions: what the attention check
// raises a transfer-suggested item for (D-130). The rule is the read
// side's (views.SuggestTransfers), so the Inbox's item and the read API's
// suggestion are the same pair.
type transferSuggestions struct {
	ledger *ledgerapp.Ledger
}

// SuggestedTransfers returns each suggested pair, in the order
// views.SuggestTransfers suggests them.
func (s transferSuggestions) SuggestedTransfers(ctx context.Context) ([]ledger.TransferPair, error) {
	accounts, err := s.ledger.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading accounts: %w", err)
	}
	var txs []ledger.Transaction
	for i := range accounts {
		got, txErr := s.ledger.CurrentTransactions(ctx, accounts[i].ID)
		if txErr != nil {
			return nil, fmt.Errorf("loading transactions of %s: %w", accounts[i].ID, txErr)
		}
		txs = append(txs, got...)
	}
	links, err := s.ledger.TransferLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading transfer links: %w", err)
	}
	suggested := views.SuggestTransfers(views.ExternalFlows(txs), links)
	out := make([]ledger.TransferPair, 0, len(suggested))
	for i := range suggested {
		out = append(out, suggested[i].Pair())
	}
	return out, nil
}
