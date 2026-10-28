package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

func TestParseFlags(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		args    []string
		want    operator.Settings
		wantErr error
	}{
		{
			name: "defaults name the openfief Services",
			want: operator.Settings{
				CoreService: "openfief-core", CorePort: "connect",
				EgressProxyService: "egress-proxy", EgressProxyPort: "proxy", PluginPort: 8080,
				CoreServiceAccount: "openfief-core",
			},
		},
		{
			name: "every setting",
			args: []string{
				"--core-service=core", "--core-port=rpc", "--egress-proxy-service=proxy",
				"--egress-proxy-port=connect", "--plugin-port=9090", "--core-service-account=core-sa",
			},
			want: operator.Settings{
				CoreService: "core", CorePort: "rpc", EgressProxyService: "proxy", EgressProxyPort: "connect", PluginPort: 9090,
				CoreServiceAccount: "core-sa",
			},
		},
		{name: "signing keys refresh 0", args: []string{"--signing-keys-refresh=0s"}, wantErr: errKeysRefresh},
		{name: "plugin port 0", args: []string{"--plugin-port=0"}, wantErr: errPluginPort},
		{name: "plugin port above 65535", args: []string{"--plugin-port=70000"}, wantErr: errPluginPort},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			o, err := parseFlags(tt.args)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("parseFlags error mismatch (-want +got):\n%s", diff)
			}
			if err != nil {
				return
			}
			if diff := cmp.Diff(tt.want, o.settings); diff != "" {
				t.Errorf("parseFlags settings mismatch (-want +got):\n%s", diff)
			}
			if !o.leaderElect || o.metricsAddr != "0" {
				t.Errorf("leader election %v, metrics %q; want leader election on and metrics off", o.leaderElect, o.metricsAddr)
			}
		})
	}
}

func TestParseFlagsSigningKeysRefresh(t *testing.T) {
	t.Parallel()
	for args, want := range map[string]time.Duration{"": 10 * time.Minute, "--signing-keys-refresh=1m": time.Minute} {
		var argv []string
		if args != "" {
			argv = []string{args}
		}
		o, err := parseFlags(argv)
		if err != nil {
			t.Fatalf("parseFlags(%q): %v", args, err)
		}
		if o.keysRefresh != want {
			t.Errorf("parseFlags(%q) signing keys refresh %v, want %v", args, o.keysRefresh, want)
		}
	}
}

func TestParseFlagsTrustedSigners(t *testing.T) {
	t.Parallel()
	for args, want := range map[string]string{"": defaultTrustedSigners, "--trusted-signers=/x.json": "/x.json"} {
		var argv []string
		if args != "" {
			argv = []string{args}
		}
		o, err := parseFlags(argv)
		if err != nil {
			t.Fatalf("parseFlags(%q): %v", args, err)
		}
		if o.trustedSigners != want {
			t.Errorf("parseFlags(%q) trusted signers %q, want %q", args, o.trustedSigners, want)
		}
	}
}

func TestTrustPolicy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return path
	}
	tests := []struct {
		name        string
		path        string
		wantSigners []string
		wantErr     error
	}{
		{name: "no signers file: CI and the catalog only", path: filepath.Join(dir, "missing.json"), wantSigners: []string{"openfief-ci"}},
		{
			name:        "owner signers are added",
			path:        write("ok.json", `{"signers":[{"name":"my-ci","issuer":"https://issuer.example.com","subject":"https://example.com/release"}]}`),
			wantSigners: []string{"openfief-ci", "my-ci"},
		},
		{name: "invalid signers file stops the operator", path: write("bad.json", `{"signers":[{"name":"x"}]}`), wantErr: plugins.ErrInvalidSigner},
		{name: "unknown field stops the operator", path: write("unknown.json", `{"allowUnsigned":true}`), wantErr: plugins.ErrInvalidCatalog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := trustPolicy(tt.path)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("trustPolicy error (-want +got):\n%s", diff)
			}
			var names []string
			for _, s := range p.Signers() {
				names = append(names, s.Name)
			}
			if diff := cmp.Diff(tt.wantSigners, names); diff != "" {
				t.Errorf("signers (-want +got):\n%s", diff)
			}
		})
	}
}

func TestImageVerifier(t *testing.T) {
	t.Parallel()
	if _, err := imageVerifier(filepath.Join(t.TempDir(), "none.json")); err != nil {
		t.Errorf("imageVerifier with the bundled trust root and catalog: %v", err)
	}
}
