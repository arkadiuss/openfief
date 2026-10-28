package main

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestPlaceOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rel  string
		want place
	}{
		{"internal/ledger", place{"ledger", layerDomain}},
		{"internal/ledger/app", place{"ledger", layerApp}},
		{"internal/ledger/postgres", place{"ledger", layerAdapter}},
		{"internal/jurisdiction/pl", place{"jurisdiction/pl", layerDomain}},
		{"internal/kernel", place{"kernel", layerDomain}},
		{"cmd/operator", place{"cmd", layerBinary}},
		{"tools/docgen", place{"tools", layerTool}},
	}
	for _, tt := range tests {
		t.Run(tt.rel, func(t *testing.T) {
			t.Parallel()
			if got := placeOf(tt.rel); got != tt.want {
				t.Errorf("placeOf(%q) = %+v, want %+v", tt.rel, got, tt.want)
			}
		})
	}
}

func TestLayeringFindings(t *testing.T) {
	t.Parallel()
	const m = "example.com/x"
	pkgs := []goPackage{
		{importPath: m + "/internal/ledger", imports: []string{m + "/internal/kernel", "fmt"}},
		{importPath: m + "/internal/ledger/app", imports: []string{m + "/internal/ledger", m + "/internal/ledger/postgres"}},
		{importPath: m + "/internal/valuation", imports: []string{m + "/internal/ledger/app"}},
		{importPath: m + "/internal/ingest/app", imports: []string{m + "/internal/ledger/postgres"}},
		{importPath: m + "/cmd/core", imports: []string{m + "/internal/ledger/postgres"}},
	}
	got := layeringFindings(m, pkgs)
	want := []string{
		"`internal/ingest/app` imports `internal/ledger/postgres`: an app package imports an adapter",
		"`internal/ledger/app` imports `internal/ledger/postgres`: an app package imports an adapter",
		"`internal/valuation` imports `internal/ledger/app`: a domain package imports an app package",
	}
	if !slices.Equal(got, want) {
		t.Errorf("layeringFindings() =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestStatusByDecision(t *testing.T) {
	t.Parallel()
	decisions := parseDecisions("## 2. Scope\n\n- **[D-1]** First.\n\n### 3.3 Ledger\n\n- **[D-18]** Log.\n  - **[D-26]** Nested.\n")
	ref := func(r string) []evidence { return []evidence{{Ref: r}} }
	rows := []coverageRow{
		{ID: "REQ-a", Evidence: ref("design-decisions:D-18#1"), DirectSymbolCount: 2, TestCount: 1},
		{ID: "REQ-b", Evidence: ref("design-decisions:D-18#2"), DirectSymbolCount: 1},
		{ID: "REQ-c", Evidence: ref("design-decisions:D-26#1")},
		{ID: "REQ-d", Evidence: ref("design-decisions:D-26#2"), Status: "closed", DirectSymbolCount: 1},
		{ID: "REQ-tooling-x", TransitiveSymbolCount: 1, TestCount: 1},
	}
	all := statusByDecision(decisions, rows)
	statuses, other := all.decisions, all.other
	labels := map[string]string{}
	for i := range statuses {
		labels[statuses[i].id] = statuses[i].label()
	}
	want := map[string]string{"D-1": statusNoReq, "D-18": statusBuilt, "D-26": statusNone}
	for id, label := range want {
		if labels[id] != label {
			t.Errorf("%s label = %q, want %q", id, labels[id], label)
		}
	}
	if statuses[1].section != "3.3 Ledger" {
		t.Errorf("D-18 section = %q, want %q", statuses[1].section, "3.3 Ledger")
	}
	if got := other["tooling"]; got.reqs != 1 || got.built != 1 || got.tested != 1 {
		t.Errorf("other[tooling] = %+v, want 1 requirement built and tested", got)
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()
	tests := []struct{ in, want string }{
		{"Short. Second sentence.", "Short."},
		{"See [DESIGN](x.md) now.", "See DESIGN now."},
		{"Uses `a | b` pipes.", "Uses `a \\| b` pipes."},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := summary(tt.in); got != tt.want {
				t.Errorf("summary(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSummaryShortensLongText(t *testing.T) {
	t.Parallel()
	got := summary(strings.Repeat("alpha beta ", 20) + "`open code")
	if len(got) > summaryLength+len(" …") || !strings.HasSuffix(got, " …") || strings.Count(got, "`")%2 != 0 {
		t.Errorf("summary() = %q, want at most %d characters ending in an ellipsis, with code spans closed", got, summaryLength)
	}
}

func TestPageRefProblems(t *testing.T) {
	t.Parallel()
	repo := fstest.MapFS{".kb/requirements/REQ-ledger-log.md": {}, ".kb/tests/TEST-ledger-log.md": {}}
	known := knownRefs{repo: repo, symbols: "  - id: SYM-ledger-append\n", tasks: map[string]bool{"test": true, "docs:build": true}}
	page := "REQ-ledger-log and TEST-ledger-log and SYM-ledger-append.\n" +
		"Patterns like TEST-tooling-* and REQ-x* are fine.\n" +
		"Run `task test` or `task docs:build`.\n" +
		"REQ-missing and `task nope`.\n"
	got := pageRefProblems("p.md", page, known)
	want := []string{"p.md:4: REQ-missing is not in kibi", `p.md:4: there is no task "nope"`}
	if !slices.Equal(got, want) {
		t.Errorf("pageRefProblems() = %v, want %v", got, want)
	}
}

func TestTasksPage(t *testing.T) {
	t.Parallel()
	var list taskList
	list.Tasks = append(list.Tasks,
		task{Name: "docs:build", Desc: "Build the site"},
		task{Name: "docs:fresh", Desc: "Gate | range"},
		task{Name: "test", Desc: "Tests"})
	got := tasksPage(list)
	for _, want := range []string{"## docs\n", "| `task docs:fresh` | Gate \\| range |", "## test\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("tasksPage() lacks %q", want)
		}
	}
}
