package treasurybonds_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// govPage is a synthetic gov.pl page in the ministry's layout: a link
// elsewhere with the spreadsheet's title, another attachment, then the
// spreadsheet, linked as the page links it, with markup in its text.
const govPage = `<html><body><h3>Materiały</h3>
<a href="https://other.example/attachment/f0">Dane dotyczące obligacji detalicznych (kopia)</a>
<a class="file-download" href="/attachment/a1" download>Rozporządzenie Ministra Finansów<br/></a>
<a class="file-download" href="/attachment/b2" target="_blank" download
aria-label="Pobierz plik Dane dotyczące obligacji detalicznych">
Dane dotyczące   obligacji detalicznych plik w formacie xls (lata 1992-2099)<br/>
<span class="extension">Dane&#8203;_dotyczace.xls</span></a>
</body></html>`

// govSite is how a synthetic gov.pl answers.
type govSite struct {
	pageStatus   int
	page         string
	noModified   bool
	requestCount atomic.Int32
}

// handler serves the page and the spreadsheet it links.
func (g *govSite) handler(tb testing.TB, file []byte) http.Handler {
	tb.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /web/finanse/obligacje-detaliczne1", func(w http.ResponseWriter, _ *http.Request) {
		g.requestCount.Add(1)
		if g.pageStatus != 0 {
			w.WriteHeader(g.pageStatus)
			return
		}
		if _, err := fmt.Fprint(w, g.page); err != nil {
			return // the client stopped reading
		}
	})
	mux.HandleFunc("GET /attachment/b2", func(w http.ResponseWriter, _ *http.Request) {
		if !g.noModified {
			w.Header().Set("Last-Modified", "Mon, 26 Oct 2099 13:53:03 GMT")
		}
		w.Header().Set("Content-Type", "application/vnd.ms-excel")
		if _, err := w.Write(file); err != nil {
			return // the client stopped reading
		}
	})
	return mux
}

// daneHandler serves a synthetic dane.gov.pl resource list, a CSV resource
// and then the spreadsheet, whose file URL is on another host and whose
// link is on its own, and the spreadsheet; status, when set, answers the
// list instead.
func daneHandler(tb testing.TB, file []byte, status int) http.Handler {
	tb.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /1.4/datasets/805/resources", func(w http.ResponseWriter, _ *http.Request) {
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		if _, err := io.WriteString(w, `{"data": [
			{"attributes": {"format": "csv", "file_url": "/media/x.csv", "data_date": "2099-09-30"}},
			{"attributes": {"title": "Sprzedaż obligacji detalicznych", "format": "xls",
			 "file_url": "https://other.example/media/resources/20991020/Dane.xls",
			 "link": "/media/resources/20991020/Dane.xls", "data_date": "2099-09-30",
			 "created": "2099-10-20T15:05:49Z"}}]}`); err != nil {
			return // the client stopped reading
		}
	})
	mux.HandleFunc("GET /media/resources/20991020/Dane.xls", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(file); err != nil {
			return // the client stopped reading
		}
	})
	return mux
}

