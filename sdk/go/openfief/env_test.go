package openfief_test

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// The tests below set the process environment, so they can't run in
// parallel.

func TestCoreURL(t *testing.T) {
	t.Setenv(openfief.EnvCoreAddress, "10.96.0.10:8080")
	if got, err := openfief.CoreURL(); err != nil || got != "http://10.96.0.10:8080" {
		t.Errorf("CoreURL() = %q, %v; want http://10.96.0.10:8080", got, err)
	}
	for _, addr := range []string{"", "core.openfief.svc", "http://10.0.0.1:80"} {
		t.Setenv(openfief.EnvCoreAddress, addr)
		if _, err := openfief.CoreURL(); !errors.Is(err, openfief.ErrNoCoreAddress) {
			t.Errorf("core address %q: error = %v, want %v", addr, err, openfief.ErrNoCoreAddress)
		}
	}
}

func TestListenAddress(t *testing.T) {
	t.Setenv(openfief.EnvListenPort, "9090")
	if got, err := openfief.ListenAddress(); err != nil || got != ":9090" {
		t.Errorf("ListenAddress() = %q, %v; want :9090", got, err)
	}
	for _, port := range []string{"", "0", "http", "70000"} {
		t.Setenv(openfief.EnvListenPort, port)
		if _, err := openfief.ListenAddress(); !errors.Is(err, openfief.ErrNoListenPort) {
			t.Errorf("port %q: error = %v, want %v", port, err, openfief.ErrNoListenPort)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	type config struct {
		Accounts []string `json:"accounts"`
	}
	t.Setenv(openfief.EnvConfig, `{"accounts":["acc-1"]}`)
	var got config
	if err := openfief.LoadConfig(&got); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !slices.Equal(got.Accounts, []string{"acc-1"}) {
		t.Errorf("LoadConfig = %+v, want accounts [acc-1]", got)
	}
	t.Setenv(openfief.EnvConfig, "")
	unchanged := config{Accounts: []string{"kept"}}
	if err := openfief.LoadConfig(&unchanged); err != nil || unchanged.Accounts[0] != "kept" {
		t.Errorf("LoadConfig without config = %+v, %v; want unchanged", unchanged, err)
	}
	t.Setenv(openfief.EnvConfig, "{")
	if err := openfief.LoadConfig(&got); err == nil {
		t.Error("LoadConfig of malformed JSON succeeded")
	}
}

func TestTokenFileFromEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	writeToken(t, path, "synthetic-token-env")
	t.Setenv(openfief.EnvIdentityFile, path)
	if got, err := (openfief.TokenFile{}).Token(); err != nil || got != "synthetic-token-env" {
		t.Errorf("Token() = %q, %v; want the token OPENFIEF_TOKEN_FILE names", got, err)
	}
}
