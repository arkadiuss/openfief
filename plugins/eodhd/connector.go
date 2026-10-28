package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/delivery"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation/eodhd"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// institution names EODHD on every delivery.
const institution = "eodhd"

// The API key's file in the eodhd-credentials Secret, as the manifest
// mounts it at keyMount.
const (
	keyMount     = "/etc/openfief/eodhd"
	apiTokenFile = "api-token"
)

// errNoAPIKey reports an API key file that is missing or empty.
var errNoAPIKey = errors.New("no EODHD API key")

// marketData is what a sync needs from EODHD: an ISIN's listings and a
// ticker's closes.
type marketData interface {
	Listings(ctx context.Context, isin ledger.ISIN) ([]eodhd.Listing, error)
	Closes(ctx context.Context, ticker string, from, to kernel.Date) ([]eodhd.Close, error)
}

// connector serves EODHD end-of-day closes and index levels (D-22, D-51)
// through the connector contract. Its watermark holds the date of the
// latest close it delivered for each instrument and index, so each sync
// asks only for the days after it, up to today, and an instrument added to
// the config later starts from the configured start.
type connector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	cfg     config
	secrets string
	source  func(token string) marketData
	today   func() kernel.Date
}

// Sync returns each configured instrument's closes and each index's levels
// since the watermark. An instrument EODHD doesn't list on a supported
// exchange, or lists ambiguously, and an index ticker it doesn't know are
// ingest problems in the delivery, and the rest is still delivered. A
// missing API key is FailedPrecondition; EODHD failing otherwise is
// Unavailable, so the core retries the batch.
func (c *connector) Sync(ctx context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	marks, err := parseWatermark(req.Msg.GetWatermark())
	if err != nil {
		return nil, status(connect.CodeInvalidArgument, err)
	}
	token, err := readSecretFile(c.secrets, apiTokenFile)
	if err != nil {
		return nil, status(connect.CodeFailedPrecondition, err)
	}
	s := syncRun{src: c.source(token), marks: marks, start: c.cfg.start, today: c.today()}
	for _, in := range c.cfg.instruments {
		if fetchErr := s.instrument(ctx, in); fetchErr != nil {
			return nil, status(connect.CodeUnavailable, fmt.Errorf("eodhd: %w", fetchErr))
		}
	}
	for _, ix := range c.cfg.indices {
		if fetchErr := s.index(ctx, ix); fetchErr != nil {
			return nil, status(connect.CodeUnavailable, fmt.Errorf("eodhd: %w", fetchErr))
		}
	}
	out, err := delivery.FromMapping(req.Msg.GetBatchId(), institution, &ingest.Mapping{Problems: s.problems})
	if err != nil {
		return nil, status(connect.CodeInternal, fmt.Errorf("eodhd: %w", err))
	}
	out.Prices, out.IndexLevels = s.prices, s.levels
	return connect.NewResponse(&connectorv1.SyncResponse{Delivery: out, Watermark: s.marks.String()}), nil
}

// status is the Connect error with code for err.
func status(code connect.Code, err error) *connect.Error { return connect.NewError(code, err) }

// syncRun collects one sync's delivery and moves its watermark.
type syncRun struct {
	src      marketData
	marks    watermark
	start    kernel.Date
	today    kernel.Date
	prices   []*connectorv1.Price
	levels   []*connectorv1.IndexLevel
	problems []ingest.Problem
}

// instrument adds in's closes since its mark: it finds the listing by ISIN,
// fetches its closes and quotes them in the listing's currency.
func (s *syncRun) instrument(ctx context.Context, in instrument) error {
	key := in.isin.String()
	from, due := nextFrom(s.marks.Prices, key, s.start, s.today)
	if !due {
		return nil
	}
	listings, err := s.src.Listings(ctx, in.isin)
	if err != nil {
		return fmt.Errorf("listings: %w", err)
	}
	listing, err := eodhd.Pick(listings, in.isin, in.exchange)
	if err != nil {
		s.problems = append(s.problems, ingest.Unsupported(key, err.Error()))
		return nil
	}
	closes, err := s.closes(ctx, key, listing.Ticker(), from)
	if err != nil || len(closes) == 0 {
		return err
	}
	symbol := in.symbol
	if symbol == "" {
		symbol = listing.Code
	}
	held := ledger.Instrument{ISIN: in.isin, Symbol: symbol}
	prices := make([]*connectorv1.Price, 0, len(closes))
	for _, cl := range closes {
		m, err := listing.Quote(cl.Close)
		if err != nil {
			s.problems = append(s.problems, ingest.Unsupported(key, err.Error()))
			return nil
		}
		prices = append(prices, delivery.Price(held, cl.Date, m))
	}
	s.prices = append(s.prices, prices...)
	s.marks.Prices[key] = closes[len(closes)-1].Date.String()
	return nil
}

// index adds ix's levels since its mark.
func (s *syncRun) index(ctx context.Context, ix index) error {
	from, due := nextFrom(s.marks.Indices, ix.name, s.start, s.today)
	if !due {
		return nil
	}
	closes, err := s.closes(ctx, ix.name, ix.ticker, from)
	if err != nil || len(closes) == 0 {
		return err
	}
	for _, cl := range closes {
		s.levels = append(s.levels, delivery.IndexLevel(ix.name, cl.Date, cl.Close))
	}
	s.marks.Indices[ix.name] = closes[len(closes)-1].Date.String()
	return nil
}

// closes fetches ticker's closes from the date from to today, in date
// order. A ticker EODHD doesn't know is an ingest problem of ref and no
// closes.
func (s *syncRun) closes(ctx context.Context, ref, ticker string, from kernel.Date) ([]eodhd.Close, error) {
	closes, err := s.src.Closes(ctx, ticker, from, s.today)
	if errors.Is(err, eodhd.ErrNotFound) {
		s.problems = append(s.problems, ingest.Unsupported(ref, "EODHD has no ticker "+ticker))
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("closes: %w", err)
	}
	return closes, nil
}

// readSecretFile reads one file of the Secret in dir, through an os.Root so
// the read can't leave dir. An error names the file, never its content.
func readSecretFile(dir, name string) (string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", fmt.Errorf("%w: %s", errNoAPIKey, dir)
	}
	b, readErr := root.ReadFile(name)
	closeErr := root.Close()
	v := strings.TrimSpace(string(b))
	if readErr != nil || closeErr != nil || v == "" {
		return "", fmt.Errorf("%w: %s", errNoAPIKey, filepath.Join(dir, name))
	}
	return v, nil
}
