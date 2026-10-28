// Command sigstore-trusted-root fetches the Sigstore public-good trusted root
// (trusted_root.json) through TUF and writes it where the operator embeds it
// (D-56). The TUF client starts from the root.json that sigstore-go embeds, so
// the file is checked against the chain of signed TUF metadata before it is
// written; nothing is taken on trust from the network. Run it through
// `task trustroot:update`, review the diff, and commit it.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
)

const (
	// trustedRootTarget is the TUF target that holds the trusted root.
	trustedRootTarget = "trusted_root.json"
	// filePerm is the written file's mode; git records it as 0644.
	filePerm = 0o644 // a committed file, not a secret
)

// errNoOutput reports a missing -o flag.
var errNoOutput = errors.New("-o is required")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "sigstore-trusted-root:", err)
		os.Exit(1)
	}
}

// run fetches the trusted root through TUF, checks that it parses, and
// writes it to the -o file.
func run(args []string) error {
	fs := flag.NewFlagSet("sigstore-trusted-root", flag.ContinueOnError)
	out := fs.String("o", "", "file to write the trusted root to")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("flags: %w", err)
	}
	if *out == "" {
		return errNoOutput
	}
	cache, err := os.MkdirTemp("", "sigstore-tuf-")
	if err != nil {
		return fmt.Errorf("TUF cache: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(cache); removeErr != nil {
			fmt.Fprintln(os.Stderr, "sigstore-trusted-root: remove TUF cache:", removeErr)
		}
	}()
	client, err := tuf.New(tuf.DefaultOptions().WithCachePath(cache))
	if err != nil {
		return fmt.Errorf("TUF client: %w", err)
	}
	data, err := client.GetTarget(trustedRootTarget)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", trustedRootTarget, err)
	}
	if _, err := root.NewTrustedRootFromJSON(data); err != nil {
		return fmt.Errorf("parse %s: %w", trustedRootTarget, err)
	}
	if err := os.WriteFile(*out, data, filePerm); err != nil {
		return fmt.Errorf("write %s: %w", *out, err)
	}
	return nil
}
