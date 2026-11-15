package main

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestAnObjectArgumentMatchesKeyByKey(t *testing.T) {
	t.Parallel()
	expected := map[string]any{"period": map[string]any{
		"from": map[string]any{"$dateBetween": []any{"2025-10-31", "2025-11-01"}}, "to": "2025-11-30",
	}}
	tests := []struct {
		name   string
		actual map[string]any
		want   bool
	}{
		{"both keys right", map[string]any{"period": map[string]any{"from": "2025-11-01", "to": "2025-11-30"}}, true},
		{"the operator's other end", map[string]any{"period": map[string]any{"from": "2025-10-31", "to": "2025-11-30"}}, true},
		{"a key the model added", map[string]any{"period": map[string]any{"from": "2025-11-01", "to": "2025-11-30", "note": "x"}}, true},
		{"the operator fails", map[string]any{"period": map[string]any{"from": "2025-11-02", "to": "2025-11-30"}}, false},
		{"a value differs", map[string]any{"period": map[string]any{"from": "2025-11-01", "to": "2025-11-29"}}, false},
		{"a key is missing", map[string]any{"period": map[string]any{"from": "2025-11-01"}}, false},
		{"not an object", map[string]any{"period": "2025-11"}, false},
		{"no period", map[string]any{"date": "2025-11-30"}, false},
	}
	for _, tt := range tests {
		if got := matchArgs(expected, tt.actual); got != tt.want {
			t.Errorf("%s: matchArgs = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestADefaultAnotherArgumentReplacesIsLeftOut checks the canonical form of a
// call (ui.AskArgReplacedBy): a value asked over a period has no date, a
// performance asked for a month no named period, and a call that gives none of
// the replacing arguments has the default.
func TestADefaultAnotherArgumentReplacesIsLeftOut(t *testing.T) {
	t.Parallel()
	latest, err := kernel.ParseDate("2026-10-28")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, tool, args, want string
	}{
		{"a value on no day", "value", `{}`, `{"date":"2026-10-28"}`},
		{"a value over a period", "value", `{"period":{"from":"2025-11-01","to":"2025-11-30"}}`, `{"period":{"from":"2025-11-01","to":"2025-11-30"}}`},
		{"a value over a period with an empty date", "value", `{"date":"","period":{"from":"2025-11-01","to":"2025-11-30"}}`, `{"date":"","period":{"from":"2025-11-01","to":"2025-11-30"}}`},
		{"a value with a null period", "value", `{"period":null}`, `{"date":"2026-10-28","period":null}`},
		{"a performance on no period", "performance", `{}`, `{"period":"this_year"}`},
		{"a performance of a month", "performance", `{"month":"2025-11"}`, `{"month":"2025-11"}`},
		{"a performance between days", "performance", `{"from":"2025-11-01","to":"2025-11-30"}`, `{"from":"2025-11-01","to":"2025-11-30"}`},
		{"a performance since a day", "performance", `{"since":"2025-11-01"}`, `{"since":"2025-11-01"}`},
		{"a performance of a year", "performance", `{"year":2025}`, `{"year":2025}`},
		{"a performance of an empty month", "performance", `{"month":""}`, `{"month":"","period":"this_year"}`},
	}
	for _, tt := range tests {
		got := canonicalArgs(tt.tool, tt.args, latest)
		var g, w any
		if err := json.Unmarshal([]byte(got), &g); err != nil {
			t.Fatalf("%s: canonical arguments %q are no JSON: %v", tt.name, got, err)
		}
		if err := json.Unmarshal([]byte(tt.want), &w); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(w, g); diff != "" {
			t.Errorf("%s: canonicalArgs (-want +got):\n%s", tt.name, diff)
		}
	}
}
