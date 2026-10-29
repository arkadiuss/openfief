package openfief_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// freePort returns a TCP port that was free a moment ago.
func freePort(t *testing.T) int {
	t.Helper()
	var lc net.ListenConfig
	l, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatalf("address %v is not TCP", l.Addr())
	}
	if err := l.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return addr.Port
}

// TestServeConnector sets the process environment, so it can't run in
// parallel. The connector it serves accepts the core's token, refuses
// another, and stops cleanly when its context is done.
func TestServeConnector(t *testing.T) {
	keys := newTestKeys(t)
	dir := t.TempDir()
	writeKeys(t, dir, testIssuer, publicJWK(&keys.ec.PublicKey, "ec-1"))
	port := freePort(t)
	t.Setenv(openfief.EnvListenPort, strconv.Itoa(port))
	t.Setenv(openfief.EnvInstance, testInstance)
	t.Setenv(openfief.EnvCoreSubject, coreSubject)
	t.Setenv(openfief.EnvKeysDir, dir)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- openfief.ServeConnector(ctx, &fakeConnector{limit: 1}) }()

	valid := claims()
	valid.Expiry = jwt.NewNumericDate(time.Now().Add(5 * time.Minute))
	valid.IssuedAt, valid.NotBefore = jwt.NewNumericDate(time.Now()), jwt.NewNumericDate(time.Now())
	client := connectorv1connect.NewConnectorServiceClient(http.DefaultClient, "http://127.0.0.1:"+strconv.Itoa(port))
	call := func(token string) error {
		req := connect.NewRequest(&connectorv1.SyncRequest{BatchId: "b-1"})
		req.Header().Set("Authorization", "Bearer "+token)
		if _, err := client.Sync(ctx, req); err != nil {
			return fmt.Errorf("sync: %w", err)
		}
		return nil
	}
	var err error
	for range 50 { // the server starts in the background
		if err = call(sign(t, jose.ES256, keys.ec, "ec-1", &valid)); connect.CodeOf(err) != connect.CodeUnavailable {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Sync with the core's token: %v", err)
	}
	if err := call("synthetic-other-token"); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("Sync with another token: %v, want Unauthenticated", err)
	}
	cancel()
	if err := <-done; err != nil {
		t.Errorf("ServeConnector after cancel = %v, want nil", err)
	}
}

// TestServeConnectorNeedsEnvironment sets the process environment, so it
// can't run in parallel.
func TestServeConnectorNeedsEnvironment(t *testing.T) {
	t.Setenv(openfief.EnvListenPort, "")
	if err := openfief.ServeConnector(t.Context(), &fakeConnector{}); !errors.Is(err, openfief.ErrNoListenPort) {
		t.Errorf("ServeConnector without a port = %v, want %v", err, openfief.ErrNoListenPort)
	}
	t.Setenv(openfief.EnvListenPort, "8080")
	t.Setenv(openfief.EnvInstance, "")
	if err := openfief.ServeConnector(t.Context(), &fakeConnector{}); !errors.Is(err, openfief.ErrNoInstance) {
		t.Errorf("ServeConnector without an instance = %v, want %v", err, openfief.ErrNoInstance)
	}
}
