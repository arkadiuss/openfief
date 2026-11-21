package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// errSynthetic is a failure of the lists a test makes.
var errSynthetic = errors.New("synthetic failure")

// quoteSources is the connector instances of a quote test, with the ones that
// declare themselves market-data connectors.
type quoteSources struct {
	instances   []string
	marketData  map[string]bool
	listErr     error
	rankedNames []string
	rankErr     error
	// held maps a query to the instance whose prices value the instrument it names.
	held map[string]string
}

func (q *quoteSources) ValuationSource(_ context.Context, query string) (string, error) {
	return q.held[query], nil
}

func (q *quoteSources) ConnectorInstances(context.Context) ([]string, error) {
	return q.instances, q.listErr
}

func (q *quoteSources) Declarations(_ context.Context, instance string) (ingest.Declarations, error) {
	return ingest.Declarations{MarketData: q.marketData[instance]}, nil
}

func (q *quoteSources) RankedPriceSources(context.Context) ([]string, error) {
	return q.rankedNames, q.rankErr
}

// quoteLookup answers per instance: what it returns, and the order it was asked in.
type quoteLookup struct {
	answers map[string]error
	asked   []string
}

func (l *quoteLookup) LookupQuote(_ context.Context, instance string, req ingest.QuoteRequest) (ingest.QuoteAnswer, error) {
	l.asked = append(l.asked, instance+":"+req.Query)
	if err := l.answers[instance]; err != nil {
		return ingest.QuoteAnswer{}, err
	}
	return ingest.QuoteAnswer{Source: instance}, nil
}

func newQuotes(s *quoteSources, l *quoteLookup) *app.Quotes {
	return &app.Quotes{Lookup: l, Instances: s, Declarations: s, Ranking: s, Held: s}
}

// TestQuotesAskTheRankedSource checks D-226 and D-94: the highest-ranked
// market-data connector is asked, the unranked ones after the ranked ones by
// name, and a connector that isn't declared and approved as a market-data one
// is never asked, as no ISIN lookup is either (D-123).
func TestQuotesAskTheRankedSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		sources   quoteSources
		answers   map[string]error
		want      []string
		wantFrom  string
		wantErr   error
		available bool
	}{
		{
			name:    "the owner's first",
			sources: quoteSources{instances: []string{"eodhd", "other", "xtb"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"other", "eodhd"}},
			want:    []string{"other:NVDA"}, wantFrom: "other", available: true,
		},
		{
			name:    "unranked ones by name",
			sources: quoteSources{instances: []string{"zeta", "alpha"}, marketData: map[string]bool{"zeta": true, "alpha": true}},
			want:    []string{"alpha:NVDA"}, wantFrom: "alpha", available: true,
		},
		{
			name:    "a source that takes no quotes is passed over",
			sources: quoteSources{instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"eodhd", "other"}},
			answers: map[string]error{"eodhd": fmt.Errorf("connector eodhd: %w", app.ErrNoEntries)},
			want:    []string{"eodhd:NVDA", "other:NVDA"}, wantFrom: "other", available: true,
		},
		{
			name:    "a source that is down is the answer, not another source",
			sources: quoteSources{instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"eodhd", "other"}},
			answers: map[string]error{"eodhd": fmt.Errorf("connector eodhd: %w", app.ErrConnectorUnavailable)},
			want:    []string{"eodhd:NVDA"}, wantErr: app.ErrConnectorUnavailable, available: true,
		},
		{
			name:    "a source with no listing is passed over for the next, which may list it",
			sources: quoteSources{instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"eodhd", "other"}},
			answers: map[string]error{"eodhd": app.ErrQuoteNotFound},
			want:    []string{"eodhd:NVDA", "other:NVDA"}, wantFrom: "other", available: true,
		},
		{
			name:    "no source lists it",
			sources: quoteSources{instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"eodhd", "other"}},
			answers: map[string]error{"eodhd": app.ErrQuoteNotFound, "other": app.ErrQuoteNotFound},
			want:    []string{"eodhd:NVDA", "other:NVDA"}, wantErr: app.ErrQuoteNotFound, available: true,
		},
		{
			name: "a held instrument is asked of the source it is valued from, though another ranks first",
			sources: quoteSources{
				instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"eodhd", "other"},
				held: map[string]string{"NVDA": "other"},
			},
			want: []string{"other:NVDA"}, wantFrom: "other", available: true,
		},
		{
			name: "a held instrument whose source finds nothing is not asked of another",
			sources: quoteSources{
				instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true}, rankedNames: []string{"eodhd", "other"},
				held: map[string]string{"NVDA": "other"},
			},
			answers: map[string]error{"other": app.ErrQuoteNotFound},
			want:    []string{"other:NVDA"}, wantErr: app.ErrQuoteNotFound, available: true,
		},
		{
			name: "a held instrument whose source is not running has no source",
			sources: quoteSources{
				instances: []string{"eodhd"}, marketData: map[string]bool{"eodhd": true},
				held: map[string]string{"NVDA": "gone"},
			},
			wantErr: app.ErrNoQuoteSource, available: true,
		},
		{
			name:    "every source takes no quotes",
			sources: quoteSources{instances: []string{"eodhd"}, marketData: map[string]bool{"eodhd": true}},
			answers: map[string]error{"eodhd": app.ErrNoEntries},
			want:    []string{"eodhd:NVDA"}, wantErr: app.ErrNoQuoteSource, available: true,
		},
		{
			name:    "no market-data connector is declared",
			sources: quoteSources{instances: []string{"xtb", "nbp"}},
			wantErr: app.ErrNoQuoteSource,
		},
		{name: "no connector runs", sources: quoteSources{}, wantErr: app.ErrNoQuoteSource},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lookup := &quoteLookup{answers: tt.answers}
			quotes := newQuotes(&tt.sources, lookup)
			got, err := quotes.Quote(t.Context(), ingest.QuoteRequest{Query: "NVDA"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Quote error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && got.Source != tt.wantFrom {
				t.Errorf("answered by %q, want %q", got.Source, tt.wantFrom)
			}
			if diff := cmp.Diff(tt.want, lookup.asked); diff != "" {
				t.Errorf("sources asked (-want +got):\n%s", diff)
			}
			ok, err := quotes.Available(t.Context())
			if err != nil || ok != tt.available {
				t.Errorf("Available = %v, %v; want %v", ok, err, tt.available)
			}
		})
	}
}

