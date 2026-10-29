// Package gus is an HTTP adapter for the public API of Statistics Poland
// (GUS), the SDP API at https://api-sdp.stat.gov.pl: it fetches consumer
// prices year on year (CPI), which set the rates of inflation-indexed
// treasury bonds after their first interest period (D-9), and maps them to
// valuation.IndexReading. GUS's JSON shape never leaves this package.
//
// The API takes anonymous requests, within limits of 5 a second, 100 in 15
// minutes and 10,000 in 7 days per client; a registered key raises them
// (the API's own description, https://api-sdp.stat.gov.pl/apidocs/index.html).
package gus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
)

const (
	// DefaultBaseURL is GUS's public SDP API.
	DefaultBaseURL = "https://api-sdp.stat.gov.pl"
	// Source names GUS CPI on every reading this adapter returns.
	Source = "gus-cpi"

	// GUS publishes CPI by the COICOP 2018 classification from January
	// 2026's data on. The SDP API keeps the two series as two indicators,
	// both "Wskaźnik cen towarów i usług konsumpcyjnych" for Poland and all
	// goods and services, way of presentation 5, "analogiczny okres roku
	// poprzedniego=100" (the same month a year before = 100):
	// indicatorCOICOP1999 (cross-section 739) up to 2025 and
	// indicatorCOICOP2018 (cross-section 1698, all households) from 2026.
	// Each answers 404 for the other's years.
	indicatorCOICOP1999 = 1832
	indicatorCOICOP2018 = 2728
	firstCOICOP2018Year = 2026
	yearOnYear          = 5
	// baseIndex is the index of the same month a year before.
	baseIndex = 100

	// januaryPeriod is the SDP period id of January's monthly data (symbol
	// M01); February to December follow it, up to M12.
	januaryPeriod = 247

	// userAgent names OpenFief.
	userAgent = "openfief"
	// maxBodyBytes bounds a response: a year of one indicator is about 5 KB.
	maxBodyBytes = 1 << 20
)

// Flags GUS marks a value with when it is not yet the final reading
// (/api/dictionaries/flag-dictionary): an estimate, a forecast and
// preliminary data.
const (
	flagEstimate    = 29
	flagForecast    = 30
	flagPreliminary = 31
)

var (
	// ErrUnexpectedStatus reports an HTTP status other than 200 or 404.
	ErrUnexpectedStatus = errors.New("unexpected GUS response status")
	// ErrMalformedData reports a response that is not the CPI indicator's
	// data for the year asked for.
	ErrMalformedData = errors.New("malformed GUS indicator data")
)

// Client fetches CPI from GUS.
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

// CPI returns the final year-on-year CPI readings GUS has published for the
// months of year, in month order, each a yearly fraction for the month it
// measures: 104.9 (the same month a year before = 100) is 0.049. It stops
// before the first month that has no value yet or only an estimate, a
// forecast or preliminary data, so later months are never returned ahead of
// it. A year with nothing published, which GUS answers with 404, has no
// readings.
func (c *Client) CPI(ctx context.Context, year int) ([]valuation.IndexReading, error) {
	url := fmt.Sprintf("%s/api/indicators/indicator-data-indicator?id-wskaznik=%d&id-rok=%d&lang=en", c.baseURL, indicator(year), year)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build GUS request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GUS CPI %d: %w", year, err)
	}
	readings, err := readResponse(resp, year)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("GUS CPI %d: %w", year, err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close GUS response: %w", closeErr)
	}
	return readings, nil
}

// readResponse maps GUS's answer: 200 carries the year's values, 404 means
// none are published, anything else is an error.
func readResponse(resp *http.Response, year int) ([]valuation.IndexReading, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		return parseValues(io.LimitReader(resp.Body, maxBodyBytes), year)
	case http.StatusNotFound:
		return []valuation.IndexReading{}, nil
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}
}

