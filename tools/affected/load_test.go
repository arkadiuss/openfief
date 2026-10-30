package main

import (
	"bytes"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

func TestParseNameStatus(t *testing.T) {
	t.Parallel()
	got := parseNameStatus([]byte("M\x00internal/kernel/money.go\x00D\x00docs/old.md\x00A\x00plugins/x/main.go\x00"))
	want := []change{{path: "internal/kernel/money.go"}, {path: "docs/old.md", deleted: true}, {path: "plugins/x/main.go"}}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(change{})); diff != "" {
		t.Errorf("parseNameStatus() mismatch (-want +got):\n%s", diff)
	}
	if got := parseNameStatus(nil); len(got) != 0 {
		t.Errorf("parseNameStatus(nil) = %v, want none", got)
	}
}

func TestScriptReads(t *testing.T) {
	t.Parallel()
	src := `cp -r "$root/deploy/demo" "$work"; bash "$root/scripts/x.sh"; cat $root/docs/DESIGN.md. "$root/deploy"`
	want := []string{"deploy", "deploy/demo", "docs/DESIGN.md", "scripts/x.sh"}
	if diff := cmp.Diff(want, scriptReads(src)); diff != "" {
		t.Errorf("scriptReads() mismatch (-want +got):\n%s", diff)
	}
}

func TestReadBuilds(t *testing.T) {
	t.Parallel()
	const d = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	got, err := readBuilds(fstest.MapFS{"builds": {Data: []byte("core aaaa " + d + "\nnbp-fx bbbb -\n")}})
	want := map[string]signedBuild{"core": {commit: "aaaa", digest: d}, "nbp-fx": {commit: "bbbb"}}
	if err != nil || !cmp.Equal(want, got, cmp.AllowUnexported(signedBuild{})) {
		t.Errorf("readBuilds() = %v, %v; want %v", got, err, want)
	}
	if got, err := readBuilds(fstest.MapFS{}); err != nil || len(got) != 0 {
		t.Errorf("readBuilds(no file) = %v, %v; want none", got, err)
	}
	if _, err := readBuilds(fstest.MapFS{"builds": {Data: []byte("core aaaa\n")}}); !errors.Is(err, errInput) {
		t.Errorf("readBuilds(two fields) error = %v, want errInput", err)
	}
	if _, err := readMatrix(fstest.MapFS{"matrix.json": {Data: []byte("{")}}); !errors.Is(err, errInput) {
		t.Errorf("readMatrix(bad JSON) error = %v, want errInput", err)
	}
}

func TestPlanLine(t *testing.T) {
	t.Parallel()
	s := step{task: "lint:go", vars: []string{"MODULE=.", "PKGS=./a ./b"}}
	if got, want := planLine(&s), "lint:go\tMODULE=.\tPKGS=./a ./b"; got != want {
		t.Errorf("planLine() = %q, want %q", got, want)
	}
}

func TestRelativeReads(t *testing.T) {
	t.Parallel()
	sources := map[string][]byte{
		"internal/plugins/operator/x_test.go": []byte(`package operator
import "path/filepath"
var crds = filepath.Join("..", "..", "..", "deploy", "crds")
var spec = "../../../gen/openapi/read.yaml"
var local = filepath.Join("testdata", "x")
var inside = "../operator/testdata"
var outside = "../../../../etc/passwd"
var dynamic = filepath.Join("..", name)
`),
		"internal/plugins/operator/broken.go": []byte("package operator\nfunc {"),
	}
	want := []string{"deploy/crds", "gen/openapi/read.yaml"}
	if diff := cmp.Diff(want, relativeReads("internal/plugins/operator", sources)); diff != "" {
		t.Errorf("relativeReads() mismatch (-want +got):\n%s", diff)
	}
}

func TestGraphHelpers(t *testing.T) {
	t.Parallel()
	g := testGraph()
	if ip, ok := g.owningPackage("internal/ledger/testdata/deep/x.json"); !ok || ip != mod+"internal/ledger" {
		t.Errorf("owningPackage() = %q, %v; want internal/ledger", ip, ok)
	}
	if _, ok := g.owningPackage("README.md"); ok {
		t.Error("owningPackage(README.md) found a package; root files belong to none")
	}
	if got := g.readers("deploy/crds/x.yaml"); len(got) != 1 || got[0] != mod+"internal/plugins/operator" {
		t.Errorf("readers() = %v, want the operator", got)
	}
	deps := g.dependencies(mod + "cmd/core")
	for _, ip := range []string{"cmd/core", "internal/ledger", "internal/kernel"} {
		if !deps[mod+ip] {
			t.Errorf("dependencies(cmd/core) lacks %s: %v", ip, deps)
		}
	}
	if deps[mod+"internal/valuation"] {
		t.Error("dependencies(cmd/core) holds internal/valuation, which it doesn't import")
	}
}

