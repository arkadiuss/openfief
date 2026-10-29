package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	identityconnect "github.com/arkadiuss/openfief/internal/identity/connectapi"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	ingestconnect "github.com/arkadiuss/openfief/internal/ingest/connectapi"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	ingestpg "github.com/arkadiuss/openfief/internal/ingest/postgres"
)

// entryCaller is the connector client the core hands the owner's manual
// entries to (D-49): none until the core has found the cluster's connector
// instances, and then the client over them. Until then, and outside a
// cluster, every call fails with ingestapp.ErrConnectorUnavailable.
type entryCaller struct {
	client atomic.Pointer[connectrpc.Client]
}

// Sync syncs instance through the client.
func (e *entryCaller) Sync(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, watermark string) (ingestapp.Delivery, error) {
	c, err := e.get()
	if err != nil {
		return ingestapp.Delivery{}, err
	}
	d, err := c.Sync(ctx, instance, id, mode, watermark)
	if err != nil {
		return ingestapp.Delivery{}, fmt.Errorf("sync: %w", err)
	}
	return d, nil
}

// ImportFile imports a file into instance through the client.
func (e *entryCaller) ImportFile(ctx context.Context, instance string, id ingest.BatchID, mode ingest.Mode, file ingestapp.NamedFile) (ingestapp.Delivery, error) {
	c, err := e.get()
	if err != nil {
		return ingestapp.Delivery{}, err
	}
	d, err := c.ImportFile(ctx, instance, id, mode, file)
	if err != nil {
		return ingestapp.Delivery{}, fmt.Errorf("import file: %w", err)
	}
	return d, nil
}

// ImportBondLots hands bond lots to instance through the client.
func (e *entryCaller) ImportBondLots(ctx context.Context, instance string, id ingest.BatchID, lots []ingest.BondLot) (ingestapp.Delivery, error) {
	c, err := e.get()
	if err != nil {
		return ingestapp.Delivery{}, err
	}
	d, err := c.ImportBondLots(ctx, instance, id, lots)
	if err != nil {
		return ingestapp.Delivery{}, fmt.Errorf("import bond lots: %w", err)
	}
	return d, nil
}

// set makes c the client every later call goes through.
func (e *entryCaller) set(c *connectrpc.Client) { e.client.Store(c) }

// get returns the client, or an error wrapping
// ingestapp.ErrConnectorUnavailable when there is none yet.
func (e *entryCaller) get() (*connectrpc.Client, error) {
	c := e.client.Load()
	if c == nil {
		return nil, fmt.Errorf("%w: no connector instances found; the core finds them only in a cluster with --connectors", ingestapp.ErrConnectorUnavailable)
	}
	return c, nil
}

// newBondLots returns the manual bond lot use case (D-9, D-49) over d: the
// owner's sessions checked by owners, the account's bound instance from the
// ingest schema, the lot handed to it through d.entries and its delivery
// imported into the ledger and valuation like a sync's, and the batch
// recorded in the ingest schema.
func newBondLots(d *connectorDeps, owners ingestapp.Owners) *ingestapp.BondLots {
	store := ingestpg.New(d.pool)
	return &ingestapp.BondLots{
		Owners: owners, Bindings: store, Batches: store, Connectors: entryConnectors(d, store),
		Today:      d.today,
		NewBatchID: newBatchID,
	}
}

// entryConnectors returns the connector use case the owner's entries and
// uploads go through: d.entries as the client, store's connector bindings
// and broker accounts (D-72),
// and the delivery imported into the ledger and valuation like a sync's,
// cross-check prices included (D-22).
func entryConnectors(d *connectorDeps, store *ingestpg.Store) *ingestapp.Connectors {
	inputs := valuationInputs{recorder: d.recorder}
	return &ingestapp.Connectors{
		Caller: d.entries, Bindings: store, Ledger: ledgerImporter{ledger: d.ledger},
		Prices: inputs, FXRates: inputs, Indices: inputs, Levels: inputs, Terms: inputs,
		CrossChecks: crossCheckInputs{checks: d.crossChecks}, Brokers: store, Today: d.today,
	}
}

// withBondLotService returns h with the BondLotService mounted next to it
// (D-49): its calls go to the bond lots' handler, admitted like the
// AuthService's, view plugins only, every other path to h. Nothing may
// store its responses.
func withBondLotService(h http.Handler, lots *ingestapp.BondLots, callers identityconnect.Callers, log *slog.Logger) http.Handler {
	path, handler := ingestconnect.New(lots, identityconnect.NewCallerInterceptor(callers, log), log)
	mux := http.NewServeMux()
	mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		handler.ServeHTTP(w, r)
	}))
	mux.Handle("/", h)
	return mux
}
