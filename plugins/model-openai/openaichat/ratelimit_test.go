package openaichat

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// A 429 is retried once, after Retry-After capped at ten seconds (D-207); every
// other status is not retried. No test sleeps: the clock and the wait are the
// test's.

// rig is what a test of the retry looks at: the bodies the server got and the
// waits the client asked for.
type rig struct {
	mu     sync.Mutex
	bodies []string
	waits  []time.Duration
}

type step struct {
	status     int
	retryAfter string
	reply      string
}

// clockNow is the fixed time of the tests' clock.
func clockNow() time.Time { return time.Date(2026, time.November, 14, 12, 0, 0, 0, time.UTC) }

// scriptedClient returns a client of a server that answers its requests in
// order with steps (the last one again when they run out), and the rig that
// records the requests and the waits; the client waits for no time.
func scriptedClient(t *testing.T, steps ...step) (*rig, *Client) {
	t.Helper()
	r := &rig{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "unreadable request", http.StatusBadRequest)
			return
		}
		r.mu.Lock()
		i := len(r.bodies)
		r.bodies = append(r.bodies, string(raw))
		r.mu.Unlock()
		st := steps[min(i, len(steps)-1)]
		if st.retryAfter != "" {
			w.Header().Set("Retry-After", st.retryAfter)
		}
		w.WriteHeader(st.status)
		if _, err := io.WriteString(w, st.reply); err != nil {
			return // the client went away; the test sees its own error
		}
	}))
	t.Cleanup(srv.Close)
	wait := func(_ context.Context, d time.Duration) error {
		r.waits = append(r.waits, d)
		return nil
	}
	return r, &Client{
		HTTP: srv.Client(), BaseURL: srv.URL + "/v1", Key: func() (string, error) { return fakeKey, nil },
		Wait: wait, Now: clockNow,
	}
}

func chatRequest() *Request {
	return &Request{Model: "m", Messages: []Message{{Role: "user", Content: "synthetic question"}}}
}

func TestA429IsRetriedOnceAfterTheWaitRetryAfterGives(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		retryAfter string
		want       time.Duration
	}{
		{name: "no header waits a short fixed time", want: time.Second},
		{name: "seconds", retryAfter: "3", want: 3 * time.Second},
		{name: "zero seconds", retryAfter: "0", want: 0},
		{name: "seconds capped at ten", retryAfter: "30", want: 10 * time.Second},
		{name: "seconds beyond a duration are neither form", retryAfter: "99999999999", want: time.Second},
		{name: "an HTTP date", retryAfter: clockNow().Add(4 * time.Second).Format(http.TimeFormat), want: 4 * time.Second},
		{name: "an HTTP date capped at ten", retryAfter: clockNow().Add(time.Hour).Format(http.TimeFormat), want: 10 * time.Second},
		{name: "an HTTP date in the past waits none", retryAfter: clockNow().Add(-time.Minute).Format(http.TimeFormat), want: 0},
		{name: "neither form waits the short fixed time", retryAfter: "soon", want: time.Second},
		{name: "negative seconds are neither form", retryAfter: "-5", want: time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r, c := scriptedClient(t, step{status: 429, retryAfter: tt.retryAfter, reply: `{}`}, step{status: 200, reply: textReply})
			reply, err := c.Complete(t.Context(), chatRequest())
			if err != nil || reply.Content != "synthetic answer" {
				t.Fatalf("reply = %+v, %v; want the answer of the retry", reply, err)
			}
			if diff := cmp.Diff([]time.Duration{tt.want}, r.waits); diff != "" {
				t.Errorf("waits (-want +got):\n%s", diff)
			}
			if len(r.bodies) != 2 || r.bodies[0] != r.bodies[1] {
				t.Errorf("requests = %d, bodies equal %v; want two with the same body", len(r.bodies), len(r.bodies) == 2 && r.bodies[0] == r.bodies[1])
			}
		})
	}
}

func TestATwice429IsAnUnavailableProviderWithTheStatus(t *testing.T) {
	t.Parallel()
	r, c := scriptedClient(t, step{status: 429, retryAfter: "2", reply: `{}`})
	_, err := c.Complete(t.Context(), chatRequest())
	var ce *CallError
	if !errors.Is(err, ErrUnavailable) || !errors.As(err, &ce) || ce.Status != 429 {
		t.Fatalf("error = %#v, want ErrUnavailable with status 429", err)
	}
	if len(r.bodies) != 2 || len(r.waits) != 1 {
		t.Errorf("requests = %d, waits = %v; want two requests and one wait: a 429 is retried once", len(r.bodies), r.waits)
	}
}

func TestOtherStatusesAreNeverRetried(t *testing.T) {
	t.Parallel()
	for _, status := range []int{400, 401, 402, 403, 404, 413, 422, 500, 502, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()
			r, c := scriptedClient(t, step{status: status, retryAfter: "1", reply: `{}`}, step{status: 200, reply: textReply})
			if _, err := c.Complete(t.Context(), chatRequest()); err == nil {
				t.Fatal("no error, want the failure of the first answer")
			}
			if len(r.bodies) != 1 || len(r.waits) != 0 {
				t.Errorf("requests = %d, waits = %v; want one request and no wait", len(r.bodies), r.waits)
			}
		})
	}
}

func TestA429IsNotRetriedWhenTheWaitDoesNotFitInTheDeadline(t *testing.T) {
	t.Parallel()
	r, c := scriptedClient(t, step{status: 429, retryAfter: "5", reply: `{}`}, step{status: 200, reply: textReply})
	c.Now = time.Now
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, err := c.Complete(ctx, chatRequest())
	var ce *CallError
	if !errors.As(err, &ce) || ce.Status != 429 {
		t.Fatalf("error = %#v, want the 429", err)
	}
	if len(r.bodies) != 1 || len(r.waits) != 0 {
		t.Errorf("requests = %d, waits = %v; want one request and no wait past the deadline", len(r.bodies), r.waits)
	}
}

func TestA429IsNotRetriedWhenTheWaitIsCut(t *testing.T) {
	t.Parallel()
	r, c := scriptedClient(t, step{status: 429, reply: `{}`}, step{status: 200, reply: textReply})
	c.Wait = func(context.Context, time.Duration) error { return context.Canceled }
	_, err := c.Complete(t.Context(), chatRequest())
	var ce *CallError
	if !errors.As(err, &ce) || ce.Status != 429 || len(r.bodies) != 1 {
		t.Errorf("error = %#v after %d requests, want the 429 after one", err, len(r.bodies))
	}
}
