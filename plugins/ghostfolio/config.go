package main

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
)

// Errors reading the config.
var (
	// errBadConfig reports a config the manifest's schema would refuse, or
	// one naming an ISIN twice; it wraps the reason.
	errBadConfig = errors.New("invalid config")
	errNotHTTPS  = errors.New("not an https URL with a host and nothing else")
	errInterval  = errors.New("shorter than the shortest interval")
	errTwice     = errors.New("listed twice")
)

// The sync interval's default and its floor.
const (
	defaultInterval = time.Hour
	minInterval     = 5 * time.Minute
)

// rawConfig is the plugin's configuration as OPENFIEF_CONFIG holds it.
type rawConfig struct {
	URL      string       `json:"url"`
	Interval string       `json:"interval"`
	Listings []rawListing `json:"listings"`
}

// rawListing says how Ghostfolio knows one listed instrument.
type rawListing struct {
	ISIN       string `json:"isin"`
	DataSource string `json:"dataSource"`
	Symbol     string `json:"symbol"`
}

// config is the checked configuration.
type config struct {
	// url is the owner's Ghostfolio, such as https://ghostfolio.example.com;
	// empty when it runs in the cluster and the plugin reaches it as its
	// in-cluster service (D-74).
	url string
	// interval is how long the plugin waits between syncs.
	interval time.Duration
	listings map[ledger.ISIN]ghostfolio.Listing
}

// parse checks the raw config: an https URL of the Ghostfolio's origin, or
// none for a Ghostfolio in the cluster (D-74), an interval of at least five
// minutes (an hour when empty), and listings with valid ISINs, none twice.
func (r *rawConfig) parse() (config, error) {
	c := config{interval: defaultInterval}
	if r.URL != "" {
		origin, ok := httpsOrigin(r.URL)
		if !ok {
			return config{}, fmt.Errorf("%w: url: %w", errBadConfig, errNotHTTPS)
		}
		c.url = origin
	}
	var err error
	if r.Interval != "" {
		if c.interval, err = time.ParseDuration(r.Interval); err != nil {
			return config{}, fmt.Errorf("%w: interval: %w", errBadConfig, err)
		}
		if c.interval < minInterval {
			return config{}, fmt.Errorf("%w: interval %s: %w", errBadConfig, c.interval, errInterval)
		}
	}
	if c.listings, err = parseListings(r.Listings); err != nil {
		return config{}, err
	}
	return c, nil
}

// httpsOrigin returns raw as an https origin, such as
// https://ghostfolio.example.com, and whether it is one: a host, no user,
// no path but "/", no query.
func httpsOrigin(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" {
		return "", false
	}
	if u.Path != "" && u.Path != "/" {
		return "", false
	}
	return "https://" + u.Host, true
}

// parseListings checks each listing's ISIN and that none is listed twice.
func parseListings(raw []rawListing) (map[ledger.ISIN]ghostfolio.Listing, error) {
	out := map[ledger.ISIN]ghostfolio.Listing{}
	for i, l := range raw {
		isin, err := ledger.ParseISIN(l.ISIN)
		if err != nil {
			return nil, fmt.Errorf("%w: listings[%d]: %w", errBadConfig, i, err)
		}
		if _, ok := out[isin]; ok {
			return nil, fmt.Errorf("%w: listings[%d]: %s %w", errBadConfig, i, isin, errTwice)
		}
		out[isin] = ghostfolio.Listing{DataSource: l.DataSource, Symbol: l.Symbol}
	}
	return out, nil
}
