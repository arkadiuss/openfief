package ui

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/design"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

var (
	// colorLiteral matches a color written out instead of taken from a
	// token: a hex code, or an rgb, hsl, hwb, lab, lch or oklch function.
	colorLiteral = regexp.MustCompile(`#[0-9A-Fa-f]{3,8}\b|\b(?:rgba?|hsla?|hwb|lab|lch|oklab|oklch)\(`)
	// customProperty matches the name of every custom property a stylesheet
	// uses, and definedProperty every one it defines.
	customProperty  = regexp.MustCompile(`var\((--[a-z0-9-]+)`)
	definedProperty = regexp.MustCompile(`(?m)^\s*(--[a-z0-9-]+)\s*:`)
)

// TestNoColorsOutsideTokens checks the design system's first rule: the
// UI's templates, scripts and stylesheet write no color of their own, so
// every color comes from design/tokens.css. The vendored libraries are not
// the UI's own.
func TestNoColorsOutsideTokens(t *testing.T) {
	t.Parallel()
	checked := 0
	err := fs.WalkDir(webFS, "web", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.HasPrefix(name, "web/static/vendor/") {
			return err
		}
		data, err := webFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		checked++
		for _, m := range colorLiteral.FindAllString(string(data), -1) {
			t.Errorf("%s writes the color %q: take it from a token in design/tokens.css", name, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no template, script or stylesheet checked")
	}
}

// TestStylesheetUsesDefinedTokens checks that every custom property the
// UI's stylesheet uses is a design token or one the stylesheet defines from
// tokens, so a misspelled token can't silently fall back to nothing.
func TestStylesheetUsesDefinedTokens(t *testing.T) {
	t.Parallel()
	tokens, err := design.Files.ReadFile("tokens.css")
	if err != nil {
		t.Fatal(err)
	}
	style, err := webFS.ReadFile("web/static/style.css")
	if err != nil {
		t.Fatal(err)
	}
	defined := map[string]bool{}
	for _, src := range [][]byte{tokens, style} {
		for _, m := range definedProperty.FindAllStringSubmatch(string(src), -1) {
			defined[m[1]] = true
		}
	}
	for _, m := range customProperty.FindAllStringSubmatch(string(style), -1) {
		if !defined[m[1]] {
			t.Errorf("style.css uses %s, which neither tokens.css nor style.css defines", m[1])
		}
	}
}

// TestEveryPageLoadsTheDesignSystem checks that every page's layout loads
// the Gate's favicon, the fonts, the tokens and then the UI's stylesheet,
// in that order, from the UI itself.
func TestEveryPageLoadsTheDesignSystem(t *testing.T) {
	t.Parallel()
	p, err := parsePages()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`href="/static/design/marks/favicon.svg"`, `href="/static/design/fonts/fonts.css"`,
		`href="/static/design/tokens.css"`, `href="/static/style.css"`,
	}
	for name, set := range p.sets {
		layout := set.Lookup("layout.html")
		if layout == nil {
			t.Errorf("page %s has no layout", name)
			continue
		}
		src := layout.Tree.Root.String()
		last := -1
		for _, w := range want {
			i := strings.Index(src, w)
			if i < 0 {
				t.Errorf("page %s doesn't load %s", name, w)
				continue
			}
			if i < last {
				t.Errorf("page %s loads %s out of order: favicon, fonts, tokens, then style.css", name, w)
			}
			last = i
		}
	}
}

// TestDesignFilesServed checks that the UI serves the design system's
// files from design/ itself, with their content types: the tokens, the
// fonts' stylesheet and a font, a mark and Kasztan, and that every font
// fonts.css names is there.
func TestDesignFilesServed(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	files := []struct{ file, ctype string }{
		{"tokens.css", "text/css; charset=utf-8"},
		{"fonts/fonts.css", "text/css; charset=utf-8"},
		{"fonts/ibm-plex-sans-latin-400-normal.woff2", "font/woff2"},
		{"fonts/bricolage-grotesque-latin-ext-wght-normal.woff2", "font/woff2"},
		{"marks/favicon.svg", "image/svg+xml"},
		{"mascot/kasztan-asleep.svg", "image/svg+xml"},
	}
	for _, f := range files {
		rec := h.serve(t, request(http.MethodGet, "/static/design/"+f.file, ""))
		wantStatus(t, rec, http.StatusOK)
		if got := rec.Header().Get("Content-Type"); got != f.ctype {
			t.Errorf("%s: Content-Type = %q, want %q", f.file, got, f.ctype)
		}
	}
	css, err := design.Files.ReadFile("fonts/fonts.css")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range regexp.MustCompile(`url\("([^"]+)"\)`).FindAllStringSubmatch(string(css), -1) {
		if _, err := fs.Stat(design.Files, path.Join("fonts", m[1])); err != nil {
			t.Errorf("fonts.css names %s, which design/fonts lacks", m[1])
		}
	}
}

// TestKasztanPoses checks that every pose a template shows is one of the
// four the design system draws.
func TestKasztanPoses(t *testing.T) {
	t.Parallel()
	poses := regexp.MustCompile(`template "kasztan" "([a-z-]+)"`)
	var used []string
	err := fs.WalkDir(webFS, "web/templates", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := webFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		for _, m := range poses.FindAllStringSubmatch(string(data), -1) {
			used = append(used, m[1])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(used)
	used = slices.Compact(used)
	if diff := cmp.Diff([]string{"asleep", "drawbridge", "on-duty", "seal"}, used); diff != "" {
		t.Errorf("poses (-want +got):\n%s", diff)
	}
	for _, pose := range used {
		if _, err := fs.Stat(design.Files, "mascot/kasztan-"+pose+".svg"); err != nil {
			t.Errorf("pose %s: %v", pose, err)
		}
	}
}

// TestTrustBadges checks D-87's four badges: every signature trust the read
// API reports falls into exactly one, and an unknown one is Blocked.
func TestTrustBadges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		trust viewsv1.SignatureTrust
		want  trustBadge
	}{
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_VERIFIED, trustBadge{Label: "Verified", Tone: "ok"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_OWNER, trustBadge{Label: "Yours", Tone: "neutral"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_OPENFIEF_CI, trustBadge{Label: "Unreviewed", Tone: "warn"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_PUBLISHER, trustBadge{Label: "Unreviewed", Tone: "warn"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_TRUSTED_SIGNER, trustBadge{Label: "Unreviewed", Tone: "warn"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_REFUSED, trustBadge{Label: "Blocked", Tone: "danger"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_NOT_CHECKED, trustBadge{Label: "Blocked", Tone: "danger"}},
		{viewsv1.SignatureTrust_SIGNATURE_TRUST_UNSPECIFIED, trustBadge{Label: "Blocked", Tone: "danger"}},
		{viewsv1.SignatureTrust(99), trustBadge{Label: "Blocked", Tone: "danger"}},
	}
	for _, tt := range tests {
		t.Run(tt.trust.String(), func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, badgeOf(tt.trust)); diff != "" {
				t.Errorf("badgeOf (-want +got):\n%s", diff)
			}
		})
	}
}
