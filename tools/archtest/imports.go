package archtest

import (
	"path"
	"slices"
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
	case slices.Contains(datasets(), to.Context) && !under(from, "cmd"):
		return "only cmd imports the " + to.Context + " dataset"
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
	case to.Layer != LayerAdapter || under(from, "cmd") || fp.Adapter == to.Adapter:
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
