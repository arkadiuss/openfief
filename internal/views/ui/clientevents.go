package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Client events (D-227): the pages report a fixed set of events about
// themselves to the UI, which logs them and counts them, so that a check that
// never completes in one browser can be debugged from what the browsers say
// rather than by guessing. What may be reported is a vocabulary, not text: every
// field is checked against a fixed set or a pattern here, and what does not fit
// is dropped. Nothing typed, no amount, identifier or page content is ever
// accepted (GR-12).
const (
	// clientEventsPath is where the pages post their events.
	clientEventsPath = "/client-events"
	// maxClientEventsBody and maxClientEvents bound one report: a batch is a few
	// short events, so anything larger is no page of ours.
	maxClientEventsBody = 4 << 10
	maxClientEvents     = 20
	// maxClientMessage is how much of an uncaught error's message is kept.
	maxClientMessage = 200
	// maxClientMs is the longest time an event may report, ten minutes.
	maxClientMs = 10 * 60 * 1000
	// maxClientLine is the largest line number kept; minHTTPStatus and maxHTTPStatus
	// bound the statuses of a failed request.
	maxClientLine = 1 << 20
	minHTTPStatus = 100
	maxHTTPStatus = 599
	// clientReportsPerMinute and clientRateAddresses limit one address's reports in
	// a minute, and how many addresses the limit remembers at once.
	clientReportsPerMinute = 30
	clientRateAddresses    = 4096
	// cloudflareScript is how an event names a script that came from Cloudflare:
	// the host only, never a path or a query.
	cloudflareScript = "challenges.cloudflare.com"
	// labelOther names what no fixed word fits: a script, a code or a browser the
	// vocabulary does not know. labelNone is a path the event did not carry.
	labelOther = "other"
	labelNone  = "none"
	// The browser families a User-Agent is read into.
	familyChrome = "chrome"
	familyEdge   = "edge"
	// decimal and msPerSecond are the bases of two conversions below.
	decimal     = 10
	msPerSecond = 1000
)

// The events a page may report. Turnstile's states carry the time since the
// check began; the others carry what the browser says of an error or a failed
// request.
const (
	evScriptLoaded      = "turnstile-script-loaded"
	evBeforeInteractive = "turnstile-before-interactive"
	evSuccess           = "turnstile-success"
	evError             = "turnstile-error"
	evExpired           = "turnstile-expired"
	evTimeout           = "turnstile-timeout"
	evUnsupported       = "turnstile-unsupported"
	evFallbackShown     = "turnstile-fallback-shown"
	evBlockedScript     = "turnstile-blocked-script"
	evSlow              = "turnstile-slow"
	evScriptError       = "script-error"
	evRequestFailed     = "request-failed"
)

// knownClientKind reports whether kind is in the fixed vocabulary; any other kind
// is dropped.
func knownClientKind(kind string) bool {
	switch kind {
	case evScriptLoaded, evBeforeInteractive, evSuccess, evError, evExpired, evTimeout,
		evUnsupported, evFallbackShown, evBlockedScript, evSlow, evScriptError, evRequestFailed:
		return true
	}
	return false
}

var (
	// errorCodePattern is what a Cloudflare error code looks like: lower-case
	// letters, digits and hyphens.
	errorCodePattern = regexp.MustCompile(`^[a-z0-9-]{1,48}$`)
	// numericCodePattern is the shape of Cloudflare's numeric error codes, the only
	// ones that become a metric label, so that a browser cannot make labels up.
	numericCodePattern = regexp.MustCompile(`^\d{6}$`)
	// scriptPathPattern is a script of the UI's own: a path under /static/.
	scriptPathPattern = regexp.MustCompile(`^/static/[A-Za-z0-9_./-]{1,100}$`)
	// plainMessagePattern is the engine's own words: letters, digits and a little
	// punctuation, without quotes, which a parse error uses to quote its input.
	plainMessagePattern = regexp.MustCompile(`^[A-Za-z0-9 _.,:;'()\[\]<>=/!?+*-]+$`)
	// digitRun finds three digits or more in a row, which an amount, an identifier or
	// a date would make; an engine's own message has none.
	digitRun = regexp.MustCompile(`\d{3,}`)
)

