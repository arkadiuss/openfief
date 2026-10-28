package main

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// coverageReport is the part of `kibi coverage --by req --include-passing --format json` that docgen reads.
type coverageReport struct {
	Rows []coverageRow `json:"rows"`
}

type coverageRow struct {
	ID                    string     `json:"id"`
	Title                 string     `json:"title"`
	Status                string     `json:"status"`
	Evidence              []evidence `json:"evidence"`
	DirectSymbolCount     int        `json:"directSymbolCount"`
	TransitiveSymbolCount int        `json:"transitiveSymbolCount"`
	TestCount             int        `json:"testCount"`
}

type evidence struct {
	Ref string `json:"ref"`
}

// decisionStatuses is every decision's status, plus the requirements that come from no decision, by area.
type decisionStatuses struct {
	decisions []decisionStatus
	other     map[string]decisionStatus
}

// decision is one D-n entry of DESIGN.md with the section it sits in.
type decision struct {
	id      string
	section string
	text    string
}

// decisionStatus counts a decision's open requirements, how many have implementing code, and how many tests.
type decisionStatus struct {
	decision

	reqs   int
	built  int
	tested int
}

var (
	decisionRe    = regexp.MustCompile(`^\s*- \*\*\[(D-\d+)\]\*\* (.+)$`)
	headingRe     = regexp.MustCompile(`^#{2,3} (.+)$`)
	decisionRefRe = regexp.MustCompile(`^design-decisions:(D-\d+)(?:#|$)`)
	mdLinkRe      = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
)

const (
	statusTested  = "Built and tested"
	statusBuilt   = "Built, some requirements untested"
	statusPartial = "Partly built"
	statusNone    = "Not built"
	statusNoReq   = "No requirement in kibi"
	summaryLength = 140
)

// parseDecisions reads DESIGN.md's decisions in document order.
func parseDecisions(design string) []decision {
	var out []decision
	section := ""
	for line := range strings.Lines(design) {
		line = strings.TrimRight(line, "\n")
		if m := headingRe.FindStringSubmatch(line); m != nil {
			section = m[1]
			continue
		}
		if m := decisionRe.FindStringSubmatch(line); m != nil {
			out = append(out, decision{id: m[1], section: section, text: m[2]})
		}
	}
	return out
}

// decisionOf returns the decision a requirement was imported from, or "" for other requirements.
func decisionOf(row *coverageRow) string {
	for _, e := range row.Evidence {
		if m := decisionRefRe.FindStringSubmatch(e.Ref); m != nil {
			return m[1]
		}
	}
	return ""
}

// statusByDecision counts each decision's open requirements; the rest are returned grouped by ID area.
func statusByDecision(decisions []decision, rows []coverageRow) decisionStatuses {
	byID := map[string]*decisionStatus{}
	out := decisionStatuses{decisions: make([]decisionStatus, len(decisions)), other: map[string]decisionStatus{}}
	for i, d := range decisions {
		out.decisions[i] = decisionStatus{decision: d}
		byID[d.id] = &out.decisions[i]
	}
	for i := range rows {
		row := &rows[i]
		if row.Status == "closed" {
			continue
		}
		if s, ok := byID[decisionOf(row)]; ok {
			s.add(row)
			continue
		}
		area := requirementArea(row.ID)
		s := out.other[area]
		s.add(row)
		out.other[area] = s
	}
	return out
}

func (s *decisionStatus) add(row *coverageRow) {
	s.reqs++
	if row.DirectSymbolCount+row.TransitiveSymbolCount > 0 {
		s.built++
	}
	if row.TestCount > 0 {
		s.tested++
	}
}

// requirementArea is the second part of a requirement ID, such as "ledger" in REQ-ledger-versioning.
func requirementArea(id string) string {
	_, rest, ok := strings.Cut(id, "-")
	if !ok {
		return id
	}
	area, _, _ := strings.Cut(rest, "-")
	return area
}

func (s *decisionStatus) label() string {
	switch {
	case s.reqs == 0:
		return statusNoReq
	case s.built == s.reqs && s.tested == s.reqs:
		return statusTested
	case s.built == s.reqs:
		return statusBuilt
	case s.built > 0:
		return statusPartial
	default:
		return statusNone
	}
}

// summary shortens a decision to its first sentence, without links, and keeps code spans closed.
func summary(text string) string {
	s := firstSentence(mdLinkRe.ReplaceAllString(text, "$1"))
	if len(s) > summaryLength {
		cut := strings.LastIndex(s[:summaryLength], " ")
		s = s[:max(cut, 0)] + " …"
	}
	if strings.Count(s, "`")%2 == 1 {
		s += "`"
	}
	return s
}

// statusPage renders the verification matrix: every decision with its requirements, code and tests.
func statusPage(decisions []decision, rows []coverageRow) string {
	all := statusByDecision(decisions, rows)
	statuses, other := all.decisions, all.other
	var b strings.Builder
	b.WriteString("---\ntitle: Status\ndescription: Every decision with the requirements derived from it, whether code implements them, and whether tests verify them.\n---\n\n")
	b.WriteString("Generated on every build from DESIGN.md and kibi's coverage of its requirements. A requirement counts as built when a kibi symbol implements it, and as tested when a test is linked to it. kibi's [requirements report](/requirements/) shows each requirement with its evidence.\n\n")
	b.WriteString(statusPie(statuses))
	section := ""
	for _, s := range statuses {
		if s.section != section {
			section = s.section
			fmt.Fprintf(&b, "\n## %s\n\n| Decision | Summary | Requirements | Built | Tested | Status |\n|---|---|---|---|---|---|\n", section)
		}
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d | %s |\n", s.id, summary(s.text), s.reqs, s.built, s.tested, s.label())
	}
	b.WriteString("\n## Other requirements\n\nRequirements that come from the development rules or from agents' own modeling rather than from one decision, grouped by the area in their ID.\n\n| Area | Requirements | Built | Tested |\n|---|---|---|---|\n")
	for _, area := range slices.Sorted(maps.Keys(other)) {
		s := other[area]
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", area, s.reqs, s.built, s.tested)
	}
	return b.String()
}

// statusColors gives each status the same pie color in light and dark mode: green for done, amber for
// partial, gray for not built, red for a decision kibi doesn't trace at all.
func statusColors() []struct{ label, color string } {
	return []struct{ label, color string }{
		{statusTested, "#3f8f47"},
		{statusBuilt, "#8bc34a"},
		{statusPartial, "#f2b134"},
		{statusNone, "#b8b8b8"},
		{statusNoReq, "#d9534f"},
	}
}

func statusPie(statuses []decisionStatus) string {
	counts := map[string]int{}
	for i := range statuses {
		counts[statuses[i].label()]++
	}
	var vars, data strings.Builder
	n := 0
	for _, s := range statusColors() {
		if counts[s.label] == 0 {
			continue
		}
		n++
		fmt.Fprintf(&vars, ", 'pie%d': '%s'", n, s.color)
		fmt.Fprintf(&data, "  \"%s\" : %d\n", s.label, counts[s.label])
	}
	return fmt.Sprintf("```mermaid\n%%%%{init: {'themeVariables': {'pieStrokeColor': '#ffffff', 'pieOuterStrokeWidth': '0px'%s}}}%%%%\npie showData\n  title Decisions by status\n%s```\n", vars.String(), data.String())
}
