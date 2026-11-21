package ui

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestModelsPageOffersOpinionsOffByDefaultWithTheirTierNote(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		`id="opinions-title"`, "Opinions in the Ask box", "whether it would buy, sell or hold something and why",
		"never investment advice", "never buys, sells or starts anything", "Off by default",
		`action="/settings/models/opinions"`, "Turn opinions on",
		// The thesis tier no longer says no thesis exists.
		"the drafts you save from the Ask box",
	)
	wantNoBody(t, rec, "Turn opinions off", "Opinions send three kinds of data", "A public demo never gives opinions", "No thesis exists yet")
}

func TestTurningOpinionsOnReachesTheCoreAndSaysWhichTiersItNeeds(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(form("/settings/models/opinions", "enabled=1"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got, want := rec.Header().Get("Location"), "/settings/models?saved=opinions-on"; got != want {
		t.Fatalf("Location = %q, want %q", got, want)
	}
	if diff := cmp.Diff([]string{"opinions on"}, callsTo(h.modelSettings)); diff != "" {
		t.Errorf("calls to the core (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, rec.Header().Get("Location"), ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		// The notice after the change names the three tiers and what happens without a model allowed for them.
		"Opinions are on. The Ask box sends your question and the cards it reads (T0), prices (T2) and a suggested thesis (T1) to its model, so its route needs a model allowed for all three.",
		"Without one, opinions stay off and the box says why.",
		// The section keeps saying it, and offers the way back.
		"Opinions send three kinds of data to the Ask box's model", "Turn opinions off",
	)
	rec = h.serve(t, withSession(form("/settings/models/opinions", "enabled=0"), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	rec = h.serve(t, withSession(request(http.MethodGet, rec.Header().Get("Location"), ""), ownerSession))
	wantBody(t, rec, "Opinions are off. The Ask box declines to say whether to buy, sell or hold.", "Turn opinions on")
}

func TestModelsPageSaysWhyOpinionsHaveNoEffectWhileTheyAreOn(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.modelSettings.locked(func() { h.modelSettings.opinions = true })
	// The switch is off in the harness.
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantBody(t, rec, "Opinions have no effect while the agents-and-models switch or the Ask box model&#39;s toggle is off.")
	wantNoBody(t, rec, "No model on the Ask box&#39;s route is allowed for T0, T1 and T2")

	h.modelSettings.locked(func() {
		h.modelSettings.enabled = true
		h.modelSettings.toggles["ask"] = true
	})
	// The Ask box's route has a model that may receive T2 only.
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantBody(t, rec, "No model on the Ask box&#39;s route is allowed for T0, T1 and T2 together, so opinions stay off in effect and the box says why.")
	wantNoBody(t, rec, "Opinions have no effect while")

	h.modelSettings.locked(func() {
		for _, tier := range []string{"T0", "T1", "T2"} {
			h.modelSettings.tiers[tier] = []string{"cloud/openai/gpt-x"}
		}
	})
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantNoBody(t, rec, "stay off in effect", "Opinions have no effect while")
	// A broken entry is no model that may be asked: allowing it for the tiers leaves the route ineligible.
	h.modelSettings.locked(func() {
		h.modelSettings.routes["ask"] = []string{"gone/m"}
		for _, tier := range []string{"T0", "T1", "T2"} {
			h.modelSettings.tiers[tier] = []string{"gone/m"}
		}
	})
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantBody(t, rec, "No model on the Ask box&#39;s route is allowed for T0, T1 and T2 together")
}

func TestAPublicDemoCannotTurnOpinionsOn(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.modelSettings.locked(func() { h.modelSettings.demo = true })
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/models", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "A public demo never gives opinions: it explains and shows data and declines to recommend. This can't be turned on here.")
	if strings.Contains(rec.Body.String(), `action="/settings/models/opinions"`) {
		t.Error("a demo's page offers a way to turn opinions on")
	}
	// Posting anyway reaches the core, which refuses, and the page shows its reason.
	rec = h.serve(t, withSession(form("/settings/models/opinions", "enabled=1"), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this: ")
	h.modelSettings.locked(func() {
		if h.modelSettings.opinions {
			t.Error("opinions are on in a demo")
		}
	})
}

func TestOpinionsFormIsTheOwnersAlone(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, token := range []string{readSession, demoSession, viewerSession} {
		rec := h.serve(t, withSession(form("/settings/models/opinions", "enabled=1"), token))
		wantStatus(t, rec, http.StatusForbidden)
	}
	if calls := callsTo(h.modelSettings); len(calls) != 0 {
		t.Errorf("a session that is not the owner's reached the core: %v", calls)
	}
	for _, body := range []string{"", "enabled=maybe"} {
		rec := h.serve(t, withSession(form("/settings/models/opinions", body), ownerSession))
		wantStatus(t, rec, http.StatusBadRequest)
	}
	if calls := callsTo(h.modelSettings); len(calls) != 0 {
		t.Errorf("a malformed form reached the core: %v", calls)
	}
}
