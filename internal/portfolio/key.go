package portfolio

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNoKey reports a name no key can be made from, such as one of only
// punctuation.
var ErrNoKey = errors.New("no key can be made from the name")

// polishLetters are the Polish letters a key spells in ASCII, and
// polishASCII the letters it spells them with, in the same order.
const (
	polishLetters = "ąćęłńóśźż"
	polishASCII   = "acelnoszz"
)

// keyRune returns the ASCII letter or digit a key spells c with, Polish
// letters included, and false for any other character.
func keyRune(c rune) (rune, bool) {
	if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
		return c, true
	}
	for i, p := range []rune(polishLetters) {
		if p == c {
			return rune(polishASCII[i]), true
		}
	}
	return 0, false
}

// KeyFromName makes a key from a name the owner typed, so the owner never
// types an ID (D-71, D-72): the name in lower case, Polish letters spelled
// in ASCII, every run of other characters that are not letters or digits
// a single hyphen, with none at either end, cut to 64 characters. It is
// the form of a KindKey, and sleeve and account IDs take it too:
// "Emergency fund" is "emergency-fund", "Konto główne" "konto-glowne".
func KeyFromName(name string) (string, error) {
	var b strings.Builder
	gap := false
	for _, c := range strings.ToLower(name) {
		k, ok := keyRune(c)
		if !ok {
			gap = true
			continue
		}
		if gap && b.Len() > 0 {
			b.WriteByte('-')
		}
		gap = false
		b.WriteRune(k)
	}
	key := b.String()
	if len(key) > maxKindKeyLen {
		key = strings.TrimRight(key[:maxKindKeyLen], "-")
	}
	if key == "" {
		return "", fmt.Errorf("%w: %q", ErrNoKey, name)
	}
	return key, nil
}
