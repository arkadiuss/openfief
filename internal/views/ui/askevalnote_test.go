package ui

import (
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

const (
	noteDisclaimer = "AI can be wrong; check the cards. Not investment advice."
	noteLinkText   = "How models did in our evaluation"
	noteDefaultURL = `href="https://openfief.dev/docs/reference/ask-eval/"`
)

// TestModelAnswersShowWhoAnsweredWithADisclaimerAndTheEvalLink checks D-222
// through the page: an answer a model wrote shows the model's name, the
// disclaimer and the link to the eval's results, and names a model the eval
// has not covered as not evaluated.
func TestModelAnswersShowWhoAnsweredWithADisclaimerAndTheEvalLink(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		model    string
		reply    string
		want     []string
		wantNone []string
	}{
		{
			name: "an evaluated model", model: "openrouter/openai/gpt-oss-120b", reply: answerText("It is 10."),
			want:     []string{"Answered by openai/gpt-oss-120b · used none of your data.", noteDisclaimer, noteLinkText, noteDefaultURL, `class="muted ask-note"`},
			wantNone: []string{"not evaluated"},
		},
		{
			name: "a model the eval has not covered", model: "p/m", reply: answerText("It is 10."),
			want: []string{"Answered by m (not evaluated by OpenFief) · used none of your data.", noteDisclaimer, noteLinkText, noteDefaultURL},
		},
		{
			name: "a refusal the model wrote is from it, as the header says", model: "p/m", reply: answerJSON(answerKindRefuse, "I can only read.", nil, nil, nil),
			want: []string{"From m (not evaluated by OpenFief).", noteDisclaimer, noteLinkText},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			h.askModel.turnOn("T0", "T2")
			h.askModel.model = tt.model
			h.askModel.say(scriptedReply{text: tt.reply})
			rec := h.asks(t, ownerSession, "tell me something")
			wantStatus(t, rec, http.StatusOK)
			wantBody(t, rec, tt.want...)
			wantNoBody(t, rec, tt.wantNone...)
		})
	}
}

// TestCommandAnswersShowNoModelNote checks that the commands' own answers,
// which no model wrote, carry none of the note (D-222).
func TestCommandAnswersShowNoModelNote(t *testing.T) {
	t.Parallel()
	for _, withModel := range []bool{false, true} {
		h := newHarness(t)
		if withModel {
			h.askModel.turnOn("T0", "T2")
		}
		questions := []string{"What needs me?", "allocation by currency"}
		if !withModel {
			// With no model, a question that is no command gets the list of commands.
			questions = append(questions, "gibberish with no command")
		}
		for _, q := range questions {
			rec := h.asks(t, ownerSession, q)
			wantStatus(t, rec, http.StatusOK)
			wantNoBody(t, rec, "ask-note", noteDisclaimer, noteLinkText, "Answered by")
		}
	}
}

// stubAnswerer answers with a fixed card.
type stubAnswerer struct{ card answerCard }

func (s *stubAnswerer) answer(*http.Request, session, string) (answerCard, error) { return s.card, nil }

