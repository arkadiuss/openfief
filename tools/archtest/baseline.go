package archtest

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrList reports a malformed baseline or exceptions line.
var ErrList = errors.New("malformed list")

// Entry is one line of the baseline or the exceptions:
//
//	<rule> | <key> | <reason>
//
// Blank lines and lines starting with # are comments.
type Entry struct {
	Rule   string
	Key    string
	Reason string
	// Line is the entry's line number in its file.
	Line int
}

// ParseList reads a baseline or exceptions file. Every entry names a known
// rule and gives a reason, and no entry is listed twice.
func ParseList(data string) ([]Entry, error) {
	var out []Entry
	sc := bufio.NewScanner(strings.NewReader(data))
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, " | ")
		if len(parts) != 3 || strings.TrimSpace(parts[2]) == "" {
			return nil, fmt.Errorf("%w: line %d: want <rule> | <key> | <reason>", ErrList, n)
		}
		e := Entry{Rule: strings.TrimSpace(parts[0]), Key: strings.TrimSpace(parts[1]), Reason: strings.TrimSpace(parts[2]), Line: n}
		if !slices.Contains(Rules(), e.Rule) {
			return nil, fmt.Errorf("%w: line %d: unknown rule %q", ErrList, n, e.Rule)
		}
		if slices.ContainsFunc(out, func(o Entry) bool { return o.Rule == e.Rule && o.Key == e.Key }) {
			return nil, fmt.Errorf("%w: line %d: %s %s is listed twice", ErrList, n, e.Rule, e.Key)
		}
		out = append(out, e)
	}
	return out, nil
}

// Rules lists every rule, in the doc's order.
func Rules() []string {
	return []string{
		RuleContextMap, RuleLayering, RulePlugins, RuleSchemas, RuleMoney, RulePureDomain,
		RulePrivacy, RuleBanned, RuleStableTests, RuleCatalogTests,
	}
}

// Check runs every rule on a repository. digests are the embedded plugin
// catalog's image digests (CatalogTests).
func Check(r *Repo, digests []string) []Violation {
	out := slices.Concat(
		ContextMap(r), Layering(r), Plugins(r), Schemas(r), Money(r), PureDomain(r),
		Privacy(r), Banned(r), StableTests(r), CatalogTests(r, digests),
	)
	slices.SortFunc(out, func(a, b Violation) int {
		return cmp.Or(cmp.Compare(a.Rule, b.Rule), cmp.Compare(a.Key, b.Key))
	})
	return out
}

// Compare matches violations against the lists that allow them. fresh are
// the violations no list names: each fails the test. stale are the entries
// that match no violation: each fails too, so the baseline only shrinks and
// both lists stay exact.
func Compare(found []Violation, lists ...[]Entry) ([]Violation, []Entry) {
	all := slices.Concat(lists...)
	var fresh []Violation
	var stale []Entry
	for _, v := range found {
		if !slices.ContainsFunc(all, func(e Entry) bool { return e.Rule == v.Rule && e.Key == v.Key }) {
			fresh = append(fresh, v)
		}
	}
	for _, e := range all {
		if !slices.ContainsFunc(found, func(v Violation) bool { return e.Rule == v.Rule && e.Key == v.Key }) {
			stale = append(stale, e)
		}
	}
	return fresh, stale
}
