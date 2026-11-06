package connectapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// getJSON serves GET target with a good token and decodes the answer.
func getJSON(t *testing.T, target string, into any) {
	t.Helper()
	_, rec := newAPI(t, newLedger(t)).serve(t, http.MethodGet, target, goodToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d\n%s", target, rec.Code, rec.Body)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), into); err != nil {
		t.Fatalf("decoding %s: %v", rec.Body, err)
	}
}

// restMoney is a Money as REST writes it.
type restMoney struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// TestSettingsOverREST reads the settings and acc-1's terms: PLN, the
// Polish default number format, four wrappers, OKI's four forms and IKE's
// earlier payments (D-147, D-149); acc-1 in IKZE with its fee rule and the
// form the fake stores.
func TestSettingsOverREST(t *testing.T) {
	t.Parallel()
	var settings struct {
		ReportingCurrency   string `json:"reportingCurrency"`
		NumberFormat        struct{ Key, DecimalSeparator string }
		NumberFormatDefault bool     `json:"numberFormatDefault"`
		TaxWrappers         []string `json:"taxWrappers"`
		WrapperTerms        []struct {
			TaxWrapper      string   `json:"taxWrapper"`
			AccountForms    []string `json:"accountForms"`
			EarlierPayments bool     `json:"earlierPayments"`
		} `json:"wrapperTerms"`
	}
	getJSON(t, "/v1/settings", &settings)
	got := []any{settings.ReportingCurrency, settings.NumberFormat.Key, settings.NumberFormat.DecimalSeparator, settings.NumberFormatDefault, len(settings.TaxWrappers)}
	if diff := cmp.Diff([]any{"PLN", "space-comma", ",", true, 4}, got); diff != "" {
		t.Errorf("settings (-want +got):\n%s", diff)
	}
	terms := map[string]string{}
	for _, w := range settings.WrapperTerms {
		terms[w.TaxWrapper] = fmt.Sprint(len(w.AccountForms), w.EarlierPayments)
	}
	if diff := cmp.Diff(map[string]string{"taxable": "0 false", "ike": "0 true", "ikze": "0 false", "oki": "4 false"}, terms); diff != "" {
		t.Errorf("wrapper terms, forms and earlier payments (-want +got):\n%s", diff)
	}
	var account struct {
		Account struct {
			TaxWrapper  string `json:"taxWrapper"`
			AccountForm string `json:"accountForm"`
			FeeRule     struct {
				Rate string `json:"rate"`
			} `json:"feeRule"`
		} `json:"account"`
	}
	getJSON(t, "/v1/accounts/acc-1", &account)
	if account.Account.TaxWrapper != "ikze" || account.Account.FeeRule.Rate != "0.0029" || account.Account.AccountForm != "brokerage-account" {
		t.Errorf("account = %+v", account)
	}
}

// accountValuesJSON is what TestValuesOverREST reads of an account's
// valuations.
type accountValuesJSON struct {
	Valuations []struct {
		NetReporting       *restMoney `json:"netReporting"`
		LiquidationMissing string     `json:"liquidationMissing"`
	} `json:"valuations"`
	Totals []struct {
		ValueReporting       restMoney  `json:"valueReporting"`
		NetReporting         *restMoney `json:"netReporting"`
		LiquidationReporting *restMoney `json:"liquidationReporting"`
		CashReporting        restMoney  `json:"cashReporting"`
		NetOffsetReporting   *restMoney `json:"netOffsetReporting"`
	} `json:"totals"`
}

// TestValuesOverREST reads the three values with acc-1 in IKZE without the
// owner's marginal rate: net values and totals are there, liquidation
// values that need the rate are missing with the reason, and cash counts
// in the totals: 1043.21 + 126 of holdings and -90 USD of cash at 4.00 are
// 809.21 gross. The rows' net values sum to 1035 + 120 - 360 = 795, but
// IKZE taxes the whole withdrawal (D-128): 10% of 809.21 is 80.92, so the
// net total is 728.29 and the losses-offset line -66.71.
func TestValuesOverREST(t *testing.T) {
	t.Parallel()
	var vals accountValuesJSON
	getJSON(t, "/v1/accounts/acc-1/valuations?from=2026-03-02&to=2026-03-02", &vals)
	got := []string{
		vals.Valuations[0].NetReporting.Amount, vals.Valuations[1].LiquidationMissing,
		vals.Totals[0].ValueReporting.Amount, vals.Totals[0].NetReporting.Amount, vals.Totals[0].CashReporting.Amount,
		vals.Totals[0].NetOffsetReporting.Amount,
	}
	if diff := cmp.Diff([]string{"1035", "marginal_rate", "809.21", "728.29", "-360", "-66.71"}, got); diff != "" {
		t.Errorf("valuations and totals (-want +got):\n%s", diff)
	}
	if vals.Totals[0].LiquidationReporting != nil {
		t.Errorf("liquidation total %v, want none without the marginal rate", vals.Totals[0].LiquidationReporting)
	}
}

