package connectapi_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// TestAuthServiceContract checks the generated AuthService: its path, and
// that a generated client reaches a generated handler over Connect.
func TestAuthServiceContract(t *testing.T) {
	t.Parallel()
	path, handler := identityv1connect.NewAuthServiceHandler(identityv1connect.UnimplementedAuthServiceHandler{})
	if path != "/openfief.identity.v1.AuthService/" {
		t.Fatalf("path = %q", path)
	}
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := identityv1connect.NewAuthServiceClient(srv.Client(), srv.URL)
	_, err := client.BeginLogin(t.Context(), connect.NewRequest(&identityv1.BeginLoginRequest{Origin: "http://localhost:8080"}))
	var cerr *connect.Error
	if !errors.As(err, &cerr) || cerr.Code() != connect.CodeUnimplemented {
		t.Errorf("BeginLogin = %v, want unimplemented", err)
	}
}
