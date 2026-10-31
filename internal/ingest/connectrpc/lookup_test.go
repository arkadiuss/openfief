package connectrpc_test

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectrpc"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// lookupConnector proposes a synthetic ISIN for TESTA.PL, once more for
// TESTA.PL, and nothing for any other ticker.
type lookupConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	extra *connectorv1.InstrumentProposal
}

func (l *lookupConnector) LookupInstruments(_ context.Context, req *connect.Request[connectorv1.LookupInstrumentsRequest]) (*connect.Response[connectorv1.LookupInstrumentsResponse], error) {
	var out []*connectorv1.InstrumentProposal
	for _, tk := range req.Msg.GetTickers() {
		if tk.GetSymbol() == "TESTA.PL" && tk.GetCode() == "TESTA" && tk.GetCountry() == "PL" {
			out = append(out,
				&connectorv1.InstrumentProposal{Symbol: "TESTA.PL", Isin: "PLTESTA00002", Name: "Test Alpha SA"},
				&connectorv1.InstrumentProposal{Symbol: "TESTA.PL", Isin: "PLTESTC00008", Name: "repeat"})
		}
	}
	if l.extra != nil {
		out = append(out, l.extra)
	}
	return connect.NewResponse(&connectorv1.LookupInstrumentsResponse{Proposals: out}), nil
}

func TestClientLookupInstruments(t *testing.T) {
	t.Parallel()
	tickers := []ingest.Ticker{{Symbol: "TESTA.PL", Code: "TESTA", Country: "PL"}, {Symbol: "TESTB.UK", Code: "TESTB", Country: "GB"}}
	got, err := newClient(t, &lookupConnector{}).LookupInstruments(t.Context(), "eodhd", tickers)
	if err != nil {
		t.Fatalf("LookupInstruments: %v", err)
	}
	isin, err := ledger.ParseISIN("PLTESTA00002")
	if err != nil {
		t.Fatal(err)
	}
	want := []ingest.ISINProposal{{Symbol: "TESTA.PL", ISIN: isin, Name: "Test Alpha SA"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("proposals (-want +got):\n%s", diff)
	}
	tests := []struct {
		name  string
		conn  connectorv1connect.ConnectorServiceHandler
		where string
		want  error
	}{
		{name: "no lookup", conn: &fakeConnector{}, where: "nbp-fx", want: app.ErrNoEntries},
		{name: "unreachable", conn: &lookupConnector{}, where: "nowhere", want: app.ErrConnectorUnavailable},
		{name: "ticker not asked", conn: &lookupConnector{extra: &connectorv1.InstrumentProposal{Symbol: "OTHER.PL", Isin: "PLTESTA00002"}}, where: "eodhd", want: connectrpc.ErrInvalidMessage},
		{name: "bad ISIN", conn: &lookupConnector{extra: &connectorv1.InstrumentProposal{Symbol: "TESTB.UK", Isin: "GBTESTB00001"}}, where: "eodhd", want: connectrpc.ErrInvalidMessage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := newClient(t, tt.conn).LookupInstruments(t.Context(), tt.where, tickers)
			if !errors.Is(err, tt.want) {
				t.Errorf("LookupInstruments() = %v, want %v", err, tt.want)
			}
		})
	}
}
