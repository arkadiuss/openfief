package treasurybonds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// ErrNotDiscovered reports a page or a resource list that doesn't link the
// retail bond spreadsheet on the host it was read from.
var ErrNotDiscovered = errors.New("no link to the retail bond spreadsheet")

// The ministry's names for the spreadsheet (D-97).
const (
	// DatasetTitle is the title the ministry gives the spreadsheet on
	// gov.pl, and the start of its link's text there.
	DatasetTitle = "Dane dotyczące obligacji detalicznych"
	// attachmentPath starts the path of every file gov.pl serves for a page.
	attachmentPath = "/attachment/"
	// xlsFormat is dane.gov.pl's format of the spreadsheet's resource.
	xlsFormat = "xls"
)

// discoverGovPL returns the address of the spreadsheet the gov.pl page at
// page links: the first link to a gov.pl attachment whose text starts with
// DatasetTitle, resolved against page. A link to another host fails with
// ErrNotDiscovered, since the endpoint may call only the one it was read
// from; so does a page without the link.
func discoverGovPL(body io.Reader, page *url.URL) (*url.URL, error) {
	z := html.NewTokenizer(body)
	l := links{page: page}
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if err := z.Err(); !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("reading the page: %w", err)
			}
			return nil, fmt.Errorf("%w on %s", ErrNotDiscovered, page.Redacted())
		}
		if u, ok := l.next(tt, z); ok {
			return u, nil
		}
	}
}

// links follows a page's links, collecting the text of the one it is in.
type links struct {
	page *url.URL
	href string
	text strings.Builder
	in   bool
}

// next takes the token z just read, of type tt, and returns the
// spreadsheet's address when it ends a link to it.
func (l *links) next(tt html.TokenType, z *html.Tokenizer) (*url.URL, bool) {
	switch tt {
	case html.StartTagToken:
		if tok := z.Token(); tok.DataAtom == atom.A {
			l.href, l.in = attr(tok, "href"), true
			l.text.Reset()
		}
	case html.TextToken:
		if l.in {
			l.text.Write(z.Text())
		}
	case html.EndTagToken:
		if tok := z.Token(); tok.DataAtom == atom.A && l.in {
			l.in = false
			return attachment(l.href, l.text.String(), l.page)
		}
	case html.ErrorToken, html.SelfClosingTagToken, html.CommentToken, html.DoctypeToken:
	}
	return nil, false
}

// attr returns tok's attribute key, or "".
func attr(tok html.Token, key string) string {
	for _, a := range tok.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// attachment resolves a link's href against page and reports whether it is
// the spreadsheet: a gov.pl attachment on page's host whose text, its spaces
// collapsed, starts with DatasetTitle.
func attachment(href, text string, page *url.URL) (*url.URL, bool) {
	if !strings.HasPrefix(strings.ToLower(strings.Join(strings.Fields(text), " ")), strings.ToLower(DatasetTitle)) {
		return nil, false
	}
	ref, err := url.Parse(href)
	if err != nil {
		return nil, false
	}
	u := page.ResolveReference(ref)
	if !sameEndpoint(page, u) || !strings.Contains(u.EscapedPath(), attachmentPath) {
		return nil, false
	}
	return u, true
}

// sameEndpoint reports whether u is https on base's host and port.
func sameEndpoint(base, u *url.URL) bool {
	return strings.EqualFold(u.Scheme, "https") && u.User == nil && strings.EqualFold(u.Host, base.Host)
}

// daneResources is the part of dane.gov.pl's resource list of a dataset
// (API 1.4) that names the files. The attributes' keys don't follow Go's
// JSON naming, so they are decoded as a map and read by these keys.
type daneResources struct {
	Data []struct {
		Attributes map[string]any `json:"attributes"`
	} `json:"data"`
}

// The keys of a resource's attributes.
const (
	keyTitle       = "title"
	keyFormat      = "format"
	keyFileURL     = "file_url"
	keyLink        = "link"
	keyDownloadURL = "download_url"
	keyDataDate    = "data_date"
	keyCreated     = "created"
)

// text returns the attribute key of attrs if it is a string, else "".
func text(attrs map[string]any, key string) string {
	s, ok := attrs[key].(string)
	if !ok {
		return ""
	}
	return s
}

// daneFile is the spreadsheet a dane.gov.pl resource list names: where it
// is, its title and the day the ministry made its data.
type daneFile struct {
	url   *url.URL
	title string
	made  kernel.Date
}

// discoverDane returns the first .xls resource of the dataset's resource
// list read from list: its file's address on list's host (the file URL,
// else the link, else the download URL), its title, and the day its data
// was made, its data date or else the day it was created. A list without
// such a resource fails with ErrNotDiscovered.
func discoverDane(body io.Reader, list *url.URL) (daneFile, error) {
	var res daneResources
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		return daneFile{}, fmt.Errorf("%w: resource list: %w", ErrNotDiscovered, err)
	}
	for _, r := range res.Data {
		if f, ok := daneResource(r.Attributes, list); ok {
			return f, nil
		}
	}
	return daneFile{}, fmt.Errorf("%w in the resource list at %s", ErrNotDiscovered, list.Redacted())
}

// daneResource reads one resource's attributes, and reports whether it is
// the spreadsheet: an .xls file on list's host with a date.
func daneResource(attrs map[string]any, list *url.URL) (daneFile, bool) {
	if !strings.EqualFold(text(attrs, keyFormat), xlsFormat) {
		return daneFile{}, false
	}
	made, ok := daneDate(text(attrs, keyDataDate), text(attrs, keyCreated))
	if !ok {
		return daneFile{}, false
	}
	for _, key := range []string{keyFileURL, keyLink, keyDownloadURL} {
		ref, err := url.Parse(text(attrs, key))
		if text(attrs, key) == "" || err != nil {
			continue
		}
		if u := list.ResolveReference(ref); sameEndpoint(list, u) {
			return daneFile{url: u, title: text(attrs, keyTitle), made: made}, true
		}
	}
	return daneFile{}, false
}

// daneDate reads the first of dates that is a date, "YYYY-MM-DD" or an
// RFC 3339 time, whose UTC day it takes.
func daneDate(dates ...string) (kernel.Date, bool) {
	for _, s := range dates {
		if d, err := kernel.ParseDate(s); err == nil {
			return d, true
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			if d, err := kernel.ParseDate(t.UTC().Format(time.DateOnly)); err == nil {
				return d, true
			}
		}
	}
	return kernel.Date{}, false
}
