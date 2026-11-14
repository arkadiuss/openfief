package archtest

import (
	"fmt"
	"slices"
	"strings"
)

// The rules, by the name the baseline and the failure messages use. Each is
// stated in docs/dev/code-organization.md#architecture-tests under the title
// ruleTitle gives.
const (
	RuleContextMap   = "context-map"
	RuleLayering     = "layering"
	RulePlugins      = "plugins"
	RuleSchemas      = "schemas"
	RuleMoney        = "money"
	RulePureDomain   = "pure-domain"
	RulePrivacy      = "privacy"
	RuleBanned       = "banned-packages"
	RuleStableTests  = "stable-tests"
	RuleCatalogTests = "catalog-tests"
	RuleNoNetwork    = "no-network-in-transactions"
	RuleBench        = "bench"
)

// Violation is one break of a rule. Key identifies it stably across edits
// (a package, file, function or import edge, never a line), so the baseline
// can list it; Detail says what is wrong.
type Violation struct {
	Rule   string
	Key    string
	Detail string
}

// String is the failure message: the rule, the key, what is wrong, and the
// doc section that states the rule.
func (v Violation) String() string {
	return fmt.Sprintf("[%s] %s: %s (docs/dev/code-organization.md#architecture-tests, %q)", v.Rule, v.Key, v.Detail, ruleTitle(v.Rule))
}

// ruleTitle is the bold title of a rule's bullet in the doc.
func ruleTitle(rule string) string {
	titles := map[string]string{
		RuleContextMap:   "Context map",
		RuleLayering:     "Layering",
		RulePlugins:      "Plugins",
		RuleSchemas:      "Schemas",
		RuleMoney:        "Money",
		RulePureDomain:   "Pure domain",
		RulePrivacy:      "Privacy (GR-12)",
		RuleBanned:       "Banned packages",
		RuleStableTests:  "Stable tests",
		RuleCatalogTests: "Catalog-independent tests",
		RuleNoNetwork:    "No network in transactions",
		RuleBench:        "Ask box bench",
	}
	if title, ok := titles[rule]; ok {
		return title
	}
	return rule
}

// The contexts and the kernel, by their directory under internal/.
const (
	ctxKernel       = "kernel"
	ctxLedger       = "ledger"
	ctxValuation    = "valuation"
	ctxIngest       = "ingest"
	ctxViews        = "views"
	ctxPlugins      = "plugins"
	ctxIdentity     = "identity"
	ctxAttention    = "attention"
	ctxPortfolio    = "portfolio"
	ctxModelRouter  = "modelrouter"
	ctxJurisdiction = "jurisdiction"
)

// Layer is where a package sits in a bounded context.
type Layer int

// The layers of docs/dev/code-organization.md#layout-p.
const (
	// LayerOutside is a package outside internal/: cmd, plugins, sdk, tools.
	LayerOutside Layer = iota
	// LayerKernel is the shared kernel, internal/kernel.
	LayerKernel
	// LayerDomain is internal/<context>, and internal/jurisdiction/<pack>.
	LayerDomain
	// LayerApp is internal/<context>/app and below.
	LayerApp
	// LayerAdapter is any other package below internal/<context>.
	LayerAdapter
)

// Place is where a repository package sits: its context (empty outside
// internal/), its layer, and for an adapter the adapter's root package, so
// internal/plugins/cosign/cosigntest belongs to internal/plugins/cosign.
type Place struct {
	Context string
	Layer   Layer
	Adapter string
}

// domainDepth is the number of path elements of a domain package:
// internal/<context>.
const domainDepth = 2

// PlaceOf classifies a repository path.
func PlaceOf(path string) Place {
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "internal" {
		return Place{Layer: LayerOutside}
	}
	ctx := parts[1]
	switch {
	case ctx == ctxKernel:
		return Place{Context: ctx, Layer: LayerKernel}
	case len(parts) == domainDepth, ctx == ctxJurisdiction && len(parts) == domainDepth+1 && parts[2] != "app":
		return Place{Context: ctx, Layer: LayerDomain}
	case parts[2] == "app":
		return Place{Context: ctx, Layer: LayerApp}
	}
	return Place{Context: ctx, Layer: LayerAdapter, Adapter: strings.Join(parts[:3], "/")}
}

// contextMap is the context map of docs/dev/code-organization.md#bounded-contexts:
// for each context, the other contexts it may import. Every context may
// import the kernel.
func contextMap() map[string][]string {
	return map[string][]string{
		ctxLedger:       nil,
		ctxPlugins:      nil,
		ctxIdentity:     nil,
		ctxJurisdiction: nil,
		ctxModelRouter:  nil,
		ctxValuation:    {ctxLedger, ctxJurisdiction},
		ctxIngest:       {ctxLedger, ctxValuation, ctxPlugins, ctxJurisdiction},
		ctxPortfolio:    {ctxLedger, ctxValuation},
		ctxAttention:    {ctxLedger, ctxValuation, ctxIngest, ctxJurisdiction},
		ctxViews:        {ctxLedger, ctxValuation, ctxPortfolio, ctxAttention, ctxIngest, ctxPlugins, ctxIdentity, ctxJurisdiction},
	}
}

// datasets are the packages under internal/ that are not contexts: the demo
// dataset (docs/dev/demo.md), a pure generator that may read every context's
// domain and that only cmd and tests import.
func datasets() []string {
	return []string{"demo"}
}

// benches are the harnesses under tools/ that drive the product's own code
// and its demo dataset, to grade it (docs/dev/code-organization.md, "Support
// packages"): the Ask box bench. Like cmd they may import adapters and the
// datasets, and nothing imports them.
func benches() []string {
	return []string{"tools/askbench"}
}

// isBench says whether a repository path is a bench or below one.
func isBench(path string) bool {
	return slices.ContainsFunc(benches(), func(b string) bool { return under(path, b) })
}

// goevalModule is the module of the eval scorer library (D-218). Only the
// benches import it, and only its packages that depend on the standard library
// alone.
const goevalModule = "github.com/datar-psa/goeval"

// lightGoeval are goeval's packages that depend on the standard library alone:
// the scorer interface, the tool-call scorer and the OpenAI-compatible judge.
// Its other packages bring the Gemini and Google Cloud client libraries.
func lightGoeval() []string {
	return []string{goevalModule + "/api", goevalModule + "/heuristic", goevalModule + "/openai"}
}

// mayImport says whether context from may import context to.
func mayImport(from, to string) bool {
	return from == to || to == ctxKernel || slices.Contains(contextMap()[from], to)
}

// isRepo says whether an import path is one of the repository's own packages
// (made relative by Load), not the standard library or a dependency: it
// starts with one of the repository's top-level Go directories.
func isRepo(path string) bool {
	top, _, _ := strings.Cut(path, "/")
	return slices.Contains([]string{"internal", "cmd", "plugins", "sdk", "tools", "design", "authz", "gen", "test"}, top)
}

// under says whether path is dir or lies below it.
func under(path, dir string) bool {
	return path == dir || strings.HasPrefix(path, dir+"/")
}
