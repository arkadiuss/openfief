package ingest

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// Binding is a connector binding: connector instance Instance may write
// to account Account.
type Binding struct {
	Instance string
	Account  ledger.AccountID
}

// ErrAccountNotBound reports a delivery that names an account its connector
// instance is not bound to.
var ErrAccountNotBound = errors.New("account not bound to the connector instance")

// UnboundAccountsError reports the accounts a connector instance's delivery
// names without a connector binding to them. The whole delivery is refused.
type UnboundAccountsError struct {
	Instance string
	Accounts []ledger.AccountID
}

// Error names the instance and the accounts.
func (e *UnboundAccountsError) Error() string {
	ids := make([]string, len(e.Accounts))
	for i, a := range e.Accounts {
		ids[i] = string(a)
	}
	return fmt.Sprintf("connector instance %s is not bound to account %s", e.Instance, strings.Join(ids, ", "))
}

// Unwrap returns ErrAccountNotBound.
func (e *UnboundAccountsError) Unwrap() error { return ErrAccountNotBound }

// CheckBound reports whether every transaction and balance assertion of m
// is in an account of bound, the accounts the connector instance has a
// connector binding to. Otherwise it returns an *UnboundAccountsError
// listing each other account once, in ID order.
func CheckBound(instance string, bound []ledger.AccountID, m *Mapping) error {
	var unbound []ledger.AccountID
	check := func(a ledger.AccountID) {
		if !slices.Contains(bound, a) && !slices.Contains(unbound, a) {
			unbound = append(unbound, a)
		}
	}
	for i := range m.Transactions {
		check(m.Transactions[i].Account)
	}
	for i := range m.Assertions {
		check(m.Assertions[i].Account)
	}
	if len(unbound) == 0 {
		return nil
	}
	slices.Sort(unbound)
	return &UnboundAccountsError{Instance: instance, Accounts: unbound}
}

// ErrInvalidInstanceName reports a connector instance name that is not a
// plugin's name.
var ErrInvalidInstanceName = errors.New("invalid connector instance name")

// maxInstanceNameLen is the longest name of a plugin instance (D-48), so
// that its Service name, "plugin-<name>", fits in a DNS label.
const maxInstanceNameLen = 56

// ValidateInstanceName reports whether name can name a connector instance:
// a lower-case DNS label of at most 56 characters that starts with a
// letter, as a Plugin's name must be.
func ValidateInstanceName(name string) error {
	if len(name) > maxInstanceNameLen || !instanceName.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrInvalidInstanceName, name)
	}
	return nil
}

// instanceName matches a lower-case DNS label that starts with a letter.
var instanceName = regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)
