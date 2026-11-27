package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
	identityapp "github.com/arkadiuss/openfief/internal/identity/app"
)

// testKey is a placeholder preshared key; it is no credential.
const testKey = "placeholder-preshared-key"

// TestParseAuthz checks --authz and SpiceDB's endpoint and key (D-164):
// spicedb is the default, and needs an endpoint, from the flag or the
// environment, and the key from the environment or its file; plaintext gRPC
// comes from the flag or the environment; the policy, the opt-out, needs
// nothing of SpiceDB.
func TestParseAuthz(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		env  map[string]string
		want authzConfig
	}{
		{
			name: "spicedb is the default",
			env:  map[string]string{spicedbEndpointEnv: "spicedb:50051"},
			want: authzConfig{Mode: authzSpiceDB, Endpoint: "spicedb:50051", KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "the core's pod: plaintext from the environment",
			env:  clusterSpiceDB(),
			want: clusterAuthz(),
		},
		{
			name: "the environment's false is TLS", env: map[string]string{spicedbEndpointEnv: "spicedb:50051", spicedbInsecureEnv: "false"},
			want: authzConfig{Mode: authzSpiceDB, Endpoint: "spicedb:50051", KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "the policy needs nothing of SpiceDB", args: []string{"--authz=policy"},
			want: authzConfig{Mode: authzPolicy, KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "the policy ignores SpiceDB's settings", args: []string{"--authz=policy"},
			env:  clusterSpiceDB(),
			want: authzConfig{Mode: authzPolicy, KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "spicedb from the environment", args: []string{"--authz=spicedb"},
			env:  map[string]string{spicedbEndpointEnv: "spicedb:50051", spicedbKeyEnv: testKey},
			want: authzConfig{Mode: authzSpiceDB, Endpoint: "spicedb:50051", Key: testKey, KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "no key yet is not an error: serve waits for the file", args: []string{"--authz=spicedb"},
			env:  map[string]string{spicedbEndpointEnv: "spicedb:50051"},
			want: authzConfig{Mode: authzSpiceDB, Endpoint: "spicedb:50051", KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "the key file flag", args: []string{"--authz=spicedb", "--spicedb-key-file=/k/key"},
			env:  map[string]string{spicedbEndpointEnv: "spicedb:50051"},
			want: authzConfig{Mode: authzSpiceDB, Endpoint: "spicedb:50051", KeyFile: "/k/key", Timeout: defaultSpiceDBTimeout},
		},
		{
			name: "the flag overrides the endpoint", args: []string{"--authz=spicedb", "--spicedb-endpoint=other:1", "--spicedb-insecure", "--spicedb-timeout=1s"},
			env:  map[string]string{spicedbEndpointEnv: "spicedb:50051", spicedbKeyEnv: testKey},
			want: authzConfig{Mode: authzSpiceDB, Endpoint: "other:1", Key: testKey, KeyFile: defaultSpiceDBKeyFile, Insecure: true, Timeout: time.Second},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseServe(tt.args, func(k string) string { return tt.env[k] })
			if err != nil {
				t.Fatalf("parseServe: %v", err)
			}
			if diff := cmp.Diff(tt.want, got.Authz); diff != "" {
				t.Errorf("Authz (-want +got):\n%s", diff)
			}
		})
	}
}

// TestParseAuthzRefuses checks what --authz refuses: spicedb, the default,
// without an endpoint, so a run without SpiceDB fails at once instead of
// waiting for it, or without a key, a mode that is none, a plaintext setting
// that is no boolean, and a key given as a flag.
func TestParseAuthzRefuses(t *testing.T) {
	t.Parallel()
	endpoint := map[string]string{spicedbEndpointEnv: "spicedb:50051"}
	key := map[string]string{spicedbKeyEnv: testKey}
	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		wantErr error // nil: any error
	}{
		{name: "no endpoint", args: []string{"--authz=spicedb"}, env: key, wantErr: errNoSpiceDBEndpoint},
		{name: "no flag and no endpoint: the default is spicedb", wantErr: errNoSpiceDBEndpoint},
		{name: "a plaintext setting that is no boolean", env: map[string]string{spicedbEndpointEnv: "spicedb:50051", spicedbInsecureEnv: "yes please"}, wantErr: errBadSpiceDBInsecure},
		{name: "no key and no file", args: []string{"--authz=spicedb", "--spicedb-key-file="}, env: endpoint, wantErr: errNoSpiceDBKey},
		{name: "an unknown mode", args: []string{"--authz=openfga"}, wantErr: errBadAuthz},
		{name: "no timeout", args: []string{"--authz=spicedb", "--spicedb-endpoint=x:1", "--spicedb-timeout=0s"}, env: key, wantErr: errBadSpiceDBTimeout},
		{name: "no key flag exists", args: []string{"--spicedb-key=" + testKey}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseServe(tt.args, func(k string) string { return tt.env[k] })
			if err == nil || (tt.wantErr != nil && !errors.Is(err, tt.wantErr)) {
				t.Errorf("parseServe error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestSpiceDBKeyStaysOutOfOutput checks that the preshared key shows in no
// formatting of the configuration and no log line.
func TestSpiceDBKeyStaysOutOfOutput(t *testing.T) {
	t.Parallel()
	cfg, err := parseServe([]string{"--authz=spicedb", "--spicedb-endpoint=spicedb:50051"}, func(k string) string {
		return map[string]string{spicedbKeyEnv: testKey}[k]
	})
	if err != nil {
		t.Fatalf("parseServe: %v", err)
	}
	var logged bytes.Buffer
	slog.New(slog.NewJSONHandler(&logged, nil)).Info("config", "authz", cfg.Authz, "key", cfg.Authz.Key)
	for name, out := range map[string]string{
		"%v": fmt.Sprintf("%v", cfg), "%+v": fmt.Sprintf("%+v", cfg), "%#v": fmt.Sprintf("%#v", cfg), "%s": cfg.Authz.Key.String(),
		"slog": logged.String(),
	} {
		if strings.Contains(out, testKey) {
			t.Errorf("%s of the config shows the preshared key: %s", name, out)
		}
	}
}

func TestNewAuthorizationPolicy(t *testing.T) {
	t.Parallel()
	for _, cfg := range []authzConfig{{}, {Mode: authzPolicy}} {
		az, err := newAuthorization(&cfg)
		if err != nil {
			t.Fatalf("newAuthorization: %v", err)
		}
		if az.authorizer != (identityapp.Policy{}) || az.syncer != nil || az.closer != nil || len(az.registryOptions()) != 0 {
			t.Errorf("newAuthorization(%q) = %+v, want the policy alone", cfg.Mode, az)
		}
	}
}

func TestNewAuthorizationSpiceDB(t *testing.T) {
	t.Parallel()
	// Dialing is lazy: nothing connects to the placeholder address.
	az, err := newAuthorization(&authzConfig{Mode: authzSpiceDB, Endpoint: "127.0.0.1:1", Key: testKey, Insecure: true, Timeout: time.Second})
	if err != nil {
		t.Fatalf("newAuthorization: %v", err)
	}
	defer func() {
		if err := az.closer.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()
	if _, isPolicy := az.authorizer.(identityapp.Policy); isPolicy || az.syncer == nil || len(az.registryOptions()) != 1 {
		t.Errorf("newAuthorization = %+v, want SpiceDB's authorizer and syncer", az)
	}
}

// TestSpiceDBDownRefusesEverything checks, with a real client, that a
// SpiceDB nobody listens at refuses every check with an error: the owner's
// too, and no holder's view (D-164).
func TestSpiceDBDownRefusesEverything(t *testing.T) {
	t.Parallel()
	az, err := newAuthorization(&authzConfig{Mode: authzSpiceDB, Endpoint: "127.0.0.1:1", Key: testKey, Insecure: true, Timeout: time.Second})
	if err != nil {
		t.Fatalf("newAuthorization: %v", err)
	}
	defer az.closer.Close() //nolint:errcheck // a test's connection to nothing
	ok, err := az.authorizer.MayAct(t.Context(), identity.Caller{Holder: identity.HolderOwner, ID: "s"}, identity.ActionView, "brokerage")
	if ok || err == nil {
		t.Errorf("MayAct with SpiceDB down = %v, %v, want false and an error", ok, err)
	}
}

// TestNoSpiceDBSaysHowToOptOut checks that a run without SpiceDB's settings,
// such as a core started outside the cluster, fails at once with what is
// missing and the opt-out, --authz=policy, instead of waiting for a SpiceDB
// that isn't there.
func TestNoSpiceDBSaysHowToOptOut(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"token", "issue", "--name=x"}, {"token", "revoke", "x"}} {
		_, err := parseArgs(args, func(k string) string { return map[string]string{dsnEnv: testDSN}[k] })
		if !errors.Is(err, errNoSpiceDBEndpoint) || !strings.Contains(err.Error(), "pass --authz=policy") {
			t.Errorf("parseArgs(%q) without SpiceDB = %v, want errNoSpiceDBEndpoint naming --authz=policy", args, err)
		}
	}
}
