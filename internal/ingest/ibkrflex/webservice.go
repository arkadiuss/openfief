package ibkrflex

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ErrFlexService reports a failure of IBKR's Flex Web Service: an error
// response, an unreadable one, or a request that didn't complete.
var ErrFlexService = errors.New("flex web service")

// ErrStatementNotReady reports a statement still being generated after the
// client's last attempt.
var ErrStatementNotReady = errors.New("flex statement not ready")

// The Flex Web Service (version 3) endpoints IBKR documents publicly.
const (
	DefaultSendRequestURL  = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService/SendRequest"
	DefaultGetStatementURL = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService/GetStatement"
)

// Client defaults.
const (
	defaultMaxAttempts = 10
	defaultWaitStep    = 5 * time.Second
	// maxServiceResponse caps a FlexStatementResponse, which is a few lines.
	maxServiceResponse = 64 << 10
	flexVersion        = "3"
	userAgent          = "openfief-ibkr-flex"
)

// retryCodes are the Flex Web Service error codes that mean "try again
// shortly": the statement is being generated, data is not ready yet, the
// server is busy, or the token made too many requests.
func retryCode(code string) bool {
	switch code {
	case "1001", "1004", "1005", "1006", "1007", "1008", "1009", "1018", "1019", "1021":
		return true
	}
	return false
}

// Client pulls a Flex Query statement from IBKR's Flex Web Service, the
// daily-pull ingest path: SendRequest with the token and query ID returns a
// reference code, and GetStatement with that code returns the statement once
// IBKR has generated it. The token is a read-only credential (DESIGN §3.5):
// it is sent only as the documented query parameter, and never appears in an
// error.
type Client struct {
	// HTTP makes the requests; it should reach IBKR through the egress proxy.
	HTTP *http.Client
	// Token is the Flex Web Service token. Secret.
	Token string
	// QueryID is the Flex Query's ID.
	QueryID string
	// SendRequestURL and GetStatementURL default to IBKR's endpoints. The
	// statement URL in IBKR's response is ignored, so the token only ever goes
	// to these two.
	SendRequestURL  string
	GetStatementURL string
	// MaxAttempts caps GetStatement calls; zero means 10.
	MaxAttempts int
	// Wait pauses before GetStatement attempt n (from 1); nil waits 5 s × n,
	// or until ctx is done.
	Wait func(ctx context.Context, attempt int) error
	// MaxBytes caps the statement's size; zero means the parser's default.
	MaxBytes int64
}

// serviceResponse is a FlexStatementResponse: a reference code on success,
// or an error code and message.
type serviceResponse struct {
	XMLName       xml.Name `xml:"FlexStatementResponse"`
	Status        string   `xml:"Status"`
	ReferenceCode string   `xml:"ReferenceCode"`
	ErrorCode     string   `xml:"ErrorCode"`
	ErrorMessage  string   `xml:"ErrorMessage"`
}

// FetchStatement requests the Flex Query and returns its statement, a
// FlexQueryResponse document, retrying while IBKR generates it.
func (c *Client) FetchStatement(ctx context.Context) ([]byte, error) {
	body, err := c.get(ctx, or(c.SendRequestURL, DefaultSendRequestURL), c.QueryID, maxServiceResponse)
	if err != nil {
		return nil, err
	}
	sent, err := parseServiceResponse(body)
	if err != nil {
		return nil, err
	}
	if sent.Status != "Success" || sent.ReferenceCode == "" {
		return nil, fmt.Errorf("%w: SendRequest failed with code %s: %s", ErrFlexService, sent.ErrorCode, sent.ErrorMessage)
	}
	return c.poll(ctx, sent.ReferenceCode)
}

// poll calls GetStatement until it returns the statement.
func (c *Client) poll(ctx context.Context, reference string) ([]byte, error) {
	attempts := c.MaxAttempts
	if attempts <= 0 {
		attempts = defaultMaxAttempts
	}
	limit := c.MaxBytes
	if limit <= 0 {
		limit = defaultMaxBytes
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := c.wait(ctx, attempt); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFlexService, err)
		}
		body, err := c.get(ctx, or(c.GetStatementURL, DefaultGetStatementURL), reference, limit)
		if err != nil {
			return nil, err
		}
		if rootElement(body) == "FlexQueryResponse" {
			return body, nil
		}
		resp, err := parseServiceResponse(body)
		if err != nil {
			return nil, err
		}
		if !retryCode(resp.ErrorCode) {
			return nil, fmt.Errorf("%w: GetStatement failed with code %s: %s", ErrFlexService, resp.ErrorCode, resp.ErrorMessage)
		}
	}
	return nil, fmt.Errorf("%w after %d attempts", ErrStatementNotReady, attempts)
}

// wait pauses before an attempt.
func (c *Client) wait(ctx context.Context, attempt int) error {
	if c.Wait != nil {
		return c.Wait(ctx, attempt)
	}
	timer := time.NewTimer(time.Duration(attempt) * defaultWaitStep)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("waiting for the statement: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

// get calls endpoint with the token and q, and returns at most limit bytes of
// a 200 response. Errors never include the URL, which carries the token.
func (c *Client) get(ctx context.Context, endpoint, q string, limit int64) ([]byte, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w: bad endpoint URL", ErrFlexService)
	}
	u.RawQuery = url.Values{"t": {c.Token}, "q": {q}, "v": {flexVersion}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: building the request failed", ErrFlexService)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return nil, fmt.Errorf("%w: %s %s: %w", ErrFlexService, http.MethodGet, u.Path, err)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if closeErr := resp.Body.Close(); err == nil {
		err = closeErr
	}
	switch {
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: %s %s: HTTP status %d", ErrFlexService, http.MethodGet, u.Path, resp.StatusCode)
	case err != nil:
		return nil, fmt.Errorf("%w: reading %s: %w", ErrFlexService, u.Path, err)
	case int64(len(body)) > limit:
		return nil, fmt.Errorf("%w: %s response over %d bytes", ErrTooLarge, u.Path, limit)
	}
	return body, nil
}

// parseServiceResponse reads a FlexStatementResponse.
func parseServiceResponse(body []byte) (serviceResponse, error) {
	var r serviceResponse
	if err := xml.Unmarshal(body, &r); err != nil {
		return serviceResponse{}, fmt.Errorf("%w: unreadable response: %w", ErrFlexService, err)
	}
	return r, nil
}

// rootElement returns the name of body's first element, or "".
func rootElement(body []byte) string {
	d := xml.NewDecoder(bytes.NewReader(body))
	for {
		tok, err := d.Token()
		if err != nil {
			return ""
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se.Name.Local
		}
	}
}

// or returns s, or def when s is empty.
func or(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
