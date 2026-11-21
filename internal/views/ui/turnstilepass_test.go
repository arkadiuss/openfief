package ui

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

// The tests of the Turnstile pass (D-208): a visitor who passed the check isn't
// checked again for 30 minutes after their last model question, and at most 2
// hours from the check, through a signed cookie tied to their session that holds
// nothing else; the limits and the budget still count every question.

const passQuestion = "how diversified am I?"

// passHarness is turnstileHarness with a clock the test moves.
func passHarness(t *testing.T) (*harness, *fakeSiteverify, *strings.Builder, *testClock) {
	t.Helper()
	return passHarnessWith(t, nil)
}

// passHarnessWith is passHarness with edit applied to the Turnstile config.
func passHarnessWith(t *testing.T, edit func(*TurnstileConfig)) (*harness, *fakeSiteverify, *strings.Builder, *testClock) {
	t.Helper()
	clock := newTestClock()
	f := newFakeSiteverify(t)
	f.hostname = testHost
	logs := &strings.Builder{}
	h := newHarnessWith(t, slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})), func(c *Config) {
		c.Turnstile = f.config()
		if edit != nil {
			edit(c.Turnstile)
		}
		c.Now = clock.now
	})
	h.askModel.turnOn("T0", "T2")
	return h, f, logs, clock
}

// asksWithPass sends a model question with no token and the pass cookie, if any,
// as the page's first request does.
func (h *harness) asksWithPass(t *testing.T, session string, pass *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := htmx(withSession(form("/ask", url.Values{"q": {passQuestion}}.Encode()), session))
	r.Header.Set(openfief.ClientIPHeader, "198.51.100.7")
	if pass != nil {
		r.AddCookie(pass)
	}
	return h.serve(t, r)
}

// passOf returns the pass cookie a response sets, or nil.
func passOf(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == checkedCookieName {
			return c
		}
	}
	return nil
}

// wantAnswered fails t unless rec is the model's answer, and wantChecked unless it
// is the card that sends the page to run the check, with the HX-Trigger event.
func wantAnswered(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Answered by m")
	if rec.Header().Get(hxTrigger) != "" {
		t.Errorf("an answered question carries HX-Trigger %q", rec.Header().Get(hxTrigger))
	}
}

func wantChecked(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "A check comes before the model answers")
	wantNoBody(t, rec, "Answered by")
	if got := rec.Header().Get(hxTrigger); got != checkNeededTrigger {
		t.Errorf("HX-Trigger = %q, want %q", got, checkNeededTrigger)
	}
}

// passSpan is when a pass's check was made and when the pass ends.
type passSpan struct{ checked, expires time.Time }

// passTimes reads when the check was made and when the pass ends, from a cookie.
func passTimes(t *testing.T, c *http.Cookie) passSpan {
	t.Helper()
	parts := strings.Split(c.Value, ".")
	if len(parts) != passFields {
		t.Fatalf("the pass %q has %d parts, want %d", c.Value, len(parts), passFields)
	}
	a, errA := strconv.ParseInt(parts[1], 10, 64)
	b, errB := strconv.ParseInt(parts[2], 10, 64)
	if errA != nil || errB != nil {
		t.Fatalf("the pass %q has times that are no numbers", c.Value)
	}
	return passSpan{checked: time.Unix(a, 0), expires: time.Unix(b, 0)}
}

// askCookie is a pass cookie with value, as a browser sends it back: only name and value matter to the server, but the
// flags are what a real one has.
func askCookie(value string) *http.Cookie {
	return &http.Cookie{Name: checkedCookieName, Value: value, Path: passCookiePath, Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}
}

