package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const mod = "github.com/arkadiuss/openfief/"

// testGraph is a small synthetic workspace: a shared kernel, a ledger on it, a valuation package whose tests
// import the ledger, a command, a plugin on the SDK module, the operator (whose tests read deploy/crds), the UI
// (which embeds a file) and the end-to-end suite, which only builds with its tag.
func testGraph() *graph {
	return newGraph([]*pkg{
		{importPath: mod + "internal/kernel", dir: "internal/kernel", module: ".", testable: true},
		{importPath: mod + "internal/ledger", dir: "internal/ledger", module: ".", imports: []string{mod + "internal/kernel", "fmt"}, testable: true},
		{importPath: mod + "internal/valuation", dir: "internal/valuation", module: ".", imports: []string{mod + "internal/kernel"}, testImports: []string{mod + "internal/ledger"}, testable: true},
		{importPath: mod + "cmd/core", dir: "cmd/core", module: ".", imports: []string{mod + "internal/ledger"}, testable: true},
		{importPath: mod + "plugins/nbp-fx", dir: "plugins/nbp-fx", module: ".", imports: []string{mod + "sdk/go/connector"}, testable: true},
		{importPath: mod + "sdk/go/connector", dir: "sdk/go/connector", module: "sdk/go", testable: true},
		{importPath: mod + "internal/plugins/operator", dir: "internal/plugins/operator", module: ".", testable: true, reads: []string{"deploy/crds"}},
		{importPath: mod + "internal/views/ui", dir: "internal/views/ui", module: ".", imports: []string{mod + "internal/kernel"}, embeds: []string{"internal/views/ui/web/app.css"}, testable: true},
		{importPath: mod + "test/e2e", dir: "test/e2e", module: ".", testImports: []string{mod + "internal/ledger"}},
	})
}

// testScripts are two script tests: one reads deploy/, one the design documents.
func testScripts() []scriptTest {
	return []scriptTest{
		{file: "scripts/tests/demo-validate.test.sh", reads: []string{"deploy"}},
		{file: "scripts/tests/check-docs.test.sh", reads: []string{"docs/DESIGN.md", "docs/proposals"}},
	}
}

// steps renders a plan as "task VAR=value ..." lines, or "full" and the reasons.
func steps(p *plan) []string {
	if p.full {
		return append([]string{"full"}, p.reasons...)
	}
	out := make([]string, 0, len(p.steps))
	for _, s := range p.steps {
		out = append(out, strings.Join(append([]string{s.task}, s.vars...), " "))
	}
	return out
}

