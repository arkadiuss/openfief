package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	stdsync "sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// Synthetic ISINs, with valid check digits: SYNT is on Xetra in euros and
// on the LSE in pence, FAKEW on Warsaw in zloty, NON on no supported
// exchange.
const (
	syntISIN  = "XX00000SYNT3"
	fakewISIN = "XX0000FAKEW1"
	nonISIN   = "XX000000NON6"
	fakeKey   = "synthetic-test-key"
)

var errEODHDDown = errors.New("synthetic: EODHD unavailable")

func day(t *testing.T, s string) kernel.Date {
	t.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// fakeEODHD lists the synthetic ISINs and publishes closes on a few days
// of January 2025 for every ticker it knows, recording what it was asked.
type fakeEODHD struct {
	mu     stdsync.Mutex
	token  string
	closes map[string]map[string]string // ticker -> date -> close
	err    error
	asked  []string
}

func newFakeEODHD(token string) *fakeEODHD {
	return &fakeEODHD{token: token, closes: map[string]map[string]string{
		"FAKEW.WAR":   {"2025-01-02": "42.10", "2025-01-03": "42.5"},
		"SYNT.XETRA":  {"2025-01-02": "100.52", "2025-01-03": "101.075"},
		"SYNT.LSE":    {"2025-01-02": "8412.5", "2025-01-03": "8420"},
		"FAKE20.INDX": {"2025-01-02": "2401.17", "2025-01-03": "2399.5"},
	}}
}

func (f *fakeEODHD) Listings(_ context.Context, isin ledger.ISIN) ([]eodhd.Listing, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.asked = append(f.asked, "search "+isin.String())
	if f.err != nil {
		return nil, f.err
	}
	switch isin.String() {
	case syntISIN:
		return []eodhd.Listing{
			{ISIN: isin, Code: "SYNT", Exchange: "XETRA", Currency: "EUR"},
			{ISIN: isin, Code: "SYNT", Exchange: "LSE", Currency: "GBX"},
		}, nil
	case fakewISIN:
		return []eodhd.Listing{{ISIN: isin, Code: "FAKEW", Exchange: "WAR", Currency: "PLN"}}, nil
	case nonISIN:
		return []eodhd.Listing{{ISIN: isin, Code: "NON", Exchange: "US", Currency: "USD"}}, nil
	}
	return nil, nil
}

func (f *fakeEODHD) Closes(_ context.Context, ticker string, from, to kernel.Date) ([]eodhd.Close, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	since := from.String()
	if from.IsZero() {
		since = "start"
	}
	f.asked = append(f.asked, fmt.Sprintf("eod %s %s..%s", ticker, since, to))
	published, ok := f.closes[ticker]
	if !ok {
		return nil, eodhd.ErrNotFound
	}
	var out []eodhd.Close
	for on, value := range published {
		d, err := kernel.ParseDate(on)
		if err != nil {
			return nil, fmt.Errorf("fake date: %w", err)
		}
		if d.Before(from) || d.After(to) {
			continue
		}
		v, err := kernel.ParseDecimal(value)
		if err != nil {
			return nil, fmt.Errorf("fake close: %w", err)
		}
		out = append(out, eodhd.Close{Date: d, Close: v})
	}
	slices.SortFunc(out, func(a, b eodhd.Close) int { return a.Date.Compare(b.Date) })
	return out, nil
}

// secretsDir is a mounted Secret holding the API key, or nothing when key
// is empty.
func secretsDir(t *testing.T, key string) string {
	t.Helper()
	dir := t.TempDir()
	if key != "" {
		if err := os.WriteFile(filepath.Join(dir, apiTokenFile), []byte(key+"\n"), 0o600); err != nil {
			t.Fatalf("write key: %v", err)
		}
	}
	return dir
}

func price(isin, symbol, on, amount, currency string) *connectorv1.Price {
	return &connectorv1.Price{
		Instrument: &connectorv1.Instrument{Isin: isin, Symbol: symbol}, Date: on,
		Close: &connectorv1.Money{Amount: amount, Currency: currency},
	}
}

func level(on, value string) *connectorv1.IndexLevel {
	return &connectorv1.IndexLevel{Index: "FAKE20", Date: on, Level: value}
}

func unsupported(ref, reason string) *connectorv1.Problem {
	return &connectorv1.Problem{Kind: connectorv1.ProblemKind_PROBLEM_KIND_UNSUPPORTED, Ref: ref, Reason: reason}
}

type syncCase struct {
	name          string
	instruments   []rawInstrument
	indices       []rawIndex
	start         string
	watermark     string
	today         string
	key           string
	err           error
	wantAsked     []string
	wantPrices    []*connectorv1.Price
	wantLevels    []*connectorv1.IndexLevel
	wantProblems  []*connectorv1.Problem
	wantWatermark string
	wantCode      connect.Code
}

func syncCases() []syncCase {
	both := []rawInstrument{{ISIN: fakewISIN}, {ISIN: syntISIN, Exchange: "LSE", Symbol: "SYNTL"}}
	fake20 := []rawIndex{{Name: "FAKE20", Ticker: "FAKE20.INDX"}}
	const (
		synced  = `{"prices":{"XX00000SYNT3":"2025-01-03","XX0000FAKEW1":"2025-01-03"},"indices":{"FAKE20":"2025-01-03"}}`
		lagging = `{"prices":{"XX00000SYNT3":"2025-01-03","XX0000FAKEW1":"2025-01-02"},"indices":{"FAKE20":"2025-01-03"}}`
	)
	return []syncCase{
		{
			name: "first sync from the configured start", instruments: both, indices: fake20, start: "2025-01-02", today: "2025-01-03",
			wantAsked: []string{
				"search " + fakewISIN, "eod FAKEW.WAR 2025-01-02..2025-01-03",
				"search " + syntISIN, "eod SYNT.LSE 2025-01-02..2025-01-03", "eod FAKE20.INDX 2025-01-02..2025-01-03",
			},
			wantPrices: []*connectorv1.Price{
				price(fakewISIN, "FAKEW", "2025-01-02", "42.1", "PLN"), price(fakewISIN, "FAKEW", "2025-01-03", "42.5", "PLN"),
				price(syntISIN, "SYNTL", "2025-01-02", "84.125", "GBP"), price(syntISIN, "SYNTL", "2025-01-03", "84.2", "GBP"),
			},
			wantLevels:    []*connectorv1.IndexLevel{level("2025-01-02", "2401.17"), level("2025-01-03", "2399.5")},
			wantWatermark: synced,
		},
		{
			name: "a later sync continues after each mark", instruments: both, indices: fake20, start: "2025-01-02",
			watermark: lagging, today: "2025-01-03",
			wantAsked:     []string{"search " + fakewISIN, "eod FAKEW.WAR 2025-01-03..2025-01-03"},
			wantPrices:    []*connectorv1.Price{price(fakewISIN, "FAKEW", "2025-01-03", "42.5", "PLN")},
			wantWatermark: synced,
		},
		{
			name: "nothing new keeps the watermark", instruments: both[:1], watermark: `{"prices":{"XX0000FAKEW1":"2025-01-03"}}`,
			today: "2025-01-06", wantAsked: []string{"search " + fakewISIN, "eod FAKEW.WAR 2025-01-04..2025-01-06"},
			wantWatermark: `{"prices":{"XX0000FAKEW1":"2025-01-03"}}`,
		},
		{
			name: "without a start, the whole history", instruments: []rawInstrument{{ISIN: syntISIN, Exchange: "XETRA"}}, today: "2025-01-02",
			wantAsked:     []string{"search " + syntISIN, "eod SYNT.XETRA start..2025-01-02"},
			wantPrices:    []*connectorv1.Price{price(syntISIN, "SYNT", "2025-01-02", "100.52", "EUR")},
			wantWatermark: `{"prices":{"XX00000SYNT3":"2025-01-02"}}`,
		},
		{
			name: "ambiguous and unlisted ISINs are problems, the rest is delivered", today: "2025-01-02", start: "2025-01-02",
			instruments: []rawInstrument{{ISIN: syntISIN}, {ISIN: nonISIN}, {ISIN: fakewISIN}},
			wantAsked: []string{
				"search " + syntISIN, "search " + nonISIN, "search " + fakewISIN, "eod FAKEW.WAR 2025-01-02..2025-01-02",
			},
			wantPrices: []*connectorv1.Price{price(fakewISIN, "FAKEW", "2025-01-02", "42.1", "PLN")},
			wantProblems: []*connectorv1.Problem{
				unsupported(syntISIN, "listed on more than one supported exchange: XX00000SYNT3 as SYNT.XETRA, SYNT.LSE; name one exchange"),
				unsupported(nonISIN, "not listed on a supported exchange: XX000000NON6"),
			},
			wantWatermark: `{"prices":{"XX0000FAKEW1":"2025-01-02"}}`,
		},
		{
			name: "an index ticker EODHD doesn't know is a problem", today: "2025-01-02", start: "2025-01-02",
			indices:       []rawIndex{{Name: "GONE", Ticker: "GONE.INDX"}},
			wantAsked:     []string{"eod GONE.INDX 2025-01-02..2025-01-02"},
			wantProblems:  []*connectorv1.Problem{unsupported("GONE", "EODHD has no ticker GONE.INDX")},
			wantWatermark: "",
		},
		{
			name: "EODHD failing is unavailable, so the core retries", instruments: both, today: "2025-01-03", err: errEODHDDown,
			wantAsked: []string{"search " + fakewISIN}, wantCode: connect.CodeUnavailable,
		},
		{name: "a watermark this connector never wrote", instruments: both, watermark: "2025-01-03", today: "2025-01-03", wantCode: connect.CodeInvalidArgument},
		{name: "a mark that is not a date", instruments: both, watermark: `{"prices":{"XX0000FAKEW1":"yesterday"}}`, today: "2025-01-03", wantCode: connect.CodeInvalidArgument},
		{name: "no API key in the Secret", instruments: both, today: "2025-01-03", key: "-", wantCode: connect.CodeFailedPrecondition},
	}
}

func TestSync(t *testing.T) {
	t.Parallel()
	for _, tt := range syncCases() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			checkSync(t, &tt)
		})
	}
}

