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

// TestParseAuthz checks --authz and SpiceDB's endpoint and key (D-164): the
// policy, which needs nothing of SpiceDB, is the default, and spicedb needs
// both an endpoint, from the flag or the environment, and the key from the
// environment alone.
func TestParseAuthz(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		env  map[string]string
		want authzConfig
	}{
		{name: "the policy is the default", want: authzConfig{Mode: authzPolicy, KeyFile: defaultSpiceDBKeyFile, Timeout: defaultSpiceDBTimeout}},
		{
			name: "the policy ignores SpiceDB's settings", args: []string{"--authz=policy"},
			env:  map[string]string{spicedbEndpointEnv: "spicedb:50051", spicedbKeyEnv: testKey},
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

// TestParseAuthzRefuses checks what --authz refuses: spicedb without an
// endpoint or a key, a mode that is none, and a key given as a flag.
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
