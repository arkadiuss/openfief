package demo

import (
	"errors"
	"fmt"
)

// ErrNotDemoDatabase reports a database that holds data but no demo marker:
// it may be a real portfolio, so the demo loader must not touch it.
var ErrNotDemoDatabase = errors.New("refusing to load the demo dataset: the database holds data and is not marked as a demo database")

// DatabaseState is what the demo loader finds in a database before it
// writes anything.
type DatabaseState struct {
	// Marked reports the demo marker: a row demo = true in the demo schema,
	// which only the demo loader writes.
	Marked bool
	// Rows counts the rows of ledger and valuation data: sleeves, accounts,
	// transactions and valuation inputs.
	Rows int64
	// Loader names the build of the loader that loaded a marked database,
	// as the marker records it; empty when the marker doesn't say.
	Loader string
}

// LoadAction is what the demo loader does with a database.
type LoadAction int

// The load actions. The zero action is not an action.
const (
	// LoadFresh marks an empty database as a demo database and loads the
	// dataset into it.
	LoadFresh LoadAction = iota + 1
	// LoadAgain loads the dataset into a demo database as it is: every
	// import is idempotent, so loading the same day's dataset again changes
	// nothing, and the ledger refuses one that would re-seed an account.
	LoadAgain
	// ResetAndLoad empties a demo database and loads the dataset.
	ResetAndLoad
)

// String returns the action's name, such as "load-fresh".
func (a LoadAction) String() string {
	switch a {
	case LoadFresh:
		return "load-fresh"
	case LoadAgain:
		return "load-again"
	case ResetAndLoad:
		return "reset-and-load"
	}
	return fmt.Sprintf("LoadAction(%d)", int(a))
}

// DecideLoad decides what the demo loader, of build loader, may do with a
// database in state st, given whether the operator asked to reset the
// demo. It is the guard that keeps the loader away from real data: a
// database with ledger or valuation data and no demo marker is refused,
// with or without reset. Only a marked demo database is ever reset: when
// asked to, or when another build of the loader filled it, so a deploy of
// changed code reseeds the demo rather than migrating its data (D-161). A
// loader that doesn't know its build (empty) never reseeds on its own. An
// empty database is loaded fresh.
func DecideLoad(st DatabaseState, reset bool, loader string) (LoadAction, error) {
	switch {
	case st.Marked && (reset || (loader != "" && st.Loader != loader)):
		return ResetAndLoad, nil
	case st.Marked:
		return LoadAgain, nil
	case st.Rows == 0:
		return LoadFresh, nil
	}
	return 0, fmt.Errorf("%w (%d rows of ledger and valuation data)", ErrNotDemoDatabase, st.Rows)
}
