package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestapp "github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of a market quote (D-226).
var (
	// ErrQuotesUnavailable reports that quotes can't be looked up here: no
	// market-data connector is installed, running and approved, or this
	// install is a demo, which answers from its synthetic data alone.
	ErrQuotesUnavailable = errors.New("market quotes are not available")
	// ErrQuoteNotFound reports a ticker or an ISIN the owner's market-data
	// source has no listing for.
	ErrQuoteNotFound = errors.New("no such instrument at the market-data source")
	// ErrQuoteSourceDown reports a market-data connector or its provider that
	// can't answer now.
	ErrQuoteSourceDown = errors.New("the market-data source can't answer now")
	// ErrQuoteRateLimited reports that the caller, or all callers together,
	// asked more lookups in a minute than the owner's provider key should spend
	// on questions.
	ErrQuoteRateLimited = errors.New("too many market quotes asked in a minute")
	// ErrQuotesForbidden reports a caller that may not read market quotes: a
	// view plugin, whose manifest does not ask for them, and a demo session.
	ErrQuotesForbidden = errors.New("this caller may not read market quotes")
)

// MarketQuotes looks a ticker or an ISIN up through the owner's market-data
// connector, the highest-ranked source of prices (D-94, D-226).
// *ingest/app.Quotes implements it.
type MarketQuotes interface {
	// Quote answers the request, or wraps ingest/app's ErrNoQuoteSource,
	// ErrQuoteNotFound, ErrConnectorUnavailable or ErrEntryRefused.
	Quote(ctx context.Context, req ingest.QuoteRequest) (ingest.QuoteAnswer, error)
	// Available reports whether any market-data connector is running and
	// approved to be asked.
	Available(ctx context.Context) (bool, error)
}

// The limits of market quote lookups in a minute unless MarketQuotesConfig says
// otherwise: per caller (a session, a script token), so one script that loops,
// or a model, can't spend the owner's provider key or starve the owner's own
// lookups, and, as a cap on the install, all callers together.
const (
	DefaultQuotesPerCaller = 10
	DefaultQuotesPerMinute = 60
)

// MarketQuotesConfig is what WithMarketQuotes needs.
type MarketQuotesConfig struct {
	// Source looks quotes up; nil leaves them unavailable, as on a demo.
	Source MarketQuotes
	// Today is the date of a lookup, in the install's time zone.
	Today func() kernel.Date
	// Now is the clock the limit counts by; nil is time.Now.
	Now func() time.Time
	// PerMinute is the limit of lookups in a minute, all callers together, and
	// PerCaller each caller's; zero is DefaultQuotesPerMinute and
	// DefaultQuotesPerCaller.
	PerMinute, PerCaller int
}

// WithMarketQuotes has the views look instruments' latest prices and
// histories up through cfg.Source (D-226). Without it, or with no source,
// MarketQuote is ErrQuotesUnavailable and MarketQuotesAvailable false.
func WithMarketQuotes(cfg *MarketQuotesConfig) Option {
	return func(v *Views) {
		if cfg.Source == nil {
			return
		}
		limiter := &quoteLimiter{now: cfg.Now, max: cfg.PerMinute, perCaller: cfg.PerCaller}
		if limiter.now == nil {
			limiter.now = time.Now
		}
		if limiter.max <= 0 {
			limiter.max = DefaultQuotesPerMinute
		}
		if limiter.perCaller <= 0 {
			limiter.perCaller = DefaultQuotesPerCaller
		}
		today := cfg.Today
		if today == nil {
			today = func() kernel.Date { return utcDate(limiter.now()) }
		}
		v.quotes = &marketQuotes{source: cfg.Source, today: today, limit: limiter}
	}
}

// utcDate is the UTC calendar date of t.
func utcDate(t time.Time) kernel.Date {
	y, m, d := t.UTC().Date()
	date, err := kernel.NewDate(y, m, d)
	if err != nil { // a calendar date of the clock is always valid
		return kernel.Date{}
	}
	return date
}