// GUS's JSON keys of one indicator value. They don't follow Go's JSON
// naming, so values are decoded as maps and read by these keys.
const (
	keyYear         = "id-daty"
	keyPeriod       = "id-okres"
	keyPresentation = "id-sposob-prezentacji-miara"
	keyFlag         = "id-flaga"
	keyIndex        = "wartosc"
)

// value is one value of an indicator. The index is kept as the number's
// text so that it never passes through a float, and is nil when GUS has
// none.
type value struct {
	year, period, presentation, flag int
	index                            *json.Number
}

// decodeValue reads one value from its JSON object.
func decodeValue(obj map[string]*json.Number) (value, error) {
	v := value{index: obj[keyIndex]}
	for key, dst := range map[string]*int{keyYear: &v.year, keyPeriod: &v.period, keyPresentation: &v.presentation, keyFlag: &v.flag} {
		n := obj[key]
		if n == nil {
			return value{}, fmt.Errorf("%w: a value without %s", ErrMalformedData, key)
		}
		i, err := strconv.Atoi(n.String())
		if err != nil {
			return value{}, fmt.Errorf("%w: %s: %w", ErrMalformedData, key, err)
		}
		*dst = i
	}
	return v, nil
}

// parseValues reads year's monthly values, in month order, up to the first
// that is not final.
func parseValues(body io.Reader, year int) ([]valuation.IndexReading, error) {
	var objs []map[string]*json.Number
	if err := json.NewDecoder(body).Decode(&objs); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedData, err)
	}
	byMonth, err := monthly(objs, year)
	if err != nil {
		return nil, err
	}
	out := make([]valuation.IndexReading, 0, len(byMonth))
	for m := time.January; m <= time.December; m++ {
		v, ok := byMonth[m]
		if !ok || !v.final() {
			break // a month missing or not final: none after it either
		}
		r, err := reading(year, m, *v.index)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// monthly returns year's values by month, leaving out values of other
// periods, such as cumulative ones.
func monthly(objs []map[string]*json.Number, year int) (map[time.Month]value, error) {
	byMonth := map[time.Month]value{}
	for _, obj := range objs {
		v, err := decodeValue(obj)
		if err != nil {
			return nil, err
		}
		month := time.Month(v.period - januaryPeriod + 1)
		if month < time.January || month > time.December {
			continue
		}
		if v.year != year || v.presentation != yearOnYear {
			return nil, fmt.Errorf("%w: a value for %d in presentation %d, want %d in %d", ErrMalformedData, v.year, v.presentation, year, yearOnYear)
		}
		if _, dup := byMonth[month]; dup {
			return nil, fmt.Errorf("%w: two values for %d-%02d", ErrMalformedData, year, int(month))
		}
		byMonth[month] = v
	}
	return byMonth, nil
}

// final reports whether v has a value that is not an estimate, a forecast or
// preliminary.
func (v *value) final() bool {
	switch v.flag {
	case flagEstimate, flagForecast, flagPreliminary:
		return false
	}
	return v.index != nil
}

// reading is the CPI reading for month of year from GUS's index, where the
// same month a year before is 100: 104.9 becomes 0.049.
func reading(year int, month time.Month, index json.Number) (valuation.IndexReading, error) {
	v, err := kernel.ParseDecimal(index.String())
	if err != nil {
		return valuation.IndexReading{}, fmt.Errorf("%w: %d-%02d: %w", ErrMalformedData, year, int(month), err)
	}
	if v.Sign() <= 0 {
		return valuation.IndexReading{}, fmt.Errorf("%w: %d-%02d: index %s is not positive", ErrMalformedData, year, int(month), v)
	}
	base, percent := kernel.NewDecimal(baseIndex, 0), kernel.NewDecimal(1, -2)
	return valuation.IndexReading{
		Kind: valuation.CPI, Year: year, Month: month,
		Value: v.Sub(base).Mul(percent), Source: Source,
	}, nil
}

// indicator is the SDP indicator that holds year's CPI.
func indicator(year int) int {
	if year >= firstCOICOP2018Year {
		return indicatorCOICOP2018
	}
	return indicatorCOICOP1999
}
