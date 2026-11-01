package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
)

// writeInputs writes judge's and marked's input files to a temporary directory and returns its path.
func writeInputs(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	image, err := proto.Marshal(baseSet())
	if err != nil {
		t.Fatalf("marshalling the base image: %v", err)
	}
	files["base.binpb"] = string(image)
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o600); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	return dir
}

func TestRun(t *testing.T) {
	t.Parallel()
	dir := writeInputs(t, map[string]string{
		"findings": rpcGone + "\n" + requestGone + "\n", "exceptions": rpcEntry + "\n" + requestEntry + "\n",
		"design": design, "message": "feat(views)!: remove ListGadgets (D-9)\n", "unmarked": "feat(views): remove ListGadgets\n",
	})
	in := func(name string) string { return filepath.Join(dir, name) }
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
	// With the entries on the base too, nothing excuses the removal any more.
	judgeArgs[6] = in("exceptions")
	if err := run(judgeArgs, &out); !errors.Is(err, errBreaking) {
		t.Errorf("judge with the entries on the base = %v, want %v", err, errBreaking)
	}
	marked := []string{"marked", "-before", in("missing"), "-after", in("exceptions"), "-message", in("message")}
	if err := run(marked, &out); err != nil {
		t.Errorf("marked with a breaking message: %v", err)
	}
	marked[6] = in("unmarked")
	if err := run(marked, &out); !errors.Is(err, errUnmarked) {
		t.Errorf("marked with an unmarked message = %v, want %v", err, errUnmarked)
	}
	for _, args := range [][]string{nil, {"other"}, {"judge", "-nope"}} {
		if err := run(args, &out); !errors.Is(err, errUsage) {
			t.Errorf("run(%q) = %v, want %v", args, err, errUsage)
		}
	}
}
