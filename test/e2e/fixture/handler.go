package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/test/e2e/modelseries"
)

// The services' paths on the fixture: each connector's baseUrl is the
// fixture's origin plus one of them.
const (
	FlexPath  = "/ibkr/AccountManagement/FlexWebService"
	EODHDPath = "/eodhd"
	NBPPath   = "/nbp"
)

// flexReference is the reference code SendRequest hands out for every query
// but the data path's (datapath.go), and GetStatement takes; made up.
const flexReference = "2222222222"

// nbpMaxDays is the longest range NBP's API serves in one call.
const nbpMaxDays = 93

// isoDate is the layout of the dates in the services' URLs and answers.
const isoDate = time.DateOnly

// flexStatement is the synthetic Flex Query statement GetStatement returns:
// fake accounts, fake IDs and synthetic ISINs, a copy of the ibkrflex
// adapter's test statement.
//
//go:embed data/flex-statement.xml
var flexStatement []byte

// dateRange is the days from first to last, both included.
type dateRange struct{ first, last time.Time }

// window is the span of days the fixture has closes and rates for, the
// statement's weeks and the weeks before.
func window() dateRange {
	return dateRange{
		first: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC),
		last:  time.Date(2026, time.March, 13, 0, 0, 0, 0, time.UTC),
	}
}

// listings are the synthetic instruments EODHD's search knows, as its search
// results, whose field names are capitalized: the ones the synthetic Flex
// statement holds.
func listings() []map[string]string {
	return append([]map[string]string{
		searchResult("TSTA", "LSE", "USD", "PLTEST000010", "Synthetic Test A"),
		searchResult("TSTB", "XETRA", "EUR", "IETEST000026", "Synthetic Test B"),
	}, dataPathListings()...)
}

// closes are the synthetic close of each EODHD ticker the fixture knows,
// the data path's included, the same every trading day, as decimal text.
func closes() map[string]string {
	c := map[string]string{"TSTA.LSE": "102.40", "TSTB.XETRA": "107.00", "FAKE20.INDX": "2500.00"}
	maps.Copy(c, dataPathCloses())
	return c
}

// midRates are the synthetic NBP table A mid rates, the same every trading
// day, as decimal text.
func midRates() [][2]string {
	return [][2]string{{"USD", "3.9512"}, {"EUR", "4.2731"}}
}

// credentials are the tokens the fixture accepts; empty accepts any
// non-empty one.
type credentials struct {
	flexToken  string
	eodhdToken string
}

// newHandler serves the services below their paths, the treasury-bonds
// connector's too (bonds.go), and logs each request's method and path,
// never its query, which holds the tokens.
func newHandler(creds credentials, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+FlexPath+"/SendRequest", func(w http.ResponseWriter, r *http.Request) {
		flexSendRequest(w, r, creds.flexToken)
	})
	mux.HandleFunc("GET "+FlexPath+"/GetStatement", func(w http.ResponseWriter, r *http.Request) {
		flexGetStatement(w, r, creds.flexToken)
	})
	mux.HandleFunc("GET "+EODHDPath+"/api/search/{query}", func(w http.ResponseWriter, r *http.Request) {
		eodhdSearch(w, r, creds.eodhdToken)
	})
	mux.HandleFunc("GET "+EODHDPath+"/api/eod/{ticker}", func(w http.ResponseWriter, r *http.Request) {
		eodhdCloses(w, r, creds.eodhdToken)
	})
	mux.HandleFunc("GET "+NBPPath+"/api/exchangerates/tables/A/{from}/{to}/", nbpTableA)
	mux.HandleFunc("GET "+NBPDataPathPath+"/api/exchangerates/tables/A/{from}/{to}/", func(w http.ResponseWriter, r *http.Request) {
		nbpTables(w, r, dataPathRates)
	})
	(&ministry{}).handle(mux)
	handleBondIndices(mux)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", "method", r.Method, "path", r.URL.Path)
		mux.ServeHTTP(w, r)
	})
}

