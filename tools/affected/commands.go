package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

var errNoBuild = errors.New("no earlier build to compare with")

// fullCheckTask is the task target with every step of the hook, run when the plan is full.
const fullCheckTask = "check:commit:full"

// runHook plans the staged change and prints the task targets to run, one per line (task, then variables,
// tab-separated), for scripts/check-commit.sh to run in order. -full names why everything runs.
func runHook(args []string, e *env) error {
	flags := flag.NewFlagSet("hook", flag.ContinueOnError)
	flags.SetOutput(e.stderr)
	in := flags.String("in", "", "the directory scripts/check-commit.sh wrote the changes and package listings to")
	modules := flags.String("modules", ".", "the Go module directories, space-separated")
	full := flags.String("full", "", "run the full check, for this reason")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("hook: %w", err)
	}
	p := plan{full: true, reasons: []string{*full}}
	if *full == "" {
		var err error
		if p, err = planStaged(e, e.dir(*in), strings.Fields(*modules)); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(e.stderr, writePlan(&p)); err != nil {
		return fmt.Errorf("hook: %w", err)
	}
	lines := []string{fullCheckTask}
	if !p.full {
		lines = lines[:0]
		for i := range p.steps {
			lines = append(lines, planLine(&p.steps[i]))
		}
	}
	return writeLines(e.stdout, lines)
}

// planStaged plans the staged change from the inputs.
func planStaged(e *env, in fs.FS, modules []string) (plan, error) {
	changes, err := readChanges(in, "changes")
	if err != nil {
		return plan{}, err
	}
	g, err := loadGraph(e.repo, in, e.root, modules)
	if err != nil {
		return plan{}, err
	}
	scripts, err := loadScriptTests(e.repo)
	if err != nil {
		return plan{}, err
	}
	return hookPlan(changes, g, scripts, modules), nil
}

// writePlan renders what the hook runs and why, and what it leaves out.
func writePlan(p *plan) string {
	var b strings.Builder
	if p.full {
		b.WriteString("affected: full check (task " + fullCheckTask + "):\n")
		for _, r := range p.reasons {
			b.WriteString("  " + r + "\n")
		}
		return b.String()
	}
	b.WriteString("affected: checks for the staged change\n")
	for _, r := range p.reasons {
		b.WriteString("  changed " + r + "\n")
	}
	for i := range p.steps {
		s := &p.steps[i]
		b.WriteString("  run  " + strings.Join(append([]string{s.task}, s.vars...), " ") + ": " + s.why + "\n")
		for _, d := range s.details {
			b.WriteString("         " + d + "\n")
		}
	}
	for _, s := range p.skipped {
		b.WriteString("  skip " + s.task + ": " + s.why + "\n")
	}
	return b.String()
}

// writeLines writes each line followed by a newline.
func writeLines(w io.Writer, lines []string) error {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	if _, err := io.WriteString(w, b.String()); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// runCI reads `git diff --name-status --no-renames -z <base> HEAD` on standard input and prints code=true when a
// CI run needs the end-to-end and PostgreSQL jobs: always outside a pull request (a push to main runs the whole
// suite), and for a pull request that changes anything other than documentation and requirements.
func runCI(args []string, e *env) error {
	flags := flag.NewFlagSet("ci", flag.ContinueOnError)
	flags.SetOutput(e.stderr)
	event := flags.String("event", "", "the GitHub event that started the run")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("ci: %w", err)
	}
	d := ciDecision{code: true, why: "a " + *event + " run runs the whole suite"}
	if *event == "pull_request" {
		data, err := io.ReadAll(e.stdin)
		if err != nil {
			return fmt.Errorf("ci: %w", err)
		}
		d = needsCode(parseNameStatus(data))
	}
	if _, err := io.WriteString(e.stderr, "affected: "+d.why+"\n"); err != nil {
		return fmt.Errorf("ci: %w", err)
	}
	return writeLines(e.stdout, []string{fmt.Sprintf("code=%t", d.code)})
}

// ciDecision is whether a CI run needs the end-to-end and PostgreSQL jobs, and why.
type ciDecision struct {
	code bool
	why  string
}

// needsCode decides whether a change set reaches beyond documentation and requirements.
func needsCode(changes []change) ciDecision {
	for _, c := range changes {
		if !docsOnly(c.path) {
			return ciDecision{code: true, why: c.path + " can change code, images or manifests: the end-to-end and PostgreSQL jobs run"}
		}
	}
	return ciDecision{why: fmt.Sprintf("all %d changed paths are documentation or requirements: the end-to-end and PostgreSQL jobs are skipped", len(changes))}
}

// runImages reads what scripts/ci-images-select.sh wrote (the image matrix, each image's earlier build and the
// diff since it) and prints two GitHub outputs: images, the ones to build, and carry, the unchanged ones whose
// signed digest CI tags for this commit. -mode says how the earlier builds were found: reuse (the newest
// signed sha-<commit> build; unchanged images are carried), base (a pull request's base; unchanged images are
// skipped) or all (no comparison; everything builds).
func runImages(args []string, e *env) error {
	flags := flag.NewFlagSet("images", flag.ContinueOnError)
	flags.SetOutput(e.stderr)
	in := flags.String("in", "", "the directory scripts/ci-images-select.sh wrote its findings to")
	modules := flags.String("modules", ".", "the Go module directories, space-separated")
	mode := flags.String("mode", "all", "reuse, base or all")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("images: %w", err)
	}
	dir := e.dir(*in)
	images, err := readMatrix(dir)
	if err != nil {
		return err
	}
	builds, err := readBuilds(dir)
	if err != nil {
		return err
	}
	g, err := loadGraph(e.repo, dir, e.root, strings.Fields(*modules))
	if err != nil {
		return err
	}
	lookup := func(img image) (signedBuild, error) {
		if b, ok := builds[img.Name]; ok {
			return b, nil
		}
		if *mode == "reuse" {
			return signedBuild{}, fmt.Errorf("%w: no signed sha-<commit> build among the recent commits", errNoBuild)
		}
		return signedBuild{}, fmt.Errorf("%w: every image builds", errNoBuild)
	}
	changesSince := func(commit string) ([]change, error) { return readChanges(dir, "diff/"+commit) }
	return writeImages(e, chooseImages(images, lookup, changesSince, g), *mode == "reuse")
}

// writeImages prints the decisions on standard error and the two outputs on standard output.
func writeImages(e *env, choices []imageChoice, reuse bool) error {
	var report strings.Builder
	build, carry := []image{}, []carried{}
	for i := range choices {
		c := &choices[i]
		switch {
		case c.build:
			build = append(build, c.img)
			report.WriteString("affected: build " + c.img.Name + ": " + c.why + "\n")
		case reuse:
			carry = append(carry, c.carry)
			report.WriteString("affected: carry " + c.img.Name + ": " + c.why + "; keeps " + c.carry.Digest + ", signed for sha-" + short(c.carry.From) + "\n")
		default:
			report.WriteString("affected: skip " + c.img.Name + ": " + c.why + "\n")
		}
	}
	if _, err := io.WriteString(e.stderr, report.String()); err != nil {
		return fmt.Errorf("images: %w", err)
	}
	images, err := json.Marshal(build)
	if err != nil {
		return fmt.Errorf("images: %w", err)
	}
	carried, err := json.Marshal(carry)
	if err != nil {
		return fmt.Errorf("images: %w", err)
	}
	return writeLines(e.stdout, []string{"images=" + string(images), "carry=" + string(carried)})
}
