package ui

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
)

// TestUIReadsAsTheSignedInPerson pins what D-163 and D-210 rest on: the UI
// reads the record with the signed-in person's session, never with its own
// plugin identity, so the core decides what each person sees. Every read a
// viewer's pages make carries the viewer's session token, none the plugin's
// token, and what a viewer sees is what the core lists for their session.
func TestUIReadsAsTheSignedInPerson(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.writeToken(t, pluginToken)
	for _, path := range []string{"/", "/accounts", "/sleeves", "/plugins", "/inbox"} {
		rec := h.serve(t, withSession(request(http.MethodGet, path, ""), viewerSession))
		wantStatus(t, rec, http.StatusOK)
	}
	bearers := seen(h.read, func() []string { return h.read.bearers })
	if len(bearers) < 5 {
		t.Fatalf("the viewer's pages made %d reads, want at least one a page", len(bearers))
	}
	for _, b := range bearers {
		if b != "Bearer "+viewerSession {
			t.Errorf("a read carried %q, want only the viewer's session token", b)
		}
	}
	rec := h.serve(t, withSession(request(http.MethodGet, "/accounts", ""), viewerSession))
	wantBody(t, rec, "Granted brokerage")
	wantNoBody(t, rec, "Synthetic brokerage")
	rec = h.serve(t, withSession(request(http.MethodGet, "/", ""), viewerSession))
	wantBody(t, rec, "Viewer access", "Alice")
	wantNoBody(t, rec, `href="/settings"`, `href="/access"`)
}

// postRoute is a route registered for POST, with the wrapper that guards it.
var postRoute = regexp.MustCompile(`HandleFunc\("POST (/[^"]*)",\s*([^)]*)`)

// ownerFreePosts returns the POST routes that take no owner session: the
// login, enrollment and sign-out ceremonies, which check what they need
// themselves, and a viewer signs in and out through them like anyone.
func ownerFreePosts() []string {
	return []string{
		"/login/passkey/begin", "/login/passkey/finish", "/login/demo", "/enroll/begin", "/enroll/finish",
		"/logout", "/passkey-check/begin", "/passkey-check/finish",
	}
}

// ownerPostRoutes returns the paths of the UI's POST routes guarded by
// withOwner, read from the source, and fails the test for a POST route that
// is neither guarded nor a sign-in or sign-out step.
func ownerPostRoutes(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("listing the UI's source: %v", err)
	}
	var routes []string
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(f) //nolint:gosec // the package's own source file, from a glob of its directory
		if readErr != nil {
			t.Fatalf("reading %s: %v", f, readErr)
		}
		for _, m := range postRoute.FindAllStringSubmatch(string(src), -1) {
			path, guard := m[1], m[2]
			switch {
			case slices.Contains(ownerFreePosts(), path):
			case strings.Contains(guard, "withOwner("):
				routes = append(routes, path)
			default:
				t.Errorf("POST %s is guarded by %q: a state-changing route takes withOwner, or is listed as a sign-in or sign-out step here", path, strings.TrimSpace(guard))
			}
		}
	}
	return routes
}

// TestViewerIsRefusedEveryChange checks that a viewer's session is refused,
// with 403, on every state-changing route of the UI: every POST route is read
// from the source, so a route added without the owner's guard fails here, and
// each one is posted to as the viewer (D-65: viewers change nothing but
// sign-out).
func TestViewerIsRefusedEveryChange(t *testing.T) {
	t.Parallel()
	routes := ownerPostRoutes(t)
	if len(routes) < 20 {
		t.Fatalf("found %d guarded POST routes, want the UI's twenty or more: the scan broke", len(routes))
	}
	h := newHarness(t)
	h.writeToken(t, pluginToken)
	placeholder := regexp.MustCompile(`\{[a-z]+\}`)
	for _, route := range routes {
		target := placeholder.ReplaceAllString(route, "x")
		rec := h.serve(t, withSession(form(target, "name=x"), viewerSession))
		if rec.Code != http.StatusForbidden {
			t.Errorf("POST %s as a viewer = %d, want 403", target, rec.Code)
		}
	}
	// None of those reached a service that changes something.
	if n := len(seen(h.sleeves, func() []string { return h.sleeves.calls })); n != 0 {
		t.Errorf("a viewer's forms reached the sleeve service %d times", n)
	}
}

// TestAskBoxOfAViewerSeesOnlyTheirGrantedAccounts checks D-163 through the Ask
// box, which runs its answerers as the asking session: the fixed commands and
// the model path's tools both read with the viewer's session token, name only
// the account the viewer was granted, and answer a question about another
// account as one about something that is not there. What a model is sent is
// what those reads returned, so it never holds the other account's name.
func TestAskBoxOfAViewerSeesOnlyTheirGrantedAccounts(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.askModel.turnOn("T0", "T2")

	// The fixed commands: the granted account is found, the other is not.
	rec := h.asks(t, viewerSession, "show Granted brokerage")
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Granted brokerage")
	rec = h.asks(t, viewerSession, "show Synthetic brokerage")
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, `"Synthetic brokerage" account`, "acc-synthetic-1")
	if strings.Contains(rec.Body.String(), `class="used-card"`) {
		t.Error("a command's answer drew a card for an account the viewer was not granted")
	}

	// The model path: the tool names the other account, finds nothing, and what
	// the model is sent never holds its name.
	h.askModel.say(
		scriptedReply{calls: []*modelsv1.ToolCall{toolCall("show", `{"name":"Synthetic brokerage","kind":"account"}`)}},
		scriptedReply{text: "I found no such account."},
	)
	rec = h.asks(t, viewerSession, "tell me about the synthetic brokerage in full")
	wantStatus(t, rec, http.StatusOK)
	for _, req := range h.askModel.requests {
		for _, m := range req {
			if strings.Contains(m.GetMessage().GetContent(), `"label":"Synthetic brokerage"`) && m.GetMessage().GetRole() == modelsv1.MessageRole_MESSAGE_ROLE_TOOL {
				t.Errorf("the model was sent the other account's name in a tool result: %s", m.GetMessage().GetContent())
			}
		}
	}
	wantAll(t, "a call to the core carried session", h.askModel.sessions, viewerSession)
	reads := seen(h.read, func() []string { return h.read.bearers })
	if len(reads) == 0 {
		t.Fatal("the Ask box read nothing")
	}
	wantAll(t, "an Ask box read carried", reads, "Bearer "+viewerSession)
}
