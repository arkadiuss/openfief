package main

import (
	"regexp"
	"strings"
)

// betaOrStable is a release tag from the first beta on: a beta, YYYY.M.N-beta.K, or a stable release,
// YYYY.M.N (docs/dev/sdlc.md#releases). Alpha tags don't count.
var betaOrStable = regexp.MustCompile(`^\d{4}\.\d{1,2}\.\d+(-beta\.\d+)?$`)

// preBeta reports whether tags, the repository's tags one a line, hold no beta or stable release yet.
// Until the first beta the contracts may break in a commit marked breaking, with no author decision;
// from it, only what a decision names in the exceptions file may break.
func preBeta(tags string) bool {
	for t := range strings.SplitSeq(tags, "\n") {
		if betaOrStable.MatchString(strings.TrimSpace(t)) {
			return false
		}
	}
	return true
}
