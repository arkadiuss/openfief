// Package eodhd is the market-data adapter for EODHD (D-22, D-51): an HTTP
// client for EODHD's search and end-of-day APIs. It finds an instrument's
// listing by ISIN on the Warsaw Stock Exchange, the LSE, Xetra or Euronext,
// and fetches the end-of-day closes of a listing or of an index. The owner
// brings the API key. EODHD's JSON shape never leaves this package, and
// every number is parsed from its text, never through a float.
package eodhd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

const (
	// DefaultBaseURL is EODHD's API.
	DefaultBaseURL = "https://eodhd.com"
	// searchLimit is the most search results asked for: enough for every
	// listing of one ISIN.
	searchLimit = 50
	// userAgent names OpenFief instead of Go's default.
	userAgent = "openfief"
	// maxBodyBytes bounds a response: thirty years of daily closes are
	// about 1.2 MB.
	maxBodyBytes = 16 << 20
)

var (
	// ErrNotFound reports a ticker EODHD doesn't know (HTTP 404).
	ErrNotFound = errors.New("EODHD has no such ticker")
	// ErrUnexpectedStatus reports any other HTTP status than 200 or 404,
	// such as 401 for a wrong API key or 402 when its plan's limit is used up.
	ErrUnexpectedStatus = errors.New("unexpected EODHD response status")
	// ErrMalformedResponse reports a body that is not what EODHD documents.
	ErrMalformedResponse = errors.New("malformed EODHD response")
)

// Close is the end-of-day close of a listing or the closing level of an
// index on one date, exactly as EODHD reports it, in the listing's quote
// currency.
type Close struct {
	Date  kernel.Date
	Close kernel.Decimal
}

// Client calls EODHD's API with the owner's API key.
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

// New returns a client that sends its requests with hc to baseURL, normally
// DefaultBaseURL, authenticated with the API key token. The caller owns hc:
// its timeout, transport and proxy (in a plugin, the egress proxy). The
// client reads no environment, and no error it returns contains the key.
func New(hc *http.Client, baseURL, token string) *Client {
	return &Client{http: hc, baseURL: baseURL, token: token}
}

// Listings returns the listings EODHD's search finds for isin, on any
// exchange; Pick chooses one. A search that finds nothing is empty.
func (c *Client) Listings(ctx context.Context, isin ledger.ISIN) ([]Listing, error) {
	q := url.Values{"limit": {strconv.Itoa(searchLimit)}}
	body, err := c.get(ctx, "/api/search/"+url.PathEscape(isin.String()), q)
	if errors.Is(err, ErrNotFound) {
		return []Listing{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("EODHD search for %s: %w", isin, err)
	}
	listings, err := parseListings(body)
	if err != nil {
		return nil, fmt.Errorf("EODHD search for %s: %w", isin, err)
	}
	return listings, nil
}

// Closes returns ticker's daily closes between the dates from and to, both
// inclusive, in date order. A zero from asks for the whole history EODHD has. Days
// without a positive close are left out. An unknown ticker is ErrNotFound.
func (c *Client) Closes(ctx context.Context, ticker string, from, to kernel.Date) ([]Close, error) {
	q := url.Values{"period": {"d"}, "order": {"a"}, "to": {to.String()}}
	if !from.IsZero() {
		q.Set("from", from.String())
	}
	body, err := c.get(ctx, "/api/eod/"+url.PathEscape(ticker), q)
	if err != nil {
		return nil, fmt.Errorf("EODHD closes of %s: %w", ticker, err)
	}
	closes, err := parseCloses(body)
	if err != nil {
		return nil, fmt.Errorf("EODHD closes of %s: %w", ticker, err)
	}
	return closes, nil
}

// get sends a GET for path with query plus the API key and JSON format, and
// returns the body of a 200. The API key is in the URL, so a transport
// error is returned without it.
func (c *Client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	query.Set("api_token", c.token)
	query.Set("fmt", "json")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+query.Encode(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", withoutURL(err))
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, withoutURL(err)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	closeErr := resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	case readErr != nil:
		return nil, fmt.Errorf("read response: %w", readErr)
	case closeErr != nil:
		return nil, fmt.Errorf("close response: %w", closeErr)
	}
	return body, nil
}

// withoutURL returns err without the request URL a *url.Error carries,
// which holds the API key.
func withoutURL(err error) error {
	if ue, ok := errors.AsType[*url.Error](err); ok {
		return fmt.Errorf("%s request: %w", ue.Op, ue.Err)
	}
	return err
}

// parseListings maps a search response: an array of results whose field
// names are capitalized, such as {"Code": "SYNT", "Exchange": "XETRA",
// "Currency": "EUR", "ISIN": "..."}. A result without a valid ISIN, such
// as a fuzzy match on the name, is left out.
func parseListings(body []byte) ([]Listing, error) {
	var results []map[string]json.RawMessage
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedResponse, err)
	}
	out := make([]Listing, 0, len(results))
	for _, r := range results {
		isin, err := ledger.ParseISIN(text(r["ISIN"]))
		if err != nil {
			continue
		}
		out = append(out, Listing{ISIN: isin, Code: text(r["Code"]), Exchange: text(r["Exchange"]), Currency: text(r["Currency"])})
	}
	return out, nil
}

// text is a JSON string's value, or "" for null, a missing field or any
// other JSON value.
func text(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}
	return s
}

// bar is one day of EODHD's end-of-day API. Close is kept as the number's
// text, so that no price ever passes through a float; adjusted_close,
// which back-adjusts for splits and dividends, is not used.
type bar struct {
	Date  string      `json:"date"`
	Close json.Number `json:"close"`
}

// parseCloses maps an end-of-day response, sorted by date. A day whose
// close is null or not positive, which EODHD reports for days without
// trading, is left out.
func parseCloses(body []byte) ([]Close, error) {
	var bars []bar
	if err := json.Unmarshal(body, &bars); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedResponse, err)
	}
	out := make([]Close, 0, len(bars))
	for _, b := range bars {
		on, err := kernel.ParseDate(b.Date)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMalformedResponse, err)
		}
		if b.Close == "" {
			continue
		}
		closing, err := kernel.ParseDecimal(b.Close.String())
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrMalformedResponse, on, err)
		}
		if closing.Sign() > 0 {
			out = append(out, Close{Date: on, Close: closing})
		}
	}
	slices.SortStableFunc(out, func(a, b Close) int { return a.Date.Compare(b.Date) })
	return out, nil
}
