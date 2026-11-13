package ui

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestAskToolsAreTheReadOnlyAnswerers checks that the tools the model may call
// are exactly the commands' answerers and nothing more (D-203): no tool
// changes anything, and every tool has a name, a description, a schema that is
// a JSON object schema, a card label and a data tier of D-206.
func TestAskToolsAreTheReadOnlyAnswerers(t *testing.T) {
	t.Parallel()
	tools := askTools()
	names := make([]string, 0, len(tools))
	for i := range tools {
		names = append(names, tools[i].Name)
		checkTool(t, &tools[i])
	}
	want := []string{"what_needs_me", "show", "sync_status", "allocation", "performance", "value"}
	if diff := cmp.Diff(want, names); diff != "" {
		t.Errorf("tools (-want +got):\n%s", diff)
	}
	for _, n := range names {
		for _, writes := range []string{"set", "create", "delete", "update", "add", "remove", "save", "confirm", "retry", "grant", "revoke"} {
			if strings.HasPrefix(n, writes) {
				t.Errorf("tool %s sounds like a write", n)
			}
		}
	}
}

// checkTool checks one tool is complete, of a tier of D-206, with a closed
// object schema.
func checkTool(t *testing.T, tool *askTool) {
	t.Helper()
	if tool.Description == "" || tool.Label == "" || tool.query == nil || tierWords(tool.Tier) == "" {
		t.Errorf("tool %s is missing its description, label or mapping, or has a tier that is none of D-206's", tool.Name)
	}
	var schema struct {
		Type                 string         `json:"type"`
		Properties           map[string]any `json:"properties"`
		AdditionalProperties *bool          `json:"additionalProperties"`
	}
	if err := json.Unmarshal([]byte(tool.Schema), &schema); err != nil {
		t.Fatalf("tool %s: schema is not JSON: %v", tool.Name, err)
	}
	if schema.Type != "object" || schema.Properties == nil || schema.AdditionalProperties == nil || *schema.AdditionalProperties {
		t.Errorf("tool %s: schema %s is not a closed object schema", tool.Name, tool.Schema)
	}
}

// TestToolsRefuseMalformedArguments checks that every tool's mapping refuses
// an argument its schema doesn't list, arguments that are no JSON object and
// trailing data, so a model can't pass anything the schema doesn't promise.
func TestToolsRefuseMalformedArguments(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	for _, tool := range askTools() {
		t.Run(tool.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := tool.query(`{"surprise":1}`, today); !errors.Is(err, errBadArgs) {
				t.Errorf("an unknown argument: error = %v, want errBadArgs", err)
			}
			if _, err := tool.query(`[]`, today); !errors.Is(err, errBadArgs) {
				t.Errorf("arguments that are no object: error = %v, want errBadArgs", err)
			}
			if _, err := tool.query(`{} {}`, today); !errors.Is(err, errBadArgs) {
				t.Errorf("two objects: error = %v, want errBadArgs", err)
			}
		})
	}
}

func TestAskToolArguments(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	tests := []struct {
		tool    string
		args    string
		want    askQuery
		wantErr error
	}{
		{tool: "what_needs_me", args: `{}`, want: askQuery{Command: askNeeds}},
		{tool: "what_needs_me", args: ``, want: askQuery{Command: askNeeds}},
		{tool: "show", args: `{"name":"Synthetic brokerage"}`, want: askQuery{Command: askShow, Term: "Synthetic brokerage"}},
		{tool: "show", args: `{"name":" bonds ","kind":"sleeve"}`, want: askQuery{Command: askShow, Term: "bonds", Show: showSleeve}},
		{tool: "show", args: `{"name":"PLTEST000010","kind":"instrument"}`, want: askQuery{Command: askShow, Term: "PLTEST000010", Show: showInstrument}},
		{tool: "show", args: `{"name":"x","kind":"account"}`, want: askQuery{Command: askShow, Term: "x", Show: showAccount}},
		{tool: "show", args: `{"kind":"account"}`, wantErr: errNoName},
		{tool: "show", args: `{"name":"  "}`, wantErr: errNoName},
		{tool: "show", args: `{"name":"x","kind":"portfolio"}`, wantErr: errBadEnum},
		{tool: "sync_status", args: `{"connector":"xtb-main"}`, want: askQuery{Command: askSync, Term: "xtb-main"}},
		{tool: "sync_status", args: `{}`, wantErr: errNoName},
		{tool: "allocation", args: `{}`, want: askQuery{Command: askAllocation, Dimension: dimAssetClass}},
		{tool: "allocation", args: `{"by":"asset_class"}`, want: askQuery{Command: askAllocation, Dimension: dimAssetClass}},
		{tool: "allocation", args: `{"by":"currency"}`, want: askQuery{Command: askAllocation, Dimension: dimCurrency}},
		{tool: "allocation", args: `{"by":"institution"}`, want: askQuery{Command: askAllocation, Dimension: dimInstitution}},
		{tool: "allocation", args: `{"by":"sleeve"}`, want: askQuery{Command: askAllocation, Dimension: dimSleeve}},
		{tool: "allocation", args: `{"by":"colour"}`, wantErr: errBadEnum},
		{tool: "performance", args: `{}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodYear}}},
		{tool: "performance", args: `{"period":"this_month"}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodMonth}}},
		{tool: "performance", args: `{"period":"last_12_months"}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodTwelveMonths}}},
		{tool: "performance", args: `{"period":"last_year"}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodLastYear}}},
		{tool: "performance", args: `{"period":"all"}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodAll}}},
		{tool: "performance", args: `{"year":2025}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodCalendarYear, Year: 2025}}},
		{
			tool: "performance", args: `{"since":"2026-01-15"}`,
			want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodSince, Since: mustDate(t, "2026-01-15")}},
		},
		{tool: "performance", args: `{"year":2025,"period":"all"}`, want: askQuery{Command: askPerformance, Period: askPeriod{Kind: periodCalendarYear, Year: 2025}}},
		{tool: "performance", args: `{"period":"fortnight"}`, wantErr: errBadEnum},
		{tool: "performance", args: `{"year":2999}`, wantErr: errToolDate},
		{tool: "performance", args: `{"year":5}`, wantErr: errToolDate},
		{tool: "performance", args: `{"since":"2026-02-30"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"since":"2027-01-01"}`, wantErr: errToolDate},
		{tool: "value", args: `{}`, want: askQuery{Command: askValue}},
		{tool: "value", args: `{"date":"2026-02-27"}`, want: askQuery{Command: askValue, Date: mustDate(t, "2026-02-27")}},
		{tool: "value", args: `{"date":"2026-02-30"}`, wantErr: errToolDate},
		{tool: "value", args: `{"date":"2027-01-01"}`, wantErr: errToolDate},
		{tool: "value", args: `{"date":"someday"}`, wantErr: errToolDate},
	}
	byName := map[string]askTool{}
	for _, tool := range askTools() {
		byName[tool.Name] = tool
	}
	for _, tt := range tests {
		t.Run(tt.tool+" "+tt.args, func(t *testing.T) {
			t.Parallel()
			got, err := byName[tt.tool].query(tt.args, today)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" && err == nil {
				t.Errorf("query (-want +got):\n%s", diff)
			}
		})
	}
}