// TestNoteAnswererNotesOnlyAnswersAModelWrote is a table test of which cards
// get a note: a model's, never a command's and never a card that says a model
// did not answer, which the spend guards' refusals are.
func TestNoteAnswererNotesOnlyAnswersAModelWrote(t *testing.T) {
	t.Parallel()
	evaluated := map[string]bool{"a/b": true}
	tests := []struct {
		name string
		card answerCard
		want *modelNote
	}{
		{"a model's answer", answerCard{Model: "inst/a/b"}, &modelNote{Name: "a/b", Evaluated: true, EvalURL: "https://docs.example/eval/"}},
		{"a model not evaluated", answerCard{Model: "inst/c"}, &modelNote{Name: "c", EvalURL: "https://docs.example/eval/"}},
		{"a reference with no provider", answerCard{Model: "solo"}, &modelNote{Name: "solo", EvalURL: "https://docs.example/eval/"}},
		{"a command's answer", answerCard{Title: "Value"}, nil},
		{"a spend guard's refusal card", answerCard{Source: "The model gave no answer"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n := noteAnswerer{next: &stubAnswerer{card: tt.card}, evaluated: evaluated, evalURL: "https://docs.example/eval/"}
			got, err := n.answer(nil, session{}, "q")
			if err != nil {
				t.Fatalf("answer() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got.Note); diff != "" {
				t.Errorf("note (-want +got):\n%s", diff)
			}
		})
	}
}

// TestEvaluatedModelsCountsOnlyFullRunsWithAVerdict checks which results make
// a model evaluated, and that the committed ones name the models of the demo.
func TestEvaluatedModelsCountsOnlyFullRunsWithAVerdict(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"evalresults/a.json": {Data: []byte(`[
 {"model":"m/pass","verdict":"pass","answered":true},
 {"model":"m/fail","verdict":"fail","answered":true},
 {"model":"m/none","verdict":"none","answered":true},
 {"model":"m/diag","verdict":"pass","answered":true,"diagnostic":["R1"]},
 {"model":"m/silent","verdict":"fail","answered":false}]`)}}
	got, err := evaluatedModels(files)
	if err != nil {
		t.Fatalf("evaluatedModels() error = %v", err)
	}
	if diff := cmp.Diff(map[string]bool{"m/pass": true, "m/fail": true}, got); diff != "" {
		t.Errorf("evaluated (-want +got):\n%s", diff)
	}
	if _, badErr := evaluatedModels(fstest.MapFS{"evalresults/bad.json": {Data: []byte(`{`)}}); badErr == nil {
		t.Error("evaluatedModels() accepted a file that is no JSON")
	}

	committed, err := evaluatedModels(evalResultFiles)
	if err != nil {
		t.Fatalf("evaluatedModels(committed) error = %v", err)
	}
	for _, model := range []string{"google/gemma-4-26b-a4b-it", "openai/gpt-oss-120b"} {
		if !committed[model] {
			t.Errorf("the committed results do not cover %s", model)
		}
	}
	for _, model := range []string{"google/gemma-4-26b-a4b-it:free", "google/gemma-4-31b-it:free"} {
		if committed[model] {
			t.Errorf("%s is evaluated, though no call of it was answered", model)
		}
	}
}

// TestAnOpinionIsEvaluatedOnlyByTheOpinionsSet is a table test: a model that passed
// the eval's own cases is evaluated for an answer without opinions and not for one
// with them, until the opinions set has a verdict for it.
func TestAnOpinionIsEvaluatedOnlyByTheOpinionsSet(t *testing.T) {
	t.Parallel()
	main := map[string]bool{"a/b": true}
	tests := []struct {
		name     string
		opinions bool
		set      map[string]bool
		want     bool
	}{
		{"main-set pass, opinions off", false, nil, true},
		{"main-set pass, opinions on, no opinions results", true, nil, false},
		{"main-set pass, opinions on, results of another model", true, map[string]bool{"x/y": true}, false},
		{"main-set pass, opinions on, an opinions verdict", true, map[string]bool{"a/b": true}, true},
		{"an opinions verdict does not evaluate an answer without opinions", false, map[string]bool{"c": true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := "inst/a/b"
			if tt.name == "an opinions verdict does not evaluate an answer without opinions" {
				model = "inst/c"
			}
			n := noteAnswerer{next: &stubAnswerer{card: answerCard{Model: model, Opinions: tt.opinions}}, evaluated: main, evaluatedOpinions: tt.set, evalURL: "u"}
			got, err := n.answer(nil, session{}, "q")
			if err != nil || got.Note == nil || got.Note.Evaluated != tt.want {
				t.Errorf("note = %+v, %v; want Evaluated %v", got.Note, err, tt.want)
			}
		})
	}
}

func TestResultsOfTheOpinionsSetNeverCountForTheEvalsOwnCases(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"evalresults/a.json": {Data: []byte(`[
 {"model":"m/main","verdict":"pass","answered":true},
 {"model":"m/op","verdict":"pass","answered":true,"set":"opinions"},
 {"model":"m/opdiag","verdict":"pass","answered":true,"set":"opinions","diagnostic":["O1"]}]`)}}
	main, err := evaluatedModels(files)
	if err != nil || !main["m/main"] || main["m/op"] {
		t.Errorf("evaluatedModels = %v, %v; want the eval's own run only", main, err)
	}
	op, err := evaluatedOpinionModels(files)
	if err != nil || !op["m/op"] || op["m/main"] || op["m/opdiag"] {
		t.Errorf("evaluatedOpinionModels = %v, %v; want the opinions set's full run only", op, err)
	}
}
