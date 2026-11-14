package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRoutesPinAModelToHostsAndLabelEachRun(t *testing.T) {
	t.Parallel()
	const gemma = "google/gemma-4-26b-a4b-it"
	prices := gemma + "=0.05:0.3"
	tests := []struct {
		name   string
		flags  modelFlags
		labels []string
	}{
		{name: "no route: the candidates, unpinned", labels: []string{"qwen/qwen3.7-flash", "openai/gpt-oss-120b"}},
		{name: "a model on one host", flags: modelFlags{names: gemma, prices: prices, routes: gemma + "=primeintellect"}, labels: []string{gemma, gemma + "@primeintellect"}},
		{
			name: "a model on two routes, each a run, and the default routing beside them", flags: modelFlags{names: gemma, prices: prices, routes: gemma + "=primeintellect," + gemma + "=novita|deepinfra"},
			labels: []string{gemma, gemma + "@primeintellect", gemma + "@novita|deepinfra"},
		},
		{name: "a route names the model without -model", flags: modelFlags{prices: prices, routes: gemma + "=novita"}, labels: []string{gemma + "@novita"}},
		{
			name: "a model with a route beside one without", flags: modelFlags{names: "openai/gpt-oss-120b," + gemma, prices: prices, routes: gemma + "=novita"},
			labels: []string{"openai/gpt-oss-120b", gemma, gemma + "@novita"},
		},
		{name: "providers are trimmed and not repeated", flags: modelFlags{names: gemma, prices: prices, routes: gemma + "= novita |novita"}, labels: []string{gemma, gemma + "@novita"}},
		{
			name: "a route alone replaces the default run, a model listed too keeps it", flags: modelFlags{names: "openai/gpt-oss-120b", prices: prices, routes: gemma + "=novita"},
			labels: []string{"openai/gpt-oss-120b", gemma + "@novita"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			models, err := modelsFrom(&tt.flags)
			if err != nil {
				t.Fatal(err)
			}
			var labels []string
			for i := range models {
				labels = append(labels, modelLabel(&models[i]))
			}
			if diff := cmp.Diff(tt.labels, labels); diff != "" {
				t.Errorf("labels (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBadRoutesAreUsageErrors(t *testing.T) {
	t.Parallel()
	for _, routes := range []string{"no-equals", "=novita", "m=", "m=|", "m=a,m=a", "m=a|b,m=a|b"} {
		if _, err := modelsFrom(&modelFlags{names: "m", prices: "m=1:2", routes: routes}); err == nil || !strings.Contains(err.Error(), "-route") {
			t.Errorf("-route %q: error = %v, want a usage error naming -route", routes, err)
		}
	}
}

func TestTheFlagsOfARoutedModelAreFoundByIDOrByLabel(t *testing.T) {
	t.Parallel()
	const id = "x/y"
	models, err := modelsFrom(&modelFlags{names: id, prices: id + "=1:2", routes: id + "=a," + id + "=b", structured: []string{id + "@a"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 3 || models[0].Structured || !models[1].Structured || models[2].Structured {
		t.Errorf("models = %+v, want the flag on the run named by its label only", models)
	}
}

func TestARouteIsInTheRoutingAndTheConfigurationHash(t *testing.T) {
	t.Parallel()
	got := routingProbe("https://openrouter.ai/api/v1", &Model{Route: []string{"a", "b"}})
	want := map[string]any{
		"provider": map[string]any{"order": []string{"a", "b"}, "allow_fallbacks": false, "require_parameters": true},
		"usage":    map[string]any{"include": true},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("routing probe (-want +got):\n%s", diff)
	}
	unpinned := map[string]any{"provider": map[string]any{"require_parameters": true}, "usage": map[string]any{"include": true}}
	if diff := cmp.Diff(unpinned, routingProbe("https://openrouter.ai/api/v1", &Model{})); diff != "" {
		t.Errorf("unpinned routing probe (-want +got):\n%s", diff)
	}
	f := newFakeModel(t)
	hash := func(routes string) string {
		return runBench(t, f.url, "-case", "R1", "-route", routes).report.Models[1].ConfigHash
	}
	a, b, again := hash("test/model=a"), hash("test/model=b"), hash("test/model=a")
	if a == b || a != again {
		t.Errorf("hashes a %s, b %s, a again %s; want the route to change the hash and nothing else to", a, b, again)
	}
}

func TestTheReportKeysAModelByItsLabelAndHoldsItsRoute(t *testing.T) {
	t.Parallel()
	f := newFakeModel(t)
	got := runBench(t, f.url, "-case", "R1", "-route", "test/model=hostA,test/model=hostB|hostC")
	if got.report == nil || len(got.report.Models) != 3 {
		t.Fatalf("report = %+v (exit %d, %s)", got.report, got.code, got.stderr)
	}
	m0, m1 := got.report.Models[1], got.report.Models[2]
	if m0.Model != "test/model@hostA" || m1.Model != "test/model@hostB|hostC" || m0.ModelID != "test/model" || len(m1.Route) != 2 {
		t.Errorf("models = %q %q, ids %q, route %v", m0.Model, m1.Model, m0.ModelID, m1.Route)
	}
	if !strings.Contains(got.summary(t), "## test/model@hostA\n") || !strings.Contains(got.stdout, "test/model@hostB|hostC: ") {
		t.Error("the summary or the console does not name the runs by their labels")
	}
}
