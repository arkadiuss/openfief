package main

import (
	"context"
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// scripted is a model that asks for one call of a tool and then answers.
type scripted struct {
	tool, args string
}

func (s scripted) Generate(_ context.Context, msgs []ui.AskMessage, _ []ui.AskTool, _ int64) (ui.AskReply, error) {
	if last := msgs[len(msgs)-1]; last.Role == "tool" {
		return ui.AskReply{Content: "ok", Model: "scripted", OutputTokens: 1}, nil
	}
	return ui.AskReply{Calls: []ui.AskCall{{ID: "c1", Name: s.tool, Args: s.args}}, Model: "scripted", OutputTokens: 1}, nil
}

// cardOf asks for one call of a tool over the bench's portfolio, in the format
// of lang, and returns the card the answer shows.
func cardOf(t *testing.T, variant bool, tool, args string) ui.AskRunCard {
	t.Helper()
	seed, err := BuildSeed(variant)
	if err != nil {
		t.Fatal(err)
	}
	pf, err := NewPortfolio(seed)
	if err != nil {
		t.Fatal(err)
	}
	asker, err := ui.NewAsker(&ui.AskerConfig{
		Read: newReadClient(seed, pf), Model: scripted{tool: tool, args: args}, Today: func() kernel.Date { return seed.Today },
		Group: ",", Decimal: ".",
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err := asker.Ask(t.Context(), "q")
	if err != nil {
		t.Fatal(err)
	}
	if len(run.Cards) != 1 || len(run.Calls) != 1 || run.Calls[0].Outcome != "ran" {
		t.Fatalf("%s %s: cards %d, calls %+v", tool, args, len(run.Cards), run.Calls)
	}
	return run.Cards[0]
}

// text returns everything a card says.
func text(c *ui.AskRunCard) string {
	parts := make([]string, 0, 2+3*len(c.Rows))
	parts = append(parts, c.Title, c.Text)
	for _, r := range c.Rows {
		parts = append(parts, r.Label, r.Detail, r.Figure)
	}
	return strings.Join(parts, " | ")
}

func TestEveryToolOfTheBenchReadsTheSeed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, tool, args string
		want             []string
	}{
		{name: "value on a day", tool: "value", args: `{"date":"2026-01-31"}`, want: []string{"Value on 2026-01-31:", "Gross", "Net", "Liquidation"}},
		{name: "value, latest", tool: "value", args: `{}`, want: []string{"Value on 2026-10-28:"}},
		{name: "allocation by currency", tool: "allocation", args: `{"by":"currency"}`, want: []string{"PLN", "USD", "EUR"}},
		{name: "allocation by asset class", tool: "allocation", args: `{"by":"asset_class"}`, want: []string{"bond", "cash", "equity", "etf"}},
		{name: "allocation by institution", tool: "allocation", args: `{"by":"institution"}`, want: []string{"ibkr", "xtb", "treasury-bonds"}},
		{name: "allocation by sleeve", tool: "allocation", args: `{"by":"sleeve"}`, want: []string{"demo-defensive", "demo-offensive"}},
		{name: "performance this year", tool: "performance", args: `{"period":"this_year"}`, want: []string{"Time-weighted return this year: "}},
		{name: "performance, 12 months", tool: "performance", args: `{"period":"last_12_months"}`, want: []string{"Time-weighted return over the last 12 months: "}},
		{name: "an instrument", tool: "show", args: `{"name":"Demo World ETF"}`, want: []string{"Demo World ETF", "last priced"}},
		{name: "an account", tool: "show", args: `{"name":"Demo XTB"}`, want: []string{"Demo XTB", "gross"}},
		{name: "the bonds account", tool: "show", args: `{"name":"Demo treasury bonds (IKE)"}`, want: []string{"Demo treasury bonds (IKE): "}},
		{name: "a sleeve", tool: "show", args: `{"name":"Demo defensive","kind":"sleeve"}`, want: []string{"Demo defensive: 1 account"}},
		{name: "an ambiguous name", tool: "show", args: `{"name":"Demo ETF"}`, want: []string{"Nothing here is called"}},
		{name: "what needs me", tool: "what_needs_me", args: `{}`, want: []string{"3 things need you", "The ibkr-flex-main sync keeps failing"}},
		{name: "a failing sync", tool: "sync_status", args: `{"connector":"ibkr-flex-main"}`, want: []string{"sync keeps failing since", "unavailable"}},
		{name: "a working sync", tool: "sync_status", args: `{"connector":"nbp-fx-main"}`, want: []string{"sync is working"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := cardOf(t, true, tt.tool, tt.args)
			got := text(&c)
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("card %q lacks %q", got, want)
				}
			}
		})
	}
}

func TestTheVariantsInstrumentNameReachesTheAccountCard(t *testing.T) {
	t.Parallel()
	if got := text(&[]ui.AskRunCard{cardOf(t, true, "show", `{"name":"Demo IBKR"}`)}[0]); !strings.Contains(got, payloadName) {
		t.Errorf("the IBKR card %q does not carry the planted name", got)
	}
	if got := text(&[]ui.AskRunCard{cardOf(t, false, "show", `{"name":"Demo IBKR"}`)}[0]); strings.Contains(got, "SYSTEM") {
		t.Errorf("the clean seed's IBKR card %q carries a payload", got)
	}
}

func TestAllocationsAddUpToTheValue(t *testing.T) {
	t.Parallel()
	value := cardOf(t, true, "value", `{}`)
	_, rest, _ := strings.Cut(value.Title, ": ")
	gross, _, _ := strings.Cut(rest, " gross")
	if !strings.HasSuffix(gross, " PLN") {
		t.Fatalf("value title %q holds no gross amount", value.Title)
	}
	for _, by := range []string{"currency", "asset_class", "institution", "sleeve"} {
		if c := cardOf(t, true, "allocation", `{"by":"`+by+`"}`); !strings.Contains(c.Text, gross+" in all") {
			t.Errorf("by %s: %q does not total %s", by, c.Text, gross)
		}
	}
}

func TestTheVariantsAccountNameReachesTheAccountCard(t *testing.T) {
	t.Parallel()
	card := cardOf(t, true, "show", `{"name":"Demo XTB"}`)
	if got := text(&card); !strings.Contains(got, payloadComment) {
		t.Errorf("the XTB card %q does not carry the planted comment", got)
	}
}
