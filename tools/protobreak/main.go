// Command protobreak decides which breaking contract changes proto:breaking lets through
// (docs/dev/pipeline.md#breaking-contract-changes). A breaking change goes in only when an author decision
// names it; scripts/proto-breaking-exceptions.txt lists what each such decision removes, and this tool
// excuses exactly that, in the change that adds the entries, and nothing else. It is a development tool,
// not part of the product, and scripts/proto-breaking.sh runs it:
//
//	protobreak judge -findings FILE -exceptions FILE -base-exceptions FILE -design FILE -base-image FILE [-tags FILE]
//	protobreak marked [-before FILE -after FILE | -breaking] -message FILE
//	protobreak stage < tags
//
// judge reads buf breaking's JSON findings against the base, the exceptions file now and at the base, the
// decisions (docs/DESIGN.md) and the base's contracts as a buf image, and fails unless every finding is
// excused by an entry new since the base, every entry is valid and every new entry excuses a finding.
// marked fails when a commit adds an entry, or with -breaking breaks a contract, but its message doesn't
// mark it breaking (Conventional Commits: `!` or a `BREAKING CHANGE:` footer). stage prints pre-beta while
// the repository's tags hold no beta or stable release, when a commit marked breaking may break the
// contracts without an author decision (docs/dev/sdlc.md#releases), and beta from then on.
//
// It never runs git or buf itself: the script collects their output into files.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// errUsage reports a missing or unknown subcommand or flag.
var errUsage = errors.New("usage: protobreak judge|marked|stage [flags]")

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "protobreak:", err)
		os.Exit(1)
	}
}

// run dispatches a subcommand, writing what it excused to out.
func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "judge":
		return runJudge(args[1:], out)
	case "marked":
		return runMarked(args[1:])
	case "stage":
		return runStage(os.Stdin, out)
	default:
		return fmt.Errorf("%w: unknown command %q", errUsage, args[0])
	}
}

// runJudge reads judge's inputs from the files its flags name and judges them.
func runJudge(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("judge", flag.ContinueOnError)
	findings := fs.String("findings", "", "buf breaking's findings, one JSON object a line")
	exceptions := fs.String("exceptions", "", "the exceptions file now; a missing file has no entries")
	base := fs.String("base-exceptions", "", "the exceptions file at the base; a missing file has no entries")
	design := fs.String("design", "", "docs/DESIGN.md")
	image := fs.String("base-image", "", "the base's contracts, built with buf build -o FILE.binpb")
	tags := fs.String("tags", "", "the repository's tags, one a line; none is before the first beta")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}
	in, err := readJudgeInput(*findings, *exceptions, *base, *design, *image)
	if err != nil {
		return err
	}
	raw, err := readFiles(*tags)
	if err != nil {
		return err
	}
	in.PreBeta = preBeta(raw[0])
	v, err := judge(&in)
	if err != nil {
		return err
	}
	return writeVerdict(out, &v)
}

// writeVerdict prints what judge let through, one line each.
func writeVerdict(out io.Writer, v *verdict) error {
	lines := make([]string, 0, len(v.Excused)+len(v.Allowed))
	for _, e := range v.Excused {
		lines = append(lines, fmt.Sprintf("proto:breaking: %s excuses %s %s", e.Decision, e.Rule, e.Element))
	}
	for _, f := range v.Allowed {
		lines = append(lines, "proto:breaking: allowed before the first beta in a commit marked breaking: "+f.String())
	}
	for _, l := range lines {
		if _, err := fmt.Fprintln(out, l); err != nil {
			return fmt.Errorf("writing the verdict: %w", err)
		}
	}
	return nil
}

// runStage prints pre-beta or beta for the tags in, one a line (preBeta).
func runStage(in io.Reader, out io.Writer) error {
	tags, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("reading the tags: %w", err)
	}
	stage := "beta"
	if preBeta(string(tags)) {
		stage = "pre-beta"
	}
	if _, err := fmt.Fprintln(out, stage); err != nil {
		return fmt.Errorf("writing the stage: %w", err)
	}
	return nil
}

// readJudgeInput reads and parses judge's input files.
func readJudgeInput(findingsPath, exceptionsPath, basePath, designPath, imagePath string) (judgeInput, error) {
	var in judgeInput
	raw, err := readFiles(findingsPath, exceptionsPath, basePath, designPath, imagePath)
	if err != nil {
		return judgeInput{}, err
	}
	if in.Findings, err = parseFindings(raw[0]); err != nil {
		return judgeInput{}, err
	}
	if in.Exceptions, err = parseExceptions(raw[1]); err != nil {
		return judgeInput{}, fmt.Errorf("%s: %w", exceptionsPath, err)
	}
	if in.BaseExceptions, err = parseExceptions(raw[2]); err != nil {
		return judgeInput{}, fmt.Errorf("%s at the base: %w", basePath, err)
	}
	in.Decisions = parseDecisions(raw[3])
	var set descriptorpb.FileDescriptorSet
	if err := (proto.UnmarshalOptions{DiscardUnknown: true}).Unmarshal([]byte(raw[4]), &set); err != nil {
		return judgeInput{}, fmt.Errorf("reading the base image %s: %w", imagePath, err)
	}
	in.Base = indexContracts(&set)
	return in, nil
}

// runMarked fails when the commit whose message -message holds adds an exceptions entry without marking
// itself breaking.
func runMarked(args []string) error {
	fs := flag.NewFlagSet("marked", flag.ContinueOnError)
	before := fs.String("before", "", "the exceptions file before the commit; a missing file has no entries")
	after := fs.String("after", "", "the exceptions file the commit leaves; a missing file has no entries")
	breaking := fs.Bool("breaking", false, "the commit breaks a contract, so it must be marked whatever the exceptions")
	message := fs.String("message", "", "the commit message")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}
	raw, err := readFiles(*before, *after, *message)
	if err != nil {
		return err
	}
	if *breaking && !markedBreaking(raw[2]) {
		return fmt.Errorf("%w (`!` or a BREAKING CHANGE: footer): it breaks a contract", errUnmarked)
	}
	return checkMarked(raw[0], raw[1], raw[2])
}

// readFiles returns each named file's contents; an empty name or a missing file reads as empty, since a
// base or a commit may have no exceptions file.
func readFiles(paths ...string) ([]string, error) {
	out := make([]string, len(paths))
	for i, p := range paths {
		if p == "" {
			continue
		}
		b, err := os.ReadFile(filepath.Clean(p))
		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return nil, fmt.Errorf("reading %s: %w", p, err)
		default:
			out[i] = string(b)
		}
	}
	return out, nil
}
