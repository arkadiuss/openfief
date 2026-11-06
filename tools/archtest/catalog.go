package archtest

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"
)

// catalogPath is the embedded plugin catalog's package (D-43), relative to
// the root module.
const catalogPath = "internal/plugins/catalog"

// CatalogTests checks that no test pins the embedded plugin catalog, which
// every release extends (D-88): a test that reads catalog.Load() never
// compares the number of entries with a literal or indexes an entry by a
// literal position; no test spells out a digest the catalog pins; and a test
// that calls a function of its package that reads the catalog derives its
// expectations from catalog.Load() itself, directly or through a helper.
// digests are the embedded catalog's image digests.
func CatalogTests(r *Repo, digests []string) []Violation {
	var out []Violation
	for _, f := range r.Tests {
		name := importName(f.Syntax, r.Module+"/"+catalogPath)
		for _, d := range f.Syntax.Decls {
			where := f.Path + " " + declName(d)
			for _, what := range pins(d, name) {
				out = append(out, Violation{RuleCatalogTests, where + " " + what, what + ": the catalog grows with every release, so derive the expectation from catalog.Load()"})
			}
			if pinsDigest(d, digests) {
				out = append(out, Violation{RuleCatalogTests, where + " digest", "spells out a digest the embedded catalog pins: read it from catalog.Load()"})
			}
		}
	}
	out = append(out, indirectReads(r)...)
	return dedupe(out)
}

// pins lists how a declaration pins the result of catalogName.Load().
func pins(d ast.Decl, catalogName string) []string {
	if catalogName == "" {
		return nil
	}
	loaded := loadedCatalogs(d, catalogName)
	var out []string
	ast.Inspect(d, func(n ast.Node) bool {
		if what := pin(n, loaded); what != "" {
			out = append(out, what)
		}
		return true
	})
	return out
}

// pin says how one node pins a loaded catalog, or "".
func pin(n ast.Node, loaded map[string]bool) string {
	switch n := n.(type) {
	case *ast.BinaryExpr:
		if isComparison(n.Op) && (lenOfEntries(n.X, loaded) && isInt(n.Y) || lenOfEntries(n.Y, loaded) && isInt(n.X)) {
			return "compares the number of catalog entries with a literal"
		}
	case *ast.IndexExpr:
		if entriesOf(n.X, loaded) && isInt(n.Index) {
			return "picks a catalog entry by a literal index"
		}
	}
	return ""
}

// loadedCatalogs names the variables a declaration assigns
// catalogName.Load()'s catalog to.
func loadedCatalogs(d ast.Decl, catalogName string) map[string]bool {
	loaded := map[string]bool{}
	ast.Inspect(d, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Rhs) != 1 || !isCall(as.Rhs[0], catalogName, "Load") {
			return true
		}
		if id, ok := as.Lhs[0].(*ast.Ident); ok {
			loaded[id.Name] = true
		}
		return true
	})
	return loaded
}

// isCall says whether e calls pkg.fn.
func isCall(e ast.Expr, pkg, fn string) bool {
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	return ok && x.Name == pkg && sel.Sel.Name == fn
}

// isComparison says whether op compares.
func isComparison(op token.Token) bool {
	return slices.Contains([]token.Token{token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ}, op)
}

// isInt says whether e is an integer literal.
func isInt(e ast.Expr) bool {
	lit, ok := ast.Unparen(e).(*ast.BasicLit)
	return ok && lit.Kind == token.INT
}

// entriesOf says whether e is v.Entries for a loaded catalog v.
func entriesOf(e ast.Expr, loaded map[string]bool) bool {
	sel, ok := ast.Unparen(e).(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Entries" {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	return ok && loaded[x.Name]
}

// lenOfEntries says whether e is len(v.Entries) for a loaded catalog v.
func lenOfEntries(e ast.Expr, loaded map[string]bool) bool {
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	fn, ok := call.Fun.(*ast.Ident)
	return ok && fn.Name == "len" && entriesOf(call.Args[0], loaded)
}

// pinsDigest says whether a declaration has a string literal holding one of
// the catalog's digests.
func pinsDigest(d ast.Decl, digests []string) bool {
	found := false
	ast.Inspect(d, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING || found {
			return !found
		}
		for _, dg := range digests {
			if _, hex, _ := strings.Cut(dg, ":"); hex != "" && strings.Contains(lit.Value, hex) {
				found = true
			}
		}
		return !found
	})
	return found
}

