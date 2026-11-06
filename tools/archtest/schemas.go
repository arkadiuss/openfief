package archtest

import (
	"go/ast"
	"go/token"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// schemaOwners maps each PostgreSQL schema a context owns to that context:
// every context's own, and River's job queue, which valuation migrates and
// runs (D-36).
func schemaOwners() map[string]string {
	owners := map[string]string{"river": ctxValuation}
	for ctx := range contextMap() {
		owners[ctx] = ctx
	}
	return owners
}

// Schemas checks that a context's SQL, in its migrations and in its Go
// packages' query strings, names only schemas the context owns.
func Schemas(r *Repo) []Violation {
	var out []Violation
	for _, f := range r.SQL {
		out = append(out, schemaBreaks(f.Path, PlaceOf(dirOf(f.Path)).Context, f.Text)...)
	}
	for _, p := range r.Packages {
		ctx := PlaceOf(p.Path).Context
		if ctx == "" {
			continue
		}
		for _, f := range p.Files {
			name := p.Fset.Position(f.Package).Filename
			for _, lit := range sqlLiterals(f) {
				out = append(out, schemaBreaks(p.Path+"/"+baseName(name), ctx, lit)...)
			}
		}
	}
	return dedupe(out)
}

// schemaBreaks lists the foreign schemas one SQL text names, once each.
func schemaBreaks(file, ctx, sql string) []Violation {
	if ctx == "" {
		return nil
	}
	owners := schemaOwners()
	ref := regexp.MustCompile(`\b([a-z_][a-z0-9_]*)\.[a-z_][a-z0-9_]*\b`)
	var out []Violation
	for _, m := range ref.FindAllStringSubmatch(stripSQLComments(sql), -1) {
		owner, ok := owners[m[1]]
		if !ok || owner == ctx {
			continue
		}
		out = append(out, Violation{RuleSchemas, file + " names " + m[1], "the " + ctx + " context's SQL names " + m[0] + ", in schema " + m[1] + " of " + owner})
	}
	return out
}

// stripSQLComments drops -- line comments and /* */ block comments.
func stripSQLComments(sql string) string {
	sql = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(sql, " ")
	return regexp.MustCompile(`--[^\n]*`).ReplaceAllString(sql, " ")
}

// sqlLiterals are a file's string literals that read as SQL statements.
func sqlLiterals(f *ast.File) []string {
	keyword := regexp.MustCompile(`(?i)\b(select|insert\s+into|update|delete\s+from|create|alter|drop)\b`)
	var out []string
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		if s, err := strconv.Unquote(lit.Value); err == nil && keyword.MatchString(s) {
			out = append(out, s)
		}
		return true
	})
	return out
}

// dirOf is a slash path's directory.
func dirOf(file string) string {
	if i := strings.LastIndex(file, "/"); i >= 0 {
		return file[:i]
	}
	return "."
}

// baseName is a path's last element, for slash or OS paths.
func baseName(file string) string {
	return file[strings.LastIndexAny(file, `/\`)+1:]
}

// dedupe keeps the first violation of each rule and key, in order.
func dedupe(vs []Violation) []Violation {
	var out []Violation
	for _, v := range vs {
		if !slices.ContainsFunc(out, func(o Violation) bool { return o.Rule == v.Rule && o.Key == v.Key }) {
			out = append(out, v)
		}
	}
	return out
}
