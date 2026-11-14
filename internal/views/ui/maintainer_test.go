package ui

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	portfoliov1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/portfolio/v1/portfoliov1connect"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// proposalID is the synthetic proposal of the fakes: its run's ID.
const proposalID = "00112233445566aa"

var (
	errMaintainerOff = errors.New("the maintainer is off: turn on the models switch and the maintainer, and the model portfolio")
	errStale         = errors.New("the model portfolio changed since the proposal was made: dismiss it")
)

// fakeMaintainer is an in-memory MaintainerService: it wants the plugin's
// token and a known session, and the owner's for everything but the note,
// which the sessions in noteDenied may not read.
type fakeMaintainer struct {
	portfoliov1connect.UnimplementedMaintainerServiceHandler

	mu sync.Mutex
	// state is what GetMaintainer answers.
	state *portfoliov1.GetMaintainerResponse
	// unavailable makes GetMaintainer fail, as a core that can't read its
	// database does.
	unavailable bool
	// noteDenied are the sessions that may not read the note, as a viewer of
	// one account may not.
	noteDenied map[string]bool
	// proposal is what GetProposal answers for proposalID.
	proposal *portfoliov1.Proposal
	// askStarted is what AskMaintainer answers; askErr, applyErr and
	// dismissErr, when set, are what each of the others return.
	askStarted                   bool
	askErr, applyErr, dismissErr error
	// asked, applied and dismissed are the session tokens of the calls.
	asked, applied, dismissed []string
	// proposalReads counts GetProposal calls.
	proposalReads int
}

func newFakeMaintainer() *fakeMaintainer {
	return &fakeMaintainer{
		askStarted: true, noteDenied: map[string]bool{},
		state: &portfoliov1.GetMaintainerResponse{
			Available: true,
			Note: &portfoliov1.Note{
				RunId: proposalID, Text: "FAKE20 is the part that drifted; the rest is where it should be.", CreatedAt: "2026-11-13T06:30:12Z",
				Model: "local/synth", Trigger: portfoliov1.RunTrigger_RUN_TRIGGER_WEEKLY,
			},
		},
		proposal: pendingProposal(),
	}
}

// pendingProposal is a proposal of two changes, one of an instrument held and
// one of an index, which is not.
func pendingProposal() *portfoliov1.Proposal {
	return &portfoliov1.Proposal{
		Id: proposalID, Status: portfoliov1.ProposalStatus_PROPOSAL_STATUS_PENDING, CreatedAt: "2026-11-13T06:30:12Z",
		Model: "local/synth", Note: "FAKE20 is the part that drifted.",
		Changes: []*portfoliov1.TargetChange{
			{Target: "instrument:XX0000000001/SYN", BeforePercent: "60", AfterPercent: "55", Held: true, CurrentPercent: "63.1", DriftPoints: "3.1", Reason: "Overweight and still rising."},
			{Target: "index:FAKE20", BeforePercent: "40", AfterPercent: "45", Reason: "Underweight."},
		},
	}
}

func (f *fakeMaintainer) GetMaintainer(_ context.Context, r *connect.Request[portfoliov1.GetMaintainerRequest]) (*connect.Response[portfoliov1.GetMaintainerResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case f.unavailable:
		return nil, codeErr(connect.CodeUnavailable, nil)
	case f.noteDenied[r.Msg.GetSessionToken()]:
		return nil, codeErr(connect.CodePermissionDenied, nil)
	}
	out := &portfoliov1.GetMaintainerResponse{Available: f.state.GetAvailable(), Note: f.state.GetNote(), Running: f.state.GetRunning()}
	if owner(r.Msg.GetSessionToken()) == nil {
		out.LastFailure, out.PendingProposals = f.state.GetLastFailure(), f.state.GetPendingProposals()
	}
	return connect.NewResponse(out), nil
}

