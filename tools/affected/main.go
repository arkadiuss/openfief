// Command affected decides which checks a change needs, so the pre-commit hook and CI run only what the
// change can affect (docs/dev/pipeline.md#affected-checks). It is a development tool, not part of the
// product, and runs from the repository root through task targets:
//
//	affected hook -in DIR -modules "DIRS" [-full REASON]
//	affected ci -event NAME < changes
//	affected images -in DIR -modules "DIRS" -mode reuse|base|all
//
// hook prints the task targets the staged change needs; ci prints whether a CI run needs the end-to-end and
// PostgreSQL jobs; images prints which images CI builds and which keep their signed digest. Every run says
// what it selected and why on standard error.
//
// It never runs git, go, cosign or task itself: scripts/check-commit.sh and scripts/ci-images-select.sh
// collect their output into DIR, and run what it prints.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

var errUsage = errors.New("usage: affected hook|ci|images [flags]")

func main() {
	root, err := os.Getwd()
	if err == nil {
		e := env{repo: os.DirFS(root), root: root, dir: os.DirFS, stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr}
		err = run(os.Args[1:], &e)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "affected:", err)
		os.Exit(1)
	}
}

// env is what a command reads and writes besides its flags. repo is the repository, root its absolute path
// (go list prints absolute directories), and dir opens an input directory. Tests replace all of them.
type env struct {
	repo   fs.FS
	root   string
	dir    func(string) fs.FS
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// run dispatches a subcommand.
func run(args []string, e *env) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "hook":
		return runHook(args[1:], e)
	case "ci":
		return runCI(args[1:], e)
	case "images":
		return runImages(args[1:], e)
	default:
		return fmt.Errorf("%w: unknown command %q", errUsage, args[0])
	}
}
