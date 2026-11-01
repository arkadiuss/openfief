package main

import (
	"errors"
	"testing"
)

func TestCheckMarked(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, before, after, msg string
		want                     error
	}{
		{name: "no exceptions file", msg: "feat(ui): add a page"},
		{name: "unchanged entries need no mark", before: rpcEntry, after: rpcEntry, msg: "fix(ui): a typo"},
		{name: "removing an entry needs no mark", before: rpcEntry, msg: "chore: drop an old exception"},
		{name: "an added entry with ! in the subject", after: rpcEntry, msg: "feat(views)!: remove ListGadgets (D-9)"},
		{name: "an added entry with ! and no scope", after: rpcEntry, msg: "feat!: remove ListGadgets"},
		{name: "an added entry with a BREAKING CHANGE footer", after: rpcEntry, msg: "feat(views): remove ListGadgets\n\nWhy.\n\nBREAKING CHANGE: ListGadgets is gone (D-9)\n"},
		{name: "an added entry with a BREAKING-CHANGE footer", after: rpcEntry, msg: "feat(views): remove ListGadgets\n\nBREAKING-CHANGE: ListGadgets is gone\n"},
		{name: "an added entry unmarked", after: rpcEntry, msg: "feat(views): remove ListGadgets (D-9)", want: errUnmarked},
		{name: "a ! in the summary only", after: rpcEntry, msg: "feat(views): remove ListGadgets!", want: errUnmarked},
		{name: "a footer only in a comment", after: rpcEntry, msg: "feat(views): remove\n# BREAKING CHANGE: no\n", want: errUnmarked},
		{name: "a malformed file", after: "D-9 nonsense", msg: "feat!: x", want: errExceptions},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := checkMarked(tt.before, tt.after, tt.msg); !errors.Is(err, tt.want) {
				t.Errorf("checkMarked = %v, want %v", err, tt.want)
			}
		})
	}
}
