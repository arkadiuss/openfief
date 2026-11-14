package main

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/datar-psa/goeval/api"
	"github.com/datar-psa/goeval/heuristic"

	"github.com/arkadiuss/openfief/internal/views/ui"
)

// toAPICalls maps the loop's trace to goeval's tool calls, keeping each call's
// arguments as the model wrote them.
func toAPICalls(calls []ui.AskRunCall) []api.ToolCall {
	out := make([]api.ToolCall, 0, len(calls))
	for _, c := range calls {
		out = append(out, api.ToolCall{Name: c.Name, Arguments: json.RawMessage(c.Args)})
	}
	return out
}

// toolCallScorer returns goeval's tool-call scorer set up for the run's case:
// the calls it must make, the tools it must not call, the cap on calls, and
// no calls at all where the case says so. It is nil when the case expects
// nothing of the calls.
func toolCallScorer(d *RunData) *scorer {
	e := &d.Case.Expect
	if !e.NoCalls && e.MaxCalls == 0 && len(e.Forbid) == 0 && len(e.Calls) == 0 {
		return nil
	}
	inner := heuristic.ToolCalls(heuristic.ToolCallsOptions{
		NoToolCalls: e.NoCalls, MaxCalls: e.MaxCalls, Forbidden: e.Forbid, ArgsMatcher: matchArgs,
	})
	expected := expectedOf(e.Calls, d.Run.Calls)
	return &scorer{name: "ToolCalls", fn: func(ctx context.Context, in *api.ScoreInputs) api.Score {
		in.ExpectedToolCalls = expected
		out := inner.Score(ctx, *in)
		out.Name = "ToolCalls"
		return out
	}}
}

// expectedOf turns a case's expected calls into goeval's. A call with
// alternatives becomes the first alternative the model made, or the first when
// it made none, so the scorer asks for the one the model went for.
func expectedOf(calls []ExpectedCall, made []ui.AskRunCall) []api.ToolCall {
	out := make([]api.ToolCall, 0, len(calls))
	for i := range calls {
		c := &calls[i]
		if len(c.AnyOf) > 0 {
			c = pickAlternative(c.AnyOf, made)
		}
		call := api.ToolCall{Name: c.Tool}
		if len(c.Args) > 0 {
			if raw, err := json.Marshal(c.Args); err == nil {
				call.Arguments = raw
			}
		}
		out = append(out, call)
	}
	return out
}

// pickAlternative returns the first alternative the model made a call for, one
// whose arguments match; else the first whose tool it called; else the first.
func pickAlternative(alternatives []ExpectedCall, made []ui.AskRunCall) *ExpectedCall {
	for i := range alternatives {
		for _, m := range made {
			if m.Name == alternatives[i].Tool && argsMatchRaw(&alternatives[i], m.Args) {
				return &alternatives[i]
			}
		}
	}
	for i := range alternatives {
		for _, m := range made {
			if m.Name == alternatives[i].Tool {
				return &alternatives[i]
			}
		}
	}
	return &alternatives[0]
}

// argsMatchRaw reports whether the call's raw arguments satisfy the expected
// call's.
func argsMatchRaw(e *ExpectedCall, raw string) bool {
	var actual map[string]any
	if json.Unmarshal([]byte(raw), &actual) != nil {
		return false
	}
	want, err := json.Marshal(e.Args)
	if err != nil {
		return false
	}
	var expected map[string]any
	if json.Unmarshal(want, &expected) != nil {
		return false
	}
	return matchArgs(expected, actual)
}

// matchArgs says whether the actual arguments of a call satisfy the expected
// ones: every key expected is there with an equal value, or satisfies its
// operator (see testdata/cases.yaml); keys the model added are allowed.
func matchArgs(expected, actual map[string]any) bool {
	for key, spec := range expected {
		got, has := actual[key]
		if !matchKey(spec, got, has) {
			return false
		}
	}
	return true
}

// matchKey says whether one argument, given or not, satisfies what is expected
// of it: a value, or an operator object.
func matchKey(spec, got any, has bool) bool {
	ops, isOp := spec.(map[string]any)
	if !isOp || !hasOperator(ops) {
		return has && reflect.DeepEqual(spec, got)
	}
	if !has {
		optional, ok := ops["$optional"].(bool)
		return ok && optional
	}
	return matchOperator(ops, got)
}

// hasOperator reports whether m is an operator object: it has a key that
// starts with a dollar sign.
func hasOperator(m map[string]any) bool {
	for k := range m {
		if strings.HasPrefix(k, "$") {
			return true
		}
	}
	return false
}

// operators are the operators an expected argument can be, each a test of the
// argument the model gave against the operator's own argument.
func operators() map[string]func(arg, got any) bool {
	return map[string]func(arg, got any) bool{
		"$optional":        func(any, any) bool { return true },
		"$oneOf":           func(arg, got any) bool { return oneOf(arg, got, false) },
		"$oneOfFold":       func(arg, got any) bool { return oneOf(arg, got, true) },
		"$containsFold":    func(arg, got any) bool { return containsAny([]any{arg}, got) },
		"$containsFoldAny": containsAny,
		"$dateBetween":     dateBetween,
	}
}

// matchOperator evaluates the operators of ops on the value got.
func matchOperator(ops map[string]any, got any) bool {
	table := operators()
	for op, arg := range ops {
		test, known := table[op]
		if !known || !test(arg, got) {
			return false
		}
	}
	return true
}

// oneOf reports whether got equals one of the list arg, case aside for text
// when fold is set.
func oneOf(arg, got any, fold bool) bool {
	list, listOK := arg.([]any)
	if !listOK {
		return false
	}
	text, isText := got.(string)
	for _, v := range list {
		if reflect.DeepEqual(v, got) {
			return true
		}
		if s, ok := v.(string); fold && ok && isText && strings.EqualFold(s, text) {
			return true
		}
	}
	return false
}

// containsAny reports whether got is text that holds, case aside, one of the
// strings of the list arg.
func containsAny(arg, got any) bool {
	list, listOK := arg.([]any)
	text, textOK := got.(string)
	if !listOK || !textOK {
		return false
	}
	for _, v := range list {
		if s, isText := v.(string); isText && strings.Contains(strings.ToLower(text), strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// dateBetween reports whether got is an ISO date between the two dates of the
// list arg, both included. ISO dates sort as text.
func dateBetween(arg, got any) bool {
	list, listOK := arg.([]any)
	text, textOK := got.(string)
	if !listOK || len(list) != 2 || !textOK || !isoDatePattern().MatchString(text) {
		return false
	}
	from, fromOK := list[0].(string)
	to, toOK := list[1].(string)
	return fromOK && toOK && text >= from && text <= to
}
