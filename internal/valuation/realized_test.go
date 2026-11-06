package valuation_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// TestRealizedGains checks each sale's gain for tax (D-128), every amount
// at the NBP rate of the day before its event: 1 950 USD sold on
// 2026-03-10 with a 5 USD fee at 4.20 (2026-03-09) is 8 190 − 21; its lots,
// 1 000 USD opened 2026-03-02 at 4.00 (2026-02-27, the last rate on or
// before 2026-03-01) with a 2 USD buy fee and 600 USD opened 2026-03-03 at
// 4.10 (2026-03-02), cost 4 000 + 8 + 2 460: the buy fee counts in the tax
// cost (D-150), so the gain is 1 701. 500 PLN for a 600 PLN lot is a loss
// of 100 in 2027. RealizedInYear sums the tax year through a day.
func TestRealizedGains(t *testing.T) {
	t.Parallel()
	etf := instrument(t, etfISIN, "SYNTH")
	key := func(id string) ledger.TransactionKey { return ledger.TransactionKey{Source: "test", ExternalID: id} }
	lot := func(opened, cost, c string) ledger.Lot {
		return ledger.Lot{Instrument: etf, Opened: date(t, opened), Quantity: kernel.NewQuantity(dec(t, "1")), Cost: money(t, cost, c), Origin: key(opened)}
	}
	withFee := func(l ledger.Lot, fee kernel.Money) ledger.Lot { l.Fees = []kernel.Money{fee}; return l }
	sales := []ledger.Sale{
		{
			Key: key("s1"), Date: date(t, "2026-03-10"), Instrument: etf, Proceeds: money(t, "1950", "USD"),
			Fees: []kernel.Money{money(t, "5", "USD")}, Consumed: []ledger.Lot{withFee(lot("2026-03-02", "1000", "USD"), money(t, "2", "USD")), lot("2026-03-03", "600", "USD")},
		},
		{Key: key("s2"), Date: date(t, "2027-01-10"), Instrument: etf, Proceeds: money(t, "500", "PLN"), Consumed: []ledger.Lot{lot("2027-01-05", "600", "PLN")}},
	}
	rates := []valuation.FXRate{rate(t, "USD", "2026-02-27", "4.00"), rate(t, "USD", "2026-03-02", "4.10"), rate(t, "USD", "2026-03-09", "4.20")}
	terms := valuation.AccountTerms{Pack: pl.NewPack(), Wrapper: "taxable"}
	got, missing, err := terms.RealizedGains(sales, rates)
	if err != nil || missing != 0 {
		t.Fatalf("RealizedGains: %v, %v", missing, err)
	}
	want := []valuation.RealizedGain{
		{Date: date(t, "2026-03-10"), Gain: money(t, "1701.00", "PLN")},
		{Date: date(t, "2027-01-10"), Gain: money(t, "-100.00", "PLN")},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("gains (-want +got):\n%s", diff)
	}
	for _, tt := range []struct{ day, want string }{
		{"2026-03-09", "0"}, {"2026-12-31", "1701.00"}, {"2027-01-09", "0"}, {"2027-01-10", "-100.00"},
	} {
		sum, sumErr := valuation.RealizedInYear(got, date(t, tt.day), cur(t, "PLN"))
		if sumErr != nil || !sum.Equal(money(t, tt.want, "PLN")) {
			t.Errorf("RealizedInYear(%s) = %s, %v; want %s", tt.day, sum, sumErr, tt.want)
		}
	}
	_, missing, err = terms.RealizedGains([]ledger.Sale{{Key: key("s3"), Date: date(t, "2026-03-10"), Proceeds: money(t, "1", "EUR")}}, rates)
	if err != nil || missing != valuation.MissingFXRate {
		t.Errorf("a sale in EUR without a rate: %v, %v; want MissingFXRate", missing, err)
	}
}
