package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// indexTickerSuffix ends the EODHD ticker of every market index.
const indexTickerSuffix = ".INDX"

// indexName is a name that can stand for its own EODHD index ticker, such
// as "WIG20", the manifest's pattern without the suffix.
var indexName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// errBadRequest reports a sync request naming an instrument this connector
// can't read.
var errBadRequest = errors.New("invalid sync request")

// targets is what one sync prices: instruments, each by ISIN once, and
// market indices, each by name once, and the requested indices it can't
// fetch.
type targets struct {
	instruments []instrument
	indices     []index
	problems    []ingest.Problem
}

// targetsFor returns what a sync of req prices (D-93): the config's
// instruments to watch, then each instrument the core asked for that the
// config doesn't list, with the ledger's symbol it asked for added to the
// symbols an ISIN's closes are delivered under; and the config's indices,
// then each index the core asked for by name, fetched as the ticker of the
// same name (D-96), such as WIG20.INDX. A requested index whose name can't
// be a ticker is a problem. An instrument that isn't a valid ISIN and
// symbol is errBadRequest.
func (c *config) targetsFor(req *connectorv1.SyncRequest) (targets, error) {
	out := targets{instruments: slices.Clone(c.instruments), indices: slices.Clone(c.indices)}
	byISIN := make(map[ledger.ISIN]int, len(out.instruments))
	for i := range out.instruments {
		byISIN[out.instruments[i].isin] = i
	}
	for _, pb := range req.GetInstruments() {
		isin, err := ledger.ParseISIN(pb.GetIsin())
		if err != nil || pb.GetSymbol() == "" {
			return targets{}, fmt.Errorf("%w: instrument %q/%q", errBadRequest, pb.GetIsin(), pb.GetSymbol())
		}
		i, ok := byISIN[isin]
		if !ok {
			byISIN[isin] = len(out.instruments)
			out.instruments = append(out.instruments, instrument{isin: isin, symbols: []string{pb.GetSymbol()}})
			continue
		}
		in := &out.instruments[i]
		if !slices.Contains(in.symbols, pb.GetSymbol()) {
			in.symbols = append(slices.Clone(in.symbols), pb.GetSymbol())
		}
	}
	for _, name := range req.GetIndices() {
		if slices.ContainsFunc(out.indices, func(ix index) bool { return ix.name == name }) {
			continue
		}
		if !indexName.MatchString(name) {
			out.problems = append(out.problems, ingest.Unsupported(name, "no EODHD index ticker for this name; list it in the config with its ticker"))
			continue
		}
		out.indices = append(out.indices, index{name: name, ticker: name + indexTickerSuffix})
	}
	return out, nil
}
