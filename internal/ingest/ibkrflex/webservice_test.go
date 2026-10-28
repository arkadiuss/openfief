package ibkrflex_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/areknoster/hypert"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
)

// recordMode stays false: there are no IBKR paper credentials, and the
// recordings in testdata/TestFetchStatement* are synthetic, written by hand
// in hypert's format. Never record against a real account (GR-12).
const recordMode = false

// syntheticFlexT is a fake Flex Web Service token, sent as the t query
// parameter; hypert replaces it with SANITIZED before comparing requests
// with the recordings.
const syntheticFlexT = "fake-flex-token-0000"

// replayClient returns a Flex client that replays this test's recordings.
func replayClient(t *testing.T, waits *[]int) *ibkrflex.Client {
	t.Helper()
	httpClient := hypert.TestClient(t, recordMode,
		hypert.WithRequestSanitizer(hypert.ComposedRequestSanitizer(
			hypert.DefaultRequestSanitizer(),
			hypert.SanitizerQueryParams("t"),
		)),
		hypert.WithRequestValidator(hypert.ComposedRequestValidator(
			hypert.PathValidator(),
			hypert.MethodValidator(),
			hypert.QueryParamsValidator(),
			hypert.SchemeValidator(),
			userAgentThenHeaders(),
		)),
	)
	return &ibkrflex.Client{
		HTTP:    httpClient,
		Token:   syntheticFlexT,
		QueryID: "000001",
		Wait: func(_ context.Context, attempt int) error {
			*waits = append(*waits, attempt)
			return nil
		},
	}
}

// userAgentThenHeaders checks the User-Agent against the recording, then the
// other headers. hypert's HeadersValidator drops User-Agent from the
// recording only, so a client that sets its own would never replay.
func userAgentThenHeaders() hypert.RequestValidatorFunc {
	headers := hypert.HeadersValidator()
	return hypert.RequestValidatorFunc(func(t hypert.T, recorded, got hypert.RequestData) error {
		if want, have := recorded.Headers.Get("User-Agent"), got.Headers.Get("User-Agent"); want != have {
			t.Errorf("User-Agent = %q, want %q", have, want)
		}
		got.Headers = got.Headers.Clone()
		got.Headers.Del("User-Agent")
		if err := headers.Validate(t, recorded, got); err != nil {
			return fmt.Errorf("headers: %w", err)
		}
		return nil
	})
}

func TestFetchStatement(t *testing.T) {
	t.Parallel()
	var waits []int
	statement, err := replayClient(t, &waits).FetchStatement(context.Background())
	if err != nil {
		t.Fatalf("FetchStatement: %v", err)
	}
	if len(waits) != 2 || waits[0] != 1 || waits[1] != 2 {
		t.Errorf("waited before attempts %v, want [1 2]", waits)
	}
	m, err := testMapper().Map(bytes.NewReader(statement), ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("Map of the fetched statement: %v", err)
	}
	if len(m.Transactions) != 1 || len(m.Assertions) != 1 || len(m.Problems) != 0 {
		t.Errorf("fetched statement maps to %s, want one buy and one cash assertion", render(&m))
	}
}

func TestFetchStatementTokenInvalid(t *testing.T) {
	t.Parallel()
	var waits []int
	_, err := replayClient(t, &waits).FetchStatement(context.Background())
	if !errors.Is(err, ibkrflex.ErrFlexService) || !strings.Contains(err.Error(), "1015") {
		t.Errorf("FetchStatement error = %v, want ErrFlexService with code 1015", err)
	}
	if len(waits) != 0 {
		t.Errorf("waited %v after a failed SendRequest", waits)
	}
}

// roundTripFunc is an http.RoundTripper made of a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// xmlResponse returns a 200 response with body.
func xmlResponse(r *http.Request, body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}

const (
	sentOK     = `<FlexStatementResponse><Status>Success</Status><ReferenceCode>1111111111</ReferenceCode></FlexStatementResponse>`
	inProgress = `<FlexStatementResponse><Status>Warn</Status><ErrorCode>1019</ErrorCode><ErrorMessage>Statement generation in progress.</ErrorMessage></FlexStatementResponse>`
)

var errNetwork = errors.New("connection refused")

func TestFetchStatementFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		statement func(r *http.Request) (*http.Response, error)
		maxBytes  int64
		wantErr   error
	}{
		{
			name:      "still generating after the last attempt",
			statement: func(r *http.Request) (*http.Response, error) { return xmlResponse(r, inProgress), nil },
			wantErr:   ibkrflex.ErrStatementNotReady,
		},
		{
			name: "error code that is not worth retrying",
			statement: func(r *http.Request) (*http.Response, error) {
				return xmlResponse(r, `<FlexStatementResponse><Status>Fail</Status><ErrorCode>1012</ErrorCode></FlexStatementResponse>`), nil
			},
			wantErr: ibkrflex.ErrFlexService,
		},
		{
			name:      "network failure",
			statement: func(*http.Request) (*http.Response, error) { return nil, errNetwork },
			wantErr:   errNetwork,
		},
		{
			name: "HTTP error status",
			statement: func(r *http.Request) (*http.Response, error) {
				resp := xmlResponse(r, "")
				resp.StatusCode = http.StatusServiceUnavailable
				return resp, nil
			},
			wantErr: ibkrflex.ErrFlexService,
		},
		{
			name:      "unreadable response",
			statement: func(r *http.Request) (*http.Response, error) { return xmlResponse(r, "<html>"), nil },
			wantErr:   ibkrflex.ErrFlexService,
		},
		{
			name: "statement over the size cap",
			statement: func(r *http.Request) (*http.Response, error) {
				return xmlResponse(r, statementXML(strings.Repeat(" ", 2048))), nil
			},
			maxBytes: 1024,
			wantErr:  ibkrflex.ErrTooLarge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &ibkrflex.Client{
				HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					if strings.HasSuffix(r.URL.Path, "/SendRequest") {
						return xmlResponse(r, sentOK), nil
					}
					return tt.statement(r)
				})},
				Token:       syntheticFlexT,
				QueryID:     "000001",
				MaxAttempts: 3,
				MaxBytes:    tt.maxBytes,
				Wait:        func(context.Context, int) error { return nil },
			}
			_, err := c.FetchStatement(context.Background())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("FetchStatement error = %v, want %v", err, tt.wantErr)
			}
			if strings.Contains(err.Error(), syntheticFlexT) {
				t.Errorf("error %q leaks the token", err)
			}
		})
	}
}

func TestFetchStatementStopsWhenCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := &ibkrflex.Client{
		HTTP: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return xmlResponse(r, sentOK), nil
		})},
		Token:   syntheticFlexT,
		QueryID: "000001",
	}
	_, err := c.FetchStatement(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("FetchStatement error = %v, want context.Canceled", err)
	}
}
