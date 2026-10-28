package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

var errNoModule = errors.New("go.mod has no module line")

const (
	dirPerm  = 0o750
	filePerm = 0o600
)

// runReference writes the generated pages into -out: status.md and reference/{code-map,custom-resources,tasks}.md.
func runReference(repo fs.FS, args []string) error {
	flags := flag.NewFlagSet("reference", flag.ContinueOnError)
	packagesFile := flags.String("packages", "", "output of `go list -json ./...`")
	tasksFile := flags.String("tasks", "", "output of `task --list-all --json`")
	coverageFile := flags.String("coverage", "", "output of `kibi coverage --by req --include-passing --format json`")
	out := flags.String("out", "", "content folder to write into")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("reference: %w", err)
	}
	pages, err := referencePages(repo, *packagesFile, *tasksFile, *coverageFile)
	if err != nil {
		return err
	}
	return writePages(*out, pages)
}

// referencePages builds every generated page from the repository and the inputs the Taskfile prepared.
func referencePages(repo fs.FS, packagesFile, tasksFile, coverageFile string) (map[string]string, error) {
	module, err := moduleName(repo)
	if err != nil {
		return nil, err
	}
	pkgData, err := fs.ReadFile(repo, packagesFile)
	if err != nil {
		return nil, fmt.Errorf("read packages: %w", err)
	}
	pkgs, err := decodePackages(bytes.NewReader(pkgData))
	if err != nil {
		return nil, err
	}
	crds, err := loadCRDs(repo)
	if err != nil {
		return nil, err
	}
	tasks, err := readTasks(repo, tasksFile)
	if err != nil {
		return nil, err
	}
	coverage, err := readCoverage(repo, coverageFile)
	if err != nil {
		return nil, err
	}
	design, err := fs.ReadFile(repo, "docs/DESIGN.md")
	if err != nil {
		return nil, fmt.Errorf("read design: %w", err)
	}
	return map[string]string{
		"status.md":                     statusPage(parseDecisions(string(design)), coverage.Rows),
		"reference/code-map.md":         codeMapPage(module, pkgs),
		"reference/custom-resources.md": crdPage(crds),
		"reference/tasks.md":            tasksPage(tasks),
	}, nil
}

func moduleName(repo fs.FS) (string, error) {
	data, err := fs.ReadFile(repo, "go.mod")
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if name, ok := strings.CutPrefix(scanner.Text(), "module "); ok {
			return strings.TrimSpace(name), nil
		}
	}
	return "", errNoModule
}

func readTasks(repo fs.FS, name string) (taskList, error) {
	var v taskList
	err := readJSON(repo, name, &v)
	return v, err
}

func readCoverage(repo fs.FS, name string) (coverageReport, error) {
	var v coverageReport
	err := readJSON(repo, name, &v)
	return v, err
}

func readJSON(repo fs.FS, name string, v any) error {
	data, err := fs.ReadFile(repo, name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}

// writePages writes each page below dir; os.Root keeps every write inside it.
func writePages(dir string, pages map[string]string) error {
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("open %s: %w", dir, err)
	}
	return errors.Join(writeAll(root, pages), root.Close())
}

func writeAll(root *os.Root, pages map[string]string) error {
	if err := root.MkdirAll("reference", dirPerm); err != nil {
		return fmt.Errorf("create reference folder: %w", err)
	}
	for name, content := range pages {
		if err := root.WriteFile(name, []byte(content), filePerm); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
