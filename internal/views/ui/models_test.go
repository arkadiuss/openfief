package ui

import (
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	identityv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/identity/v1"
	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// callsTo returns the changes the fake core was asked for.
func callsTo(f *fakeModelSettings) []string {
	var out []string
	f.locked(func() { out = slices.Clone(f.calls) })
	return out
}

func TestModelsPageShowsTheSwitchTheProvidersTheTiersAndTheRoutes(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		// The switch is off, and says what that means.
		`id="switch-title"`, "Turn agents and models on", "nothing calls a model",
		// The features, the MCP server's plainly not built.
		"Ask box model", "Maintainer", "MCP server", "Not built yet", "no MCP endpoint exists",
		`action="/settings/models/features/mcp"`,
		// The providers with the terms they declare, prices and flags.
		`id="providers-title"`, "cloud", "Its models run in a third party&#39;s cloud.", "keeps a request and its answer for 30 days",
		"does not train models on them", "openai/gpt-x", "0.15 USD", "0.6 USD", "128000",
		"Its models run on hardware you run.", "keeps nothing of a request", "qwen3:8b", "no price", "not stated",
		"Not answering now", "sleepy",
		// The tiers, explained, and one box per model and tier, ticked where allowed.
		`id="tiers-title"`, "Whole portfolio", "Public data", "Nothing is allowed until you tick it",
		`name="allow" value="T0|cloud/openai/gpt-x"`, `name="allow" value="T2|home/qwen3:8b"`,
		`aria-label="Allow cloud/openai/gpt-x to receive T2" checked`,
		// The routes.
		`id="routes-title"`, `id="route-ask"`, `id="route-maintainer"`, "Primary", "Fallback 1", "gone/m",
		`action="/settings/models/routes/ask"`,
		// The settings navigation names the page.
		`<a href="/settings/models" aria-current="page">Models</a>`,
	)
	wantNoBody(t, rec, `aria-label="Allow cloud/openai/gpt-x to receive T0" checked`)
}

func TestModelsPageShowsABrokenRouteEntryWithAFix(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.modelSettings.locked(func() {
		h.modelSettings.routes["ask"] = []string{"gone/m", "sleepy/x", "cloud/missing", "cloud/openai/gpt-x"}
		h.modelSettings.broken = []*modelsv1.BrokenEntry{
			{Role: "ask", Model: "gone/m", Reason: "missing"},
			{Role: "maintainer", Model: "home/qwen3:8b", Reason: "not-model-provider"},
		}
	})
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`<span class="pill pill-danger">Broken</span>`, "Broken: no plugin has that name.", "Broken: that plugin is not a model provider.",
		"Fix it: remove this entry and choose another model, or install the plugin on the", `<a href="/plugins">Plugins</a>`,
		"Its provider is not answering now.", "Its provider does not list this model.",
	)
}

func TestModelsPageWarnsWhenNoModelOfARouteMayReceiveWhatTheRoleSends(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantBody(t, rec, "No model of this route may receive T0 and T2")
	h.modelSettings.locked(func() {
		h.modelSettings.tiers["T0"] = []string{"home/qwen3:8b"}
		h.modelSettings.tiers["T2"] = []string{"home/qwen3:8b", "cloud/openai/gpt-x"}
	})
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	// The maintainer's only model may now receive both; the Ask box's primary still may not receive T0.
	if got := strings.Count(rec.Body.String(), "No model of this route may receive"); got != 1 {
		t.Errorf("%d warnings, want one (the Ask box's route, whose models may not receive T0)", got)
	}
	wantBody(t, rec, `title="This model may not receive T0"`)
}

