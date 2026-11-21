package ui

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"

	modelsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/models/v1"
	"github.com/arkadiuss/openfief/sdk/go/openfief"
)

const (
	// siteKeyForTests and secretForTests stand for Turnstile's keys: the first is
	// public, the second must reach Cloudflare and nothing else.
	siteKeyForTests = "1x00000000000000000000AA"
	secretForTests  = "SECRET-TURNSTILE-KEY-FOR-TESTS"
)

// fakeSiteverify stands for Cloudflare's siteverify. A token that starts with
// "ok-" is good once; any other is refused as invalid, a used one as a
// duplicate, and a call with another secret as one with a wrong secret. It
// keeps every form it was sent.
type fakeSiteverify struct {
	mu       sync.Mutex
	forms    []url.Values
	used     map[string]bool
	hostname string
	// status and body, when set, are what it answers instead.
	status int
	body   string
	srv    *httptest.Server
}

func newFakeSiteverify(t *testing.T) *fakeSiteverify {
	t.Helper()
	f := &fakeSiteverify{used: map[string]bool{}}
	f.srv = httptest.NewTLSServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeSiteverify) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		http.Error(w, "not a siteverify request", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	f.forms = append(f.forms, r.PostForm)
	if f.status != 0 {
		w.WriteHeader(f.status)
	}
	if f.body != "" {
		writeString(w, f.body)
		return
	}
	token := r.PostForm.Get("response")
	switch {
	case r.PostForm.Get("secret") != secretForTests:
		writeString(w, `{"success":false,"error-codes":["invalid-input-secret"]}`)
	case f.used[token]:
		writeString(w, `{"success":false,"error-codes":["timeout-or-duplicate"]}`)
	case !strings.HasPrefix(token, "ok-"):
		writeString(w, `{"success":false,"error-codes":["invalid-input-response"]}`)
	default:
		f.used[token] = true
		writeString(w, `{"success":true,"hostname":"`+f.hostname+`","error-codes":[]}`)
	}
}

func writeString(w http.ResponseWriter, s string) {
	if _, err := w.Write([]byte(s)); err != nil {
		return // the client went away
	}
}

func (f *fakeSiteverify) calls() []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.forms...)
}

// config is the Turnstile config that calls this fake with the test secret.
func (f *fakeSiteverify) config() *TurnstileConfig {
	return &TurnstileConfig{
		SiteKey: siteKeyForTests, VerifyURL: f.srv.URL, HTTP: f.srv.Client(),
		Secret: func() (string, error) { return secretForTests, nil },
	}
}

func (f *fakeSiteverify) check(log *slog.Logger) *turnstileCheck {
	return &turnstileCheck{cfg: *f.config(), log: log}
}