func (f *fakeMaintainer) AskMaintainer(_ context.Context, r *connect.Request[portfoliov1.AskMaintainerRequest]) (*connect.Response[portfoliov1.AskMaintainerResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.asked = append(f.asked, r.Msg.GetSessionToken())
	if f.askErr != nil {
		return nil, codeErr(connect.CodeFailedPrecondition, f.askErr)
	}
	return connect.NewResponse(&portfoliov1.AskMaintainerResponse{Started: f.askStarted}), nil
}

func (f *fakeMaintainer) GetProposal(_ context.Context, r *connect.Request[portfoliov1.GetProposalRequest]) (*connect.Response[portfoliov1.GetProposalResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.proposalReads++
	if r.Msg.GetId() != f.proposal.GetId() {
		return nil, codeErr(connect.CodeNotFound, nil)
	}
	return connect.NewResponse(&portfoliov1.GetProposalResponse{Proposal: f.proposal}), nil
}

func (f *fakeMaintainer) ApplyProposal(_ context.Context, r *connect.Request[portfoliov1.ApplyProposalRequest]) (*connect.Response[portfoliov1.ApplyProposalResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.applied = append(f.applied, r.Msg.GetId())
	if f.applyErr != nil {
		return nil, codeErr(connect.CodeFailedPrecondition, f.applyErr)
	}
	return connect.NewResponse(&portfoliov1.ApplyProposalResponse{Proposal: f.proposal}), nil
}

func (f *fakeMaintainer) DismissProposal(_ context.Context, r *connect.Request[portfoliov1.DismissProposalRequest]) (*connect.Response[portfoliov1.DismissProposalResponse], error) {
	if err := f.admit(r.Header(), r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	if err := owner(r.Msg.GetSessionToken()); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dismissed = append(f.dismissed, r.Msg.GetId())
	if f.dismissErr != nil {
		return nil, codeErr(connect.CodeFailedPrecondition, f.dismissErr)
	}
	return connect.NewResponse(&portfoliov1.DismissProposalResponse{Proposal: f.proposal}), nil
}

// admit checks the plugin's token and returns the session's error, if any.
func (f *fakeMaintainer) admit(h http.Header, token string) error {
	if h.Get("Authorization") != "Bearer "+pluginToken {
		return codeErr(connect.CodePermissionDenied, nil)
	}
	_, err := scopeOf(token)
	return err
}

func (f *fakeMaintainer) locked(fn func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn()
}

// proposalItem is the Inbox's item of the fake proposal.
func proposalItem(state viewsv1.AttentionItemState) *viewsv1.AttentionItem {
	return &viewsv1.AttentionItem{
		Id: 70, Kind: kindTargetProposal, Subject: proposalID, Detail: "The maintainer proposes changing 2 target weight(s). Apply or dismiss it.",
		State: state, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS,
	}
}

func TestModelPageShowsTargetsAndTheLatestNote(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.read.locked(func() { h.read.modelOn = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/model", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`id="targets-title"`, "SYN (XX0000000001)", "60%", "FAKE20 (index, PLN)", "40%",
		`id="maintainer-note"`, "FAKE20 is the part that drifted; the rest is where it should be.",
		"2026-11-13 06:30 UTC", "local/synth", "the weekly run",
	)
	wantNoBody(t, rec, "Ask the maintainer", `action="/model/maintainer/ask"`)
	if got := seenOne(h.maintainer, func() string { return strings.Join(h.maintainer.asked, ",") }); got != "" {
		t.Errorf("a page view asked the maintainer: %q", got)
	}
}

func TestModelPageOffersTheOwnerToAsk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mod     func(f *fakeMaintainer)
		wantAsk bool
		want    []string
		not     []string
	}{
		{name: "available", mod: func(*fakeMaintainer) {}, wantAsk: true},
		{
			name: "a run under way", mod: func(f *fakeMaintainer) { f.state.Running = true },
			want: []string{"A run is under way"}, not: []string{"Ask the maintainer"},
		},
		{
			name: "the maintainer off", mod: func(f *fakeMaintainer) { f.state.Available, f.state.Note = false, nil },
			want: []string{"The maintainer is off", "No note yet"}, not: []string{"Ask the maintainer"},
		},
		{
			name: "a run that left no note and proposals waiting", mod: func(f *fakeMaintainer) {
				f.state.LastFailure = &portfoliov1.RunFailure{At: "2026-11-13T06:30:12Z", ErrorClass: "denied", Trigger: portfoliov1.RunTrigger_RUN_TRIGGER_WEEKLY}
				f.state.PendingProposals = 2
			},
			wantAsk: true, want: []string{"ended without a note", "whole-portfolio data (T0)", "2 proposal(s) of new target weights wait in the Inbox"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.read.locked(func() { h.read.modelOn = true })
			h.maintainer.locked(func() { tt.mod(h.maintainer) })
			rec := h.serve(t, withSession(request(http.MethodGet, "/model", ""), ownerSession))
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, tt.want...)
			wantNoBody(t, rec, tt.not...)
			if has := strings.Contains(rec.Body.String(), `action="/model/maintainer/ask"`); has != tt.wantAsk {
				t.Errorf("the ask form is shown = %v, want %v", has, tt.wantAsk)
			}
		})
	}
}

func TestModelPageHidesWhatItMayNotShow(t *testing.T) {
	t.Parallel()
	t.Run("the model portfolio off shows nothing of it", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		rec := h.serve(t, withSession(request(http.MethodGet, "/model", ""), ownerSession))
		wantStatus(t, rec, http.StatusNotFound)
		wantNoBody(t, rec, "FAKE20", "maintainer")
	})
	t.Run("a session that may not read the note sees the targets only", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.read.locked(func() { h.read.modelOn = true })
		h.maintainer.locked(func() { h.maintainer.noteDenied[readSession] = true })
		rec := h.serve(t, withSession(request(http.MethodGet, "/model", ""), readSession))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "FAKE20 (index, PLN)")
		wantNoBody(t, rec, "maintainer-note", "FAKE20 is the part that drifted", `id="maintainer-title"`)
	})
	t.Run("a core that can't say still shows the targets", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.read.locked(func() { h.read.modelOn = true })
		h.maintainer.locked(func() { h.maintainer.unavailable = true })
		rec := h.serve(t, withSession(request(http.MethodGet, "/model", ""), ownerSession))
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "FAKE20 (index, PLN)", "could not say how the maintainer stands")
	})
	t.Run("without a session", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		rec := h.serve(t, request(http.MethodGet, "/model", ""))
		wantStatus(t, rec, http.StatusSeeOther)
	})
}

