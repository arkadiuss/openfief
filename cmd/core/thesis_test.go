package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	thesisv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1/thesisv1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// TestDraftThesesAreTheOwnersAlone runs D-229 through the real core: the owner
// saves a draft, edits it into a second version and reads both; a viewer's
// session, a demo session and a script token as the caller, as the session or
// as a session's holder are each refused every call, and the read API serves
// no thesis a grant could reach (D-137, D-209).
func TestDraftThesesAreTheOwnersAlone(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	link, _, _ := strings.Cut(enrollCmd(t, dsn), "\n")
	code, ok := strings.CutPrefix(link, localUI+"/enroll#code=")
	if !ok {
		t.Fatalf("enroll-link printed %q", link)
	}
	_, srv, pool := startCoreForUI(t, dsn, func(cfg *serveConfig) { cfg.Auth.DemoSessions = true })
	loadDemo(t, pool)
	if _, err := pool.Exec(t.Context(), `CREATE SCHEMA demo;
		CREATE TABLE demo.marker (demo boolean PRIMARY KEY CHECK (demo), seed text NOT NULL, loaded_on date NOT NULL);
		INSERT INTO demo.marker VALUES (true, '1', '2026-10-28')`); err != nil {
		t.Fatalf("writing the demo marker: %v", err)
	}
	auth := authClient(srv.URL, srv.Client(), uiCaller)
	owner := passkeyLogin(t, auth, passkeyEnroll(t, auth, code))
	_, reg := viewerEnroll(t, auth, inviteAlice(t, auth, owner))
	demoSession, err := auth.CreateDemoSession(t.Context(), connect.NewRequest(&identityv1.CreateDemoSessionRequest{Origin: localUI}))
	if err != nil {
		t.Fatalf("CreateDemoSession: %v", err)
	}
	token, err := auth.CreateScriptToken(t.Context(), connect.NewRequest(&identityv1.CreateScriptTokenRequest{Origin: localUI, SessionToken: owner, Name: "nightly"}))
	if err != nil {
		t.Fatalf("CreateScriptToken: %v", err)
	}

	saved := checkOwnerKeepsDrafts(t, srv, owner)
	refused := []struct {
		name, caller, session string
		want                  connect.Code
	}{
		{"a viewer's session", uiCaller, reg.GetSessionToken(), connect.CodePermissionDenied},
		{"a demo session", uiCaller, demoSession.Msg.GetSessionToken(), connect.CodePermissionDenied},
		{"a script token as the session", uiCaller, token.Msg.GetSecret(), connect.CodeUnauthenticated},
		{"a script token as the caller", token.Msg.GetSecret(), owner, connect.CodePermissionDenied},
		{"a caller that is no view plugin", "wrong", owner, connect.CodePermissionDenied},
		{"no session", uiCaller, "", connect.CodeUnauthenticated},
	}
	for _, tt := range refused {
		c := thesisv1connect.NewThesisServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(callerToken(tt.caller)))
		draft := &thesisv1.ThesisDraft{}
		calls := map[string]func() error{
			"SaveThesis": func() error {
				return outcome(c.SaveThesis(t.Context(), connect.NewRequest(&thesisv1.SaveThesisRequest{Origin: localUI, SessionToken: tt.session, Draft: draft})))
			},
			"EditThesis": func() error {
				return outcome(c.EditThesis(t.Context(), connect.NewRequest(&thesisv1.EditThesisRequest{
					Origin: localUI, SessionToken: tt.session, ThesisId: saved, BaseVersion: 2, Draft: draft,
				})))
			},
			"GetThesis": func() error {
				return outcome(c.GetThesis(t.Context(), connect.NewRequest(&thesisv1.GetThesisRequest{Origin: localUI, SessionToken: tt.session, ThesisId: saved})))
			},
			"ListTheses": func() error {
				return outcome(c.ListTheses(t.Context(), connect.NewRequest(&thesisv1.ListThesesRequest{Origin: localUI, SessionToken: tt.session})))
			},
		}
		for name, call := range calls {
			if got := connect.CodeOf(call()); got != tt.want {
				t.Errorf("%s by %s = %s, want %s", name, tt.name, got, tt.want)
			}
		}
	}
	methods := viewsv1.File_openfief_views_v1_read_service_proto.Services().Get(0).Methods()
	for i := range methods.Len() {
		if name := string(methods.Get(i).Name()); strings.Contains(strings.ToLower(name), "thesis") {
			t.Errorf("the read API serves %s: a draft thesis is on no read path (D-137)", name)
		}
	}
}

// checkOwnerKeepsDrafts has the owner of session save a draft, edit it and
// read it, its list and its versions back, and returns the thesis's ID.
func checkOwnerKeepsDrafts(t *testing.T, srv *httptest.Server, owner string) string {
	t.Helper()
	ctx := t.Context()
	c := thesisv1connect.NewThesisServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(callerToken(uiCaller)))
	draft := &thesisv1.ThesisDraft{
		Statement: "Services revenue keeps growing.", Instruments: []string{"EXMP"}, Horizon: thesisv1.Horizon_HORIZON_LONG_TERM,
		Invalidation: []string{"Growth stalls for two quarters."}, Sources: []string{"The holdings card."},
	}
	saved, err := c.SaveThesis(ctx, connect.NewRequest(&thesisv1.SaveThesisRequest{Origin: localUI, SessionToken: owner, Draft: draft}))
	if err != nil {
		t.Fatalf("SaveThesis: %v", err)
	}
	id := saved.Msg.GetThesis().GetId()
	draft.Statement = "Services revenue keeps growing and margins widen."
	if _, editErr := c.EditThesis(ctx, connect.NewRequest(&thesisv1.EditThesisRequest{
		Origin: localUI, SessionToken: owner, ThesisId: id, BaseVersion: 1, Draft: draft,
	})); editErr != nil {
		t.Fatalf("EditThesis: %v", editErr)
	}
	got, err := c.GetThesis(ctx, connect.NewRequest(&thesisv1.GetThesisRequest{Origin: localUI, SessionToken: owner, ThesisId: id}))
	if err != nil || len(got.Msg.GetThesis().GetVersions()) != 2 || got.Msg.GetThesis().GetDataTier() != "T1" {
		t.Fatalf("GetThesis = %v, %v, want two versions of T1 data", got, err)
	}
	list, err := c.ListTheses(ctx, connect.NewRequest(&thesisv1.ListThesesRequest{Origin: localUI, SessionToken: owner}))
	if err != nil || len(list.Msg.GetTheses()) != 1 || list.Msg.GetTheses()[0].GetVersion() != 2 {
		t.Fatalf("ListTheses = %v, %v, want the one thesis at version 2", list, err)
	}
	return id
}
