package ui

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestCheckPathKeepsOnlyTheTwoFixedWords(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, raw, want string }{
		{"invisible", "invisible", "invisible"},
		{"visible fallback", "visible-fallback", "visible-fallback"},
		{"absent", "", "unknown"},
		{"other case", "Invisible", "unknown"},
		{"padded", " invisible", "unknown"},
		{"typed data", "private-note-9f3", "unknown"},
		{"long", strings.Repeat("a", 5000), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := checkPath(tt.raw); got != tt.want {
				t.Errorf("checkPath(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestAskBoxLogsWhichPathMadeTheTokenAndNeverEchoesIt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, field, wantLog string
	}{
		{"invisible", "invisible", "path=invisible"},
		{"visible fallback", "visible-fallback", "path=visible-fallback"},
		{"absent", "", "path=unknown"},
		{"typed data", "private-note-9f3", "path=unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, _, logs := turnstileHarness(t)
			q := url.Values{"q": {"how diversified am I?"}, turnstileField: {"ok-1"}}
			if tt.field != "" {
				q.Set(checkPathField, tt.field)
			}
			rec := h.serve(t, htmx(withSession(form("/ask", q.Encode()), readSession)))
			wantStatus(t, rec, http.StatusOK)
			if !strings.Contains(logs.String(), "verdict=passed") || !strings.Contains(logs.String(), tt.wantLog) {
				t.Errorf("the log does not say the verdict and %q:\n%s", tt.wantLog, logs)
			}
			if tt.name == "typed data" && (strings.Contains(logs.String(), tt.field) || strings.Contains(rec.Body.String(), tt.field)) {
				t.Errorf("the typed value %q reached the log or the page", tt.field)
			}
		})
	}
}