// rawClientEvent is one event as the page sent it, every field optional.
type rawClientEvent struct {
	Kind   string `json:"k"`
	Ms     *int64 `json:"ms"`
	Code   string `json:"code"`
	Path   string `json:"path"`
	Msg    string `json:"msg"`
	Src    string `json:"src"`
	Line   *int64 `json:"line"`
	Status *int64 `json:"status"`
	Brave  bool   `json:"brave"`
}

// clientEvent is an event that passed the checks: only fixed words, patterns and
// numbers in range.
type clientEvent struct {
	Kind    string
	Ms      int64 // -1 when absent
	Code    string
	Path    string
	Message string
	Source  string
	Line    int64 // -1 when absent
	Status  int64 // -1 when absent
	Brave   bool
}

// validClientEvent returns the event with every field that does not pass the
// checks cleared, and false for one whose kind is not in the vocabulary or whose
// numbers are out of range.
func validClientEvent(raw *rawClientEvent) (clientEvent, bool) {
	if !knownClientKind(raw.Kind) {
		return clientEvent{}, false
	}
	ev := clientEvent{Kind: raw.Kind, Ms: -1, Line: -1, Status: -1, Brave: raw.Brave}
	if !ev.setNumbers(raw) {
		return clientEvent{}, false
	}
	if errorCodePattern.MatchString(raw.Code) {
		ev.Code = raw.Code
	}
	if raw.Path == checkPathInvisible || raw.Path == checkPathVisible {
		ev.Path = raw.Path
	}
	if raw.Kind == evScriptError {
		report := scriptAndMessage(raw.Src, raw.Msg)
		ev.Source, ev.Message = report.source, report.message
	}
	return ev, true
}

// setNumbers copies the numbers of raw that are in range into ev, and says false
// for a time or a status that is out of range, which drops the whole event; a line
// out of range is only left out.
func (ev *clientEvent) setNumbers(raw *rawClientEvent) bool {
	if raw.Ms != nil {
		ev.Ms = *raw.Ms
	}
	if raw.Status != nil {
		ev.Status = *raw.Status
	}
	if raw.Line != nil && inRange(*raw.Line, 0, maxClientLine) {
		ev.Line = *raw.Line
	}
	// An absent number is -1, which passes; a status is 0 (no response) or an HTTP status.
	return inRange(ev.Ms, -1, maxClientMs) && (inRange(ev.Status, -1, 0) || inRange(ev.Status, minHTTPStatus, maxHTTPStatus))
}

// inRange reports whether n is from lo to hi, both included.
func inRange(n, lo, hi int64) bool {
	return n >= lo && n <= hi
}

// scriptReport is what is kept of where an error came from: the script, and its
// message when that is safe.
type scriptReport struct{ source, message string }

// scriptAndMessage returns the script an error came from, as one of the UI's own
// paths, or Cloudflare's host, or "other", and its message only for a script of
// the UI's own and only when it is the engine's plain words: a script from
// anywhere else reports "Script error." and nothing a page could hold, and a
// message that quotes something (a parse error quotes its input) or holds
// digits that might be an amount is dropped, not trimmed. What is kept is cut
// to maxClientMessage characters.
func scriptAndMessage(src, msg string) scriptReport {
	switch {
	case scriptPathPattern.MatchString(src):
		if utf8.RuneCountInString(msg) > maxClientMessage {
			msg = string([]rune(msg)[:maxClientMessage])
		}
		if !plainMessagePattern.MatchString(msg) || digitRun.MatchString(msg) {
			return scriptReport{source: src}
		}
		return scriptReport{source: src, message: msg}
	case strings.HasPrefix(src, "https://"+cloudflareScript+"/"):
		return scriptReport{source: cloudflareScript}
	}
	return scriptReport{source: labelOther}
}

