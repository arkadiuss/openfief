package ingest

import (
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidInstitution reports an institution without an ID or a name.
var ErrInvalidInstitution = errors.New("invalid institution")

// ErrInvalidBatch reports an import batch without an ID, an institution, a
// source, a known mode or a start date.
var ErrInvalidBatch = errors.New("invalid import batch")

// InstitutionID identifies an institution, such as "ibkr".
type InstitutionID string

// Institution is a broker, bank or registry that holds assets: IBKR, XTB,
// the treasury bond registry.
type Institution struct {
	ID   InstitutionID
	Name string
}

// Validate reports whether in has an ID and a name.
func (in Institution) Validate() error {
	switch {
	case in.ID == "":
		return fmt.Errorf("%w: empty ID", ErrInvalidInstitution)
	case in.Name == "":
		return fmt.Errorf("%w %s: empty name", ErrInvalidInstitution, in.ID)
	}
	return nil
}

// Mode says what an import batch takes from its source (D-23).
type Mode int

// The ingest modes. The zero Mode is not a mode.
const (
	// ModeOpening seeds an account at go-live: opening lots from the
	// institution's current lot-level positions and the opening cash, plus
	// balance assertions. Transactions in the source are skipped, because the
	// opening lots already hold their effect.
	ModeOpening Mode = iota + 1
	// ModeTransactions imports transactions from go-live on, plus balance
	// assertions.
	ModeTransactions
)

// String returns "opening" or "transactions", or a placeholder for an
// invalid mode.
func (m Mode) String() string {
	switch m {
	case ModeOpening:
		return "opening"
	case ModeTransactions:
		return "transactions"
	}
	return fmt.Sprintf("Mode(%d)", int(m))
}

// BatchID identifies an import batch.
type BatchID string

// ImportSummary is what the ledger did with a batch's transactions and
// balance assertions (D-18).
type ImportSummary struct {
	// New counts transactions whose key was new.
	New int
	// NewVersions counts transactions that superseded a changed current
	// version.
	NewVersions int
	// Unchanged counts re-imported transactions equal to their current
	// version.
	Unchanged int
	// FailedAssertions counts balance assertions the derived balance did not
	// match: each is a flagged error.
	FailedAssertions int
	// AlreadySeeded are the accounts an opening batch found already seeded
	// with the same content: for them the batch changed nothing (D-23).
	AlreadySeeded []ledger.AccountID
}

// Batch is one run of a connector's ingest path: what was read from where,
// when, and what came of it.
type Batch struct {
	ID          BatchID
	Institution InstitutionID
	// Source names the ingest path that ran, such as "sync" or
	// "upload" (the app package's Source constants).
	Source  string
	Mode    Mode
	Started kernel.Date
	// Transactions and Assertions count what the mapping produced.
	Transactions int
	Assertions   int
	Imported     ImportSummary
	// Problems are the source rows the mapping left out, with their reasons.
	Problems []Problem
}

// NewBatch starts an import batch, before anything is read.
func NewBatch(id BatchID, institution InstitutionID, source string, mode Mode, started kernel.Date) (Batch, error) {
	switch {
	case id == "":
		return Batch{}, fmt.Errorf("%w: empty ID", ErrInvalidBatch)
	case institution == "":
		return Batch{}, fmt.Errorf("%w %s: no institution", ErrInvalidBatch, id)
	case source == "":
		return Batch{}, fmt.Errorf("%w %s: no source", ErrInvalidBatch, id)
	case mode != ModeOpening && mode != ModeTransactions:
		return Batch{}, fmt.Errorf("%w %s: %s", ErrInvalidBatch, id, mode)
	case started.IsZero():
		return Batch{}, fmt.Errorf("%w %s: no start date", ErrInvalidBatch, id)
	}
	return Batch{ID: id, Institution: institution, Source: source, Mode: mode, Started: started}, nil
}

// Complete records in b the results of its mapping m and of importing it
// into the ledger.
func (b *Batch) Complete(m *Mapping, imported ImportSummary) {
	b.Transactions = len(m.Transactions)
	b.Assertions = len(m.Assertions)
	b.Imported = imported
	b.Problems = append([]Problem(nil), m.Problems...)
}

// Mapping is what a connector makes of one source: ledger transactions,
// balance assertions, the rows it left out as ingest problems, and the
// institution's own prices as cross-check prices, which the ledger never
// sees.
type Mapping struct {
	Transactions     []ledger.Transaction
	Assertions       []ledger.BalanceAssertion
	Problems         []Problem
	CrossCheckPrices []CrossCheckPrice
}
