package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeyFromMount(t *testing.T) {
	t.Parallel()
	write := func(t *testing.T, content string) string {
		t.Helper()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, keyFile), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	tests := []struct {
		name    string
		dir     func(t *testing.T) string
		want    string
		wantErr bool
	}{
		{name: "the key, trimmed", dir: func(t *testing.T) string { t.Helper(); return write(t, fakeKey+"\n") }, want: fakeKey},
		{name: "no Secret mounted", dir: func(t *testing.T) string { t.Helper(); return filepath.Join(t.TempDir(), "absent") }},
		{name: "an empty mount", dir: func(t *testing.T) string { t.Helper(); return t.TempDir() }},
		{name: "an empty key", dir: func(t *testing.T) string { t.Helper(); return write(t, " \n") }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := keyFromMount(tt.dir(t))()
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("key = %q, %v; want %q, error %v", got, err, tt.want, tt.wantErr)
			}
			if err != nil && strings.Contains(err.Error(), fakeKey) {
				t.Errorf("error %q holds the key", err)
			}
		})
	}
}
