package kernel

import "errors"

// ErrAuthzUnavailable reports that the authorization model could not answer
// now, such as a SpiceDB that is down or too slow (D-164): fail closed. It is
// no refusal, since no one was refused on the merits, so a caller that sees
// it answers "unavailable, try again" and never "forbidden" or "failed". Its
// text is what a client reads, so it names nothing.
var ErrAuthzUnavailable = errors.New("authorization is unavailable")