// browser is the family and major version the User-Agent names, as fixed words
// and a number; nothing else of the header is kept.
type browser struct {
	Family string
	Major  int
	Mobile bool
}

// browserOf reads the browser out of a User-Agent: Edge, Opera, Samsung Internet,
// Firefox, Chrome (and so the browsers built on it) or Safari; anything else is
// "other" with major 0. brave says the page found Brave's own object, which its
// User-Agent hides.
func browserOf(ua string, brave bool) browser {
	b := browser{Family: labelOther, Mobile: strings.Contains(ua, "Mobile")}
	for _, f := range []struct{ family, token string }{
		{familyEdge, "Edg/"},
		{familyEdge, "EdgA/"},
		{familyEdge, "EdgiOS/"},
		{"opera", "OPR/"},
		{"samsung", "SamsungBrowser/"},
		{"firefox", "Firefox/"},
		{"firefox", "FxiOS/"},
		{familyChrome, "Chrome/"},
		{familyChrome, "CriOS/"},
		{"safari", "Version/"},
	} {
		if major, ok := majorAfter(ua, f.token); ok {
			b.Family, b.Major = f.family, major
			break
		}
	}
	if b.Family == "safari" && !strings.Contains(ua, "Safari/") {
		b.Family, b.Major = labelOther, 0
	}
	if brave && b.Family == familyChrome {
		b.Family = "brave"
	}
	return b
}

// majorAfter returns the whole number that follows token in ua, up to three digits.
func majorAfter(ua, token string) (int, bool) {
	_, after, found := strings.Cut(ua, token)
	if !found {
		return 0, false
	}
	n, digits := 0, 0
	for _, c := range after {
		if c < '0' || c > '9' || digits == 3 {
			break
		}
		n, digits = n*decimal+int(c-'0'), digits+1
	}
	return n, digits > 0
}

// addressLimit lets one address make clientReportsPerMinute reports a minute. It
// keeps at most clientRateAddresses addresses, forgetting the ones whose minute
// has ended; when all are current, an address it does not know is refused, so
// that many addresses cannot grow it. The address is used for this and never
// recorded.
type addressLimit struct {
	mu      sync.Mutex
	now     func() time.Time
	windows map[string]clientWindow
}

type clientWindow struct {
	start time.Time
	count int
}

// allow counts a report from address and says whether it is within the limit.
func (l *addressLimit) allow(address string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	w, known := l.windows[address]
	if !known && len(l.windows) >= clientRateAddresses {
		for a, old := range l.windows {
			if now.Sub(old.start) >= time.Minute {
				delete(l.windows, a)
			}
		}
		if len(l.windows) >= clientRateAddresses {
			return false
		}
	}
	if now.Sub(w.start) >= time.Minute {
		w = clientWindow{start: now}
	}
	w.count++
	l.windows[address] = w
	return w.count <= clientReportsPerMinute
}

// clientEvents receives the pages' events (D-227), logs each as a record without
// the address and counts it.
type clientEvents struct {
	log       *slog.Logger
	limit     *addressLimit
	events    metric.Int64Counter
	tokenTime metric.Float64Histogram
}

// newClientEvents returns the receiver, counting on meter, with now for its limit.
func newClientEvents(log *slog.Logger, meter metric.Meter, now func() time.Time) (*clientEvents, error) {
	events, err := meter.Int64Counter("openfief.ui.client_events",
		metric.WithDescription("Events the UI's pages reported about themselves, by kind, browser family and, for a Turnstile error, its code"))
	if err != nil {
		return nil, fmt.Errorf("ui: client events counter: %w", err)
	}
	tokenTime, err := meter.Float64Histogram("openfief.ui.turnstile.token_seconds",
		metric.WithDescription("Seconds from a Turnstile check starting to its token arriving, by browser family and path"),
		metric.WithUnit("s"), metric.WithExplicitBucketBoundaries(tokenBuckets()...))
	if err != nil {
		return nil, fmt.Errorf("ui: client events histogram: %w", err)
	}
	return &clientEvents{log: log, limit: &addressLimit{now: now, windows: map[string]clientWindow{}}, events: events, tokenTime: tokenTime}, nil
}