func TestPassIsIssuedWithTheCheckAndHoldsNothingElse(t *testing.T) {
	t.Parallel()
	h, _, _, clock := passHarness(t)
	start := clock.now()
	rec := h.asksWithToken(t, readSession, passQuestion, "ok-1")
	wantAnswered(t, rec)
	pass := passOf(rec)
	if pass == nil {
		t.Fatal("a question that passed the check set no pass")
	}
	checkPassFlags(t, pass)
	span := passTimes(t, pass)
	if !span.checked.Equal(start) || !span.expires.Equal(start.Add(defaultPassIdle)) || pass.MaxAge != int(defaultPassIdle/time.Second) {
		t.Errorf("the pass was made at %v and ends %v (max-age %d), want %v and %v (%d)",
			span.checked, span.expires, pass.MaxAge, start, start.Add(defaultPassIdle), int(defaultPassIdle/time.Second))
	}
	// It names no one: neither the session, nor its token or ID, nor the token the widget made, nor the address, nor the question.
	for _, secret := range []string{"readSession", readSession, "s-" + readSession, "ok-1", "198.51.100.7", "diversified", secretForTests} {
		if strings.Contains(pass.Value, secret) {
			t.Errorf("the pass %q holds %q", pass.Value, secret)
		}
	}
	if strings.Count(rec.Header().Get("Set-Cookie"), "openfief_ask_pass") != 1 {
		t.Errorf("Set-Cookie = %q, want the one pass", rec.Header().Get("Set-Cookie"))
	}
}

// checkPassFlags fails t unless the pass cookie has the flags D-208's pass has: HttpOnly, Secure, SameSite=Strict, a path
// scoped to the Ask box's own request, and the __Secure- prefix that makes a browser insist on Secure.
func checkPassFlags(t *testing.T, pass *http.Cookie) {
	t.Helper()
	if !pass.HttpOnly || !pass.Secure || pass.SameSite != http.SameSiteStrictMode || pass.Path != passCookiePath {
		t.Errorf("the pass has HttpOnly=%v Secure=%v SameSite=%v Path=%q, want HttpOnly, Secure, Strict and %q",
			pass.HttpOnly, pass.Secure, pass.SameSite, pass.Path, passCookiePath)
	}
	if !strings.HasPrefix(pass.Name, "__Secure-") {
		t.Errorf("the pass is named %q, want the __Secure- prefix that makes browsers insist on Secure", pass.Name)
	}
}

func TestAValidPassSkipsSiteverifyAndTheNextQuestionIsStillCounted(t *testing.T) {
	t.Parallel()
	h, f, _, _ := passHarness(t)
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	if n := len(f.calls()); n != 1 {
		t.Fatalf("Cloudflare was asked %d times for the first question, want once", n)
	}
	generated, admitted := h.askModel.generated(), len(h.askModel.admitted)
	for i := range 3 {
		wantAnswered(t, h.asksWithPass(t, readSession, pass))
		if n := len(f.calls()); n != 1 {
			t.Fatalf("after question %d with the pass Cloudflare was asked %d times, want it not asked again", i+2, n)
		}
	}
	if got := h.askModel.generated(); got != generated+3 {
		t.Errorf("the model answered %d more times, want 3", got-generated)
	}
	// The pass skips the check and never the core's admission: each question counted once, with the address.
	if got := len(h.askModel.admitted); got != admitted+3 {
		t.Errorf("the core counted %d more questions, want 3", got-admitted)
	}
	for _, a := range h.askModel.admitted {
		if a != "198.51.100.7" && a != "" {
			t.Errorf("the core was told the address %q", a)
		}
	}
}

