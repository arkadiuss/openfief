package ui

import (
	"net/http"
	"strings"
	"testing"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// thesisReply is the model's thesis answer naming the card its first tool call made.
func thesisReply() scriptedReply {
	th := &structuredThesis{
		Statement: "Synthetic Fund keeps compounding.", Instruments: []string{"Synthetic Fund"}, Horizon: horizonLongTerm,
		Invalidation: []string{"Its fees double."}, Sources: []string{"The holdings card for Synthetic Fund."},
	}
	return scriptedReply{text: opinionsJSON(answerKindThesis, "Here is a thesis on it.", []string{"c1"}, th)}
}

func showFund() scriptedReply {
	return scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"Synthetic Fund","kind":"instrument"}`)}}
}

func TestTheOwnersAskBoxShowsAThesisCardWithASaveButtonAndSavesItInPlace(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T1", "T2")
	h.askModel.opinions = modelsv1.AskOpinions_ASK_OPINIONS_ON
	h.askModel.say(showFund(), thesisReply())
	rec := h.asks(t, ownerSession, "suggest a thesis on Synthetic Fund")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`aria-label="Suggested thesis"`, "Synthetic Fund keeps compounding.", "Wrong if", "Its fees double.", "The holdings card for Synthetic Fund.",
		"A suggestion from an AI model, not investment advice. It holds no lease size",
		`action="/settings/theses" hx-post="/settings/theses" hx-target="this" hx-swap="outerHTML"`,
		`name="statement" value="Synthetic Fund keeps compounding."`, `name="horizon" value="long_term"`,
		`name="instruments" value="Synthetic Fund"`, `name="invalidation" value="Its fees double."`,
		"Save as draft thesis",
		// The answer's note says what an opinion is.
		"AI can be wrong; check the cards. Not investment advice. This is an AI model&#39;s opinion from your cards and public prices, and it never acts.",
	)
	// Stop at suggest and save: nothing on the card starts an agent or a fief.
	wantNoBody(t, rec, "Start an agent", "Spawn", "Not investment advice. <a")
	first := h.askModel.requests[0][0]
	if !strings.Contains(first.GetMessage().GetContent(), "Refuse only these two things") || first.GetDataTier() != "instructions" {
		t.Errorf("the model's first message = %q (%s), want the opinions prompt as instructions", first.GetMessage().GetContent(), first.GetDataTier())
	}
	// Saving posts what the card holds, with htmx, and the card's place says it was saved.
	save := h.serve(t, htmx(withSession(form("/settings/theses", strings.Join([]string{
		"statement=Synthetic+Fund+keeps+compounding.", "horizon=long_term", "instruments=Synthetic+Fund",
		"invalidation=Its+fees+double.", "sources=The+holdings+card+for+Synthetic+Fund.",
	}, "&")), ownerSession)))
	wantStatus(t, save, http.StatusOK)
	wantBody(t, save, "Saved as a draft thesis.", "Open it on the Theses page")
	list := h.serve(t, withSession(request(http.MethodGet, "/settings/theses", ""), ownerSession))
	wantBody(t, list, "Synthetic Fund keeps compounding.")
}

func TestAViewerAndAnOwnerWithOpinionsOffGetNoThesisAndNoSaveButton(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		session string
		state   modelsv1.AskOpinions
	}{
		"a viewer, with opinions on for the owner": {viewerSession, modelsv1.AskOpinions_ASK_OPINIONS_ON},
		"a read session":               {readSession, modelsv1.AskOpinions_ASK_OPINIONS_ON},
		"the owner, with opinions off": {ownerSession, modelsv1.AskOpinions_ASK_OPINIONS_UNSPECIFIED},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.askModel.turnOn("T0", "T1", "T2")
			h.askModel.opinions = tt.state
			// A model that suggests a thesis anyway is not shown: its kind is none the plain answer has.
			h.askModel.say(showFund(), thesisReply(), thesisReply(), thesisReply())
			rec := h.asks(t, tt.session, "suggest a thesis on Synthetic Fund")
			wantStatus(t, rec, http.StatusOK)
			wantNoBody(t, rec, "Save as draft thesis", "Suggested thesis", "/settings/theses", "AI model&#39;s opinion")
			if first := h.askModel.requests[0][0].GetMessage().GetContent(); strings.Contains(first, "thesis") {
				t.Errorf("the prompt of %s mentions a thesis", name)
			}
		})
	}
}

func TestARouteThatKeepsOpinionsOffSaysSoOnTheOwnersRefusal(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")
	h.askModel.opinions = modelsv1.AskOpinions_ASK_OPINIONS_ROUTE_INELIGIBLE
	h.askModel.say(scriptedReply{text: answerJSON(answerKindRefuse, "I can't say whether to buy it.", nil, nil, nil)})
	rec := h.asks(t, ownerSession, "should I buy Synthetic Fund?")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "I can&#39;t say whether to buy it.", "Opinions are on in Settings, but no model on the Ask box&#39;s route is allowed to receive",
		"Allow one on the Models page.")
}

func TestAThesisAnswerWithoutTheCoresThesesServiceHasNoSaveButton(t *testing.T) {
	t.Parallel()
	h := newHarnessWith(t, nil, func(cfg *Config) { cfg.Theses = nil })
	h.askModel.turnOn("T0", "T1", "T2")
	h.askModel.opinions = modelsv1.AskOpinions_ASK_OPINIONS_ON
	h.askModel.say(showFund(), thesisReply())
	rec := h.asks(t, ownerSession, "suggest a thesis on Synthetic Fund")
	wantBody(t, rec, "Suggested thesis", "Synthetic Fund keeps compounding.")
	wantNoBody(t, rec, "Save as draft thesis")
}
