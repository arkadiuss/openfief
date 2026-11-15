package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

func TestTheAnswerIsAskedStrictlyOfAModelWithTheFlagAndAsJSONModeOfOneWithout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "with the flag", args: []string{"-structured", "test/model"}, want: "json_schema"},
		{name: "without it", want: "json_object"},
		{name: "taken away from a default", args: []string{"-structured", "test/model", "-json-mode", "test/model"}, want: "json_object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeModel(t)
			got := runBench(t, f.url, append([]string{"-case", "A17", "-parallel", "1"}, tt.args...)...)
			if got.report == nil {
				t.Fatalf("no report: exit %d, %s", got.code, got.stderr)
			}
			seen := f.formatsAndTools(t)
			// "Say hello" never stops asking for tools, so each of its five runs reaches the final round.
			if want := []string{tt.want, tt.want, tt.want, tt.want, tt.want}; !slices.Equal(seen.formats, want) || seen.withTools != 25 {
				t.Errorf("response formats %v and %d requests that offered tools, want %v and 25", seen.formats, seen.withTools, want)
			}
			if got.report.Models[0].Config.StructuredOutput != (tt.want == "json_schema") {
				t.Errorf("the configuration says structured output = %v", got.report.Models[0].Config.StructuredOutput)
			}
		})
	}
}

// formatsAndTools returns the type of the response format of each request that
// asked for one, and how many offered tools; a request that did both is an error.
func (f *fakeModel) formatsAndTools(t *testing.T) requestFormats {
	t.Helper()
	var seen requestFormats
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, body := range f.requests {
		format, hasFormat := body["response_format"].(map[string]any)
		offered := body["tools"] != nil
		if offered && hasFormat {
			t.Errorf("a request that offered tools asked for a response format: %v", format)
		}
		if offered {
			seen.withTools++
		}
		if hasFormat {
			seen.formats = append(seen.formats, fmt.Sprint(format["type"]))
		}
	}
	return seen
}

// requestFormats is what the fake model's requests asked for: the type of each
// response format, and how many requests offered tools.
type requestFormats struct {
	formats   []string
	withTools int
}

// TestTheStrictFormatCarriesTheSchema checks the strict form is the box's schema, strict.
func TestTheStrictFormatCarriesTheSchema(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	if got := runBench(t, f.url, "-case", "A17", "-parallel", "1", "-structured", "test/model"); got.report == nil {
		t.Fatalf("no report: exit %d, %s", got.code, got.stderr)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, body := range f.requests {
		format, ok := body["response_format"].(map[string]any)
		if !ok {
			continue
		}
		schema, isMap := format["json_schema"].(map[string]any)
		if !isMap {
			t.Fatalf("response_format = %v, want a schema", format)
		}
		var want any
		if err := json.Unmarshal([]byte(ui.AskAnswerSchema()), &want); err != nil {
			t.Fatal(err)
		}
		if schema["strict"] != true || !reflect.DeepEqual(schema["schema"], want) {
			t.Fatalf("response_format = %v, want the answer's schema, strict", format)
		}
	}
}

func TestTheStructuredFlagIsInTheConfigurationHash(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	strict := runBench(t, f.url, "-case", "R1", "-structured", "test/model").report.Models[0].ConfigHash
	jsonMode := runBench(t, f.url, "-case", "R1").report.Models[0].ConfigHash
	again := runBench(t, f.url, "-case", "R1", "-structured", "test/model").report.Models[0].ConfigHash
	if strict == jsonMode || strict != again {
		t.Errorf("hashes: strict %s, JSON mode %s, strict again %s; want the flag to change the hash and nothing else to", strict, jsonMode, again)
	}
}

func TestTheDefaultModelsKeepTheirFlags(t *testing.T) {
	t.Parallel()
	models, err := modelsFrom(&modelFlags{})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, m := range models {
		got[m.ID] = m.Structured
	}
	if !got["openai/gpt-oss-120b"] || got["qwen/qwen3.7-flash"] {
		t.Errorf("structured by default = %v, want gpt-oss with the flag and qwen without it (JSON mode)", got)
	}
	gemma, err := modelsFrom(&modelFlags{names: "google/gemma-4-26b-a4b-it", prices: "google/gemma-4-26b-a4b-it=0.05:0.3"})
	if err != nil || !gemma[0].Structured {
		t.Errorf("gemma = %+v, %v; want the flag by default", gemma, err)
	}
	other, err := modelsFrom(&modelFlags{names: "x/y", prices: "x/y=1:2", structured: []string{"x/y"}})
	if err != nil || !other[0].Structured {
		t.Errorf("-structured: %+v, %v", other, err)
	}
}

func TestAReplyThatIsNoStructuredAnswerIsASchemaFailure(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	c := runBench(t, f.url, "-case", "A13").caseOf(t, "A13")
	if c.Passed != 0 || len(c.Detail) != 5 {
		t.Fatalf("A13 = %+v, want 0 of 5", c)
	}
	run := c.Detail[0]
	if run.Answered || run.SchemaError == "" || run.Invalid != 2 || run.ModelCalls != 3 {
		t.Errorf("run = answered %v, schema error %q, %d invalid, %d calls; want a schema failure after the final round and the retry", run.Answered, run.SchemaError, run.Invalid, run.ModelCalls)
	}
	reasons := make([]string, 0, len(c.Failures))
	for _, fl := range c.Failures {
		reasons = append(reasons, fl.Reason)
	}
	if !slices.ContainsFunc(reasons, func(r string) bool { return strings.HasPrefix(r, "Schema: schema failure: it is not one JSON object") }) {
		t.Errorf("failures = %q, want the schema failure named", reasons)
	}
}

// TestTheKindDecidesTheRefusalsAndTheAmbiguousName checks the cases the schema
// decides: R1 and R2 by the kind with text, whatever the words (R2's Polish
// refusal is in none of the phrase list's), R5 by the kind and the box's
// commands it offers, A18 by a clarifying answer that lists its candidates and
// guessed nothing.
func TestTheKindDecidesTheRefusalsAndTheAmbiguousName(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	got := runBench(t, f.url, "-case", "R1,R2,R5,A18")
	for _, id := range []string{"R1", "R2", "R5", "A18"} {
		if c := got.caseOf(t, id); !c.PassAll {
			t.Errorf("%s: pass^5 = false (%+v)", id, c.Failures)
		}
	}
	if got.code != exitOK {
		t.Errorf("exit %d, want 0", got.code)
	}
	r5 := got.caseOf(t, "R5").Detail[0]
	if r5.Kind != "refuse" || len(r5.Suggestions) != len(ui.AskRefusalSuggestions(kernel.Date{})) {
		t.Errorf("R5's run = kind %q, suggestions %v", r5.Kind, r5.Suggestions)
	}
	if a18 := got.caseOf(t, "A18").Detail[0]; a18.Kind != "clarify" || len(a18.Candidates) != 2 {
		t.Errorf("A18's run = kind %q, candidates %v", a18.Kind, a18.Candidates)
	}
}
