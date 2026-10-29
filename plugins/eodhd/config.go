package main

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
)

// Errors reading the config.
var (
	// errBadConfig reports a config the manifest's schema would refuse, or
	// one naming an instrument or index twice; it wraps the reason.
	errBadConfig           = errors.New("invalid config")
	errUnsupportedExchange = errors.New("not a supported exchange")
	errListedTwice         = errors.New("instrument listed twice")
)

// rawConfig is the plugin's configuration as OPENFIEF_CONFIG holds it.
type rawConfig struct {
	Start       string          `json:"start"`
	Instruments []rawInstrument `json:"instruments"`
	Indices     []rawIndex      `json:"indices"`
}

type rawInstrument struct {
	ISIN     string `json:"isin"`
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

type rawIndex struct {
	Name   string `json:"name"`
	Ticker string `json:"ticker"`
}

// config is the checked configuration.
type config struct {
	// start is the first day of a first sync; zero is EODHD's whole history.
	start       kernel.Date
	instruments []instrument
	indices     []index
}

// instrument is one instrument to fetch closes for: its ISIN, the EODHD
// exchange code of its listing when the owner named one, and the venue
// symbols in the ledger its closes are delivered under; none means EODHD's
// ticker code.
type instrument struct {
	isin     ledger.ISIN
	exchange string
	symbols  []string
}

// index is one market index to fetch levels for: its name in OpenFief and
// its EODHD ticker, such as "FAKE20.INDX".
type index struct {
	name   string
	ticker string
}

// parse checks the raw config: a date start, valid ISINs on supported
// exchanges, named indices with tickers, and no ISIN or index name twice.
func (r *rawConfig) parse() (config, error) {
	var c config
	if r.Start != "" {
		d, err := kernel.ParseDate(r.Start)
		if err != nil {
			return config{}, fmt.Errorf("%w: start: %w", errBadConfig, err)
		}
		c.start = d
	}
	seen := map[ledger.ISIN]bool{}
	for i, in := range r.Instruments {
		parsed, err := in.parse(seen)
		if err != nil {
			return config{}, fmt.Errorf("%w: instruments[%d]: %w", errBadConfig, i, err)
		}
		c.instruments = append(c.instruments, parsed)
	}
	names := map[string]bool{}
	for i, ix := range r.Indices {
		if ix.Name == "" || ix.Ticker == "" || names[ix.Name] {
			return config{}, fmt.Errorf("%w: indices[%d]: needs a unique name and a ticker", errBadConfig, i)
		}
		names[ix.Name] = true
		c.indices = append(c.indices, index{name: ix.Name, ticker: ix.Ticker})
	}
	return c, nil
}

// parse checks one instrument: a valid ISIN not in seen, which it adds,
// and a supported exchange when one is named.
func (r rawInstrument) parse(seen map[ledger.ISIN]bool) (instrument, error) {
	isin, err := ledger.ParseISIN(r.ISIN)
	if err != nil {
		return instrument{}, fmt.Errorf("isin: %w", err)
	}
	if r.Exchange != "" && !eodhd.SupportedExchange(r.Exchange) {
		return instrument{}, fmt.Errorf("%w: %q", errUnsupportedExchange, r.Exchange)
	}
	if seen[isin] {
		return instrument{}, fmt.Errorf("%w: %s", errListedTwice, isin)
	}
	seen[isin] = true
	in := instrument{isin: isin, exchange: r.Exchange}
	if r.Symbol != "" {
		in.symbols = []string{r.Symbol}
	}
	return in, nil
}