func TestTurnstileVerifyByVerdict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		token    string
		remote   string
		host     string
		hostname string
		status   int
		body     string
		secret   func() (string, error)
		want     turnstileVerdict
		calls    int
	}{
		{name: "a good token", token: "ok-1", remote: "198.51.100.7", host: testHost, want: turnstilePassed, calls: 1},
		{name: "a good token made on the host that asks, with its port", token: "ok-1", host: "localhost:8080", hostname: "localhost", want: turnstilePassed, calls: 1},
		{name: "a good token made on another host", token: "ok-1", host: testHost, hostname: "evil.example.com", want: turnstileWrongHost, calls: 1},
		{name: "a good token, the host in another case", token: "ok-1", host: "Ui.Example.com", hostname: "ui.example.com", want: turnstilePassed, calls: 1},
		{name: "an invalid token", token: "forged", host: testHost, want: turnstileRejected, calls: 1},
		{name: "no token", token: "", host: testHost, want: turnstileNoToken, calls: 0},
		{name: "a token longer than Cloudflare issues", token: strings.Repeat("a", maxTurnstileToken+1), host: testHost, want: turnstileNoToken, calls: 0},
		{name: "no secret to ask with", token: "ok-1", host: testHost, secret: func() (string, error) { return "", nil }, want: turnstileUnavailable, calls: 0},
		{name: "a secret that can't be read", token: "ok-1", host: testHost, secret: func() (string, error) { return "", errBoomTurnstile }, want: turnstileUnavailable, calls: 0},
		{name: "Cloudflare answers an error", token: "ok-1", host: testHost, status: http.StatusInternalServerError, body: "upstream trouble", want: turnstileUnavailable, calls: 1},
		{name: "Cloudflare answers no JSON", token: "ok-1", host: testHost, body: "<html>blocked</html>", want: turnstileUnavailable, calls: 1},
		{name: "Cloudflare refuses the secret", token: "ok-1", host: testHost, body: `{"success":false,"error-codes":["invalid-input-secret"]}`, want: turnstileUnavailable, calls: 1},
		{name: "Cloudflare finds no secret", token: "ok-1", host: testHost, body: `{"success":false,"error-codes":["missing-input-secret"]}`, want: turnstileUnavailable, calls: 1},
		{name: "a token Cloudflare calls a duplicate", token: "ok-1", host: testHost, body: `{"success":false,"error-codes":["timeout-or-duplicate"]}`, want: turnstileRejected, calls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeSiteverify(t)
			f.hostname, f.status, f.body = tt.hostname, tt.status, tt.body
			check := f.check(slog.New(slog.DiscardHandler))
			if tt.secret != nil {
				check.cfg.Secret = tt.secret
			}
			if got := check.verify(t.Context(), tt.token, tt.remote, tt.host).verdict; got != tt.want {
				t.Errorf("verdict = %d, want %d", got, tt.want)
			}
			if got := len(f.calls()); got != tt.calls {
				t.Errorf("calls to Cloudflare = %d, want %d", got, tt.calls)
			}
		})
	}
}

var errBoomTurnstile = errors.New("the secret file is not there")

func TestTurnstileVerifySendsTheSecretTheTokenAndTheAddressOnly(t *testing.T) {
	t.Parallel()
	f := newFakeSiteverify(t)
	check := f.check(slog.New(slog.DiscardHandler))
	if got := check.verify(t.Context(), "ok-abc", "198.51.100.7", testHost).verdict; got != turnstilePassed {
		t.Fatalf("verdict = %d, want passed", got)
	}
	want := url.Values{"secret": {secretForTests}, "response": {"ok-abc"}, "remoteip": {"198.51.100.7"}}
	if diff := cmp.Diff([]url.Values{want}, f.calls()); diff != "" {
		t.Errorf("what Cloudflare was sent (-want +got):\n%s", diff)
	}
	// With no address known, none is sent.
	if got := check.verify(t.Context(), "ok-def", "", testHost).verdict; got != turnstilePassed {
		t.Fatalf("verdict = %d, want passed", got)
	}
	if got := f.calls()[1]; got.Has("remoteip") || len(got) != 2 {
		t.Errorf("without an address Cloudflare was sent %v", got)
	}
}

func TestTurnstileVerifyOfAUsedTokenIsRefused(t *testing.T) {
	t.Parallel()
	f := newFakeSiteverify(t)
	check := f.check(slog.New(slog.DiscardHandler))
	if got := check.verify(t.Context(), "ok-once", "", testHost).verdict; got != turnstilePassed {
		t.Fatalf("first use: %d", got)
	}
	if got := check.verify(t.Context(), "ok-once", "", testHost).verdict; got != turnstileRejected {
		t.Errorf("second use: %d, want rejected", got)
	}
}

func TestTurnstileVerifyWithCloudflareUnreachableFailsClosed(t *testing.T) {
	t.Parallel()
	f := newFakeSiteverify(t)
	check := f.check(slog.New(slog.DiscardHandler))
	f.srv.Close()
	if got := check.verify(t.Context(), "ok-1", "", testHost).verdict; got != turnstileUnavailable {
		t.Errorf("verdict = %d, want unavailable", got)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	f2 := newFakeSiteverify(t)
	if got := f2.check(slog.New(slog.DiscardHandler)).verify(ctx, "ok-1", "", testHost).verdict; got != turnstileUnavailable {
		t.Errorf("a cancelled request: verdict = %d, want unavailable", got)
	}
}

func TestTurnstileNeverLogsTheTokenTheSecretOrTheQuestion(t *testing.T) {
	t.Parallel()
	logs := &strings.Builder{}
	f := newFakeSiteverify(t)
	check := f.check(slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	check.verify(t.Context(), "forged-token-for-tests", "198.51.100.7", testHost)
	f.body = `{"success":false,"error-codes":["invalid-input-secret"]}`
	check.verify(t.Context(), "ok-2", "198.51.100.7", testHost)
	f.body = "garbage"
	check.verify(t.Context(), "ok-3", "198.51.100.7", testHost)
	check.cfg.Secret = func() (string, error) { return "", errBoomTurnstile }
	check.verify(t.Context(), "ok-4", "198.51.100.7", testHost)
	for _, secret := range []string{"forged-token-for-tests", secretForTests, "ok-2", "ok-3", "ok-4", "198.51.100.7"} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("the log holds %q:\n%s", secret, logs)
		}
	}
	if logs.Len() == 0 {
		t.Error("nothing was logged for the failures")
	}
}

