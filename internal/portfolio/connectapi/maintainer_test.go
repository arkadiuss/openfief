package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	gocmp "github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/portfolio/app"
	"github.com/arkadiuss/openfief/internal/portfolio/connectapi"
	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
)

const (
	readerSession  = "ofsess_reader"
	limitedSession = "ofsess_limited"
	proposalID     = "00000000000000p1"
)

var errCore = errors.New("a core failure that quotes ETFA 63.1")

// noonUTC is the fixed time of the stubs.
func noonUTC() time.Time { return time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC) }

// codeOf is the Connect code of err, and zero for no error.
func codeOf(err error) connect.Code {
	if err == nil {
		return 0
	}
	return connect.CodeOf(err)
}

func decimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// stubStore is a maintainer store with one note and one proposal whose status
// and failures a test sets.
type stubStore struct {
	status app.ProposalStatus
	err    error
	run    app.Run
	queued []string
}

func (s *stubStore) QueueRun(_ context.Context, r *app.Run) error {
	if len(s.queued) > 0 {
		return app.ErrRunInProgress
	}
	s.queued = append(s.queued, r.ID)
	return nil
}

func (s *stubStore) StartRun(context.Context, string, time.Time) (app.Run, bool, error) {
	return app.Run{}, false, nil
}
func (s *stubStore) Run(context.Context, string) (app.Run, bool, error) { return s.run, true, nil }
func (s *stubStore) FinishRun(context.Context, *app.RunFinish) ([]string, error) {
	return nil, nil
}

func (s *stubStore) LatestNote(context.Context) (app.Note, bool, error) {
	return app.Note{RunID: "00000000000000n1", Text: "ETFA drifted up.", Model: "local/synth", At: noonUTC(), Trigger: app.TriggerWeekly}, true, s.err
}
func (s *stubStore) LatestRun(context.Context) (app.Run, bool, error) { return s.run, true, nil }

func (s *stubStore) PendingProposals(context.Context) ([]string, error) {
	return []string{proposalID, "00000000000000p2"}, nil
}

func (s *stubStore) stored(t *testing.T) app.StoredProposal {
	t.Helper()
	return app.StoredProposal{
		ID: proposalID, Status: s.status, CreatedAt: noonUTC(), Model: "local/synth", Note: "ETFA drifted up.",
		Proposal: portfolio.Proposal{Changes: []portfolio.Change{
			{Target: "instrument:IETEST000026/ETFA", Before: decimal(t, "60"), After: decimal(t, "55"), Reason: "Trim.", Held: true, Current: decimal(t, "63.1"), Drift: decimal(t, "3.1")},
			{Target: "index:FAKE20", Before: decimal(t, "40"), After: decimal(t, "45"), Reason: "Add."},
		}},
	}
}

// proposals serves the stored proposal for proposalID only.
type proposals struct {
	*stubStore

	t       *testing.T
	applied error
}

func (p proposals) Proposal(_ context.Context, id string) (app.StoredProposal, bool, error) {
	if id != proposalID {
		return app.StoredProposal{}, false, nil
	}
	return p.stored(p.t), true, nil
}

func (p proposals) ApplyProposal(context.Context, string, time.Time, func(*portfolio.Model, *portfolio.Proposal) (portfolio.Model, error), func(context.Context) error) error {
	return p.applied
}

func (p proposals) DismissProposal(_ context.Context, id string, _ time.Time, _ func(context.Context) error) error {
	if id != proposalID {
		return app.ErrUnknownProposal
	}
	return p.applied
}

type onGate bool

func (g onGate) MaintainerOn(context.Context) (bool, error) { return bool(g), nil }

type quietDecisions struct{}

func (quietDecisions) Raise(context.Context, string, string, string) error { return nil }
func (quietDecisions) Clear(context.Context, string, string) error         { return nil }

type quietAudit struct{}

func (quietAudit) Record(context.Context, *app.AuditRecord) error { return nil }

type quietQueue struct{}

func (quietQueue) Enqueue(context.Context, string) error { return nil }

type threeSessions struct{}

func (threeSessions) Reader(_ context.Context, _, token string) (app.Reader, error) {
	switch token {
	case ownerSession:
		return app.Reader{SessionID: "00000000000000a1", Owner: true, ViewsAll: true}, nil
	case readerSession:
		return app.Reader{SessionID: "00000000000000b2", ViewsAll: true}, nil
	case limitedSession:
		return app.Reader{SessionID: "00000000000000c3"}, nil
	}
	return app.Reader{}, app.ErrNoSession
}

// maintainerClient is the generated client, as a struct so helpers return no
// interface.
type maintainerClient struct {
	portfoliov1connect.MaintainerServiceClient
}

