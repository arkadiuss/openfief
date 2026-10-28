package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

const testCRD = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: plugins.example.io
spec:
  group: example.io
  names:
    kind: Plugin
    plural: plugins
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        type: object
        properties:
          apiVersion:
            type: string
          spec:
            type: object
            required: [image]
            properties:
              image:
                type: string
                description: Image by digest. Tags are refused.
                pattern: '^[a-z]+@sha256:[a-f0-9]{64}$'
              mode:
                type: string
                enum: [live, paper]
              hosts:
                type: array
                items:
                  type: string
`

func TestDecodePackages(t *testing.T) {
	t.Parallel()
	stream := `{"ImportPath":"m/internal/ledger","Doc":"Package ledger records.","GoFiles":["a.go","b.go"],"TestGoFiles":["a_test.go"],"Imports":["m/internal/kernel","fmt"]}
{"ImportPath":"m/cmd/core","GoFiles":["main.go"]}`
	pkgs, err := decodePackages(strings.NewReader(stream))
	if err != nil {
		t.Fatalf("decodePackages() error = %v", err)
	}
	if len(pkgs) != 2 || pkgs[0].importPath != "m/internal/ledger" || len(pkgs[0].goFiles) != 2 || len(pkgs[0].imports) != 2 || pkgs[1].doc != "" {
		t.Errorf("decodePackages() = %+v", pkgs)
	}
	if _, err := decodePackages(strings.NewReader(`{"GoFiles": 3}`)); err == nil {
		t.Error("decodePackages() accepted a number for GoFiles")
	}
}

func TestCRDPage(t *testing.T) {
	t.Parallel()
	crds, err := loadCRDs(fstest.MapFS{"deploy/crds/plugins.yaml": {Data: []byte(testCRD)}, "deploy/crds/kustomization.yaml": {Data: []byte("kind: Kustomization\n")}})
	if err != nil {
		t.Fatalf("loadCRDs() error = %v", err)
	}
	got := crdPage(crds)
	for _, want := range []string{
		"## Plugin\n\n`example.io/v1alpha1`, namespaced-scoped, plural `plugins`.",
		"| `spec.image` | string matching `^[a-z]+@sha256:[a-f0-9]{64}$` | yes | Image by digest. |",
		"| `spec.mode` | string: live, paper |  |  |",
		"| `spec.hosts` | []string |  |  |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("crdPage() lacks %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "`apiVersion`") {
		t.Error("crdPage() lists the standard apiVersion field")
	}
}

func TestCodeMapPage(t *testing.T) {
	t.Parallel()
	pkgs := []goPackage{
		{importPath: "m/internal/ledger", doc: "Package ledger records. More.", goFiles: []string{"a.go"}, imports: []string{"m/internal/kernel", "m/internal/portfolio"}},
		{importPath: "m/internal/kernel"},
		{importPath: "m/internal/portfolio"},
	}
	got := codeMapPage("m", pkgs)
	for _, want := range []string{"c_ledger -.-> c_kernel", "c_ledger --> c_portfolio", "No violations.", "| [`internal/ledger`](../../../../internal/ledger/) | domain | 1 | 0 | Package ledger records. |"} {
		if !strings.Contains(got, want) {
			t.Errorf("codeMapPage() lacks %q", want)
		}
	}
}

func TestStatusPage(t *testing.T) {
	t.Parallel()
	rows := []coverageRow{{ID: "REQ-a", Evidence: []evidence{{Ref: "design-decisions:D-18#1"}}, DirectSymbolCount: 1, TestCount: 1}}
	got := statusPage(parseDecisions("### 3.3 Ledger\n\n- **[D-18]** The log is append-only. More text.\n- **[D-19]** Postings.\n"), rows)
	for _, want := range []string{"'pie1': '#3f8f47'", "\"Built and tested\" : 1", "\"No requirement in kibi\" : 1", "| D-18 | The log is append-only. | 1 | 1 | 1 | Built and tested |"} {
		if !strings.Contains(got, want) {
			t.Errorf("statusPage() lacks %q in:\n%s", want, got)
		}
	}
}

func TestRunReference(t *testing.T) {
	t.Parallel()
	repo := fstest.MapFS{
		"go.mod":                   {Data: []byte("module m\n\ngo 1.26\n")},
		"docs/DESIGN.md":           {Data: []byte("- **[D-1]** Scope.\n")},
		"deploy/crds/plugins.yaml": {Data: []byte(testCRD)},
		"in/packages.json":         {Data: []byte(`{"ImportPath":"m/internal/kernel"}`)},
		"in/tasks.json":            {Data: []byte(`{"tasks":[{"name":"test","desc":"Tests"}]}`)},
		"in/coverage.json":         {Data: []byte(`{"rows":[]}`)},
	}
	out := t.TempDir()
	err := runReference(repo, []string{"-packages", "in/packages.json", "-tasks", "in/tasks.json", "-coverage", "in/coverage.json", "-out", out})
	if err != nil {
		t.Fatalf("runReference() error = %v", err)
	}
	written := os.DirFS(out)
	for _, name := range []string{"status.md", "reference/code-map.md", "reference/custom-resources.md", "reference/tasks.md"} {
		data, err := fs.ReadFile(written, name)
		if err != nil || !strings.HasPrefix(string(data), "---\ntitle: ") {
			t.Errorf("%s: err = %v, content starts %q", name, err, string(data[:min(len(data), 20)]))
		}
	}
	delete(repo, "go.mod")
	if err := runReference(repo, []string{"-packages", "in/packages.json", "-tasks", "in/tasks.json", "-coverage", "in/coverage.json", "-out", out}); err == nil {
		t.Error("runReference() succeeded without go.mod")
	}
}

func TestRunRefs(t *testing.T) {
	t.Parallel()
	repo := fstest.MapFS{
		".kb/symbols.yaml":              {Data: []byte("symbols:\n  - id: SYM-a\n")},
		".kb/requirements/REQ-a.md":     {},
		"in/tasks.json":                 {Data: []byte(`{"tasks":[{"name":"test"}]}`)},
		"docs/concepts/good.md":         {Data: []byte("REQ-a, SYM-a and `task test`.\n")},
		"site/content/_index.md":        {Data: []byte("Nothing to check.\n")},
		"docs/concepts/sub/notpage.txt": {Data: []byte("REQ-missing\n")},
	}
	var out bytes.Buffer
	if err := runRefs(repo, []string{"-tasks", "in/tasks.json"}, &out); err != nil {
		t.Fatalf("runRefs() error = %v; output:\n%s", err, out.String())
	}
	repo["docs/concepts/bad.md"] = &fstest.MapFile{Data: []byte("SCEN-missing\n")}
	out.Reset()
	err := runRefs(repo, []string{"-tasks", "in/tasks.json"}, &out)
	if !errors.Is(err, errBrokenRefs) || !strings.Contains(out.String(), "docs/concepts/bad.md:1: SCEN-missing is not in kibi") {
		t.Errorf("runRefs() error = %v; output:\n%s", err, out.String())
	}
}

func TestRunUsage(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"nope"}} {
		if err := run(args, strings.NewReader(""), &bytes.Buffer{}); !errors.Is(err, errUsage) {
			t.Errorf("run(%q) error = %v, want a usage error", args, err)
		}
	}
}