func TestTurnstileConfigValidate(t *testing.T) {
	t.Parallel()
	good := func() TurnstileConfig {
		return TurnstileConfig{
			SiteKey: siteKeyForTests, VerifyURL: "https://challenges.cloudflare.com/turnstile/v0/siteverify", HTTP: http.DefaultClient,
			Secret: func() (string, error) { return "s", nil },
		}
	}
	tests := []struct {
		name string
		edit func(c *TurnstileConfig)
		ok   bool
	}{
		{name: "complete", edit: func(*TurnstileConfig) {}, ok: true},
		{name: "no site key", edit: func(c *TurnstileConfig) { c.SiteKey = "" }},
		{name: "no secret", edit: func(c *TurnstileConfig) { c.Secret = nil }},
		{name: "no client", edit: func(c *TurnstileConfig) { c.HTTP = nil }},
		{name: "plain http", edit: func(c *TurnstileConfig) { c.VerifyURL = "http://challenges.cloudflare.com/x" }},
		{name: "not a URL", edit: func(c *TurnstileConfig) { c.VerifyURL = "siteverify" }},
		{name: "a user in the URL", edit: func(c *TurnstileConfig) { c.VerifyURL = "https://user@challenges.cloudflare.com/x" }},
		{name: "no URL", edit: func(c *TurnstileConfig) { c.VerifyURL = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := good()
			tt.edit(&c)
			err := c.validate()
			if (err == nil) != tt.ok || (err != nil && !errors.Is(err, ErrBadTurnstile)) {
				t.Errorf("validate = %v, want ok: %v", err, tt.ok)
			}
		})
	}
}

// turnstileHarness is the UI over its fakes with a model on and Turnstile
// guarding it, a fake Cloudflare, and a log the test reads.
func turnstileHarness(t *testing.T) (*harness, *fakeSiteverify, *strings.Builder) {
	t.Helper()
	f := newFakeSiteverify(t)
	f.hostname = testHost
	logs := &strings.Builder{}
	h := newHarnessWith(t, slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug})), func(c *Config) { c.Turnstile = f.config() })
	h.askModel.turnOn("T0", "T2")
	return h, f, logs
}

// asksWithToken sends question with the widget's token, as the form does.
func (h *harness) asksWithToken(t *testing.T, session, question, token string) *httptest.ResponseRecorder {
	t.Helper()
	q := url.Values{"q": {question}}
	if token != "" {
		q.Set(turnstileField, token)
	}
	return h.serve(t, htmx(withSession(form("/ask", q.Encode()), session)))
}

func TestAskBoxShowsTheWidgetOnlyWhereTurnstileIsOnAndAModelAnswers(t *testing.T) {
	t.Parallel()
	const script = `<script src="` + turnstileScript + `" async defer></script>`
	h, _, _ := turnstileHarness(t)
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, `class="ask-check"`, `id="ask-check"`, `data-sitekey="`+siteKeyForTests+`"`, script, `data-check="1"`, `id="ask-pending"`, `role="status"`)
	// The widget is rendered on demand by ask.js, outside the form, in the explicit form: nothing renders it by itself, and the page
	// has no status line for it.
	wantNoBody(t, rec, "cf-turnstile\"", "data-appearance", "data-callback", "ask-check-status")
	// Without Turnstile there is no widget and no script.
	plain := newHarness(t)
	plain.askModel.turnOn("T0", "T2")
	rec = plain.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "cf-turnstile", "challenges.cloudflare.com", "data-sitekey", "ask-check")
}

