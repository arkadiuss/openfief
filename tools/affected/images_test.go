package main

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var errNoTag = errors.New("no signed tag")

func TestChooseImages(t *testing.T) {
	t.Parallel()
	core := image{Name: "core", Dockerfile: "cmd/core/Dockerfile", Image: "ghcr.io/example/openfief/core"}
	nbp := image{Name: "nbp-fx", Dockerfile: "plugins/nbp-fx/Dockerfile", Image: "ghcr.io/example/openfief/nbp-fx"}
	ui := image{Name: "ui", Dockerfile: "internal/views/ui/Dockerfile", Image: "ghcr.io/example/openfief/ui"}
	const digest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	found := func(image) (signedBuild, error) { return signedBuild{commit: "aaaaaaaaaa", digest: digest}, nil }
	tests := []struct {
		name    string
		lookup  func(image) (signedBuild, error)
		changes []change
		want    []string // "build <name>" or "carry <name> <digest> <from>"
	}{
		{
			"docs only: every image keeps its digest", found,
			[]change{{path: "docs/x.md"}, {path: ".kb/symbols.yaml"}},
			[]string{"carry core", "carry nbp-fx", "carry ui"},
		},
		{
			"a package in one closure rebuilds that image", found,
			[]change{{path: "internal/ledger/ledger.go"}},
			[]string{"build core", "carry nbp-fx", "carry ui"},
		},
		{
			"a shared package rebuilds every image importing it", found,
			[]change{{path: "internal/kernel/money.go"}},
			[]string{"build core", "carry nbp-fx", "build ui"},
		},
		{
			"a test file changes no image", found,
			[]change{{path: "internal/kernel/money_test.go"}, {path: "internal/kernel/testdata/x.go"}},
			[]string{"carry core", "carry nbp-fx", "carry ui"},
		},
		{
			"a module in the SDK rebuilds the plugin on it", found,
			[]change{{path: "sdk/go/connector/server.go"}},
			[]string{"carry core", "build nbp-fx", "carry ui"},
		},
		{
			"an embedded file rebuilds the image embedding it", found,
			[]change{{path: "internal/views/ui/web/app.css"}},
			[]string{"carry core", "carry nbp-fx", "build ui"},
		},
		{
			"a Dockerfile rebuilds its image", found,
			[]change{{path: "plugins/nbp-fx/Dockerfile"}},
			[]string{"carry core", "build nbp-fx", "carry ui"},
		},
		{
			"go.sum rebuilds everything", found,
			[]change{{path: "go.sum"}},
			[]string{"build core", "build nbp-fx", "build ui"},
		},
		{
			"the images workflow rebuilds everything", found,
			[]change{{path: ".github/workflows/images.yml"}},
			[]string{"build core", "build nbp-fx", "build ui"},
		},
		{
			"a Go file in no package rebuilds everything", found,
			[]change{{path: "internal/gone/gone.go", deleted: true}},
			[]string{"build core", "build nbp-fx", "build ui"},
		},
		{
			"without a signed earlier build, it builds", func(image) (signedBuild, error) { return signedBuild{}, errNoTag }, nil,
			[]string{"build core", "build nbp-fx", "build ui"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			changesSince := func(rev string) ([]change, error) {
				if rev != "aaaaaaaaaa" {
					t.Errorf("changesSince(%q), want the signed build's commit", rev)
				}
				return tt.changes, nil
			}
			got := render(t, chooseImages([]image{core, nbp, ui}, tt.lookup, changesSince, testGraph()), digest)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("chooseImages() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// render lists the choices as "build <name>" or "carry <name>", checking that a carry keeps the signed build.
func render(t *testing.T, choices []imageChoice, digest string) []string {
	t.Helper()
	got := make([]string, 0, len(choices))
	for i := range choices {
		c := &choices[i]
		switch {
		case c.build:
			got = append(got, "build "+c.img.Name)
		case c.carry.Digest == digest && c.carry.From == "aaaaaaaaaa" && c.carry.Image == c.img.Image:
			got = append(got, "carry "+c.img.Name)
		default:
			t.Errorf("carry %+v doesn't carry the signed build", c.carry)
		}
	}
	return got
}

func TestChooseImagesBuildsWhenItCannotDiff(t *testing.T) {
	t.Parallel()
	img := image{Name: "core", Dockerfile: "cmd/core/Dockerfile", Image: "ghcr.io/example/openfief/core"}
	lookup := func(image) (signedBuild, error) { return signedBuild{commit: "gone"}, nil }
	failing := func(string) ([]change, error) { return nil, errNoTag }
	got := chooseImages([]image{img}, lookup, failing, testGraph())
	if len(got) != 1 || !got[0].build {
		t.Errorf("chooseImages() = %+v, want a build", got)
	}
}

func TestImageWithoutPackageBuilds(t *testing.T) {
	t.Parallel()
	img := image{Name: "x", Dockerfile: "cmd/x/Dockerfile", Image: "ghcr.io/example/openfief/x"}
	if why := imageChanged(img, nil, testGraph()); why == "" {
		t.Error("imageChanged() = \"\", want a reason to build")
	}
}