func TestNewPkg(t *testing.T) {
	t.Parallel()
	root := filepath.FromSlash("/repo")
	listed, err := decodeListed(strings.NewReader(`{"Dir":"/repo/internal/views/ui","ImportPath":"m/ui","Imports":["m/kernel"],
		"TestImports":["m/ui","testing"],"XTestImports":["m/ledger"],"EmbedFiles":["web/app.css"],"TestEmbedFiles":["testdata/x"]}
		{"Dir":"/elsewhere/x","ImportPath":"m/x"}`))
	if err != nil || len(listed) != 2 {
		t.Fatalf("decodeListed() = %v, %v", listed, err)
	}
	p, err := newPkg(root, ".", &listed[0], true)
	if err != nil {
		t.Fatalf("newPkg() error = %v", err)
	}
	want := &pkg{
		importPath: "m/ui", dir: "internal/views/ui", module: ".", imports: []string{"m/kernel"},
		testImports: []string{"testing", "m/ledger"}, embeds: []string{"internal/views/ui/web/app.css"},
		testEmbeds: []string{"internal/views/ui/testdata/x"}, testable: true,
	}
	if diff := cmp.Diff(want, p, cmp.AllowUnexported(pkg{})); diff != "" {
		t.Errorf("newPkg() mismatch (-want +got):\n%s", diff)
	}
	if _, err := newPkg(root, ".", &listed[1], true); !errors.Is(err, errOutsideRepo) {
		t.Errorf("newPkg(outside) error = %v, want errOutsideRepo", err)
	}
	if _, err := decodeListed(strings.NewReader("{")); err == nil {
		t.Error("decodeListed(truncated) succeeded")
	}
}

// fakeRoot is the absolute repository path the fake package listings use.
const fakeRoot = "/repo"

// fakeRepo is a repository with a kernel, a command whose tests read deploy/crds, and two script tests; and an
// input directory holding their package listings, as scripts/check-commit.sh and ci-images-select.sh write it.
func fakeRepo() fakeFS {
	repo := fstest.MapFS{
		"internal/kernel/money.go":         {Data: []byte("package kernel\n")},
		"cmd/core/main.go":                 {Data: []byte("package main\n")},
		"cmd/core/main_test.go":            {Data: []byte("package main\nvar crds = \"../../deploy/crds\"\n")},
		"scripts/tests/demo.test.sh":       {Data: []byte("cat \"$root/deploy/demo/stack.yaml\"\n")},
		"scripts/tests/check-docs.test.sh": {Data: []byte("cat \"$root/docs/DESIGN.md\"\n")},
	}
	listing := []byte(`{"Dir":"/repo/internal/kernel","ImportPath":"github.com/arkadiuss/openfief/internal/kernel","GoFiles":["money.go"]}
{"Dir":"/repo/cmd/core","ImportPath":"github.com/arkadiuss/openfief/cmd/core","Imports":["github.com/arkadiuss/openfief/internal/kernel"],"GoFiles":["main.go"],"TestGoFiles":["main_test.go"]}`)
	in := fstest.MapFS{
		"go-list/0-tagged.json":   {Data: listing},
		"go-list/0-untagged.json": {Data: listing},
		"matrix.json":             {Data: []byte(`[{"name":"core","dockerfile":"cmd/core/Dockerfile","image":"ghcr.io/example/openfief/core"}]`)},
	}
	return fakeFS{repo: repo, in: in}
}

// fakeFS is the fake repository and input directory.
type fakeFS struct{ repo, in fstest.MapFS }

// output is what a command printed.
type output struct{ stdout, stderr string }

// runFake runs a command against the fake repository and input directory.
func runFake(t *testing.T, f fakeFS, stdin string, args ...string) output {
	t.Helper()
	var out, errOut bytes.Buffer
	e := env{
		repo: f.repo, root: fakeRoot, dir: func(string) fs.FS { return f.in },
		stdin: strings.NewReader(stdin), stdout: &out, stderr: &errOut,
	}
	if err := run(args, &e); err != nil {
		t.Fatalf("run(%v) error = %v; stderr:\n%s", args, err, errOut.String())
	}
	return output{stdout: out.String(), stderr: errOut.String()}
}

