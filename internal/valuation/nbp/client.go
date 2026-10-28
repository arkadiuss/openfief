// Package nbp is the default FX rate provider (D-20): an HTTP adapter for
// the National Bank of Poland's public Web API that fetches table A, the
// daily average exchange rates, and maps it to valuation.FXRate. NBP's JSON
// shape never leaves this package.
package nbp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

const (
	// DefaultBaseURL is NBP's public Web API.
	DefaultBaseURL = "https://api.nbp.pl"
	// MaxRangeDays is the most days, inclusive, that NBP serves in one
	// request; pass it to app.FetchRates.
	MaxRangeDays = 93
	// Source names NBP table A on every rate this adapter returns.
	Source = "nbp-a"

	tableA = "A"
	// userAgent names OpenFief: api.nbp.pl answers 403 to Go's default
	// "Go-http-client/1.1".
	userAgent = "openfief"
	// maxBodyBytes bounds a response: a full 93-day table A is about 200 KB.
	maxBodyBytes = 4 << 20
)

var (
	// ErrUnexpectedStatus reports an HTTP status other than 200 or 404.
	ErrUnexpectedStatus = errors.New("unexpected NBP response status")
	// ErrMalformedTable reports a response that is not a valid table A.
	ErrMalformedTable = errors.New("malformed NBP table A")
)

// Client fetches table A from NBP. It implements app.RateProvider.
type Client struct {
	http    *http.Client
	baseURL string
}

// New returns a client that sends its requests with hc to baseURL, normally
// DefaultBaseURL. The caller owns hc: its timeout, transport and proxy (in a
// plugin, HTTPS_PROXY). The client reads no environment itself.
func New(hc *http.Client, baseURL string) *Client {
	return &Client{http: hc, baseURL: baseURL}
}

// Rates returns table A's rates for each business day in [from, to], both
// inclusive, a range of at most MaxRangeDays days. NBP answers 404 when no table falls
// in the range, such as a weekend or a holiday; that is an empty result.
func (c *Client) Rates(ctx context.Context, from, to kernel.Date) ([]valuation.FXRate, error) {
	if to.Before(from) || from.DaysUntil(to) >= MaxRangeDays {
		return nil, fmt.Errorf("%w: %s to %s, NBP serves at most %d days", app.ErrInvalidRange, from, to, MaxRangeDays)
	}
	url := fmt.Sprintf("%s/api/exchangerates/tables/%s/%s/%s/?format=json", c.baseURL, tableA, from, to)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build NBP request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NBP table A %s to %s: %w", from, to, err)
	}
	rates, err := readResponse(resp)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("NBP table A %s to %s: %w", from, to, err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close NBP response: %w", closeErr)
	}
	return rates, nil
}

// readResponse maps NBP's answer: 200 carries tables, 404 means no table
// in the range, anything else is an error.
func readResponse(resp *http.Response) ([]valuation.FXRate, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		return parseTables(io.LimitReader(resp.Body, maxBodyBytes))
	case http.StatusNotFound:
		return []valuation.FXRate{}, nil
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}
}

// table and rate are NBP's JSON shape. Mid is kept as the number's text so
// that no rate ever passes through a float.
type table struct {
	Table         string `json:"table"`
	EffectiveDate string `json:"effectiveDate"`
	Rates         []rate `json:"rates"`
}

type rate struct {
	Code string      `json:"code"`
	Mid  json.Number `json:"mid"`
}

func parseTables(body io.Reader) ([]valuation.FXRate, error) {
	var tables []table
	if err := json.NewDecoder(body).Decode(&tables); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedTable, err)
	}
	var out []valuation.FXRate
	for _, t := range tables {
		if t.Table != tableA {
			return nil, fmt.Errorf("%w: table %q", ErrMalformedTable, t.Table)
		}
		on, err := kernel.ParseDate(t.EffectiveDate)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMalformedTable, err)
		}
		for _, r := range t.Rates {
			fx, err := mapRate(on, r)
			if err != nil {
				return nil, err
			}
			out = append(out, fx)
		}
	}
	return out, nil
}

// mapRate translates one NBP rate. Table A quotes every currency per one
// unit, so no scaling applies.
func mapRate(on kernel.Date, r rate) (valuation.FXRate, error) {
	c, err := kernel.ParseCurrency(r.Code)
	if err != nil {
		return valuation.FXRate{}, fmt.Errorf("%w: %s: %w", ErrMalformedTable, on, err)
	}
	mid, err := kernel.ParseDecimal(r.Mid.String())
	if err != nil {
		return valuation.FXRate{}, fmt.Errorf("%w: %s %s: %w", ErrMalformedTable, c, on, err)
	}
	if mid.Sign() <= 0 {
		return valuation.FXRate{}, fmt.Errorf("%w: %s %s: rate %s is not positive", ErrMalformedTable, c, on, mid)
	}
	return valuation.FXRate{Currency: c, Date: on, Rate: mid, Source: Source}, nil
}