// newMaintainerClient serves a Maintainer over the stubs; the model portfolio
// is on and the maintainer's gate as on says.
func newMaintainerClient(t *testing.T, st *proposals, on bool, caller string) maintainerClient {
	t.Helper()
	model := portfolio.Model{Enabled: true, Targets: []portfolio.Target{{Kind: portfolio.TargetIndex, Name: "FAKE20", Currency: mustPLN(t), Weight: decimal(t, "100")}}}
	m := app.NewMaintainer(&app.MaintainerPorts{
		Runs: st, Proposals: st, Model: &memStore{model: model}, Gate: onGate(on), Decisions: quietDecisions{}, Audit: quietAudit{},
		Queue: quietQueue{}, Sessions: threeSessions{},
		IDs: func() (string, error) { return "00000000000000d1", nil }, Now: noonUTC,
	})
	_, handler := connectapi.NewMaintainer(m, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return maintainerClient{portfoliov1connect.NewMaintainerServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

func mustPLN(t *testing.T) kernel.Currency {
	t.Helper()
	c, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestGetMaintainerByReader(t *testing.T) {
	t.Parallel()
	st := &stubStore{run: app.Run{ID: "00000000000000f1", Trigger: app.TriggerWeekly, State: app.RunFailed, Class: "unavailable", At: noonUTC()}}
	c := newMaintainerClient(t, &proposals{stubStore: st, t: t}, true, viewCaller)
	ctx := t.Context()
	wantNote := &portfoliov1.Note{
		RunId: "00000000000000n1", Text: "ETFA drifted up.", CreatedAt: "2026-11-13T12:00:00Z", Model: "local/synth",
		Trigger: portfoliov1.RunTrigger_RUN_TRIGGER_WEEKLY,
	}
	owner, err := c.GetMaintainer(ctx, connect.NewRequest(&portfoliov1.GetMaintainerRequest{Origin: local, SessionToken: ownerSession}))
	if err != nil {
		t.Fatalf("GetMaintainer as the owner: %v", err)
	}
	wantOwner := &portfoliov1.GetMaintainerResponse{
		Available: true, Note: wantNote, PendingProposals: 2,
		LastFailure: &portfoliov1.RunFailure{At: "2026-11-13T12:00:00Z", ErrorClass: "unavailable", Trigger: portfoliov1.RunTrigger_RUN_TRIGGER_WEEKLY},
	}
	if diff := gocmp.Diff(wantOwner, owner.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("owner's view (-want +got):\n%s", diff)
	}
	reader, err := c.GetMaintainer(ctx, connect.NewRequest(&portfoliov1.GetMaintainerRequest{Origin: local, SessionToken: readerSession}))
	if err != nil {
		t.Fatalf("GetMaintainer as a reader: %v", err)
	}
	if diff := gocmp.Diff(&portfoliov1.GetMaintainerResponse{Available: true, Note: wantNote}, reader.Msg, protocmp.Transform()); diff != "" {
		t.Errorf("reader's view has the owner's failure or count (-want +got):\n%s", diff)
	}
	for token, want := range map[string]connect.Code{limitedSession: connect.CodePermissionDenied, "nope": connect.CodeUnauthenticated} {
		_, err := c.GetMaintainer(ctx, connect.NewRequest(&portfoliov1.GetMaintainerRequest{Origin: local, SessionToken: token}))
		if connect.CodeOf(err) != want {
			t.Errorf("GetMaintainer as %s: %v, want %v", token, err, want)
		}
	}
}

func TestOnlyViewPluginsMayCallTheMaintainer(t *testing.T) {
	t.Parallel()
	c := newMaintainerClient(t, &proposals{stubStore: &stubStore{}, t: t}, true, "some-other-token")
	_, err := c.GetMaintainer(t.Context(), connect.NewRequest(&portfoliov1.GetMaintainerRequest{Origin: local, SessionToken: ownerSession}))
	if connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("a caller that is no view plugin got %v, want PermissionDenied", err)
	}
}

func TestAskMaintainer(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		token   string
		on      bool
		want    connect.Code
		started bool
	}{
		{name: "the owner starts a run", token: ownerSession, on: true, started: true},
		{name: "a reader who may view everything is refused", token: readerSession, on: true, want: connect.CodePermissionDenied},
		{name: "no session", token: "nope", on: true, want: connect.CodeUnauthenticated},
		{name: "with the maintainer off", token: ownerSession, on: false, want: connect.CodeFailedPrecondition},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newMaintainerClient(t, &proposals{stubStore: &stubStore{}, t: t}, tt.on, viewCaller)
			got, err := c.AskMaintainer(t.Context(), connect.NewRequest(&portfoliov1.AskMaintainerRequest{Origin: local, SessionToken: tt.token}))
			if codeOf(err) != tt.want {
				t.Fatalf("AskMaintainer = %v, want %v", err, tt.want)
			}
			if err == nil && got.Msg.GetStarted() != tt.started {
				t.Errorf("started = %v, want %v", got.Msg.GetStarted(), tt.started)
			}
		})
	}
	t.Run("a run already queued is not started again", func(t *testing.T) {
		t.Parallel()
		c := newMaintainerClient(t, &proposals{stubStore: &stubStore{queued: []string{"x"}}, t: t}, true, viewCaller)
		got, err := c.AskMaintainer(t.Context(), connect.NewRequest(&portfoliov1.AskMaintainerRequest{Origin: local, SessionToken: ownerSession}))
		if err != nil || got.Msg.GetStarted() {
			t.Errorf("AskMaintainer = %v, %v; want started false and no error", got, err)
		}
	})
}