func TestLimitsAndTheBudgetStillStopAQuestionWithAPass(t *testing.T) {
	t.Parallel()
	h, f, _, _ := passHarness(t)
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	for _, tt := range []struct {
		name      string
		admission modelsv1.AskAdmission
		want      string
	}{
		{"the session's limit", modelsv1.AskAdmission_ASK_ADMISSION_SESSION_LIMIT, "You&#39;ve asked the model enough for now"},
		{"the address's limit", modelsv1.AskAdmission_ASK_ADMISSION_ADDRESS_LIMIT, "Too many questions from your network"},
		{"the day's budget", modelsv1.AskAdmission_ASK_ADMISSION_BUDGET_SPENT, "budget"},
	} {
		h.askModel.admission = tt.admission
		before := h.askModel.generated()
		admitted := len(h.askModel.admitted)
		rec := h.asksWithPass(t, readSession, pass)
		wantStatus(t, rec, http.StatusOK)
		if !strings.Contains(strings.ToLower(rec.Body.String()), strings.ToLower(tt.want)) {
			t.Errorf("%s: the answer lacks %q:\n%s", tt.name, tt.want, rec.Body.String())
		}
		if h.askModel.generated() != before {
			t.Errorf("%s: the model was called for a question the core refused", tt.name)
		}
		if got := len(h.askModel.admitted); got != admitted+1 {
			t.Errorf("%s: the core was asked %d times, want once: the pass skips the check, not the core", tt.name, got-admitted)
		}
		// A question the core refuses does not reach the model, so it does not renew the pass.
		if passOf(rec) != nil {
			t.Errorf("%s: a refused question renewed the pass", tt.name)
		}
	}
	if n := len(f.calls()); n != 1 {
		t.Errorf("Cloudflare was asked %d times, want only for the first question", n)
	}
}

func TestPassSlidesWithEachModelQuestionUpToTwoHoursFromTheCheck(t *testing.T) {
	t.Parallel()
	h, f, _, clock := passHarness(t)
	start := clock.now()
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	pass = askEvery20Minutes(t, h, clock, start, pass)
	// 100 minutes in, the window would reach 130; it is cut at 2 hours from the check.
	if span := passTimes(t, pass); !span.expires.Equal(start.Add(defaultPassMax)) {
		t.Fatalf("the pass ends %v, want the cap at %v", span.expires, start.Add(defaultPassMax))
	}
	// Another question at 119 minutes still passes and still ends at the cap.
	clock.advance(19 * time.Minute)
	rec := h.asksWithPass(t, readSession, pass)
	wantAnswered(t, rec)
	if span := passTimes(t, passOf(rec)); !span.expires.Equal(start.Add(defaultPassMax)) {
		t.Errorf("at 119 minutes the pass ends %v, want it still capped at %v", span.expires, start.Add(defaultPassMax))
	}
	// Past the cap the same pass is refused, although the person asked a minute ago.
	clock.advance(2 * time.Minute)
	wantChecked(t, h.asksWithPass(t, readSession, passOf(rec)))
	if n := len(f.calls()); n != 1 {
		t.Errorf("Cloudflare was asked %d times, want once", n)
	}
	// And the check makes a new pass, that starts the windows again.
	again := h.asksWithToken(t, readSession, passQuestion, "ok-2")
	wantAnswered(t, again)
	span := passTimes(t, passOf(again))
	if !span.checked.Equal(clock.now()) || !span.expires.Equal(clock.now().Add(defaultPassIdle)) {
		t.Errorf("the new pass was made at %v and ends %v, want %v and %v", span.checked, span.expires, clock.now(), clock.now().Add(defaultPassIdle))
	}
}

// askEvery20Minutes asks a model question every 20 minutes, five times, with the pass the last answer set. Each renews
// the 30 minutes: the pass ends 30 minutes after the last question, never later than 2 hours from the check at start.
// It returns the last pass.
func askEvery20Minutes(t *testing.T, h *harness, clock *testClock, start time.Time, pass *http.Cookie) *http.Cookie {
	t.Helper()
	for _, minutes := range []int{20, 40, 60, 80, 100} {
		clock.advance(20 * time.Minute)
		rec := h.asksWithPass(t, readSession, pass)
		wantAnswered(t, rec)
		renewed := passOf(rec)
		if renewed == nil {
			t.Fatalf("at %d minutes the pass was not renewed", minutes)
		}
		span := passTimes(t, renewed)
		want := start.Add(time.Duration(minutes)*time.Minute + defaultPassIdle)
		if limit := start.Add(defaultPassMax); want.After(limit) {
			want = limit
		}
		if !span.checked.Equal(start) {
			t.Errorf("at %d minutes the renewed pass says the check was at %v, want the original %v", minutes, span.checked, start)
		}
		if !span.expires.Equal(want) {
			t.Errorf("at %d minutes the renewed pass ends %v, want %v", minutes, span.expires, want)
		}
		pass = renewed
	}
	return pass
}

