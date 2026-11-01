package main

import (
	"errors"
	"fmt"
	"strings"
)

// errBreaking reports breaking changes no entry excuses, or entries that are not allowed.
var errBreaking = errors.New("breaking contract changes")

// judgeInput is everything judge decides on.
type judgeInput struct {
	// Findings are buf breaking's findings against the base.
	Findings []Finding
	// Exceptions are the exceptions file's entries now, BaseExceptions the base's.
	Exceptions, BaseExceptions []Entry
	// Decisions are the author decisions' texts by ID (docs/DESIGN.md).
	Decisions map[string]string
	// Base indexes the base's contracts.
	Base *contracts
}

// judge returns the new entries that excuse the findings, or an error wrapping errBreaking that lists
// every problem:
//
//   - every entry, old or new, cites an author decision that exists, and a method entry's decision names
//     the method as code, such as "`ListBrokerAccounts`";
//   - a new entry, one the base's exceptions file lacks, removes what the base has: a method entry a
//     method, and a message entry a message that a new method entry of the same decision holds, as its
//     request, its response or a message they hold, so a decision excuses a method and its messages only;
//   - every finding is a removal that a new entry names, and every new entry excuses a finding.
//
// So an entry excuses its removal only in the change that adds it: once both are on the base, buf no
// longer reports the removal and the entry excuses nothing again.
func judge(in *judgeInput) ([]Entry, error) {
	fresh := newEntries(in.BaseExceptions, in.Exceptions)
	problems := checkDecisions(in.Exceptions, in.Decisions)
	problems = append(problems, checkFresh(fresh, in.Base)...)
	used := map[Entry]bool{}
	for i := range in.Findings {
		e, ok := excuse(&in.Findings[i], fresh, in.Base)
		if !ok {
			problems = append(problems, "not excused: "+in.Findings[i].String())
			continue
		}
		used[e] = true
	}
	for _, e := range fresh {
		if !used[e] {
			problems = append(problems, e.String()+": excuses nothing this change breaks; an entry goes in with its removal only")
		}
	}
	if len(problems) > 0 {
		return nil, fmt.Errorf("%w:\n  %s", errBreaking, strings.Join(problems, "\n  "))
	}
	return fresh, nil
}

// checkDecisions returns a problem for each entry whose decision doesn't exist, or doesn't name its
// method.
func checkDecisions(entries []Entry, decisions map[string]string) []string {
	var out []string
	for _, e := range entries {
		text, ok := decisions[e.Decision]
		switch {
		case !ok:
			out = append(out, e.String()+": no author decision "+e.Decision+" in docs/DESIGN.md")
		case e.Rule == ruleRPCDelete && !names(text, shortName(e.Element)):
			out = append(out, e.String()+": "+e.Decision+" doesn't name `"+shortName(e.Element)+"`")
		}
	}
	return out
}

// checkFresh returns a problem for each new entry that doesn't remove what the base has, or a message no
// new method entry of its decision holds.
func checkFresh(fresh []Entry, base *contracts) []string {
	methods := map[string][]string{}
	for _, e := range fresh {
		if e.Rule == ruleRPCDelete {
			methods[e.Decision] = append(methods[e.Decision], e.Element)
		}
	}
	var out []string
	for _, e := range fresh {
		switch {
		case e.Rule == ruleRPCDelete && base.methods[e.Element] == nil:
			out = append(out, e.String()+": the base has no such method")
		case e.Rule == ruleMessageDelete && !base.hasMessage(e.Element):
			out = append(out, e.String()+": the base has no such message")
		case e.Rule == ruleMessageDelete && !base.reachable(methods[e.Decision])[e.Element]:
			out = append(out, e.String()+": no method "+e.Decision+" removes holds this message")
		}
	}
	return out
}

// excuse returns the new entry that excuses f.
func excuse(f *Finding, fresh []Entry, base *contracts) (Entry, bool) {
	name, ok := removed(f, base)
	if !ok {
		return Entry{}, false
	}
	for _, e := range fresh {
		if e.Rule == f.Rule && e.Element == name {
			return e, true
		}
	}
	return Entry{}, false
}

// shortName returns the last part of a fully qualified name.
func shortName(name string) string {
	return name[strings.LastIndex(name, ".")+1:]
}