// TestYearlyTaxesOverREST reads OKI's yearly tax, whose act isn't in force.
func TestYearlyTaxesOverREST(t *testing.T) {
	t.Parallel()
	var taxes struct {
		Taxes []struct {
			TaxWrapper string `json:"taxWrapper"`
			Year       int    `json:"year"`
			InForce    bool   `json:"inForce"`
		} `json:"taxes"`
	}
	getJSON(t, "/v1/yearly-taxes?year=2026", &taxes)
	if len(taxes.Taxes) != 1 || taxes.Taxes[0].TaxWrapper != "oki" || taxes.Taxes[0].Year != 2026 || taxes.Taxes[0].InForce {
		t.Errorf("yearly taxes = %+v", taxes)
	}
}

// TestYearlyTaxUnderTheActOverREST reads 2027, under the act (D-127): no
// account is in OKI, so every class averages zero and nothing is taxed,
// at 0.85%.
func TestYearlyTaxUnderTheActOverREST(t *testing.T) {
	t.Parallel()
	var y2027 struct {
		Taxes []struct {
			InForce  bool   `json:"inForce"`
			Rate     string `json:"rate"`
			Averages []struct {
				ExemptionClass string    `json:"exemptionClass"`
				Amount         restMoney `json:"amount"`
			} `json:"averages"`
			Exempt []struct {
				ExemptionClass string `json:"exemptionClass"`
			} `json:"exempt"`
			Tax *restMoney `json:"tax"`
		} `json:"taxes"`
	}
	getJSON(t, "/v1/yearly-taxes?year=2027", &y2027)
	if len(y2027.Taxes) != 1 {
		t.Fatalf("2027 = %+v", y2027)
	}
	got := y2027.Taxes[0]
	if !got.InForce || got.Rate != "0.0085" || len(got.Averages) != 3 || got.Averages[0].ExemptionClass != "savings" ||
		len(got.Exempt) != 3 || got.Tax == nil || got.Tax.Amount != "0" {
		t.Errorf("2027 = %+v", got)
	}
}

// TestAllocationOfAKindOverREST splits the liquidation value: acc-1's ETF1
// and its IKZE cash are left out for want of the marginal rate.
func TestAllocationOfAKindOverREST(t *testing.T) {
	t.Parallel()
	var alloc struct {
		Value    string `json:"value"`
		Unvalued []struct {
			Missing string     `json:"missing"`
			Cash    *restMoney `json:"cash"`
		} `json:"unvalued"`
	}
	getJSON(t, "/v1/allocation?date=2026-03-02&by=ALLOCATION_DIMENSION_SLEEVE&value=VALUE_KIND_LIQUIDATION", &alloc)
	if alloc.Value != "VALUE_KIND_LIQUIDATION" || len(alloc.Unvalued) != 2 || alloc.Unvalued[0].Missing != "marginal_rate" || alloc.Unvalued[1].Cash == nil {
		t.Errorf("liquidation allocation = %+v", alloc)
	}
}

// TestRESTHoldingsAndLotsMarked reads acc-1's holdings and lots with the
// buy of ETF1 from the removed instance ibkr-old: the ETF1 holding and its
// lot are no longer updated and name ibkr-old; the bond, opened by the
// installed ibkr-main, is not marked. Both answers match the contract.
func TestRESTHoldingsAndLotsMarked(t *testing.T) {
	t.Parallel()
	c := loadContract(t)
	a := api{handler: pluginAPI(t, installed())}
	type marked struct {
		Instrument struct {
			Symbol string `json:"symbol"`
		} `json:"instrument"`
		NoLongerUpdated  bool     `json:"noLongerUpdated"`
		RemovedInstances []string `json:"removedInstances"`
	}
	want := map[string]marked{}
	etf := marked{NoLongerUpdated: true, RemovedInstances: []string{"ibkr-old"}}
	etf.Instrument.Symbol = "ETF1"
	bond := marked{}
	bond.Instrument.Symbol = "EDO0334"
	want["ETF1"], want["EDO0334"] = etf, bond
	for _, target := range []string{"/v1/accounts/acc-1/holdings?date=2026-03-02", "/v1/accounts/acc-1/lots?date=2026-03-02"} {
		req, rec := a.serve(t, http.MethodGet, target, goodToken)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s = %d\n%s", target, rec.Code, rec.Body)
		}
		c.check(t, req, rec)
		var body struct {
			Holdings []marked `json:"holdings"`
			Lots     []marked `json:"lots"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decoding %s: %v", rec.Body, err)
		}
		got := map[string]marked{}
		for _, m := range append(body.Holdings, body.Lots...) {
			got[m.Instrument.Symbol] = m
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("GET %s marks (-want +got):\n%s", target, diff)
		}
	}
}