func TestAnExpiredPassIsRefusedAndTheQuestionIsChecked(t *testing.T) {
	t.Parallel()
	h, f, _, clock := passHarness(t)
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	clock.advance(defaultPassIdle - time.Second)
	wantAnswered(t, h.asksWithPass(t, readSession, pass))
	clock.advance(defaultPassIdle + time.Second) // 30 minutes after the pass was made: idle for longer than the window
	generated := h.askModel.generated()
	wantChecked(t, h.asksWithPass(t, readSession, pass))
	if h.askModel.generated() != generated {
		t.Error("an expired pass let a question through to the model")
	}
	// The fallback is the on-demand check: the page asks again with a token and is answered.
	wantAnswered(t, h.asksWithToken(t, readSession, passQuestion, "ok-2"))
	if n := len(f.calls()); n != 2 {
		t.Errorf("Cloudflare was asked %d times, want 2: the first question and the one after the pass expired", n)
	}
}

func TestAPassFromAnotherSessionIsRefused(t *testing.T) {
	t.Parallel()
	h, f, _, _ := passHarness(t)
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	// Another visitor's session, with the first one's cookie.
	generated := h.askModel.generated()
	wantChecked(t, h.asksWithPass(t, demoSession, pass))
	wantChecked(t, h.asksWithPass(t, ownerSession, pass))
	if h.askModel.generated() != generated {
		t.Error("a pass made for one session let another's question through")
	}
	// The session it was made for still has it.
	wantAnswered(t, h.asksWithPass(t, readSession, pass))
	if n := len(f.calls()); n != 1 {
		t.Errorf("Cloudflare was asked %d times, want once", n)
	}
}

func TestATamperedOrMalformedPassIsRefusedWithoutAnError(t *testing.T) {
	t.Parallel()
	h, f, _, clock := passHarness(t)
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	parts := strings.Split(pass.Value, ".")
	span := passTimes(t, pass)
	other := func(c byte) string { // a character other than c, from the base64 alphabet
		if c == 'A' {
			return "B"
		}
		return "A"
	}
	mac := parts[3]
	later := strconv.FormatInt(span.expires.Add(time.Hour).Unix(), 10)
	earlier := strconv.FormatInt(span.checked.Add(-time.Hour).Unix(), 10)
	forgedTime := strconv.FormatInt(clock.now().Add(2*passSkew).Unix(), 10)
	tests := []struct{ name, value string }{
		{"a changed MAC", strings.Join([]string{parts[0], parts[1], parts[2], other(mac[0]) + mac[1:]}, ".")},
		{"a truncated MAC", strings.Join([]string{parts[0], parts[1], parts[2], mac[:len(mac)-2]}, ".")},
		{"a later end with the old MAC", strings.Join([]string{parts[0], parts[1], later, mac}, ".")},
		{"an earlier check with the old MAC", strings.Join([]string{parts[0], earlier, parts[2], mac}, ".")},
		{"another version", strings.Join([]string{"v2", parts[1], parts[2], mac}, ".")},
		{"an extra part", pass.Value + ".x"},
		{"a missing part", strings.Join(parts[:3], ".")},
		{"times that are no numbers", strings.Join([]string{parts[0], "abc", "def", mac}, ".")},
		{"a MAC that is no base64", strings.Join([]string{parts[0], parts[1], parts[2], "!!!"}, ".")},
		{"a check made in the future", strings.Join([]string{parts[0], forgedTime, later, mac}, ".")},
		{"nothing", ""},
		{"garbage", "not-a-pass"},
		{"dots", "...."},
		{"far too long", strings.Repeat("a", 5000)},
		{"the session's own token", readSession},
	}
	generated := h.askModel.generated()
	for _, tt := range tests {
		wantChecked(t, h.asksWithPass(t, readSession, askCookie(tt.value)))
		if n := len(f.calls()); n != 1 {
			t.Fatalf("%s: Cloudflare was asked %d times, want only for the first question", tt.name, n)
		}
	}
	if h.askModel.generated() != generated {
		t.Error("a pass that does not verify let a question through to the model")
	}
}

