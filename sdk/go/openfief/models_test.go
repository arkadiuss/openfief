package openfief_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// fakeProvider answers Describe with a fixed location.
type fakeProvider struct {
	modelsv1connect.UnimplementedModelProviderServiceHandler
}

func (fakeProvider) Describe(context.Context, *connect.Request[modelsv1.DescribeRequest]) (*connect.Response[modelsv1.DescribeResponse], error) {
	return connect.NewResponse(&modelsv1.DescribeResponse{Location: modelsv1.Location_LOCATION_OWNER_HARDWARE}), nil
}

// TestModelProviderServerIsForTheCoreOnly checks that the model-provider
// contract is served behind the core check: the core's token reaches the
// handler, any other token and no token do not.
func TestModelProviderServerIsForTheCoreOnly(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(openfief.NewModelProviderServer("", fakeProvider{}, fakeVerifier{}).Handler)
	t.Cleanup(srv.Close)
	client := modelsv1connect.NewModelProviderServiceClient(http.DefaultClient, srv.URL)
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{name: "the core's token", token: coreToken, want: true},
		{name: "another token", token: "synthetic-other-token"},
		{name: "no token"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := connect.NewRequest(&modelsv1.DescribeRequest{})
			if tt.token != "" {
				req.Header().Set("Authorization", "Bearer "+tt.token)
			}
			resp, err := client.Describe(t.Context(), req)
			if (err == nil) != tt.want {
				t.Fatalf("Describe error = %v, want success %v", err, tt.want)
			}
			if err == nil && resp.Msg.GetLocation() != modelsv1.Location_LOCATION_OWNER_HARDWARE {
				t.Errorf("location = %v, want the owner's hardware", resp.Msg.GetLocation())
			}
		})
	}
}