func TestAskTheMaintainer(t *testing.T) {
	t.Parallel()
	t.Run("the owner starts a run", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.read.locked(func() { h.read.modelOn = true })
		rec := h.serve(t, withSession(form("/model/maintainer/ask", ""), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/model?asked=1" {
			t.Errorf("Location = %q, want /model?asked=1", got)
		}
		if got := seen(h.maintainer, func() []string { return h.maintainer.asked }); len(got) != 1 || got[0] != ownerSession {
			t.Errorf("the core was asked by %v, want the owner's session once", got)
		}
		page := h.serve(t, withSession(request(http.MethodGet, "/model?asked=1", ""), ownerSession))
		wantBody(t, page, "The maintainer was asked")
	})
	t.Run("a run already queued is said, not started again", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.maintainer.locked(func() { h.maintainer.askStarted = false })
		rec := h.serve(t, withSession(form("/model/maintainer/ask", ""), ownerSession))
		wantStatus(t, rec, http.StatusSeeOther)
		if got := rec.Header().Get("Location"); got != "/model?asked=busy" {
			t.Errorf("Location = %q, want /model?asked=busy", got)
		}
	})
	t.Run("a read session is refused by the page and never reaches the core", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		wantStatus(t, h.serve(t, withSession(form("/model/maintainer/ask", ""), readSession)), http.StatusForbidden)
		if got := seen(h.maintainer, func() []string { return h.maintainer.asked }); len(got) != 0 {
			t.Errorf("the core was asked by %v", got)
		}
	})
	t.Run("the core's refusal is shown with its reason", func(t *testing.T) {
		t.Parallel()
		h := newHarness(t)
		h.read.locked(func() { h.read.modelOn = true })
		h.maintainer.locked(func() { h.maintainer.askErr = errMaintainerOff })
		rec := h.serve(t, withSession(form("/model/maintainer/ask", ""), ownerSession))
		wantStatus(t, rec, http.StatusBadRequest)
		wantBody(t, rec, "The core refused this", "the maintainer is off")
	})
}

func TestInboxShowsTheProposalWithItsFigures(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(proposalItem(viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN))
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"The maintainer proposes new target weights",
		"XX0000000001/SYN", "FAKE20 (index)", "Overweight and still rising.", "63.1%", "&#43;3.1 pp", "not held", ">60%<", ">55%<", ">40%<", ">45%<",
		`action="/inbox/proposals/apply"`, `action="/inbox/proposals/dismiss"`, `name="id" value="`+proposalID+`"`,
		"nothing is bought or sold", "The maintainer's note", "Open the model portfolio",
	)
}

func TestInboxShowsAViewerNoFiguresAndNoButtons(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(proposalItem(viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN))
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), readSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "The maintainer proposes new target weights", "Waiting for the owner.")
	wantNoBody(t, rec, "63.1%", "Overweight and still rising.", `action="/inbox/proposals/apply"`, `action="/inbox/proposals/dismiss"`)
	if got := seenOne(h.maintainer, func() string { return strings.Repeat("x", h.maintainer.proposalReads) }); got != "" {
		t.Errorf("a read session's Inbox asked the core for the proposal %d times", len(got))
	}
}

func TestInboxShowsADecidedProposalWithoutButtons(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.maintainer.locked(func() {
		h.maintainer.proposal.Status, h.maintainer.proposal.DecidedAt = portfoliov1.ProposalStatus_PROPOSAL_STATUS_APPLIED, "2026-11-14T08:00:00Z"
	})
	h.attention.add(proposalItem(viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_RESOLVED))
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox?view=done", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Applied on 2026-11-14 08:00 UTC", "Nothing was bought or sold")
	wantNoBody(t, rec, `action="/inbox/proposals/apply"`)
}

