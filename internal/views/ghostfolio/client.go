package ghostfolio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// SupportedVersions is the range of Ghostfolio versions the view supports
// (D-64): the versions its contract tests' recordings come from. The
// accounts, activities and import endpoints are Ghostfolio's internal API,
// renamed within /v1 without notice (R-15), so a version outside the range
// may not work. The account balance endpoints, which the view uses since
// D-108, are internal too.
const SupportedVersions = ">=3.93.0 <3.94.0"

// Client errors.
var (
	// ErrHTTP reports a response with a status other than success; it is
	// wrapped by an *HTTPError.
	ErrHTTP = errors.New("error response from Ghostfolio")
	// ErrNoAccountFilter reports a call that would act on every account:
	// listing or bulk-deleting activities with no account to filter by.
	// Ghostfolio's bulk delete with no filter deletes every activity the
	// user has (D-64), so the client never sends one.
	ErrNoAccountFilter = errors.New("no account filter")
	// ErrBadResponse reports a response body the client can't read.
	ErrBadResponse = errors.New("unreadable Ghostfolio response")
)

// HTTPError is a Ghostfolio response with an error status. It names the
// call and the status only: Ghostfolio's message can quote the data sent.
type HTTPError struct {
	Op     string
	Status int
}

// Error names the call and the status.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s: HTTP %d", ErrHTTP, e.Op, e.Status)
}

// Unwrap returns ErrHTTP.
func (*HTTPError) Unwrap() error { return ErrHTTP }

// maxResponse is the largest response body the client reads.
const maxResponse = 64 << 20

// TokenSource returns Ghostfolio's security token, read from the plugin's
// Secret on every call so a rotated token is picked up.
type TokenSource func() (string, error)

// Client calls Ghostfolio's HTTP API at a base URL, such as
// https://ghostfolio.example.com, with the bearer token Login got.
type Client struct {
	http    *http.Client
	base    string
	token   TokenSource
	session string
}

// NewClient returns a client for the Ghostfolio at base, logging in with the
// security token from token.
func NewClient(hc *http.Client, base string, token TokenSource) *Client {
	return &Client{http: hc, base: strings.TrimSuffix(base, "/"), token: token}
}

// User is the part of Ghostfolio's user the view reads.
type User struct {
	// ID is the user's ID; the view's MANUAL assets' symbols derive from it.
	ID string
	// BaseCurrency is the currency the view's new accounts get.
	BaseCurrency kernel.Currency
	// RestrictedView is the setting that blanks comments, fees and
	// quantities in responses.
	RestrictedView bool
}

// Login exchanges the security token for a bearer token, used by every
// later call.
func (c *Client) Login(ctx context.Context) error {
	token, err := c.token()
	if err != nil {
		return fmt.Errorf("reading the security token: %w", err)
	}
	var out struct {
		AuthToken string `json:"authToken"`
	}
	body := struct {
		AccessToken string `json:"accessToken"`
	}{AccessToken: token}
	if err := c.do(ctx, "log in", http.MethodPost, "/api/v1/auth/anonymous", nil, body, &out); err != nil {
		return err
	}
	if out.AuthToken == "" {
		return fmt.Errorf("%w: log in: no token", ErrBadResponse)
	}
	c.session = out.AuthToken
	return nil
}

// User returns the user's ID, base currency and restricted view setting.
func (c *Client) User(ctx context.Context) (User, error) {
	var out struct {
		ID       string `json:"id"`
		Settings struct {
			BaseCurrency     string `json:"baseCurrency"`
			IsRestrictedView bool   `json:"isRestrictedView"`
		} `json:"settings"`
	}
	if err := c.do(ctx, "read the user", http.MethodGet, "/api/v1/user", nil, nil, &out); err != nil {
		return User{}, err
	}
	cur, err := kernel.ParseCurrency(out.Settings.BaseCurrency)
	if err != nil {
		return User{}, fmt.Errorf("%w: base currency: %w", ErrBadResponse, err)
	}
	if out.ID == "" {
		return User{}, fmt.Errorf("%w: read the user: no ID", ErrBadResponse)
	}
	return User{ID: out.ID, BaseCurrency: cur, RestrictedView: out.Settings.IsRestrictedView}, nil
}

