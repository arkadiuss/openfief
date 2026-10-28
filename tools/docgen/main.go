// Command docgen generates the docs site's reference pages and checks that hand-written pages keep up with
// the code they cover (docs/dev/docs.md). It is a development tool, not part of the product, and runs from
// the repository root through task targets:
//
//	docgen reference -packages FILE -tasks FILE -coverage FILE -out DIR
//	docgen fresh [-message FILE] < changes
//	docgen refs -tasks FILE
//
// It never runs git itself: the Taskfile pipes git's output in.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var errUsage = errors.New("usage: docgen reference|fresh|refs [flags]")

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "docgen:", err)
		os.Exit(1)
	}
}

// run dispatches a subcommand against the repository in the working directory.
func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return errUsage
	}
	repo := os.DirFS(".")
	switch args[0] {
	case "reference":
		return runReference(repo, args[1:])
	case "fresh":
		return runFresh(repo, args[1:], stdin, stdout)
	case "refs":
		return runRefs(repo, args[1:], stdout)
	default:
		return fmt.Errorf("%w: unknown command %q", errUsage, args[0])
	}
}
