package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectapi"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// The synthetic caller token, sessions, origin and account.
const (
	viewCaller   = "view-plugin-token"
	ownerSession = "ofsess_owner"
	readSession  = "ofsess_read"
	local        = "http://localhost:8080"
	bondsAccount = "acc-bonds"
)

var (
	errDown    = errors.New("synthetic: ledger down")
	errNotView = errors.New("not a view plugin")
)

// owners accepts ownerSession, refuses readSession as not the owner's and
// any other token as no session.
type owners struct{}

func (owners) CheckOwner(_ context.Context, _, token string) error {
	switch token {
	case ownerSession:
		return nil
	case readSession:
		return app.ErrNotOwner
	}
	return app.ErrNoSession
}

// bindings binds bonds-main to acc-bonds only.
type bindings struct{}

func (bindings) BoundInstances(_ context.Context, account ledger.AccountID) ([]string, error) {
	if account == bondsAccount {
		return []string{"bonds-main"}, nil
	}
	return nil, nil
}

func (bindings) BoundAccounts(context.Context, string) ([]ledger.AccountID, error) {
	return []ledger.AccountID{bondsAccount}, nil
}

// caller records the lots it is handed and delivers a deposit of their
// cost, or fails with err.
type caller struct {
	err     error
	lots    []ingest.BondLot
	deleted []ingest.BondLotKey
}

func (c *caller) Sync(context.Context, string, *app.SyncCall) (app.Delivery, error) {
	return app.Delivery{}, c.err
}

func (c *caller) ImportFile(context.Context, string, ingest.BatchID, ingest.Mode, app.NamedFile) (app.Delivery, error) {
	return app.Delivery{}, c.err
}

func (c *caller) ImportBondLots(_ context.Context, _ string, _ ingest.BatchID, e *ingest.BondLotEntries) (app.Delivery, error) {
	c.lots, c.deleted = e.Lots, e.Deleted
	if c.err != nil {
		return app.Delivery{}, c.err
	}
	h := ledger.TransactionHeader{Key: ledger.TransactionKey{Source: "bonds", ExternalID: "d"}, Account: bondsAccount}
	if len(e.Lots) == 0 {
		h.Date = e.Deleted[0].Purchased
		t, err := ledger.NewDeletion(h)
		if err != nil {
			return app.Delivery{}, fmt.Errorf("deletion: %w", err)
		}
		return app.Delivery{Institution: "treasury-bonds", Mapping: ingest.Mapping{Transactions: []ledger.Transaction{t}}}, nil
	}
	h.Date = e.Lots[0].Purchased
	t, err := ledger.NewDeposit(h, e.Lots[0].Cost())
	if err != nil {
		return app.Delivery{}, fmt.Errorf("fake caller: %w", err)
	}
	return app.Delivery{Institution: "treasury-bonds", Mapping: ingest.Mapping{Transactions: []ledger.Transaction{t}}}, nil
}

// importer imports anything as new, or fails with err.
type importer struct {
	err error
}

func (i importer) Import(_ context.Context, txs []ledger.Transaction, _ []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	return ingest.ImportSummary{New: len(txs)}, i.err
}

func (i importer) Seed(ctx context.Context, txs []ledger.Transaction, a []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	return i.Import(ctx, txs, a)
}

func (i importer) AddOpeningLots(ctx context.Context, txs []ledger.Transaction, a []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	return i.Import(ctx, txs, a)
}

// recorder takes every batch.
type recorder struct{}

func (recorder) RecordBatch(context.Context, string, *ingest.Batch) error { return nil }

// viewOnly admits calls bearing viewCaller, like cmd/core's view-plugin
// check.
func viewOnly() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if req.Header().Get("Authorization") != "Bearer "+viewCaller {
				return nil, connect.NewError(connect.CodePermissionDenied, errNotView)
			}
			return next(ctx, req)
		}
	}
}

// bearer sets the Authorization header of every call.
func bearer(token string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

// client is the generated client, as a struct so helpers return no
// interface.
type client struct {
	ingestv1connect.BondLotServiceClient
}

// newClient serves the service over c and l, today 2026-10-29, and
// returns a client calling it with the caller token.
func newClient(t *testing.T, c *caller, l importer, callerToken string) client {
	t.Helper()
	today, err := kernel.ParseDate("2026-10-29")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	lots := &app.BondLots{
		Owners: owners{}, Bindings: bindings{}, Batches: recorder{},
		Connectors: &app.Connectors{Caller: c, Bindings: bindings{}, Ledger: l, Today: func() kernel.Date { return today }},
		Today:      func() kernel.Date { return today },
		NewBatchID: func() ingest.BatchID { return "b-1" },
	}
	_, handler := connectapi.New(lots, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return client{ingestv1connect.NewBondLotServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(callerToken)))}
}

