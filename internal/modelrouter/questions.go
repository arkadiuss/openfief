package modelrouter

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

// ErrInvalidQuestionLimits reports a limit that is negative.
var ErrInvalidQuestionLimits = errors.New("invalid question limits")

// QuestionLimits are how many questions beyond the Ask box's commands one
// subject, a session or an address, may ask a model in a rolling hour and in a
// rolling day (D-208). A zero limit is no limit, which is what an installation
// that sets none gets; the demo sets D-208's 10 an hour and 30 a day.
type QuestionLimits struct {
	PerHour int
	PerDay  int
}

// DefaultViewerLimits are the limits a viewer's session has unless the owner
// sets others: 10 questions an hour and 30 a day, D-208's values. A viewer
// (D-65) is not the owner, and each of their questions is a model call the
// owner pays for, so their allowance is limited by default; the owner's own is
// limited only when set.
func DefaultViewerLimits() QuestionLimits {
	return QuestionLimits{PerHour: defaultViewerPerHour, PerDay: defaultViewerPerDay}
}

// The viewer's default allowance.
const (
	defaultViewerPerHour = 10
	defaultViewerPerDay  = 30
)

// Stricter returns the limits that hold when both l and o do: in each window
// the lower of the two limits, and the other's when one has none.
func (l QuestionLimits) Stricter(o QuestionLimits) QuestionLimits {
	return QuestionLimits{PerHour: stricter(l.PerHour, o.PerHour), PerDay: stricter(l.PerDay, o.PerDay)}
}

// stricter returns the lower of two limits, where zero is no limit.
func stricter(a, b int) int {
	switch {
	case a <= 0:
		return b
	case b <= 0:
		return a
	}
	return min(a, b)
}

// Enabled reports whether any limit is set.
func (l QuestionLimits) Enabled() bool { return l.PerHour > 0 || l.PerDay > 0 }

// Validate returns nil, or an error wrapping ErrInvalidQuestionLimits for a
// negative limit or an hourly limit above the daily one.
func (l QuestionLimits) Validate() error {
	if l.PerHour < 0 || l.PerDay < 0 {
		return fmt.Errorf("%w: a limit is negative", ErrInvalidQuestionLimits)
	}
	if l.PerHour > 0 && l.PerDay > 0 && l.PerHour > l.PerDay {
		return fmt.Errorf("%w: more questions an hour than a day", ErrInvalidQuestionLimits)
	}
	return nil
}

// The windows the limits count questions in.
const (
	hourWindow = time.Hour
	dayWindow  = 24 * time.Hour
	// QuestionWindow is how long a question counts: the longest window, after
	// which a stored question can be forgotten.
	QuestionWindow = dayWindow
)

// QuestionVerdict is the decision on one more question of one subject.
type QuestionVerdict struct {
	// Allowed is true when the question may go to the model.
	Allowed bool
	// ResetsAt, when the question is not allowed, is the first moment it
	// would be: when enough of the subject's earlier questions have left the
	// window of every limit it reached.
	ResetsAt time.Time
}

// DecideQuestion decides whether a subject that asked at the given times may
// ask another at now, under l. The times may be in any order, and times in
// the future or older than a day are ignored. Windows roll: the hour's limit
// holds for any 60 minutes, so a burst at the turn of an hour gains nothing,
// and the answer to "when" is exact: the moment the oldest question that
// still counts leaves the window.
func DecideQuestion(times []time.Time, now time.Time, l QuestionLimits) QuestionVerdict {
	recent := questionsInDay(times, now)
	var resets time.Time
	for _, w := range []questionWindow{{l.PerHour, hourWindow}, {l.PerDay, dayWindow}} {
		if at := w.reachedUntil(recent, now); at.After(resets) {
			resets = at
		}
	}
	return QuestionVerdict{Allowed: resets.IsZero(), ResetsAt: resets}
}

// questionsInDay returns the times of times that count at now: not in the
// future and less than a day old, newest first.
func questionsInDay(times []time.Time, now time.Time) []time.Time {
	recent := make([]time.Time, 0, len(times))
	for _, t := range times {
		if !t.After(now) && now.Sub(t) < dayWindow {
			recent = append(recent, t)
		}
	}
	slices.SortFunc(recent, func(a, b time.Time) int { return b.Compare(a) })
	return recent
}

// questionWindow is one limit: how many questions fit in a rolling length of
// time. A limit of zero or less is no limit.
type questionWindow struct {
	limit  int
	length time.Duration
}

// reachedUntil returns the moment a question that does not fit in w fits again,
// or the zero time when one fits now. recent is newest first. The question that
// must leave the window for one more to fit is the limit-th newest: after it
// goes, limit-1 remain.
func (w questionWindow) reachedUntil(recent []time.Time, now time.Time) time.Time {
	if w.limit <= 0 {
		return time.Time{}
	}
	inWindow := 0
	for _, t := range recent {
		if now.Sub(t) < w.length {
			inWindow++
		}
	}
	if inWindow < w.limit {
		return time.Time{}
	}
	return recent[w.limit-1].Add(w.length)
}
