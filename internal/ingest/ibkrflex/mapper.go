package ibkrflex

import (
	"errors"
	"fmt"
	"io"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrUnknownMode reports an ingest mode the mapper doesn't know.
var ErrUnknownMode = errors.New("unknown ingest mode")

// SourcePrefix starts the source of every transaction key this connector
// writes; the ledger account ID follows it, never the IBKR account number.
const SourcePrefix = "ibkr-flex:"

// Mapper translates Flex Query statements into the ledger. It knows which
// ledger account each IBKR account feeds.
type Mapper struct {
	// Accounts maps an IBKR account ID, such as U0000000, to the ledger
	// account it feeds. A statement for any other account is skipped.
	Accounts map[string]ledger.AccountID
	// Limits caps what a statement may contain; zero fields take defaults.
	Limits Limits
}

// problemError is a row-level failure that becomes an ingest problem.
type problemError struct {
	kind   ingest.ProblemKind
	reason string
}

// Error returns the reason.
func (e *problemError) Error() string { return e.reason }

// unsupported returns the error of a row OpenFief cannot book yet.
func unsupported(format string, args ...any) error {
	return &problemError{kind: ingest.ProblemUnsupported, reason: fmt.Sprintf(format, args...)}
}

// skipped returns the error of a row left out on purpose.
func skipped(format string, args ...any) error {
	return &problemError{kind: ingest.ProblemSkipped, reason: fmt.Sprintf(format, args...)}
}

// errLedger reports a row whose transaction the ledger rejects, such as a
// withholding tax above its dividend.
var errLedger = errors.New("not a valid ledger transaction")

// booked returns t with description desc, or err from building it, marked as
// a ledger rejection.
func booked(t *ledger.Transaction, err error, desc string) (ledger.Transaction, error) {
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("%w: %w", errLedger, err)
	}
	t.Description = desc
	return *t, nil
}

// problemOf turns a row's mapping error into an ingest problem. Ledger
// errors are reduced to their kind, since they can carry amounts.
func problemOf(ref string, err error) ingest.Problem {
	var pe *problemError
	switch {
	case errors.As(err, &pe):
		return ingest.Problem{Kind: pe.kind, Ref: ref, Reason: pe.reason}
	case errors.Is(err, errField):
		return ingest.Unsupported(ref, err.Error())
	}
	return ingest.Unsupported(ref, errLedger.Error())
}

// stmt is what every row of one FlexStatement is mapped with.
type stmt struct {
	index   int
	account ledger.AccountID
	mode    ingest.Mode
	// to is the statement period's last day: the date of its balance
	// assertions, and in opening mode the go-live date.
	to  kernel.Date
	out *ingest.Mapping
}

// header returns the transaction header for the institution's own ID.
func (s *stmt) header(externalID string, on kernel.Date) ledger.TransactionHeader {
	return ledger.TransactionHeader{
		Key:     ledger.TransactionKey{Source: SourcePrefix + string(s.account), ExternalID: externalID},
		Account: s.account,
		Date:    on,
	}
}

// ref names a row in problems: its statement, section, place and, when it
// has one, the institution's ID.
func (s *stmt) ref(section string, r *row, idAttr string) string {
	ref := fmt.Sprintf("FlexStatement#%d/%s#%d", s.index, section, r.index)
	if idAttr != "" && r.text(idAttr) != "" {
		ref += "[" + idAttr + "=" + r.text(idAttr) + "]"
	}
	return ref
}

// add appends t, or the problem err makes of the row, to the mapping.
func (s *stmt) add(ref string, t *ledger.Transaction, err error) {
	if err != nil {
		s.problem(ref, err)
		return
	}
	s.out.Transactions = append(s.out.Transactions, *t)
}

// problem records the row's error as an ingest problem.
func (s *stmt) problem(ref string, err error) {
	s.out.Problems = append(s.out.Problems, problemOf(ref, err))
}

// Map reads one Flex Query response and translates it in mode. It fails only
// when the input is not a readable Flex statement within the limits; rows it
// cannot book become ingest problems.
func (m Mapper) Map(src io.Reader, mode ingest.Mode) (ingest.Mapping, error) {
	if mode != ingest.ModeOpening && mode != ingest.ModeTransactions {
		return ingest.Mapping{}, fmt.Errorf("%w: %s", ErrUnknownMode, mode)
	}
	statements, err := decode(src, m.Limits)
	if err != nil {
		return ingest.Mapping{}, err
	}
	var out ingest.Mapping
	for i := range statements {
		if err := m.mapStatement(&statements[i], i+1, mode, &out); err != nil {
			return ingest.Mapping{}, err
		}
	}
	return out, nil
}

// mapStatement translates one FlexStatement into out.
func (m Mapper) mapStatement(st *statement, index int, mode ingest.Mode, out *ingest.Mapping) error {
	ref := fmt.Sprintf("FlexStatement#%d", index)
	account, ok := m.Accounts[st.attrs["accountId"]]
	if !ok {
		out.Problems = append(out.Problems, ingest.Skipped(ref, "statement of an IBKR account that feeds no ledger account"))
		return nil
	}
	to, err := (&row{attrs: st.attrs}).date("toDate")
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrMalformed, ref, err)
	}
	s := stmt{index: index, account: account, mode: mode, to: to, out: out}
	s.openPositions(st.sections["OpenPositions"])
	s.cashReport(st.sections["CashReport"])
	s.trades(st.sections["Trades"])
	s.cashTransactions(st.sections["CashTransactions"])
	return nil
}
