package answer

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestAskToolsAreTheReadOnlyAnswerers checks that the tools the model may call
// are exactly the commands' answerers and nothing more (D-203): no tool
// changes anything, and every tool has a name, a description, a schema that is
// a JSON object schema, a card label and a data tier of D-206.
func TestAskToolsAreTheReadOnlyAnswerers(t *testing.T) {
	t.Parallel()
	tools := Tools()
	names := make([]string, 0, len(tools))
	for i := range tools {
		names = append(names, tools[i].Name)
		checkTool(t, &tools[i])
	}
	want := []string{"what_needs_me", "show", "sync_status", "allocation", "performance", "value", "market_price"}
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
func checkTool(t *testing.T, tool *Tool) {
	t.Helper()
	if tool.Description == "" || tool.Label == "" || tool.query == nil || !slices.Contains([]string{TierT0, TierTF, TierT1, TierT2}, tool.Tier) {
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
	today := mustDay(t, askToday)
	for _, tool := range Tools() {
		t.Run(tool.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := tool.Query(`{"surprise":1}`, today); !errors.Is(err, errBadArgs) {
				t.Errorf("an unknown argument: error = %v, want errBadArgs", err)
			}
			if _, err := tool.Query(`[]`, today); !errors.Is(err, errBadArgs) {
				t.Errorf("arguments that are no object: error = %v, want errBadArgs", err)
			}
			if _, err := tool.Query(`{} {}`, today); !errors.Is(err, errBadArgs) {
				t.Errorf("two objects: error = %v, want errBadArgs", err)
			}
		})
	}
}

func TestAskToolArguments(t *testing.T) {
	t.Parallel()
	today := mustDay(t, askToday)
	tests := []struct {
		tool    string
		args    string
		want    Query
		wantErr error
	}{
		{tool: "what_needs_me", args: `{}`, want: Query{Command: CommandNeeds}},
		{tool: "what_needs_me", args: ``, want: Query{Command: CommandNeeds}},
		{tool: "show", args: `{"name":"Synthetic brokerage"}`, want: Query{Command: CommandShow, Term: "Synthetic brokerage"}},
		{tool: "show", args: `{"name":" bonds ","kind":"sleeve"}`, want: Query{Command: CommandShow, Term: "bonds", Show: ShowSleeve}},
		{tool: "show", args: `{"name":"PLTEST000010","kind":"instrument"}`, want: Query{Command: CommandShow, Term: "PLTEST000010", Show: ShowInstrument}},
		{tool: "show", args: `{"name":"x","kind":"account"}`, want: Query{Command: CommandShow, Term: "x", Show: ShowAccount}},
		{tool: "show", args: `{"kind":"account"}`, wantErr: errNoName},
		{tool: "show", args: `{"name":"  "}`, wantErr: errNoName},
		{tool: "show", args: `{"name":"x","kind":"portfolio"}`, wantErr: errBadEnum},
		{tool: "sync_status", args: `{"connector":"xtb-main"}`, want: Query{Command: CommandSync, Term: "xtb-main"}},
		{tool: "sync_status", args: `{}`, wantErr: errNoName},
		{tool: "allocation", args: `{}`, want: Query{Command: CommandAllocation, Dimension: DimAssetClass}},
		{tool: "allocation", args: `{"by":"asset_class"}`, want: Query{Command: CommandAllocation, Dimension: DimAssetClass}},
		{tool: "allocation", args: `{"by":"currency"}`, want: Query{Command: CommandAllocation, Dimension: DimCurrency}},
		{tool: "allocation", args: `{"by":"institution"}`, want: Query{Command: CommandAllocation, Dimension: DimInstitution}},
		{tool: "allocation", args: `{"by":"sleeve"}`, want: Query{Command: CommandAllocation, Dimension: DimSleeve}},
		{tool: "allocation", args: `{"by":"colour"}`, wantErr: errBadEnum},
		{tool: "performance", args: `{}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodYear}}},
		{tool: "performance", args: `{"period":"this_month"}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonth}}},
		{tool: "performance", args: `{"period":"last_12_months"}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodTwelveMonths}}},
		{tool: "performance", args: `{"period":"last_year"}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodLastYear}}},
		{tool: "performance", args: `{"period":"all"}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodAll}}},
		{tool: "performance", args: `{"year":2025}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodCalendarYear, Year: 2025}}},
		{
			tool: "performance", args: `{"since":"2026-01-15"}`,
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodSince, Since: mustDay(t, "2026-01-15")}},
		},
		{
			tool: "performance", args: `{"month":"2025-11"}`,
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.November}},
		},
		{
			tool: "performance", args: `{"from":"2025-11-01","to":"2025-11-30"}`,
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodBetween, Since: mustDay(t, "2025-11-01"), To: mustDay(t, "2025-11-30")}},
		},
		{
			tool: "performance", args: `{"from":"2025-11-01","to":"2025-11-30","month":"2025-10","year":2024,"since":"2026-01-01","period":"all"}`,
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodBetween, Since: mustDay(t, "2025-11-01"), To: mustDay(t, "2025-11-30")}},
		},
		{
			tool: "performance", args: `{"month":"2025-10","year":2024}`,
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodMonthOf, Year: 2025, Month: time.October}},
		},
		{
			tool: "performance", args: `{"from":"2025-11-30","to":"2025-11-01"}`,
			want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodBetween, Since: mustDay(t, "2025-11-30"), To: mustDay(t, "2025-11-01")}},
		},
		{tool: "performance", args: `{"from":"2025-11-01"}`, wantErr: errPeriodBounds},
		{tool: "performance", args: `{"to":"2025-11-30"}`, wantErr: errPeriodBounds},
		{tool: "performance", args: `{"from":"2025-11-31","to":"2025-12-01"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"from":"soon","to":"2025-12-01"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"month":"2025-13"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"month":"2025"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"month":"2025-11-01 to 2025-11-30"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"year":2025,"period":"all"}`, want: Query{Command: CommandPerformance, Period: Period{Kind: PeriodCalendarYear, Year: 2025}}},
		{tool: "performance", args: `{"period":"fortnight"}`, wantErr: errBadEnum},
		{tool: "performance", args: `{"year":2999}`, wantErr: errToolDate},
		{tool: "performance", args: `{"year":5}`, wantErr: errToolDate},
		{tool: "performance", args: `{"since":"2026-02-30"}`, wantErr: errToolDate},
		{tool: "performance", args: `{"since":"2027-01-01"}`, wantErr: errToolDate},
		{tool: "value", args: `{}`, want: Query{Command: CommandValue}},
		{tool: "value", args: `{"date":"2026-02-27"}`, want: Query{Command: CommandValue, Date: mustDay(t, "2026-02-27")}},
		{
			tool: "value", args: `{"period":{"from":"2025-11-01","to":"2025-11-30"}}`,
			want: Query{Command: CommandValue, Change: &Period{Kind: PeriodBetween, Since: mustDay(t, "2025-11-01"), To: mustDay(t, "2025-11-30")}},
		},
		{tool: "value", args: `{"period":{"from":"2025-11-01"}}`, wantErr: errPeriodBounds},
		{tool: "value", args: `{"period":{}}`, wantErr: errPeriodBounds},
		{tool: "value", args: `{"period":{"from":"2025-11-01","to":"never"}}`, wantErr: errToolDate},
		{tool: "value", args: `{"period":{"from":"2025-11-01","to":"2025-11-30","extra":1}}`, wantErr: errBadArgs},
		{tool: "value", args: `{"period":"2025-11"}`, wantErr: errBadArgs},
		{tool: "value", args: `{"date":"2025-11-30","period":{"from":"2025-11-01","to":"2025-11-30"}}`, wantErr: errDateAndPeriod},
		{tool: "value", args: `{"date":"2026-02-30"}`, wantErr: errToolDate},
		{tool: "value", args: `{"date":"2027-01-01"}`, wantErr: errToolDate},
		{tool: "value", args: `{"date":"someday"}`, wantErr: errToolDate},
	}
	byName := map[string]Tool{}
	for _, tool := range Tools() {
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
	today := mustDay(t, askToday)
	const secret = "SECRET-ARGUMENT"
	for _, tool := range Tools() {
		for _, args := range []string{
			`{"name":"` + secret + `","kind":"` + secret + `"}`, `{"by":"` + secret + `"}`, `{"date":"` + secret + `"}`,
			`{"since":"` + secret + `"}`, `{"period":"` + secret + `"}`, `{"` + secret + `":1}`, `{"connector":" "}`, secret,
		} {
			if _, err := tool.Query(args, today); err != nil && strings.Contains(err.Error(), secret) {
				t.Errorf("tool %s, args %s: the error quotes the argument: %v", tool.Name, args, err)
			}
		}
	}
}