// TestToolArgumentErrorsQuoteNothing checks that no error a tool's arguments
// cause holds the text of an argument: the model gets it back as the tool's
// result and logs may carry it (GR-12).
func TestToolArgumentErrorsQuoteNothing(t *testing.T) {
	t.Parallel()
	today := mustDate(t, askToday)
	const secret = "SECRET-ARGUMENT"
	for _, tool := range askTools() {
		for _, args := range []string{
			`{"name":"` + secret + `","kind":"` + secret + `"}`, `{"by":"` + secret + `"}`, `{"date":"` + secret + `"}`,
			`{"since":"` + secret + `"}`, `{"period":"` + secret + `"}`, `{"` + secret + `":1}`, `{"connector":" "}`, secret,
		} {
			if _, err := tool.query(args, today); err != nil && strings.Contains(err.Error(), secret) {
				t.Errorf("tool %s, args %s: the error quotes the argument: %v", tool.Name, args, err)
			}
		}
	}
}

func TestCardResultCarriesTheDataAndNothingToDrawIt(t *testing.T) {
	t.Parallel()
	card := answerCard{
		Question: "SECRET-QUESTION", Title: "By currency, gross", Text: "Holdings at their stored valuations.",
		Rows: []answerRow{
			{Label: "PLN", Detail: "66.7%", Figure: "1 000,00 PLN", Href: "/holdings?by=currency"},
			{Label: "EUR", Figure: "500,00 EUR"},
		},
		Link: &answerLink{Href: "/?by=currency", Text: "Open the allocation"},
		Help: knownCommands(),
	}
	got := cardResult(&card)
	want := `{"title":"By currency, gross","text":"Holdings at their stored valuations.","rows":[` +
		`{"label":"PLN","detail":"66.7%","figure":"1 000,00 PLN"},{"label":"EUR","figure":"500,00 EUR"}]}`
	if got != want {
		t.Errorf("cardResult =\n%s\nwant\n%s", got, want)
	}
	for _, leak := range []string{"SECRET-QUESTION", "/holdings", "Open the allocation", "What needs me"} {
		if strings.Contains(got, leak) {
			t.Errorf("the tool result holds %q", leak)
		}
	}
	if got := cardResult(&answerCard{Title: "A <b> & c"}); got != `{"title":"A <b> & c"}` {
		t.Errorf("cardResult escaped HTML: %s", got)
	}
}

func TestToolFailureIsJSONWithNoData(t *testing.T) {
	t.Parallel()
	if got := toolFailure("withheld: this model may not receive %s", "amounts"); got != `{"error":"withheld: this model may not receive amounts"}` {
		t.Errorf("toolFailure = %s", got)
	}
}

func TestMostSensitive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		tiers map[string]bool
		want  string
	}{
		{nil, "T2"},
		{map[string]bool{"T2": true}, "T2"},
		{map[string]bool{"T2": true, "T0": true}, "T0"},
		{map[string]bool{"T1": true, "TF": true}, "TF"},
	}
	for _, tt := range tests {
		if got := mostSensitive(tt.tiers); got != tt.want {
			t.Errorf("mostSensitive(%v) = %s, want %s", tt.tiers, got, tt.want)
		}
	}
	if got := tierList(map[string]bool{"T2": true, "T0": true, "T1": true}); !slices.Equal(got, []string{"T0", "T1", "T2"}) {
		t.Errorf("tierList = %v, want most sensitive first", got)
	}
}
