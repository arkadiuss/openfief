package main

import (
	"regexp"
	"strings"
)

// decisionLine is how docs/DESIGN.md states an author decision: a list item starting with its bold
// bracketed ID, such as "- **[D-157]** The read API's ...".
var decisionLine = regexp.MustCompile(`^\s*- \*\*\[(D-[1-9]\d*)\]\*\*\s*(.*)$`)

// parseDecisions returns the text of each author decision in design, docs/DESIGN.md, by ID.
func parseDecisions(design string) map[string]string {
	out := map[string]string{}
	for line := range strings.SplitSeq(design, "\n") {
		if m := decisionLine.FindStringSubmatch(line); m != nil {
			out[m[1]] = m[2]
		}
	}
	return out
}

// names reports whether a decision's text names the element whose short name is short, written as code:
// "`ListBrokerAccounts`".
func names(text, short string) bool {
	return strings.Contains(text, "`"+short+"`")
}