func TestAskBoxShowsNoWidgetWhenNoModelAnswers(t *testing.T) {
	t.Parallel()
	f := newFakeSiteverify(t)
	h := newHarnessWith(t, nil, func(c *Config) { c.Turnstile = f.config() })
	// The fake core's model is off until a test turns it on.
	rec := h.serve(t, withSession(request(http.MethodGet, "/", ""), ownerSession))
	wantStatus(t, rec, http.StatusOK)
	wantNoBody(t, rec, "cf-turnstile", "challenges.cloudflare.com")
}

func TestUIPolicyAllowsTurnstileOnlyWhereItIsOn(t *testing.T) {
	t.Parallel()
	h, _, _ := turnstileHarness(t)
	csp := h.serve(t, request(http.MethodGet, "/login", "")).Header().Get("Content-Security-Policy")
	for _, want := range []string{
		"script-src 'self' https://challenges.cloudflare.com;", "frame-src https://challenges.cloudflare.com;", "default-src 'none'",
		"style-src 'self'", "connect-src 'self'", "frame-ancestors 'none'", "form-action 'self'",
	} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP %q lacks %q", csp, want)
		}
	}
	// Turnstile's origin is the only one beyond the UI's own, and nothing is loosened.
	if got := strings.Count(csp, "https://"); got != 2 || strings.Contains(csp, "unsafe") || strings.Contains(csp, "*") {
		t.Errorf("CSP %q names %d outside origins or is loose, want exactly Turnstile's two uses", csp, got)
	}
	plain := newHarness(t).serve(t, request(http.MethodGet, "/login", "")).Header().Get("Content-Security-Policy")
	if plain != contentSecurityPolicy || strings.Contains(plain, "cloudflare") {
		t.Errorf("without Turnstile the CSP is %q, want the strict one", plain)
	}
}

func TestAskBoxPassesAQuestionWithAGoodToken(t *testing.T) {
	t.Parallel()
	h, f, _ := turnstileHarness(t)
	req := htmx(withSession(form("/ask", url.Values{"q": {"how diversified am I?"}, turnstileField: {"ok-1"}}.Encode()), readSession))
	req.Header.Set(openfief.ClientIPHeader, "198.51.100.7")
	rec := h.serve(t, req)
	wantStatus(t, rec, http.StatusOK)
	wantBody(t, rec, "Answered by m")
	if h.askModel.generated() == 0 {
		t.Error("the model was not asked after the check passed")
	}
	calls := f.calls()
	if len(calls) != 1 || calls[0].Get("remoteip") != "198.51.100.7" || calls[0].Get("response") != "ok-1" {
		t.Errorf("Cloudflare was asked %v, want once, with the token and the address the core handed over", calls)
	}
}