// TestAPassMadeUnderAnotherSecretIsRefused: the pass's key comes from the widget's
// secret, so a rotated secret ends every pass and the questions are checked again.
func TestAPassMadeUnderAnotherSecretIsRefused(t *testing.T) {
	t.Parallel()
	var secret atomic.Value
	secret.Store(secretForTests)
	h, f, _, _ := passHarnessWith(t, func(c *TurnstileConfig) {
		c.Secret = func() (string, error) {
			s, ok := secret.Load().(string)
			if !ok {
				return "", errBoomTurnstile
			}
			return s, nil
		}
	})
	pass := passOf(h.asksWithToken(t, readSession, passQuestion, "ok-1"))
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	wantAnswered(t, h.asksWithPass(t, readSession, pass))
	secret.Store("ANOTHER-SECRET-FOR-TESTS")
	wantChecked(t, h.asksWithPass(t, readSession, pass))
	if n := len(f.calls()); n != 1 {
		t.Errorf("Cloudflare was asked %d times, want once", n)
	}
}

func TestACommandNeverNeedsAPass(t *testing.T) {
	t.Parallel()
	h, f, _, _ := passHarness(t)
	rec := h.serve(t, htmx(withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {"allocation by currency"}}.Encode(), ""), readSession)))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "A check comes before the model answers")
	if rec.Header().Get(hxTrigger) != "" || passOf(rec) != nil || len(f.calls()) != 0 {
		t.Errorf("a command was checked (HX-Trigger %q, pass %v, %d calls to Cloudflare)", rec.Header().Get(hxTrigger), passOf(rec), len(f.calls()))
	}
}

func TestPassIsNotRenewedByAQuestionTheCheckStops(t *testing.T) {
	t.Parallel()
	h, f, _, _ := passHarness(t)
	// A token Cloudflare refuses makes no pass, and the response sets no cookie.
	rec := h.asksWithToken(t, readSession, passQuestion, "forged")
	wantStatus(t, rec, http.StatusOK)
	if passOf(rec) != nil || len(f.calls()) != 1 {
		t.Errorf("a refused token earned a pass (%v) or was not asked of Cloudflare once (%d)", passOf(rec), len(f.calls()))
	}
	// A question with no token and no pass sets none either.
	if passOf(h.asksWithPass(t, readSession, nil)) != nil {
		t.Error("a question with no token and no pass was given one")
	}
}

func TestPassNeverReachesTheLog(t *testing.T) {
	t.Parallel()
	h, _, logs, clock := passHarness(t)
	rec := h.asksWithToken(t, readSession, passQuestion, "ok-1")
	pass := passOf(rec)
	if pass == nil {
		t.Fatal("no pass was issued")
	}
	clock.advance(time.Minute)
	pass = passOf(h.asksWithPass(t, readSession, pass)) // used and renewed
	tampered := askCookie(pass.Value[:len(pass.Value)-3] + "AAA")
	h.asksWithPass(t, readSession, tampered) // refused
	h.asksWithPass(t, demoSession, pass)     // another session's
	clock.advance(3 * time.Hour)
	h.asksWithPass(t, readSession, pass) // expired
	mac := strings.Split(pass.Value, ".")[3]
	for _, secret := range []string{pass.Value, mac, tampered.Value, "s-" + readSession, "s-" + demoSession, checkedCookieName, secretForTests, "ok-1"} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("the log holds %q:\n%s", secret, logs)
		}
	}
	// What is logged of a question that used the pass is the fixed words of the check, like any verdict.
	if !strings.Contains(logs.String(), "verdict=pass path=none") {
		t.Errorf("the log does not say a question went on with a pass:\n%s", logs)
	}
}

