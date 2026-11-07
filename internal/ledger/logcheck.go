package ledger

import (
	"sort"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// LogCheck is the result of checking the whole transaction log, every
// stored version of every transaction, current or superseded (D-19, D-200
// criterion 2): what it holds, by account, and every version whose postings
// don't sum to zero in a currency. Balanced reports that there is none, and
// then Sums, the postings of the whole log added up per currency, are all
// zero.
type LogCheck struct {
	// Accounts counts the log per account, in ID order.
	Accounts []AccountLogCount
	// Transactions, Versions and Superseded count the whole log: the
	// transactions by key, every stored version, and the versions a later
	// one superseded.
	Transactions, Versions, Superseded int
	// Unbalanced lists every version whose postings don't sum to zero in a
	// currency, one entry per currency, in account, key, version and
	// currency order.
	Unbalanced []UnbalancedVersion
	// Sums adds up the postings of every version per currency, in currency
	// order.
	Sums []kernel.Money
	// Logged lists every stored version, in account, key and version
	// order, without its postings.
	Logged []LoggedVersion
}

// LoggedVersion is one stored version of a transaction: its account, key,
// version and kind, and whether a later version superseded it.
type LoggedVersion struct {
	Account    AccountID
	Key        TransactionKey
	Version    int
	Kind       TransactionKind
	Superseded bool
}

// Balanced reports whether every version of every transaction sums to zero
// in every currency, and so the whole log does.
func (c *LogCheck) Balanced() bool {
	if len(c.Unbalanced) > 0 {
		return false
	}
	for _, s := range c.Sums {
		if !s.IsZero() {
			return false
		}
	}
	return true
}

// AccountLogCount is one account's part of the log: its transactions by
// key, their stored versions and the superseded ones among them, each
// counted in the account its own version names.
type AccountLogCount struct {
	Account                            AccountID
	Transactions, Versions, Superseded int
}

// UnbalancedVersion is one version of a transaction whose postings sum to
// Sum, not zero, in Sum's currency.
type UnbalancedVersion struct {
	Account AccountID
	Key     TransactionKey
	Version int
	Sum     kernel.Money
}

// versionKey names one stored version.
type versionKey struct {
	key     TransactionKey
	version int
}

// CheckLog checks log, every stored version of every transaction in any
// order. A version listed twice, as the logs of two accounts list a
// transaction that a correction moved between them, counts once. It adds
// up every version's postings per currency and lists each version that
// doesn't balance; a version is superseded when a later version of its key
// is stored, a deletion included.
func CheckLog(log []VersionedTransaction) LogCheck {
	d := distinctVersions(log)
	versions, latest := d.versions, d.latest
	counts := map[AccountID]*AccountLogCount{}
	keysIn := map[AccountID]map[TransactionKey]bool{}
	totals := map[kernel.Currency]kernel.Decimal{}
	var c LogCheck
	for i := range versions {
		v := &versions[i]
		account := v.Transaction.Account
		if counts[account] == nil {
			counts[account] = &AccountLogCount{Account: account}
			keysIn[account] = map[TransactionKey]bool{}
		}
		superseded := v.Version < latest[v.Transaction.Key]
		counts[account].count(!keysIn[account][v.Transaction.Key], superseded)
		keysIn[account][v.Transaction.Key] = true
		if superseded {
			c.Superseded++
		}
		c.Logged = append(c.Logged, LoggedVersion{
			Account: account, Key: v.Transaction.Key, Version: v.Version, Kind: v.Transaction.Kind, Superseded: superseded,
		})
		c.Unbalanced = append(c.Unbalanced, addPostings(totals, v)...)
	}
	c.Transactions, c.Versions = len(latest), len(versions)
	for _, n := range counts {
		c.Accounts = append(c.Accounts, *n)
	}
	sort.Slice(c.Accounts, func(i, j int) bool { return c.Accounts[i].Account < c.Accounts[j].Account })
	sort.Slice(c.Unbalanced, func(i, j int) bool { return unbalancedBefore(&c.Unbalanced[i], &c.Unbalanced[j]) })
	sort.Slice(c.Logged, func(i, j int) bool { return loggedBefore(&c.Logged[i], &c.Logged[j]) })
	for cur, d := range totals {
		c.Sums = append(c.Sums, kernel.NewMoney(d, cur))
	}
	sort.Slice(c.Sums, func(i, j int) bool { return c.Sums[i].Currency().String() < c.Sums[j].Currency().String() })
	return c
}

// distinct is each version of a log once, in the order first listed, and
// the latest version number of each key.
type distinct struct {
	versions []VersionedTransaction
	latest   map[TransactionKey]int
}

// distinctVersions returns each version in log once and each key's latest
// version number.
func distinctVersions(log []VersionedTransaction) distinct {
	seen := map[versionKey]bool{}
	latest := map[TransactionKey]int{}
	var versions []VersionedTransaction
	for i := range log {
		k := versionKey{key: log[i].Transaction.Key, version: log[i].Version}
		if seen[k] {
			continue
		}
		seen[k] = true
		versions = append(versions, log[i])
		latest[k.key] = max(latest[k.key], k.version)
	}
	return distinct{versions: versions, latest: latest}
}

// count counts one version of the account: of a key not counted before
// when newKey, and superseded when superseded.
func (n *AccountLogCount) count(newKey, superseded bool) {
	n.Versions++
	if newKey {
		n.Transactions++
	}
	if superseded {
		n.Superseded++
	}
}

// addPostings adds v's postings to totals per currency and returns v as
// unbalanced in each currency its postings don't sum to zero in.
func addPostings(totals map[kernel.Currency]kernel.Decimal, v *VersionedTransaction) []UnbalancedVersion {
	var out []UnbalancedVersion
	for _, sum := range sumsByCurrency(v.Transaction.Postings) {
		totals[sum.Currency()] = totals[sum.Currency()].Add(sum.Amount())
		if !sum.IsZero() {
			out = append(out, UnbalancedVersion{Account: v.Transaction.Account, Key: v.Transaction.Key, Version: v.Version, Sum: sum})
		}
	}
	return out
}

// unbalancedBefore orders unbalanced versions by account, key, version and
// currency.
func unbalancedBefore(a, b *UnbalancedVersion) bool {
	if c := compareVersions(versionRef{a.Account, a.Key, a.Version}, versionRef{b.Account, b.Key, b.Version}); c != 0 {
		return c < 0
	}
	return a.Sum.Currency().String() < b.Sum.Currency().String()
}

// loggedBefore orders logged versions by account, key and version.
func loggedBefore(a, b *LoggedVersion) bool {
	return compareVersions(versionRef{a.Account, a.Key, a.Version}, versionRef{b.Account, b.Key, b.Version}) < 0
}

// versionRef names a version in an account, for ordering.
type versionRef struct {
	account AccountID
	key     TransactionKey
	version int
}

// compareVersions compares two versions by account, key source, external
// ID and version number.
func compareVersions(a, b versionRef) int {
	switch {
	case a.account != b.account:
		return strings.Compare(string(a.account), string(b.account))
	case a.key.Source != b.key.Source:
		return strings.Compare(a.key.Source, b.key.Source)
	case a.key.ExternalID != b.key.ExternalID:
		return strings.Compare(a.key.ExternalID, b.key.ExternalID)
	}
	return a.version - b.version
}
