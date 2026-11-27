package app_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"
)

// unauthorizedByAccount are the read procedures of Views that don't ask the
// authorizer about the account they answer for, each with why that is safe.
// Those marked household are household-level reads (S5): they name no account,
// so what a holder of read access gets of them is the household's, or public
// data, whatever it is granted, documented as such in docs/concepts/read-api.md
// ("Household-level reads"). A procedure may be listed though it also asks
// about accounts for its own purpose (ModelPortfolio resolves its range from the
// accounts the caller may view, Instrument filters the transactions on the
// instrument): the list says the data it returns is not filtered by account.
// A new read procedure that asks the authorizer about no account must be added
// here, or fail TestEveryReadProcedureAsksTheAuthorizerOrIsHouseholdLevel.
func unauthorizedByAccount() map[string]string {
	const household = "household: "
	const passed = "of an account an authorized read already returned: "
	return map[string]string{
		"ModelPortfolio":        household + "the model portfolio is the owner's plan, not a view of an account",
		"PriceChecks":           household + "cross-checks of market prices between sources, tied to instruments, not accounts",
		"Instrument":            household + "an instrument's identity and prices are market data; the transactions on it are filtered per account",
		"InstrumentPrices":      household + "an instrument's price history is market data",
		"Plugins":               household + "which plugins are installed is a fact of the install, not of an account",
		"InstrumentNames":       household + "a name is served only with an instrument an answer already names (D-197)",
		"Settings":              household + "the owner's number format and the jurisdiction's wrappers (D-78)",
		"MarketQuote":           household + "public quotes, decided by who holds the token (identity.MayReadQuotes, D-226), not by an account",
		"MarketQuotesAvailable": household + "whether quotes are served at all, decided like MarketQuote (D-226)",
		"AccountSettings":       passed + "the read API calls it with the account Account or Accounts returned for the caller",
		"AccountTerms":          passed + "the read API and the tax reads call it with an account they already authorized",
	}
}

// asksTheAuthorizer reports whether name is a selector by which the views ask
// the authorizer about an account.
func asksTheAuthorizer(name string) bool { return name == "MayAct" || name == "ActableAccounts" }

// readScan is what the source of the package says of its functions.
type readScan struct {
	// procedures are the exported read procedures of Views.
	procedures []string
	// calls are the package's own functions each function calls, by name.
	calls map[string]map[string]bool
	// asks are the functions that ask the authorizer themselves.
	asks map[string]bool
}

// scanViews reads the non-test source of the package in the current directory.
func scanViews(t *testing.T) *readScan {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package: %v", err)
	}
	scan := &readScan{calls: map[string]map[string]bool{}, asks: map[string]bool{}}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, parseErr := parser.ParseFile(fset, e.Name(), nil, 0)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", e.Name(), parseErr)
		}
		for _, d := range f.Decls {
			if fn, isFunc := d.(*ast.FuncDecl); isFunc && fn.Body != nil {
				scan.add(fn)
			}
		}
	}
	return scan
}

// add records fn: its calls, whether it asks the authorizer, and whether it is
// a read procedure.
func (s *readScan) add(fn *ast.FuncDecl) {
	name := fn.Name.Name
	s.calls[name] = map[string]bool{}
	recv := receiverName(fn)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, isCall := n.(*ast.CallExpr)
		if !isCall {
			return true
		}
		switch f := call.Fun.(type) {
		case *ast.SelectorExpr:
			if asksTheAuthorizer(f.Sel.Name) {
				s.asks[name] = true
			}
			// A method of Views called on the receiver.
			if x, isIdent := f.X.(*ast.Ident); isIdent && recv != "" && x.Name == recv {
				s.calls[name][f.Sel.Name] = true
			}
		case *ast.Ident:
			s.calls[name][f.Name] = true
		}
		return true
	})
	if isReadProcedure(fn) {
		s.procedures = append(s.procedures, name)
	}
}

// reaches reports whether name is, or calls through the package's functions by
// name, one that asks the authorizer.
func (s *readScan) reaches(name string) bool {
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(n string) bool {
		if s.asks[n] {
			return true
		}
		if seen[n] {
			return false
		}
		seen[n] = true
		for callee := range s.calls[n] {
			if _, declared := s.calls[callee]; declared && callee != n && walk(callee) {
				return true
			}
		}
		return false
	}
	return walk(name)
}

// TestEveryReadProcedureAsksTheAuthorizerOrIsHouseholdLevel is the guard of
// S5: every exported read procedure of Views (a method taking a context) either
// reaches, through the package's own functions, a call that asks the
// authorizer about an account (MayAct or ActableAccounts), or is on the list
// above, so a new read can't skip both. It reads the package's source, since
// no run of a procedure tells whether it asked.
func TestEveryReadProcedureAsksTheAuthorizerOrIsHouseholdLevel(t *testing.T) {
	t.Parallel()
	scan := scanViews(t)
	const atLeast = 30
	if len(scan.procedures) < atLeast {
		t.Fatalf("found %d read procedures, want the %d and more of Views: the scan broke", len(scan.procedures), atLeast)
	}
	listed := unauthorizedByAccount()
	for _, p := range scan.procedures {
		if _, isListed := listed[p]; !isListed && !scan.reaches(p) {
			t.Errorf("%s asks the authorizer about no account and is not listed in unauthorizedByAccount: authorize it per account, "+
				"or list it with its reason and, if it is household-level, say so in docs/concepts/read-api.md (S5)", p)
		}
	}
	for p := range listed {
		if !slices.Contains(scan.procedures, p) {
			t.Errorf("%s is listed in unauthorizedByAccount but is no read procedure of Views", p)
		}
	}
}

// receiverName returns the name of fn's receiver, or "" for a function.
func receiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) != 1 || len(fn.Recv.List[0].Names) != 1 {
		return ""
	}
	return fn.Recv.List[0].Names[0].Name
}

// isReadProcedure reports whether fn is an exported method of Views whose
// first parameter is a context.
func isReadProcedure(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || !fn.Name.IsExported() || len(fn.Recv.List) != 1 || fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return false
	}
	recv := fn.Recv.List[0].Type
	if star, isStar := recv.(*ast.StarExpr); isStar {
		recv = star.X
	}
	if id, isIdent := recv.(*ast.Ident); !isIdent || id.Name != "Views" {
		return false
	}
	sel, isSel := fn.Type.Params.List[0].Type.(*ast.SelectorExpr)
	return isSel && sel.Sel.Name == "Context"
}
