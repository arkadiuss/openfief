package kernel

import "errors"

// ErrAuthzUnavailable reports that the authorization model could not answer
// now, such as a SpiceDB that is down or too slow (D-164): fail closed. It is
// no refusal, since no one was refused on the merits, so a caller that sees
// it answers "unavailable, try again" and never "forbidden" or "failed". Its
// text is what a client reads, so it names nothing.
var ErrAuthzUnavailable = errors.New("authorization is unavailable")

// Unavailable returns, when err says the authorization model or the audit log
// could not answer now, the one of ErrAuthzUnavailable and ErrAuditUnavailable
// it wraps, and nil otherwise: what a Connect adapter answers as "unavailable,
// try again", with a text that names nothing.
func Unavailable(err error) error {
	for _, e := range []error{ErrAuthzUnavailable, ErrAuditUnavailable} {
		if errors.Is(err, e) {
			return e
		}
	}
	return nil
}

// ErrAuditUnavailable reports that the audit log could not take the entry of
// an owner's change, so the change was not made (D-211): fail closed. Like
// ErrAuthzUnavailable it is no refusal, and a caller that sees it answers
// "unavailable, try again". Its text names nothing.
var ErrAuditUnavailable = errors.New("the audit log is unavailable")
