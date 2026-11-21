package ui

import (
	"maps"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	thesisv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/thesis/v1"
)

// exampleDraft is a complete draft thesis for the tests: no real company, no
// amount.
func exampleDraft(statement string) *thesisv1.ThesisDraft {
	return &thesisv1.ThesisDraft{
		Statement: statement, Instruments: []string{"EXMP", "OTHR"}, Horizon: thesisv1.Horizon_HORIZON_LONG_TERM,
		Invalidation: []string{"Growth stalls for two quarters."}, Sources: []string{"The holdings card.", "The market price card for EXMP."},
	}
}

// draftForm is a posted form of a complete draft.
func draftBody(statement string, extra url.Values) string {
	v := url.Values{
		"statement": {statement}, "horizon": {"long_term"}, "instruments": {"EXMP\nOTHR"},
		"invalidation": {"Growth stalls for two quarters."}, "sources": {"The holdings card."},
	}
	maps.Copy(v, extra)
	return v.Encode()
}

func TestThesesPageListsTheDraftsNewestFirstWithAnEmptyState(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/theses", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "No drafts yet.", `<a href="/settings/models">Models</a>`, `<a href="/settings/theses" aria-current="page">Theses</a>`,
		"A draft is only text: no agent, no fief and no lease, and no lease size.")
	first := h.theses.seed(exampleDraft("Services revenue keeps growing."))
	short := exampleDraft("A short-lived catalyst.")
	short.Horizon = thesisv1.Horizon_HORIZON_SHORT_TERM
	second := h.theses.seed(short)
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/theses", ""), ownerSession))
	body := rec.Body.String()
	wantBody(t, rec, `href="/settings/theses/`+first+`"`, `href="/settings/theses/`+second+`"`, "Services revenue keeps growing.",
		"Short-term", "Long-term", `<span class="pill pill-neutral mono">EXMP</span>`, "Version 1, saved")
	if strings.Index(body, second) > strings.Index(body, first) {
		t.Error("the list shows the older draft first")
	}
	wantNoBody(t, rec, "No drafts yet.")
	// A long statement is cut in the list and whole on its page.
	long := strings.Repeat("word ", 80)
	id := h.theses.seed(exampleDraft(long))
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/theses", ""), ownerSession))
	wantBody(t, rec, "…")
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/theses/"+id, ""), ownerSession))
	wantBody(t, rec, strings.TrimSpace(long))
}

func TestThesisPageShowsTheDraftItsFormItsHistoryAndNoAgentButton(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	id := h.theses.seed(exampleDraft("Services revenue keeps growing."))
	rec := h.serve(t, withSession(request(http.MethodGet, "/settings/theses/"+id, ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec,
		"Services revenue keeps growing.", "Wrong if", "Growth stalls for two quarters.", "The market price card for EXMP.",
		"T1 data: a thesis's text", "Start an agent: arrives with the control plane.",
		`name="base_version" value="1"`, `name="statement"`, `name="horizon"`, `name="instruments"`, `name="invalidation"`, `name="sources"`,
		"Version history", "Version 1 (current)", "never the text",
	)
	// Stop at suggest and save: no agent, fief, lease or spawn button.
	wantNoBody(t, rec, "Start an agent</button>", "disabled", "Start the agent", "Spawn", "Create a fief", "Lease it")
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/theses/ffffffffffffffff", ""), ownerSession))
	wantStatus(t, rec, http.StatusNotFound)
}

func TestAnEditAddsAVersionAndKeepsTheEarlierOnes(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	id := h.theses.seed(exampleDraft("Services revenue keeps growing."))
	rec := h.serve(t, withSession(form("/settings/theses/"+id, draftBody("Services revenue keeps growing, and margins widen.", url.Values{
		"base_version": {"1"}, "horizon": {"short_term"}, "instruments": {"EXMP\r\n\r\n OTHR "}, "sources": {"The holdings card.\nThe market price card for EXMP."},
	})), ownerSession))
	wantStatus(t, rec, http.StatusSeeOther)
	if got, want := rec.Header().Get("Location"), "/settings/theses/"+id+"?saved=edit"; got != want {
		t.Fatalf("Location = %q, want %q", got, want)
	}
	if diff := cmp.Diff([]string{"edit " + id + " v1"}, h.callsOf()); diff != "" {
		t.Errorf("calls to the core (-want +got):\n%s", diff)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, "/settings/theses/"+id+"?saved=edit", ""), ownerSession))
	wantBody(t, rec,
		"Saved as version 2. The earlier versions are kept below", "Version 2 (current)", "Version 1 ·",
		"Services revenue keeps growing, and margins widen.", "Short-term", `name="base_version" value="2"`,
	)
	var got *thesisv1.ThesisDraft
	h.theses.locked(func() { got = h.theses.byID[id][1].GetDraft() })
	if diff := cmp.Diff([]string{"EXMP", "OTHR"}, got.GetInstruments()); diff != "" {
		t.Errorf("the posted lines (-want +got):\n%s", diff)
	}
	// The first version is still the first text.
	h.theses.locked(func() {
		if first := h.theses.byID[id][0].GetDraft().GetStatement(); first != "Services revenue keeps growing." {
			t.Errorf("version 1 now says %q", first)
		}
	})
}

func (h *harness) callsOf() []string {
	var out []string
	h.theses.locked(func() { out = append(out, h.theses.calls...) })
	return out
}

