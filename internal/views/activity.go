package views

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// ActivityKind is what happened, as the overview's recent activity lists it
// (D-193).
type ActivityKind int

// The kinds of activity. The zero ActivityKind is none.
const (
	// ActivityImport is an import batch the owner started that completed,
	// such as an uploaded statement or a manual entry.
	ActivityImport ActivityKind = iota + 1
	// ActivitySync is a connector sync whose batch completed.
	ActivitySync
	// ActivitySyncFailed is a connector sync attempt that failed.
	ActivitySyncFailed
	// ActivityCoupon is a coupon booked: the first version of an interest
	// transaction (D-101).
	ActivityCoupon
	// ActivityCorrection is a correction: a later version of a transaction,
	// a deletion included (D-18, D-100).
	ActivityCorrection
)

// ActivitySource is where an activity comes from; each is paged on its own.
type ActivitySource int

// The sources of activity, in the order activities at the same time are
// listed.
const (
	// SourceBatches is the completed import batches (ActivityImport,
	// ActivitySync).
	SourceBatches ActivitySource = iota
	// SourceFailures is the failed sync attempts (ActivitySyncFailed).
	SourceFailures
	// SourceChanges is the ledger's coupons and corrections.
	SourceChanges
	activitySources
)

// Activity is one entry of the recent activity, newest first by At. Exactly
// one of Batch, Attempt and Change is set, as Kind says.
type Activity struct {
	Kind ActivityKind
	At   time.Time
	// Batch is a completed batch, for ActivityImport and ActivitySync.
	Batch *ingest.BatchRecord
	// Attempt is a failed sync attempt, for ActivitySyncFailed.
	Attempt *ingest.SyncAttempt
	// Change is a ledger version, with the row ID that orders versions
	// recorded at the same time, for ActivityCoupon and ActivityCorrection.
	Change   *ledger.VersionedTransaction
	ChangeID int64
	// Hidden is a change in an account the caller may not view, or a batch
	// that booked only into such accounts (D-163): it
	// is paged past but not listed.
	Hidden bool
}

// BatchActivity returns the activity of a completed batch: a sync, or an
// import of any other source.
func BatchActivity(b *ingest.BatchRecord) Activity {
	kind := ActivityImport
	if b.Source == "sync" {
		kind = ActivitySync
	}
	return Activity{Kind: kind, At: b.Finished, Batch: b}
}

// ScopeBatch returns b as a caller who may view visible, those of its
// accounts in their order, sees it (D-199): naming only those accounts,
// with only their counts. A caller who may view every account it booked
// into sees its totals; one who may view only some sees the sums of their
// counts, and none of its left-out rows, which are tied to no account. A
// batch that booked into no account is unchanged.
func ScopeBatch(b *ingest.BatchRecord, visible []ledger.AccountID) ingest.BatchRecord {
	out := *b
	if len(b.Accounts) == 0 {
		return out
	}
	out.Accounts = visible
	out.PerAccount = nil
	for _, c := range b.PerAccount {
		if slices.Contains(visible, c.Account) {
			out.PerAccount = append(out.PerAccount, c)
		}
	}
	if len(visible) == len(b.Accounts) {
		return out
	}
	out.Transactions, out.New, out.NewVersions, out.Unchanged, out.FailedAssertions = 0, 0, 0, 0, 0
	out.Problems, out.Unsupported = 0, 0
	for _, c := range out.PerAccount {
		out.Transactions += c.Transactions
		out.New += c.New
		out.NewVersions += c.NewVersions
		out.Unchanged += c.Unchanged
		out.FailedAssertions += c.FailedAssertions
	}
	return out
}

// FailureActivity returns the activity of a failed sync attempt.
func FailureActivity(a *ingest.SyncAttempt) Activity {
	return Activity{Kind: ActivitySyncFailed, At: a.Ended, Attempt: a}
}

// ChangeActivity returns the activity of a ledger version recorded at the
// time at with row ID id: a correction for a version after the first, otherwise a
// coupon.
func ChangeActivity(v *ledger.VersionedTransaction, id int64, at time.Time) Activity {
	kind := ActivityCoupon
	if v.Version > 1 {
		kind = ActivityCorrection
	}
	return Activity{Kind: kind, At: at, Change: v, ChangeID: id}
}

// bound returns the position a takes in its source's order.
func (a *Activity) bound() ActivityBound {
	switch {
	case a.Batch != nil:
		return ActivityBound{At: a.At, Key: string(a.Batch.ID)}
	case a.Attempt != nil:
		return ActivityBound{At: a.At, ID: a.Attempt.ID}
	}
	return ActivityBound{At: a.At, ID: a.ChangeID}
}

// ActivityBound is where a page of one source of activity ended: the time
// and the ID or key of its last entry. The zero bound is the start.
type ActivityBound struct {
	At  time.Time
	ID  int64
	Key string
}

