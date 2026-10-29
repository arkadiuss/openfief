package ingest_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
)

func TestFileNameHint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, want string
	}{
		{name: "a plain name", in: "statement.xml", want: "statement.xml"},
		{name: "a Unicode name", in: "wyciąg 2026.csv", want: "wyciąg 2026.csv"},
		{name: "a Unix path", in: "../../etc/passwd", want: "passwd"},
		{name: "a Windows path", in: `C:\Users\owner\statement.xml`, want: "statement.xml"},
		{name: "a trailing slash", in: "statements/", want: ""},
		{name: "dot", in: ".", want: ""},
		{name: "two dots", in: "..", want: ""},
		{name: "control characters", in: "state\x00ment\r\n.xml", want: "statement.xml"},
		{name: "invalid UTF-8", in: "state\xffment.xml", want: "statement.xml"},
		{name: "surrounding spaces", in: "  statement.xml\t", want: "statement.xml"},
		{name: "empty", in: "", want: ""},
		{name: "too long", in: strings.Repeat("a", 300), want: strings.Repeat("a", 255)},
		{name: "too long, cut on a character boundary", in: strings.Repeat("a", 254) + "ą", want: strings.Repeat("a", 254)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, ingest.FileNameHint(tt.in)); diff != "" {
				t.Errorf("FileNameHint(%q) (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}
