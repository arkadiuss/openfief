package openfief

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
)

// The environment of a plugin container. The operator sets the first four
// in every plugin; see EnvEgressIdentityFile for the fifth.
const (
	// EnvConfig holds the plugin's configuration as JSON.
	EnvConfig = "OPENFIEF_CONFIG"
	// EnvIdentityFile holds the path of the plugin's projected
	// ServiceAccount token for the core (audience openfief-core, D-27).
	EnvIdentityFile = "OPENFIEF_TOKEN_FILE"
	// EnvCoreAddress holds the core's address as ClusterIP:port; plugins
	// have no DNS.
	EnvCoreAddress = "OPENFIEF_CORE_ADDRESS"
	// EnvListenPort holds the port the plugin serves Connect RPC on.
	EnvListenPort = "OPENFIEF_LISTEN_PORT"
	// EnvEgressIdentityFile holds the path of the plugin's projected token
	// for the egress proxy (audience openfief-egress, D-63). The operator
	// doesn't mount that token yet; this name is the SDK's proposal.
	EnvEgressIdentityFile = "OPENFIEF_EGRESS_TOKEN_FILE"
)

// Environment errors.
var (
	// ErrNoCoreAddress reports a missing or malformed OPENFIEF_CORE_ADDRESS.
	ErrNoCoreAddress = errors.New("no core address (OPENFIEF_CORE_ADDRESS is not host:port)")
	// ErrNoListenPort reports a missing or malformed OPENFIEF_LISTEN_PORT.
	ErrNoListenPort = errors.New("no listen port (OPENFIEF_LISTEN_PORT is not a port)")
)

// CoreURL returns the core's base URL, "http://" plus OPENFIEF_CORE_ADDRESS.
func CoreURL() (string, error) {
	addr := os.Getenv(EnvCoreAddress)
	if _, _, err := net.SplitHostPort(addr); err != nil || addr == "" {
		return "", fmt.Errorf("%w: %q", ErrNoCoreAddress, addr)
	}
	return "http://" + addr, nil
}

// ListenAddress returns the address the plugin serves Connect RPC on,
// ":<OPENFIEF_LISTEN_PORT>".
func ListenAddress() (string, error) {
	raw := os.Getenv(EnvListenPort)
	port, err := strconv.ParseUint(raw, 10, 16)
	if err != nil || port == 0 {
		return "", fmt.Errorf("%w: %q", ErrNoListenPort, raw)
	}
	return ":" + raw, nil
}

// LoadConfig decodes the plugin's configuration, OPENFIEF_CONFIG, into v. A
// plugin without configuration leaves v unchanged.
func LoadConfig(v any) error {
	raw := os.Getenv(EnvConfig)
	if raw == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(raw), v); err != nil {
		return fmt.Errorf("decode %s: %w", EnvConfig, err)
	}
	return nil
}