// syncWith runs one Sync of a connector configured as tt says, against a
// fake EODHD, and returns the fake it called, if any.
func syncWith(t *testing.T, tt *syncCase) (*connectorv1.SyncResponse, *fakeEODHD, error) {
	t.Helper()
	raw := rawConfig{Start: tt.start, Instruments: tt.instruments, Indices: tt.indices}
	cfg, err := raw.parse()
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	key := fakeKey
	if tt.key == "-" {
		key = ""
	}
	var src *fakeEODHD
	today := day(t, tt.today)
	c := &connector{
		cfg: cfg, secrets: secretsDir(t, key), today: func() kernel.Date { return today },
		source: func(token string) marketData { src = newFakeEODHD(token); src.err = tt.err; return src },
	}
	resp, err := c.Sync(t.Context(), connect.NewRequest(&connectorv1.SyncRequest{
		BatchId: "b-1", Mode: connectorv1.Mode_MODE_TRANSACTIONS, Watermark: tt.watermark,
	}))
	if err != nil {
		return nil, src, err
	}
	return resp.Msg, src, nil
}

// checkSync checks one case of TestSync.
func checkSync(t *testing.T, tt *syncCase) {
	t.Helper()
	got, src, err := syncWith(t, tt)
	if tt.wantCode != 0 {
		if connect.CodeOf(err) != tt.wantCode || strings.Contains(fmt.Sprint(err), fakeKey) {
			t.Fatalf("Sync error = %v, want code %v and no API key", err, tt.wantCode)
		}
		return
	}
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if src.token != fakeKey {
		t.Errorf("EODHD called with key %q, want the Secret's", src.token)
	}
	if diff := cmp.Diff(tt.wantAsked, src.asked); diff != "" {
		t.Errorf("requests (-want +got):\n%s", diff)
	}
	want := &connectorv1.SyncResponse{
		Delivery: &connectorv1.Delivery{
			BatchId: "b-1", Institution: "eodhd",
			Prices: tt.wantPrices, IndexLevels: tt.wantLevels, Problems: tt.wantProblems,
		},
		Watermark: tt.wantWatermark,
	}
	if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
		t.Errorf("Sync (-want +got):\n%s", diff)
	}
}

