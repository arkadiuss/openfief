package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTurnstileFromConfig(t *testing.T) {
	t.Parallel()
	const verify = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	tests := []struct {
		name    string
		raw     rawConfig
		want    bool
		wantErr bool
	}{
		{name: "neither set: Turnstile is off", raw: rawConfig{}},
		{name: "both set", raw: rawConfig{TurnstileSiteKey: "0x4AAAAAAAxxxxxxxxxxxxxx", TurnstileVerifyURL: verify}, want: true},
		{name: "Cloudflare's always-passing test key", raw: rawConfig{TurnstileSiteKey: "1x00000000000000000000AA", TurnstileVerifyURL: verify}, want: true},
		{name: "a site key alone", raw: rawConfig{TurnstileSiteKey: "0x4AAAAAAAxxxxxxxxxxxxxx"}, wantErr: true},
		{name: "a URL alone", raw: rawConfig{TurnstileVerifyURL: verify}, wantErr: true},
		{name: "a placeholder is no site key", raw: rawConfig{TurnstileSiteKey: "REPLACE-WITH-THE-SITE-KEY", TurnstileVerifyURL: verify}, wantErr: true},
		{name: "a site key with a space", raw: rawConfig{TurnstileSiteKey: "0x4AAAA AAAxxxxxxxxxxxxxx", TurnstileVerifyURL: verify}, wantErr: true},
		{name: "a site key that is too short", raw: rawConfig{TurnstileSiteKey: "0x4AAA", TurnstileVerifyURL: verify}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.raw.turnstileOn()
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, errBadConfig)) {
				t.Fatalf("error = %v, want error: %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("on = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTurnstileOffBuildsNothing(t *testing.T) {
	t.Parallel()
	cfg, err := (&rawConfig{}).turnstile()
	if err != nil || cfg != nil {
		t.Errorf("turnstile() = %v, %v, want none and no error", cfg, err)
	}
}

// writeKey writes the key file of a mounted Secret in dir.
func writeKey(t *testing.T, dir, key string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, turnstileKeyFile), []byte(key), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestKeyFromMountReadsItOnEveryCallTrimmed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	read := keyFromMount(dir)
	writeKey(t, dir, "  the-first-key-for-tests\n")
	if got, err := read(); err != nil || got != "the-first-key-for-tests" {
		t.Fatalf("read = %q, %v, want the key trimmed", got, err)
	}
	// A rotated Secret works at once: nothing was kept.
	writeKey(t, dir, "the-second-key-for-tests")
	if got, err := read(); err != nil || got != "the-second-key-for-tests" {
		t.Errorf("after rotation read = %q, %v", got, err)
	}
}

func TestKeyFromMountErrorsNameTheDirectoryAndNeverTheKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tests := []struct {
		name  string
		read  func() (string, error)
		setup func()
	}{
		{name: "no mount", read: keyFromMount(filepath.Join(dir, "absent"))},
		{name: "no file", read: keyFromMount(dir)},
		{name: "an empty file", read: keyFromMount(dir), setup: func() { writeKey(t, dir, " \n") }},
	}
	for _, tt := range tests {
		if tt.setup != nil {
			tt.setup()
		}
		_, err := tt.read()
		if !errors.Is(err, errNoTurnstileKey) || strings.Contains(err.Error(), "key-for-tests") {
			t.Errorf("%s: error = %v, want errNoTurnstileKey", tt.name, err)
		}
	}
}