// lot is a synthetic lot as the owner types it: 40 EDO0536 bonds bought on
// 2026-05-04, 6.80% then a 2.00% margin, a 2.00 PLN fee.
func lot() *ingestv1.BondLot {
	return &ingestv1.BondLot{
		Series: "EDO0536", Quantity: "40", PurchaseDate: "2026-05-04",
		FirstPeriodRatePercent: "6.80", MarginPercent: "2.00", EarlyRedemptionFeePln: "2.00",
	}
}

func TestAddBondLot(t *testing.T) {
	t.Parallel()
	c := &caller{}
	got, err := newClient(t, c, importer{}, viewCaller).AddBondLot(t.Context(), connect.NewRequest(&ingestv1.AddBondLotRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Lot: lot(),
	}))
	if err != nil {
		t.Fatalf("AddBondLot: %v", err)
	}
	if got.Msg.GetBatchId() != "b-1" || got.Msg.GetNew() != 1 {
		t.Errorf("AddBondLot = %v, want batch b-1 with one new transaction", got.Msg)
	}
	if len(c.lots) != 1 {
		t.Fatalf("handed %d lots to the connector, want 1", len(c.lots))
	}
	l := c.lots[0]
	if l.Terms.FirstPeriodRate.String() != "0.068" || l.Terms.Margin.String() != "0.02" || l.Terms.EarlyRedemptionFee.String() != "2 PLN" {
		t.Errorf("handed terms %s, %s, %s; want the percentages as fractions 0.068, 0.02 and a 2 PLN fee",
			l.Terms.FirstPeriodRate, l.Terms.Margin, l.Terms.EarlyRedemptionFee)
	}
}

func TestAddBondLotRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		caller   string
		session  string
		account  string
		edit     func(l *ingestv1.BondLot)
		callErr  error
		ledger   error
		want     connect.Code
		wantText string
	}{
		{name: "not a view plugin", caller: "nope", want: connect.CodePermissionDenied},
		{name: "no session", session: "ofsess_unknown", want: connect.CodeUnauthenticated},
		{name: "a read-scoped session", session: readSession, want: connect.CodePermissionDenied},
		{name: "a series that isn't one", edit: func(l *ingestv1.BondLot) { l.Series = "EDO536" }, want: connect.CodeInvalidArgument, wantText: "series"},
		{name: "a quantity that isn't a number", edit: func(l *ingestv1.BondLot) { l.Quantity = "forty" }, want: connect.CodeInvalidArgument, wantText: "quantity"},
		{name: "a date that isn't one", edit: func(l *ingestv1.BondLot) { l.PurchaseDate = "4.5.2026" }, want: connect.CodeInvalidArgument, wantText: "purchase date"},
		{name: "a rate that isn't a number", edit: func(l *ingestv1.BondLot) { l.FirstPeriodRatePercent = "6,8" }, want: connect.CodeInvalidArgument, wantText: "first-period rate"},
		{name: "a margin that isn't a number", edit: func(l *ingestv1.BondLot) { l.MarginPercent = "" }, want: connect.CodeInvalidArgument, wantText: "margin"},
		{name: "a fee that isn't a number", edit: func(l *ingestv1.BondLot) { l.EarlyRedemptionFeePln = "2 zł" }, want: connect.CodeInvalidArgument, wantText: "fee"},
		{name: "a purchase date of another series", edit: func(l *ingestv1.BondLot) { l.PurchaseDate = "2026-06-04" }, want: connect.CodeInvalidArgument, wantText: "EDO0536"},
		{name: "an account without a bond connector", account: "acc-ibkr", want: connect.CodeFailedPrecondition, wantText: "acc-ibkr"},
		{name: "the connector refuses", callErr: fmt.Errorf("%w: lot 1: synthetic reason", app.ErrEntryRefused), want: connect.CodeInvalidArgument, wantText: "synthetic reason"},
		{name: "the connector takes no bond lots", callErr: app.ErrNoEntries, want: connect.CodeFailedPrecondition},
		{name: "the connector can't be reached", callErr: app.ErrConnectorUnavailable, want: connect.CodeUnavailable},
		{name: "the ledger fails", ledger: errDown, want: connect.CodeInternal},
		{name: "a purchase that isn't one", edit: func(l *ingestv1.BondLot) { l.Purchase = ingestv1.BondPurchase(9) }, want: connect.CodeInvalidArgument, wantText: "purchase"},
		{name: "a lot that would count earlier transactions twice", ledger: fmt.Errorf("adding: %w", ledger.ErrBookedBeforeGoLive), want: connect.CodeFailedPrecondition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := &ingestv1.AddBondLotRequest{Origin: local, SessionToken: or(tt.session, ownerSession), AccountId: or(tt.account, bondsAccount), Lot: lot()}
			apply(tt.edit, req.GetLot())
			_, err := newClient(t, &caller{err: tt.callErr}, importer{err: tt.ledger}, or(tt.caller, viewCaller)).AddBondLot(t.Context(), connect.NewRequest(req))
			if connect.CodeOf(err) != tt.want || !strings.Contains(fmt.Sprint(err), tt.wantText) || strings.Contains(fmt.Sprint(err), errDown.Error()) {
				t.Errorf("AddBondLot = %v, want %s with %q and never an internal cause", err, tt.want, tt.wantText)
			}
		})
	}
}