// flexSendRequest answers SendRequest: a reference code for a valid token
// and query, or the Flex Web Service's error for an invalid token.
func flexSendRequest(w http.ResponseWriter, r *http.Request, token string) {
	q := r.URL.Query()
	switch {
	case !accepted(q.Get("t"), token):
		flexError(w, "1015", "Token is invalid.")
	case q.Get("q") == "" || q.Get("v") != "3":
		flexError(w, "1014", "Query is invalid.")
	default:
		writeXML(w, "<FlexStatementResponse><Status>Success</Status><ReferenceCode>"+referenceFor(q.Get("q"))+
			"</ReferenceCode><Url>https://elsewhere.example/AccountManagement/FlexWebService/GetStatement</Url>"+
			"</FlexStatementResponse>")
	}
}

// flexGetStatement answers GetStatement: the synthetic statement for the
// reference code SendRequest handed out.
func flexGetStatement(w http.ResponseWriter, r *http.Request, token string) {
	q := r.URL.Query()
	statement, known := statementOf(q.Get("q"))
	switch {
	case !accepted(q.Get("t"), token):
		flexError(w, "1015", "Token is invalid.")
	case !known || q.Get("v") != "3":
		flexError(w, "1014", "Query is invalid.")
	default:
		writeBody(w, "text/xml;charset=UTF-8", statement)
	}
}

// flexError writes a FlexStatementResponse that failed with code.
func flexError(w http.ResponseWriter, code, message string) {
	writeXML(w, "<FlexStatementResponse><Status>Fail</Status><ErrorCode>"+code+"</ErrorCode><ErrorMessage>"+
		message+"</ErrorMessage></FlexStatementResponse>")
}

// writeXML writes a 200 with an XML body, as the Flex Web Service answers
// even its errors.
func writeXML(w http.ResponseWriter, body string) {
	writeBody(w, "text/xml;charset=UTF-8", []byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+body+"\n"))
}

// writeBody writes a 200 with body of contentType. A client that went away
// gets nothing more; the failure is only logged.
func writeBody(w http.ResponseWriter, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	if _, err := w.Write(body); err != nil {
		slog.Default().Debug("writing the answer failed", "error", kernel.LogSafe(err))
	}
}

// eodhdSearch answers EODHD's search with the listings of the ISIN or code
// searched for, on the exchange the query names, if any; none is an empty
// array.
func eodhdSearch(w http.ResponseWriter, r *http.Request, token string) {
	if !eodhdAuthorized(w, r, token) {
		return
	}
	query, exchange := r.PathValue("query"), r.URL.Query().Get("exchange")
	out := []map[string]string{}
	for _, l := range listings() {
		if (query == l["ISIN"] || query == l["Code"]) && (exchange == "" || exchange == l["Exchange"]) {
			out = append(out, l)
		}
	}
	writeJSON(w, out)
}

// eodhdBar is one day of EODHD's end-of-day API.
type eodhdBar struct {
	Date  string      `json:"date"`
	Close json.Number `json:"close"`
}

