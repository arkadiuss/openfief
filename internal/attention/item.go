package attention

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Errors of attention items.
var (
	// ErrInvalidCondition reports a condition without a valid kind or a
	// subject, or with a subject or detail that is too long.
	ErrInvalidCondition = errors.New("invalid attention condition")
	// ErrInvalidTransition reports an owner's action an item's state does
	// not allow, such as acknowledging a resolved item.
	ErrInvalidTransition = errors.New("invalid attention item transition")
	// ErrNotOwnerResolvable reports the owner resolving an item of a kind
	// that only clears when its condition does (OwnerResolves).
	ErrNotOwnerResolvable = errors.New("this kind of attention item resolves when its check clears")
)

// Limits of a condition's text.
const (
	// MaxSubjectLen caps a subject, in bytes.
	MaxSubjectLen = 300
	// MaxDetailLen caps a detail, in bytes.
	MaxDetailLen = 500
)

// kindPattern is what a kind looks like: lower-case words joined by
// hyphens, such as "price-check-flagged".
var kindPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Kind says which check raised an item. Packages that raise items define
// their own kinds; the ones below are the checks the core has.
type Kind string

// The core's kinds of attention items (D-130, D-131).
const (
	// KindAssertionFlagged is a balance assertion that doesn't match the
	// balance derived from the log (D-18). Subject: the assertion.
	KindAssertionFlagged Kind = "balance-assertion-flagged"
	// KindPriceCheckFlagged is an instrument whose latest price check is
	// flagged and not acknowledged (D-110..D-112). Subject: the instrument.
	KindPriceCheckFlagged Kind = "price-check-flagged"
	// KindImportProblems is an import batch that left rows out because
	// OpenFief can't book them yet. Subject: the batch.
	KindImportProblems Kind = "import-problems"
	// KindBrokerAccountUnmapped is a broker account a connector reported
	// that no OpenFief account is mapped to, so its transactions are not
	// booked (D-72). Subject: the instance and the broker account.
	KindBrokerAccountUnmapped Kind = "broker-account-unmapped"
	// KindTickerUnconfirmed is a ticker whose ISIN the owner hasn't
	// confirmed, so its rows wait (D-116). Subject: the instance and the
	// ticker.
	KindTickerUnconfirmed Kind = "ticker-unconfirmed"
	// KindSettingMissing is an owner's setting a value needs and that is
	// not set, such as the marginal income tax rate for IKZE (D-78).
	// Subject: the setting.
	KindSettingMissing Kind = "setting-missing"
	// KindTransferBroken is a confirmed transfer whose legs no longer
	// match after a correction (D-131). Subject: the pair.
	KindTransferBroken Kind = "transfer-broken"
)

// OwnerResolves reports whether the owner may resolve an item of kind k by
// hand: a flagged balance assertion the owner has explained, or an import
// batch's left-out rows. Every other kind resolves when its check clears,
// such as when the owner maps the broker account or sets the setting.
func OwnerResolves(k Kind) bool { return k == KindAssertionFlagged || k == KindImportProblems }

// ReadSubject returns the subject of an item of kind k as read access may
// see it (D-137): an unmapped broker account's item names only the
// connector instance, since read access never includes a broker account
// number (D-121); every other subject is returned as it is. The owner
// sees the broker account through the account service.
func ReadSubject(k Kind, subject string) string {
	if k == KindBrokerAccountUnmapped {
		instance, _, _ := strings.Cut(subject, "/")
		return instance
	}
	return subject
}

// Validate reports whether k is a kind: lower-case words joined by hyphens.
func (k Kind) Validate() error {
	if !kindPattern.MatchString(string(k)) {
		return fmt.Errorf("%w: kind %q", ErrInvalidCondition, k)
	}
	return nil
}

// Condition is what a check found that needs the owner: its kind, its
// subject, unique within the kind, and a short detail, such as the
// difference found. A changed detail of a standing condition opens its item
// again (Observe).
type Condition struct {
	Kind    Kind
	Subject string
	Detail  string
}

// Validate reports whether c has a valid kind, a subject of at most
// MaxSubjectLen bytes and a detail of at most MaxDetailLen.
func (c *Condition) Validate() error {
	if err := c.Kind.Validate(); err != nil {
		return err
	}
	switch {
	case c.Subject == "", len(c.Subject) > MaxSubjectLen:
		return fmt.Errorf("%w: %s: subject must have 1 to %d bytes", ErrInvalidCondition, c.Kind, MaxSubjectLen)
	case len(c.Detail) > MaxDetailLen:
		return fmt.Errorf("%w: %s: detail must have at most %d bytes", ErrInvalidCondition, c.Kind, MaxDetailLen)
	}
	return nil
}

// State is where an item stands.
type State int

// The states. The zero state is not a state.
const (
	// StateOpen means the item needs the owner.
	StateOpen State = iota + 1
	// StateAcknowledged means the owner has seen it and its condition
	// still stands.
	StateAcknowledged
	// StateResolved means its condition cleared, or the owner resolved it.
	StateResolved
)

// String returns "open", "acknowledged" or "resolved", or a placeholder for
// an invalid state.
func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateAcknowledged:
		return "acknowledged"
	case StateResolved:
		return "resolved"
	}
	return fmt.Sprintf("State(%d)", int(s))
}

// Reason says why an item changed.
type Reason int

// The reasons. The zero reason is not a reason.
const (
	// ReasonRaised: a check found a new condition.
	ReasonRaised Reason = iota + 1
	// ReasonChanged: the condition stands with another detail.
	ReasonChanged
	// ReasonReopened: a cleared condition came back.
	ReasonReopened
	// ReasonCleared: the check no longer finds the condition.
	ReasonCleared
	// ReasonAcknowledged: the owner acknowledged the item.
	ReasonAcknowledged
	// ReasonResolved: the owner resolved the item.
	ReasonResolved
)

// String returns the reason's name, such as "raised", or a placeholder for
// an invalid reason.
func (r Reason) String() string {
	switch r {
	case ReasonRaised:
		return "raised"
	case ReasonChanged:
		return "changed"
	case ReasonReopened:
		return "reopened"
	case ReasonCleared:
		return "cleared"
	case ReasonAcknowledged:
		return "acknowledged"
	case ReasonResolved:
		return "resolved"
	}
	return fmt.Sprintf("Reason(%d)", int(r))
}

// ByOwner reports whether r is the owner's action rather than a check's.
func (r Reason) ByOwner() bool { return r == ReasonAcknowledged || r == ReasonResolved }

// Change is one entry of an item's history: the state it moved to, why,
// the condition's detail then, and when. Seq numbers an item's changes
// from 1.
type Change struct {
	Seq    int
	State  State
	Reason Reason
	Detail string
	At     time.Time
}

// Item is an attention item with its history, oldest change first. ID is
// the store's; zero for an item not stored yet.
type Item struct {
	ID      int64
	Kind    Kind
	Subject string
	Changes []Change
}

// Current returns the item's latest change; the zero Change when it has
// none.
func (it *Item) Current() Change {
	if len(it.Changes) == 0 {
		return Change{}
	}
	return it.Changes[len(it.Changes)-1]
}

// State returns the item's current state.
func (it *Item) State() State { return it.Current().State }

// next returns the change after the item's current one.
func (it *Item) next(state State, reason Reason, detail string, at time.Time) Change {
	return Change{Seq: it.Current().Seq + 1, State: state, Reason: reason, Detail: detail, At: at}
}