func TestAnEditTheCoreRefusesShowsTheOwnersTextAndTheReason(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	id := h.theses.seed(exampleDraft("Services revenue keeps growing."))
	// No statement: the core refuses, and the page keeps what the owner typed in the other fields.
	rec := h.serve(t, withSession(form("/settings/theses/"+id, draftBody("  ", url.Values{
		"base_version": {"1"}, "instruments": {"KEEPME"},
	})), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this edit: ", "KEEPME")
	// A stale edit says the draft has a newer version, and keeps the text too.
	if first := h.serve(t, withSession(form("/settings/theses/"+id, draftBody("Second.", url.Values{"base_version": {"1"}})), ownerSession)); first.Code != http.StatusSeeOther {
		t.Fatalf("the first edit: status %d", first.Code)
	}
	rec = h.serve(t, withSession(form("/settings/theses/"+id, draftBody("Typed on the old version.", url.Values{"base_version": {"1"}})), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "This draft has a newer version, saved since you opened it.", "Typed on the old version.")
	h.theses.locked(func() {
		if n := len(h.theses.byID[id]); n != 2 {
			t.Errorf("the draft has %d versions, want the 2 saved", n)
		}
	})
	rec = h.serve(t, withSession(form("/settings/theses/ffffffffffffffff", draftBody("Nothing.", url.Values{"base_version": {"1"}})), ownerSession))
	wantStatus(t, rec, http.StatusNotFound)
	for _, body := range []string{"statement=x", "base_version=one&statement=x"} {
		rec = h.serve(t, withSession(form("/settings/theses/"+id, body), ownerSession))
		wantStatus(t, rec, http.StatusBadRequest)
	}
}

func TestTheAskBoxCardSavesADraftWithHTMXOrWithout(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	post := func(body string, htmx bool) *http.Request {
		r := withSession(form("/settings/theses", body), ownerSession)
		if htmx {
			r.Header.Set(hxRequest, "true")
		}
		return r
	}
	rec := h.serve(t, post(draftBody("Services revenue keeps growing.", nil), true))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Saved as a draft thesis.", `<a href="/settings/theses/0000000000000001">Open it on the Theses page</a>`, "It has no agent, fief or lease.")
	// Without htmx the page redirects to the draft.
	rec = h.serve(t, post(draftBody("A second.", nil), false))
	wantStatus(t, rec, http.StatusSeeOther)
	if got, want := rec.Header().Get("Location"), "/settings/theses/0000000000000002?saved=1"; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
	rec = h.serve(t, withSession(request(http.MethodGet, rec.Header().Get("Location"), ""), ownerSession))
	wantBody(t, rec, "Saved as a draft thesis, version 1.", "A second.")
	// A draft the core refuses says why in the card.
	rec = h.serve(t, post(draftBody("", nil), true))
	wantStatus(t, rec, http.StatusBadRequest)
	wantBody(t, rec, "The core refused this draft: ", `role="alert"`)
	rec = h.serve(t, post("horizon=nonsense&statement=x", true))
	wantStatus(t, rec, http.StatusBadRequest)
	if diff := cmp.Diff([]string{"save 0000000000000001", "save 0000000000000002"}, h.callsOf()); diff != "" {
		t.Errorf("saves reaching the core (-want +got):\n%s", diff)
	}
}

func TestThesesAreTheOwnersAloneNoOtherSessionSeesThemOrTheirLink(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	id := h.theses.seed(exampleDraft("Services revenue keeps growing."))
	for _, token := range []string{readSession, demoSession, viewerSession} {
		for _, target := range []string{"/settings/theses", "/settings/theses/" + id} {
			rec := h.serve(t, withSession(request(http.MethodGet, target, ""), token))
			wantStatus(t, rec, http.StatusForbidden)
			wantNoBody(t, rec, "Services revenue keeps growing.", "EXMP")
		}
		for _, target := range []string{"/settings/theses", "/settings/theses/" + id} {
			rec := h.serve(t, withSession(form(target, draftBody("Mine.", url.Values{"base_version": {"1"}})), token))
			wantStatus(t, rec, http.StatusForbidden)
		}
		rec := h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), token))
		wantStatus(t, rec, http.StatusOK)
		wantNoBody(t, rec, `href="/settings/theses"`)
	}
	if calls := h.callsOf(); len(calls) != 0 {
		t.Errorf("a session that is not the owner's changed drafts: %v", calls)
	}
	// Signed out, a visitor goes to the login page and the link shows only for the owner.
	rec := h.serve(t, request(http.MethodGet, "/settings/theses", ""))
	wantStatus(t, rec, http.StatusSeeOther)
	rec = h.serve(t, withSession(request(http.MethodGet, "/sleeves", ""), ownerSession))
	wantBody(t, rec, `href="/settings/theses"`)
}

func TestWithoutADraftThesesServiceThePageIsNotFound(t *testing.T) {
	t.Parallel()
	h := newHarnessWith(t, nil, func(cfg *Config) { cfg.Theses = nil })
	for _, target := range []string{"/settings/theses", "/settings/theses/0000000000000001"} {
		rec := h.serve(t, withSession(request(http.MethodGet, target, ""), ownerSession))
		wantStatus(t, rec, http.StatusNotFound)
	}
	rec := h.serve(t, withSession(form("/settings/theses", draftBody("x", nil)), ownerSession))
	wantStatus(t, rec, http.StatusBadRequest)
}

func TestPostedLinesAreTrimmedAndSplit(t *testing.T) {
	t.Parallel()
	got := postedLines([]string{" a \r\n\r\n b ", "", "c\nd", "  "})
	if diff := cmp.Diff([]string{"a", "b", "c", "d"}, got); diff != "" {
		t.Errorf("postedLines (-want +got):\n%s", diff)
	}
	if got := preview(strings.Repeat("é", listPreviewRunes+10)); len([]rune(got)) != listPreviewRunes+1 || !strings.HasSuffix(got, "…") {
		t.Errorf("preview of a long text has %d characters, want %d and a mark", len([]rune(got)), listPreviewRunes+1)
	}
}
