package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"sigs.k8s.io/yaml"
)

var errFrontMatter = errors.New("unterminated or invalid front matter")

// frontMatter is the part of a page's YAML front matter that docgen reads.
type frontMatter struct {
	Title  string   `json:"title"`
	Covers []string `json:"covers"`
}

// skippedDir says whether a folder is never searched for Go packages: tooling, generated output and dependencies.
func skippedDir(name string) bool {
	switch name {
	case "node_modules", "site", "testdata", "vendor", "dist":
		return true
	default:
		return strings.HasPrefix(name, ".")
	}
}

// loadPages reads every Markdown page under docs/ that says which code it covers.
func loadPages(repo fs.FS) ([]page, error) {
	var pages []page
	err := fs.WalkDir(repo, docsRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".md" {
			return err
		}
		data, err := fs.ReadFile(repo, p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		fm, err := parseFrontMatter(data)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if len(fm.Covers) > 0 {
			pages = append(pages, page{path: p, covers: fm.Covers})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load pages: %w", err)
	}
	return pages, nil
}

// parseFrontMatter reads YAML front matter between leading `---` lines; a page without any has none.
func parseFrontMatter(data []byte) (frontMatter, error) {
	var fm frontMatter
	text := string(data)
	if !strings.HasPrefix(text, "---\n") {
		return fm, nil
	}
	body, _, found := strings.Cut(text[len("---\n"):], "\n---\n")
	if !found {
		return fm, errFrontMatter
	}
	if err := yaml.Unmarshal([]byte(body), &fm); err != nil {
		return fm, fmt.Errorf("%w: %w", errFrontMatter, err)
	}
	return fm, nil
}

// goSourceFiles lists the repository's non-test Go files, skipping hidden, generated and dependency folders.
func goSourceFiles(repo fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(repo, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != "." && skippedDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if path.Ext(p) == ".go" && !isTestFile(p) {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list Go files: %w", err)
	}
	return files, nil
}
