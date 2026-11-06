package connectapi_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	gocmp "github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/connectapi"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// checkpointBook derives 100 of any currency on every date, or fails with
// err; it books a deposit of the difference.
type checkpointBook struct{ err error }

func (b checkpointBook) RecordCashCheckpoint(_ context.Context, account ledger.AccountID, on kernel.Date, cash kernel.Money) (app.AssertionCheck, error) {
	if b.err != nil {
		return app.AssertionCheck{}, fmt.Errorf("recording: %w", b.err)
	}
	derived, err := kernel.ParseDecimal("100")
	if err != nil {
		return app.AssertionCheck{}, fmt.Errorf("parsing: %w", err)
	}
	status := app.AssertionMatched
	if !derived.Equal(cash.Amount()) {
		status = app.AssertionFlagged
	}
	return app.AssertionCheck{Assertion: ledger.CashCheckpoint(account, on, cash), Status: status, Derived: derived}, nil
}

func (b checkpointBook) BookCashDifference(_ context.Context, account ledger.AccountID, on kernel.Date, c kernel.Currency) (ledger.Transaction, error) {
	if b.err != nil {
		return ledger.Transaction{}, fmt.Errorf("booking: %w", b.err)
	}
	five, err := kernel.ParseMoney("-5", c)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("parsing: %w", err)
	}
	t, err := ledger.NewWithdrawal(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: ledger.CashCheckpointSource, ExternalID: on.String() + "/" + c.String() + "/95/100"}, Account: account, Date: on,
	}, five.Neg())
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("building: %w", err)
	}
	return t, nil
}

// checkpointAPI is the generated client, as a struct so helpers return no
// interface.
type checkpointAPI struct {
	ledgerv1connect.CashCheckpointServiceClient
}

// checkpointClient serves the CashCheckpointService over book and returns
// a client that calls it as caller.
func checkpointClient(t *testing.T, book app.CashCheckpointBook, caller string) checkpointAPI {
	t.Helper()
	_, handler := connectapi.NewCashCheckpoints(&app.CashCheckpoints{Owners: owners{}, Ledger: book}, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return checkpointAPI{ledgerv1connect.NewCashCheckpointServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

func TestCashCheckpointService(t *testing.T) {
	t.Parallel()
	c := checkpointClient(t, checkpointBook{}, viewCaller)
	ctx := t.Context()
	rec, err := c.RecordCashCheckpoint(ctx, connect.NewRequest(&ledgerv1.RecordCashCheckpointRequest{
		Origin: local, SessionToken: ownerSession, AccountId: "acc-1", Date: "2026-03-31", Currency: "PLN", Amount: "95.00",
	}))
	if err != nil {
		t.Fatalf("RecordCashCheckpoint: %v", err)
	}
	want := &ledgerv1.CashCheckpoint{AccountId: "acc-1", Date: "2026-03-31", Currency: "PLN", Shown: "95", Derived: "100"}
	if diff := gocmp.Diff(want, rec.Msg.GetCheckpoint(), protocmp.Transform()); diff != "" {
		t.Errorf("checkpoint (-want +got):\n%s", diff)
	}
	booked, err := c.BookCashDifference(ctx, connect.NewRequest(&ledgerv1.BookCashDifferenceRequest{
		Origin: local, SessionToken: ownerSession, AccountId: "acc-1", Date: "2026-03-31", Currency: "PLN",
	}))
	if err != nil {
		t.Fatalf("BookCashDifference: %v", err)
	}
	wantBooked := &ledgerv1.BookCashDifferenceResponse{
		Key:  &ledgerv1.TransactionKey{Source: "owner:cash-checkpoint", ExternalId: "2026-03-31/PLN/95/100"},
		Kind: "withdrawal", Amount: "5", Currency: "PLN",
	}
	if diff := gocmp.Diff(wantBooked, booked.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("booked (-want +got):\n%s", diff)
	}
}

func TestCashCheckpointServiceRefusals(t *testing.T) {
	t.Parallel()
	record := func(session, date, currency, amount string) *ledgerv1.RecordCashCheckpointRequest {
		return &ledgerv1.RecordCashCheckpointRequest{Origin: local, SessionToken: session, AccountId: "acc-1", Date: date, Currency: currency, Amount: amount}
	}
	tests := []struct {
		name   string
		book   checkpointBook
		caller string
		req    *ledgerv1.RecordCashCheckpointRequest
		book2  bool
		want   connect.Code
	}{
		{name: "not a view plugin", caller: "other", req: record(ownerSession, "2026-03-31", "PLN", "1"), want: connect.CodePermissionDenied},
		{name: "no session", caller: viewCaller, req: record("", "2026-03-31", "PLN", "1"), want: connect.CodeUnauthenticated},
		{name: "a read session", caller: viewCaller, req: record(readSession, "2026-03-31", "PLN", "1"), want: connect.CodePermissionDenied},
		{name: "a malformed date", caller: viewCaller, req: record(ownerSession, "31.03.2026", "PLN", "1"), want: connect.CodeInvalidArgument},
		{name: "a malformed currency", caller: viewCaller, req: record(ownerSession, "2026-03-31", "zloty", "1"), want: connect.CodeInvalidArgument},
		{name: "a malformed amount", caller: viewCaller, req: record(ownerSession, "2026-03-31", "PLN", "1,5"), want: connect.CodeInvalidArgument},
		{name: "an unknown account", book: checkpointBook{err: app.ErrUnknownAccount}, caller: viewCaller, req: record(ownerSession, "2026-03-31", "PLN", "1"), want: connect.CodeNotFound},
		{name: "an unknown checkpoint", book: checkpointBook{err: app.ErrUnknownCheckpoint}, book2: true, caller: viewCaller, req: record(ownerSession, "2026-03-31", "PLN", "1"), want: connect.CodeNotFound},
		{name: "nothing to book", book: checkpointBook{err: ledger.ErrNoCashDifference}, book2: true, caller: viewCaller, req: record(ownerSession, "2026-03-31", "PLN", "1"), want: connect.CodeFailedPrecondition},
		{name: "a store failure", book: checkpointBook{err: errDown}, caller: viewCaller, req: record(ownerSession, "2026-03-31", "PLN", "1"), want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := checkpointClient(t, tt.book, tt.caller)
			var err error
			if tt.book2 {
				_, err = c.BookCashDifference(t.Context(), connect.NewRequest(&ledgerv1.BookCashDifferenceRequest{
					Origin: local, SessionToken: tt.req.GetSessionToken(), AccountId: "acc-1", Date: tt.req.GetDate(), Currency: tt.req.GetCurrency(),
				}))
			} else {
				_, err = c.RecordCashCheckpoint(t.Context(), connect.NewRequest(tt.req))
			}
			if got := connect.CodeOf(err); got != tt.want {
				t.Errorf("code = %v (%v), want %v", got, err, tt.want)
			}
		})
	}
}