func TestCardResultCarriesTheDataAndNothingToDrawIt(t *testing.T) {
	t.Parallel()
	card := Card{
		Title: "By currency, gross", Text: "Holdings at their stored valuations.",
		Rows: []Row{
			{Label: "PLN", Detail: "66.7%", Figure: "1 000,00 PLN", Href: "/holdings?by=currency"},
			{Label: "EUR", Figure: "500,00 EUR"},
		},
		Link: &Link{Href: "/?by=currency", Text: "Open the allocation"},
		Help: KnownCommands(),
	}
	got := CardResult(&card, "c2")
	want := `{"id":"c2","title":"By currency, gross","text":"Holdings at their stored valuations.","rows":[` +
		`{"label":"PLN","detail":"66.7%","figure":"1 000,00 PLN"},{"label":"EUR","figure":"500,00 EUR"}]}`
	if got != want {
		t.Errorf("CardResult =\n%s\nwant\n%s", got, want)
	}
	for _, leak := range []string{"/holdings", "Open the allocation", "What needs me"} {
		if strings.Contains(got, leak) {
			t.Errorf("the tool result holds %q", leak)
		}
	}
	if got := CardResult(&Card{Title: "A <b> & c"}, "c1"); got != `{"id":"c1","title":"A <b> & c"}` {
		t.Errorf("CardResult escaped HTML: %s", got)
	}
}

func TestToolFailureIsJSONWithNoData(t *testing.T) {
	t.Parallel()
	if got := ToolFailure("withheld: this model may not receive %s", "amounts"); got != `{"error":"withheld: this model may not receive amounts"}` {
		t.Errorf("ToolFailure = %s", got)
	}
}

// TestThePerformanceToolIsFoundByReturnAndPerformanceInEitherLanguage pins that
// the description says the tool answers questions about return and performance
// over a period and names the Polish terms, so that a model reaches for it when
// asked for "stopa zwrotu" or "wynik" (D-203).
func TestThePerformanceToolIsFoundByReturnAndPerformanceInEitherLanguage(t *testing.T) {
	t.Parallel()
	tool := performanceTool()
	for _, want := range []string{"return and performance over a period", `"stopa zwrotu"`, `"wynik"`} {
		if !strings.Contains(tool.Description, want) {
			t.Errorf("the description %q lacks %q", tool.Description, want)
		}
	}
}
