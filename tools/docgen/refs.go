package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

var errBrokenRefs = errors.New("docs name kibi entities or tasks that don't exist")

var (
	kbIDRe    = regexp.MustCompile(`\b(REQ|TEST|SCEN|SYM)-[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9]\b(-?\*)?`)
	taskRefRe = regexp.MustCompile("`task ([a-z][a-z0-9:_-]*)")
)

// kbFolder is the folder holding one Markdown file per kibi entity with the given ID prefix.
func kbFolder(prefix string) string {
	switch prefix {
	case "REQ":
		return ".kb/requirements"
	case "TEST":
		return ".kb/tests"
	default:
		return ".kb/scenarios"
	}
}

// runRefs checks that every kibi ID and `task` name the pages mention exists.
func runRefs(repo fs.FS, args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("refs", flag.ContinueOnError)
	tasksFile := flags.String("tasks", "", "output of `task --list-all --json`")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("refs: %w", err)
	}
	list, err := readTasks(repo, *tasksFile)
	if err != nil {
		return err
	}
	symbols, err := fs.ReadFile(repo, ".kb/symbols.yaml")
	if err != nil {
		return fmt.Errorf("read kibi symbols: %w", err)
	}
	known := knownRefs{repo: repo, symbols: string(symbols), tasks: map[string]bool{}}
	for _, t := range list.Tasks {
		known.tasks[t.Name] = true
	}
	var problems []string
	for _, root := range []string{docsRoot, "site/content"} {
		p, err := checkPages(repo, root, known)
		if err != nil {
			return err
		}
		problems = append(problems, p...)
	}
	report := "docs: every kibi ID and task name in the pages exists\n"
	if len(problems) > 0 {
		report = "docs: " + strings.Join(problems, "\ndocs: ") + "\n"
	}
	if _, err := io.WriteString(stdout, report); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %d", errBrokenRefs, len(problems))
	}
	return nil
}

// knownRefs answers whether a kibi entity or a task exists.
type knownRefs struct {
	repo    fs.FS
	symbols string
	tasks   map[string]bool
}

func (k knownRefs) hasEntity(id string) bool {
	prefix, _, _ := strings.Cut(id, "-")
	if prefix == "SYM" {
		return strings.Contains(k.symbols, "id: "+id+"\n")
	}
	_, err := fs.Stat(k.repo, path.Join(kbFolder(prefix), id+".md"))
	return err == nil
}

func checkPages(repo fs.FS, root string, known knownRefs) ([]string, error) {
	var problems []string
	err := fs.WalkDir(repo, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".md" {
			return err
		}
		data, err := fs.ReadFile(repo, p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		problems = append(problems, pageRefProblems(p, string(data), known)...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("check %s: %w", root, err)
	}
	return problems, nil
}

// pageRefProblems lists, with line numbers, the kibi IDs and task names on one page that don't exist.
// An ID followed by `*` or `-*` is a pattern, not a reference.
func pageRefProblems(name, text string, known knownRefs) []string {
	var problems []string
	n := 0
	for line := range strings.Lines(text) {
		n++
		for _, m := range kbIDRe.FindAllStringSubmatch(line, -1) {
			if m[2] == "" && !known.hasEntity(m[0]) {
				problems = append(problems, fmt.Sprintf("%s:%d: %s is not in kibi", name, n, m[0]))
			}
		}
		for _, m := range taskRefRe.FindAllStringSubmatch(line, -1) {
			if !known.tasks[m[1]] {
				problems = append(problems, fmt.Sprintf("%s:%d: there is no task %q", name, n, m[1]))
			}
		}
	}
	return slices.Compact(problems)
}
