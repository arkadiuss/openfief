package jurisdiction_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestFeeRuleValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		rate, minimum, minCur string
		conversion            string
		wantErr               error
	}{
		{name: "the zero rule", rate: "0", minimum: "0", conversion: "0"},
		{name: "0.29% at least 14 PLN, 0.5% conversion", rate: "0.0029", minimum: "14", minCur: "PLN", conversion: "0.005"},
		{name: "negative rate", rate: "-0.001", minimum: "0", conversion: "0", wantErr: jurisdiction.ErrInvalidFeeRule},
		{name: "rate of 100%", rate: "1", minimum: "0", conversion: "0", wantErr: jurisdiction.ErrInvalidFeeRule},
		{name: "negative conversion rate", rate: "0", minimum: "0", conversion: "-0.01", wantErr: jurisdiction.ErrInvalidFeeRule},
		{name: "conversion rate above 100%", rate: "0", minimum: "0", conversion: "1.5", wantErr: jurisdiction.ErrInvalidFeeRule},
		{name: "negative minimum", rate: "0", minimum: "-1", minCur: "PLN", conversion: "0", wantErr: jurisdiction.ErrInvalidFeeRule},
		{name: "minimum without a currency", rate: "0", minimum: "5", conversion: "0", wantErr: jurisdiction.ErrInvalidFeeRule},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := jurisdiction.FeeRule{Rate: dec(t, tt.rate), Minimum: money(t, tt.minimum, tt.minCur), ConversionRate: dec(t, tt.conversion)}
			if diff := cmp.Diff(tt.wantErr, r.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestFeeRuleSaleCost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		rate, min, convRate string
		gross               string
		foreign             bool
		want                string
	}{
		{name: "the zero rule charges nothing", rate: "0", min: "0", convRate: "0", gross: "10000", want: "0"},
		{name: "0.29% of 10 000 = 29.00, above the minimum", rate: "0.0029", min: "14", convRate: "0", gross: "10000", want: "29.00"},
		{name: "0.29% of 1 000 = 2.90 is below the 14.00 minimum", rate: "0.0029", min: "14", convRate: "0", gross: "1000", want: "14"},
		{name: "foreign: commission plus 0.5% conversion", rate: "0.0029", min: "14", convRate: "0.005", gross: "10000", foreign: true, want: "79.00"},
		{name: "conversion only for a foreign instrument", rate: "0", min: "0", convRate: "0.005", gross: "10000", want: "0"},
		{name: "rounds half up to grosze: 0.25% of 1.02 = 0.00255", rate: "0.0025", min: "0", convRate: "0", gross: "1.02", want: "0.00"},
		{name: "rounds half up: 0.5% of 1.01 = 0.00505", rate: "0.005", min: "0", convRate: "0", gross: "1.01", want: "0.01"},
		{name: "nothing sold, nothing charged", rate: "0.0029", min: "14", convRate: "0", gross: "0", want: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := jurisdiction.FeeRule{Rate: dec(t, tt.rate), Minimum: money(t, tt.min, "PLN"), ConversionRate: dec(t, tt.convRate)}
			got, err := r.SaleCost(money(t, tt.gross, "PLN"), r.Minimum, tt.foreign)
			if err != nil {
				t.Fatalf("SaleCost: %v", err)
			}
			if diff := cmp.Diff(money(t, tt.want, "PLN"), got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
	t.Run("minimum in another currency than the value", func(t *testing.T) {
		t.Parallel()
		r := jurisdiction.FeeRule{Rate: dec(t, "0.001"), Minimum: money(t, "1", "USD")}
		if _, err := r.SaleCost(money(t, "100", "PLN"), r.Minimum, false); err == nil {
			t.Error("SaleCost took a USD minimum for a PLN value")
		}
	})
}

// TestSaleCostProperties draws fee rules and values from a seeded
// generator: a sale never costs less than the rule's minimum or its rate of
// the value, and a foreign sale never costs less than a domestic one.
func TestSaleCostProperties(t *testing.T) {
	t.Parallel()
	r := &rng{state: propertySeed}
	for i := range propertyCases {
		rule := jurisdiction.FeeRule{
			Rate:           kernel.NewDecimal(int64(r.intn(100)), -4),
			Minimum:        money(t, kernel.NewDecimal(int64(r.intn(3000)), -2).String(), "PLN"),
			ConversionRate: kernel.NewDecimal(int64(r.intn(100)), -4),
		}
		gross := money(t, kernel.NewDecimal(int64(r.intn(10_000_000)+1), -2).String(), "PLN")
		domestic, err := rule.SaleCost(gross, rule.Minimum, false)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		foreign, err := rule.SaleCost(gross, rule.Minimum, true)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		floor, err := gross.Amount().Mul(rule.Rate).Round(jurisdiction.FeeRulePlaces, kernel.RoundDown)
		if err != nil {
			t.Fatal(err)
		}
		if domestic.Amount().Cmp(rule.Minimum.Amount()) < 0 || domestic.Amount().Cmp(floor) < 0 {
			t.Errorf("case %d: %s on %s costs %s, below its minimum or rate", i, rule.Rate, gross, domestic)
		}
		if foreign.Amount().Cmp(domestic.Amount()) < 0 {
			t.Errorf("case %d: foreign sale %s costs less than domestic %s", i, foreign, domestic)
		}
	}
}

func TestFeeRuleIsZero(t *testing.T) {
	t.Parallel()
	if r := (jurisdiction.FeeRule{}); !r.IsZero() {
		t.Error("the zero rule is not zero")
	}
	if r := (jurisdiction.FeeRule{Minimum: money(t, "1", "PLN")}); r.IsZero() {
		t.Error("a rule with a minimum is zero")
	}
}
