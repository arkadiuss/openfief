//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// errSettingsPost reports a post of Settings' form that did not send the browser back to Settings.
var errSettingsPost = errors.New("the settings form was not accepted")

const (
	// eventsScriptTag is what the UI's layout loads while the client events are on.
	eventsScriptTag = `src="/static/events.js"`
	// Distinct times for the reports the test posts, so that each is found in the UI's log
	// on its own: the first is taken, the second is sent while the owner has the events off,
	// the third once they are on again.
	eventsMsOn, eventsMsOff, eventsMsBack = 4242, 4343, 4444
)

// TestClientEventsThroughTheFrontDoor checks D-227's client events on the real install
// (SCEN-e2e-ui-client-events): a page's report posted to the core's browser port, which
// proxies it to the built-in UI, is taken, with no session, and becomes a log record of
// the UI; the owner turns the events off in Settings' own setting, which the core stores
// and the UI reads, and the page then loads no reporting script and a report is answered
// 204 and dropped, so it is no log record; turned on again, reports are taken again.
//
// It changes a setting the whole install reads, so it puts it back on when it ends; no
// other test posts a report or reads the setting.
func TestClientEventsThroughTheFrontDoor(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	waitPluginReady(t, k, uiPlugin)

	base := forwardCore(t, k)
	owner := newOwnerAPI(t, k)
	set := func(ctx context.Context, enabled bool) error {
		return setClientEvents(ctx, base, owner.session, enabled)
	}
	// The setting is on until the owner turns it off; whatever happens, leave it on, in a context of its own
	// since a Cleanup runs after the test's is canceled.
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), requestTimeout)
		defer cancel()
		if err := set(ctx, true); err != nil {
			t.Errorf("putting the client events back on: %v", err)
		}
	})

	// On by default: the login page loads the reporting script, and a report is taken.
	waitLoginLoadsEvents(t, base, true)
	if r := postClientEvent(t, base, eventsMsOn); r.status != http.StatusNoContent {
		t.Fatalf("POST /client-events with the events on = %d %s, want 204", r.status, r.body)
	}
	waitClientEventLogged(t, k, eventsMsOn)

	// The owner turns them off: the page stops loading the script, the settings page says so, and a report is
	// answered 204 and dropped.
	if err := set(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	waitLoginLoadsEvents(t, base, false)
	settingsPage := uiPage(t, base, "/settings", owner.session)
	for _, want := range []string{"Client events", "Turn client events on", "nothing you type and no amounts"} {
		if !strings.Contains(settingsPage, want) {
			t.Errorf("the settings page with the events off lacks %q", want)
		}
	}
	if r := postClientEvent(t, base, eventsMsOff); r.status != http.StatusNoContent {
		t.Fatalf("POST /client-events with the events off = %d %s, want 204 and nothing kept", r.status, r.body)
	}

	// On again: reports are taken once more, and the one sent while they were off never was.
	if err := set(t.Context(), true); err != nil {
		t.Fatal(err)
	}
	waitLoginLoadsEvents(t, base, true)
	if r := postClientEvent(t, base, eventsMsBack); r.status != http.StatusNoContent {
		t.Fatalf("POST /client-events after turning them on again = %d %s, want 204", r.status, r.body)
	}
	waitClientEventLogged(t, k, eventsMsBack)
	if clientEventLogged(t.Context(), t, k, eventsMsOff) {
		t.Errorf("the report sent while the owner had the client events off (ms %d) is in the UI's log", eventsMsOff)
	}
}

// setClientEvents posts Settings' form that turns the client events on or off to the core's browser port, as
// the owner's browser does: with the owner's session cookie, on the allowed origin, marked same-origin. The
// UI answers by sending the browser back to Settings.
func setClientEvents(ctx context.Context, base, session string, enabled bool) error {
	value := "0"
	if enabled {
		value = "1"
	}
	body := url.Values{"enabled": {value}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/settings/client-events", strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST /settings/client-events: %w", err)
	}
	req.Host = enrollHost
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.AddCookie(&http.Cookie{Name: "__Host-openfief_session", Value: session, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }, Timeout: 2 * requestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("POST /settings/client-events: %w", err)
	}
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, readAPIBodyLimit)); err != nil {
		return fmt.Errorf("POST /settings/client-events: reading the answer: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("POST /settings/client-events: %w", err)
	}
	if resp.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("%w: POST /settings/client-events = %d, want 303 back to Settings", errSettingsPost, resp.StatusCode)
	}
	return nil
}

// postClientEvent posts one report of a Turnstile success taking ms milliseconds to the core's browser
// port, as a page on the allowed origin does: with no session, as the login page reports too, and marked
// same-origin as a browser marks a fetch from its own page.
func postClientEvent(t *testing.T, base string, ms int) reply {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), requestTimeout)
	defer cancel()
	body := fmt.Sprintf(`{"events":[{"k":"turnstile-success","ms":%d,"path":"invisible"}]}`, ms)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/client-events", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /client-events: %v", err)
	}
	req.Host = enrollHost
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	return doReply(t, req)
}

// waitLoginLoadsEvents waits until the login page, which the UI shows without a session, loads the reporting
// script when on is true and does not when it is false. The UI keeps the core's answer for a few seconds, so a
// change the owner made reaches it after that.
func waitLoginLoadsEvents(t *testing.T, base string, on bool) {
	t.Helper()
	err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/login", http.NoBody)
		if err != nil {
			return "", false, fmt.Errorf("GET /login: %w", err)
		}
		req.Host = enrollHost
		r := doReply(t, req)
		loads := bytes.Contains(r.body, []byte(eventsScriptTag))
		return fmt.Sprintf("the login page (%d) loads the reporting script: %v", r.status, loads), r.status == http.StatusOK && loads == on, nil
	})
	if err != nil {
		t.Fatalf("waiting for the login page to load the reporting script = %v: %v", on, err)
	}
}

// uiLog returns the built-in UI's log.
func uiLog(ctx context.Context, t *testing.T, k *kube) []byte {
	t.Helper()
	pod, err := k.runningPluginPod(ctx, uiPlugin)
	if err != nil {
		t.Fatalf("finding the UI's pod: %v", err)
	}
	out, err := k.clients.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{}).DoRaw(ctx)
	if err != nil {
		t.Fatalf("reading the UI's log: %v", err)
	}
	return out
}

// clientEventLogged reports whether the UI's log holds the record of a client event that took ms
// milliseconds: one line with the message and the number.
func clientEventLogged(ctx context.Context, t *testing.T, k *kube, ms int) bool {
	t.Helper()
	want := fmt.Sprintf(`"ms":%d`, ms)
	for line := range bytes.Lines(uiLog(ctx, t, k)) {
		if bytes.Contains(line, []byte(`"msg":"client event"`)) && bytes.Contains(line, []byte(want)) {
			return true
		}
	}
	return false
}

// waitClientEventLogged waits until the UI's log holds the record of the client event that took ms.
func waitClientEventLogged(t *testing.T, k *kube, ms int) {
	t.Helper()
	err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
		return fmt.Sprintf("no record of the event of %d ms yet", ms), clientEventLogged(ctx, t, k, ms), nil
	})
	if err != nil {
		t.Fatalf("waiting for the UI to log the client event: %v", err)
	}
}
