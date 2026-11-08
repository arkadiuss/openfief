package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// What the treasury-bonds connector reads, each below its own path, which
// its config's base URLs name (plugins/treasury-bonds): the Ministry of
// Finance's retail bond page on gov.pl and the file it links, the same
// dataset's resource list on dane.gov.pl's API, GUS's CPI from its SDP API
// and NBP's interest rate archive.
const (
	GovPLPath      = "/gov"
	DanePath       = "/dane"
	GUSPath        = "/gus"
	NBPArchivePath = "/nbp-static"
	// MinistryControlPath is where the suite publishes the ministry's data
	// (POST), which the fixture withholds until then.
	MinistryControlPath = "/control/ministry"
)

// The ministry's addresses on gov.pl and dane.gov.pl, below GovPLPath and
// DanePath, as the connector looks for them; the attachment and the
// resource's file are made-up names.
const (
	govPLPage      = "/web/finanse/obligacje-detaliczne1"
	govPLFile      = "/attachment/0e2e0000-0000-4000-8000-000000000805"
	daneResources  = "/1.4/datasets/805/resources"
	daneFile       = "/media/resources/e2e/obligacje-detaliczne.xls"
	ministryTitle  = "Dane dotyczące obligacji detalicznych"
	ministryMade   = "2026-11-02"
	ministryLastMo = "Mon, 02 Nov 2026 07:00:00 GMT"
)

// The bond series the end-to-end data path holds (test/e2e/bonds_test.go),
// with the synthetic registry ISINs the fixture's spreadsheet lists for
// them: Polish, with a valid check digit, and no provisional ISIN
// (pl.IsProvisional), since their national numbers hold no series code.
const (
	SeriesTOS = "TOS0329"
	SeriesEDO = "EDO0534"
	SeriesCOI = "COI1029"
	ISINTOS   = "PL9999001016"
	ISINEDO   = "PL9999001024"
	ISINCOI   = "PL9999001032"
)

// ministry is the Ministry of Finance's published data as the fixture
// serves it: withheld, every address answering 404, until the suite
// publishes it, so the lots entered before then are booked under their
// provisional ISINs and a later sync re-keys them (D-214).
type ministry struct {
	published atomic.Bool
}

// The columns of the synthetic spreadsheet's bond type sheets.
const (
	colSeries = iota
	colISIN
	colMaturity
	colRate
	colMargin
)

// ministrySheets are the synthetic spreadsheet's sheets in the ministry's
// layout: a description sheet, then one sheet per bond type with the
// series, its registry ISIN ("Kod ISIN"), its first-period rate (the first
// "Oprocentowanie" column) and, for indexed types, its margin ("Marża"),
// as fractions. Every rate is made up.
func ministrySheets() []xlsSheet {
	header := func(margin bool) []xlsCell {
		row := []xlsCell{
			textCell(colSeries, "Seria"), textCell(colISIN, "Kod ISIN"), textCell(colMaturity, "Data wykupu"),
			textCell(colRate, "Oprocentowanie"),
		}
		if margin {
			row = append(row, textCell(colMargin, "Marża"))
		}
		return row
	}
	return []xlsSheet{
		{name: "Opis", rows: [][]xlsCell{{textCell(colSeries, ministryTitle)}, {textCell(colSeries, "Dane syntetyczne zestawu testów OpenFief")}}},
		{name: "TOS", rows: [][]xlsCell{
			header(false),
			{textCell(colSeries, SeriesTOS), textCell(colISIN, ISINTOS), textCell(colMaturity, "2029-03-10"), numberCell(colRate, "0.0565")},
		}},
		{name: "COI", rows: [][]xlsCell{
			header(true),
			{
				textCell(colSeries, SeriesCOI), textCell(colISIN, ISINCOI), textCell(colMaturity, "2029-10-06"),
				numberCell(colRate, "0.056"), numberCell(colMargin, "0.015"),
			},
		}},
		{name: "EDO", rows: [][]xlsCell{
			header(true),
			{
				textCell(colSeries, SeriesEDO), textCell(colISIN, ISINEDO), textCell(colMaturity, "2034-05-14"),
				numberCell(colRate, "0.065"), numberCell(colMargin, "0.0175"),
			},
		}},
	}
}

// handle serves the ministry's addresses and the control the suite
// publishes the data with on mux.
func (m *ministry) handle(mux *http.ServeMux) {
	workbook := xlsFile(ministrySheets()...)
	mux.HandleFunc("GET "+GovPLPath+govPLPage, m.withheld(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(w, "text/html; charset=utf-8", []byte(`<!doctype html><html lang="pl"><body><h1>Obligacje detaliczne</h1>`+
			`<ul><li><a href="`+GovPLPath+govPLFile+`">`+ministryTitle+` (XLS)</a></li></ul></body></html>`))
	}))
	mux.HandleFunc("GET "+GovPLPath+govPLFile, m.withheld(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Last-Modified", ministryLastMo)
		writeBody(w, "application/vnd.ms-excel", workbook)
	}))
	mux.HandleFunc("GET "+DanePath+daneResources, m.withheld(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"data": []any{map[string]any{"attributes": map[string]any{
			"title": ministryTitle, "format": "xls", "file_url": DanePath + daneFile, "data_date": ministryMade,
		}}}})
	}))
	mux.HandleFunc("GET "+DanePath+daneFile, m.withheld(func(w http.ResponseWriter, _ *http.Request) {
		writeBody(w, "application/vnd.ms-excel", workbook)
	}))
	mux.HandleFunc("POST "+MinistryControlPath, func(w http.ResponseWriter, _ *http.Request) {
		m.published.Store(true)
		w.WriteHeader(http.StatusNoContent)
	})
}