// TestQuotesWithoutAClientAskNobody checks that Quotes without a lookup, as in
// a demo, finds no source and asks nothing.
func TestQuotesWithoutAClientAskNobody(t *testing.T) {
	t.Parallel()
	s := &quoteSources{instances: []string{"eodhd"}, marketData: map[string]bool{"eodhd": true}}
	quotes := &app.Quotes{Instances: s, Declarations: s, Ranking: s}
	if _, err := quotes.Quote(t.Context(), ingest.QuoteRequest{Query: "NVDA"}); !errors.Is(err, app.ErrNoQuoteSource) {
		t.Errorf("Quote error = %v, want ErrNoQuoteSource", err)
	}
	if ok, err := quotes.Available(t.Context()); ok || err != nil {
		t.Errorf("Available = %v, %v; want false", ok, err)
	}
}

func TestQuotesFailWhenTheInstancesOrTheRankingCannotBeRead(t *testing.T) {
	t.Parallel()
	boom := errSynthetic
	base := quoteSources{instances: []string{"eodhd"}, marketData: map[string]bool{"eodhd": true}}
	list, rank := base, base
	list.listErr, rank.rankErr = boom, boom
	for name, s := range map[string]*quoteSources{"instances": &list, "ranking": &rank} {
		lookup := &quoteLookup{}
		if _, err := newQuotes(s, lookup).Quote(t.Context(), ingest.QuoteRequest{Query: "NVDA"}); !errors.Is(err, boom) {
			t.Errorf("%s: Quote error = %v, want the failure", name, err)
		}
		if len(lookup.asked) != 0 {
			t.Errorf("%s: asked %v, want nobody", name, lookup.asked)
		}
	}
}

// pricedLookup answers every query with the one close a source has for it, the
// latest price that source would value the instrument at.
type pricedLookup struct {
	closes map[string]string
}

func (l pricedLookup) LookupQuote(_ context.Context, instance string, _ ingest.QuoteRequest) (ingest.QuoteAnswer, error) {
	usd, err := kernel.ParseCurrency("USD")
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("currency: %w", err)
	}
	m, err := kernel.ParseMoney(l.closes[instance], usd)
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("close: %w", err)
	}
	day, err := kernel.ParseDate("2026-03-02")
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("day: %w", err)
	}
	return ingest.QuoteAnswer{Source: instance, Quote: &ingest.Quote{History: []ingest.QuotePoint{{Date: day, Close: m}}}}, nil
}

// TestAHeldQuoteIsTheValuationSourcesLatestPrice checks D-226 and D-94: the
// quote of an instrument the owner holds is the latest price of the source its
// valuation uses, though another source ranks first and would say another
// price, so the model never sees two prices of one instrument; an instrument
// the owner does not hold is the ranked first source's.
func TestAHeldQuoteIsTheValuationSourcesLatestPrice(t *testing.T) {
	t.Parallel()
	s := &quoteSources{
		instances: []string{"eodhd", "other"}, marketData: map[string]bool{"eodhd": true, "other": true},
		rankedNames: []string{"eodhd", "other"}, held: map[string]string{"NVDA": "other"},
	}
	quotes := &app.Quotes{
		Lookup:    pricedLookup{closes: map[string]string{"eodhd": "101.5", "other": "99.25"}},
		Instances: s, Declarations: s, Ranking: s, Held: s,
	}
	for query, want := range map[string]string{"NVDA": "99.25", "NOVA": "101.5"} {
		got, err := quotes.Quote(t.Context(), ingest.QuoteRequest{Query: query})
		if err != nil {
			t.Fatalf("Quote(%s): %v", query, err)
		}
		latest, ok := got.Quote.Latest()
		if !ok || latest.Close.String() != want+" USD" {
			t.Errorf("Quote(%s) latest = %v, want %s USD", query, latest.Close, want)
		}
	}
}
