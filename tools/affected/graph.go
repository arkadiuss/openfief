package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// listedPackage is the part of `go list -json` the selector reads; paths are as go list prints them.
type listedPackage struct {
	dir             string
	importPath      string
	imports         []string
	testImports     []string
	xTestImports    []string
	goFiles         []string
	cgoFiles        []string
	testGoFiles     []string
	xTestGoFiles    []string
	embedFiles      []string
	testEmbedFiles  []string
	xTestEmbedFiles []string
}

// fields maps go list's JSON keys, which are Go field names, to where each value goes; like
// tools/docgen/codemap.go, objects are read key by key rather than into a tagged struct.
func (p *listedPackage) fields() map[string]any {
	return map[string]any{
		"Dir": &p.dir, "ImportPath": &p.importPath, "Imports": &p.imports, "TestImports": &p.testImports,
		"XTestImports": &p.xTestImports, "GoFiles": &p.goFiles, "CgoFiles": &p.cgoFiles, "TestGoFiles": &p.testGoFiles,
		"XTestGoFiles": &p.xTestGoFiles, "EmbedFiles": &p.embedFiles, "TestEmbedFiles": &p.testEmbedFiles,
		"XTestEmbedFiles": &p.xTestEmbedFiles,
	}
}

// pkg is one Go package of the repository's workspace. Paths are relative to the repository root.
type pkg struct {
	importPath string
	dir        string
	module     string
	imports    []string
	// testImports are what the package's tests import beyond the package itself.
	testImports []string
	// embeds are the files compiled into the package with go:embed; testEmbeds only into its tests.
	embeds     []string
	testEmbeds []string
	// testable is true when `go test` without build tags finds files in the package (test/e2e needs -tags e2e).
	testable bool
	// reads are paths outside the package directory that its code or tests name as relative string literals,
	// such as filepath.Join("..", "..", "deploy", "crds"): a change under one of them reruns its tests.
	reads []string
}

// graph is the import graph of every package in the workspace's modules.
type graph struct {
	pkgs map[string]*pkg
	// byDir maps a package directory to its import path.
	byDir map[string]string
	// importers and testImporters are the reverse edges of imports and testImports.
	importers     map[string][]string
	testImporters map[string][]string
	// embedOwners maps an embedded file to the packages that embed it, tests included.
	embedOwners map[string][]string
}

var errOutsideRepo = errors.New("package directory is outside the repository")

// decodeListed reads the concatenated JSON objects `go list -json` prints.
func decodeListed(r io.Reader) ([]listedPackage, error) {
	var out []listedPackage
	dec := json.NewDecoder(r)
	for {
		var raw map[string]json.RawMessage
		err := dec.Decode(&raw)
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("decode go list: %w", err)
		}
		var p listedPackage
		for key, dst := range p.fields() {
			if v, ok := raw[key]; ok {
				if err := json.Unmarshal(v, dst); err != nil {
					return nil, fmt.Errorf("decode go list %s: %w", key, err)
				}
			}
		}
		out = append(out, p)
	}
}

// relDir turns go list's absolute directory into a slash path relative to the repository root.
func relDir(root, dir string) (string, error) {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return "", fmt.Errorf("%s: %w", dir, err)
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("%w: %s", errOutsideRepo, dir)
	}
	return rel, nil
}

// newPkg builds a package from its tagged listing (imports, embeds) and whether the untagged listing has files.
func newPkg(root, module string, lp *listedPackage, testable bool) (*pkg, error) {
	dir, err := relDir(root, lp.dir)
	if err != nil {
		return nil, err
	}
	in := func(files []string) []string {
		out := make([]string, 0, len(files))
		for _, f := range files {
			out = append(out, path.Join(dir, f))
		}
		return out
	}
	testImports := slices.Concat(lp.testImports, lp.xTestImports)
	testImports = slices.DeleteFunc(testImports, func(ip string) bool { return ip == lp.importPath })
	return &pkg{
		importPath:  lp.importPath,
		dir:         dir,
		module:      module,
		imports:     lp.imports,
		testImports: testImports,
		embeds:      in(lp.embedFiles),
		testEmbeds:  in(slices.Concat(lp.testEmbedFiles, lp.xTestEmbedFiles)),
		testable:    testable,
	}, nil
}