func TestHookPlan(t *testing.T) {
	t.Parallel()
	always := func(mid ...string) []string {
		return append(append([]string{}, mid...), "docs:check", "kibi:staged")
	}
	tests := []struct {
		name    string
		changes []change
		want    []string
	}{
		{
			name:    "a docs-only change skips Go but builds the docs",
			changes: []change{{path: "docs/concepts/kernel.md"}},
			want:    []string{"docs:check", "docs:build", "kibi:staged"},
		},
		{
			name:    "a kibi-only change builds the docs too",
			changes: []change{{path: ".kb/requirements/REQ-x.md"}, {path: ".kb/symbols.yaml"}},
			want:    []string{"docs:check", "docs:build", "kibi:staged"},
		},
		{
			name:    "a change to a design document runs the script test that reads it",
			changes: []change{{path: "docs/DESIGN.md"}},
			want:    []string{"test:script SCRIPT=scripts/tests/check-docs.test.sh", "docs:check", "docs:build", "kibi:staged"},
		},
		{
			name:    "one plugin checks only that plugin",
			changes: []change{{path: "plugins/nbp-fx/main.go"}},
			want: always("fmt:check:files MODULE=. FILES=plugins/nbp-fx/main.go", "lint:config",
				"lint:go MODULE=. PKGS=./plugins/nbp-fx", "vuln:go MODULE=. PKGS=./plugins/nbp-fx", "test:go MODULE=. PKGS=./plugins/nbp-fx"),
		},
		{
			name:    "a shared package checks every package that imports it, and lints the ones whose tests do",
			changes: []change{{path: "internal/kernel/money.go"}},
			want: always("fmt:check:files MODULE=. FILES=internal/kernel/money.go", "lint:config",
				"lint:go MODULE=. PKGS=./cmd/core ./internal/kernel ./internal/ledger ./internal/valuation ./internal/views/ui ./test/e2e",
				"vuln:go MODULE=. PKGS=./cmd/core ./internal/kernel ./internal/ledger ./internal/valuation ./internal/views/ui",
				"test:go MODULE=. PKGS=./cmd/core ./internal/kernel ./internal/ledger ./internal/valuation ./internal/views/ui"),
		},
		{
			name:    "an SDK change checks the SDK module and the root packages importing it",
			changes: []change{{path: "sdk/go/connector/server.go"}},
			want: always("fmt:check:files MODULE=sdk/go FILES=connector/server.go", "lint:config",
				"lint:go MODULE=. PKGS=./plugins/nbp-fx", "vuln:go MODULE=. PKGS=./plugins/nbp-fx", "test:go MODULE=. PKGS=./plugins/nbp-fx",
				"lint:go MODULE=sdk/go PKGS=./connector", "vuln:go MODULE=sdk/go PKGS=./connector", "test:go MODULE=sdk/go PKGS=./connector"),
		},
		{
			name:    "an embedded file changes its package like Go source",
			changes: []change{{path: "internal/views/ui/web/app.css"}},
			want: always("lint:config", "lint:go MODULE=. PKGS=./internal/views/ui",
				"vuln:go MODULE=. PKGS=./internal/views/ui", "test:go MODULE=. PKGS=./internal/views/ui"),
		},
		{
			name:    "test data reruns only its package's tests",
			changes: []change{{path: "internal/ledger/testdata/golden.json"}},
			want:    always("test:go MODULE=. PKGS=./internal/ledger"),
		},
		{
			name:    "a generated manifest regenerates, validates and reruns the tests that read it",
			changes: []change{{path: "deploy/crds/plugins.yaml"}},
			want: []string{
				"gen:check", "test:go MODULE=. PKGS=./internal/plugins/operator",
				"test:script SCRIPT=scripts/tests/demo-validate.test.sh", "docs:check", "demo:validate", "kibi:staged",
			},
		},
		{
			name:    "a workflow change runs actionlint",
			changes: []change{{path: ".github/workflows/ci.yml"}},
			want:    always("lint:ci"),
		},
		{
			name:    "a deleted file builds the docs, whose links may name it",
			changes: []change{{path: ".claude/skills/x/SKILL.md", deleted: true}},
			want:    []string{"docs:check", "docs:build", "kibi:staged"},
		},
		{
			name:    "a plugin manifest checks the labels and the plugin's tests",
			changes: []change{{path: "plugins/nbp-fx/manifest.json"}},
			want:    always("plugins:labels:check", "test:go MODULE=. PKGS=./plugins/nbp-fx"),
		},
		{
			name:    "go.mod selects the full check",
			changes: []change{{path: "internal/kernel/money.go"}, {path: "go.mod"}},
			want:    []string{"full", "go.mod: module dependencies affect every package"},
		},
		{
			name:    "an unknown root file selects the full check",
			changes: []change{{path: ".gitattributes"}},
			want:    []string{"full", ".gitattributes: no rule names this file; when in doubt, everything runs"},
		},
		{
			name:    "a Go file outside every package selects the full check",
			changes: []change{{path: "internal/gone/gone.go", deleted: true}},
			want:    []string{"full", "internal/gone/gone.go: Go file outside any package (a removed or new package?); when in doubt, everything runs"},
		},
		{
			name:    "a file no package holds or reads selects the full check",
			changes: []change{{path: "sdk/go/LICENSE"}},
			want:    []string{"full", "sdk/go/LICENSE: no package holds or reads it; when in doubt, everything runs"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := hookPlan(tt.changes, testGraph(), testScripts(), []string{".", "sdk/go"})
			got := steps(&p)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("hookPlan() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHookPlanSaysWhy(t *testing.T) {
	t.Parallel()
	p := hookPlan([]change{{path: "internal/kernel/money.go"}}, testGraph(), testScripts(), []string{".", "sdk/go"})
	var test step
	for _, s := range p.steps {
		if s.task == "test:go" {
			test = s
		}
	}
	want := []string{
		"cmd/core: imports internal/ledger",
		"internal/kernel: changed internal/kernel/money.go",
		"internal/ledger: imports internal/kernel",
		"internal/valuation: imports internal/kernel",
		"internal/views/ui: imports internal/kernel",
	}
	if diff := cmp.Diff(want, test.details); diff != "" {
		t.Errorf("test:go details mismatch (-want +got):\n%s", diff)
	}
	skipped := make([]string, 0, len(p.skipped))
	for _, s := range p.skipped {
		skipped = append(skipped, s.task)
	}
	for _, task := range []string{"docs:build", "demo:validate", "gen:check", "lint:ci"} {
		if !strings.Contains(strings.Join(skipped, " "), task) {
			t.Errorf("skipped = %v, want it to name %s with a reason", skipped, task)
		}
	}
}

func TestClassify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want area
	}{
		{"go.mod", areaFull},
		{"sdk/go/go.sum", areaFull},
		{"Taskfile.yml", areaFull},
		{"tools/docgen/main.go", areaFull},
		{"scripts/ci-images.sh", areaFull},
		{"proto/openfief/connector/v1/connector.proto", areaFull},
		{"sdk/go/gen/openfief/connector/v1/connector.pb.go", areaFull},
		{"design/tokens.css", areaFull},
		{"docs/DESIGN.md", areaDocs},
		{"README.md", areaDocs},
		{"LICENSE", areaDocs},
		{"site/hugo.toml", areaDocs},
		{".kb/symbols.yaml", areaKibi},
		{".github/workflows/ci.yml", areaWorkflows},
		{"deploy/demo/stack.yaml", areaDeploy},
		{"renovate.json", areaQuiet},
		{".claude/skills/design-docs/SKILL.md", areaQuiet},
		{"internal/kernel/money.go", areaCode},
		{"cmd/core/Dockerfile", areaCode},
		{".gitattributes", areaUnknown},
		{"go.mod.bak", areaUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			if got := classify(tt.path).area; got != tt.want {
				t.Errorf("classify(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestNeedsCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		changes []change
		want    bool
	}{
		{"docs and requirements only", []change{{path: "docs/x.md"}, {path: ".kb/symbols.yaml"}, {path: "README.md"}}, false},
		{"agent instructions only", []change{{path: ".claude/skills/x/SKILL.md"}}, false},
		{"a Go file", []change{{path: "docs/x.md"}, {path: "internal/kernel/money.go"}}, true},
		{"a manifest", []change{{path: "deploy/core/deployment.yaml"}}, true},
		{"a workflow", []change{{path: ".github/workflows/ci.yml"}}, true},
		{"an unknown file", []change{{path: ".gitattributes"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := needsCode(tt.changes).code; got != tt.want {
				t.Errorf("needsCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