func TestModelsPageIsTheOwnersAloneAndNoOtherSessionSeesItsLink(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	posts := map[string]string{
		"/settings/models/switch":            "enabled=1",
		"/settings/models/features/ask":      "enabled=1",
		"/settings/models/tiers":             "allow=T0%7Chome%2Fqwen3%3A8b",
		"/settings/models/routes/maintainer": "op=remove&index=0",
	}
	for _, token := range []string{readSession, demoSession, viewerSession} {
		rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), token))
		wantStatus(t, rec, http.StatusForbidden)
		wantNoBody(t, rec, "Turn agents and models on", `id="switch-title"`)
		for target, body := range posts {
			posted := h.serve(t, withSession(form(target, body), token))
			wantStatus(t, posted, http.StatusForbidden)
		}
		// Every page that shows the settings navigation leaves the link out.
		rec = h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), token))
		wantStatus(t, rec, http.StatusOK)
		wantNoBody(t, rec, `href="/settings/models"`)
	}
	if calls := callsTo(h.modelSettings); len(calls) != 0 {
		t.Errorf("a session that is not the owner's reached the core: %v", calls)
	}
	rec := h.serve(t, request(http.MethodGet, "/settings/models", ""))
	wantStatus(t, rec, http.StatusSeeOther)
	rec = h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), ownerSession))
	wantBody(t, rec, `href="/settings/models"`)
}

func TestModelsFormsReachTheCoreWithTheOwnersSession(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, target, body string
		wantCalls          []string
		wantSaved          string
		check              func(t *testing.T, f *fakeModelSettings)
	}{
		{
			name: "the switch on", target: "/settings/models/switch", body: "enabled=1", wantCalls: []string{"switch on"}, wantSaved: "switch",
			check: func(t *testing.T, f *fakeModelSettings) {
				t.Helper()
				if !f.enabled {
					t.Error("the switch is off")
				}
			},
		},
		{name: "the switch off", target: "/settings/models/switch", body: "enabled=0", wantCalls: []string{"switch off"}, wantSaved: "switch"},
		{
			name: "a toggle on", target: "/settings/models/features/maintainer", body: "enabled=1", wantCalls: []string{"toggle maintainer on"}, wantSaved: "feature",
			check: func(t *testing.T, f *fakeModelSettings) {
				t.Helper()
				if !f.toggles["maintainer"] {
					t.Error("the maintainer's toggle is off")
				}
			},
		},
		{name: "the MCP toggle, stored only", target: "/settings/models/features/mcp", body: "enabled=1", wantCalls: []string{"toggle mcp on"}, wantSaved: "feature"},
		{
			name: "a model added to a route from the choices", target: "/settings/models/routes/maintainer", body: "op=add&model=cloud%2Fopenai%2Fgpt-x&ref=",
			wantCalls: []string{"route maintainer home/qwen3:8b,cloud/openai/gpt-x"}, wantSaved: "route",
		},
		{
			name: "a model added to a route by typing it", target: "/settings/models/routes/maintainer", body: "op=add&model=&ref=+other%2Fm+",
			wantCalls: []string{"route maintainer home/qwen3:8b,other/m"}, wantSaved: "route",
		},
		{
			name: "an entry moved down", target: "/settings/models/routes/ask", body: "op=down&index=0",
			wantCalls: []string{"route ask gone/m,cloud/openai/gpt-x"}, wantSaved: "route",
		},
		{
			name: "an entry moved up", target: "/settings/models/routes/ask", body: "op=up&index=1",
			wantCalls: []string{"route ask gone/m,cloud/openai/gpt-x"}, wantSaved: "route",
		},
		{
			name: "a broken entry removed", target: "/settings/models/routes/ask", body: "op=remove&index=1",
			wantCalls: []string{"route ask cloud/openai/gpt-x"}, wantSaved: "route",
		},
		{
			name: "the last entry removed clears the route", target: "/settings/models/routes/maintainer", body: "op=remove&index=0",
			wantCalls: []string{"route maintainer "}, wantSaved: "route",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, withSession(form(tt.target, tt.body), ownerSession))
			wantStatus(t, rec, http.StatusSeeOther)
			if got, want := rec.Header().Get("Location"), "/settings/models?saved="+tt.wantSaved; got != want {
				t.Errorf("Location = %q, want %q", got, want)
			}
			if diff := cmp.Diff(tt.wantCalls, callsTo(h.modelSettings)); diff != "" {
				t.Errorf("calls to the core (-want +got):\n%s", diff)
			}
			if tt.check != nil {
				h.modelSettings.locked(func() { tt.check(t, h.modelSettings) })
			}
		})
	}
}