func TestAskBoxStopsAQuestionTheCheckRefusesBeforeAnythingIsSentOrCounted(t *testing.T) {
	t.Parallel()
	const never = "Answered by"
	tests := []struct {
		name     string
		token    string
		body     string
		hostname string
		down     bool
		want     []string
		wantNo   []string
	}{
		{
			name:   "no token",
			want:   []string{"A check comes before the model answers", "No result of the check reached OpenFief", "reload the page", "challenges.cloudflare.com", "in Brave: turn Shields off for it"},
			wantNo: []string{"Cloudflare refused the check", "another host", "error codes"},
		},
		{
			name: "a forged token", token: "forged",
			want:   []string{"Cloudflare refused the check", "error codes: invalid-input-response", "Ask again to run a new check"},
			wantNo: []string{"A check comes before the model answers", "forged"},
		},
		{
			name: "a used token", token: "ok-1", body: `{"success":false,"error-codes":["timeout-or-duplicate","internal-error"]}`,
			want:   []string{"Cloudflare refused the check", "error codes: timeout-or-duplicate, internal-error"},
			wantNo: []string{"A check comes before the model answers", "ok-1"},
		},
		{
			name: "a refusal with no code", token: "ok-1", body: `{"success":false}`,
			want: []string{"Cloudflare refused the check", "it gave no error code"},
		},
		{
			name: "codes that are no fixed words", token: "ok-1", body: `{"success":false,"error-codes":["bad-token","<script>alert(1)</script>","SECRET value"]}`,
			want:   []string{"error codes: bad-token)"},
			wantNo: []string{"<script>", "SECRET"},
		},
		{
			name: "a token made for another host", token: "ok-1", hostname: "evil.example.com",
			want:   []string{"The check was made for another host", "made for evil.example.com", "came from " + testHost},
			wantNo: []string{"A check comes before the model answers", "Cloudflare refused the check"},
		},
		{
			name: "Cloudflare down", token: "ok-1", down: true,
			want:   []string{"The check couldn&#39;t run", "nothing was sent to the model"},
			wantNo: []string{"Cloudflare refused the check", "A check comes before the model answers"},
		},
	}
	for _, tt := range tests {
		for _, session := range []string{ownerSession, readSession} {
			t.Run(tt.name+" "+session, func(t *testing.T) {
				t.Parallel()
				h, f, logs := turnstileHarness(t)
				f.body = tt.body
				if tt.hostname != "" {
					f.hostname = tt.hostname
				}
				if tt.down {
					f.srv.Close()
				}
				rec := h.asksWithToken(t, session, "how diversified am I?", tt.token)
				wantStatus(t, rec, http.StatusOK)
				wantBody(t, rec, append(tt.want, "The model gave no answer", `class="ask-help"`, "What needs me")...)
				wantNoBody(t, rec, append(tt.wantNo, never)...)
				wantNothingSentOrCounted(t, h)
				wantNothingLeaked(t, rec, logs.String(), tt.token)
			})
		}
	}
}

// wantNothingLeaked checks that neither the secret key nor the question reached
// the log or the page's card, and that the token did not reach the log.
func wantNothingLeaked(t *testing.T, rec *httptest.ResponseRecorder, logs, token string) {
	t.Helper()
	if strings.Contains(logs, secretForTests) || strings.Contains(rec.Body.String(), secretForTests) {
		t.Error("the secret key reached the log or the page")
	}
	if strings.Contains(logs, "how diversified") {
		t.Errorf("the question reached the log:\n%s", logs)
	}
	if token != "" && strings.Contains(logs, token) {
		t.Errorf("the token reached the log:\n%s", logs)
	}
}

func TestTurnstileOutcomeCarriesOnlyWhatMayBeShown(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
		host string
		want turnstileOutcome
	}{
		{
			name: "codes in order", body: `{"success":false,"error-codes":["timeout-or-duplicate","invalid-input-response"]}`, host: testHost,
			want: turnstileOutcome{verdict: turnstileRejected, codes: []string{"timeout-or-duplicate", "invalid-input-response"}},
		},
		{
			name: "codes that are not fixed words are dropped", body: `{"success":false,"error-codes":["Has Space","ok-code",""]}`, host: testHost,
			want: turnstileOutcome{verdict: turnstileRejected, codes: []string{"ok-code"}},
		},
		{name: "no codes", body: `{"success":false}`, host: testHost, want: turnstileOutcome{verdict: turnstileRejected}},
		{
			name: "the other host and the asking host, without its port", body: `{"success":true,"hostname":"other.example.org"}`, host: "ui.example.com:8443",
			want: turnstileOutcome{verdict: turnstileWrongHost, reported: "other.example.org", asked: "ui.example.com"},
		},
		{name: "a passed token shows nothing", body: `{"success":true,"hostname":"ui.example.com"}`, host: testHost, want: turnstileOutcome{verdict: turnstilePassed}},
		{
			name: "an unavailable check shows nothing", body: `{"success":false,"error-codes":["invalid-input-secret","timeout-or-duplicate"]}`, host: testHost,
			want: turnstileOutcome{verdict: turnstileUnavailable},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := newFakeSiteverify(t)
			f.body = tt.body
			got := f.check(slog.New(slog.DiscardHandler)).verify(t.Context(), "ok-1", "", tt.host)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(turnstileOutcome{})); diff != "" {
				t.Errorf("outcome (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTurnstileAnswerByVerdict(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		outcome turnstileOutcome
		stopped bool
		title   string
		text    []string
	}{
		{
			name: "no token", outcome: turnstileOutcome{verdict: turnstileNoToken}, stopped: true, title: "A check comes before the model answers",
			text: []string{"No result of the check reached OpenFief", "JavaScript", "challenges.cloudflare.com", "in Brave: turn Shields off for it"},
		},
		{
			name: "refused with codes", outcome: turnstileOutcome{verdict: turnstileRejected, codes: []string{"timeout-or-duplicate"}}, stopped: true, title: "Cloudflare refused the check",
			text: []string{"error codes: timeout-or-duplicate", "Ask again to run a new check"},
		},
		{
			name: "refused with none", outcome: turnstileOutcome{verdict: turnstileRejected}, stopped: true, title: "Cloudflare refused the check",
			text: []string{"it gave no error code"},
		},
		{
			name: "another host", outcome: turnstileOutcome{verdict: turnstileWrongHost, reported: "a.example.org", asked: "b.example.org"}, stopped: true,
			title: "The check was made for another host", text: []string{"made for a.example.org", "came from b.example.org"},
		},
		{
			name: "unavailable", outcome: turnstileOutcome{verdict: turnstileUnavailable}, stopped: true, title: "The check couldn't run",
			text: []string{"nothing was sent to the model"},
		},
		{name: "passed", outcome: turnstileOutcome{verdict: turnstilePassed}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card, stopped := turnstileAnswer("how diversified am I?", tt.outcome)
			if stopped != tt.stopped {
				t.Fatalf("stopped = %v, want %v", stopped, tt.stopped)
			}
			if card.Title != tt.title {
				t.Errorf("title = %q, want %q", card.Title, tt.title)
			}
			for _, want := range tt.text {
				if !strings.Contains(card.Text, want) {
					t.Errorf("text %q lacks %q", card.Text, want)
				}
			}
		})
	}
}

