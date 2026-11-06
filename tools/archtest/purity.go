package archtest

import (
	"go/ast"
	"go/token"
)

// PureDomain checks what depguard and forbidigo cannot see in domain and
// kernel packages: no go statement, and no package-level variable other than
// an error (a sentinel) or the blank identifier.
func PureDomain(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		if l := PlaceOf(p.Path).Layer; l != LayerDomain && l != LayerKernel {
			continue
		}
		for _, f := range p.Files {
			for _, d := range f.Decls {
				out = append(out, goStatements(p.Path, d)...)
				out = append(out, packageVars(p, d)...)
			}
		}
	}
	return dedupe(out)
}

// goStatements reports a declaration that starts a goroutine.
func goStatements(pkg string, d ast.Decl) []Violation {
	var out []Violation
	ast.Inspect(d, func(n ast.Node) bool {
		if _, ok := n.(*ast.GoStmt); ok && out == nil {
			key := pkg + " " + declName(d) + " go"
			out = append(out, Violation{RulePureDomain, key, "a go statement in a domain package: concurrency belongs in app or adapters"})
		}
		return out == nil
	})
	return out
}

// packageVars reports the package-level variables of a declaration that
// are not errors.
func packageVars(p *Package, d ast.Decl) []Violation {
	g, ok := d.(*ast.GenDecl)
	if !ok || g.Tok != token.VAR {
		return nil
	}
	var out []Violation
	for _, s := range g.Specs {
		vs, ok := s.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			obj := p.Info.Defs[name]
			if name.Name == "_" || obj == nil || isError(obj.Type()) {
				continue
			}
			key := p.Path + " var " + name.Name
			out = append(out, Violation{RulePureDomain, key, "a package-level variable in a domain package: take it as a parameter, or make it a constant or a function"})
		}
	}
	return out
}
