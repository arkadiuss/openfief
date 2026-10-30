package jurisdiction_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
)

func TestPackWrappers(t *testing.T) {
	t.Parallel()
	p := pl.NewPack()
	if got := jurisdiction.DefaultWrapper(p); got != "taxable" {
		t.Errorf("DefaultWrapper = %q, want taxable", got)
	}
	for _, w := range []jurisdiction.Wrapper{"taxable", "ike", "ikze", "oki"} {
		if !jurisdiction.HasWrapper(p, w) {
			t.Errorf("HasWrapper(%q) = false", w)
		}
	}
	for _, w := range []jurisdiction.Wrapper{"", "IKE", "ppk"} {
		if jurisdiction.HasWrapper(p, w) {
			t.Errorf("HasWrapper(%q) = true", w)
		}
	}
}

func TestValidateMarginalRate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		rate    string
		wantErr error
	}{
		{rate: "0.12"},
		{rate: "0.32"},
		{rate: "0.999"},
		{rate: "0", wantErr: jurisdiction.ErrInvalidMarginalRate},
		{rate: "-0.12", wantErr: jurisdiction.ErrInvalidMarginalRate},
		{rate: "1", wantErr: jurisdiction.ErrInvalidMarginalRate},
		{rate: "32", wantErr: jurisdiction.ErrInvalidMarginalRate},
	}
	for _, tt := range tests {
		t.Run(tt.rate, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, jurisdiction.ValidateMarginalRate(dec(t, tt.rate)), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}