// wantNothingSentOrCounted checks that the model was not asked and the core was
// not asked to count a question: a bot that passes no check spends no allowance.
func wantNothingSentOrCounted(t *testing.T, h *harness) {
	t.Helper()
	if n := h.askModel.generated(); n != 0 {
		t.Errorf("the model was asked %d times without a passed check", n)
	}
	if n := len(h.askModel.admitted); n != 0 {
		t.Errorf("the core was asked to count the question %d times", n)
	}
}

func TestAskBoxASingleUseTokenCannotAskTwice(t *testing.T) {
	t.Parallel()
	h, _, _ := turnstileHarness(t)
	first := h.asksWithToken(t, ownerSession, "how diversified am I?", "ok-1")
	wantBody(t, first, "Answered by m")
	again := h.asksWithToken(t, ownerSession, "and by currency?", "ok-1")
	wantBody(t, again, "Cloudflare refused the check", "timeout-or-duplicate")
	wantNoBody(t, again, "Answered by")
	if got := h.askModel.generated(); got == 0 {
		t.Fatal("the first question never reached the model")
	}
}

func TestAskBoxCommandsNeedNoTokenAndNeverCallCloudflare(t *testing.T) {
	t.Parallel()
	h, f, _ := turnstileHarness(t)
	for _, q := range []string{"What needs me?", "allocation by currency", "value on 2026-03-31"} {
		rec := h.asksWithToken(t, readSession, q, "")
		wantStatus(t, rec, http.StatusOK)
		wantBody(t, rec, "Answered from your data · no AI model")
	}
	if n := len(f.calls()); n != 0 {
		t.Errorf("a command called Cloudflare %d times", n)
	}
	if n := h.askModel.generated(); n != 0 {
		t.Errorf("a command reached the model %d times", n)
	}
}

func TestAskBoxChecksBeforeTheCoreCountsAndTheCoreStillRefuses(t *testing.T) {
	t.Parallel()
	h, f, _ := turnstileHarness(t)
	h.askModel.admission = modelsv1.AskAdmission_ASK_ADMISSION_SESSION_LIMIT
	rec := h.asksWithToken(t, ownerSession, "how diversified am I?", "ok-1")
	wantBody(t, rec, "You&#39;ve asked the model enough for now")
	if n := len(f.calls()); n != 1 {
		t.Errorf("Cloudflare was asked %d times, want once, before the core was", n)
	}
	if n := len(h.askModel.admitted); n != 1 {
		t.Errorf("the core was asked %d times, want once after the check passed", n)
	}
}

