package portfolio_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestKeyFromName(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("abcd ", 20)
	tests := []struct {
		name, in, want string
		wantErr        error
	}{
		{name: "one word", in: "Core", want: "core"},
		{name: "words", in: "Emergency fund", want: "emergency-fund"},
		{name: "Polish letters", in: "Konto główne ŻÓŁĆ", want: "konto-glowne-zolc"},
		{name: "punctuation and spaces collapse", in: "  IBKR -- main (USD)! ", want: "ibkr-main-usd"},
		{name: "digits", in: "IKE 2026", want: "ike-2026"},
		{name: "other scripts dropped", in: "Ünïcode Ω fund", want: "n-code-fund"},
		{name: "cut to 64 without a trailing hyphen", in: long, want: strings.TrimRight(strings.Repeat("abcd-", 13)[:64], "-")},
		{name: "empty", in: "", wantErr: portfolio.ErrNoKey},
		{name: "punctuation only", in: "--- !", wantErr: portfolio.ErrNoKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.KeyFromName(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("KeyFromName(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if got != tt.want {
				t.Errorf("KeyFromName(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if err == nil {
				if verr := portfolio.KindKey(got).Validate(); verr != nil {
					t.Errorf("KeyFromName(%q) = %q, not a valid key: %v", tt.in, got, verr)
				}
			}
		})
	}
}
