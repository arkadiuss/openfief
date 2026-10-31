package connectapi

import (
	"context"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/portfolio"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// ListSleeveKinds answers GET /v1/sleeve-kinds.
func (s *readService) ListSleeveKinds(ctx context.Context, _ *connect.Request[viewsv1.ListSleeveKindsRequest]) (*connect.Response[viewsv1.ListSleeveKindsResponse], error) {
	kinds, err := s.views.SleeveKinds(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListSleeveKindsResponse{Kinds: make([]*viewsv1.SleeveKind, 0, len(kinds))}
	for i := range kinds {
		out.Kinds = append(out.Kinds, pbSleeveKind(&kinds[i]))
	}
	return connect.NewResponse(out), nil
}

// ListConnectorBindings answers GET /v1/connector-bindings.
func (s *readService) ListConnectorBindings(ctx context.Context, _ *connect.Request[viewsv1.ListConnectorBindingsRequest]) (*connect.Response[viewsv1.ListConnectorBindingsResponse], error) {
	bindings, err := s.views.ConnectorBindings(ctx)
	if err != nil {
		return nil, toConnectError(ctx, s.log, err)
	}
	out := &viewsv1.ListConnectorBindingsResponse{Bindings: make([]*viewsv1.ConnectorBinding, 0, len(bindings))}
	for _, b := range bindings {
		out.Bindings = append(out.Bindings, &viewsv1.ConnectorBinding{ConnectorInstance: b.Instance, AccountId: string(b.Account)})
	}
	return connect.NewResponse(out), nil
}

// pbSleeveKind converts a sleeve kind.
func pbSleeveKind(k *portfolio.Kind) *viewsv1.SleeveKind {
	return &viewsv1.SleeveKind{Key: string(k.Key), Name: k.Name, DefaultRole: viewsv1.SleeveRole(enumNumber(int(k.DefaultRole)))}
}