// or returns s, or def when s is empty.
func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// apply changes l with edit, if there is one.
func apply(edit func(l *ingestv1.BondLot), l *ingestv1.BondLot) {
	if edit != nil {
		edit(l)
	}
}

func TestAddBondLotByExchange(t *testing.T) {
	t.Parallel()
	c := &caller{}
	l := lot()
	l.Purchase = ingestv1.BondPurchase_BOND_PURCHASE_EXCHANGE
	_, err := newClient(t, c, importer{}, viewCaller).AddBondLot(t.Context(), connect.NewRequest(&ingestv1.AddBondLotRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Lot: l,
	}))
	if err != nil {
		t.Fatalf("AddBondLot: %v", err)
	}
	if len(c.lots) != 1 || c.lots[0].Purchase != ingest.PurchaseExchange || c.lots[0].Cost().Amount().Cmp(kernel.NewDecimal(3996, 0)) != 0 {
		t.Errorf("handed %v; want one lot bought by exchange for 40 x 99.90 = 3996.00 PLN (D-100)", c.lots)
	}
}

func TestDeleteBondLot(t *testing.T) {
	t.Parallel()
	c := &caller{}
	got, err := newClient(t, c, importer{}, viewCaller).DeleteBondLot(t.Context(), connect.NewRequest(&ingestv1.DeleteBondLotRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Series: "EDO0536", PurchaseDate: "2026-05-04",
	}))
	if err != nil {
		t.Fatalf("DeleteBondLot: %v", err)
	}
	if got.Msg.GetBatchId() != "b-1" {
		t.Errorf("DeleteBondLot = %v, want batch b-1", got.Msg)
	}
	if len(c.lots) != 0 || len(c.deleted) != 1 || c.deleted[0].Series.String() != "EDO0536" || c.deleted[0].Purchased.String() != "2026-05-04" {
		t.Errorf("handed lots %v and deleted %v; want only EDO0536 of 2026-05-04 deleted", c.lots, c.deleted)
	}
}

func TestDeleteBondLotRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		session string
		series  string
		date    string
		ledger  error
		want    connect.Code
	}{
		{name: "a read-scoped session", session: readSession, want: connect.CodePermissionDenied},
		{name: "a series that isn't one", series: "EDO536", want: connect.CodeInvalidArgument},
		{name: "a date that isn't one", date: "4.5.2026", want: connect.CodeInvalidArgument},
		{name: "a lot never entered", ledger: fmt.Errorf("importing: %w", ledger.ErrNothingToDelete), want: connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := newClient(t, &caller{}, importer{err: tt.ledger}, viewCaller).DeleteBondLot(t.Context(), connect.NewRequest(&ingestv1.DeleteBondLotRequest{
				Origin: local, SessionToken: or(tt.session, ownerSession), AccountId: bondsAccount,
				Series: or(tt.series, "EDO0536"), PurchaseDate: or(tt.date, "2026-05-04"),
			}))
			if connect.CodeOf(err) != tt.want {
				t.Errorf("DeleteBondLot = %v, want %s", err, tt.want)
			}
		})
	}
}
