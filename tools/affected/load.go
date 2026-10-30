package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var errInput = errors.New("malformed input")

// rootReadRe finds the repository paths a script test reads: "$root/<path>".
var rootReadRe = regexp.MustCompile(`\$root/([A-Za-z0-9_][A-Za-z0-9_./-]*)`)

// loadGraph reads the package listings scripts/go-packages.sh wrote into the input directory: for module i,
// go-list/<i>-tagged.json (with the e2e build tag, for imports and embeds, so the end-to-end suite is in the
// graph) and go-list/<i>-untagged.json (to know which packages `go test` can run). It reads the packages'
// Go files from the repository for the relative paths they name.
func loadGraph(repo, in fs.FS, root string, modules []string) (*graph, error) {
	var pkgs []*pkg
	for i, module := range modules {
		tagged, err := readListing(in, fmt.Sprintf("go-list/%d-tagged.json", i))
		if err != nil {
			return nil, err
		}
		untagged, err := readListing(in, fmt.Sprintf("go-list/%d-untagged.json", i))
		if err != nil {
			return nil, err
		}
		testable := map[string]bool{}
		for i := range untagged {
			lp := &untagged[i]
			testable[lp.importPath] = len(lp.goFiles)+len(lp.cgoFiles)+len(lp.testGoFiles)+len(lp.xTestGoFiles) > 0
		}
		for i := range tagged {
			p, err := newPkg(root, module, &tagged[i], testable[tagged[i].importPath])
			if err != nil {
				return nil, err
			}
			p.reads = relativeReads(p.dir, sourcesOf(repo, p.dir, &tagged[i]))
			pkgs = append(pkgs, p)
		}
	}
	return newGraph(pkgs), nil
}

// readListing decodes one `go list -e -json` output file.
func readListing(in fs.FS, name string) ([]listedPackage, error) {
	data, err := fs.ReadFile(in, name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return decodeListed(bytes.NewReader(data))
}

// sourcesOf reads a package's Go files, tests included, keyed by their repository path.
func sourcesOf(repo fs.FS, dir string, lp *listedPackage) map[string][]byte {
	sources := map[string][]byte{}
	for _, f := range slices.Concat(lp.goFiles, lp.cgoFiles, lp.testGoFiles, lp.xTestGoFiles) {
		name := path.Join(dir, f)
		if src, err := fs.ReadFile(repo, name); err == nil {
			sources[name] = src
		}
	}
	return sources
}

// loadScriptTests finds the script tests and the repository paths each reads.
func loadScriptTests(repo fs.FS) ([]scriptTest, error) {
	files, err := fs.Glob(repo, "scripts/tests/*.test.sh")
	if err != nil {
		return nil, fmt.Errorf("script tests: %w", err)
	}
	out := make([]scriptTest, 0, len(files))
	for _, f := range files {
		src, err := fs.ReadFile(repo, f)
		if err != nil {
			return nil, fmt.Errorf("script tests: %w", err)
		}
		out = append(out, scriptTest{file: f, reads: scriptReads(string(src))})
	}
	return out, nil
}

// scriptReads lists the $root/<path> references of a script, cleaned and deduplicated.
func scriptReads(src string) []string {
	var out []string
	for _, m := range rootReadRe.FindAllStringSubmatch(src, -1) {
		p := path.Clean(strings.TrimRight(m[1], "."))
		if !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out
}

// parseNameStatus reads `git diff --name-status --no-renames -z`: a status, then a path, each NUL-terminated.
func parseNameStatus(out []byte) []change {
	fields := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	var changes []change
	for i := 0; i+1 < len(fields); i += 2 {
		changes = append(changes, change{path: fields[i+1], deleted: strings.HasPrefix(fields[i], "D")})
	}
	return changes
}

// readChanges reads a `git diff --name-status --no-renames -z` output file.
func readChanges(in fs.FS, name string) ([]change, error) {
	data, err := fs.ReadFile(in, name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return parseNameStatus(data), nil
}

// readBuilds reads the "builds" file of scripts/ci-images-select.sh: one line per image whose earlier build it
// found, "<name> <commit> <digest>", with "-" for the digest when it compares with a base revision instead.
func readBuilds(in fs.FS) (map[string]signedBuild, error) {
	data, err := fs.ReadFile(in, "builds")
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]signedBuild{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read builds: %w", err)
	}
	builds := map[string]signedBuild{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; scanner.Scan(); n++ {
		const buildFields = 3
		f := strings.Fields(scanner.Text())
		if len(f) != buildFields {
			return nil, fmt.Errorf("%w: builds:%d: want <name> <commit> <digest>", errInput, n)
		}
		build := signedBuild{commit: f[1]}
		if f[2] != "-" {
			build.digest = f[2]
		}
		builds[f[0]] = build
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read builds: %w", err)
	}
	return builds, nil
}

// readMatrix reads the image matrix `task ci:images` printed.
func readMatrix(in fs.FS) ([]image, error) {
	data, err := fs.ReadFile(in, "matrix.json")
	if err != nil {
		return nil, fmt.Errorf("read matrix.json: %w", err)
	}
	var images []image
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, fmt.Errorf("%w: matrix.json: %w", errInput, err)
	}
	return images, nil
}

// planLine renders a step as the hook's runner reads it: the task, then each variable, tab-separated.
func planLine(s *step) string {
	fields := append([]string{s.task}, s.vars...)
	for i, f := range fields {
		if strings.ContainsAny(f, "\t\n") {
			fields[i] = strconv.Quote(f)
		}
	}
	return strings.Join(fields, "\t")
}
