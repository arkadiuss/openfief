package main

import (
	"bufio"
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

var (
	errStale     = errors.New("docs are stale")
	errUncovered = errors.New("packages without a covering docs page")
)

const (
	// docsRoot holds every hand-written page; only pages with `covers` front matter take part in the gate.
	docsRoot = "docs"
	// logFields are the parts of one git log record: hash, message and changed files.
	logFields = 3
)

// trailerRe finds the commit trailer that excuses a commit from updating its pages.
var trailerRe = regexp.MustCompile(`(?m)^Docs-Unchanged:[ \t]*\S`)

// page is a hand-written docs page and the code paths it covers.
type page struct {
	path   string
	covers []string
}

// commit is one commit's changed files and whether it says why its pages stay unchanged.
type commit struct {
	id     string
	files  []string
	exempt bool
}

// finding is a covered file that changed while none of the pages covering it did.
type finding struct {
	commit string
	file   string
	pages  []string
}

// runFresh is the staleness gate: changes come from stdin, either `git log` records (a pushed range) or,
// with -message, the staged file names and the commit message being written.
func runFresh(repo fs.FS, args []string, stdin io.Reader, stdout io.Writer) error {
	flags := flag.NewFlagSet("fresh", flag.ContinueOnError)
	message := flags.String("message", "", "commit message file (staged mode); stdin then lists staged files")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("fresh: %w", err)
	}
	commits, err := readChanges(repo, *message, stdin)
	if err != nil {
		return err
	}
	pages, err := loadPages(repo)
	if err != nil {
		return err
	}
	goFiles, err := goSourceFiles(repo)
	if err != nil {
		return err
	}
	return reportFresh(stdout, staleFindings(pages, commits), uncoveredPackages(pages, goFiles))
}

// readChanges turns the gate's input into commits.
func readChanges(repo fs.FS, messageFile string, stdin io.Reader) ([]commit, error) {
	data, err := io.ReadAll(stdin)
	if err != nil {
		return nil, fmt.Errorf("read changes: %w", err)
	}
	if messageFile == "" {
		return parseLog(string(data)), nil
	}
	msg, err := fs.ReadFile(repo, messageFile)
	if err != nil {
		return nil, fmt.Errorf("read commit message: %w", err)
	}
	return []commit{{id: "the staged change", files: nonEmptyLines(string(data)), exempt: trailerRe.Match(msg)}}, nil
}

// parseLog reads `git log --no-merges --format=%x1e%H%x1f%B%x1f --name-only` output.
func parseLog(log string) []commit {
	var commits []commit
	for record := range strings.SplitSeq(log, "\x1e") {
		parts := strings.SplitN(record, "\x1f", logFields)
		if len(parts) != logFields {
			continue
		}
		commits = append(commits, commit{
			id:     shortID(strings.TrimSpace(parts[0])),
			files:  nonEmptyLines(parts[2]),
			exempt: trailerRe.MatchString(parts[1]),
		})
	}
	return commits
}

func shortID(id string) string {
	const short = 12
	if len(id) > short {
		return id[:short]
	}
	return id
}

func nonEmptyLines(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// staleFindings lists covered files changed by commits without a trailer, where no page covering the file
// changed anywhere in the same set of commits. Tests and test data never count: they don't change behavior.
func staleFindings(pages []page, commits []commit) []finding {
	changed := map[string]bool{}
	for _, c := range commits {
		for _, f := range c.files {
			changed[f] = true
		}
	}
	var findings []finding
	for _, c := range commits {
		if c.exempt {
			continue
		}
		for _, f := range c.files {
			covering := coveringPages(pages, f)
			if isTestFile(f) || len(covering) == 0 || slices.ContainsFunc(covering, func(p string) bool { return changed[p] }) {
				continue
			}
			findings = append(findings, finding{commit: c.id, file: f, pages: covering})
		}
	}
	return findings
}

func isTestFile(file string) bool {
	return strings.HasSuffix(file, "_test.go") || slices.Contains(strings.Split(file, "/"), "testdata")
}

func coveringPages(pages []page, file string) []string {
	var paths []string
	for _, p := range pages {
		if slices.ContainsFunc(p.covers, func(glob string) bool { return matchGlob(glob, file) }) {
			paths = append(paths, p.path)
		}
	}
	return paths
}

// uncoveredPackages lists the directories holding non-test Go files that no page covers.
func uncoveredPackages(pages []page, goFiles []string) []string {
	covered := map[string]bool{}
	dirs := map[string]bool{}
	for _, f := range goFiles {
		dir := path.Dir(f)
		dirs[dir] = true
		if len(coveringPages(pages, f)) > 0 {
			covered[dir] = true
		}
	}
	var missing []string
	for dir := range dirs {
		if !covered[dir] {
			missing = append(missing, dir)
		}
	}
	slices.Sort(missing)
	return missing
}

func reportFresh(w io.Writer, findings []finding, uncovered []string) error {
	var b strings.Builder
	var errs []error
	for _, f := range findings {
		fmt.Fprintf(&b, "docs: %s changed in %s, but none of its pages did: %s\n", f.file, f.commit, strings.Join(f.pages, ", "))
	}
	if len(findings) > 0 {
		b.WriteString("docs: update one of those pages in the same push, or add a 'Docs-Unchanged: <reason>' trailer to the commit\n")
		errs = append(errs, fmt.Errorf("%w: %d covered files changed without their pages", errStale, len(findings)))
	}
	for _, dir := range uncovered {
		fmt.Fprintf(&b, "docs: no page covers %s; add it to a page's covers (docs/dev/docs.md)\n", dir)
	}
	if len(uncovered) > 0 {
		errs = append(errs, fmt.Errorf("%w: %d", errUncovered, len(uncovered)))
	}
	if len(errs) == 0 {
		b.WriteString("docs: every covered change updated its pages, and every Go package is covered\n")
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		errs = append(errs, fmt.Errorf("write report: %w", err))
	}
	return errors.Join(errs...)
}

// matchGlob matches a slash-separated path against a pattern where `*` matches within one path segment
// and a `**` segment matches any number of segments, including none.
func matchGlob(pattern, name string) bool {
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

func matchSegments(pattern, name []string) bool {
	if len(pattern) == 0 {
		return len(name) == 0
	}
	if pattern[0] == "**" {
		for i := range len(name) + 1 {
			if matchSegments(pattern[1:], name[i:]) {
				return true
			}
		}
		return false
	}
	if len(name) == 0 {
		return false
	}
	ok, err := path.Match(pattern[0], name[0])
	return err == nil && ok && matchSegments(pattern[1:], name[1:])
}
