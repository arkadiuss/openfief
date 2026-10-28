package ingest

import (
	"errors"
	"fmt"
	"time"
)

// ErrInvalidSyncPolicy reports a sync policy without a positive interval and
// first retry delay, with a maximum retry delay below the first, or with a
// negative jitter.
var ErrInvalidSyncPolicy = errors.New("invalid sync policy")

// SyncPolicy is when the core syncs each connector instance (D-49): every
// Interval after a successful sync, plus up to Jitter so that instances
// don't all sync at once, and after a failed one again after a delay that
// starts at RetryFirst and doubles with each consecutive failure, up to
// RetryMax.
type SyncPolicy struct {
	Interval   time.Duration
	Jitter     time.Duration
	RetryFirst time.Duration
	RetryMax   time.Duration
}

// Validate reports whether p can schedule syncs.
func (p *SyncPolicy) Validate() error {
	switch {
	case p.Interval <= 0:
		return fmt.Errorf("%w: interval %s is not positive", ErrInvalidSyncPolicy, p.Interval)
	case p.Jitter < 0:
		return fmt.Errorf("%w: jitter %s is negative", ErrInvalidSyncPolicy, p.Jitter)
	case p.RetryFirst <= 0:
		return fmt.Errorf("%w: first retry %s is not positive", ErrInvalidSyncPolicy, p.RetryFirst)
	case p.RetryMax < p.RetryFirst:
		return fmt.Errorf("%w: maximum retry %s is below the first %s", ErrInvalidSyncPolicy, p.RetryMax, p.RetryFirst)
	}
	return nil
}

// NextSync returns when an instance's next sync is due, after an attempt
// that ended at ended and left failures consecutive failed attempts: zero
// after a success, which waits Interval plus jitter, clamped to [0, Jitter];
// otherwise RetryFirst doubled for each failure after the first, at most
// RetryMax, with no jitter.
func (p *SyncPolicy) NextSync(ended time.Time, failures int, jitter time.Duration) time.Time {
	if failures <= 0 {
		return ended.Add(p.Interval + min(max(jitter, 0), p.Jitter))
	}
	delay := p.RetryFirst
	for i := 1; i < failures && delay < p.RetryMax; i++ {
		delay *= 2
	}
	return ended.Add(min(delay, p.RetryMax))
}