// proposalCall is one of the owner's calls on a proposal.
type proposalCall struct {
	name string
	call func(ctx context.Context, c maintainerClient, token, id string) error
}

// wrapped describes the failure of a call, nil as nil.
func wrapped(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("the call: %w", err)
}

// proposalCalls are the three calls on a proposal, as errors alone.
func proposalCalls() []proposalCall {
	return []proposalCall{
		{"get", func(ctx context.Context, c maintainerClient, token, id string) error {
			_, err := c.GetProposal(ctx, connect.NewRequest(&portfoliov1.GetProposalRequest{Origin: local, SessionToken: token, Id: id}))
			return wrapped(err)
		}},
		{"apply", func(ctx context.Context, c maintainerClient, token, id string) error {
			_, err := c.ApplyProposal(ctx, connect.NewRequest(&portfoliov1.ApplyProposalRequest{Origin: local, SessionToken: token, Id: id}))
			return wrapped(err)
		}},
		{"dismiss", func(ctx context.Context, c maintainerClient, token, id string) error {
			_, err := c.DismissProposal(ctx, connect.NewRequest(&portfoliov1.DismissProposalRequest{Origin: local, SessionToken: token, Id: id}))
			return wrapped(err)
		}},
	}
}

func TestGetProposalReturnsTheFiguresToTheOwner(t *testing.T) {
	t.Parallel()
	c := newMaintainerClient(t, &proposals{stubStore: &stubStore{status: app.ProposalPending}, t: t}, true, viewCaller)
	got, err := c.GetProposal(t.Context(), connect.NewRequest(&portfoliov1.GetProposalRequest{Origin: local, SessionToken: ownerSession, Id: proposalID}))
	if err != nil {
		t.Fatalf("GetProposal: %v", err)
	}
	want := &portfoliov1.Proposal{
		Id: proposalID, Status: portfoliov1.ProposalStatus_PROPOSAL_STATUS_PENDING, CreatedAt: "2026-11-13T12:00:00Z", Model: "local/synth", Note: "ETFA drifted up.",
		Changes: []*portfoliov1.TargetChange{
			{Target: "instrument:IETEST000026/ETFA", BeforePercent: "60", AfterPercent: "55", Held: true, CurrentPercent: "63.1", DriftPoints: "3.1", Reason: "Trim."},
			{Target: "index:FAKE20", BeforePercent: "40", AfterPercent: "45", Reason: "Add."},
		},
	}
	if diff := gocmp.Diff(want, got.Msg.GetProposal(), protocmp.Transform()); diff != "" {
		t.Errorf("proposal (-want +got):\n%s", diff)
	}
}

func TestApplyAndDismissMapTheUseCasesErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		apply   error
		wantErr connect.Code
	}{
		{name: "taken"},
		{name: "already decided", apply: app.ErrProposalDecided, wantErr: connect.CodeFailedPrecondition},
		{name: "the model changed", apply: portfolio.ErrStaleProposal, wantErr: connect.CodeFailedPrecondition},
		{name: "no valid model", apply: portfolio.ErrInvalidProposal, wantErr: connect.CodeFailedPrecondition},
		{name: "a core failure", apply: errCore, wantErr: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newMaintainerClient(t, &proposals{stubStore: &stubStore{status: app.ProposalApplied}, t: t, applied: tt.apply}, true, viewCaller)
			for _, call := range proposalCalls()[1:] {
				err := call.call(t.Context(), c, ownerSession, proposalID)
				if codeOf(err) != tt.wantErr {
					t.Errorf("%s: %v, want %v", call.name, err, tt.wantErr)
				}
				if err != nil && (strings.Contains(err.Error(), "63.1") || strings.Contains(err.Error(), "ETFA")) {
					t.Errorf("%s: the error quotes a figure or a name: %v", call.name, err)
				}
			}
		})
	}
}

func TestAnUnknownProposalIsNotFound(t *testing.T) {
	t.Parallel()
	c := newMaintainerClient(t, &proposals{stubStore: &stubStore{}, t: t}, true, viewCaller)
	for _, call := range []proposalCall{proposalCalls()[0], proposalCalls()[2]} {
		if err := call.call(t.Context(), c, ownerSession, "ffffffffffffffff"); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("%s of an unknown ID: %v, want NotFound", call.name, err)
		}
	}
}

func TestProposalCallsAreTheOwnersAlone(t *testing.T) {
	t.Parallel()
	c := newMaintainerClient(t, &proposals{stubStore: &stubStore{}, t: t}, true, viewCaller)
	for _, token := range []string{readerSession, limitedSession} {
		for _, call := range proposalCalls() {
			if err := call.call(t.Context(), c, token, proposalID); connect.CodeOf(err) != connect.CodePermissionDenied {
				t.Errorf("%s as %s: %v, want PermissionDenied", call.name, token, err)
			}
		}
	}
}
