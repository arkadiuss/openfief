package main

import (
	"path"
	"strings"
)

// area is what kind of input a changed path is, which decides the checks it needs.
type area int

const (
	// areaUnknown is a path no rule names; the selector falls back to the full check for it.
	areaUnknown area = iota
	// areaFull can change the result of any check: dependencies, lint and task configuration, the tools
	// and scripts that implement the checks, contracts and generated code, embedded design assets.
	areaFull
	// areaDocs is hand-written documentation and the docs site.
	areaDocs
	// areaKibi is the requirements knowledge base.
	areaKibi
	// areaWorkflows is the GitHub Actions configuration.
	areaWorkflows
	// areaDeploy is the Kubernetes manifests, including the demo's GitOps layers.
	areaDeploy
	// areaQuiet is read by no check: agent instructions and the Renovate configuration.
	areaQuiet
	// areaCode is any other file: Go source, or data a Go package embeds or its tests read.
	areaCode
)

// rule maps a path prefix (ending in "/") or an exact path to an area, with the reason the plan prints.
type rule struct {
	prefix string
	area   area
	why    string
}

// rules lists the path rules in the order they are tried; the first match wins.
func rules() []rule {
	full := func(prefix, why string) rule { return rule{prefix: prefix, area: areaFull, why: why} }
	return []rule{
		full("go.mod", "module dependencies affect every package"),
		full("go.sum", "module dependencies affect every package"),
		full("go.work", "the workspace affects every package"),
		full("go.work.sum", "the workspace affects every package"),
		full("sdk/go/go.mod", "SDK dependencies affect every package that imports the SDK"),
		full("sdk/go/go.sum", "SDK dependencies affect every package that imports the SDK"),
		full(".golangci.yml", "lint configuration applies to every package"),
		full("Taskfile.yml", "the task targets define every check"),
		full("lefthook.yml", "the git hooks define the local checks"),
		full("tools/", "the docs generator and this selector run in the checks"),
		full("scripts/", "scripts implement the checks and their tests"),
		full("proto/", "contracts generate code in sdk/go/gen and gen/"),
		full("buf.yaml", "contract lint and generation settings"),
		full("buf.lock", "contract dependencies"),
		full("buf.gen.yaml", "contract code generation settings"),
		full("gen/", "generated code is checked against its contracts"),
		full("sdk/go/gen/", "generated code is checked against its contracts"),
		full("design/", "design assets are embedded in the UI and mounted into the docs site"),
		full("package.json", "npm packages: kibi, Mermaid, playwright-core"),
		full("package-lock.json", "npm packages: kibi, Mermaid, playwright-core"),
		{prefix: "docs/", area: areaDocs, why: "documentation"},
		{prefix: "site/", area: areaDocs, why: "the docs site"},
		{prefix: ".kb/", area: areaKibi, why: "requirements knowledge base"},
		{prefix: ".github/", area: areaWorkflows, why: "GitHub Actions configuration"},
		{prefix: "deploy/", area: areaDeploy, why: "Kubernetes manifests"},
		{prefix: ".claude/", area: areaQuiet, why: "agent instructions, read by no check"},
		{prefix: ".mcp.json", area: areaQuiet, why: "agent configuration, read by no check"},
		{prefix: "renovate.json", area: areaQuiet, why: "Renovate configuration, read by no check"},
	}
}

// classify returns the rule a path falls under. Root-level Markdown and licence files are documentation; every
// other path under a directory is code, which the graph then resolves; any other root file is unknown.
func classify(file string) rule {
	if strings.ContainsAny(file, " \t\n") {
		return rule{area: areaUnknown, why: "a path with whitespace, which the task variables can't carry"}
	}
	for _, r := range rules() {
		if file == r.prefix || (strings.HasSuffix(r.prefix, "/") && strings.HasPrefix(file, r.prefix)) {
			return r
		}
	}
	if !strings.Contains(file, "/") {
		if path.Ext(file) == ".md" || strings.HasPrefix(file, "LICENSE") {
			return rule{area: areaDocs, why: "documentation"}
		}
		return rule{area: areaUnknown, why: "no rule names this file"}
	}
	return rule{area: areaCode}
}

// docsOnly reports whether a path can't change any Go package, image or manifest: CI then skips the
// end-to-end and PostgreSQL jobs for a pull request made only of such paths.
func docsOnly(file string) bool {
	a := classify(file).area
	return a == areaDocs || a == areaKibi || a == areaQuiet
}
