package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// errExceptions reports an exceptions file or entry that is malformed or not allowed.
var errExceptions = errors.New("invalid breaking-change exception")

// The rules an entry may excuse: removals only. An author decision removes a method or a message; any
// other breaking change, such as a changed field type, has no exception.
const (
	ruleRPCDelete     = "RPC_NO_DELETE"
	ruleMessageDelete = "MESSAGE_NO_DELETE"
)

// entryFields is how many fields an entry line has: decision, rule, element.
const entryFields = 3

// decisionID is what an author decision's ID looks like (docs/DESIGN.md), and elementName a fully
// qualified protobuf name.
var (
	decisionID  = regexp.MustCompile(`^D-[1-9]\d*$`)
	elementName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)+$`)
)

// Entry is one exception: the author decision that names the removal, the buf rule it excuses, and the
// fully qualified name of the removed method or message.
type Entry struct {
	Decision string
	Rule     string
	Element  string
}

// String returns the entry as its line.
func (e Entry) String() string { return e.Decision + " " + e.Rule + " " + e.Element }

// parseExceptions reads an exceptions file: one entry a line, as "D-157 RPC_NO_DELETE
// openfief.views.v1.ReadService.ListBrokerAccounts"; blank lines and lines starting with # are ignored.
// A malformed line, a rule other than a removal, or an entry listed twice fails.
func parseExceptions(text string) ([]Entry, error) {
	var out []Entry
	seen := map[Entry]bool{}
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		e, err := parseEntry(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n+1, err)
		}
		if seen[e] {
			return nil, fmt.Errorf("%w: line %d: %s is listed twice", errExceptions, n+1, e)
		}
		seen[e] = true
		out = append(out, e)
	}
	return out, nil
}

// parseEntry reads one entry line.
func parseEntry(line string) (Entry, error) {
	f := strings.Fields(line)
	if len(f) != entryFields {
		return Entry{}, fmt.Errorf("%w: %q: want DECISION RULE ELEMENT", errExceptions, line)
	}
	e := Entry{Decision: f[0], Rule: f[1], Element: f[2]}
	switch {
	case !decisionID.MatchString(e.Decision):
		return Entry{}, fmt.Errorf("%w: %q: %q is no author decision ID such as D-157", errExceptions, line, e.Decision)
	case e.Rule != ruleRPCDelete && e.Rule != ruleMessageDelete:
		return Entry{}, fmt.Errorf("%w: %q: only %s and %s can be excused", errExceptions, line, ruleRPCDelete, ruleMessageDelete)
	case !elementName.MatchString(e.Element):
		return Entry{}, fmt.Errorf("%w: %q: %q is no fully qualified protobuf name", errExceptions, line, e.Element)
	}
	return e, nil
}

// newEntries returns the entries of now that base doesn't have, in now's order. Only these excuse
// anything: an entry already on the base excused its removal when it went in.
func newEntries(base, now []Entry) []Entry {
	old := make(map[Entry]bool, len(base))
	for _, e := range base {
		old[e] = true
	}
	var out []Entry
	for _, e := range now {
		if !old[e] {
			out = append(out, e)
		}
	}
	return out
}
