package plugins_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// reviewPayload is the second security review's: in a terminal it moves
// the cursor six lines up, erases that line and writes a fake one over the
// install's consent screen.
const reviewPayload = "\x1b[6A\r\x1b[2K  Egress hosts: none"

// TestManifestRefusesUnsafeText: a manifest string with a control
// character or a bidirectional override, in any field or anywhere in the
// config schema, is refused with ErrUnsafeText; the same text without one
// passes (D-46, D-90).
func TestManifestRefusesUnsafeText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		change  func(m *plugins.Manifest)
		wantErr error
	}{
		{name: "plain text and a schema with titles", change: func(m *plugins.Manifest) {
			m.ClusterService = &plugins.ClusterServiceRequest{Description: "a Ghostfolio — in this cluster"}
			m.ConfigSchema = []byte(`{"type":"object","properties":{"start":{"title":"Start","description":"First day, ISO 8601",` +
				`"enum":["2024-01-02"]}}}`)
		}},
		{name: "the review's payload in a schema property name", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.ConfigSchema = []byte(`{"type":"object","properties":{"\u001b[6A\r\u001b[2K  Egress hosts: none":{"type":"string"}}}`)
		}},
		{name: "a C1 control in a schema title", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.ConfigSchema = []byte(`{"type":"object","properties":{"a":{"title":"x\u009b2Ky"}}}`)
		}},
		{name: "a right-to-left override in an enum value", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.ConfigSchema = []byte(`{"properties":{"a":{"enum":["safe","\u202egnp.exe"]}}}`)
		}},
		{name: "a bidi isolate in a schema description", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.ConfigSchema = []byte(`{"description":"\u2066hidden\u2069"}`)
		}},
		{name: "invalid UTF-8 in the schema", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.ConfigSchema = []byte("{\"title\":\"\x9b2K\"}")
		}},
		{name: "the review's payload in a cluster service description", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.ClusterService = &plugins.ClusterServiceRequest{Description: reviewPayload}
		}},
		{name: "an escape in a mount path", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.Secrets[0].MountPath = "/etc/openfief/\x1b[2Kx"
		}},
		{name: "a tab in a config field name", wantErr: plugins.ErrUnsafeText, change: func(m *plugins.Manifest) {
			m.EgressFromConfig = []string{"base\turl"}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := samplePlugin()
			tt.change(&p.Manifest)
			err := p.Manifest.Validate()
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestSafeText: C0 and C1 controls, DEL, bidirectional embeddings,
// overrides and isolates and invalid UTF-8 are unsafe; other text,
// non-ASCII letters and punctuation included, is safe.
func TestSafeText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		s    string
		want bool
	}{
		{"Zażółć gęślą jaźń — ok", true},
		{"", true},
		{"a\nb", false},
		{"del\x7f", false},
		{"c1\u0085", false},
		{"lre\u202a", false},
		{"rlo\u202e", false},
		{"pdi\u2069", false},
		{"lrm\u200e", true},
		{"bad\xff", false},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			t.Parallel()
			if got := plugins.SafeText(tt.s); got != tt.want {
				t.Errorf("SafeText(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}
