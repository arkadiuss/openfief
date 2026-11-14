package archtest

import (
	"path"
	"slices"
	"strconv"
	"strings"
)

// edge is one import between two repository packages.
func edge(from, to string) string {
	return from + " -> " + to
}

// ContextMap checks that no context imports another the context map doesn't
// name, that the kernel imports no context, that every package under
// internal/ is in a context or a listed dataset, and that a dataset reads
// only domains and is imported only by cmd.
func ContextMap(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		from := PlaceOf(p.Path)
		if from.Context != "" && !known(from.Context) {
			out = append(out, Violation{RuleContextMap, p.Path, "internal/" + from.Context + " is neither a context in the table nor a dataset"})
		}
		for _, imp := range p.Imports {
			if !isRepo(imp) {
				continue
			}
			if detail := contextMapBreak(p.Path, from, imp); detail != "" {
				out = append(out, Violation{RuleContextMap, edge(p.Path, imp), detail})
			}
		}
	}
	return out
}

// known says whether a name under internal/ is a context, the kernel or a
// dataset.
func known(ctx string) bool {
	_, ok := contextMap()[ctx]
	return ok || ctx == ctxKernel || slices.Contains(datasets(), ctx)
}

// contextMapBreak says what is wrong with one import, or "".
func contextMapBreak(from string, fp Place, imp string) string {
	to := PlaceOf(imp)
	if slices.Contains(datasets(), to.Context) || slices.Contains(datasets(), fp.Context) {
		return datasetBreak(from, fp, to)
	}
	switch {
	case to.Context == "" || fp.Context == "" || !known(to.Context) || !known(fp.Context):
		return ""
	case fp.Layer == LayerKernel && to.Layer != LayerKernel:
		return "the kernel imports no context"
	case !mayImport(fp.Context, to.Context):
		return "the context map doesn't let " + fp.Context + " import " + to.Context
	}
	return ""
}

// datasetBreak says what is wrong with an import from or of a dataset, or
// "": a dataset reads only domains, and only cmd imports one.
func datasetBreak(from string, fp, to Place) string {
	switch {
	case fp.Context == to.Context:
		return ""
	case slices.Contains(datasets(), to.Context) && !under(from, "cmd") && !isBench(from):
		return "only cmd and the benches import the " + to.Context + " dataset"
	case slices.Contains(datasets(), fp.Context) && to.Layer != LayerDomain && to.Layer != LayerKernel:
		return "the " + fp.Context + " dataset reads only domain packages"
	}
	return ""
}

// Layering checks the layers of docs/dev/code-organization.md#rules: a domain
// package imports only domains and the kernel, an app package no adapter, an
// adapter no other adapter, and nothing outside internal/ but cmd an adapter.
// Plugins and the SDK are the Plugins rule's.
func Layering(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		from := PlaceOf(p.Path)
		if under(p.Path, "plugins") || under(p.Path, "sdk") {
			continue
		}
		for _, imp := range p.Imports {
			if !isRepo(imp) {
				continue
			}
			if detail := layeringBreak(p.Path, from, imp); detail != "" {
				out = append(out, Violation{RuleLayering, edge(p.Path, imp), detail})
			}
		}
	}
	return out
}

// layeringBreak says what is wrong with one import, or "".
func layeringBreak(from string, fp Place, imp string) string {
	to := PlaceOf(imp)
	switch {
	case under(imp, "cmd") && !under(from, "cmd"):
		return "only cmd imports cmd"
	case fp.Layer == LayerDomain && to.Layer != LayerDomain && to.Layer != LayerKernel:
		return "a domain package imports only the kernel and domain packages"
	case to.Layer != LayerAdapter || under(from, "cmd") || isBench(from) || fp.Adapter == to.Adapter:
		return ""
	}
	return adapterBreak(fp, to)
}

// adapterBreak says what is wrong with an import of an adapter from outside
// cmd and outside the adapter itself.
func adapterBreak(fp, to Place) string {
	switch {
	case fp.Layer == LayerApp:
		return "an app package never imports an adapter"
	case fp.Layer == LayerAdapter && fp.Context != to.Context:
		return "no context imports another context's adapters"
	}
	return "only cmd imports adapters"
}

