package ingest

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxFileNameBytes caps a file name hint, the usual limit of one path
// element.
const maxFileNameBytes = 255

// FileNameHint returns what of name, a statement file's name as the owner's
// browser or tool gave it, may be passed to a connector as a hint: its last
// path element (after the last slash or backslash), without control
// characters, invalid UTF-8 or surrounding spaces, cut to 255 bytes on a
// character boundary. "." and ".." become empty. The result is never a
// path, and nothing may use it as one.
func FileNameHint(name string) string {
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}
	var b strings.Builder
	for _, r := range name {
		if r == utf8.RuneError || unicode.IsControl(r) {
			continue
		}
		if b.Len()+utf8.RuneLen(r) > maxFileNameBytes {
			break
		}
		b.WriteRune(r)
	}
	hint := strings.TrimSpace(b.String())
	if hint == "." || hint == ".." {
		return ""
	}
	return hint
}