// Balances returns the cash balances of the given account, from its
// balance history (GET /api/v1/account/{id}/balances).
func (c *Client) Balances(ctx context.Context, account string) ([]Balance, error) {
	var out struct {
		Balances []struct {
			AccountID string      `json:"accountId"`
			Date      string      `json:"date"`
			ID        string      `json:"id"`
			Value     json.Number `json:"value"`
		} `json:"balances"`
	}
	if err := c.do(ctx, "list balances", http.MethodGet, "/api/v1/account/"+url.PathEscape(account)+"/balances", nil, nil, &out); err != nil {
		return nil, err
	}
	bs := make([]Balance, 0, len(out.Balances))
	for _, w := range out.Balances {
		d, err := wireDateOf(w.Date)
		if err != nil {
			return nil, fmt.Errorf("%w: balance date: %w", ErrBadResponse, err)
		}
		v, err := kernel.ParseDecimal(w.Value.String())
		if err != nil {
			return nil, fmt.Errorf("%w: balance: %w", ErrBadResponse, err)
		}
		bs = append(bs, Balance{ID: w.ID, AccountID: w.AccountID, Date: d, Value: v})
	}
	return bs, nil
}

// SetBalance sets an owned account's cash balance on b's date, replacing
// the one that date has (POST /api/v1/account-balance).
func (c *Client) SetBalance(ctx context.Context, b *Balance) error {
	body := struct {
		AccountID string      `json:"accountId"`
		Balance   json.Number `json:"balance"`
		Date      string      `json:"date"`
	}{AccountID: b.AccountID, Balance: json.Number(b.Value.String()), Date: wireDate(b.Date)}
	return c.do(ctx, "set a balance", http.MethodPost, "/api/v1/account-balance", nil, body, nil)
}

// DeleteBalance deletes one cash balance by its ID.
func (c *Client) DeleteBalance(ctx context.Context, id string) error {
	return c.do(ctx, "delete a balance", http.MethodDelete, "/api/v1/account-balance/"+url.PathEscape(id), nil, nil, nil)
}

// Accounts returns the IDs of every account the user has.
func (c *Client) Accounts(ctx context.Context) ([]string, error) {
	var out struct {
		Accounts []struct {
			ID string `json:"id"`
		} `json:"accounts"`
	}
	if err := c.do(ctx, "list accounts", http.MethodGet, "/api/v1/account", nil, nil, &out); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(out.Accounts))
	for _, a := range out.Accounts {
		ids = append(ids, a.ID)
	}
	return ids, nil
}

// CreateAccount creates a with the ID the view chose.
func (c *Client) CreateAccount(ctx context.Context, a Account) error {
	body := struct {
		Currency   string  `json:"currency"`
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		PlatformID *string `json:"platformId"`
	}{Currency: a.Currency.String(), ID: a.ID, Name: a.Name}
	return c.do(ctx, "create an account", http.MethodPost, "/api/v1/account", nil, body, nil)
}

// wireActivity is an activity as Ghostfolio's activities and import
// responses hold it.
type wireActivity struct {
	ID           string      `json:"id"`
	AccountID    *string     `json:"accountId"`
	Comment      *string     `json:"comment"`
	Currency     *string     `json:"currency"`
	Date         string      `json:"date"`
	Fee          json.Number `json:"fee"`
	Quantity     json.Number `json:"quantity"`
	Type         string      `json:"type"`
	UnitPrice    json.Number `json:"unitPrice"`
	AssetProfile struct {
		Currency   *string `json:"currency"`
		DataSource string  `json:"dataSource"`
		Symbol     string  `json:"symbol"`
	} `json:"assetProfile"`
	Error *struct {
		Code string `json:"code"`
	} `json:"error"`
}

