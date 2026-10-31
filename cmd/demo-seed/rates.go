package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
	"github.com/arkadiuss/openfief/internal/valuation/nbp"
)

// committedRates are NBP table A rates for the demo's currencies, committed
// for a load without network access. They are public data.
//
//go:embed nbp-table-a.csv
var committedRates []byte

// errBadRates reports a malformed line of the committed rates.
var errBadRates = errors.New("malformed committed NBP rates")

// rateLookback is how far before the first day that needs a rate the rates
// start, so that day has one even after a long weekend or, for a lot's tax
// cost, on the business day before it (pl.TaxRateCutoff).
const rateLookback = 14

// nbpTimeout bounds each request to NBP.
const nbpTimeout = 30 * time.Second

// demoCurrencies are the currencies the demo's instruments are quoted in
// besides PLN.
func demoCurrencies() []string { return []string{"EUR", "USD"} }

// fxRates returns the NBP table A rates of the demo's currencies from
// before ds.FXFrom, the earliest day the dataset needs one for, to its
// today: fetched from NBP through its adapter (D-20), or, offline or when
// NBP fails, the committed ones. Either way they are NBP's own published
// rates.
func fxRates(ctx context.Context, cfg *config, ds *demo.Dataset, log *slog.Logger) ([]valuation.FXRate, error) {
	from := ds.FXFrom.AddDays(-rateLookback)
	if !cfg.Offline {
		client := nbp.New(&http.Client{Timeout: nbpTimeout}, cfg.NBPURL)
		rates, err := app.FetchRates(ctx, client, from, ds.Today, nbp.MaxRangeDays)
		if err == nil {
			log.Info("fetched FX rates from NBP", "rates", len(rates))
			return onlyDemoCurrencies(rates), nil
		}
		log.Warn("NBP unreachable: using the committed NBP rates", "error", err)
	}
	rates, err := parseRates(committedRates, from, ds.Today)
	if err != nil {
		return nil, err
	}
	log.Info("using the committed NBP rates", "rates", len(rates))
	return rates, nil
}

// onlyDemoCurrencies returns the rates of demoCurrencies among rates.
func onlyDemoCurrencies(rates []valuation.FXRate) []valuation.FXRate {
	var out []valuation.FXRate
	for _, r := range rates {
		for _, c := range demoCurrencies() {
			if r.Currency.String() == c {
				out = append(out, r)
			}
		}
	}
	return out
}

// parseRates reads "date,currency,rate" lines, skipping empty lines and
// "#" comments, and returns the rates dated in [from, to], with NBP's
// source name.
func parseRates(data []byte, from, to kernel.Date) ([]valuation.FXRate, error) {
	var out []valuation.FXRate
	sc := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		r, err := parseRate(line)
		if err != nil {
			return nil, fmt.Errorf("%w: line %d: %w", errBadRates, n, err)
		}
		if !r.Date.Before(from) && !r.Date.After(to) {
			out = append(out, r)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errBadRates, err)
	}
	return out, nil
}

// parseRate parses one "date,currency,rate" line.
func parseRate(line string) (valuation.FXRate, error) {
	const fields = 3
	parts := strings.Split(line, ",")
	if len(parts) != fields {
		return valuation.FXRate{}, fmt.Errorf("%w: want %d fields", errBadRates, fields)
	}
	d, err1 := kernel.ParseDate(parts[0])
	c, err2 := kernel.ParseCurrency(parts[1])
	rate, err3 := kernel.ParseDecimal(parts[2])
	if err := errors.Join(err1, err2, err3); err != nil {
		return valuation.FXRate{}, fmt.Errorf("%w: %w", errBadRates, err)
	}
	if rate.Sign() <= 0 {
		return valuation.FXRate{}, fmt.Errorf("%w: rate not positive", errBadRates)
	}
	return valuation.FXRate{Currency: c, Date: d, Rate: rate, Source: nbp.Source}, nil
}
