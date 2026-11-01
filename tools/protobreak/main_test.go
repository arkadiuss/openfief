package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
)

// writeInputs writes judge's and marked's input files to a temporary directory and returns a function
// naming a file in it.
func writeInputs(t *testing.T) func(string) string {
	t.Helper()
	dir := t.TempDir()
	image, err := proto.Marshal(baseSet())
	if err != nil {
		t.Fatalf("marshalling the base image: %v", err)
	}
	files := map[string]string{
		"findings": rpcGone + "\n" + requestGone + "\n", "exceptions": rpcEntry + "\n" + requestEntry + "\n",
		"design": design, "message": "feat(views)!: remove ListGadgets (D-9)\n", "unmarked": "feat(views): remove ListGadgets\n",
		"beta-tags": "2026.11.0-alpha.1\n2026.11.0-beta.1\n", "base.binpb": string(image),
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o600); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	return func(name string) string { return filepath.Join(dir, name) }
}

// TestRunJudge drives judge over files as the script writes them: new entries excuse their removal;
// once on the base, the removal fails from the first beta and is allowed, and said so, before it.
func TestRunJudge(t *testing.T) {
	t.Parallel()
	in := writeInputs(t)
	judgeArgs := []string{
		"judge", "-findings", in("findings"), "-exceptions", in("exceptions"), "-base-exceptions", in("missing"),
		"-design", in("design"), "-base-image", in("base.binpb"),
	}
	var out bytes.Buffer
	if err := run(judgeArgs, &out); err != nil {
		t.Fatalf("judge: %v", err)
	}
	if want := "proto:breaking: D-9 excuses RPC_NO_DELETE test.v1.Things.ListGadgets\n"; !strings.HasPrefix(out.String(), want) {
		t.Errorf("judge printed %q, want it to start with %q", out.String(), want)
	}
	judgeArgs[6] = in("exceptions")
	if err := run(slices.Concat(judgeArgs, []string{"-tags", in("beta-tags")}), &out); !errors.Is(err, errBreaking) {
		t.Errorf("judge with the entries on the base = %v, want %v", err, errBreaking)
	}
	out.Reset()
	if err := run(judgeArgs, &out); err != nil || !strings.Contains(out.String(), "allowed before the first beta") {
		t.Errorf("judge before the first beta = %v, %q; want the removal allowed", err, out.String())
	}
}

// TestRunMarked drives marked and the usage errors: an added entry or, with -breaking, any change needs
// a message marked breaking.
func TestRunMarked(t *testing.T) {
	t.Parallel()
	in := writeInputs(t)
	var out bytes.Buffer
	for _, tt := range []struct {
		args []string
		want error
	}{
		{[]string{"marked", "-before", in("missing"), "-after", in("exceptions"), "-message", in("message")}, nil},
		{[]string{"marked", "-before", in("missing"), "-after", in("exceptions"), "-message", in("unmarked")}, errUnmarked},
		{[]string{"marked", "-breaking", "-message", in("message")}, nil},
		{[]string{"marked", "-breaking", "-message", in("unmarked")}, errUnmarked},
		{nil, errUsage},
		{[]string{"other"}, errUsage},
		{[]string{"judge", "-nope"}, errUsage},
	} {
		if err := run(tt.args, &out); !errors.Is(err, tt.want) {
			t.Errorf("run(%q) = %v, want %v", tt.args, err, tt.want)
		}
	}
}
