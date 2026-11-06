package archtest

import (
	"go/ast"
	"strconv"
)

// StableTests checks that no test sleeps (time.Sleep) or skips (Skip, Skipf
// or SkipNow on a *testing.T, B or F, or a testing.TB). The few that wait on
// something outside the process, or need a tool a contributor may lack, are
// listed in the exceptions with their reasons.
func StableTests(r *Repo) []Violation {
	var out []Violation
	for _, f := range r.Tests {
		timePkg := importName(f.Syntax, "time")
		for _, d := range f.Syntax.Decls {
			testers := testerNames(d)
			ast.Inspect(d, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if what := unstable(call, timePkg, testers); what != "" {
					key := f.Path + " " + declName(d) + " " + what
					out = append(out, Violation{RuleStableTests, key, what + " in a test: wait on a condition with a deadline, and never skip"})
				}
				return true
			})
		}
	}
	return dedupe(out)
}

// unstable names what a call does that a stable test must not, or "".
func unstable(call *ast.CallExpr, timePkg string, testers map[string]bool) string {
	sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	x, ok := sel.X.(*ast.Ident)
	switch {
	case !ok:
		return ""
	case x.Name == timePkg && sel.Sel.Name == "Sleep":
		return "time.Sleep"
	case testers[x.Name] && (sel.Sel.Name == "Skip" || sel.Sel.Name == "Skipf" || sel.Sel.Name == "SkipNow"):
		return "Skip"
	}
	return ""
}

// importName is the name a file uses for an import path, or "" when it
// does not import it.
func importName(f *ast.File, path string) string {
	for _, imp := range f.Imports {
		if p, err := strconv.Unquote(imp.Path.Value); err != nil || p != path {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name
		}
		return baseName(path)
	}
	return ""
}

// testerNames are the names a declaration gives its *testing.T, B and F
// and testing.TB parameters, closures' included.
func testerNames(d ast.Decl) map[string]bool {
	names := map[string]bool{}
	ast.Inspect(d, func(n ast.Node) bool {
		var ft *ast.FuncType
		switch n := n.(type) {
		case *ast.FuncDecl:
			ft = n.Type
		case *ast.FuncLit:
			ft = n.Type
		default:
			return true
		}
		for _, field := range ft.Params.List {
			if isTester(field.Type) {
				for _, name := range field.Names {
					names[name.Name] = true
				}
			}
		}
		return true
	})
	return names
}

// isTester says whether a parameter type is *testing.T, B or F, or
// testing.TB.
func isTester(e ast.Expr) bool {
	if star, ok := e.(*ast.StarExpr); ok {
		e = star.X
	}
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "testing" && (sel.Sel.Name == "T" || sel.Sel.Name == "B" || sel.Sel.Name == "F" || sel.Sel.Name == "TB")
}
