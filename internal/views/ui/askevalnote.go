package ui

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

// defaultAskEvalURL is the docs page that reports how models did in the Ask
// box's eval (D-222), on the public docs site, where the site publishes the
// docs of main. Config.AskEvalURL overrides it, for a deployment that serves
// its own docs.
const defaultAskEvalURL = "https://openfief.dev/docs/reference/ask-eval/"

// evalResultFiles are the trimmed results of the Ask box bench's runs, the
// files docgen renders that page from, so the box's "not evaluated" note and
// the page cannot disagree.
//
//go:embed evalresults/*.json
var evalResultFiles embed.FS

// modelNote is what a model's answer says under it (D-222): which model
// answered, that it can be wrong, and where to read how it did in the eval.
type modelNote struct {
	// Name is the model's id at its provider, such as "openai/gpt-oss-120b":
	// the entry of the model settings that answered, without the provider
	// instance.
	Name string
	// Evaluated is false for a model no full run of the eval has covered.
	Evaluated bool
	// EvalURL is the page that reports the eval's results.
	EvalURL string
	// Opinions is true for an answer given with opinions on (D-225): the note
	// then says what an opinion is.
	Opinions bool
}

// Advice is the sentence the note says about advice: every model answer says it
// is not investment advice, and with opinions on, that it is an AI's opinion
// from the cards and public prices, and that it never acts (D-225).
func (n *modelNote) Advice() string {
	if n.Opinions {
		return opinionsNotAdviceNote
	}
	return notAdviceNote
}

// evalResultRow is the part of a result file the note reads.
type evalResultRow struct {
	Model      string   `json:"model"`
	Verdict    string   `json:"verdict"`
	Answered   bool     `json:"answered"`
	Diagnostic []string `json:"diagnostic"`
	// Set names the set of cases the run asked when it is not the eval's own:
	// "opinions" for the opinions set (D-225, D-229); empty for the eval's.
	Set string `json:"set"`
}

// opinionsResultSet is the Set of a result of the opinions set.
const opinionsResultSet = "opinions"

// evaluatedModels returns the model ids a full run of the eval has covered
// with a verdict: a run of some cases only, one no call answered and one from
// before the verdict existed do not count, and neither does a run of the
// opinions set, which says nothing of the model's answers to the eval's cases.
func evaluatedModels(files fs.FS) (map[string]bool, error) { return evaluatedIn(files, "") }

// evaluatedOpinionModels returns the model ids a full run of the opinions set
// has covered with a verdict, on the same terms. An opinion's note reads only
// these: a pass of the eval's own cases says nothing of an opinion.
func evaluatedOpinionModels(files fs.FS) (map[string]bool, error) {
	return evaluatedIn(files, opinionsResultSet)
}

// evaluatedIn is the models the full runs with a verdict of the named set covered.
func evaluatedIn(files fs.FS, set string) (map[string]bool, error) {
	rows, err := resultRows(files)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, r := range rows {
		if r.Set == set && r.Answered && len(r.Diagnostic) == 0 && r.Verdict != "none" && r.Verdict != "" {
			out[r.Model] = true
		}
	}
	return out, nil
}

// resultRows reads every run of every committed result file.
func resultRows(files fs.FS) ([]evalResultRow, error) {
	names, err := fs.Glob(files, "evalresults/*.json")
	if err != nil {
		return nil, fmt.Errorf("list eval results: %w", err)
	}
	var out []evalResultRow
	for _, name := range names {
		data, readErr := fs.ReadFile(files, name)
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", name, readErr)
		}
		var rows []evalResultRow
		if jsonErr := json.Unmarshal(data, &rows); jsonErr != nil {
			return nil, fmt.Errorf("decode %s: %w", name, jsonErr)
		}
		out = append(out, rows...)
	}
	return out, nil
}

// noteAnswerer adds the model note to every answer a model wrote: a card
// whose Model is set. The commands' own answers, and a card that says a model
// did not answer (the spend guards' refusals), carry none.
type noteAnswerer struct {
	next      answerer
	evaluated map[string]bool
	// evaluatedOpinions are the models the opinions set has a verdict for: an
	// answer given with opinions on is evaluated only by them (D-225).
	evaluatedOpinions map[string]bool
	evalURL           string
}

func (n noteAnswerer) answer(r *http.Request, sess session, question string) (answerCard, error) {
	card, err := n.next.answer(r, sess, question)
	if err != nil {
		return card, err
	}
	if card.Model != "" {
		name := modelName(card.Model)
		evaluated := n.evaluated[name]
		if card.Opinions {
			evaluated = n.evaluatedOpinions[name]
		}
		card.Note = &modelNote{Name: name, Evaluated: evaluated, EvalURL: n.evalURL, Opinions: card.Opinions}
	}
	return card, nil
}

// modelName is a model's name as the note shows it: the "provider/model"
// reference of the settings entry without the provider instance.
func modelName(ref string) string {
	if _, model, ok := strings.Cut(ref, "/"); ok && model != "" {
		return model
	}
	return ref
}