// newGraph indexes packages and their reverse edges. Imports of packages outside the workspace are kept on
// the package but never followed.
func newGraph(pkgs []*pkg) *graph {
	g := &graph{
		pkgs:          map[string]*pkg{},
		byDir:         map[string]string{},
		importers:     map[string][]string{},
		testImporters: map[string][]string{},
		embedOwners:   map[string][]string{},
	}
	for _, p := range pkgs {
		g.pkgs[p.importPath] = p
		g.byDir[p.dir] = p.importPath
		for _, f := range slices.Concat(p.embeds, p.testEmbeds) {
			g.embedOwners[f] = append(g.embedOwners[f], p.importPath)
		}
	}
	for _, p := range pkgs {
		for _, ip := range p.imports {
			g.importers[ip] = append(g.importers[ip], p.importPath)
		}
		for _, ip := range p.testImports {
			g.testImporters[ip] = append(g.testImporters[ip], p.importPath)
		}
	}
	return g
}

// owningPackage is the package whose directory holds path, or the nearest enclosing one, for data files
// such as testdata. Files directly in the repository root belong to no package.
func (g *graph) owningPackage(file string) (string, bool) {
	for dir := path.Dir(file); dir != "." && dir != "/"; dir = path.Dir(dir) {
		if ip, ok := g.byDir[dir]; ok {
			return ip, true
		}
	}
	return "", false
}

// readers are the packages whose code or tests name a path that contains file.
func (g *graph) readers(file string) []string {
	var out []string
	for _, p := range g.pkgs {
		for _, r := range p.reads {
			if within(file, r) {
				out = append(out, p.importPath)
				break
			}
		}
	}
	slices.Sort(out)
	return out
}

// within reports whether file is dir itself or lies below it.
func within(file, dir string) bool {
	return dir == "." || file == dir || strings.HasPrefix(file, dir+"/")
}

// dependencies is the import closure of a package inside the workspace, the package included.
func (g *graph) dependencies(importPath string) map[string]bool {
	seen := map[string]bool{}
	queue := []string{importPath}
	for len(queue) > 0 {
		ip := queue[0]
		queue = queue[1:]
		p, ok := g.pkgs[ip]
		if !ok || seen[ip] {
			continue
		}
		seen[ip] = true
		queue = append(queue, p.imports...)
	}
	return seen
}

// relativeReads lists the repository paths a package's Go files name as relative string literals that leave
// the package directory: "../x" and filepath.Join("..", "x") with only literal arguments. Paths that would
// leave the repository are dropped.
func relativeReads(dir string, sources map[string][]byte) []string {
	var out []string
	add := func(rel string) {
		joined := path.Clean(path.Join(dir, rel))
		if joined == ".." || strings.HasPrefix(joined, "../") || within(joined, dir) {
			return
		}
		if !slices.Contains(out, joined) {
			out = append(out, joined)
		}
	}
	for name, src := range sources {
		file, err := parser.ParseFile(token.NewFileSet(), name, src, parser.SkipObjectResolution)
		if err != nil {
			continue // a file that doesn't parse fails the build or lint, which report it
		}
		ast.Inspect(file, func(n ast.Node) bool {
			for _, rel := range literalPaths(n) {
				add(rel)
			}
			return true
		})
	}
	slices.Sort(out)
	return out
}

// literalPaths returns the relative path a node spells out: a string literal starting with "../", or a
// *.Join call whose arguments are all string literals and whose first is "..".
func literalPaths(n ast.Node) []string {
	switch n := n.(type) {
	case *ast.BasicLit:
		if s, ok := stringLit(n); ok && strings.HasPrefix(s, "../") {
			return []string{s}
		}
	case *ast.CallExpr:
		return joinedPath(n)
	}
	return nil
}

// joinedPath is the path a *.Join call of string literals spells out, when its first argument is "..".
func joinedPath(call *ast.CallExpr) []string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Join" || len(call.Args) == 0 {
		return nil
	}
	parts := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		s, ok := stringLit(arg)
		if !ok {
			return nil
		}
		parts = append(parts, s)
	}
	if parts[0] != ".." {
		return nil
	}
	return []string{path.Join(parts...)}
}

// stringLit is the value of a string literal expression.
func stringLit(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}