// withheld answers 404 until the data is published, then h.
func (m *ministry) withheld(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.published.Load() {
			http.NotFound(w, r)
			return
		}
		h(w, r)
	}
}

// GUS's SDP API: the CPI indicators by the year they hold (internal/valuation/gus),
// the period id of January, the "same month a year before = 100" way of
// presentation, and a flag that marks a final value.
const (
	gusCOICOP1999 = 1832
	gusCOICOP2018 = 2728
	gusFirst2018  = 2026
	gusJanuary    = 247
	gusYearOnYear = 5
	gusFinalFlag  = 1
	// cpiFirstYear is the year of cpiTenths' first row.
	cpiFirstYear = 2023
)

// cpiTenths are the CPI readings the demo dataset seeds
// (internal/demo/bonds.go, cpiTenths), the same month of the previous year
// = 100, in tenths, from January 2023 to September 2026: public GUS
// statistics, not anyone's financial data. The fixture serves the same
// readings, so that the suite's CPI agrees with the demo's whichever
// source the core values bonds from (D-94); its test checks the two agree.
// Only cmd imports the demo dataset, so they are copied here.
func cpiTenths() [][]int64 {
	return [][]int64{
		{1166, 1184, 1161, 1147, 1130, 1115, 1108, 1101, 1082, 1066, 1066, 1062}, // 2023
		{1037, 1028, 1020, 1024, 1025, 1026, 1042, 1043, 1049, 1050, 1047, 1047}, // 2024
		{1049, 1049, 1049, 1043, 1040, 1041, 1031, 1029, 1029, 1028, 1025, 1024}, // 2025
		{1021, 1021, 1030, 1032, 1031, 1025, 1030, 1034, 1041},                   // 2026, to September
	}
}

// gusCPI answers GUS's indicator data for the CPI indicator of a year: one
// value per month of cpiTenths, as the index of the same month a year
// before = 100, or 404 for a year without any or the other indicator, as
// GUS answers.
func gusCPI(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year, yearErr := strconv.Atoi(q.Get("id-rok"))
	indicator, indicatorErr := strconv.Atoi(q.Get("id-wskaznik"))
	want := gusCOICOP1999
	if year >= gusFirst2018 {
		want = gusCOICOP2018
	}
	if yearErr != nil || indicatorErr != nil || indicator != want {
		http.NotFound(w, r)
		return
	}
	values := cpiValues(year)
	if len(values) == 0 {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, values)
}

// cpiValues are cpiTenths' values of year as GUS's API writes them.
func cpiValues(year int) []map[string]json.Number {
	rows := cpiTenths()
	if year < cpiFirstYear || year >= cpiFirstYear+len(rows) {
		return nil
	}
	months := rows[year-cpiFirstYear]
	out := make([]map[string]json.Number, 0, len(months))
	for m, tenths := range months {
		out = append(out, map[string]json.Number{
			"id-daty":                     json.Number(strconv.Itoa(year)),
			"id-okres":                    json.Number(strconv.Itoa(gusJanuary + m)),
			"id-sposob-prezentacji-miara": json.Number(strconv.Itoa(gusYearOnYear)),
			"id-flaga":                    json.Number(strconv.Itoa(gusFinalFlag)),
			"wartosc":                     json.Number(kernel.NewDecimal(tenths, -1).String()),
		})
	}
	return out
}

// nbpArchive is a synthetic NBP interest rate archive: two made-up changes
// of the reference rate, each listing another rate too, with the byte order
// mark NBP starts it with.
const nbpArchive = "\xEF\xBB\xBF" + `<?xml version="1.0" encoding="UTF-8"?>
<stopy_procentowe_archiwum>
<pozycje obowiazuje_od="2022-01-05"><pozycja id="ref" oprocentowanie="5,00"/><pozycja id="lom" oprocentowanie="5,50"/></pozycje>
<pozycje obowiazuje_od="2025-07-03"><pozycja id="ref" oprocentowanie="4,50"/><pozycja id="lom" oprocentowanie="5,00"/></pozycje>
</stopy_procentowe_archiwum>
`

// nbpRateArchive answers NBP's interest rate archive.
func nbpRateArchive(w http.ResponseWriter, _ *http.Request) {
	writeBody(w, "text/xml", []byte(nbpArchive))
}

// handleBondIndices serves GUS's CPI and NBP's archive on mux.
func handleBondIndices(mux *http.ServeMux) {
	mux.HandleFunc("GET "+GUSPath+"/api/indicators/indicator-data-indicator", gusCPI)
	mux.HandleFunc("GET "+NBPArchivePath+"/dane/stopy/stopy_procentowe_archiwum.xml", nbpRateArchive)
}