// appliedBy and dismissedBy are the proposals the fake core was asked to apply
// and to dismiss.
func appliedBy(h *harness) []string {
	return seen(h.maintainer, func() []string { return h.maintainer.applied })
}

func dismissedBy(h *harness) []string {
	return seen(h.maintainer, func() []string { return h.maintainer.dismissed })
}

func TestApplyReachesTheCoreWithTheOwnersSessionAndSaysNothingWasTraded(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/inbox/proposals/apply", "id="+proposalID), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/inbox?applied=1" {
		t.Errorf("Location = %q, want /inbox?applied=1", got)
	}
	if applied := appliedBy(h); len(applied) != 1 || applied[0] != proposalID {
		t.Errorf("applied = %v, want the proposal once", applied)
	}
	page := h.serve(t, withSession(request(http.MethodGet, "/inbox?applied=1", ""), ownerSession))
	wantBody(t, page, "The model portfolio&#39;s targets changed. Nothing was bought or sold.")
}

func TestDismissReachesTheCore(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/inbox/proposals/dismiss", "id="+proposalID), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/inbox?dismissed=1" {
		t.Errorf("Location = %q, want /inbox?dismissed=1", got)
	}
	if dismissed := dismissedBy(h); len(dismissed) != 1 || dismissed[0] != proposalID {
		t.Errorf("dismissed = %v, want the proposal once", dismissed)
	}
}

func TestADecisionOfAReadSessionIsRefusedBeforeTheCore(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, target := range []string{"/inbox/proposals/apply", "/inbox/proposals/dismiss"} {
		wantStatus(t, h.serve(t, withSession(form(target, "id="+proposalID), readSession)), http.StatusForbidden)
	}
	if a, d := appliedBy(h), dismissedBy(h); len(a)+len(d) != 0 {
		t.Errorf("the core was called: applied %v, dismissed %v", a, d)
	}
}

func TestAMalformedProposalIDNeverReachesTheCore(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, id := range []string{"", "nope", proposalID + "00", "00112233445566AA", "../../inbox"} {
		wantStatus(t, h.serve(t, withSession(form("/inbox/proposals/apply", "id="+id), ownerSession)), http.StatusBadRequest)
		wantStatus(t, h.serve(t, withSession(form("/inbox/proposals/dismiss", "id="+id), ownerSession)), http.StatusBadRequest)
	}
	if a, d := appliedBy(h), dismissedBy(h); len(a)+len(d) != 0 {
		t.Errorf("the core was asked to apply %v and dismiss %v", a, d)
	}
}

func TestACoreRefusalOfADecisionShowsItsReasonInTheInbox(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.maintainer.locked(func() { h.maintainer.applyErr = errStale })
	rec := h.serve(t, withSession(form("/inbox/proposals/apply", "id="+proposalID), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this", "the model portfolio changed since the proposal was made: dismiss it")
}

func TestMaintainerKindsHaveTheirWordsInTheInbox(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.attention.add(&viewsv1.AttentionItem{
		Id: 71, Kind: kindMaintainerFailing, Subject: "maintainer", Detail: "The maintainer's weekly run has ended without a note 2 times in a row; the last one ended as \"denied\".",
		State: viewsv1.AttentionItemState_ATTENTION_ITEM_STATE_OPEN, Group: viewsv1.AttentionItemGroup_ATTENTION_ITEM_GROUP_DECISIONS,
	})
	rec := h.serve(t, withSession(request(http.MethodGet, "/inbox", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "The maintainer&#39;s weekly run keeps failing", `href="/model"`, "Open the model portfolio")
}

func TestProposalWords(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"instrument:XX0000000001/SYN": "XX0000000001/SYN",
		"index:FAKE20":                "FAKE20 (index)",
		"other":                       "other",
	} {
		if got := proposalTarget(in); got != want {
			t.Errorf("proposalTarget(%q) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{"3.1": "+3.1", "-20": "-20", "0": "0", "": ""} {
		if got := driftPoints(in); got != want {
			t.Errorf("driftPoints(%q) = %q, want %q", in, got, want)
		}
	}
	if got := failureWords("nonsense"); got != "an unknown reason" {
		t.Errorf("failureWords(nonsense) = %q", got)
	}
	if got := whenOf("not a time"); got != "not a time" {
		t.Errorf("whenOf(not a time) = %q, want it as it is", got)
	}
}
