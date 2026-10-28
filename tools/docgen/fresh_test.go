package main

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMatchGlob(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern, name string
		want          bool
	}{
		{"internal/ledger/**", "internal/ledger/ledger.go", true},
		{"internal/ledger/**", "internal/ledger/app/import.go", true},
		{"internal/ledger/**", "internal/ledgerx/a.go", false},
		{"internal/plugins/*.go", "internal/plugins/manifest.go", true},
		{"internal/plugins/*.go", "internal/plugins/app/plan.go", false},
		{"Taskfile.yml", "Taskfile.yml", true},
		{"Taskfile.yml", "site/Taskfile.yml", false},
		{"**/doc.go", "doc.go", true},
		{".github/**", ".github/workflows/ci.yml", true},
		{"deploy/crds/**", "deploy/crds", true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+" "+tt.name, func(t *testing.T) {
			t.Parallel()
			if got := matchGlob(tt.pattern, tt.name); got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestParseLog(t *testing.T) {
	t.Parallel()
	log := "\x1eaaaaaaaaaaaaaaaa\x1ffeat: one\n\nDocs-Unchanged: rename only\n\x1f\n\ninternal/a/a.go\ninternal/a/b.go\n" +
		"\x1ebbbb\x1ffix: two\n\nDocs-Unchanged:\n\x1f\n\ninternal/b/b.go\n"
	got := parseLog(log)
	want := []commit{
		{id: "aaaaaaaaaaaa", files: []string{"internal/a/a.go", "internal/a/b.go"}, exempt: true},
		{id: "bbbb", files: []string{"internal/b/b.go"}, exempt: false},
	}
	if len(got) != len(want) {
		t.Fatalf("parseLog() = %d commits, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].id != want[i].id || got[i].exempt != want[i].exempt || !slices.Equal(got[i].files, want[i].files) {
			t.Errorf("commit %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestStaleFindings(t *testing.T) {
	t.Parallel()
	pages := []page{
		{path: "docs/concepts/ledger.md", covers: []string{"internal/ledger/**"}},
		{path: "docs/architecture.md", covers: []string{"cmd/**", "internal/ledger/postgres/**"}},
	}
	tests := []struct {
		name    string
		commits []commit
		want    []string
	}{
		{"covered code without its page fails", []commit{{id: "c1", files: []string{"internal/ledger/ledger.go"}}}, []string{"internal/ledger/ledger.go"}},
		{"page changed in the same commit", []commit{{id: "c1", files: []string{"internal/ledger/ledger.go", "docs/concepts/ledger.md"}}}, nil},
		{"page changed later in the push", []commit{{id: "c1", files: []string{"internal/ledger/ledger.go"}}, {id: "c2", files: []string{"docs/concepts/ledger.md"}}}, nil},
		{"any covering page counts", []commit{{id: "c1", files: []string{"internal/ledger/postgres/store.go", "docs/architecture.md"}}}, nil},
		{"trailer exempts the commit", []commit{{id: "c1", files: []string{"internal/ledger/ledger.go"}, exempt: true}}, nil},
		{"tests never count", []commit{{id: "c1", files: []string{"internal/ledger/ledger_test.go", "internal/ledger/testdata/x.json"}}}, nil},
		{"uncovered files never count", []commit{{id: "c1", files: []string{"README.md"}}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			for _, f := range staleFindings(pages, tt.commits) {
				got = append(got, f.file)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("staleFindings() files = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUncoveredPackages(t *testing.T) {
	t.Parallel()
	pages := []page{{path: "docs/a.md", covers: []string{"internal/ledger/**", "internal/plugins/*.go"}}}
	files := []string{"internal/ledger/a.go", "internal/ledger/app/b.go", "internal/plugins/c.go", "internal/plugins/kube/d.go", "tools/docgen/main.go"}
	want := []string{"internal/plugins/kube", "tools/docgen"}
	if got := uncoveredPackages(pages, files); !slices.Equal(got, want) {
		t.Errorf("uncoveredPackages() = %v, want %v", got, want)
	}
}

func TestParseFrontMatter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		covers  []string
		wantErr bool
	}{
		{"covers list", "---\ntitle: Ledger\ncovers:\n  - internal/ledger/**\n---\n\nBody\n", []string{"internal/ledger/**"}, false},
		{"no front matter", "# Design\n\nBody\n", nil, false},
		{"unterminated", "---\ntitle: x\n", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fm, err := parseFrontMatter([]byte(tt.in))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseFrontMatter() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !slices.Equal(fm.Covers, tt.covers) {
				t.Errorf("covers = %v, want %v", fm.Covers, tt.covers)
			}
		})
	}
}

func TestRunFreshStaged(t *testing.T) {
	t.Parallel()
	repo := fstest.MapFS{
		"docs/concepts/ledger.md":   {Data: []byte("---\ncovers:\n  - internal/ledger/**\n---\n")},
		"internal/ledger/ledger.go": {Data: []byte("package ledger\n")},
		"msg-plain":                 {Data: []byte("feat: change\n")},
		"msg-exempt":                {Data: []byte("refactor: rename\n\nDocs-Unchanged: rename only\n")},
	}
	tests := []struct {
		name      string
		message   string
		wantStale bool
	}{
		{"covered change without page fails", "msg-plain", true},
		{"trailer passes", "msg-exempt", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			err := runFresh(repo, []string{"-message", tt.message}, strings.NewReader("internal/ledger/ledger.go\n"), &out)
			if got := errors.Is(err, errStale); got != tt.wantStale {
				t.Errorf("runFresh() error = %v, want stale %v; output:\n%s", err, tt.wantStale, out.String())
			}
		})
	}
}
