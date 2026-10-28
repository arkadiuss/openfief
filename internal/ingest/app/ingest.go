// Package app holds the ingest context's use cases: running a connector's
// ingest path, from a file or a pull, through its mapper into the ledger
// (D-7, D-18, D-23). The ports it needs are declared here, by the consumer.
package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownInstitution reports an institution with no mapper.
var ErrUnknownInstitution = errors.New("no connector for institution")

// ErrSeedingOutsideOpening reports a mapping in transactions mode that
// holds opening lots or opening cash: only an opening batch seeds an
// account (D-23).
var ErrSeedingOutsideOpening = errors.New("seeding transaction outside an opening batch")

// The ingest paths a batch records as its source.
const (
	// SourceFile is a statement file, whether uploaded, passed to the CLI or
	// dropped in a watched folder.
	SourceFile = "file"
	// SourcePull is a statement the connector pulled from the institution.
	SourcePull = "pull"
)

// LedgerImporter imports mapped transactions and balance assertions into the
// ledger: idempotently, superseding changed versions, and checking each
// assertion against the derived balance (D-18).
type LedgerImporter interface {
	Import(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error)
	// Seed imports an opening batch like Import, in one database
	// transaction that refuses, with an error wrapping a
	// *ledger.AlreadySeededError, to seed an account that is already
	// seeded, unless the seeding changes nothing for it: then the summary
	// lists it in AlreadySeeded (D-23).
	Seed(ctx context.Context, txs []ledger.Transaction, assertions []ledger.BalanceAssertion) (ingest.ImportSummary, error)
}

// Mapper translates one institution's statement into the ledger in a mode.
// It fails only for a statement it cannot read; rows it cannot book are
// ingest problems in the mapping.
type Mapper interface {
	Map(src io.Reader, mode ingest.Mode) (ingest.Mapping, error)
}

// StatementFetcher pulls the current statement from an institution, such
// as a Flex Query from IBKR's Flex Web Service.
type StatementFetcher interface {
	FetchStatement(ctx context.Context) ([]byte, error)
}

// Service runs connectors' ingest paths. Data enters the ledger only through
// it (D-7).
type Service struct {
	Ledger LedgerImporter
	// Mappers holds each institution's mapper.
	Mappers map[ingest.InstitutionID]Mapper
	// Today returns the date a batch starts on.
	Today func() kernel.Date
	// NewBatchID returns a new, unique import batch ID.
	NewBatchID func() ingest.BatchID
}

// IngestFile reads one institution's statement file, maps it in mode,
// imports the result into the ledger and returns the completed batch, whose
// problems list the rows left out.
func (s *Service) IngestFile(ctx context.Context, institution ingest.InstitutionID, mode ingest.Mode, src io.Reader) (ingest.Batch, error) {
	return s.ingest(ctx, institution, SourceFile, mode, src)
}

// Pull fetches one institution's current statement with f and ingests it
// like IngestFile.
func (s *Service) Pull(ctx context.Context, institution ingest.InstitutionID, mode ingest.Mode, f StatementFetcher) (ingest.Batch, error) {
	if _, ok := s.Mappers[institution]; !ok {
		return ingest.Batch{}, fmt.Errorf("%w %s", ErrUnknownInstitution, institution)
	}
	statement, err := f.FetchStatement(ctx)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("pull %s: %w", institution, err)
	}
	return s.ingest(ctx, institution, SourcePull, mode, bytes.NewReader(statement))
}

// ingest runs one import batch.
func (s *Service) ingest(ctx context.Context, institution ingest.InstitutionID, source string, mode ingest.Mode, src io.Reader) (ingest.Batch, error) {
	mapper, ok := s.Mappers[institution]
	if !ok {
		return ingest.Batch{}, fmt.Errorf("%w %s", ErrUnknownInstitution, institution)
	}
	batch, err := ingest.NewBatch(s.NewBatchID(), institution, source, mode, s.Today())
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("ingest %s: %w", institution, err)
	}
	m, err := mapper.Map(src, mode)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("ingest %s batch %s: %w", institution, batch.ID, err)
	}
	imported, err := importMapping(ctx, s.Ledger, mode, &m)
	if err != nil {
		return ingest.Batch{}, fmt.Errorf("ingest %s batch %s: %w", institution, batch.ID, err)
	}
	batch.Complete(&m, imported)
	return batch, nil
}

// importMapping imports m into l: an opening batch through Seed, which
// refuses to seed an account twice, and a transactions batch through
// Import, after refusing any seeding transaction in it (D-23).
func importMapping(ctx context.Context, l LedgerImporter, mode ingest.Mode, m *ingest.Mapping) (ingest.ImportSummary, error) {
	if mode == ingest.ModeOpening {
		sum, err := l.Seed(ctx, m.Transactions, m.Assertions)
		if err != nil {
			return ingest.ImportSummary{}, fmt.Errorf("seeding: %w", err)
		}
		return sum, nil
	}
	for i := range m.Transactions {
		if t := &m.Transactions[i]; ledger.IsSeeding(t) {
			return ingest.ImportSummary{}, fmt.Errorf("%w: %s", ErrSeedingOutsideOpening, t.Key)
		}
	}
	sum, err := l.Import(ctx, m.Transactions, m.Assertions)
	if err != nil {
		return ingest.ImportSummary{}, fmt.Errorf("importing: %w", err)
	}
	return sum, nil
}