// day is a synthetic date.
func day(tb testing.TB, s string) kernel.Date {
	tb.Helper()
	d, err := kernel.ParseDate(s)
	if err != nil {
		tb.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

// TestClientPublished reads the published data from gov.pl, finding the
// spreadsheet's link on the page, and from dane.gov.pl whenever anything
// about gov.pl fails, with the attribution the re-use terms ask for.
func TestClientPublished(t *testing.T) {
	t.Parallel()
	file := syntheticWorkbook(t)
	now := time.Date(2099, time.November, 7, 6, 0, 0, 0, time.UTC)
	wantRates, wantISINs, err := treasurybonds.ParseWorkbook(file)
	if err != nil {
		t.Fatalf("ParseWorkbook: %v", err)
	}
	tests := []struct {
		name       string
		gov        *govSite
		daneStatus int
		wantFrom   string // "gov" or "dane"
		wantErrs   []error
	}{
		{name: "from gov.pl", gov: &govSite{page: govPage}, wantFrom: "gov"},
		{name: "gov.pl is down", gov: &govSite{pageStatus: http.StatusServiceUnavailable}, wantFrom: "dane"},
		{name: "the page links no spreadsheet", gov: &govSite{page: "<a href=\"/attachment/a1\">Inne dane</a>"}, wantFrom: "dane"},
		{
			name: "the page links it only on another host", wantFrom: "dane",
			gov: &govSite{page: `<a href="https://other.example/attachment/b2">Dane dotyczące obligacji detalicznych</a>`},
		},
		{name: "the file has no date", gov: &govSite{page: govPage, noModified: true}, wantFrom: "dane"},
		{name: "the page is too large", gov: &govSite{page: strings.Repeat(" ", 2<<20) + govPage}, wantFrom: "dane"},
		{
			name: "both fail", gov: &govSite{page: "<html></html>"}, daneStatus: http.StatusBadGateway,
			wantErrs: []error{treasurybonds.ErrNotDiscovered, treasurybonds.ErrUnexpectedStatus},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gov := httptest.NewTLSServer(tt.gov.handler(t, file))
			defer gov.Close()
			dane := httptest.NewTLSServer(daneHandler(t, file, tt.daneStatus))
			defer dane.Close()
			c := treasurybonds.NewMinistryClient(treasurybonds.MinistryConfig{
				GovPL: gov.Client(), GovPLBaseURL: gov.URL, Dane: dane.Client(), DaneBaseURL: dane.URL,
				Now: func() time.Time { return now },
			})
			got, err := c.Published(context.Background())
			for _, want := range tt.wantErrs {
				if diff := cmp.Diff(want, err, cmpopts.EquateErrors()); diff != "" {
					t.Errorf("Published error (-want +got):\n%s", diff)
				}
			}
			if tt.wantErrs == nil {
				wantPublished(t, got, err, &pl.Published{Rates: wantRates, ISINs: wantISINs, Attribution: wantAttribution(t, tt.wantFrom, gov.URL, dane.URL, now)})
			}
		})
	}
}

// wantAttribution is the attribution of the synthetic data read from gov.pl
// at govURL or, when from is "dane", from dane.gov.pl at daneURL, at now.
func wantAttribution(tb testing.TB, from, govURL, daneURL string, now time.Time) pl.Attribution {
	tb.Helper()
	if from == "dane" {
		return pl.Attribution{
			Publisher: "Ministerstwo Finansów", Dataset: "Sprzedaż obligacji detalicznych", License: "CC BY 4.0",
			URL: daneURL + "/media/resources/20991020/Dane.xls", Made: day(tb, "2099-09-30"), Fetched: now, Processing: treasurybonds.PublishedProcessing,
		}
	}
	return pl.Attribution{
		Publisher: "Ministerstwo Finansów", Dataset: treasurybonds.DatasetTitle, License: "CC BY 4.0",
		URL: govURL + "/attachment/b2", Made: day(tb, "2099-10-26"), Fetched: now, Processing: treasurybonds.PublishedProcessing,
	}
}

// wantPublished fails t unless Published returned want without an error.
func wantPublished(t *testing.T, got *pl.Published, err error, want *pl.Published) {
	t.Helper()
	if err != nil {
		t.Fatalf("Published: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Published (-want +got):\n%s", diff)
	}
}

// TestClientPublishedCaches fetches the data once for six hours, then
// again.
func TestClientPublishedCaches(t *testing.T) {
	t.Parallel()
	site := &govSite{page: govPage}
	gov := httptest.NewTLSServer(site.handler(t, syntheticWorkbook(t)))
	defer gov.Close()
	now := time.Date(2099, time.November, 7, 6, 0, 0, 0, time.UTC)
	c := treasurybonds.NewMinistryClient(treasurybonds.MinistryConfig{
		GovPL: gov.Client(), GovPLBaseURL: gov.URL, Dane: gov.Client(), DaneBaseURL: gov.URL,
		Now: func() time.Time { return now },
	})
	for _, step := range []struct {
		after time.Duration
		want  int32
	}{{0, 1}, {time.Hour, 1}, {6 * time.Hour, 2}} {
		now = now.Add(step.after)
		if _, err := c.Published(context.Background()); err != nil {
			t.Fatalf("Published: %v", err)
		}
		if got := site.requestCount.Load(); got != step.want {
			t.Errorf("after %s: %d page requests, want %d", step.after, got, step.want)
		}
	}
}
