package ui

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

func TestHumanWaitRoundsUpAndSpeaksOfHoursFromTwo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		d    time.Duration
		want string
	}{
		{-time.Minute, "less than a minute"},
		{0, "less than a minute"},
		{30 * time.Second, "less than a minute"},
		{time.Minute, "less than a minute"},
		{61 * time.Second, "about 2 minutes"},
		{40 * time.Minute, "about 40 minutes"},
		{90*time.Minute + time.Second, "about 91 minutes"},
		{119 * time.Minute, "about 119 minutes"},
		{2 * time.Hour, "about 2 hours"},
		{2*time.Hour + time.Second, "about 3 hours"},
		{23*time.Hour + 59*time.Minute, "about 24 hours"},
	}
	for _, tt := range tests {
		if got := humanWait(tt.d); got != tt.want {
			t.Errorf("humanWait(%s) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestAdmissionAnswerEachRefusalSaysWhyAndWhenAndKeepsTheCommands(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	in := now.Add
	tests := []struct {
		name      string
		adm       askAdmission
		wantTitle string
		wantText  string
		refused   bool
	}{
		{name: "allowed goes on", adm: askAdmission{Kind: admitAllowed}},
		{
			name: "the budget is spent", adm: askAdmission{Kind: admitBudgetSpent, ResetsAt: in(5 * time.Hour)}, refused: true,
			wantTitle: "The day's budget for the model is used up", wantText: "starts over in about 5 hours",
		},
		{
			name: "the session's limit", adm: askAdmission{Kind: admitSessionLimit, ResetsAt: in(40 * time.Minute)}, refused: true,
			wantTitle: "You've asked the model enough for now", wantText: "ask it again in about 40 minutes",
		},
		{
			name: "the address's limit", adm: askAdmission{Kind: admitAddressLimit, ResetsAt: in(10 * time.Second)}, refused: true,
			wantTitle: "Too many questions from your network", wantText: "ask it again in less than a minute",
		},
		{
			name: "a reset time that has passed", adm: askAdmission{Kind: admitSessionLimit, ResetsAt: in(-time.Minute)}, refused: true,
			wantTitle: "You've asked the model enough for now", wantText: "ask it again later",
		},
		{
			name: "no reset time", adm: askAdmission{Kind: admitBudgetSpent}, refused: true,
			wantTitle: "The day's budget for the model is used up", wantText: "starts over later",
		},
		{
			name: "a core that decided nothing allows nothing", adm: askAdmission{}, refused: true,
			wantTitle: "The model couldn't be checked", wantText: "couldn't confirm",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card, refused := admissionAnswer("how diversified am I?", tt.adm, now)
			if refused != tt.refused {
				t.Fatalf("refused = %v, want %v", refused, tt.refused)
			}
			if !refused {
				return
			}
			if card.Title != tt.wantTitle || !strings.Contains(card.Text, tt.wantText) {
				t.Errorf("card = %q / %q, want %q with %q", card.Title, card.Text, tt.wantTitle, tt.wantText)
			}
			if len(card.Help) == 0 || card.Source != "The model gave no answer" || card.Question != "how diversified am I?" {
				t.Errorf("card keeps no commands or no source: %+v", card)
			}
		})
	}
}

func TestAskBoxBudgetAndLimitsAreSaidAndCommandsStillWork(t *testing.T) {
	t.Parallel()
	soon := time.Now().Add(3*time.Hour + 30*time.Minute)
	tests := []struct {
		name      string
		admission modelsv1.AskAdmission
		wantTitle string
		wantText  string
	}{
		{"the budget is spent", modelsv1.AskAdmission_ASK_ADMISSION_BUDGET_SPENT, "The day&#39;s budget for the model is used up", "starts over in about 4 hours"},
		{"the session's limit", modelsv1.AskAdmission_ASK_ADMISSION_SESSION_LIMIT, "You&#39;ve asked the model enough for now", "ask it again in about 4 hours"},
		{"the address's limit", modelsv1.AskAdmission_ASK_ADMISSION_ADDRESS_LIMIT, "Too many questions from your network", "ask it again in about 4 hours"},
	}
	for _, tt := range tests {
		for _, session := range []string{ownerSession, readSession} {
			t.Run(tt.name+" "+session, func(t *testing.T) {
				t.Parallel()
				h := newHarness(t)
				h.askModel.turnOn("T0", "T2")
				h.askModel.admission, h.askModel.resetsAt = tt.admission, soon
				rec := h.asks(t, session, "how diversified am I?")
				wantStatus(t, rec, http.StatusOK)
				wantBody(t, rec, tt.wantTitle, tt.wantText, "The model gave no answer", `class="ask-help"`, "What needs me", "Allocation by currency")
				wantNoBody(t, rec, "Answered by", `class="used-card"`)
				if n := h.askModel.generated(); n != 0 {
					t.Fatalf("the model was asked %d times after the core refused the question", n)
				}
				// A fixed command is answered from the record and never counted against any limit.
				before := len(h.askModel.admitted)
				cmd := h.asks(t, session, "What needs me?")
				wantBody(t, cmd, "Answered from your data · no AI model")
				wantNoBody(t, cmd, "The model gave no answer")
				if got := len(h.askModel.admitted); got != before {
					t.Errorf("a command asked the core for admission (%d calls, was %d)", got, before)
				}
			})
		}
	}
}

func TestAskBoxAsksForAdmissionOncePerQuestionWithTheBrowsersAddress(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("allocation", `{"by":"currency"}`)}},
		scriptedReply{text: answerText("Two thirds of it is listed.", "c1")},
	)
	req := askPost("how diversified am I?", ownerSession)
	req.Header.Set(openfief.ClientIPHeader, "198.51.100.7")
	rec := h.serve(t, req)
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Two thirds of it is listed.")
	if got := h.askModel.generated(); got != 2 {
		t.Fatalf("the model was asked %d times, want a loop of 2 calls", got)
	}
	if len(h.askModel.admitted) != 1 || h.askModel.admitted[0] != "198.51.100.7" {
		t.Errorf("admission asked for %v, want once, with the address the core handed over", h.askModel.admitted)
	}
	// A browser that sends the header itself gets nothing from it: the core replaces it, so the UI
	// reads what it is given; with none, the address is empty and the core leaves the limit out.
	h2 := newHarness(t)
	h2.askModel.turnOn("T0", "T2")
	h2.asks(t, ownerSession, "how diversified am I?")
	if len(h2.askModel.admitted) != 1 || h2.askModel.admitted[0] != "" {
		t.Errorf("without the header admission asked for %v, want one empty address", h2.askModel.admitted)
	}
}

