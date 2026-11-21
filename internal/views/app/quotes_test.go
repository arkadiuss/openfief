package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/app"
)

// errSynthetic is a failure of a source's own a test makes up.
var errSynthetic = errors.New("synthetic")

// fakeQuotes answers a lookup with err, or with an answer from "eodhd", and
// keeps the requests.
type fakeQuotes struct {
	err       error
	available bool
	availErr  error
	asked     []ingest.QuoteRequest
}

func (f *fakeQuotes) Quote(_ context.Context, req ingest.QuoteRequest) (ingest.QuoteAnswer, error) {
	f.asked = append(f.asked, req)
	if f.err != nil {
		return ingest.QuoteAnswer{}, f.err
	}
	return ingest.QuoteAnswer{Source: "eodhd"}, nil
}

func (f *fakeQuotes) Available(context.Context) (bool, error) { return f.available, f.availErr }

// quoteViews returns the views over a fixture with the quote source and a
// clock that stands on 2026-11-20.
func quoteViews(t *testing.T, cfg *app.MarketQuotesConfig) *app.Views {
	t.Helper()
	f := newFixture(t)
	cfg.Today = func() kernel.Date { return date(t, "2026-11-20") }
	v, err := app.New(f.ledger, f.valuations, f.statuses, f.rates, cur(t, "PLN"), cur(t, "PLN"), app.WithMarketQuotes(cfg))
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	return v
}

// TestMarketQuote checks D-226's lookup: the source is asked with the
// query and the period resolved on today, 30 days by default, and what it
// answers comes back with its source.
func TestMarketQuote(t *testing.T) {
	t.Parallel()
	src := &fakeQuotes{available: true}
	v := quoteViews(t, &app.MarketQuotesConfig{Source: src})
	got, err := v.MarketQuote(reader(t), " NVDA ", kernel.Date{}, kernel.Date{})
	if err != nil || got.Source != "eodhd" {
		t.Fatalf("MarketQuote = %+v, %v; want an answer from eodhd", got, err)
	}
	if len(src.asked) != 1 || src.asked[0].Query != "NVDA" || src.asked[0].From.String() != "2026-10-22" || src.asked[0].To.String() != "2026-11-20" {
		t.Errorf("the source was asked %+v, want NVDA over the last 30 days", src.asked)
	}
	if !v.MarketQuotesAvailable(reader(t)) {
		t.Error("MarketQuotesAvailable = false with a running source")
	}
}

func TestMarketQuoteRefuses(t *testing.T) {
	t.Parallel()
	day := date
	tests := []struct {
		name     string
		query    string
		from, to string
		err      error
		want     error
		asked    bool
	}{
		{name: "no query", query: "", want: ingest.ErrInvalidQuoteQuery},
		{name: "a query that is no ticker", query: "NVDA?x=1", want: ingest.ErrInvalidQuoteQuery},
		{name: "a period in the future", query: "NVDA", to: "2026-11-21", want: ingest.ErrInvalidQuotePeriod},
		{name: "a period over a year", query: "NVDA", from: "2025-01-01", to: "2026-11-20", want: ingest.ErrInvalidQuotePeriod},
		{name: "no market-data connector", query: "NVDA", err: fmt.Errorf("x: %w", ingestapp.ErrNoQuoteSource), want: app.ErrQuotesUnavailable, asked: true},
		{name: "no listing", query: "NVDA", err: fmt.Errorf("x: %w", ingestapp.ErrQuoteNotFound), want: app.ErrQuoteNotFound, asked: true},
		{name: "the connector is down", query: "NVDA", err: fmt.Errorf("x: %w", ingestapp.ErrConnectorUnavailable), want: app.ErrQuoteSourceDown, asked: true},
		{name: "the connector refuses", query: "NVDA", err: fmt.Errorf("x: %w", ingestapp.ErrEntryRefused), want: app.ErrQuoteSourceDown, asked: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			src := &fakeQuotes{err: tt.err, available: true}
			v := quoteViews(t, &app.MarketQuotesConfig{Source: src})
			var from, to kernel.Date
			if tt.from != "" {
				from = day(t, tt.from)
			}
			if tt.to != "" {
				to = day(t, tt.to)
			}
			if _, err := v.MarketQuote(reader(t), tt.query, from, to); !errors.Is(err, tt.want) {
				t.Errorf("MarketQuote error = %v, want %v", err, tt.want)
			}
			if (len(src.asked) > 0) != tt.asked {
				t.Errorf("the source was asked %d times, want asked %v: an invalid request must never reach the connector", len(src.asked), tt.asked)
			}
		})
	}
}

