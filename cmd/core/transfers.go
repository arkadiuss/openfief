package main

import (
	"log/slog"
	"net/http"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerconnect "github.com/arkadiuss/openfief/internal/ledger/connectapi"
)

// ledgerOwners is ownerSessions answering with the ledger context's errors.
func ledgerOwners(auth *identityapp.Auth, origins identity.Origins) ownerSessions {
	return ownerSessions{auth: auth, origins: origins, noSession: ledgerapp.ErrNoSession, notOwner: ledgerapp.ErrNotOwner}
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
