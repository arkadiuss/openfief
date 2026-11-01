package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// errUnmarked reports a commit that adds a breaking-change exception without marking itself breaking.
var errUnmarked = errors.New("a commit that adds a breaking-change exception must be marked breaking")

// breakingSubject is a Conventional Commits subject marked breaking, "type(scope)!: summary", and
// breakingFooter the footer that marks one, "BREAKING CHANGE: what".
var (
	breakingSubject = regexp.MustCompile(`^[a-z]+(\([^()]*\))?!: \S`)
	breakingFooter  = regexp.MustCompile(`^BREAKING[ -]CHANGE: \S`)
)

// markedBreaking reports whether msg, a commit message, marks its commit breaking: `!` before the
// subject's colon, or a BREAKING CHANGE footer. Comment lines, as git's editor template leaves them, are
// not the message.
func markedBreaking(msg string) bool {
	lines := strings.Split(msg, "\n")
	if breakingSubject.MatchString(strings.TrimSpace(lines[0])) {
		return true
	}
	for _, l := range lines[1:] {
		if !strings.HasPrefix(l, "#") && breakingFooter.MatchString(l) {
			return true
		}
	}
	return false
}

// checkMarked fails when after, the exceptions file a commit leaves, has an entry before lacks, and
// msg doesn't mark the commit breaking. A malformed file fails too; removing entries needs no mark.
func checkMarked(before, after, msg string) error {
	old, err := parseExceptions(before)
	if err != nil {
		return fmt.Errorf("the exceptions file before the commit: %w", err)
	}
	now, err := parseExceptions(after)
	if err != nil {
		return fmt.Errorf("the exceptions file of the commit: %w", err)
	}
	fresh := newEntries(old, now)
	if len(fresh) == 0 || markedBreaking(msg) {
		return nil
	}
	return fmt.Errorf("%w (`!` or a BREAKING CHANGE: footer): it adds %s", errUnmarked, fresh[0])
}
