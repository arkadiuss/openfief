package treasurybonds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

const (
	// DefaultGovPLBaseURL is gov.pl, which hosts the ministry's page.
	DefaultGovPLBaseURL = "https://www.gov.pl"
	// DefaultDaneBaseURL is the API of dane.gov.pl, Poland's open data
	// portal.
	DefaultDaneBaseURL = "https://api.dane.gov.pl"
	// govPLPage is the path of the ministry's retail bond page on gov.pl.
	govPLPage = "/web/finanse/obligacje-detaliczne1"
	// daneResourcesPath lists the resources of dataset 805, "Podstawowe
	// informacje na temat obligacji detalicznych".
	daneResourcesPath = "/1.4/datasets/805/resources"

	// Publisher is the source the attribution names: the ministry.
	Publisher = "Ministerstwo Finansów"
	// License is the licence of the ministry's data.
	License = "CC BY 4.0"
	// PublishedProcessing is what the reader does to the data.
	PublishedProcessing = "OpenFief read each series' code, first-period rate (the first \"Oprocentowanie\" column) " +
		"and margin (\"Marża\") from the sheet of its bond type, rounded them half even to six decimal places " +
		"as fractions, and left out every other column"

	// userAgent names OpenFief.
	userAgent = "openfief"
	// maxPageBytes bounds the gov.pl page (about 40 KB) and maxListBytes
	// the dane.gov.pl resource list (about 2 KB).
	maxPageBytes = 2 << 20
	maxListBytes = 1 << 20
	// maxFileBytes bounds the spreadsheet: about 1 MB in 2026, holding
	// every series since 1992.
	maxFileBytes = 16 << 20
	// cacheFor is how long a fetched spreadsheet is reused: the ministry
	// adds the next month's series once a month.
	cacheFor = 6 * time.Hour
)

// ErrUnexpectedStatus reports an HTTP status other than 200.
var ErrUnexpectedStatus = errors.New("unexpected response status")

// MinistryConfig is where MinistryClient reads the ministry's data from.
type MinistryConfig struct {
	// GovPL sends requests to GovPLBaseURL, normally
	// DefaultGovPLBaseURL; Dane to DaneBaseURL, normally DefaultDaneBaseURL.
	GovPL        *http.Client
	GovPLBaseURL string
	Dane         *http.Client
	DaneBaseURL  string
	// Now is the clock, which dates the fetch.
	Now func() time.Time
}

// MinistryClient fetches the ministry's published retail bond data (D-97),
// and keeps it for cacheFor; it is the PublishedSource the plugin uses.
type MinistryClient struct {
	cfg MinistryConfig

	mu      sync.Mutex
	cached  *pl.Published
	fetched time.Time
}

// NewMinistryClient returns a client reading as cfg says.
func NewMinistryClient(cfg MinistryConfig) *MinistryClient { return &MinistryClient{cfg: cfg} }

// Published returns the ministry's published data: from the gov.pl page's
// spreadsheet, or, when anything about that fails, from dane.gov.pl's. Data
// fetched less than cacheFor ago is returned again. When both fail, the
// error joins both failures.
func (c *MinistryClient) Published(ctx context.Context) (*pl.Published, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.cfg.Now()
	if c.cached != nil && now.Sub(c.fetched) < cacheFor {
		return c.cached, nil
	}
	p, govErr := c.fromGovPL(ctx, now)
	if govErr != nil {
		var daneErr error
		if p, daneErr = c.fromDane(ctx, now); daneErr != nil {
			return nil, fmt.Errorf("ministry retail bond data: %w", errors.Join(govErr, daneErr))
		}
	}
	c.cached, c.fetched = p, now
	return p, nil
}

// fromGovPL reads the spreadsheet the gov.pl page links, made on the day
// its Last-Modified header names.
func (c *MinistryClient) fromGovPL(ctx context.Context, now time.Time) (*pl.Published, error) {
	page, err := url.Parse(c.cfg.GovPLBaseURL + govPLPage)
	if err != nil {
		return nil, fmt.Errorf("gov.pl page: %w", err)
	}
	body, _, err := get(ctx, c.cfg.GovPL, page, maxPageBytes)
	if err != nil {
		return nil, fmt.Errorf("gov.pl page: %w", err)
	}
	file, err := discoverGovPL(bytes.NewReader(body), page)
	if err != nil {
		return nil, fmt.Errorf("gov.pl: %w", err)
	}
	data, header, err := get(ctx, c.cfg.GovPL, file, maxFileBytes)
	if err != nil {
		return nil, fmt.Errorf("gov.pl spreadsheet: %w", err)
	}
	modified, err := http.ParseTime(header.Get("Last-Modified"))
	if err != nil {
		return nil, fmt.Errorf("gov.pl spreadsheet: no Last-Modified date: %w", err)
	}
	made, err := kernel.ParseDate(modified.UTC().Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("gov.pl spreadsheet date: %w", err)
	}
	return published(data, &pl.Attribution{
		Publisher: Publisher, Dataset: DatasetTitle, License: License, URL: file.String(),
		Made: made, Fetched: now.UTC(), Processing: PublishedProcessing,
	})
}

// fromDane reads the spreadsheet dane.gov.pl's dataset lists, made on the
// day its resource says.
func (c *MinistryClient) fromDane(ctx context.Context, now time.Time) (*pl.Published, error) {
	list, err := url.Parse(c.cfg.DaneBaseURL + daneResourcesPath)
	if err != nil {
		return nil, fmt.Errorf("dane.gov.pl resource list: %w", err)
	}
	body, _, err := get(ctx, c.cfg.Dane, list, maxListBytes)
	if err != nil {
		return nil, fmt.Errorf("dane.gov.pl resource list: %w", err)
	}
	file, err := discoverDane(bytes.NewReader(body), list)
	if err != nil {
		return nil, fmt.Errorf("dane.gov.pl: %w", err)
	}
	data, _, err := get(ctx, c.cfg.Dane, file.url, maxFileBytes)
	if err != nil {
		return nil, fmt.Errorf("dane.gov.pl spreadsheet: %w", err)
	}
	dataset := DatasetTitle
	if file.title != "" {
		dataset = file.title
	}
	return published(data, &pl.Attribution{
		Publisher: Publisher, Dataset: dataset, License: License, URL: file.url.String(),
		Made: file.made, Fetched: now.UTC(), Processing: PublishedProcessing,
	})
}

// published parses the spreadsheet data into the published data with att.
func published(data []byte, att *pl.Attribution) (*pl.Published, error) {
	rates, err := ParseWorkbook(data)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet %s: %w", att.URL, err)
	}
	if err := att.Validate(); err != nil {
		return nil, fmt.Errorf("spreadsheet %s: %w", att.URL, err)
	}
	return &pl.Published{Rates: rates, Attribution: *att}, nil
}

// get fetches u with hc and returns its body, at most limit bytes, and its
// headers. A status other than 200 fails with ErrUnexpectedStatus, a longer
// body with ErrTooLarge.
func get(ctx context.Context, hc *http.Client, u *url.URL, limit int64) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := hc.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", u.Redacted(), err)
	}
	body, err := read(resp, limit)
	if closeErr := resp.Body.Close(); err == nil && closeErr != nil {
		err = fmt.Errorf("close: %w", closeErr)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", u.Redacted(), err)
	}
	return body, resp.Header, nil
}

// read reads resp's body, at most limit bytes, if its status is 200.
func read(resp *http.Response, limit int64) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("reading: %w", err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%w: more than %d bytes", ErrTooLarge, limit)
	}
	return body, nil
}
