package connectapi_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	"github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/modelrouter/connectapi"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1/modelsv1connect"
)

// ownerOrViewer admits the token "owner" as the owner's session and "viewer"
// as a viewer's.
type ownerOrViewer struct{}

func (ownerOrViewer) Asker(_ context.Context, _, token string) (app.Actor, error) {
	switch token {
	case "owner":
		return app.Actor{Kind: app.ActorOwner, ID: "session-owner"}, nil
	case "viewer":
		return app.Actor{Kind: app.ActorViewer, ID: "session-viewer"}, nil
	}
	return app.Actor{}, app.ErrNoSession
}

// opinionsFixture serves both services over one store, with opinions turned
// on, the switch and the Ask box's toggle on and a route that may receive T0,
// T1 and T2.
type opinionsFixture struct {
	ask      modelsv1connect.AskModelServiceClient
	settings modelsv1connect.ModelSettingsServiceClient
	store    *memStore
}

func newOpinionsFixture(t *testing.T, demo bool) *opinionsFixture {
	t.Helper()
	m := modelrouter.Ref{Instance: "p", Model: "m"}
	store := &memStore{s: modelrouter.Settings{
		Switch:   modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}},
		Routes:   map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: {Role: modelrouter.RoleAsk, Models: []modelrouter.Ref{m}}},
		Policy:   modelrouter.DataTierPolicy{modelrouter.DataTierT0: {m}, modelrouter.DataTierT1: {m}, modelrouter.DataTierT2: {m}},
		Opinions: true,
	}}
	router := app.NewRouter(&app.RouterPorts{
		Settings: store, Providers: &provider{}, Audit: &audit{}, Attention: noAttention{},
		CallIDs: func() (string, error) { return "0000000000000001", nil },
		Now:     func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
	})
	log := slog.New(slog.DiscardHandler)
	askPath, askHandler := connectapi.NewAsk(app.NewAsk(ownerOrViewer{}, store, router).WithDemo(demo), passThrough{}, log)
	setPath, setHandler := connectapi.NewSettings(app.NewSettings(store, owners{}).WithDemo(demo), passThrough{}, log)
	mux := http.NewServeMux()
	mux.Handle(askPath, askHandler)
	mux.Handle(setPath, setHandler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &opinionsFixture{
		ask: modelsv1connect.NewAskModelServiceClient(srv.Client(), srv.URL), store: store,
		settings: modelsv1connect.NewModelSettingsServiceClient(srv.Client(), srv.URL),
	}
}

func (f *opinionsFixture) opinions(token string) (modelsv1.AskOpinions, connect.Code) {
	resp, err := f.ask.GetAskModel(t0(), connect.NewRequest(&modelsv1.GetAskModelRequest{Origin: "o", SessionToken: token}))
	if err != nil {
		return 0, connect.CodeOf(err)
	}
	return resp.Msg.GetOpinions(), 0
}

func t0() context.Context { return context.Background() }

func TestGetAskModelSaysWhetherOpinionsWork(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		demo  bool
		token string
		mod   func(s *modelrouter.Settings)
		want  modelsv1.AskOpinions
		code  connect.Code
	}{
		{"the owner, everything on", false, "owner", func(*modelrouter.Settings) {}, modelsv1.AskOpinions_ASK_OPINIONS_ON, 0},
		{"a viewer", false, "viewer", func(*modelrouter.Settings) {}, modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED, 0},
		{"a demo", true, "owner", func(*modelrouter.Settings) {}, modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED, 0},
		{"turned off", false, "owner", func(s *modelrouter.Settings) { s.Opinions = false }, modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED, 0},
		{"the switch off", false, "owner", func(s *modelrouter.Settings) { s.Switch.Enabled = false }, modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED, 0},
		{
			"a route that may not receive T1", false, "owner", func(s *modelrouter.Settings) { delete(s.Policy, modelrouter.DataTierT1) },
			modelsv1.AskOpinions_ASK_OPINIONS_ROUTE_INELIGIBLE, 0,
		},
		{"no session", false, "nobody", func(*modelrouter.Settings) {}, 0, connect.CodeUnauthenticated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newOpinionsFixture(t, tt.demo)
			tt.mod(&f.store.s)
			got, code := f.opinions(tt.token)
			if got != tt.want || code != tt.code {
				t.Errorf("opinions = %v, %v; want %v, %v", got, code, tt.want, tt.code)
			}
		})
	}
}

func TestOwnerTurnsOpinionsOnAndOffAndSettingsSayIt(t *testing.T) {
	t.Parallel()
	f := newOpinionsFixture(t, false)
	f.store.s.Opinions = false
	ctx := t.Context()
	get := func() *modelsv1.GetModelSettingsResponse {
		resp, err := f.settings.GetModelSettings(ctx, connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: "owner"}))
		if err != nil {
			t.Fatalf("GetModelSettings: %v", err)
		}
		return resp.Msg
	}
	if get().GetOpinionsEnabled() || get().GetDemo() {
		t.Fatalf("settings = %v, want opinions off by default and no demo", get())
	}
	if _, err := f.settings.SetOpinionsEnabled(ctx, connect.NewRequest(&modelsv1.SetOpinionsEnabledRequest{Origin: "o", SessionToken: "owner", Enabled: true})); err != nil {
		t.Fatalf("SetOpinionsEnabled(on): %v", err)
	}
	if !get().GetOpinionsEnabled() {
		t.Error("opinions are not on after turning them on")
	}
	if _, err := f.settings.SetOpinionsEnabled(ctx, connect.NewRequest(&modelsv1.SetOpinionsEnabledRequest{Origin: "o", SessionToken: "owner"})); err != nil {
		t.Fatalf("SetOpinionsEnabled(off): %v", err)
	}
	if get().GetOpinionsEnabled() {
		t.Error("opinions are on after turning them off")
	}
}

func TestOnlyTheOwnerTurnsOpinionsOnAndNeverADemo(t *testing.T) {
	t.Parallel()
	for token, want := range map[string]connect.Code{"": connect.CodeUnauthenticated, "read": connect.CodePermissionDenied} {
		f := newOpinionsFixture(t, false)
		f.store.s.Opinions = false
		_, err := f.settings.SetOpinionsEnabled(t.Context(), connect.NewRequest(&modelsv1.SetOpinionsEnabledRequest{Origin: "o", SessionToken: token, Enabled: true}))
		if connect.CodeOf(err) != want || f.store.s.Opinions {
			t.Errorf("SetOpinionsEnabled with %q = %v, stored %v; want %v and nothing stored", token, err, f.store.s.Opinions, want)
		}
	}
	f := newOpinionsFixture(t, true)
	f.store.s.Opinions = false
	_, err := f.settings.SetOpinionsEnabled(t.Context(), connect.NewRequest(&modelsv1.SetOpinionsEnabledRequest{Origin: "o", SessionToken: "owner", Enabled: true}))
	if connect.CodeOf(err) != connect.CodeFailedPrecondition || f.store.s.Opinions {
		t.Errorf("SetOpinionsEnabled(on) on a demo = %v, stored %v; want failed_precondition and nothing stored", err, f.store.s.Opinions)
	}
	resp, err := f.settings.GetModelSettings(t.Context(), connect.NewRequest(&modelsv1.GetModelSettingsRequest{Origin: "o", SessionToken: "owner"}))
	if err != nil || !resp.Msg.GetDemo() {
		t.Errorf("GetModelSettings on a demo = %v, %v; want demo set", resp, err)
	}
}
