package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// testClock is a clock the test moves: the UI's Config.Now.
type testClock struct {
	mu sync.Mutex
	t  time.Time
}

func newTestClock() *testClock {
	return &testClock{t: time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)}
}

func (c *testClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *testClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

const eventsScript = `<script src="/static/events.js"></script>`

// postClientEvents posts the owner's choice for the client events as the form does.
func (h *harness) postClientEvents(t *testing.T, session, enabled string) *httptest.ResponseRecorder {
	t.Helper()
	return h.serve(t, withSession(form("/settings/client-events", "enabled="+enabled), session))
}

func TestClientEventsFollowTheOwnersSettingInTheCore(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	h, logs, _ := clientEventsHarness(t, func(c *Config) { c.Now = clock.now })
	const body = `{"events":[{"k":"turnstile-success","ms":5}]}`
	login := func() string { return h.serve(t, request(http.MethodGet, "/login", "")).Body.String() }

	// On by default (D-227): the page loads the script and the endpoint takes the report.
	if !strings.Contains(login(), eventsScript) {
		t.Error("by default the login page does not load events.js")
	}
	wantStatus(t, h.serve(t, report(body)), http.StatusNoContent)
	if got := strings.Count(logs.String(), "client event"); got != 1 {
		t.Fatalf("%d reports were logged with client events on, want 1", got)
	}

	// The owner turns them off in Settings: the core stores it, and this replica follows at once.
	rec := h.postClientEvents(t, ownerSession, "0")
	wantStatus(t, rec, http.StatusSeeOther)
	if h.settings.events.sets.Load() != 1 || !h.settings.events.off.Load() {
		t.Fatalf("the core was set %d times and holds off = %v, want once and off", h.settings.events.sets.Load(), h.settings.events.off.Load())
	}
	if strings.Contains(login(), eventsScript) {
		t.Error("with client events off the login page still loads events.js")
	}
	owner := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, owner, http.StatusOK)
	wantNoBody(t, owner, eventsScript)
	// The endpoint answers 204 as before and keeps nothing: no log record, and nothing for a prober to tell it from on.
	wantStatus(t, h.serve(t, report(body)), http.StatusNoContent)
	if got := strings.Count(logs.String(), "client event"); got != 1 {
		t.Errorf("%d reports in the log after one was sent with client events off, want 1", got)
	}

	// And on again.
	wantStatus(t, h.postClientEvents(t, ownerSession, "1"), http.StatusSeeOther)
	if !strings.Contains(login(), eventsScript) {
		t.Error("after turning them on the login page does not load events.js")
	}
	wantStatus(t, h.serve(t, report(body)), http.StatusNoContent)
	if got := strings.Count(logs.String(), "client event"); got != 2 {
		t.Errorf("%d reports in the log after turning them on again, want 2", got)
	}
}

func TestClientEventsSettingIsTheOwnersAlone(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	for _, tt := range []struct {
		name    string
		session string
		want    int
	}{
		{"a read session", readSession, http.StatusForbidden},
		{"a viewer", viewerSession, http.StatusForbidden},
	} {
		wantStatus(t, h.postClientEvents(t, tt.session, "0"), tt.want)
	}
	rec := h.serve(t, form("/settings/client-events", "enabled=0"))
	wantStatus(t, rec, http.StatusSeeOther) // no session goes to the login page
	if got := rec.Header().Get("Location"); got != "/login" {
		t.Errorf("without a session the post went to %q, want /login", got)
	}
	if h.settings.events.sets.Load() != 0 || h.settings.events.off.Load() {
		t.Error("a session that is not the owner's changed the setting")
	}
}

func TestClientEventsSettingShowsInSettingsWithWhatIsCollected(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	page := func() *httptest.ResponseRecorder {
		return h.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	}
	const line = "The pages report a fixed list of events about themselves, such as the person check's states, script errors and failed requests: " +
		"nothing you type and no amounts. They stay on this install unless OpenTelemetry export is set."
	on := page()
	wantStatus(t, on, http.StatusOK)
	wantBody(t, on, `id="client-events-title"`, "Client events", `<span class="pill pill-ok">On</span>`, line,
		`action="/settings/client-events"`, `name="enabled" value="0"`, "Turn client events off")
	wantNoBody(t, on, "Turn client events on")

	wantStatus(t, h.postClientEvents(t, ownerSession, "0"), http.StatusSeeOther)
	off := page()
	wantBody(t, off, `<span class="pill pill-neutral">Off</span>`, `name="enabled" value="1"`, "Turn client events on", line)
	wantNoBody(t, off, "Turn client events off")

	// The install's config can only turn them off, and says so instead of offering a switch.
	forced := newHarnessWith(t, nil, func(c *Config) { c.DisableClientEvents = true })
	rec := forced.serve(t, withSession(request(http.MethodGet, "/settings", ""), ownerSession))
	wantBody(t, rec, "This install's UI configuration turns them off.", `<span class="pill pill-neutral">Off</span>`)
	wantNoBody(t, rec, `action="/settings/client-events"`)
}

func TestClientEventsSettingIsReadFromTheCoreOnceEveryInterval(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	h := newHarnessWith(t, nil, func(c *Config) { c.Now = clock.now })
	login := func() string { return h.serve(t, request(http.MethodGet, "/login", "")).Body.String() }
	for range 5 {
		login()
	}
	if got := h.settings.events.reads.Load(); got != 1 {
		t.Fatalf("five pages read the setting %d times, want once", got)
	}
	// Another replica's change, made in the core, reaches this one after the interval and not before.
	h.settings.events.off.Store(true)
	clock.advance(clientEventsRefresh - time.Second)
	if !strings.Contains(login(), eventsScript) {
		t.Error("the change showed before the interval was over")
	}
	clock.advance(2 * time.Second)
	if strings.Contains(login(), eventsScript) {
		t.Error("the change did not show after the interval")
	}
	if got := h.settings.events.reads.Load(); got != 2 {
		t.Errorf("the setting was read %d times, want twice", got)
	}
}

func TestClientEventsSettingWhenTheCoreCantBeAsked(t *testing.T) {
	t.Parallel()
	clock := newTestClock()
	h := newHarnessWith(t, nil, func(c *Config) { c.Now = clock.now })
	// With no answer ever, the UI collects nothing.
	h.settings.events.down.Store(true)
	if strings.Contains(h.serve(t, request(http.MethodGet, "/login", "")).Body.String(), eventsScript) {
		t.Error("with no answer from the core the page loads events.js")
	}
	// With an answer, it keeps the last one while the core is down.
	h.settings.events.down.Store(false)
	clock.advance(clientEventsRefresh)
	if !strings.Contains(h.serve(t, request(http.MethodGet, "/login", "")).Body.String(), eventsScript) {
		t.Fatal("with the core answering the page does not load events.js")
	}
	h.settings.events.down.Store(true)
	clock.advance(clientEventsRefresh)
	if !strings.Contains(h.serve(t, request(http.MethodGet, "/login", "")).Body.String(), eventsScript) {
		t.Error("with the core down the UI forgot its last answer")
	}
}
