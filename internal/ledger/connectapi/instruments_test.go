package connectapi_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/ledger/connectapi"
	ledgerv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ledger/v1/ledgerv1connect"
)

// Synthetic ISIN with a valid check digit; no real instrument is implied.
const nameISIN = "PLTEST000010"

// nameStore keeps names in memory, or fails with errDown when down.
type nameStore struct {
	down  bool
	names []ledger.InstrumentName
}

func (s *nameStore) AppendInstrumentNames(_ context.Context, names []ledger.InstrumentName) error {
	if s.down {
		return errDown
	}
	s.names = append(s.names, names...)
	return nil
}

func (s *nameStore) InstrumentNames(context.Context) ([]ledger.InstrumentName, error) {
	return s.names, nil
}

// postings has acc-1 hold nothing: no account posts to the instrument.
type postings struct{}

func (postings) Accounts(context.Context) ([]ledger.Account, error) {
	return []ledger.Account{{ID: "acc-1"}}, nil
}

func (postings) CurrentTransactions(context.Context, ledger.AccountID) ([]ledger.Transaction, error) {
	return nil, nil
}

// checkRefused fails t unless err is a Connect error of code want, an
// internal one saying nothing of its cause.
func checkRefused(t *testing.T, err error, want connect.Code) {
	t.Helper()
	if connect.CodeOf(err) != want {
		t.Errorf("RenameInstrument = %v, want %s", err, want)
	}
	if ce, ok := errors.AsType[*connect.Error](err); ok && want == connect.CodeInternal && ce.Message() != "internal error" {
		t.Errorf("internal error message = %q", ce.Message())
	}
}

// TestInstrumentService checks the owner's rename of an instrument
// (D-197, D-79): the owner's view stores the cleaned name and gets it back;
// a read session, another caller, a malformed ISIN, an empty name and a
// store that fails are refused with their codes, the failure with nothing
// of its text.
func TestInstrumentService(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, caller, session, isin, newName string
		down                                 bool
		want                                 connect.Code
	}{
		{name: "the owner", caller: viewCaller, session: ownerSession, isin: nameISIN, newName: "  My  fund ", want: 0},
		{name: "a read session", caller: viewCaller, session: readSession, isin: nameISIN, newName: "x", want: connect.CodePermissionDenied},
		{name: "no session", caller: viewCaller, session: "ofsess_gone", isin: nameISIN, newName: "x", want: connect.CodeUnauthenticated},
		{name: "not a view plugin", caller: "other", session: ownerSession, isin: nameISIN, newName: "x", want: connect.CodePermissionDenied},
		{name: "a malformed ISIN", caller: viewCaller, session: ownerSession, isin: "PL123", newName: "x", want: connect.CodeInvalidArgument},
		{name: "an empty name", caller: viewCaller, session: ownerSession, isin: nameISIN, newName: " ", want: connect.CodeInvalidArgument},
		{name: "a store down", caller: viewCaller, session: ownerSession, isin: nameISIN, newName: "x", down: true, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &nameStore{down: tt.down}
			_, handler := connectapi.NewInstruments(&app.Names{Store: store, Owners: owners{}, Postings: postings{}}, viewOnly(), slog.New(slog.DiscardHandler))
			srv := httptest.NewServer(handler)
			t.Cleanup(srv.Close)
			c := ledgerv1connect.NewInstrumentServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(tt.caller)))
			resp, err := c.RenameInstrument(t.Context(), connect.NewRequest(&ledgerv1.RenameInstrumentRequest{
				Origin: local, SessionToken: tt.session, Isin: tt.isin, Name: tt.newName,
			}))
			if tt.want == 0 {
				if err != nil || resp.Msg.GetName() != "My fund" || len(store.names) != 1 || !store.names[0].ByOwner {
					t.Errorf("RenameInstrument = %v, %v; stored %v", resp, err, store.names)
				}
				return
			}
			checkRefused(t, err, tt.want)
		})
	}
}
