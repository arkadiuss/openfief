package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestLoggedAttention checks the stand-in for the attention store: it logs
// the kind and subject of each item raised and cleared, and fails nothing.
func TestLoggedAttention(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	a := loggedAttention{log: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))}
	if err := a.Raise(t.Context(), "instrument-unclassified", "PLTEST000010", "detail"); err != nil {
		t.Fatalf("Raise: %v", err)
	}
	if err := a.Clear(t.Context(), "instrument-unclassified", "PLTEST000010"); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	for _, want := range []string{`msg="attention item" kind=instrument-unclassified subject=PLTEST000010`, `msg="attention item cleared"`} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("log lacks %q:\n%s", want, buf.String())
		}
	}
}
