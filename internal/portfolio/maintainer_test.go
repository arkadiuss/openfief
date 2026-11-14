package portfolio_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// maintainerFixture is the model of driftModel with the drift of a portfolio
// holding 63.10% and 20% of its value in the two instruments.
func maintainerFixture(t *testing.T) (*portfolio.Model, *portfolio.Drift) {
	t.Helper()
	m := driftModel(t)
	d, err := portfolio.DriftOf(m, map[string]kernel.Decimal{etfA: mustDec(t, "6310"), etfB: mustDec(t, "2000")}, mustDec(t, "10000"))
	if err != nil {
		t.Fatalf("DriftOf: %v", err)
	}
	return m, &d
}

// answer writes the JSON a model would answer with.
func answer(t *testing.T, note string, changes ...map[string]string) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{"note": note, "changes": changes})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// change is one change of an answer.
func change(target, current, proposed, reason string) map[string]string {
	return map[string]string{
		"target": target, "currentTargetPercent": current, "proposedTargetPercent": proposed, "reason": reason,
	}
}

// validPair moves two points from ETFA to ETFB: 58 + 32 + 10 = 100.
func validPair() []map[string]string {
	return []map[string]string{
		change("instrument:"+etfA, "60", "58", "Trim the overweight."), change("instrument:"+etfB, "30", "32", "Add where it is light."),
	}
}

// answerCase is one answer and what ParseAnswer makes of it.
type answerCase struct {
	name        string
	content     string
	wantInvalid bool
	wantDropped bool
	wantChanges int
}

const answerNote = "ETFA drifted above its target."

// changeCount is how many changes p holds; none for no proposal.
func changeCount(p *portfolio.Proposal) int {
	if p == nil {
		return 0
	}
	return len(p.Changes)
}