func TestRunHook(t *testing.T) {
	t.Parallel()
	f := fakeRepo()
	f.in["changes"] = &fstest.MapFile{Data: []byte("M\x00internal/kernel/money.go\x00M\x00deploy/crds/x.yaml\x00M\x00deploy/demo/stack.yaml\x00")}
	o := runFake(t, f, "", "hook", "-in", "x", "-modules", ".")
	wantOut := strings.Join([]string{
		"gen:check", "plugins:labels:check", "fmt:check:files\tMODULE=.\tFILES=internal/kernel/money.go", "lint:config",
		"lint:go\tMODULE=.\tPKGS=./cmd/core ./internal/kernel", "vuln:go\tMODULE=.\tPKGS=./cmd/core ./internal/kernel",
		"test:go\tMODULE=.\tPKGS=./cmd/core ./internal/kernel", "test:script\tSCRIPT=scripts/tests/demo.test.sh",
		"docs:check", "demo:validate", "kibi:staged",
	}, "\n") + "\n"
	if diff := cmp.Diff(wantOut, o.stdout); diff != "" {
		t.Errorf("hook output mismatch (-want +got):\n%s", diff)
	}
	for _, want := range []string{"cmd/core: imports internal/kernel", "skip docs:build: "} {
		if !strings.Contains(o.stderr, want) {
			t.Errorf("hook report lacks %q:\n%s", want, o.stderr)
		}
	}
	f.in["changes"] = &fstest.MapFile{Data: []byte("M\x00go.mod\x00")}
	if o := runFake(t, f, "", "hook", "-in", "x"); o.stdout != "check:commit:full\n" || !strings.Contains(o.stderr, "go.mod: module dependencies") {
		t.Errorf("hook for go.mod = %+v; want the full check and why", o)
	}
	if o := runFake(t, f, "", "hook", "-full", "FULL is set"); o.stdout != "check:commit:full\n" || !strings.Contains(o.stderr, "FULL is set") {
		t.Errorf("hook -full = %+v; want the full check and why", o)
	}
}

func TestRunCI(t *testing.T) {
	t.Parallel()
	f := fakeRepo()
	tests := []struct {
		name, stdin, event, want string
	}{
		{"a push runs everything", "", "push", "code=true\n"},
		{"a docs-only pull request", "M\x00docs/x.md\x00M\x00.kb/symbols.yaml\x00", "pull_request", "code=false\n"},
		{"a code pull request", "M\x00docs/x.md\x00M\x00cmd/core/main.go\x00", "pull_request", "code=true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if o := runFake(t, f, tt.stdin, "ci", "-event", tt.event); o.stdout != tt.want {
				t.Errorf("ci = %q, want %q", o.stdout, tt.want)
			}
		})
	}
}

func TestRunImages(t *testing.T) {
	t.Parallel()
	const digest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	carried := `carry=[{"name":"core","image":"ghcr.io/example/openfief/core","digest":"` + digest + `","from":"bbbbbbbbbb"}]`
	built := `images=[{"name":"core","dockerfile":"cmd/core/Dockerfile","image":"ghcr.io/example/openfief/core"}]`
	tests := []struct {
		name, mode, builds, diff string
		want                     string
		report                   string
	}{
		{"reuse: unchanged is carried", "reuse", "core bbbbbbbbbb " + digest + "\n", "M\x00docs/x.md\x00", "images=[]\n" + carried + "\n", "carry core: no input changed since bbbbbbb"},
		{"reuse: changed is built", "reuse", "core bbbbbbbbbb " + digest + "\n", "M\x00internal/kernel/money.go\x00", built + "\ncarry=[]\n", "build core: internal/kernel/money.go is in internal/kernel"},
		{"reuse: no signed build", "reuse", "", "", built + "\ncarry=[]\n", "no signed sha-<commit> build"},
		{"base: unchanged is skipped", "base", "core bbbbbbbbbb -\n", "M\x00docs/x.md\x00", "images=[]\ncarry=[]\n", "skip core"},
		{"all: everything builds", "all", "", "", built + "\ncarry=[]\n", "every image builds"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := fakeRepo()
			if tt.builds != "" {
				f.in["builds"] = &fstest.MapFile{Data: []byte(tt.builds)}
				f.in["diff/bbbbbbbbbb"] = &fstest.MapFile{Data: []byte(tt.diff)}
			}
			o := runFake(t, f, "", "images", "-in", "x", "-mode", tt.mode)
			if o.stdout != tt.want || !strings.Contains(o.stderr, tt.report) {
				t.Errorf("images -mode %s = %+v; want %q and a report naming %q", tt.mode, o, tt.want, tt.report)
			}
		})
	}
}

func TestRunUsage(t *testing.T) {
	t.Parallel()
	f := fakeRepo()
	e := env{repo: f.repo, root: fakeRoot, dir: func(string) fs.FS { return f.in }, stdin: strings.NewReader(""), stdout: &bytes.Buffer{}, stderr: &bytes.Buffer{}}
	for _, args := range [][]string{nil, {"nope"}, {"hook", "-nope"}, {"ci", "-nope"}, {"images", "-nope"}, {"hook", "-in", "x"}} {
		if err := run(args, &e); err == nil {
			t.Errorf("run(%v) succeeded, want an error", args)
		}
	}
}