func TestAskBoxTurnstileWithACoreThatFailsStillKeepsTheCheckFirst(t *testing.T) {
	t.Parallel()
	h, _, _ := turnstileHarness(t)
	h.askModel.admitErr = codeErr(connect.CodeInternal, nil)
	rec := h.asksWithToken(t, ownerSession, "how diversified am I?", "ok-1")
	wantBody(t, rec, "The model couldn&#39;t be checked")
}

func TestNewRefusesTurnstileWithoutAModelToGuard(t *testing.T) {
	t.Parallel()
	f := newFakeSiteverify(t)
	// Config.AskModel is nil here: a UI with Turnstile and nothing to protect is a mistake, not a setting.
	s := &server{log: slog.New(slog.DiscardHandler)}
	if _, err := s.withTurnstile(&Config{Turnstile: f.config()}); !errors.Is(err, ErrBadTurnstile) {
		t.Errorf("withTurnstile = %v, want ErrBadTurnstile", err)
	}
	if csp, err := s.withTurnstile(&Config{}); err != nil || csp != contentSecurityPolicy {
		t.Errorf("without Turnstile: %q, %v, want the strict policy", csp, err)
	}
	if s.turnstile != nil {
		t.Error("a refused config left a check behind")
	}
}

// TestOnlyAModelQuestionWithNoTokenIsToldAChecksNeeded checks the signal the Ask box's script runs the check on (D-208): a response
// carries the HX-Trigger event "openfief-check-needed" only when a question that goes to the model came with no token. A command, a
// question with a token Cloudflare accepts, one it refuses, a plain request without htmx and an install with no Turnstile never get
// it, so the page never runs a check the server does not ask for, and the server still stops the no-token question itself.
func TestOnlyAModelQuestionWithNoTokenIsToldAChecksNeeded(t *testing.T) {
	t.Parallel()
	h, _, _ := turnstileHarness(t)
	trigger := func(rec *httptest.ResponseRecorder) string { return rec.Header().Get(hxTrigger) }

	rec := h.asksWithToken(t, ownerSession, "how diversified am I?", "")
	wantStatus(t, rec, http.StatusOK)
	if got := trigger(rec); got != checkNeededTrigger || checkNeededTrigger != "openfief-check-needed" {
		t.Errorf("a model question with no token: HX-Trigger = %q, want %q", got, "openfief-check-needed")
	}
	// The card is still the no-token card, and nothing reached the model: a page that cannot run the check shows it.
	wantBody(t, rec, "A check comes before the model answers")
	wantNoBody(t, rec, "Answered by")
	wantNothingSentOrCounted(t, h)

	for name, rec := range map[string]*httptest.ResponseRecorder{
		"a command":           h.asksWithToken(t, ownerSession, "what needs me", ""),
		"a good token":        h.asksWithToken(t, ownerSession, "how diversified am I?", "ok-1"),
		"a forged token":      h.asksWithToken(t, ownerSession, "how diversified am I?", "forged"),
		"a reused token":      h.asksWithToken(t, ownerSession, "how diversified am I?", "ok-1"),
		"a question too long": h.asksWithToken(t, ownerSession, strings.Repeat("x", maxModelQuestionRunes+1), ""),
	} {
		if got := trigger(rec); got != "" {
			t.Errorf("%s: HX-Trigger = %q, want none", name, got)
		}
	}

	// Without htmx the box redirects and tells nothing; without Turnstile a model question is not stopped.
	plain := h.serve(t, withSession(request(http.MethodGet, "/ask?"+url.Values{"q": {"how diversified am I?"}}.Encode(), ""), ownerSession))
	if got := trigger(plain); got != "" {
		t.Errorf("a request without htmx: HX-Trigger = %q, want none", got)
	}
	open := newHarness(t)
	open.askModel.turnOn("T0", "T2")
	if got := trigger(open.asks(t, ownerSession, "how diversified am I?")); got != "" {
		t.Errorf("an install with no Turnstile: HX-Trigger = %q, want none", got)
	}
}