// activity converts w.
func (w *wireActivity) activity() (Activity, error) {
	a := Activity{
		ID: w.ID, AccountID: deref(w.AccountID), Comment: deref(w.Comment), Type: ActivityType(w.Type),
		DataSource: w.AssetProfile.DataSource, Symbol: w.AssetProfile.Symbol,
	}
	var err error
	if a.Date, err = wireDateOf(w.Date); err != nil {
		return Activity{}, fmt.Errorf("%w: activity date: %w", ErrBadResponse, err)
	}
	code := deref(w.Currency)
	if code == "" {
		code = deref(w.AssetProfile.Currency)
	}
	if a.Currency, err = kernel.ParseCurrency(code); err != nil {
		return Activity{}, fmt.Errorf("%w: activity currency: %w", ErrBadResponse, err)
	}
	for _, f := range []struct {
		dst *kernel.Decimal
		src json.Number
	}{{&a.Fee, w.Fee}, {&a.Quantity, w.Quantity}, {&a.UnitPrice, w.UnitPrice}} {
		if f.src == "" {
			continue
		}
		if *f.dst, err = kernel.ParseDecimal(f.src.String()); err != nil {
			return Activity{}, fmt.Errorf("%w: activity number: %w", ErrBadResponse, err)
		}
	}
	return a, nil
}

// Activities returns every activity in the given accounts. It refuses an
// empty list, which Ghostfolio reads as every account.
func (c *Client) Activities(ctx context.Context, accounts []string) ([]Activity, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("list activities: %w", ErrNoAccountFilter)
	}
	var out struct {
		Activities []wireActivity `json:"activities"`
	}
	q := url.Values{"accounts": {strings.Join(accounts, ",")}}
	if err := c.do(ctx, "list activities", http.MethodGet, "/api/v1/activities", q, nil, &out); err != nil {
		return nil, err
	}
	acts := make([]Activity, 0, len(out.Activities))
	for i := range out.Activities {
		a, err := out.Activities[i].activity()
		if err != nil {
			return nil, err
		}
		acts = append(acts, a)
	}
	return acts, nil
}

// Import imports b, or with dryRun only validates it and returns the
// preview (D-64). Ghostfolio skips a duplicate silently and doesn't roll
// back a failed import, so every caller checks a dry run first.
func (c *Client) Import(ctx context.Context, b *Batch, dryRun bool) ([]ImportedActivity, error) {
	body := importBody(b)
	var out struct {
		Activities []wireActivity `json:"activities"`
	}
	q := url.Values{"dryRun": {strconv.FormatBool(dryRun)}}
	op := "import"
	if dryRun {
		op = "dry-run an import"
	}
	if err := c.do(ctx, op, http.MethodPost, "/api/v1/import", q, body, &out); err != nil {
		return nil, err
	}
	res := make([]ImportedActivity, 0, len(out.Activities))
	for i := range out.Activities {
		a, err := out.Activities[i].activity()
		if err != nil {
			return nil, err
		}
		res = append(res, ImportedActivity{Activity: a, Duplicate: out.Activities[i].Error != nil})
	}
	return res, nil
}

// DeleteActivity deletes one activity; one already gone counts as deleted.
func (c *Client) DeleteActivity(ctx context.Context, id string) error {
	err := c.do(ctx, "delete an activity", http.MethodDelete, "/api/v1/activities/"+url.PathEscape(id), nil, nil, nil)
	if he, ok := errors.AsType[*HTTPError](err); ok && he.Status == http.StatusNotFound {
		return nil
	}
	return err
}

// DeleteActivitiesInAccounts deletes every activity in the given accounts
// with one bulk delete. It refuses an empty list: without an account
// filter Ghostfolio deletes every activity the user has (D-64).
func (c *Client) DeleteActivitiesInAccounts(ctx context.Context, accounts []string) error {
	if len(accounts) == 0 {
		return fmt.Errorf("bulk delete: %w", ErrNoAccountFilter)
	}
	q := url.Values{"accounts": {strings.Join(accounts, ",")}}
	return c.do(ctx, "bulk delete", http.MethodDelete, "/api/v1/activities", q, nil, nil)
}

// importRequest is the body of Ghostfolio's import.
type importRequest struct {
	Activities    []importActivity `json:"activities"`
	AssetProfiles []importAsset    `json:"assetProfiles,omitempty"`
}

