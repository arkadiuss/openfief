package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ErrInvalidLog reports stored versions that break the log's rules: versions
// of different keys passed as one transaction's history, or a version number
// that is missing, repeated or below 1.
var ErrInvalidLog = errors.New("invalid transaction log")

// VersionedTransaction is one stored version of a transaction. Versions of
// one key are numbered from 1; the highest is current and supersedes all
// earlier ones, which stay in the log (D-18).
type VersionedTransaction struct {
	Transaction Transaction
	Version     int
}

// ImportDecision says what to do with an incoming transaction.
type ImportDecision int

// The import decisions. The zero decision is not a decision.
const (
	// ImportUnchanged means the current version already has the same content:
	// re-importing it changes nothing.
	ImportUnchanged ImportDecision = iota + 1
	// ImportNew means the key is new: append it as version 1.
	ImportNew
	// ImportNewVersion means the content changed: append it as a new version
	// that supersedes the current one, which is kept.
	ImportNewVersion
)

// String returns the decision's name, such as "unchanged", or a placeholder
// for an invalid decision.
func (d ImportDecision) String() string {
	switch d {
	case ImportUnchanged:
		return "unchanged"
	case ImportNew:
		return "new"
	case ImportNewVersion:
		return "new-version"
	}
	return fmt.Sprintf("ImportDecision(%d)", int(d))
}

// ImportResult is DecideImport's answer.
type ImportResult struct {
	Decision ImportDecision
	// Version is the incoming transaction's version number: the one to
	// append, or for ImportUnchanged the current one it matches.
	Version int
	// Supersedes is the version a new version supersedes, or 0.
	Supersedes int
}

// DecideImport decides how incoming joins the log, given every stored version
// of its key (in any order; none for a new key). Re-imports are idempotent:
// content equal to the current version, by Fingerprint, is ImportUnchanged.
// Changed content becomes the next version, superseding the current one.
// incoming must be valid.
func DecideImport(versions []VersionedTransaction, incoming *Transaction) (ImportResult, error) {
	if err := incoming.Validate(); err != nil {
		return ImportResult{}, err
	}
	current, err := currentOf(incoming.Key, versions)
	if err != nil {
		return ImportResult{}, err
	}
	switch {
	case current == nil:
		return ImportResult{Decision: ImportNew, Version: 1}, nil
	case Fingerprint(&current.Transaction) == Fingerprint(incoming):
		return ImportResult{Decision: ImportUnchanged, Version: current.Version}, nil
	}
	return ImportResult{Decision: ImportNewVersion, Version: current.Version + 1, Supersedes: current.Version}, nil
}

// currentOf returns the highest version among versions, which must all have
// key and be numbered 1..n without gaps or repeats, or nil if there are none.
func currentOf(key TransactionKey, versions []VersionedTransaction) (*VersionedTransaction, error) {
	seen := make(map[int]bool, len(versions))
	var current *VersionedTransaction
	for i := range versions {
		v := &versions[i]
		if v.Transaction.Key != key {
			return nil, fmt.Errorf("%w: version of %s among versions of %s", ErrInvalidLog, v.Transaction.Key, key)
		}
		if v.Version < 1 || v.Version > len(versions) || seen[v.Version] {
			return nil, fmt.Errorf("%w: %s has versions that are not 1..%d", ErrInvalidLog, key, len(versions))
		}
		seen[v.Version] = true
		if current == nil || v.Version > current.Version {
			current = v
		}
	}
	return current, nil
}

// CurrentTransactions returns the current version of every key in log, in
// date order, then by key. Superseded versions are left out; they stay in
// the log but never count toward holdings, lots or balances.
func CurrentTransactions(log []VersionedTransaction) ([]Transaction, error) {
	byKey := map[TransactionKey][]VersionedTransaction{}
	for i := range log {
		byKey[log[i].Transaction.Key] = append(byKey[log[i].Transaction.Key], log[i])
	}
	out := make([]Transaction, 0, len(byKey))
	for key, versions := range byKey {
		current, err := currentOf(key, versions)
		if err != nil {
			return nil, err
		}
		out = append(out, current.Transaction)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := &out[i], &out[j]
		if c := a.Date.Compare(b.Date); c != 0 {
			return c < 0
		}
		if a.Key.Source != b.Key.Source {
			return a.Key.Source < b.Key.Source
		}
		return a.Key.ExternalID < b.Key.ExternalID
	})
	return out, nil
}

// Fingerprint returns a SHA-256 hex digest of t's content: every field, with
// postings in a canonical order and numbers in canonical form, so that two
// transactions have the same fingerprint exactly when they record the same
// thing. Posting order and trailing zeros (1.50 against 1.5) don't matter.
func Fingerprint(t *Transaction) string {
	postings := make([]string, len(t.Postings))
	for i := range t.Postings {
		p := &t.Postings[i]
		postings[i] = field(p.Book.String()) + field(p.Amount.Currency().String()) +
			field(p.Amount.Amount().String()) + field(p.Quantity.String())
	}
	sort.Strings(postings)
	var b strings.Builder
	for _, f := range []string{
		t.Key.Source, t.Key.ExternalID, string(t.Account), t.Kind.String(),
		t.Date.String(), dateOrEmpty(t), t.Description,
	} {
		b.WriteString(field(f))
	}
	for _, p := range postings {
		b.WriteString(field(p))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// field encodes s with its length, so no two field lists encode alike.
func field(s string) string { return strconv.Itoa(len(s)) + ":" + s }

// dateOrEmpty returns t's lot open date, or "" when it has none.
func dateOrEmpty(t *Transaction) string {
	if t.LotOpened.IsZero() {
		return ""
	}
	return t.LotOpened.String()
}
