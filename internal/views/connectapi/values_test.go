package connectapi_test

import (
	"encoding/json"
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
// Polish default number format, four wrappers; acc-1 in IKZE with its fee
// rule.
func TestSettingsOverREST(t *testing.T) {
	t.Parallel()
	var settings struct {
		ReportingCurrency   string `json:"reportingCurrency"`
		NumberFormat        struct{ Key, DecimalSeparator string }
		NumberFormatDefault bool     `json:"numberFormatDefault"`
		TaxWrappers         []string `json:"taxWrappers"`
	}
	getJSON(t, "/v1/settings", &settings)
	got := []any{settings.ReportingCurrency, settings.NumberFormat.Key, settings.NumberFormat.DecimalSeparator, settings.NumberFormatDefault, len(settings.TaxWrappers)}
	if diff := cmp.Diff([]any{"PLN", "space-comma", ",", true, 4}, got); diff != "" {
		t.Errorf("settings (-want +got):\n%s", diff)
	}
	var account struct {
		Account struct {
			TaxWrapper string `json:"taxWrapper"`
			FeeRule    struct {
				Rate string `json:"rate"`
			} `json:"feeRule"`
		} `json:"account"`
	}
	getJSON(t, "/v1/accounts/acc-1", &account)
	if account.Account.TaxWrapper != "ikze" || account.Account.FeeRule.Rate != "0.0029" {
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
	} `json:"totals"`
}

// TestValuesOverREST reads the three values with acc-1 in IKZE without the
// owner's marginal rate: net values and totals are there, liquidation
// values that need the rate are missing with the reason, and cash counts
// in the totals: 1043.21 + 126 of holdings and -90 USD of cash at 4.00 are
// 809.21 gross, 1035 + 120 - 360 = 795 net.
func TestValuesOverREST(t *testing.T) {
	t.Parallel()
	var vals accountValuesJSON
	getJSON(t, "/v1/accounts/acc-1/valuations?from=2026-03-02&to=2026-03-02", &vals)
	got := []string{
		vals.Valuations[0].NetReporting.Amount, vals.Valuations[1].LiquidationMissing,
		vals.Totals[0].ValueReporting.Amount, vals.Totals[0].NetReporting.Amount, vals.Totals[0].CashReporting.Amount,
	}
	if diff := cmp.Diff([]string{"1035", "marginal_rate", "809.21", "795", "-360"}, got); diff != "" {
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