// importActivity is an activity as Ghostfolio's import takes it. Numbers
// are JSON numbers with the activity's exact digits.
type importActivity struct {
	AccountID  string      `json:"accountId"`
	Comment    string      `json:"comment"`
	Currency   string      `json:"currency"`
	DataSource string      `json:"dataSource"`
	Date       string      `json:"date"`
	Fee        json.Number `json:"fee"`
	Quantity   json.Number `json:"quantity"`
	Symbol     string      `json:"symbol"`
	Type       string      `json:"type"`
	UnitPrice  json.Number `json:"unitPrice"`
}

// importAsset is a MANUAL asset profile with its market data, as
// Ghostfolio's import takes it.
type importAsset struct {
	AssetClass    string        `json:"assetClass,omitempty"`
	AssetSubClass string        `json:"assetSubClass,omitempty"`
	Currency      string        `json:"currency"`
	DataSource    string        `json:"dataSource"`
	ISIN          string        `json:"isin,omitempty"`
	MarketData    []importPrice `json:"marketData"`
	Name          string        `json:"name"`
	Symbol        string        `json:"symbol"`
}

// importPrice is one day of an asset's market data.
type importPrice struct {
	Date        string      `json:"date"`
	MarketPrice json.Number `json:"marketPrice"`
}

// importBody is b as Ghostfolio's import takes it.
func importBody(b *Batch) importRequest {
	r := importRequest{Activities: make([]importActivity, 0, len(b.Activities))}
	for i := range b.Activities {
		a := &b.Activities[i]
		r.Activities = append(r.Activities, importActivity{
			AccountID: a.AccountID, Comment: a.Comment, Currency: a.Currency.String(), DataSource: a.DataSource,
			Date: wireDate(a.Date), Fee: json.Number(a.Fee.String()), Quantity: json.Number(a.Quantity.String()),
			Symbol: a.Symbol, Type: string(a.Type), UnitPrice: json.Number(a.UnitPrice.String()),
		})
	}
	for i := range b.Assets {
		as := &b.Assets[i]
		md := make([]importPrice, 0, len(as.MarketData))
		for _, p := range as.MarketData {
			md = append(md, importPrice{Date: wireDate(p.Date), MarketPrice: json.Number(p.Price.String())})
		}
		r.AssetProfiles = append(r.AssetProfiles, importAsset{
			AssetClass: as.AssetClass, AssetSubClass: as.AssetSubClass, Currency: as.Currency.String(),
			DataSource: DataSourceManual, ISIN: as.ISIN, MarketData: md, Name: as.Name, Symbol: as.Symbol,
		})
	}
	return r
}

// wireDate is d at midnight UTC in ISO 8601, as Ghostfolio returns dates.
func wireDate(d kernel.Date) string { return d.String() + "T00:00:00.000Z" }

// wireDateOf is the UTC date of s, a time in ISO 8601 as Ghostfolio
// returns it.
func wireDateOf(s string) (kernel.Date, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return kernel.Date{}, fmt.Errorf("parsing %q: %w", s, err)
	}
	t = t.UTC()
	d, err := kernel.NewDate(t.Year(), t.Month(), t.Day())
	if err != nil {
		return kernel.Date{}, fmt.Errorf("parsing %q: %w", s, err)
	}
	return d, nil
}

// do sends one request and decodes a JSON response into out, unless out is
// nil. Any status but 2xx is an *HTTPError.
func (c *Client) do(ctx context.Context, op, method, path string, q url.Values, body, out any) error {
	req, err := c.request(ctx, method, path, q, body)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	closeErr := resp.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return fmt.Errorf("%s: reading: %w", op, err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return &HTTPError{Op: op, Status: resp.StatusCode}
	}
	if out == nil {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrBadResponse, op, err)
	}
	return nil
}

// request builds a request to path with query q and, unless body is nil,
// body as JSON, carrying the bearer token once Login has one.
func (c *Client) request(ctx context.Context, method, path string, q url.Values, body any) (*http.Request, error) {
	u := c.base + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	var rd io.Reader = http.NoBody
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding: %w", err)
		}
		rd = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.session != "" {
		req.Header.Set("Authorization", "Bearer "+c.session)
	}
	return req, nil
}

// deref returns *s, or "" for nil.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
