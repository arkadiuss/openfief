package connectapi_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"slices"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"

	gocmp "github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/portfolio/connectapi"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

// memSleeves keeps kinds and sleeves in memory, checked by the portfolio
// rules as the ledger checks them, or fails with err.
type memSleeves struct {
	kinds   []portfolio.Kind
	sleeves []portfolio.Sleeve
	err     error
}

func (m *memSleeves) Kinds(context.Context) ([]portfolio.Kind, error) { return m.kinds, m.err }

func (m *memSleeves) SaveKind(_ context.Context, k portfolio.Kind) error {
	if err := portfolio.CheckKind(m.kinds, k); err != nil {
		return fmt.Errorf("checking the kind: %w", err)
	}
	m.kinds = append(slices.DeleteFunc(m.kinds, func(o portfolio.Kind) bool { return o.Key == k.Key }), k)
	return nil
}

func (m *memSleeves) Sleeves(context.Context) ([]portfolio.Sleeve, error) { return m.sleeves, m.err }

func (m *memSleeves) SaveSleeve(_ context.Context, s portfolio.Sleeve) error {
	checked, err := portfolio.CheckSleeve(m.kinds, s)
	if err != nil {
		return fmt.Errorf("checking the sleeve: %w", err)
	}
	m.sleeves = append(slices.DeleteFunc(m.sleeves, func(o portfolio.Sleeve) bool { return o.ID == s.ID }), checked)
	return nil
}

// sleeveClient is the generated SleeveService client.
type sleeveClient struct {
	portfoliov1connect.SleeveServiceClient
}

