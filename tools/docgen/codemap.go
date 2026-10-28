package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// goPackage is the part of `go list -json` output docgen reads.
type goPackage struct {
	importPath  string
	doc         string
	goFiles     []string
	testGoFiles []string
	imports     []string
}

// place is where a package sits in the layout of docs/dev/code-organization.md.
type place struct {
	context string // internal/<context>, jurisdiction/<country>, cmd or tools
	layer   string // domain, app, adapter, binary or tool
}

const (
	layerDomain  = "domain"
	layerApp     = "app"
	layerAdapter = "adapter"
	layerBinary  = "binary"
	layerTool    = "tool"
)

// decodePackages reads the stream of JSON objects `go list -json ./...` prints. go list names its keys
// like Go fields (ImportPath), so each object is read key by key rather than into a tagged struct.
func decodePackages(r io.Reader) ([]goPackage, error) {
	var pkgs []goPackage
	dec := json.NewDecoder(r)
	for {
		var raw map[string]json.RawMessage
		err := dec.Decode(&raw)
		if errors.Is(err, io.EOF) {
			return pkgs, nil
		}
		if err != nil {
			return nil, fmt.Errorf("decode go list output: %w", err)
		}
		p, err := packageFromJSON(raw)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, p)
	}
}

func packageFromJSON(raw map[string]json.RawMessage) (goPackage, error) {
	var p goPackage
	fields := map[string]any{"ImportPath": &p.importPath, "Doc": &p.doc, "GoFiles": &p.goFiles, "TestGoFiles": &p.testGoFiles, "Imports": &p.imports}
	for key, dst := range fields {
		value, ok := raw[key]
		if !ok {
			continue
		}
		if err := json.Unmarshal(value, dst); err != nil {
			return p, fmt.Errorf("decode go list %s: %w", key, err)
		}
	}
	return p, nil
}

// placeOf classifies a module-relative package path, such as internal/ledger/app.
func placeOf(rel string) place {
	parts := strings.Split(rel, "/")
	switch {
	case parts[0] == "cmd":
		return place{context: "cmd", layer: layerBinary}
	case parts[0] != "internal" || len(parts) < 2:
		return place{context: parts[0], layer: layerTool}
	}
	ctxLen := 2
	if parts[1] == "jurisdiction" && len(parts) > 2 {
		ctxLen = 3
	}
	ctx := strings.Join(parts[1:min(ctxLen, len(parts))], "/")
	switch {
	case len(parts) <= ctxLen:
		return place{context: ctx, layer: layerDomain}
	case parts[ctxLen] == "app":
		return place{context: ctx, layer: layerApp}
	default:
		return place{context: ctx, layer: layerAdapter}
	}
}

// layeringFindings applies code-organization.md's layering rules to every import inside the module.
func layeringFindings(module string, pkgs []goPackage) []string {
	var findings []string
	for _, p := range pkgs {
		from := placeOf(relPath(module, p.importPath))
		for _, imp := range p.imports {
			if !strings.HasPrefix(imp, module+"/") {
				continue
			}
			if why := layeringViolation(from, placeOf(relPath(module, imp))); why != "" {
				findings = append(findings, fmt.Sprintf("%#q imports %#q: %s", relPath(module, p.importPath), relPath(module, imp), why))
			}
		}
	}
	slices.Sort(findings)
	return findings
}

func layeringViolation(from, to place) string {
	switch {
	case from.layer == layerBinary || from.layer == layerTool:
		return ""
	case from.layer == layerDomain && to.layer != layerDomain:
		return "a domain package imports " + article(to.layer) + " package"
	case from.layer == layerApp && to.layer == layerAdapter:
		return "an app package imports an adapter"
	case from.context != to.context && to.layer != layerDomain:
		return "a context imports another context's " + to.layer + " package"
	default:
		return ""
	}
}

// article prefixes a layer name with "a" or "an".
func article(layer string) string {
	if strings.ContainsAny(layer[:1], "aeiou") {
		return "an " + layer
	}
	return "a " + layer
}

func relPath(module, importPath string) string {
	return strings.TrimPrefix(strings.TrimPrefix(importPath, module), "/")
}

