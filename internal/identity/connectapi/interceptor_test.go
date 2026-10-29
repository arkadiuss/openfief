package connectapi_test

import (
	"log/slog"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/identity/connectapi"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
)

// TestCallerInterceptor puts CallerInterceptor in front of a handler that
// implements nothing: a view plugin's call gets through to it
// (unimplemented), any other caller is refused before it.
func TestCallerInterceptor(t *testing.T) {
	t.Parallel()
	path, handler := identityv1connect.NewAuthServiceHandler(identityv1connect.UnimplementedAuthServiceHandler{},
		connect.WithInterceptors(connectapi.NewCallerInterceptor(callers{}, slog.New(slog.DiscardHandler))))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	tests := []struct {
		name, token string
		want        connect.Code
	}{
		{name: "a view plugin", token: viewCaller, want: connect.CodeUnimplemented},
		{name: "no caller token", want: connect.CodePermissionDenied},
		{name: "not a view plugin", token: "connector-token", want: connect.CodePermissionDenied},
		{name: "reviewer down", token: brokenCaller, want: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var opts []connect.ClientOption
			if tt.token != "" {
				opts = append(opts, connect.WithInterceptors(bearer(tt.token)))
			}
			client := identityv1connect.NewAuthServiceClient(srv.Client(), srv.URL, opts...)
			_, err := client.WhoAmI(t.Context(), connect.NewRequest(&identityv1.WhoAmIRequest{Origin: local}))
			wantCode(t, path, err, tt.want)
		})
	}
}