// newSleeveClient serves the SleeveService over a fresh install's kinds
// and sleeves and returns a client calling it with the caller token.
func newSleeveClient(t *testing.T, store *memSleeves, caller string) sleeveClient {
	t.Helper()
	_, handler := connectapi.NewSleeves(&app.Sleeves{Owners: owners{}, Store: store}, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return sleeveClient{portfoliov1connect.NewSleeveServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

func freshInstall() *memSleeves {
	return &memSleeves{kinds: portfolio.DefaultKinds(), sleeves: portfolio.DefaultSleeves()}
}

func TestSleeveServiceRoundTrip(t *testing.T) {
	t.Parallel()
	c := newSleeveClient(t, freshInstall(), viewCaller)
	ctx := t.Context()
	kind, err := c.AddSleeveKind(ctx, connect.NewRequest(&portfoliov1.AddSleeveKindRequest{
		Origin: local, SessionToken: ownerSession, Name: "Gold", DefaultRole: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE,
	}))
	want := &portfoliov1.SleeveKind{Key: "gold", Name: "Gold", DefaultRole: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE}
	if err != nil || !gocmp.Equal(want, kind.Msg.GetKind(), protocmp.Transform()) {
		t.Fatalf("AddSleeveKind = %v, %v; want %v", kind, err, want)
	}
	renamed, err := c.RenameSleeveKind(ctx, connect.NewRequest(&portfoliov1.RenameSleeveKindRequest{
		Origin: local, SessionToken: ownerSession, Key: "satellite", Name: "Play",
	}))
	if err != nil || renamed.Msg.GetKind().GetDefaultRole() != portfoliov1.SleeveRole_SLEEVE_ROLE_OFFENSIVE {
		t.Fatalf("RenameSleeveKind = %v, %v; want satellite renamed, still offensive", renamed, err)
	}
	added, err := c.AddSleeve(ctx, connect.NewRequest(&portfoliov1.AddSleeveRequest{
		Origin: local, SessionToken: ownerSession, Name: "Gold bars", Kind: "gold",
	}))
	wantSleeve := &portfoliov1.Sleeve{Id: "gold-bars", Name: "Gold bars", Kind: "gold", Role: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE}
	if err != nil || !gocmp.Equal(wantSleeve, added.Msg.GetSleeve(), protocmp.Transform()) {
		t.Fatalf("AddSleeve = %v, %v; want %v", added, err, wantSleeve)
	}
	updated, err := c.UpdateSleeve(ctx, connect.NewRequest(&portfoliov1.UpdateSleeveRequest{
		Origin: local, SessionToken: ownerSession, Id: "satellite", Name: "Satellite", Kind: "satellite",
		Role: portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE, RoleOverride: true,
	}))
	if err != nil || !updated.Msg.GetSleeve().GetRoleOverride() || updated.Msg.GetSleeve().GetRole() != portfoliov1.SleeveRole_SLEEVE_ROLE_DEFENSIVE {
		t.Errorf("UpdateSleeve = %v, %v; want satellite defensive by the owner's choice", updated, err)
	}
}

func TestSleeveServiceErrors(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	add := func(name, kind string, role portfoliov1.SleeveRole) *connect.Request[portfoliov1.AddSleeveRequest] {
		return connect.NewRequest(&portfoliov1.AddSleeveRequest{Origin: local, SessionToken: ownerSession, Name: name, Kind: kind, Role: role})
	}
	tests := []struct {
		name   string
		caller string
		store  *memSleeves
		call   func(c sleeveClient) connect.Code
		want   connect.Code
	}{
		{
			name: "not a view plugin", caller: "another-plugin", store: freshInstall(), want: connect.CodePermissionDenied,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeve(ctx, add("Gold", "core", 0))
				return connect.CodeOf(err)
			},
		},
		{
			name: "no session", caller: viewCaller, store: freshInstall(), want: connect.CodeUnauthenticated,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeve(ctx, connect.NewRequest(&portfoliov1.AddSleeveRequest{Origin: local, SessionToken: "nope", Name: "Gold", Kind: "core"}))
				return connect.CodeOf(err)
			},
		},
		{
			name: "a read session", caller: viewCaller, store: freshInstall(), want: connect.CodePermissionDenied,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeve(ctx, connect.NewRequest(&portfoliov1.AddSleeveRequest{Origin: local, SessionToken: readSession, Name: "Gold", Kind: "core"}))
				return connect.CodeOf(err)
			},
		},
		{
			name: "another role without the override", caller: viewCaller, store: freshInstall(), want: connect.CodeInvalidArgument,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeve(ctx, add("Play", "core", portfoliov1.SleeveRole_SLEEVE_ROLE_OFFENSIVE))
				return connect.CodeOf(err)
			},
		},
		{
			name: "an unknown kind", caller: viewCaller, store: freshInstall(), want: connect.CodeInvalidArgument,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeve(ctx, add("Gold", "gold", 0))
				return connect.CodeOf(err)
			},
		},
		{
			name: "a taken ID", caller: viewCaller, store: freshInstall(), want: connect.CodeAlreadyExists,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeve(ctx, add("Core", "core", 0))
				return connect.CodeOf(err)
			},
		},
		{
			name: "an unknown sleeve", caller: viewCaller, store: freshInstall(), want: connect.CodeNotFound,
			call: func(c sleeveClient) connect.Code {
				_, err := c.UpdateSleeve(ctx, connect.NewRequest(&portfoliov1.UpdateSleeveRequest{Origin: local, SessionToken: ownerSession, Id: "gold", Name: "Gold", Kind: "core"}))
				return connect.CodeOf(err)
			},
		},
		{
			name: "a kind without a role", caller: viewCaller, store: freshInstall(), want: connect.CodeInvalidArgument,
			call: func(c sleeveClient) connect.Code {
				_, err := c.AddSleeveKind(ctx, connect.NewRequest(&portfoliov1.AddSleeveKindRequest{Origin: local, SessionToken: ownerSession, Name: "Gold"}))
				return connect.CodeOf(err)
			},
		},
		{
			name: "the store down", caller: viewCaller, store: &memSleeves{err: errDown}, want: connect.CodeInternal,
			call: func(c sleeveClient) connect.Code {
				_, err := c.RenameSleeveKind(ctx, connect.NewRequest(&portfoliov1.RenameSleeveKindRequest{Origin: local, SessionToken: ownerSession, Key: "core", Name: "X"}))
				return connect.CodeOf(err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.call(newSleeveClient(t, tt.store, tt.caller)); got != tt.want {
				t.Errorf("code = %v, want %v", got, tt.want)
			}
		})
	}
}
