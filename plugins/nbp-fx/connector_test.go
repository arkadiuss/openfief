package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

var errNBPDown = errors.New("synthetic: NBP unavailable")

func day(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// fakeRates publishes one synthetic USD rate on each listed day and records
// the ranges it was asked for.
type fakeRates struct {
	mu        sync.Mutex
	published map[string]string // date -> rate
	err       error
	asked     [][2]string
}

func (f *fakeRates) Rates(_ context.Context, from, to kernel.Date) ([]valuation.FXRate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.asked = append(f.asked, [2]string{from.String(), to.String()})
	if f.err != nil {
		return nil, f.err
	}
	var out []valuation.FXRate
	for d := from; !d.After(to); d = d.AddDays(1) {
		rate, ok := f.published[d.String()]
		if !ok {
			continue
		}
		r, err := kernel.ParseDecimal(rate)
		if err != nil {
			return nil, fmt.Errorf("fake rate: %w", err)
		}
		usd, err := kernel.ParseCurrency("USD")
		if err != nil {
			return nil, fmt.Errorf("fake currency: %w", err)
		}
		out = append(out, valuation.FXRate{Currency: usd, Date: d, Rate: r, Source: "nbp-a"})
	}
	return out, nil
}

func TestSync(t *testing.T) {
	t.Parallel()
	published := map[string]string{"2025-01-02": "4.1219", "2025-01-03": "4.1512", "2025-01-07": "4.0732"}
	tests := []struct {
		name      string
		watermark string
		today     string
		err       error
		wantAsked [][2]string
		wantRates []*connectorv1.FxRate
		wantMark  string
		wantCode  connect.Code
	}{
		{
			name: "first sync starts at the configured start", today: "2025-01-03",
			wantAsked: [][2]string{{"2025-01-01", "2025-01-03"}},
			wantRates: []*connectorv1.FxRate{
				{Currency: "USD", Date: "2025-01-02", Rate: "4.1219"},
				{Currency: "USD", Date: "2025-01-03", Rate: "4.1512"},
			},
			wantMark: "2025-01-03",
		},
		{
			name: "a later sync continues after the watermark", watermark: "2025-01-03", today: "2025-01-08",
			wantAsked: [][2]string{{"2025-01-04", "2025-01-08"}},
			wantRates: []*connectorv1.FxRate{{Currency: "USD", Date: "2025-01-07", Rate: "4.0732"}},
			wantMark:  "2025-01-07",
		},
		{
			name: "nothing published since keeps the watermark", watermark: "2025-01-03", today: "2025-01-06",
			wantAsked: [][2]string{{"2025-01-04", "2025-01-06"}}, wantMark: "2025-01-03",
		},
		{name: "already synced up to today asks nothing", watermark: "2025-01-07", today: "2025-01-07", wantMark: "2025-01-07"},
		{name: "a watermark this connector never wrote", watermark: "yesterday", today: "2025-01-07", wantCode: connect.CodeInvalidArgument},
		{
			name: "NBP failing is unavailable, so the core retries", today: "2025-01-03", err: errNBPDown,
			wantAsked: [][2]string{{"2025-01-01", "2025-01-03"}}, wantCode: connect.CodeUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rates := &fakeRates{published: published, err: tt.err}
			today := day(t, tt.today)
			c := &connector{rates: rates, today: func() kernel.Date { return today }, start: day(t, "2025-01-01")}
			resp, err := c.Sync(t.Context(), connect.NewRequest(&connectorv1.SyncRequest{
				BatchId: "b-1", Mode: connectorv1.Mode_MODE_TRANSACTIONS, Watermark: tt.watermark,
			}))
			if diff := cmp.Diff(tt.wantAsked, rates.asked); diff != "" {
				t.Errorf("ranges asked (-want +got):\n%s", diff)
			}
			switch {
			case tt.wantCode != 0 && connect.CodeOf(err) != tt.wantCode:
				t.Fatalf("Sync error = %v, want code %v", err, tt.wantCode)
			case tt.wantCode != 0:
				return
			case err != nil:
				t.Fatalf("Sync: %v", err)
			}
			want := &connectorv1.SyncResponse{
				Delivery:  &connectorv1.Delivery{BatchId: "b-1", Institution: "nbp", FxRates: tt.wantRates},
				Watermark: tt.wantMark,
			}
			if diff := cmp.Diff(want, resp.Msg, protocmp.Transform()); diff != "" {
				t.Errorf("Sync (-want +got):\n%s", diff)
			}
		})
	}
}

// coreToken is what the fake verifier takes for the core's token.
const coreToken = "synthetic-core-token"

type fakeVerifier struct{}

func (fakeVerifier) VerifyCore(_ context.Context, token string) error {
	if token != coreToken {
		return openfief.ErrNotCore
	}
	return nil
}

// TestServedConnector serves the connector with the SDK, as the plugin
// does, and calls it as the core: Sync answers, ImportFile is not served.
func TestServedConnector(t *testing.T) {
	t.Parallel()
	today := day(t, "2025-01-03")
	c := &connector{
		rates: &fakeRates{published: map[string]string{"2025-01-02": "4.1219"}},
		today: func() kernel.Date { return today }, start: day(t, "2025-01-01"),
	}
	server := openfief.NewConnectorServer("", c, fakeVerifier{})
	srv := httptest.NewUnstartedServer(server.Handler)
	srv.Config.Protocols = server.Protocols
	srv.Start()
	t.Cleanup(srv.Close)
	var p http.Protocols
	p.SetUnencryptedHTTP2(true)
	client := connectorv1connect.NewConnectorServiceClient(&http.Client{Transport: &http.Transport{Protocols: &p}}, srv.URL)

	req := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-1"})
	req.Header().Set("Authorization", "Bearer "+coreToken)
	resp, err := client.Sync(t.Context(), req)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if got := resp.Msg.GetDelivery().GetFxRates(); len(got) != 1 || resp.Msg.GetWatermark() != "2025-01-02" {
		t.Errorf("Sync = %v, want one rate and watermark 2025-01-02", resp.Msg)
	}
	unauthorized := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-2"})
	if _, err := client.Sync(t.Context(), unauthorized); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("Sync without the core's token: %v, want Unauthenticated", err)
	}
	stream := client.ImportFile(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+coreToken)
	if err := stream.Send(&connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Header{
		Header: &connectorv1.FileHeader{BatchId: "b-3", FileName: "rates.csv"},
	}}); err != nil && !errors.Is(err, context.Canceled) {
		t.Logf("send: %v", err)
	}
	if _, err := stream.CloseAndReceive(); connect.CodeOf(err) != connect.CodeUnimplemented {
		t.Errorf("ImportFile: %v, want Unimplemented", err)
	}
}
