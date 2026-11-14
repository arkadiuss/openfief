package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// The session timeouts of the core's database connections (D-211). They are
// set when each connection starts, so they hold for everything the core runs
// on its pool, whichever database it is: the bundled one, the demo's, or the
// owner's own. A transaction a change leaves open and idle, or a statement that
// waits on a lock for ever, would otherwise hold the database up for every
// other change.
const (
	// defaultDBStatementTimeout is how long any one statement may run: far
	// above what a statement of the core takes, and the bound on one that
	// waits on a lock or a failing disk. A connector's delivery, whose
	// statements are many and larger, sets its own for its transaction
	// (--sync-delivery-timeout).
	defaultDBStatementTimeout = 2 * time.Minute
	// defaultDBIdleInTransactionTimeout is how long a transaction may sit idle
	// between two statements before the database ends it and rolls it back: a
	// change whose process hangs holds no locks beyond it.
	defaultDBIdleInTransactionTimeout = 2 * time.Minute
	// dbStatementTimeoutEnv and dbIdleInTransactionTimeoutEnv name the
	// variables that set them, as Go durations, such as "90s"; "0" turns the
	// timeout off. A parameter the connection string itself names
	// (statement_timeout=...) wins over both.
	dbStatementTimeoutEnv         = "OPENFIEF_DB_STATEMENT_TIMEOUT"
	dbIdleInTransactionTimeoutEnv = "OPENFIEF_DB_IDLE_IN_TRANSACTION_TIMEOUT"
)

// errBadDBTimeout reports a database timeout that is not a duration of zero or
// more.
var errBadDBTimeout = errors.New("a database timeout must be a duration of zero or more, such as 90s; 0 turns it off")

// dbTimeouts are the session timeouts of the core's database connections; zero
// turns one off.
type dbTimeouts struct {
	Statement         time.Duration
	IdleInTransaction time.Duration
}

// dbTimeoutsFromEnv reads the timeouts from the environment through getenv,
// the defaults for those it does not set.
func dbTimeoutsFromEnv(getenv func(string) string) (dbTimeouts, error) {
	t := dbTimeouts{Statement: defaultDBStatementTimeout, IdleInTransaction: defaultDBIdleInTransactionTimeout}
	for _, p := range []struct {
		env string
		out *time.Duration
	}{{dbStatementTimeoutEnv, &t.Statement}, {dbIdleInTransactionTimeoutEnv, &t.IdleInTransaction}} {
		v := strings.TrimSpace(getenv(p.env))
		if v == "" {
			continue
		}
		d, err := parseDBTimeout(v)
		if err != nil {
			return dbTimeouts{}, fmt.Errorf("%s: %w", p.env, err)
		}
		*p.out = d
	}
	return t, nil
}

// parseDBTimeout reads a Go duration of zero or more; a bare 0 is none.
func parseDBTimeout(v string) (time.Duration, error) {
	if v == "0" {
		return 0, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil || d < 0 {
		return 0, fmt.Errorf("%w: %q", errBadDBTimeout, v)
	}
	return d, nil
}

// apply sets the timeouts in params, the runtime parameters a connection
// starts with, leaving alone any the connection string already names.
func (t dbTimeouts) apply(params map[string]string) {
	for name, d := range map[string]time.Duration{
		"statement_timeout":                   t.Statement,
		"idle_in_transaction_session_timeout": t.IdleInTransaction,
	} {
		if _, named := params[name]; !named {
			params[name] = strconv.FormatInt(d.Milliseconds(), 10)
		}
	}
}
