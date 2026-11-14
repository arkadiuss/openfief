package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerconnect "github.com/arkadiuss/openfief/internal/ledger/connectapi"
)

// newCashCheckpoints returns the owner's cash checkpoints (D-158, D-79)
// over the ledger, with the owner's sessions checked by owners.
func newCashCheckpoints(l *ledgerapp.Ledger, owners *ownerSessions) *ledgerapp.CashCheckpoints {
	return &ledgerapp.CashCheckpoints{Owners: owners, Ledger: l, Audit: owners}
}

// withCashCheckpointService returns h with the CashCheckpointService
// mounted next to it (D-158, D-79): its calls go to the checkpoints'
// handler, admitted like the AuthService's, view plugins only, every other
// path to h. Nothing may store its responses.
func withCashCheckpointService(h http.Handler, checkpoints *ledgerapp.CashCheckpoints, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := ledgerconnect.NewCashCheckpoints(checkpoints, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, noStore(handler))
	mux.Handle("/", h)
	return mux
}

// ledgerAttention adapts the attention items to the ledger's
// AttentionRaiser, so a cash checkpoint the owner enters raises its item
// at once, and booking its difference clears it (D-158). The ledger
// declares its port with plain strings and the attention context its own
// condition, so neither imports the other.
type ledgerAttention struct {
	items attentionItems
}

// Raise raises the item of kind and subject with detail, or opens it
// again when its detail changed.
func (a ledgerAttention) Raise(ctx context.Context, kind, subject, detail string) error {
	if err := a.items.Raise(ctx, attentiondomain.Condition{Kind: attentiondomain.Kind(kind), Subject: subject, Detail: detail}); err != nil {
		return fmt.Errorf("raising a ledger attention item: %w", err)
	}
	return nil
}

// Clear resolves the open or acknowledged item of kind and subject.
func (a ledgerAttention) Clear(ctx context.Context, kind, subject string) error {
	if err := a.items.Clear(ctx, attentiondomain.Kind(kind), subject); err != nil {
		return fmt.Errorf("clearing a ledger attention item: %w", err)
	}
	return nil
}