// checkAnswer fails the test unless got and err are what c wants.
func checkAnswer(t *testing.T, c *answerCase, got portfolio.Answer, err error) {
	t.Helper()
	if c.wantInvalid {
		if !errors.Is(err, portfolio.ErrInvalidAnswer) {
			t.Fatalf("ParseAnswer error = %v, want ErrInvalidAnswer", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("ParseAnswer: %v", err)
	}
	if got.Note != answerNote || got.Dropped != c.wantDropped {
		t.Errorf("Note = %q, Dropped = %v; want the note kept and Dropped %v", got.Note, got.Dropped, c.wantDropped)
	}
	if n := changeCount(got.Proposal); n != c.wantChanges {
		t.Errorf("Proposal has %d changes, want %d: never a partial proposal", n, c.wantChanges)
	}
}

func TestParseAnswerAcceptsAndRefuses(t *testing.T) {
	t.Parallel()
	tests := []answerCase{
		{name: "a note and no changes", content: answer(t, answerNote)},
		{name: "a note and a valid change set", content: answer(t, answerNote, validPair()...), wantChanges: 2},
		{name: "not JSON", content: "ETFA drifted", wantInvalid: true},
		{name: "JSON of another shape", content: `{"summary":"x","changes":[]}`, wantInvalid: true},
		{name: "text after the JSON", content: answer(t, answerNote) + " thanks", wantInvalid: true},
		{name: "no note", content: answer(t, "  "), wantInvalid: true},
		{name: "a note that is too long", content: answer(t, strings.Repeat("a", portfolio.MaxNoteRunes+1)), wantInvalid: true},
		{
			name:        "a weight that is a number, not a decimal string",
			content:     `{"note":"` + answerNote + `","changes":[{"target":"instrument:` + etfA + `","currentTargetPercent":60,"proposedTargetPercent":58,"reason":"x"}]}`,
			wantInvalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, d := maintainerFixture(t)
			got, err := portfolio.ParseAnswer(tt.content, m, d)
			checkAnswer(t, &tt, got, err)
		})
	}
}

func TestParseAnswerDropsAProposalThatBreaksARule(t *testing.T) {
	t.Parallel()
	a, b := "instrument:"+etfA, "instrument:"+etfB
	tests := []struct {
		name    string
		changes []map[string]string
	}{
		{"weights that do not sum to 100", []map[string]string{change(a, "60", "58", "Trim.")}},
		{"a target that is not in the model", []map[string]string{change("instrument:XX0000000CC3/ETFC", "10", "20", "New.")}},
		{"a current weight that is not the model's", []map[string]string{change(a, "55", "58", "Trim."), change(b, "30", "32", "Add.")}},
		{"a weight that is not a decimal", []map[string]string{change(a, "60", "fifty", "Trim.")}},
		{"more than two decimal places", []map[string]string{change(a, "60", "58.005", "Trim."), change(b, "30", "31.995", "Add.")}},
		{"a weight of zero, which removes the target", []map[string]string{change(b, "30", "0", "Drop it."), change(a, "60", "90", "Take it.")}},
		{"a change that changes nothing", []map[string]string{change(a, "60", "60", "Keep.")}},
		{"a target changed twice", []map[string]string{change(a, "60", "58", "Trim."), change(a, "60", "62", "Add.")}},
		{"a reason of two lines", []map[string]string{change(a, "60", "58", "Trim.\nAnd more."), change(b, "30", "32", "Add.")}},
		{"no reason", []map[string]string{change(a, "60", "58", " "), change(b, "30", "32", "Add.")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, d := maintainerFixture(t)
			c := answerCase{wantDropped: true}
			got, err := portfolio.ParseAnswer(answer(t, answerNote, tt.changes...), m, d)
			checkAnswer(t, &c, got, err)
		})
	}
}

// figureRow is what a change shows of its figures.
type figureRow struct {
	Target                        string
	Before, After, Current, Drift string
	Held                          bool
	Reason                        string
}

func TestParseAnswerFillsTheFiguresBehindEachChange(t *testing.T) {
	t.Parallel()
	m, d := maintainerFixture(t)
	got, err := portfolio.ParseAnswer(answer(t, answerNote, change("instrument:"+etfA, "60", "58", "Trim."), change("instrument:"+etfB, "30", "32", "Add.")), m, d)
	if err != nil || got.Proposal == nil {
		t.Fatalf("ParseAnswer = %+v, %v", got, err)
	}
	have := make([]figureRow, 0, len(got.Proposal.Changes))
	for _, c := range got.Proposal.Changes {
		have = append(have, figureRow{c.Target, c.Before.String(), c.After.String(), c.Current.String(), c.Drift.String(), c.Held, c.Reason})
	}
	want := []figureRow{
		{"instrument:" + etfA, "60", "58", "63.1", "3.1", true, "Trim."},
		{"instrument:" + etfB, "30", "32", "20", "-10", true, "Add."},
	}
	if diff := cmp.Diff(want, have); diff != "" {
		t.Errorf("figures mismatch (-want +have):\n%s", diff)
	}
}

// proposed is a valid proposal over the fixture's model.
func proposed(t *testing.T) (*portfolio.Model, *portfolio.Proposal) {
	t.Helper()
	m, d := maintainerFixture(t)
	got, err := portfolio.ParseAnswer(answer(t, answerNote, validPair()...), m, d)
	if err != nil || got.Proposal == nil {
		t.Fatalf("ParseAnswer = %+v, %v", got, err)
	}
	return m, got.Proposal
}

// weightsOf lists the weights of m's targets.
func weightsOf(m *portfolio.Model) []string {
	out := make([]string, 0, len(m.Targets))
	for i := range m.Targets {
		out = append(out, m.Targets[i].Weight.String())
	}
	return out
}

func TestProposalApplyChangesOnlyTheNamedWeights(t *testing.T) {
	t.Parallel()
	m, p := proposed(t)
	applied, err := p.Apply(m)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if diff := cmp.Diff([]string{"58", "32", "10"}, weightsOf(&applied)); diff != "" {
		t.Errorf("weights mismatch (-want +got):\n%s", diff)
	}
	if !applied.Enabled || applied.Validate() != nil {
		t.Errorf("the applied model is on and valid, got enabled=%v, validate=%v", applied.Enabled, applied.Validate())
	}
	if diff := cmp.Diff([]string{"60", "30", "10"}, weightsOf(m)); diff != "" {
		t.Errorf("Apply changed the model it was given (-want +got):\n%s", diff)
	}
}

func TestProposalApplyKeepsAModelThatIsOffOff(t *testing.T) {
	t.Parallel()
	m, p := proposed(t)
	off := *m
	off.Enabled = false
	applied, err := p.Apply(&off)
	if err != nil || applied.Enabled {
		t.Errorf("Apply = enabled %v, %v; want off and no error", applied.Enabled, err)
	}
}

func TestProposalApplyRefusesAModelThatChangedSince(t *testing.T) {
	t.Parallel()
	_, p := proposed(t)
	pln := mustCurrency(t, "PLN")
	tests := []struct {
		name  string
		model portfolio.Model
	}{
		{"weights the owner edited", portfolio.Model{Enabled: true, Targets: []portfolio.Target{
			{Kind: portfolio.TargetInstrument, Name: etfA, Weight: mustDec(t, "50")},
			{Kind: portfolio.TargetInstrument, Name: etfB, Weight: mustDec(t, "40")},
			{Kind: portfolio.TargetIndex, Name: index1, Currency: pln, Weight: mustDec(t, "10")},
		}}},
		{"a target removed", portfolio.Model{Enabled: true, Targets: []portfolio.Target{
			{Kind: portfolio.TargetInstrument, Name: etfA, Weight: mustDec(t, "100")},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := p.Apply(&tt.model); !errors.Is(err, portfolio.ErrStaleProposal) {
				t.Errorf("Apply error = %v, want ErrStaleProposal", err)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	t.Parallel()
	_, d := maintainerFixture(t)
	asOf, err := kernel.ParseDate("2026-11-13")
	if err != nil {
		t.Fatal(err)
	}
	got := portfolio.BuildPrompt(d, asOf, mustCurrency(t, "PLN"))
	wantFigures := "Portfolio as of 2026-11-13. Weights are percent of the whole portfolio's value in PLN.\n" +
		"Targets, one per line: key | target weight | weight held now | drift in percentage points.\n" +
		"instrument:" + etfA + " | 60 | 63.1 | +3.1\n" +
		"instrument:" + etfB + " | 30 | 20 | -10\n" +
		"index:" + index1 + " | 10 | not held | -\n" +
		"Held outside the targets (cash and other holdings): 16.9\n"
	if diff := cmp.Diff(wantFigures, got.Figures); diff != "" {
		t.Errorf("figures mismatch (-want +got):\n%s", diff)
	}
	if strings.Contains(got.Instructions, "ETFA") || strings.Contains(got.Instructions, "63.1") {
		t.Error("the instructions, which go as public data (T2), quote a figure of the portfolio")
	}
	for _, must := range []string{"cannot trade", "sum to exactly 100", "JSON"} {
		if !strings.Contains(got.Instructions, must) {
			t.Errorf("the instructions don't say %q", must)
		}
	}
}

func TestResponseSchemaIsJSONAndStrict(t *testing.T) {
	t.Parallel()
	var schema map[string]any
	if err := json.Unmarshal([]byte(portfolio.ResponseSchema()), &schema); err != nil {
		t.Fatalf("the schema is not JSON: %v", err)
	}
	if schema["additionalProperties"] != false {
		t.Error("the schema allows properties it doesn't list; strict structured output needs false")
	}
	if diff := cmp.Diff([]any{"note", "changes"}, schema["required"]); diff != "" {
		t.Errorf("required mismatch (-want +got):\n%s", diff)
	}
}
