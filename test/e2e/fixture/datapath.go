package main

import (
	_ "embed"
	"time"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// The data path's Flex queries (test/e2e/datapath_test.go): the query ID an
// instance's flex-token Secret names picks the statement, so the suite moves
// an instance from one statement to the next by changing that Secret, as an
// owner's statement moves on. Each has its own made-up reference code.
const (
	// DataPathOpening is the statement that seeds the data path's IBKR
	// account: lot-level open positions and cash on Friday, March 6, 2026.
	DataPathOpening = "e2e-datapath-opening"
	// DataPathWeek is the week after it, March 9 to 13: a deposit, a buy, a
	// dividend with tax withheld and a fee.
	DataPathWeek = "e2e-datapath-week"
	// DataPathCorrected is the same week with IBKR's correction of one row:
	// the deposit's amount, under the same transaction ID.
	DataPathCorrected = "e2e-datapath-week-corrected"
)

// NBPDataPathPath is where the fixture serves the data path's NBP table A:
// only GBP, CHF and NOK, so the data path's NBP connector instance delivers
// no rate of a currency the demo's own FX source already serves.
const NBPDataPathPath = "/nbp-datapath"

// The data path's synthetic statements: fake account U0000002, fake IDs and
// synthetic ISINs, in GBP and CHF, currencies no other dataset of the suite
// holds, so their FX rates have one source only.
var (
	//go:embed data/datapath-opening.xml
	dataPathOpening []byte
	//go:embed data/datapath-week.xml
	dataPathWeek []byte
	//go:embed data/datapath-week-corrected.xml
	dataPathCorrected []byte
)

// flexQuery is one statement the Flex Web Service hands out: the reference
// code SendRequest answers its query with, and the statement GetStatement
// answers that code with.
type flexQuery struct {
	reference string
	statement []byte
}

// dataPathQueries are the data path's Flex queries by query ID.
func dataPathQueries() map[string]flexQuery {
	return map[string]flexQuery{
		DataPathOpening:   {reference: "3333333301", statement: dataPathOpening},
		DataPathWeek:      {reference: "3333333302", statement: dataPathWeek},
		DataPathCorrected: {reference: "3333333303", statement: dataPathCorrected},
	}
}

// referenceFor returns the reference code SendRequest hands out for query:
// a data path query's own, or flexReference for any other.
func referenceFor(query string) string {
	if q, ok := dataPathQueries()[query]; ok {
		return q.reference
	}
	return flexReference
}

// statementOf returns the statement GetStatement answers reference with,
// and false for a reference SendRequest never handed out.
func statementOf(reference string) ([]byte, bool) {
	if reference == flexReference {
		return flexStatement, true
	}
	for _, q := range dataPathQueries() {
		if q.reference == reference {
			return q.statement, true
		}
	}
	return nil, false
}

// dataPathListings are the data path's synthetic instruments as EODHD's
// search lists them: the two the IBKR statements hold, both on the London
// Stock Exchange, TSTH in CHF, since the statements name no exchange and
// the eodhd connector prices only on the exchanges D-51 names, and the one
// the synthetic XTB export names by ticker TSTJ.NO, on Oslo's exchange,
// where the eodhd connector looks a Norwegian ticker up.
func dataPathListings() []map[string]string {
	return []map[string]string{
		searchResult("TSTG", "LSE", "GBP", "GBTEST000047", "Synthetic Test G"),
		searchResult("TSTH", "LSE", "CHF", "CHTEST000052", "Synthetic Test H"),
		searchResult("TSTJ", "OL", "NOK", "NOTEST000068", "Synthetic Test J"),
	}
}

// searchResult is one listing as EODHD's search answers it.
func searchResult(code, exchange, currency, isin, name string) map[string]string {
	return map[string]string{"Code": code, "Exchange": exchange, "Currency": currency, "ISIN": isin, "Name": name}
}

// dataPathCloses are the data path's synthetic closes, the same every
// trading day.
func dataPathCloses() map[string]string {
	return map[string]string{"TSTG.LSE": "105.00", "TSTH.LSE": "102.00", "TSTJ.OL": "55.00"}
}

// dataPathRate is one data path currency's synthetic NBP mid rate: base
// plus step for each weekday since the window's first, so each day's rate
// differs from the day before's and a conversion at the wrong day's rate
// shows.
type dataPathRate struct {
	code, base, step string
}

// dataPathRateRules are the data path's currencies' rates: GBP 5.0000,
// CHF 4.5000 and NOK 0.3600 on Friday, January 2, 2026, each weekday after
// it 0.0100, 0.0100 and 0.0010 more, so March 6 is 5.4500, 4.9500 and
// 0.4050.
func dataPathRateRules() []dataPathRate {
	return []dataPathRate{
		{code: "GBP", base: "5.0000", step: "0.0100"},
		{code: "CHF", base: "4.5000", step: "0.0100"},
		{code: "NOK", base: "0.3600", step: "0.0010"},
	}
}

// dataPathRates returns the data path's mid rates on weekday d of the
// window, as currency code and decimal text. A rule whose decimals don't
// parse gives no rate.
func dataPathRates(d time.Time) [][2]string {
	n := kernel.NewDecimal(int64(len(weekdays(dateRange{first: window().first, last: d}))-1), 0)
	rules := dataPathRateRules()
	out := make([][2]string, 0, len(rules))
	for _, r := range rules {
		base, baseErr := kernel.ParseDecimal(r.base)
		step, stepErr := kernel.ParseDecimal(r.step)
		if baseErr != nil || stepErr != nil {
			continue
		}
		out = append(out, [2]string{r.code, base.Add(step.Mul(n)).String()})
	}
	return out
}
