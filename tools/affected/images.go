package main

import (
	"path"
	"slices"
	"strings"
)

// image is one entry of the matrix `task ci:images` prints: a name, its Dockerfile and its registry reference.
type image struct {
	Name       string `json:"name"`
	Dockerfile string `json:"dockerfile"`
	Image      string `json:"image"`
}

// carried is an unchanged image that keeps the digest signed for an earlier commit; CI tags that digest
// for this commit instead of building it again.
type carried struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Digest string `json:"digest"`
	// From is the commit whose sha-<commit> tag carries the signed digest.
	From string `json:"from"`
}

// signedBuild is the newest earlier build of an image whose signature verified: its commit and digest.
type signedBuild struct {
	commit string
	digest string
}

// imageChoice is one image's decision with the reason the report prints.
type imageChoice struct {
	img   image
	build bool
	carry carried
	why   string
}

// imageInputs are the paths every image is built from besides its own package closure and Dockerfile:
// module files the Dockerfiles copy, the workflow and action that build them, and this selector.
func imageInputs() []string {
	return []string{
		"go.mod", "go.sum", "go.work", "go.work.sum", "sdk/go/go.mod", "sdk/go/go.sum", ".dockerignore",
		".github/workflows/images.yml", ".github/actions/setup", "scripts/ci-images.sh", "tools/affected",
	}
}

// imageChanged says why a change set can change an image, or "" when it can't: a shared build input, its Dockerfile, a
// non-test Go file in a package its binary imports (directly or not), or a file one of those packages
// embeds. A Go file in a directory that is no package any more can't be placed, so it rebuilds everything.
func imageChanged(img image, changes []change, g *graph) string {
	mainDir := path.Dir(img.Dockerfile)
	mainPkg, ok := g.byDir[mainDir]
	if !ok {
		return "no Go package at " + mainDir
	}
	closure := g.dependencies(mainPkg)
	for _, c := range changes {
		if why := imageHit(c, img, closure, g); why != "" {
			return why
		}
	}
	return ""
}

// imageHit says why one changed path is an input of the image, or "" when it isn't.
func imageHit(c change, img image, closure map[string]bool, g *graph) string {
	switch {
	case slices.ContainsFunc(imageInputs(), func(in string) bool { return within(c.path, in) }):
		return c.path + " is a shared build input"
	case c.path == img.Dockerfile:
		return "its Dockerfile changed"
	case isGoSource(c.path):
		ip, ok := g.byDir[path.Dir(c.path)]
		if !ok {
			return c.path + " is in no package; when in doubt, rebuild"
		}
		if closure[ip] {
			return c.path + " is in " + shortPath(ip) + ", which it imports"
		}
	default:
		for _, ip := range g.embedOwners[c.path] {
			if closure[ip] && slices.Contains(g.pkgs[ip].embeds, c.path) {
				return c.path + " is embedded in " + shortPath(ip)
			}
		}
	}
	return ""
}

// isGoSource is a Go file that can be compiled into a binary: not a test, not test data.
func isGoSource(file string) bool {
	return path.Ext(file) == ".go" && !strings.HasSuffix(file, "_test.go") && !strings.Contains(file, "/testdata/")
}

// chooseImages decides for each image whether CI builds it. Without an earlier build to compare with
// (lookup says why), or without a way to diff, it builds. With one, it builds only if the commits since that build changed an
// input of the image, and otherwise carries the signed digest forward.
func chooseImages(images []image, lookup func(image) (signedBuild, error), changesSince func(string) ([]change, error), g *graph) []imageChoice {
	out := make([]imageChoice, 0, len(images))
	for _, img := range images {
		prev, err := lookup(img)
		if err != nil {
			out = append(out, imageChoice{img: img, build: true, why: err.Error()})
			continue
		}
		changes, err := changesSince(prev.commit)
		if err != nil {
			out = append(out, imageChoice{img: img, build: true, why: "can't diff against " + prev.commit + ": " + err.Error()})
			continue
		}
		if why := imageChanged(img, changes, g); why != "" {
			out = append(out, imageChoice{img: img, build: true, why: why + " since " + short(prev.commit)})
			continue
		}
		out = append(out, imageChoice{
			img:   img,
			carry: carried{Name: img.Name, Image: img.Image, Digest: prev.digest, From: prev.commit},
			why:   "no input changed since " + short(prev.commit),
		})
	}
	return out
}

// short is a commit's 7-character abbreviation, as the sha-<commit> image tags use it.
func short(commit string) string {
	const shortSHA = 7
	if len(commit) < shortSHA {
		return commit
	}
	return commit[:shortSHA]
}
