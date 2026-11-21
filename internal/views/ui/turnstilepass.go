package ui

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// The pass (D-208): a visitor who passed Turnstile isn't checked again for 30
// minutes after their last model question, and at most 2 hours from the check,
// through a signed cookie tied to their session that holds nothing else. The
// rate limits and the budget still count every question: the pass skips only
// the check that a person is asking, never the core's admission.
//
// The cookie's value is "v1.<checked>.<expires>.<mac>": the time of the check
// and the time the pass ends, both Unix seconds, and an HMAC-SHA256 over them
// and the ID of the session it was made for, in unpadded URL-safe base64. It
// names no one: the session ID is in the MAC and not in the cookie, so the
// cookie alone says nothing about whose it is.
const (
	// checkedCookieName is the pass cookie's name. The __Secure- prefix makes the
	// browser insist on Secure; __Host- would insist on Path=/, which the pass
	// does not use.
	checkedCookieName = "__Secure-openfief_ask_pass"
	// passCookiePath scopes the cookie to the Ask box's own request, the only
	// one that reads it.
	passCookiePath = "/ask"
	// defaultPassIdle and defaultPassMax are D-208's windows: 30 minutes after the
	// last model question, and 2 hours from the check.
	defaultPassIdle = 30 * time.Minute
	defaultPassMax  = 2 * time.Hour
	// passVersion starts the cookie's value, so the format can change.
	passVersion = "v1"
	// passFields is the number of parts of a cookie's value.
	passFields = 4
	// maxPassValue bounds what is parsed of a cookie's value: a real one is about 80
	// bytes.
	maxPassValue = 128
	// passSkew is how far in the future a check time may be, for clocks of
	// replicas that differ a little.
	passSkew = time.Minute
	// derivationLabel separates the pass's key from every other use of the widget's
	// secret, and signedLabel the MAC's input from every other message.
	derivationLabel = "openfief ui ask pass key v1"
	signedLabel     = "openfief ui ask pass v1"
	// verdictPass is what the log says of a question that went on with a pass: a
	// fixed word, like the verdicts' (turnstileVerdict.String).
	verdictPass = "pass"
)

// windows are the pass's two lengths: how long after the last model question it
// lasts, and how long after the check at the most.
type windows struct{ idle, limit time.Duration }

// passWindows returns the pass's two windows, each its default for a zero config
// value.
func (c *TurnstileConfig) passWindows() windows {
	w := windows{idle: c.PassIdle, limit: c.PassMax}
	if w.idle <= 0 {
		w.idle = defaultPassIdle
	}
	if w.limit <= 0 {
		w.limit = defaultPassMax
	}
	return w
}

// passKey derives the key the pass is signed with from the widget's secret key
// (HMAC-SHA256 under a label), so that the install has one without a new Secret
// and the pass's key is not the secret itself. A rotated secret ends every pass,
// which then falls back to the check. It is false when the secret can't be read.
func (t *turnstileCheck) passKey() ([]byte, bool) {
	secret, err := t.cfg.Secret()
	if err != nil || secret == "" {
		return nil, false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(derivationLabel))
	return mac.Sum(nil), true
}

// passMAC signs a pass for sessionID: the check time and the end, in Unix
// seconds, each field ended so that none can run into the next.
func passMAC(key []byte, sessionID string, checked, expires int64) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signedLabel + "\n" + sessionID + "\n" + strconv.FormatInt(checked, 10) + "\n" + strconv.FormatInt(expires, 10)))
	return mac.Sum(nil)
}

// issue returns the pass cookie for sessionID's check made at checked: it ends 30
// minutes from now, and never later than 2 hours from the check, which is what
// makes the window slide up to a cap. It is nil for a session with no ID, a
// secret that can't be read, or a pass that would already have ended.
func (t *turnstileCheck) issue(sessionID string, checked time.Time) *http.Cookie {
	if sessionID == "" {
		return nil
	}
	key, ok := t.passKey()
	if !ok {
		return nil
	}
	w := t.cfg.passWindows()
	now := t.now()
	expires := now.Add(w.idle)
	if end := checked.Add(w.limit); end.Before(expires) {
		expires = end
	}
	if !expires.After(now) {
		return nil
	}
	value := strings.Join([]string{
		passVersion, strconv.FormatInt(checked.Unix(), 10), strconv.FormatInt(expires.Unix(), 10),
		base64.RawURLEncoding.EncodeToString(passMAC(key, sessionID, checked.Unix(), expires.Unix())),
	}, ".")
	return &http.Cookie{
		Name: checkedCookieName, Value: value, Path: passCookiePath, Expires: expires, MaxAge: int(expires.Sub(now) / time.Second),
		HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
	}
}

// usePass reads r's pass cookie and returns when the check it vouches for was
// made, or false for none: no cookie, one that is malformed, whose MAC fails
// (forged, or made for another session or under another key), that has ended or
// is past 2 hours from its check, or that claims to be from the future. Nothing
// of it is ever logged, and a pass that fails is no error: the question goes on
// to the check.
func (t *turnstileCheck) usePass(r *http.Request, sessionID string) (time.Time, bool) {
	c, err := r.Cookie(checkedCookieName)
	if err != nil || sessionID == "" {
		return time.Time{}, false
	}
	p, ok := parsePass(c.Value)
	if !ok {
		return time.Time{}, false
	}
	key, ok := t.passKey()
	if !ok || !hmac.Equal(p.mac, passMAC(key, sessionID, p.checked, p.expires)) {
		return time.Time{}, false
	}
	now := t.now()
	checkedAt, end := time.Unix(p.checked, 0), time.Unix(p.expires, 0)
	if !now.Before(end) || !now.Before(checkedAt.Add(t.cfg.passWindows().limit)) || checkedAt.After(now.Add(passSkew)) {
		return time.Time{}, false
	}
	return checkedAt, true
}

// parsedPass is a cookie's value read apart: the two times in Unix seconds and the
// MAC they were signed with, none of them yet checked.
type parsedPass struct {
	checked, expires int64
	mac              []byte
}

// parsePass reads a cookie's value, "v1.<checked>.<expires>.<mac>", or says false
// for one that is over-long or not of that shape.
func parsePass(value string) (parsedPass, bool) {
	if len(value) > maxPassValue {
		return parsedPass{}, false
	}
	parts := strings.Split(value, ".")
	if len(parts) != passFields || parts[0] != passVersion {
		return parsedPass{}, false
	}
	checked, errChecked := strconv.ParseInt(parts[1], 10, 64)
	expires, errExpires := strconv.ParseInt(parts[2], 10, 64)
	mac, errMAC := base64.RawURLEncoding.DecodeString(parts[3])
	if errChecked != nil || errExpires != nil || errMAC != nil {
		return parsedPass{}, false
	}
	return parsedPass{checked: checked, expires: expires, mac: mac}, true
}