func TestPassIssueAndUseByUnits(t *testing.T) {
	t.Parallel()
	check, clock := unitCheck(t)
	checked := clock.now()
	pass := check.issue("session-1", checked)
	if pass == nil {
		t.Fatal("no pass for a session with an ID")
	}
	if at, ok := usePassOf(check, "session-1", pass); !ok || !at.Equal(checked) {
		t.Errorf("use = %v, %v, want the check time %v", at, ok, checked)
	}
	if _, ok := usePassOf(check, "session-2", pass); ok {
		t.Error("another session used the pass")
	}
	if _, ok := usePassOf(check, "", pass); ok {
		t.Error("a session with no ID used a pass")
	}
	if _, ok := usePassOf(check, "session-1", nil); ok {
		t.Error("no cookie was a pass")
	}
	if check.issue("", checked) != nil {
		t.Error("a pass was issued for a session with no ID")
	}
	// A check older than the cap earns nothing.
	if check.issue("session-1", checked.Add(-defaultPassMax)) != nil {
		t.Error("a pass was issued for a check made 2 hours ago")
	}
}

func TestPassNeedsTheWidgetsSecretToIssueAndToUse(t *testing.T) {
	t.Parallel()
	check, clock := unitCheck(t)
	pass := check.issue("session-1", clock.now())
	if pass == nil {
		t.Fatal("no pass for a session with an ID")
	}
	check.cfg.Secret = func() (string, error) { return "", errBoomTurnstile }
	if check.issue("session-1", clock.now()) != nil {
		t.Error("a pass was issued with no secret to sign with")
	}
	if _, ok := usePassOf(check, "session-1", pass); ok {
		t.Error("a pass was accepted with no secret to check it with")
	}
}

// unitCheck is a check over a fake siteverify with a clock the test moves.
func unitCheck(t *testing.T) (*turnstileCheck, *testClock) {
	t.Helper()
	clock := newTestClock()
	check := newFakeSiteverify(t).check(slog.New(slog.DiscardHandler))
	check.now = clock.now
	return check, clock
}

// usePassOf asks check whether a request for session id with cookie c, if any, holds a valid pass.
func usePassOf(check *turnstileCheck, id string, c *http.Cookie) (time.Time, bool) {
	r := request(http.MethodGet, "/ask", "")
	if c != nil {
		r.AddCookie(c)
	}
	return check.usePass(r, id)
}

func TestTurnstileConfigRefusesNegativePassWindows(t *testing.T) {
	t.Parallel()
	f := newFakeSiteverify(t)
	for _, edit := range []func(*TurnstileConfig){
		func(c *TurnstileConfig) { c.PassIdle = -time.Second },
		func(c *TurnstileConfig) { c.PassMax = -time.Second },
	} {
		c := *f.config()
		edit(&c)
		if err := c.validate(); err == nil {
			t.Errorf("validate accepted %+v", c)
		}
	}
	c := f.config()
	if err := c.validate(); err != nil {
		t.Fatalf("validate refused the default windows: %v", err)
	}
	if w := c.passWindows(); w.idle != 30*time.Minute || w.limit != 2*time.Hour {
		t.Errorf("the default windows are %v and %v, want D-208's 30 minutes and 2 hours", w.idle, w.limit)
	}
}
