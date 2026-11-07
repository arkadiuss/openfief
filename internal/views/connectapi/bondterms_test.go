package connectapi_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// fakeSeriesTerms stores EDO0334's terms: published rates with their
// attribution and a fee carried forward as assumed (D-97).
type fakeSeriesTerms struct{}

func (fakeSeriesTerms) BondTerms(_ context.Context, isin ledger.ISIN) (pl.Terms, bool, error) {
	if isin.String() != isinPL {
		return pl.Terms{}, false, nil
	}
	s, err1 := pl.ParseSeries("EDO0334")
	made, err2 := kernel.ParseDate("2026-10-26")
	fee, err3 := kernel.ParseMoney("3.00", pl.Nominal().Currency())
	if err := errors.Join(err1, err2, err3); err != nil {
		return pl.Terms{}, false, fmt.Errorf("the synthetic terms: %w", err)
	}
	return pl.Terms{
		Series: s, FirstPeriodRate: kernel.NewDecimal(611, -4), Margin: kernel.NewDecimal(2, -2), EarlyRedemptionFee: fee,
		FeeSource: pl.FeeAssumed,
		Attribution: &pl.Attribution{
			Publisher: "Ministerstwo Finansów", Dataset: "Dane dotyczące obligacji detalicznych", License: "CC BY 4.0",
			URL: "https://www.example.org/attachment/x", Made: made, Fetched: time.Date(2026, time.November, 7, 6, 0, 0, 0, time.UTC),
			Processing: "read",
		},
	}, true, nil
}

// bondTermsJSON is what TestRESTBondTerms reads of an instrument.
type bondTermsJSON struct {
	Instrument struct {
		BondTerms *struct {
			Series             string `json:"series"`
			FirstPeriodRate    string `json:"firstPeriodRate"`
			Margin             string `json:"margin"`
			RatesSource        string `json:"ratesSource"`
			EarlyRedemptionFee struct {
				Amount string `json:"amount"`
			} `json:"earlyRedemptionFee"`
			FeeSource   string `json:"feeSource"`
			Attribution struct {
				Publisher string `json:"publisher"`
				License   string `json:"license"`
				Made      string `json:"made"`
				Fetched   string `json:"fetched"`
			} `json:"attribution"`
		} `json:"bondTerms"`
	} `json:"instrument"`
}

// TestRESTBondTerms reads a treasury bond's stored terms with where each
// came from and the published data's attribution (D-97), and none for an
// instrument that is no bond.
func TestRESTBondTerms(t *testing.T) {
	t.Parallel()
	a := uiv2API(t)
	var got bondTermsJSON
	serveJSON(t, a, "/v1/instruments/"+isinPL, http.StatusOK, &got)
	b := got.Instrument.BondTerms
	if b == nil {
		t.Fatal("the bond has no terms")
	}
	want := [][2]string{
		{"series", "EDO0334"},
		{"first period rate", "0.0611"},
		{"margin", "0.02"},
		{"rates source", "published"},
		{"fee", "3"},
		{"fee source", "assumed"},
		{"publisher", "Ministerstwo Finansów"},
		{"licence", "CC BY 4.0"},
		{"made", "2026-10-26"},
		{"fetched", "2026-11-07T06:00:00Z"},
	}
	gotPairs := [][2]string{
		{"series", b.Series},
		{"first period rate", b.FirstPeriodRate},
		{"margin", b.Margin},
		{"rates source", b.RatesSource},
		{"fee", b.EarlyRedemptionFee.Amount},
		{"fee source", b.FeeSource},
		{"publisher", b.Attribution.Publisher},
		{"licence", b.Attribution.License},
		{"made", b.Attribution.Made},
		{"fetched", b.Attribution.Fetched},
	}
	if diff := cmp.Diff(want, gotPairs); diff != "" {
		t.Errorf("bond terms (-want +got):\n%s", diff)
	}
	var etf bondTermsJSON
	serveJSON(t, a, "/v1/instruments/"+isinIE, http.StatusOK, &etf)
	if etf.Instrument.BondTerms != nil {
		t.Errorf("an ETF has bond terms %+v", etf.Instrument.BondTerms)
	}
}