// tokenBuckets are the bounds, in seconds, of the time-to-token histogram's buckets.
func tokenBuckets() []float64 {
	return []float64{0.25, 0.5, 1, 2, 5, 10, 20, 60}
}

// handler answers a page's report: 204 when it was taken or discarded, 400 for a
// body that is no report, 413 for one over the cap and 429 past the address's
// limit. The cap and the limit come before anything is read or kept. With
// reporting off, every report is discarded unread.
func (c *clientEvents) handler(on bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !on {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !c.limit.allow(clientAddress(r)) {
			http.Error(w, "too many reports", http.StatusTooManyRequests)
			return
		}
		events, ok := readReport(w, r)
		if !ok {
			return
		}
		for i := range events {
			if i == maxClientEvents {
				break
			}
			if ev, valid := validClientEvent(&events[i]); valid {
				c.record(r.Context(), &ev, browserOf(r.UserAgent(), ev.Brave))
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// readReport decodes a report's events, or answers 413 for a body over the cap and
// 400 for one that is no report, and says false.
func readReport(w http.ResponseWriter, r *http.Request) ([]rawClientEvent, bool) {
	var body struct {
		Events []rawClientEvent `json:"events"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxClientEventsBody)).Decode(&body); err != nil {
		if _, tooBig := errors.AsType[*http.MaxBytesError](err); tooBig {
			http.Error(w, "report too large", http.StatusRequestEntityTooLarge)
			return nil, false
		}
		http.Error(w, "not a report", http.StatusBadRequest)
		return nil, false
	}
	return body.Events, true
}

// record logs one valid event, without an address, and counts it. The metric's
// labels are the kind, the browser family and, for a Turnstile error, its code
// when it is a six-digit one; a time to token is observed from a success.
func (c *clientEvents) record(ctx context.Context, ev *clientEvent, b browser) {
	c.log.LogAttrs(ctx, slog.LevelInfo, "client event", logAttrs(ev, b)...)
	labels := []attribute.KeyValue{attribute.String("kind", ev.Kind), attribute.String("browser", b.Family)}
	if ev.Code != "" {
		code := labelOther
		if numericCodePattern.MatchString(ev.Code) {
			code = ev.Code
		}
		labels = append(labels, attribute.String("code", code))
	}
	c.events.Add(ctx, 1, metric.WithAttributes(labels...))
	if ev.Kind == evSuccess && ev.Ms >= 0 {
		path := ev.Path
		if path == "" {
			path = labelNone
		}
		c.tokenTime.Record(ctx, float64(ev.Ms)/msPerSecond, metric.WithAttributes(attribute.String("browser", b.Family), attribute.String("path", path)))
	}
}

// logAttrs are the fields of an event's log record: only what passed the checks,
// and never an address.
func logAttrs(ev *clientEvent, b browser) []slog.Attr {
	attrs := []slog.Attr{
		slog.String("kind", ev.Kind), slog.String("browser", b.Family), slog.Int("major", b.Major), slog.Bool("mobile", b.Mobile),
	}
	for _, f := range []struct {
		key string
		n   int64
	}{{"ms", ev.Ms}, {"line", ev.Line}, {"status", ev.Status}} {
		if f.n >= 0 {
			attrs = append(attrs, slog.Int64(f.key, f.n))
		}
	}
	for _, f := range []struct{ key, value string }{
		{"code", ev.Code}, {"path", ev.Path}, {"script", ev.Source}, {"message", ev.Message},
	} {
		if f.value != "" {
			attrs = append(attrs, slog.String(f.key, f.value))
		}
	}
	return attrs
}
