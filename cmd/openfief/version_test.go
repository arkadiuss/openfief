package main

import (
	"bytes"
	"errors"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins/cli"
)

// rev is a synthetic 40-hex commit.
const rev = "0123456789abcdef0123456789abcdef01234567"

func TestVersionLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		version  string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name:    "a release build names its version and commit",
			version: "2026.11.0-alpha.2",
			settings: []debug.BuildSetting{
				{Key: "vcs", Value: "git"}, {Key: "vcs.revision", Value: rev}, {Key: "vcs.modified", Value: "false"},
			},
			want: "openfief 2026.11.0-alpha.2 commit " + rev,
		},
		{
			name:     "a local build is dev",
			version:  "dev",
			settings: []debug.BuildSetting{{Key: "vcs.revision", Value: rev}},
			want:     "openfief dev commit " + rev,
		},
		{
			name:     "uncommitted changes are marked",
			version:  "dev",
			settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "true"}, {Key: "vcs.revision", Value: rev}},
			want:     "openfief dev commit " + rev + "-modified",
		},
		{
			name:     "no version control information",
			version:  "dev",
			settings: nil,
			want:     "openfief dev commit unknown",
		},
		{
			name:     "an empty revision is unknown and never marked modified",
			version:  "dev",
			settings: []debug.BuildSetting{{Key: "vcs.revision", Value: ""}, {Key: "vcs.modified", Value: "true"}},
			want:     "openfief dev commit unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, versionLine(tt.version, tt.settings)); diff != "" {
				t.Errorf("versionLine (-want +got):\n%s", diff)
			}
		})
	}
}

// TestVersionCommand runs `openfief version` through the command line's
// entry point: it needs no cluster or catalog, and a test binary isn't
// stamped, so it is dev.
func TestVersionCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{name: "prints one line", args: nil, want: "openfief dev commit "},
		{name: "refuses arguments", args: []string{"extra"}, wantErr: cli.ErrUsage},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			err := runVersion(tt.args, &out)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("openfief version %v: error %v, want %v", tt.args, err, tt.wantErr)
			}
			got := out.String()
			if tt.want == "" && got != "" {
				t.Errorf("openfief version %v printed %q, want nothing", tt.args, got)
			}
			if tt.want != "" && (!strings.HasPrefix(got, tt.want) || strings.Count(got, "\n") != 1) {
				t.Errorf("openfief version printed %q, want one line starting %q", got, tt.want)
			}
		})
	}
}
