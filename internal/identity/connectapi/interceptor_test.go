package connectapi_test

import (
	"log/slog"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/identity/connectapi"
	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1/identityv1connect"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
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

// TestCallerInterceptorStreams puts CallerInterceptor in front of a
// client-streaming service that implements nothing, the statement upload:
// a view plugin's stream gets through to it (unimplemented), any other
// caller's is refused before it.
func TestCallerInterceptorStreams(t *testing.T) {
	t.Parallel()
	path, handler := ingestv1connect.NewStatementServiceHandler(ingestv1connect.UnimplementedStatementServiceHandler{},
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
			stream := ingestv1connect.NewStatementServiceClient(srv.Client(), srv.URL).UploadStatement(t.Context())
			if tt.token != "" {
				stream.RequestHeader().Set("Authorization", "Bearer "+tt.token)
			}
			header := &ingestv1.StatementHeader{Origin: local, AccountId: "acc-1"}
			if err := stream.Send(&ingestv1.UploadStatementRequest{Part: &ingestv1.UploadStatementRequest_Header{Header: header}}); err != nil {
				t.Logf("Send: %v (the server may have answered already)", err)
			}
			_, err := stream.CloseAndReceive()
			wantCode(t, path, err, tt.want)
		})
	}
}