// indirectReads reports Test functions that call a function of their own
// package that reads the embedded catalog, without reading it themselves.
func indirectReads(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		readers := closure(p.Files, r.Module)
		if len(readers) == 0 {
			continue
		}
		var tests []*ast.File
		var paths []string
		for _, f := range r.Tests {
			if dirOf(f.Path) == p.Path && f.Syntax.Name.Name == p.Name {
				tests = append(tests, f.Syntax)
				paths = append(paths, f.Path)
			}
		}
		derived := closure(tests, r.Module)
		for i, f := range tests {
			out = append(out, indirectTests(paths[i], f, readers, derived)...)
		}
	}
	return out
}

// indirectTests checks the Test functions of one file: a test that inspects
// what a catalog reader returns, beyond its error, must derive from the
// catalog too.
func indirectTests(path string, f *ast.File, readers, derived map[string]bool) []Violation {
	var out []Violation
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(fd.Name.Name, "Test") || derived[fd.Name.Name] {
			continue
		}
		for _, name := range inspectedCalls(fd, readers) {
			where := path + " " + fd.Name.Name
			out = append(out, Violation{RuleCatalogTests, where + " calls " + name, "inspects what " + name + " returns, which reads the embedded catalog, but derives nothing from catalog.Load()"})
		}
	}
	return out
}

// inspectedCalls names the calls to readers whose results a function
// keeps: every call but a statement of its own and an assignment of its
// results to blanks and a last error.
func inspectedCalls(fd *ast.FuncDecl, readers map[string]bool) []string {
	ignored := ignoredCalls(fd)
	var out []string
	ast.Inspect(fd, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || ignored[call] {
			return true
		}
		if name := calleeName(call); readers[name] && !slices.Contains(out, name) {
			out = append(out, name)
		}
		return true
	})
	return out
}

// ignoredCalls are the calls whose results a function drops: a call that is
// a statement of its own, and one assigned only to blanks and an error.
func ignoredCalls(fd *ast.FuncDecl) map[*ast.CallExpr]bool {
	ignored := map[*ast.CallExpr]bool{}
	ast.Inspect(fd, func(n ast.Node) bool {
		var e ast.Expr
		switch n := n.(type) {
		case *ast.ExprStmt:
			e = n.X
		case *ast.AssignStmt:
			if len(n.Rhs) == 1 && blanksButLast(n.Lhs) {
				e = n.Rhs[0]
			}
		}
		if call, ok := ast.Unparen(e).(*ast.CallExpr); ok && e != nil {
			ignored[call] = true
		}
		return true
	})
	return ignored
}

// blanksButLast says whether every assigned name but the last, the error,
// is blank; a lone name must be err.
func blanksButLast(lhs []ast.Expr) bool {
	for _, e := range lhs[:len(lhs)-1] {
		if id, ok := e.(*ast.Ident); !ok || id.Name != "_" {
			return false
		}
	}
	last, ok := lhs[len(lhs)-1].(*ast.Ident)
	return len(lhs) > 1 || ok && last.Name == "err"
}

// calleeName is the name a call calls: f for f() and m for x.m().
func calleeName(call *ast.CallExpr) string {
	switch fn := ast.Unparen(call.Fun).(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fn.Sel.Name
	}
	return ""
}

// closure names the functions of files that call catalog.Load(), directly
// or through other functions of the same files, called by name.
func closure(files []*ast.File, module string) map[string]bool {
	calls := map[string]map[string]bool{}
	reads := map[string]bool{}
	for _, f := range files {
		name := importName(f, module+"/"+catalogPath)
		for fn, callees := range callGraph(f) {
			calls[fn] = callees
			if name != "" && callees[name+".Load"] {
				reads[fn] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for fn, callees := range calls {
			if !reads[fn] && callsAny(callees, reads) {
				reads[fn], changed = true, true
			}
		}
	}
	return reads
}

// callGraph maps each function of a file to the names it calls.
func callGraph(f *ast.File) map[string]map[string]bool {
	calls := map[string]map[string]bool{}
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			calls[fd.Name.Name] = calledNames(fd)
		}
	}
	return calls
}

// callsAny says whether any callee is in fns.
func callsAny(callees, fns map[string]bool) bool {
	for c := range callees {
		if fns[c] {
			return true
		}
	}
	return false
}

// calledNames are the names a function calls: f for f(), m for x.m(), and
// pkg.F as well for a package-qualified call.
func calledNames(fd *ast.FuncDecl) map[string]bool {
	names := map[string]bool{}
	ast.Inspect(fd, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := ast.Unparen(call.Fun).(type) {
		case *ast.Ident:
			names[fn.Name] = true
		case *ast.SelectorExpr:
			names[fn.Sel.Name] = true
			if x, ok := fn.X.(*ast.Ident); ok {
				names[x.Name+"."+fn.Sel.Name] = true
			}
		}
		return true
	})
	return names
}