// Plugins checks that nothing under plugins/ or sdk/ imports internal/ or
// cmd/: a plugin builds on the SDK alone.
func Plugins(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		if !under(p.Path, "plugins") && !under(p.Path, "sdk") {
			continue
		}
		for _, imp := range p.Imports {
			if under(imp, "internal") || under(imp, "cmd") {
				out = append(out, Violation{RulePlugins, edge(p.Path, imp), "a plugin or the SDK imports only the SDK, never internal/ or cmd/"})
			}
		}
	}
	return out
}

// reachesNetwork reports whether the package imp is an HTTP, Connect, gRPC (and
// so SpiceDB's) or Kubernetes client: one that reaches the network.
func reachesNetwork(imp string) bool {
	if imp == "net/http" || imp == "net/rpc" {
		return true
	}
	for _, prefix := range []string{"connectrpc.com/", "google.golang.org/grpc", "github.com/authzed/", "k8s.io/"} {
		if strings.HasPrefix(imp, prefix) {
			return true
		}
	}
	return false
}

// NoNetworkInTransactions checks that no PostgreSQL adapter, the code that
// runs in a change's transaction, imports a client that reaches the network:
// the database's own driver is all it talks to. A change that needs an answer
// from a connector, a model provider, SpiceDB or the cluster asks for it
// before its transaction opens, in a use case, or defers it until the
// transaction has committed (D-211).
func NoNetworkInTransactions(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		if !strings.HasPrefix(p.Path, "internal/") || !slices.Contains(strings.Split(p.Path, "/"), "postgres") {
			continue
		}
		for _, imp := range p.Imports {
			if reachesNetwork(imp) {
				out = append(out, Violation{RuleNoNetwork, edge(p.Path, imp), "a PostgreSQL adapter runs in a change's transaction and makes no network call: ask before the transaction opens, or defer until it commits"})
			}
		}
	}
	return out
}

// Bench checks that only the Ask box bench imports goeval (D-218), and only
// the packages of it that depend on the standard library alone: the root
// package and the Gemini, judge and embedding ones would bring the Gemini and
// Google Cloud client libraries into the module's builds.
func Bench(r *Repo) []Violation {
	var out []Violation
	check := func(key, from string, imports []string) {
		for _, imp := range imports {
			switch {
			case !under(imp, goevalModule):
			case !isBench(from):
				out = append(out, Violation{RuleBench, key + " -> " + imp, "only the Ask box bench imports goeval"})
			case !slices.Contains(lightGoeval(), imp):
				out = append(out, Violation{RuleBench, key + " -> " + imp, "the bench imports only goeval/api, goeval/heuristic and goeval/openai, which need only the standard library"})
			}
		}
	}
	for _, p := range r.Packages {
		check(p.Path, p.Path, p.Imports)
	}
	for _, f := range r.Tests {
		var imports []string
		for _, spec := range f.Syntax.Imports {
			if imp, err := strconv.Unquote(spec.Path.Value); err == nil {
				imports = append(imports, imp)
			}
		}
		check(f.Path, path.Dir(f.Path), imports)
	}
	return dedupe(out)
}

// Banned checks that no package, test-only ones included, is named util,
// common, helpers, models or types.
func Banned(r *Repo) []Violation {
	banned := []string{"util", "common", "helpers", "models", "types"}
	seen := map[string]bool{}
	var out []Violation
	check := func(dir, name string) {
		name = strings.TrimSuffix(name, "_test")
		if seen[dir] || (!slices.Contains(banned, name) && !slices.Contains(banned, path.Base(dir))) {
			return
		}
		seen[dir] = true
		out = append(out, Violation{RuleBanned, dir, "a package named " + name + " says nothing: find the context the code belongs to"})
	}
	for _, p := range r.Packages {
		check(p.Path, p.Name)
	}
	for _, f := range r.Tests {
		check(path.Dir(f.Path), f.Syntax.Name.Name)
	}
	return out
}