// eodhdCloses answers EODHD's end-of-day API for a known ticker: its close
// on every weekday from the query's from to its to, within the window.
func eodhdCloses(w http.ResponseWriter, r *http.Request, token string) {
	if !eodhdAuthorized(w, r, token) {
		return
	}
	if series, ok := modelseries.Find(r.PathValue("ticker")); ok {
		modelCloses(w, r, series)
		return
	}
	price, ok := closes()[r.PathValue("ticker")]
	if !ok {
		http.NotFound(w, r)
		return
	}
	asked, err := span(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	bars := []eodhdBar{}
	for _, d := range weekdays(asked) {
		bars = append(bars, eodhdBar{Date: d.Format(isoDate), Close: json.Number(price)})
	}
	writeJSON(w, bars)
}

// modelCloses answers EODHD's end-of-day API for one of the model
// portfolio's synthetic index series: its level on every weekday within the
// query's from and to, each optional, over the series' own days, which the
// window doesn't bound (the series start in December 2025, and the model
// portfolio's backtest crosses a year end).
func modelCloses(w http.ResponseWriter, r *http.Request, series modelseries.Series) {
	var from, to time.Time
	for key, dst := range map[string]*time.Time{"from": &from, "to": &to} {
		raw := r.URL.Query().Get(key)
		if raw == "" {
			continue
		}
		d, err := time.Parse(isoDate, raw)
		if err != nil {
			http.Error(w, key+": "+err.Error(), http.StatusUnprocessableEntity)
			return
		}
		*dst = d
	}
	bars := []eodhdBar{}
	for _, c := range series.Closes(from, to) {
		bars = append(bars, eodhdBar{Date: c.Date, Close: json.Number(c.Level)})
	}
	writeJSON(w, bars)
}

// eodhdAuthorized answers 401 and reports false unless r carries an
// accepted api_token and asks for JSON.
func eodhdAuthorized(w http.ResponseWriter, r *http.Request, token string) bool {
	q := r.URL.Query()
	if !accepted(q.Get("api_token"), token) || q.Get("fmt") != "json" {
		http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		return false
	}
	return true
}

// span reads an EODHD from and to, each optional: the window's ends by
// default.
func span(rawFrom, rawTo string) (dateRange, error) {
	r := window()
	var err error
	if rawFrom != "" {
		if r.first, err = time.Parse(isoDate, rawFrom); err != nil {
			return dateRange{}, fmt.Errorf("from: %w", err)
		}
	}
	if rawTo != "" {
		if r.last, err = time.Parse(isoDate, rawTo); err != nil {
			return dateRange{}, fmt.Errorf("to: %w", err)
		}
	}
	return r, nil
}

// nbpTable is one NBP table A, its mid rates as decimal text.
type nbpTable struct {
	Table         string    `json:"table"`
	No            string    `json:"no"`
	EffectiveDate string    `json:"effectiveDate"`
	Rates         []nbpRate `json:"rates"`
}

type nbpRate struct {
	Currency string      `json:"currency"`
	Code     string      `json:"code"`
	Mid      json.Number `json:"mid"`
}

// nbpTableA answers NBP's table A API for a range of at most 93 days: one
// table per weekday within the window, or 404 when none falls in the range,
// as NBP answers a weekend.
func nbpTableA(w http.ResponseWriter, r *http.Request) {
	nbpTables(w, r, func(time.Time) [][2]string { return midRates() })
}

// nbpTables answers NBP's table A API as nbpTableA does, with each day's
// mid rates from rates.
func nbpTables(w http.ResponseWriter, r *http.Request, rates func(time.Time) [][2]string) {
	if r.URL.Query().Get("format") != "json" {
		http.Error(w, "400 BadRequest - Nieprawidłowy format", http.StatusBadRequest)
		return
	}
	from, fromErr := time.Parse(isoDate, r.PathValue("from"))
	to, toErr := time.Parse(isoDate, r.PathValue("to"))
	if fromErr != nil || toErr != nil || to.Before(from) || to.Sub(from) >= nbpMaxDays*24*time.Hour {
		http.Error(w, "400 BadRequest - Błędny zakres dat", http.StatusBadRequest)
		return
	}
	days := weekdays(dateRange{first: from, last: to})
	if len(days) == 0 {
		http.Error(w, "404 NotFound - Brak danych", http.StatusNotFound)
		return
	}
	tables := make([]nbpTable, 0, len(days))
	for i, d := range days {
		t := nbpTable{Table: "A", No: fmt.Sprintf("%03d/A/NBP/%d", i+1, d.Year()), EffectiveDate: d.Format(isoDate)}
		for _, m := range rates(d) {
			t.Rates = append(t.Rates, nbpRate{Currency: strings.ToLower(m[0]), Code: m[0], Mid: json.Number(m[1])})
		}
		tables = append(tables, t)
	}
	writeJSON(w, tables)
}

// weekdays are the weekdays of r within the window.
func weekdays(r dateRange) []time.Time {
	from, to := r.first, r.last
	if w := window(); from.Before(w.first) {
		from = w.first
	}
	if w := window(); to.After(w.last) {
		to = w.last
	}
	var out []time.Time
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			out = append(out, d)
		}
	}
	return out
}

// accepted reports whether got is a token the fixture takes: want, or any
// non-empty one when want is empty.
func accepted(got, want string) bool {
	return got != "" && (want == "" || got == want)
}

// writeJSON writes v as a 200 JSON body.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "encoding failed", http.StatusInternalServerError)
	}
}