// IsZero reports whether b is the start, before any entry.
func (b ActivityBound) IsZero() bool { return b.At.IsZero() }

// ActivityCursor is where a page of the activity ended, in each source.
type ActivityCursor [activitySources]ActivityBound

// activityCursorJSON is an encoded bound: the time in microseconds since
// the epoch, as PostgreSQL keeps it, the ID and the key.
type activityCursorJSON struct {
	At  int64  `json:"t,omitempty"`
	ID  int64  `json:"i,omitempty"`
	Key string `json:"k,omitempty"`
}

// Encode returns c as an opaque, URL-safe string that ParseActivityCursor
// reads back.
func (c *ActivityCursor) Encode() string {
	out := make([]activityCursorJSON, len(c))
	for i, b := range c {
		if !b.IsZero() {
			out[i] = activityCursorJSON{At: b.At.UnixMicro(), ID: b.ID, Key: b.Key}
		}
	}
	data, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

// ParseActivityCursor reads a cursor that Encode wrote; anything else is an
// error wrapping ErrInvalidCursor.
func ParseActivityCursor(s string) (ActivityCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ActivityCursor{}, fmt.Errorf("%w: %w", ErrInvalidCursor, err)
	}
	var in []activityCursorJSON
	if jsonErr := json.Unmarshal(data, &in); jsonErr != nil {
		return ActivityCursor{}, fmt.Errorf("%w: %w", ErrInvalidCursor, jsonErr)
	}
	var c ActivityCursor
	if len(in) != len(c) {
		return ActivityCursor{}, fmt.Errorf("%w: %d sources", ErrInvalidCursor, len(in))
	}
	for i, b := range in {
		if b.At != 0 {
			c[i] = ActivityBound{At: time.UnixMicro(b.At).UTC(), ID: b.ID, Key: b.Key}
		}
	}
	return c, nil
}

// Encode returns b as an opaque, URL-safe string that ParseActivityBound
// reads back: the page token of one source's pages, such as an instance's
// sync attempts.
func (b ActivityBound) Encode() string {
	c := ActivityCursor{SourceFailures: b}
	return c.Encode()
}

// ParseActivityBound reads a bound that ActivityBound.Encode wrote;
// anything else is an error wrapping ErrInvalidCursor.
func ParseActivityBound(s string) (ActivityBound, error) {
	c, err := ParseActivityCursor(s)
	if err != nil {
		return ActivityBound{}, err
	}
	if c[SourceFailures].IsZero() {
		return ActivityBound{}, fmt.Errorf("%w: no bound", ErrInvalidCursor)
	}
	return c[SourceFailures], nil
}

// ActivityPage is one page of the activity, newest first, and where the
// next starts; Next is nil on the last page.
type ActivityPage struct {
	Activities []Activity
	Next       *ActivityCursor
}

// MergeActivity returns the page of at most size activities that follows
// after, newest first, from fetched: what each source holds after its bound
// in after, newest first, each source's list holding at most fetch entries
// and fewer only when the source has no more. Activities at the same time
// come in source order. A hidden activity is paged past but not listed.
// The merge stops where a source's fetched entries run out while it may
// have more, so nothing is skipped; the next page continues from there. A
// size outside 1..MaxPageSize is an error.
func MergeActivity(after *ActivityCursor, fetched *[activitySources][]Activity, fetch, size int) (ActivityPage, error) {
	if size < 1 || size > MaxPageSize {
		return ActivityPage{}, fmt.Errorf("%w: %d", ErrInvalidPageSize, size)
	}
	next := *after
	var used [activitySources]int
	var out []Activity
	for len(out) < size {
		pick, ok := nextSource(fetched, &used, fetch)
		if !ok {
			break
		}
		a := fetched[pick][used[pick]]
		used[pick]++
		next[pick] = a.bound()
		if !a.Hidden {
			out = append(out, a)
		}
	}
	page := ActivityPage{Activities: out}
	for s := range fetched {
		if used[s] < len(fetched[s]) || len(fetched[s]) >= fetch {
			page.Next = &next
			break
		}
	}
	return page, nil
}

// nextSource returns the source whose next unused activity is the newest,
// or false when none has one or when a source that may have more has run
// out of fetched ones.
func nextSource(fetched *[activitySources][]Activity, used *[activitySources]int, fetch int) (ActivitySource, bool) {
	best, found := ActivitySource(0), false
	for s := range activitySources {
		if used[s] == len(fetched[s]) {
			if len(fetched[s]) >= fetch {
				return 0, false
			}
			continue
		}
		if !found || fetched[s][used[s]].At.After(fetched[best][used[best]].At) {
			best, found = s, true
		}
	}
	return best, found
}
