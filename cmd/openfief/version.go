package main

import (
	"fmt"
	"io"
	"runtime/debug"

	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// version is the release this CLI was built for. The release stamps it with
// -ldflags "-X main.version=<version>" (scripts/release-cli.sh); any other
// build is "dev".
var version = "dev"

// unknownCommit stands for a build without version control information, such
// as a test binary.
const unknownCommit = "unknown"

// runVersion is `openfief version`: it prints versionLine for this binary,
// with the build settings Go recorded in it (-buildvcs), and takes no
// arguments.
func runVersion(args []string, out io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("%w: openfief version takes no arguments\n%s", cli.ErrUsage, cli.Usage)
	}
	var settings []debug.BuildSetting
	if info, ok := debug.ReadBuildInfo(); ok {
		settings = info.Settings
	}
	if _, err := fmt.Fprintln(out, versionLine(version, settings)); err != nil {
		return fmt.Errorf("version: %w", err)
	}
	return nil
}

// versionLine is "openfief <version> commit <revision>": the revision is the
// build's vcs.revision setting, "unknown" without one, and ends in
// "-modified" when the working tree had uncommitted changes. So the owner can
// tell which release, and which commit of it, a downloaded binary is.
func versionLine(v string, settings []debug.BuildSetting) string {
	commit, modified := unknownCommit, false
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			if s.Value != "" {
				commit = s.Value
			}
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if modified && commit != unknownCommit {
		commit += "-modified"
	}
	return "openfief " + v + " commit " + commit
}