// contextEdges lists which contexts import which, from every package's imports inside the module.
func contextEdges(module string, pkgs []goPackage) map[[2]string]bool {
	edges := map[[2]string]bool{}
	for _, p := range pkgs {
		from := placeOf(relPath(module, p.importPath))
		for _, imp := range p.imports {
			if !strings.HasPrefix(imp, module+"/") {
				continue
			}
			to := placeOf(relPath(module, imp))
			if from.context != to.context && from.layer != layerTool {
				edges[[2]string{from.context, to.context}] = true
			}
		}
	}
	return edges
}

// codeMapPage renders the as-built code map: a context diagram, the layering check and every package.
func codeMapPage(module string, pkgs []goPackage) string {
	var b strings.Builder
	b.WriteString("---\ntitle: Code map\ndescription: Every Go package as built, grouped by bounded context, with the dependencies between contexts drawn from the imports.\nweight: 10\n---\n\n")
	b.WriteString("Generated from `go list` on every build. Compare the diagram with the designed context map in [code-organization.md](../../../../docs/dev/code-organization.md): an arrow there that is missing here is not built yet, and an arrow here that is missing there needs a decision.\n\n")
	b.WriteString("## Contexts\n\nAn arrow means at least one package of the first context imports a package of the second. Arrows into the shared kernel are dashed.\n\n")
	b.WriteString(contextDiagram(contextEdges(module, pkgs)))
	b.WriteString("\n## Layering check\n\nThe rules of code-organization.md: domain packages import only domain packages, app packages never import adapters, and no context imports another context's app or adapters. Only `cmd` wires adapters together.\n\n")
	if findings := layeringFindings(module, pkgs); len(findings) > 0 {
		for _, f := range findings {
			b.WriteString("- " + f + "\n")
		}
	} else {
		b.WriteString("No violations.\n")
	}
	b.WriteString("\n## Packages\n\n")
	writePackageTables(&b, module, pkgs)
	return b.String()
}

func contextDiagram(edges map[[2]string]bool) string {
	var b strings.Builder
	b.WriteString("```mermaid\nflowchart LR\n")
	nodes := map[string]bool{}
	for e := range edges {
		nodes[e[0]], nodes[e[1]] = true, true
	}
	for _, n := range slices.Sorted(maps.Keys(nodes)) {
		fmt.Fprintf(&b, "  %s[\"%s\"]\n", nodeID(n), n)
	}
	keys := slices.SortedFunc(maps.Keys(edges), func(a, c [2]string) int { return strings.Compare(a[0]+" "+a[1], c[0]+" "+c[1]) })
	for _, e := range keys {
		arrow := "-->"
		if e[1] == "kernel" {
			arrow = "-.->"
		}
		fmt.Fprintf(&b, "  %s %s %s\n", nodeID(e[0]), arrow, nodeID(e[1]))
	}
	b.WriteString("```\n")
	return b.String()
}

func nodeID(name string) string {
	return "c_" + strings.NewReplacer("/", "_", "-", "_", ".", "_").Replace(name)
}

func writePackageTables(b *strings.Builder, module string, pkgs []goPackage) {
	byContext := map[string][]goPackage{}
	for _, p := range pkgs {
		ctx := placeOf(relPath(module, p.importPath)).context
		byContext[ctx] = append(byContext[ctx], p)
	}
	for _, ctx := range slices.Sorted(maps.Keys(byContext)) {
		fmt.Fprintf(b, "### %s\n\n| Package | Layer | Files | Test files | Summary |\n|---|---|---|---|---|\n", ctx)
		for _, p := range byContext[ctx] {
			rel := relPath(module, p.importPath)
			fmt.Fprintf(b, "| [`%s`](../../../../%s/) | %s | %d | %d | %s |\n",
				rel, rel, placeOf(rel).layer, len(p.goFiles), len(p.testGoFiles), firstSentence(p.doc))
		}
		b.WriteString("\n")
	}
}

// firstSentence returns a package comment's first sentence, flattened onto one table-safe line.
func firstSentence(doc string) string {
	doc = strings.Join(strings.Fields(doc), " ")
	if i := strings.Index(doc, ". "); i >= 0 {
		doc = doc[:i+1]
	}
	return tableCell(doc)
}

// tableCell makes text safe in a Markdown table cell: pipes are escaped, and angle brackets outside code
// spans, such as CONFIRM=<ENV>, become entities instead of raw HTML.
func tableCell(text string) string {
	parts := strings.Split(text, "`")
	for i := 0; i < len(parts); i += 2 {
		parts[i] = strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(parts[i])
	}
	return strings.ReplaceAll(strings.Join(parts, "`"), "|", "\\|")
}
