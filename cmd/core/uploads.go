package main

import (
	"log/slog"
	"net/http"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	ingestconnect "github.com/arkadiuss/openfief/internal/ingest/connectapi"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
)

// newUploads returns the statement upload use case (D-7, D-49, D-109) over
// d: the owner's sessions checked by owners, the account's bound instance
// from the ingest schema, its seeding from the ledger for the mode, the
// file streamed to it through d.entries with its media type sniffed from
// its first bytes, its delivery imported into the ledger and valuation
// like a sync's, and the batch recorded in the ingest schema. The accounts
// that take files come from the running instances whose manifests declare
// statement files, through d.entries once discovery runs. A file may be up
// to ingestapp.DefaultMaxUploadBytes.
func newUploads(d *connectorDeps, owners ingestapp.Owners) *ingestapp.Uploads {
	store := ingestpg.New(d.pool)
	connectors := entryConnectors(d, store)
	connectors.Seeding = d.ledger
	return &ingestapp.Uploads{
		Owners: owners, Bindings: store, Batches: store, Connectors: connectors, Instances: d.entries,
		NewBatchID: newBatchID,
		Sniff:      http.DetectContentType,
	}
}

// withStatementService returns h with the StatementService mounted next to
// it (D-49): its calls go to the uploads' handler, admitted like the
// AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withStatementService(h http.Handler, uploads *ingestapp.Uploads, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := ingestconnect.NewStatements(uploads, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	}))
	mux.Handle("/", h)
	return mux
}