func TestNextFrom(t *testing.T) {
	t.Parallel()
	marks := map[string]string{"A": "2025-01-03"}
	tests := []struct {
		name, key, start, today string
		wantFrom                string
		wantDue                 bool
	}{
		{name: "the day after the mark", key: "A", start: "2024-01-01", today: "2025-01-06", wantFrom: "2025-01-04", wantDue: true},
		{name: "marked today is not due", key: "A", start: "2024-01-01", today: "2025-01-03", wantFrom: "2025-01-04"},
		{name: "no mark starts at start", key: "B", start: "2024-01-01", today: "2025-01-06", wantFrom: "2024-01-01", wantDue: true},
		{name: "a start after today is not due", key: "B", start: "2025-02-01", today: "2025-01-06", wantFrom: "2025-02-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			from, due := nextFrom(marks, tt.key, day(t, tt.start), day(t, tt.today))
			if from.String() != tt.wantFrom || due != tt.wantDue {
				t.Errorf("nextFrom = %s, %v; want %s, %v", from, due, tt.wantFrom, tt.wantDue)
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
	cfg, err := (&rawConfig{Start: "2025-01-02", Instruments: []rawInstrument{{ISIN: fakewISIN}}}).parse()
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	today := day(t, "2025-01-03")
	c := &connector{
		cfg: cfg, secrets: secretsDir(t, fakeKey), today: func() kernel.Date { return today },
		source: func(token string) marketData { return newFakeEODHD(token) },
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
	if got := resp.Msg.GetDelivery().GetPrices(); len(got) != 2 || resp.Msg.GetWatermark() != `{"prices":{"XX0000FAKEW1":"2025-01-03"}}` {
		t.Errorf("Sync = %v, want two prices and FAKEW marked on 2025-01-03", resp.Msg)
	}
	unauthorized := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-2"})
	if _, err := client.Sync(t.Context(), unauthorized); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("Sync without the core's token: %v, want Unauthenticated", err)
	}
	stream := client.ImportFile(t.Context())
	stream.RequestHeader().Set("Authorization", "Bearer "+coreToken)
	if err := stream.Send(&connectorv1.ImportFileRequest{Part: &connectorv1.ImportFileRequest_Header{
		Header: &connectorv1.FileHeader{BatchId: "b-3", FileName: "prices.csv"},
	}}); err != nil && !errors.Is(err, context.Canceled) {
		t.Logf("send: %v", err)
	}
	if _, err := stream.CloseAndReceive(); connect.CodeOf(err) != connect.CodeUnimplemented {
		t.Errorf("ImportFile: %v, want Unimplemented", err)
	}
}
