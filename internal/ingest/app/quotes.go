package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
)

// Errors of a quote lookup (D-226).
var (
	// ErrNoQuoteSource reports that no market-data connector can look a quote
	// up: none is installed, running and approved (D-123), or none that is
	// answers quotes.
	ErrNoQuoteSource = errors.New("no market-data connector can look quotes up")
	// ErrQuoteNotFound reports a ticker or an ISIN the market-data source
	// has no listing for.
	ErrQuoteNotFound = errors.New("the market-data source has no such instrument")
)

// QuoteLookup asks a market-data connector instance for a quote (D-226).
// *connectrpc.Client implements it. A connector that can't quote wraps
// ErrNoEntries, one that can't be reached ErrConnectorUnavailable, an
// instrument it finds nothing for ErrQuoteNotFound, and a request it refuses
// ErrEntryRefused.
type QuoteLookup interface {
	LookupQuote(ctx context.Context, instance string, req ingest.QuoteRequest) (ingest.QuoteAnswer, error)
}

// PriceRanking tells the owner's ranking of the sources of prices (D-94): the
// connector instances' names, the highest-ranked first, as the owner placed
// them and then by default. *valuation/postgres.Store implements it through
// the ranking it keeps.
type PriceRanking interface {
	RankedPriceSources(ctx context.Context) ([]string, error)
}

// HeldSources tells which connector instance values an instrument the owner
// holds (D-94: one source of its values), so a quote of it comes from the same
// source and the model never sees two prices of one instrument.
// *valuation/postgres.Store implements it through the sources of its series.
type HeldSources interface {
	// ValuationSource returns the connector instance whose prices value the
	// instrument query names, a ticker or an ISIN, and "" when no stored
	// series of prices is of it.
	ValuationSource(ctx context.Context, query string) (string, error)
}

// Quotes looks a ticker or an ISIN up for its latest price and history
// through the owner's market-data connector (D-226). The core makes no
// request of its own: the connector plugin alone reaches its provider, under
// its egress grant. Nothing is stored, and no lookup is a sync.
type Quotes struct {
	Lookup QuoteLookup
	// Instances and Declarations find the market-data connectors: the running
	// instances whose manifest declares one and whose grant approves it
	// (D-123), as for an ISIN lookup. Without them none is found.
	Instances    InstanceLister
	Declarations ConnectorDeclarations
	// Ranking orders them; nil orders them by name.
	Ranking PriceRanking
	// Held finds the source of an instrument the owner holds; nil knows none.
	Held HeldSources
}

// Sources returns the running market-data connector instances in the order
// the owner ranks the sources of prices (D-94): the ranked ones first, in
// their order, then the others by name.
func (q *Quotes) Sources(ctx context.Context) ([]string, error) {
	if q.Instances == nil {
		return nil, nil
	}
	found, err := marketDataInstances(ctx, q.Instances, q.Declarations)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	var ranked []string
	if q.Ranking != nil {
		if ranked, err = q.Ranking.RankedPriceSources(ctx); err != nil {
			return nil, fmt.Errorf("ranking the sources of prices: %w", err)
		}
	}
	rank := func(name string) int {
		if i := slices.Index(ranked, name); i >= 0 {
			return i
		}
		return len(ranked)
	}
	slices.SortStableFunc(found, func(a, b string) int {
		if d := rank(a) - rank(b); d != 0 {
			return d
		}
		return cmp.Compare(a, b)
	})
	return found, nil
}

// Available reports whether a quote can be looked up: a market-data
// connector is running and approved.
func (q *Quotes) Available(ctx context.Context) (bool, error) {
	if q.Lookup == nil {
		return false, nil
	}
	sources, err := q.Sources(ctx)
	if err != nil {
		return false, err
	}
	return len(sources) > 0, nil
}

// Quote looks a quote up (D-226, D-94). An instrument the owner holds is asked
// of the source of its valuation and of no other, so its quote is the price it
// is valued at and the model never sees two prices of one instrument; if that
// source is not running or takes no quotes the lookup is ErrNoQuoteSource, and
// if it is down or finds nothing that is the answer. Any other instrument is
// asked of the owner's ranked sources in order, the first that knows it
// answering: a source that takes no quotes (ErrNoEntries) or has no listing
// (ErrQuoteNotFound) is passed over for the next, since another source may list
// it, but a source that is down stops the lookup, since an answer from a lower
// source would hide that the owner's first choice failed and the price would be
// one the owner did not rank first. With every source passed over it is
// ErrQuoteNotFound when one had no listing, else ErrNoQuoteSource.
func (q *Quotes) Quote(ctx context.Context, req ingest.QuoteRequest) (ingest.QuoteAnswer, error) {
	if q.Lookup == nil {
		return ingest.QuoteAnswer{}, ErrNoQuoteSource
	}
	sources, err := q.Sources(ctx)
	if err != nil {
		return ingest.QuoteAnswer{}, err
	}
	sources, err = q.heldOnly(ctx, req.Query, sources)
	if err != nil {
		return ingest.QuoteAnswer{}, err
	}
	notFound := false
	for _, source := range sources {
		answer, lookupErr := q.Lookup.LookupQuote(ctx, source, req)
		switch {
		case errors.Is(lookupErr, ErrNoEntries):
			continue
		case errors.Is(lookupErr, ErrQuoteNotFound):
			notFound = true
			continue
		case lookupErr != nil:
			return ingest.QuoteAnswer{}, fmt.Errorf("looking up a quote: %w", lookupErr)
		}
		return answer, nil
	}
	if notFound {
		return ingest.QuoteAnswer{}, ErrQuoteNotFound
	}
	return ingest.QuoteAnswer{}, ErrNoQuoteSource
}

// heldOnly narrows sources to the one source of the instrument query names
// when the owner holds it, and leaves them as they are when not. A held
// instrument whose source is not among them has no source.
func (q *Quotes) heldOnly(ctx context.Context, query string, sources []string) ([]string, error) {
	if q.Held == nil {
		return sources, nil
	}
	held, err := q.Held.ValuationSource(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("finding the source of a held instrument: %w", err)
	}
	switch {
	case held == "":
		return sources, nil
	case !slices.Contains(sources, held):
		return nil, ErrNoQuoteSource
	}
	return []string{held}, nil
}
