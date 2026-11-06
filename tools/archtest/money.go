package archtest

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"
)

// Money checks that no domain, app or kernel package uses float32 or float64,
// and that no contract under proto/ has a float or double field. A package
// that only draws, such as chart coordinates, is listed in the exceptions.
func Money(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		if l := PlaceOf(p.Path).Layer; l == LayerDomain || l == LayerApp || l == LayerKernel {
			out = append(out, floats(p)...)
		}
	}
	field := regexp.MustCompile(`(?m)^\s*(?:repeated\s+|optional\s+)?(?:float|double|map<[^>]*,\s*(?:float|double)\s*>)\s+(\w+)\s*=`)
	for _, f := range r.Protos {
		for _, m := range field.FindAllStringSubmatch(f.Text, -1) {
			key := f.Path + " " + m[1]
			out = append(out, Violation{RuleMoney, key, "field " + m[1] + " is a float or double: contracts carry amounts as decimal strings"})
		}
	}
	return dedupe(out)
}

// floats reports the declarations of a package that use a float.
func floats(p *Package) []Violation {
	var out []Violation
	for _, f := range p.Files {
		for _, d := range f.Decls {
			if usesFloat(p.Info, d) {
				key := p.Path + " " + declName(d)
				out = append(out, Violation{RuleMoney, key, "float32 or float64 in a domain, app or kernel package: use the kernel's decimal types"})
			}
		}
	}
	return out
}

// usesFloat says whether any expression in a declaration has a float type.
func usesFloat(info *types.Info, d ast.Decl) bool {
	found := false
	ast.Inspect(d, func(n ast.Node) bool {
		if e, ok := n.(ast.Expr); ok && !found {
			found = isFloat(info.TypeOf(e))
		}
		return !found
	})
	return found
}

// isFloat says whether t is float32 or float64, or a named type over one.
func isFloat(t types.Type) bool {
	if t == nil {
		return false
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsFloat != 0 && b.Kind() != types.UntypedFloat
}

// declName names a top-level declaration: a function, a method as
// Type.Method, or a type, var or const by its first name.
func declName(d ast.Decl) string {
	switch d := d.(type) {
	case *ast.FuncDecl:
		if d.Recv != nil && len(d.Recv.List) > 0 {
			return recvName(d.Recv.List[0].Type) + "." + d.Name.Name
		}
		return d.Name.Name
	case *ast.GenDecl:
		return genDeclName(d)
	}
	return "?"
}

// genDeclName names a type, var or const declaration by its first name.
func genDeclName(d *ast.GenDecl) string {
	for _, s := range d.Specs {
		switch s := s.(type) {
		case *ast.TypeSpec:
			return s.Name.Name
		case *ast.ValueSpec:
			return s.Names[0].Name
		}
	}
	return strings.ToLower(d.Tok.String())
}

// recvName is a receiver's type name, without pointer or type parameters.
func recvName(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.StarExpr:
		return recvName(e.X)
	case *ast.IndexExpr:
		return recvName(e.X)
	case *ast.IndexListExpr:
		return recvName(e.X)
	case *ast.Ident:
		return e.Name
	}
	return "?"
}
