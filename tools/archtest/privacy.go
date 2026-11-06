package archtest

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"
)

// errorMethod is the name of error's method, and of the logging calls at
// error level.
const errorMethod = "Error"

// infoMethod is the name of the logging calls at info level.
const infoMethod = "Info"

// sensitiveTypes are the types whose values a log line never carries
// (GR-12), by repository path and name: amounts, quantities, prices, rates,
// ISINs and tokens.
func sensitiveTypes() map[string]string {
	return map[string]string{
		"internal/kernel.Money":         "money",
		"internal/kernel.Quantity":      "a quantity",
		"internal/kernel.Decimal":       "a decimal (a price, rate or factor)",
		"internal/ledger.ISIN":          "an ISIN",
		"internal/valuation.Price":      "a price",
		"internal/valuation.FXRate":     "an FX rate",
		"internal/identity.ScriptToken": "a script token",
	}
}

// loggers are the logging functions and methods, by package path: a call
// whose callee is one of these is a logging call.
func loggers() map[string][]string {
	slogNames := []string{
		"Debug", infoMethod, "Warn", errorMethod, "DebugContext", "InfoContext", "WarnContext", "ErrorContext", "Log", "LogAttrs",
	}
	return map[string][]string{
		"log/slog":                slogNames,
		"log":                     {"Print", "Printf", "Println", "Fatal", "Fatalf", "Fatalln", "Panic", "Panicf", "Panicln", "Output"},
		"github.com/go-logr/logr": {infoMethod, errorMethod},
		"github.com/sirupsen/logrus": {
			"Print", "Printf", "Println", "Debug", "Debugf", infoMethod, "Infof", "Warn", "Warnf", errorMethod, "Errorf",
			"Fatal", "Fatalf", "WithField", "WithFields", "WithError",
		},
	}
}

// Privacy checks that no logging call takes a value of a sensitive type, or
// an error (whose text may quote one), as an argument: an error is logged
// only through a LogSafe function, such as kernel.LogSafe. Attribute
// constructors (slog.String, slog.Any) and fmt.Sprint* are looked through,
// and so are .String() and .Error() calls.
func Privacy(r *Repo) []Violation {
	var out []Violation
	for _, p := range r.Packages {
		if under(p.Path, "sdk/go/gen") || !isRepo(p.Path) {
			continue
		}
		for _, f := range p.Files {
			for _, d := range f.Decls {
				out = append(out, logBreaks(r, p, d)...)
			}
		}
	}
	return dedupe(out)
}

// logBreaks checks the logging calls of one declaration.
func logBreaks(r *Repo, p *Package, d ast.Decl) []Violation {
	where := p.Path + " " + declName(d)
	var out []Violation
	ast.Inspect(d, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isLogging(p.Info, call) {
			return true
		}
		for _, arg := range call.Args {
			for _, what := range leaks(r, p.Info, arg) {
				out = append(out, Violation{RulePrivacy, where + " logs " + what, "a log line carries " + what + " (GR-12): log internal IDs, and errors through a LogSafe function"})
			}
		}
		return true
	})
	return out
}

// callee is the function or method a call names, or nil.
func callee(info *types.Info, call *ast.CallExpr) *types.Func {
	var id *ast.Ident
	switch fn := ast.Unparen(call.Fun).(type) {
	case *ast.Ident:
		id = fn
	case *ast.SelectorExpr:
		id = fn.Sel
	default:
		return nil
	}
	if f, ok := info.Uses[id].(*types.Func); ok {
		return f
	}
	return nil
}

// isLogging says whether a call is a logging call.
func isLogging(info *types.Info, call *ast.CallExpr) bool {
	f := callee(info, call)
	return f != nil && f.Pkg() != nil && slices.Contains(loggers()[f.Pkg().Path()], f.Name())
}

// leaks lists what an argument of a logging call carries that it must not.
func leaks(r *Repo, info *types.Info, e ast.Expr) []string {
	e = ast.Unparen(e)
	if call, ok := e.(*ast.CallExpr); ok {
		if inner, ok := lookThrough(info, call); ok {
			out := make([]string, 0, len(inner))
			for _, a := range inner {
				out = append(out, leaks(r, info, a)...)
			}
			return out
		}
	}
	if what := sensitive(r, info.TypeOf(e)); what != "" {
		return []string{what}
	}
	return nil
}

// lookThrough returns the expressions a call passes its values on from:
// none for a LogSafe call, the arguments of slog's attribute constructors
// and fmt.Sprint*, and the receiver of .String() and .Error(). ok is false
// for any other call, whose result is checked as it is.
func lookThrough(info *types.Info, call *ast.CallExpr) ([]ast.Expr, bool) {
	f := callee(info, call)
	switch {
	case f == nil:
		return nil, false
	case f.Name() == "LogSafe":
		return nil, true
	case isFormatter(f):
		return call.Args, true
	case (f.Name() == "String" || f.Name() == errorMethod) && len(call.Args) == 0:
		if sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
			return []ast.Expr{sel.X}, true
		}
	}
	return nil, false
}

// isFormatter says whether f puts its arguments into a log line as they
// are: one of slog's attribute constructors, or fmt.Sprint*.
func isFormatter(f *types.Func) bool {
	if f.Pkg() == nil {
		return false
	}
	path := f.Pkg().Path()
	return path == "log/slog" && f.Signature().Recv() == nil || path == "fmt" && strings.HasPrefix(f.Name(), "Sprint")
}

// sensitive says what a value of type t would leak into a log line: a
// sensitive type's description, "an error" for any error, or "".
func sensitive(r *Repo, t types.Type) string {
	if _, basic := t.(*types.Basic); t == nil || basic {
		return ""
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	if named, ok := t.(*types.Named); ok && named.Obj().Pkg() != nil {
		if what, ok := sensitiveTypes()[r.rel(named.Obj().Pkg().Path())+"."+named.Obj().Name()]; ok {
			return what
		}
	}
	if isError(t) || isError(types.NewPointer(t)) {
		return "an error"
	}
	return ""
}

// isError says whether t implements error.
func isError(t types.Type) bool {
	errType, ok := types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
	return ok && types.Implements(t, errType)
}
