package ingest

import (
	"errors"
	"fmt"
	"slices"
	"unicode"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ErrInvalidBrokerAccount reports a broker account ID that is empty, too
// long or holds anything but printable characters without spaces.
var ErrInvalidBrokerAccount = errors.New("invalid broker account")

// maxBrokerAccountLen bounds a broker account ID.
const maxBrokerAccountLen = 64

// BrokerAccountID is an institution's own ID for one of the owner's
// accounts, such as an IBKR Flex account ID. A broker connector reports the
// broker accounts it saw (D-72), so the owner picks one instead of typing
// it. It is an account number: it is stored to map it, shown to the owner,
// and never logged.
type BrokerAccountID string

// Validate reports whether id is 1 to 64 printable characters without
// spaces.
func (id BrokerAccountID) Validate() error {
	s := string(id)
	if s == "" || len(s) > maxBrokerAccountLen {
		return fmt.Errorf("%w: must have 1 to %d characters", ErrInvalidBrokerAccount, maxBrokerAccountLen)
	}
	for _, c := range s {
		if !unicode.IsPrint(c) || unicode.IsSpace(c) {
			return fmt.Errorf("%w: only printable characters without spaces", ErrInvalidBrokerAccount)
		}
	}
	return nil
}

// BrokerAccount is a broker account a connector instance reported, and the
// OpenFief account it is mapped to, if any (D-72).
type BrokerAccount struct {
	Instance    string
	ID          BrokerAccountID
	Institution InstitutionID
	// Account is the OpenFief account the owner mapped it to; empty while
	// it is unmapped, and then its transactions are not booked.
	Account ledger.AccountID
	// LastSeen is the date of the last delivery that reported it, and
	// InLastDelivery whether the instance's last delivery that reported
	// broker accounts reported it.
	LastSeen       kernel.Date
	InLastDelivery bool
	// Seeded reports whether an opening batch of the instance completed
	// after it was mapped, so the next opening batch leaves its seeding out
	// (D-23).
	Seeded bool
}

// Mapped reports whether the owner mapped b to an OpenFief account.
func (b *BrokerAccount) Mapped() bool { return b.Account != "" }

// BrokerEntries are the transactions and balance assertions a connector
// delivered for one broker account it named, before the core knows the
// OpenFief account: their Account is empty, and their key's source names no
// account (D-72).
type BrokerEntries struct {
	Account      BrokerAccountID
	Transactions []ledger.Transaction
	Assertions   []ledger.BalanceAssertion
}

// Resolution is a mapping with its broker accounts resolved: what to
// import, the broker accounts left unmapped, and the mapped ones an opening
// batch seeds.
type Resolution struct {
	Mapping Mapping
	// Unmapped are the broker accounts the delivery named that are not
	// mapped to an OpenFief account; none of their entries is booked.
	Unmapped []BrokerAccountID
	// Seeding are the broker accounts mapped and not yet seeded when the
	// mapping was resolved in an opening batch: once it completes they are
	// seeded. Empty in a transactions batch.
	Seeding []BrokerAccountID
}

// ResolveBrokerAccounts books the entries m names by broker account into the
// OpenFief accounts known maps them to (D-72): each transaction and balance
// assertion gets the account, and each transaction's key source gets ":"
// and the account ID, so that keys stay per account as connectors key them
// (ledger.TransactionKey). Entries of a broker account that is not mapped
// are left out and the account listed in Unmapped. In an opening batch the
// seeding transactions of a broker account already seeded are left out too,
// so seeding one newly mapped account never seeds another again; their
// balance assertions are kept. The rest of m is kept as it is.
func ResolveBrokerAccounts(m *Mapping, known []BrokerAccount, mode Mode) Resolution {
	out := Resolution{Mapping: Mapping{
		Transactions:     slices.Clone(m.Transactions),
		Assertions:       slices.Clone(m.Assertions),
		Problems:         slices.Clone(m.Problems),
		CrossCheckPrices: slices.Clone(m.CrossCheckPrices),
	}}
	byID := make(map[BrokerAccountID]BrokerAccount, len(known))
	for i := range known {
		b := &known[i]
		byID[b.ID] = *b
		if mode == ModeOpening && b.Mapped() && !b.Seeded {
			out.Seeding = append(out.Seeding, b.ID)
		}
	}
	for i := range m.ByBrokerAccount {
		e := &m.ByBrokerAccount[i]
		b, ok := byID[e.Account]
		if !ok || !b.Mapped() {
			if !slices.Contains(out.Unmapped, e.Account) {
				out.Unmapped = append(out.Unmapped, e.Account)
			}
			continue
		}
		book(e, b.Account, mode == ModeOpening && b.Seeded, &out.Mapping)
	}
	slices.Sort(out.Unmapped)
	return out
}

// book adds e's entries to out in account, each transaction's key source
// with ":" and the account appended, and with reseeding without its
// seeding transactions.
func book(e *BrokerEntries, account ledger.AccountID, reseeding bool, out *Mapping) {
	for j := range e.Transactions {
		if reseeding && ledger.IsSeeding(&e.Transactions[j]) {
			continue
		}
		t := e.Transactions[j]
		t.Account = account
		t.Key.Source += ":" + string(account)
		out.Transactions = append(out.Transactions, t)
	}
	for _, a := range e.Assertions {
		a.Account = account
		out.Assertions = append(out.Assertions, a)
	}
}

// ReportedBrokerAccounts returns every broker account m reports, in the
// order it first names them: those it lists as seen, then any other that
// names entries.
func ReportedBrokerAccounts(m *Mapping) []BrokerAccountID {
	out := slices.Clone(m.BrokerAccounts)
	for i := range m.ByBrokerAccount {
		if id := m.ByBrokerAccount[i].Account; !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out
}