// marketQuotes is the views' market quote lookup.
type marketQuotes struct {
	source MarketQuotes
	today  func() kernel.Date
	limit  *quoteLimiter
}

// quoteLimiter counts the lookups of the current minute, in all and per caller.
type quoteLimiter struct {
	now       func() time.Time
	max       int
	perCaller int

	mu     sync.Mutex
	window time.Time
	count  int
	byCall map[string]int
}

// allow counts a lookup of the caller named key and reports whether it may go
// on: neither the caller's minute nor the install's is spent.
func (l *quoteLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if now := l.now().Truncate(time.Minute); !now.Equal(l.window) {
		l.window, l.count, l.byCall = now, 0, map[string]int{}
	}
	if l.count >= l.max || l.byCall[key] >= l.perCaller {
		return false
	}
	l.count++
	l.byCall[key]++
	return true
}

// mayReadQuotes reports whether the caller in ctx may read market quotes
// (identity.MayReadQuotes, D-226).
func mayReadQuotes(ctx context.Context) bool {
	return identity.MayReadQuotes(identityapp.CallerFrom(ctx))
}

// callerKey names the caller in ctx for the limit.
func callerKey(ctx context.Context) string {
	c := identityapp.CallerFrom(ctx)
	return strconv.Itoa(int(c.Holder)) + ":" + c.ID
}

// MarketQuotesAvailable reports whether a market quote can be looked up: the
// views have a source and a market-data connector is running and approved. A
// source that can't tell is not available.
func (v *Views) MarketQuotesAvailable(ctx context.Context) bool {
	if v.quotes == nil || !mayReadQuotes(ctx) {
		return false
	}
	ok, err := v.quotes.source.Available(ctx)
	return err == nil && ok
}

// MarketQuote looks up the instrument query names, a ticker or an ISIN, and
// returns its latest end-of-day price with the closes of the period from the
// first day to the last (D-226): by default the last 30 days, at most 366, ending today or earlier.
// The lookup goes to the owner's highest-ranked market-data connector, whose
// plugin alone reaches the provider; the answer is public market data (tier
// T2, D-206), carries the source's name, and is never stored. An invalid query
// or period wraps ingest.ErrInvalidQuoteQuery or ingest.ErrInvalidQuotePeriod;
// without a source ErrQuotesUnavailable; a query the source finds nothing for
// ErrQuoteNotFound; a source that is down ErrQuoteSourceDown; and a minute's
// lookups spent ErrQuoteRateLimited.
func (v *Views) MarketQuote(ctx context.Context, query string, from, to kernel.Date) (ingest.QuoteAnswer, error) {
	if v.quotes == nil {
		return ingest.QuoteAnswer{}, ErrQuotesUnavailable
	}
	if !mayReadQuotes(ctx) {
		return ingest.QuoteAnswer{}, ErrQuotesForbidden
	}
	q, err := ingest.ParseQuoteQuery(query)
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("looking up a quote: %w", err)
	}
	period, err := ingest.QuotePeriod(from, to, v.quotes.today())
	if err != nil {
		return ingest.QuoteAnswer{}, fmt.Errorf("looking up a quote: %w", err)
	}
	if !v.quotes.limit.allow(callerKey(ctx)) {
		return ingest.QuoteAnswer{}, ErrQuoteRateLimited
	}
	answer, err := v.quotes.source.Quote(ctx, ingest.QuoteRequest{Query: q, From: period.From, To: period.To})
	switch {
	case errors.Is(err, ingestapp.ErrNoQuoteSource):
		return ingest.QuoteAnswer{}, ErrQuotesUnavailable
	case errors.Is(err, ingestapp.ErrQuoteNotFound):
		return ingest.QuoteAnswer{}, ErrQuoteNotFound
	case errors.Is(err, ingestapp.ErrConnectorUnavailable), errors.Is(err, ingestapp.ErrEntryRefused):
		return ingest.QuoteAnswer{}, ErrQuoteSourceDown
	case err != nil:
		return ingest.QuoteAnswer{}, fmt.Errorf("looking up a quote: %w", err)
	}
	return answer, nil
}