func TestModelsTierGridChangesOnlyTheTiersThatChanged(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	// T2 stays as it is; T0 gains the home model; TF gains two; nothing else posted clears nothing else.
	body := url.Values{"allow": {
		"T2|cloud/openai/gpt-x", "T0|home/qwen3:8b", "TF|home/qwen3:8b", "TF|cloud/openai/gpt-x",
	}}.Encode()
	rec := h.serve(t, withSession(form("/settings/models/tiers", body), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); !strings.HasPrefix(got, "/settings/models?saved=tiers") {
		t.Errorf("Location = %q", got)
	}
	got := callsTo(h.modelSettings)
	slices.Sort(got)
	want := []string{"tier T0 home/qwen3:8b", "tier TF home/qwen3:8b,cloud/openai/gpt-x"}
	if diff := cmp.Diff(want, got, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
		t.Errorf("calls to the core (-want +got):\n%s", diff)
	}
	// Unticking everything clears the tiers that had models: deny by default.
	rec = h.serve(t, withSession(form("/settings/models/tiers", ""), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	h.modelSettings.locked(func() {
		for tier, models := range h.modelSettings.tiers {
			if len(models) != 0 {
				t.Errorf("tier %s still allows %v after unticking everything", tier, models)
			}
		}
	})
}

func TestModelsMalformedFormsNeverReachTheCore(t *testing.T) {
	t.Parallel()
	posts := []struct{ name, target, body string }{
		{"a switch without its state", "/settings/models/switch", ""},
		{"a switch with a word", "/settings/models/switch", "enabled=yes"},
		{"an unknown feature", "/settings/models/features/trade", "enabled=1"},
		{"a toggle without its state", "/settings/models/features/ask", "enabled=2"},
		{"a tier box of no tier", "/settings/models/tiers", "allow=T9%7Cp%2Fm"},
		{"a tier box without a tier", "/settings/models/tiers", "allow=p%2Fm"},
		{"an unknown role", "/settings/models/routes/trader", "op=add&ref=p%2Fm"},
		{"an unknown operation", "/settings/models/routes/ask", "op=swap&index=0"},
		{"an add naming nothing", "/settings/models/routes/ask", "op=add&model=&ref=+"},
		{"a remove without an index", "/settings/models/routes/ask", "op=remove"},
		{"a remove past the end", "/settings/models/routes/ask", "op=remove&index=2"},
		{"a remove of a negative entry", "/settings/models/routes/ask", "op=remove&index=-1"},
		{"the first entry moved up", "/settings/models/routes/ask", "op=up&index=0"},
		{"the last entry moved down", "/settings/models/routes/ask", "op=down&index=1"},
	}
	for _, tt := range posts {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := newHarness(t)
			rec := h.serve(t, withSession(form(tt.target, tt.body), ownerSession))
			wantStatus(t, rec, http.StatusBadRequest)
			if calls := callsTo(h.modelSettings); len(calls) != 0 {
				t.Errorf("a malformed form reached the core: %v", calls)
			}
		})
	}
}

func TestModelsTheCoresRefusalShowsItsReasonOnThePage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	// The core refuses a route that names a model twice, and a reference that is no provider/model.
	for _, body := range []string{"op=add&model=cloud%2Fopenai%2Fgpt-x&ref=", "op=add&model=&ref=nomodel"} {
		rec := h.serve(t, withSession(form("/settings/models/routes/ask", body), ownerSession))
		wantStatus(t, rec, http.StatusBadRequest)
		wantBody(t, rec, "The core refused this:", `id="routes-title"`)
	}
	h.modelSettings.locked(func() { h.modelSettings.refuse = errStale })
	rec := h.serve(t, withSession(form("/settings/models/switch", "enabled=1"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this:", errStale.Error())
	h.modelSettings.locked(func() {
		if h.modelSettings.enabled {
			t.Error("a refused switch is on")
		}
	})
}

func TestModelsACoreThatIsDownIsABadGatewayNotAnEmptyPage(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.modelSettings.locked(func() { h.modelSettings.unavailable = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusBadGateway)
	rec = h.serve(t, withSession(form("/settings/models/switch", "enabled=1"), ownerSession))
	wantStatus(t, rec, http.StatusBadGateway)
}

func TestModelsNoticeWords(t *testing.T) {
	t.Parallel()
	for saved, want := range map[string]string{
		"switch": "The agents-and-models switch is saved.", "feature": "The toggle is saved.",
		"tiers": "The models allowed for each data tier are saved.", "route": "The route is saved.", "x": "",
	} {
		if got := modelsNotice(url.Values{"saved": {saved}}); got != want {
			t.Errorf("modelsNotice(%q) = %q, want %q", saved, got, want)
		}
	}
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models?saved=route", ""), ownerSession))
	wantBody(t, rec, "The route is saved.")
}

func TestEditRoute(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		route []string
		form  string
		want  []string
		ok    bool
	}{
		{"add from the choices", []string{"a/1"}, "op=add&model=b%2F2&ref=c%2F3", []string{"a/1", "b/2"}, true},
		{"add typed", []string{"a/1"}, "op=add&model=&ref=c%2F3", []string{"a/1", "c/3"}, true},
		{"add to an empty route", nil, "op=add&ref=a%2F1", []string{"a/1"}, true},
		{"remove the middle", []string{"a/1", "b/2", "c/3"}, "op=remove&index=1", []string{"a/1", "c/3"}, true},
		{"up", []string{"a/1", "b/2", "c/3"}, "op=up&index=2", []string{"a/1", "c/3", "b/2"}, true},
		{"down", []string{"a/1", "b/2", "c/3"}, "op=down&index=0", []string{"b/2", "a/1", "c/3"}, true},
		{"up at the top", []string{"a/1", "b/2"}, "op=up&index=0", nil, false},
		{"down at the bottom", []string{"a/1", "b/2"}, "op=down&index=1", nil, false},
		{"an index that is not a number", []string{"a/1"}, "op=remove&index=x", nil, false},
		{"no operation", []string{"a/1"}, "", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, err := url.ParseQuery(tt.form)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := editRoute(slices.Clone(tt.route), f)
			if ok != tt.ok || cmp.Diff(tt.want, got, cmpopts.EquateEmpty()) != "" {
				t.Errorf("editRoute = %v, %v; want %v, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestRetentionWords(t *testing.T) {
	t.Parallel()
	tests := []struct {
		r    *modelsv1.Retention
		want string
	}{
		{&modelsv1.Retention{}, "It declares it keeps nothing of a request or its answer, and that it does not train models on them."},
		{&modelsv1.Retention{Days: 1}, "It declares it keeps a request and its answer for 1 day, and that it does not train models on them."},
		{&modelsv1.Retention{Days: 30, MayTrain: true}, "It declares it keeps a request and its answer for 30 days, and that it may train models on them."},
	}
	for _, tt := range tests {
		if got := retentionWords(tt.r); got != tt.want {
			t.Errorf("retentionWords(%v) = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestModelsGridListsAModelNoProviderOffersAnyMoreSoItCanBeUnticked(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.modelSettings.locked(func() { h.modelSettings.tiers["T0"] = []string{"retired/old-model", "sleepy/x"} })
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`value="T0|retired/old-model" aria-label="Allow retired/old-model to receive T0" checked`, "No model provider has that name.",
		`value="T0|sleepy/x" aria-label="Allow sleepy/x to receive T0" checked`, "Its provider is not answering now.",
	)
}

func TestAuditLogNamesASettingsChangeAndItsSetting(t *testing.T) {
	t.Parallel()
	if got := actionWords(identityv1.AuditAction_AUDIT_ACTION_MODEL_SETTINGS_CHANGED); got != "Models settings changed" {
		t.Errorf("actionWords = %q", got)
	}
	ref := &identityv1.AuditRef{Kind: identityv1.AuditRefKind_AUDIT_REF_KIND_MODEL_SETTING, Id: "route:ask"}
	if got := (&auditNames{}).ref(ref); got != "Setting route:ask" {
		t.Errorf("ref = %q, want the setting by its name", got)
	}
}