func TestAskBoxWhenTheCoreCannotAdmitFallsBackToTheCommands(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	h.askModel.admitErr = codeErr(connect.CodeInternal, nil)
	rec := h.asks(t, ownerSession, "how diversified am I?")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "The model couldn&#39;t be checked", "The model gave no answer", `class="ask-help"`)
	if n := h.askModel.generated(); n != 0 {
		t.Errorf("the model was asked %d times without an admission", n)
	}
}

func TestAskBoxGuardRefusalsMidQuestionAreSaid(t *testing.T) {
	t.Parallel()
	resets := time.Now().Add(90 * time.Minute).UTC()
	budget := connect.NewError(connect.CodeResourceExhausted, nil)
	budget.Meta().Set(refusalHeader, refusalBudget)
	budget.Meta().Set(resetsAtHeader, resets.Format(time.RFC3339))
	busy := connect.NewError(connect.CodeResourceExhausted, nil)
	busy.Meta().Set(refusalHeader, refusalBusy)
	other := connect.NewError(connect.CodeResourceExhausted, nil)
	tests := []struct {
		name     string
		err      error
		want     []string
		wantNone []string
	}{
		{"the budget ran out", budget, []string{"The day&#39;s budget for the model is used up", "starts over in about 90 minutes"}, nil},
		{"the calls at once were taken", busy, []string{"The model is busy", "Try again in a few seconds"}, nil},
		{"a refusal of no guard is a failure", other, []string{"The model couldn&#39;t answer"}, []string{"budget", "busy"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.askModel.turnOn("T0", "T2")
			h.askModel.failWith = tt.err
			rec := h.asks(t, ownerSession, "how diversified am I?")
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, append(tt.want, "The model gave no answer", `class="ask-help"`)...)
			wantNoBody(t, rec, tt.wantNone...)
		})
	}
}