// TestMarketQuoteUnavailableWithoutASource checks that views with no source,
// as on a demo, or one that can't say whether a connector runs, offer no
// quotes and ask nothing.
func TestMarketQuoteUnavailableWithoutASource(t *testing.T) {
	t.Parallel()
	for name, cfg := range map[string]*app.MarketQuotesConfig{
		"no source":           {},
		"availability failed": {Source: &fakeQuotes{availErr: errSynthetic}},
		"no connector runs":   {Source: &fakeQuotes{}},
	} {
		v := quoteViews(t, cfg)
		if v.MarketQuotesAvailable(reader(t)) {
			t.Errorf("%s: MarketQuotesAvailable = true", name)
		}
	}
	v := quoteViews(t, &app.MarketQuotesConfig{})
	if _, err := v.MarketQuote(reader(t), "NVDA", kernel.Date{}, kernel.Date{}); !errors.Is(err, app.ErrQuotesUnavailable) {
		t.Errorf("MarketQuote with no source = %v, want ErrQuotesUnavailable", err)
	}
}

// TestMarketQuoteIsLimitedPerMinute checks that lookups beyond the limit of a
// minute are refused without asking the source, and that the next minute
// counts afresh: a looping token or model can't spend the owner's provider key.
func TestMarketQuoteIsLimitedPerMinute(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 20, 10, 0, 5, 0, time.UTC)
	src := &fakeQuotes{available: true}
	v := quoteViews(t, &app.MarketQuotesConfig{Source: src, PerMinute: 2, Now: func() time.Time { return now }})
	ask := func() error {
		if _, err := v.MarketQuote(reader(t), "NVDA", kernel.Date{}, kernel.Date{}); err != nil {
			return fmt.Errorf("lookup: %w", err)
		}
		return nil
	}
	for i := range 2 {
		if err := ask(); err != nil {
			t.Fatalf("lookup %d: %v", i+1, err)
		}
	}
	if err := ask(); !errors.Is(err, app.ErrQuoteRateLimited) {
		t.Fatalf("the third lookup of a minute = %v, want ErrQuoteRateLimited", err)
	}
	if len(src.asked) != 2 {
		t.Errorf("the source was asked %d times, want 2", len(src.asked))
	}
	now = now.Add(time.Minute)
	if err := ask(); err != nil {
		t.Errorf("a lookup of the next minute = %v, want it allowed", err)
	}
}

// TestMarketQuoteIsLimitedPerCaller checks that one caller spending its minute
// leaves the others theirs, and that all callers together are capped too.
func TestMarketQuoteIsLimitedPerCaller(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 20, 10, 0, 5, 0, time.UTC)
	src := &fakeQuotes{available: true}
	v := quoteViews(t, &app.MarketQuotesConfig{Source: src, PerCaller: 1, PerMinute: 3, Now: func() time.Time { return now }})
	as := func(id string) context.Context {
		return identityapp.WithCaller(t.Context(), identity.Caller{Holder: identity.HolderScriptToken, ID: id})
	}
	ask := func(ctx context.Context) error {
		_, err := v.MarketQuote(ctx, "NVDA", kernel.Date{}, kernel.Date{})
		if err != nil {
			return fmt.Errorf("lookup: %w", err)
		}
		return nil
	}
	if err := ask(as("a")); err != nil {
		t.Fatalf("first lookup of a: %v", err)
	}
	if err := ask(as("a")); !errors.Is(err, app.ErrQuoteRateLimited) {
		t.Errorf("second lookup of a = %v, want ErrQuoteRateLimited", err)
	}
	for _, id := range []string{"b", "c"} {
		if err := ask(as(id)); err != nil {
			t.Errorf("first lookup of %s = %v, want it allowed: a's spending is not theirs", id, err)
		}
	}
	if err := ask(as("d")); !errors.Is(err, app.ErrQuoteRateLimited) {
		t.Errorf("lookup past the install's cap = %v, want ErrQuoteRateLimited", err)
	}
}

// TestMarketQuoteIsReadByWhoD226Names checks the owner, a script token and a
// viewer may read quotes, and a view plugin and a demo session may not, with
// the source never asked for them.
func TestMarketQuoteIsReadByWhoD226Names(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		holder  identity.Holder
		allowed bool
	}{
		{identity.HolderOwner, true},
		{identity.HolderScriptToken, true},
		{identity.HolderViewer, true},
		{identity.HolderViewPlugin, false},
		{identity.HolderDemoSession, false},
	} {
		src := &fakeQuotes{available: true}
		v := quoteViews(t, &app.MarketQuotesConfig{Source: src})
		ctx := identityapp.WithCaller(t.Context(), identity.Caller{Holder: tt.holder, ID: "x"})
		_, err := v.MarketQuote(ctx, "NVDA", kernel.Date{}, kernel.Date{})
		if tt.allowed != (err == nil) || (!tt.allowed && !errors.Is(err, app.ErrQuotesForbidden)) {
			t.Errorf("holder %v: error = %v, want allowed %v", tt.holder, err, tt.allowed)
		}
		if got := v.MarketQuotesAvailable(ctx); got != tt.allowed {
			t.Errorf("holder %v: MarketQuotesAvailable = %v, want %v", tt.holder, got, tt.allowed)
		}
		if !tt.allowed && len(src.asked) != 0 {
			t.Errorf("holder %v: the source was asked", tt.holder)
		}
	}
}
