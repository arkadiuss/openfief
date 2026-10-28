package ingest

import "fmt"

// ProblemKind says why a source row was left out of the ledger.
type ProblemKind int

// The problem kinds. The zero kind is not a kind.
const (
	// ProblemUnsupported marks a row OpenFief cannot book yet, such as an
	// option trade or broker interest: its effect is missing from the ledger,
	// so balance assertions may fail until it is supported.
	ProblemUnsupported ProblemKind = iota + 1
	// ProblemSkipped marks a row left out on purpose because its effect is
	// already booked elsewhere or doesn't belong in the ledger, such as a
	// summary row or a transaction before go-live.
	ProblemSkipped
)

// String returns "unsupported" or "skipped", or a placeholder for an invalid
// kind.
func (k ProblemKind) String() string {
	switch k {
	case ProblemUnsupported:
		return "unsupported"
	case ProblemSkipped:
		return "skipped"
	}
	return fmt.Sprintf("ProblemKind(%d)", int(k))
}

// Problem is a source row a mapping left out of the ledger. Ref names the
// row by its place in the source and the institution's own ID, never by an
// account number or an amount (GR-12).
type Problem struct {
	Kind   ProblemKind
	Ref    string
	Reason string
}

// Unsupported returns the problem of a row OpenFief cannot book yet.
func Unsupported(ref, reason string) Problem {
	return Problem{Kind: ProblemUnsupported, Ref: ref, Reason: reason}
}

// Skipped returns the problem of a row left out on purpose.
func Skipped(ref, reason string) Problem {
	return Problem{Kind: ProblemSkipped, Ref: ref, Reason: reason}
}

// String returns "kind ref: reason".
func (p Problem) String() string { return p.Kind.String() + " " + p.Ref + ": " + p.Reason }
