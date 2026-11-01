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
// from the ingest schema, its seeding and whether it holds transactions
// from the ledger for the mode (D-109, D-136), the
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
		NewBatchID: newBatchID, History: d.ledger,
		Sniff: http.DetectContentType,
	}
}

// newTickers returns the owner's ticker confirmations (D-116) over d: the
// owner's sessions checked by owners, the tickers, proposals and waiting
// entries in the ingest schema, and a confirmed ticker's entries booked
// into the ledger.
func newTickers(d *connectorDeps, owners ingestapp.Owners) *ingestapp.Tickers {
	return &ingestapp.Tickers{Owners: owners, Store: ingestpg.New(d.pool), Ledger: ledgerImporter{ledger: d.ledger}}
}

// withStatementService returns h with the StatementService and the
// TickerService mounted next to it (D-49, D-79, D-116): their calls go to
// the uploads' and the tickers' handlers, admitted like the AuthService's,
// view plugins only, every other path to h. Nothing may store their
// responses.
func withStatementService(h http.Handler, uploads *ingestapp.Uploads, t *ingestapp.Tickers, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	statementPath, statementHandler := ingestconnect.NewStatements(uploads, identityconnect.NewCallerInterceptor(callers, log), log)
	tickerPath, tickerHandler := ingestconnect.NewTickers(t, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(statementPath, noStore(statementHandler))
	mux.Handle(tickerPath, noStore(tickerHandler))
	mux.Handle("/", h)
	return mux
}
