package cli

import (
	"bytes"
	"testing"
)

// TestSayEscapesControls: whatever reaches say, a control character other
// than a newline or a tab, a bidirectional override or invalid UTF-8 is
// written as a visible escape, so no string the CLI prints can steer the
// owner's terminal (D-46, D-90).
func TestSayEscapesControls(t *testing.T) {
	t.Parallel()
	tests := []struct{ in, want string }{
		{"plain\ttext\n", "plain\ttext\n"},
		{"\x1b[6A\r\x1b[2K  Egress hosts: none\n", `\u001b[6A\u000d\u001b[2K  Egress hosts: none` + "\n"},
		{"c1\u009b and rlo\u202e\n", `c1\u009b and rlo\u202e` + "\n"},
		{"bad\xff\n", `bad\ufffd` + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			var b bytes.Buffer
			if err := say(&b, tt.in); err != nil {
				t.Fatalf("say: %v", err)
			}
			if got := b.String(); got != tt.want {
				t.Errorf("say(%q) wrote %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
