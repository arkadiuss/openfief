package archtest

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ErrLoad reports a repository the rules cannot read: a package that does not
// build or type-check, or a directory that cannot be walked.
var ErrLoad = errors.New("load repository")

// Config says which repository Load reads.
type Config struct {
	// Root is the repository's root directory, where go.mod (and go.work) is.
	Root string
	// Module is the root module's path; package paths are made relative to it.
	Module string
	// Patterns are the go list patterns, relative to Root, such as "./...".
	Patterns []string
	// Walk lists directories to scan for tests (*_test.go, whatever their
	// build tags), SQL (*.sql) and contracts (*.proto).
	Walk []string
	// Env is the go command's environment; nil means the process's own.
	Env []string
}

// Package is one Go package of the repository as the rules see it: its
// non-test files, parsed and type-checked.
type Package struct {
	// Path is the package's path relative to the root module, such as
	// "internal/ledger/app" or "sdk/go/openfief".
	Path string
	// Name is the package clause's name.
	Name string
	// Imports are the package's imports: repository paths for the
	// repository's own packages, import paths for the rest.
	Imports []string
	Files   []*ast.File
	Fset    *token.FileSet
	Info    *types.Info
}

// SourceFile is a test, SQL or contract file found by the walk.
type SourceFile struct {
	// Path is the file's path relative to the repository root, with slashes.
	Path string
	Text string
	// Syntax is the parsed file, for Go files only.
	Syntax *ast.File
	Fset   *token.FileSet
}

// Repo is a loaded repository.
type Repo struct {
	Module   string
	Packages []*Package
	Tests    []SourceFile
	SQL      []SourceFile
	Protos   []SourceFile
}

// Load reads the repository cfg names.
func Load(cfg *Config) (*Repo, error) {
	root, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoad, err)
	}
	pc := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Dir: root,
		Env: cfg.Env,
	}
	pkgs, err := packages.Load(pc, cfg.Patterns...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoad, err)
	}
	repo := &Repo{Module: cfg.Module}
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("%w: %s: %w", ErrLoad, p.PkgPath, p.Errors[0])
		}
		repo.Packages = append(repo.Packages, repo.convert(p))
	}
	for _, dir := range cfg.Walk {
		if err := repo.walk(root, dir); err != nil {
			return nil, err
		}
	}
	return repo, nil
}

// convert keeps what the rules read of a loaded package.
func (r *Repo) convert(p *packages.Package) *Package {
	out := &Package{Path: r.rel(p.PkgPath), Name: p.Name, Files: p.Syntax, Fset: p.Fset, Info: p.TypesInfo}
	for ip := range p.Imports {
		out.Imports = append(out.Imports, r.rel(ip))
	}
	return out
}

// rel makes an import path relative to the root module when it lies in it.
func (r *Repo) rel(importPath string) string {
	if importPath == r.Module {
		return "."
	}
	if after, ok := strings.CutPrefix(importPath, r.Module+"/"); ok {
		return after
	}
	return importPath
}

// walk collects the tests, SQL and contracts below dir, skipping hidden
// directories, testdata and node_modules.
func (r *Repo) walk(root, dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLoad, err)
	}
	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != abs && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		return r.add(root, path)
	})
	if err != nil {
		return fmt.Errorf("%w: walk %s: %w", ErrLoad, dir, err)
	}
	return nil
}

// skipDir names the directories the walk never enters.
func skipDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "testdata" || name == "node_modules"
}

// add reads one walked file if a rule looks at its kind.
func (r *Repo) add(root, path string) error {
	name := filepath.Base(path)
	isSQL, isProto, isTest := strings.HasSuffix(name, ".sql"), strings.HasSuffix(name, ".proto"), strings.HasSuffix(name, "_test.go")
	if !isSQL && !isProto && !isTest {
		return nil
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return fmt.Errorf("relative path: %w", err)
	}
	f := SourceFile{Path: filepath.ToSlash(rel), Text: string(data)}
	switch {
	case isSQL:
		r.SQL = append(r.SQL, f)
	case isProto:
		r.Protos = append(r.Protos, f)
	default:
		f.Fset = token.NewFileSet()
		f.Syntax, err = parser.ParseFile(f.Fset, f.Path, data, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		r.Tests = append(r.Tests, f)
	}
	return nil
}
