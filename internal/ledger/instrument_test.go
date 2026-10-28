package ledger_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger"
)

func TestParseISIN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		wantErr error
	}{
		{name: "public example with letters in the check sum", in: "US0378331005"},
		{name: "synthetic PL", in: isinA},
		{name: "synthetic IE", in: isinB},
		{name: "synthetic XS", in: isinC},
		{name: "wrong check digit", in: "US0378331006", wantErr: ledger.ErrInvalidISIN},
		{name: "two digits swapped", in: "US0378331050", wantErr: ledger.ErrInvalidISIN},
		{name: "lower case", in: "us0378331005", wantErr: ledger.ErrInvalidISIN},
		{name: "digit in the country code", in: "U10378331005", wantErr: ledger.ErrInvalidISIN},
		{name: "letter as check digit", in: "US037833100A", wantErr: ledger.ErrInvalidISIN},
		{name: "too short", in: "US037833100", wantErr: ledger.ErrInvalidISIN},
		{name: "too long", in: "US03783310050", wantErr: ledger.ErrInvalidISIN},
		{name: "empty", in: "", wantErr: ledger.ErrInvalidISIN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.ParseISIN(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseISIN(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if err != nil {
				if !got.IsZero() {
					t.Errorf("ParseISIN(%q) = %q on error, want the zero ISIN", tt.in, got)
				}
				return
			}
			if diff := cmp.Diff(tt.in, got.String()); diff != "" {
				t.Errorf("String() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInstrumentValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		isin    string
		symbol  string
		wantErr error
	}{
		{name: "ISIN and symbol", isin: isinA, symbol: "TSTA"},
		{name: "no symbol", isin: isinA, wantErr: ledger.ErrInvalidInstrument},
		{name: "no ISIN", symbol: "TSTA", wantErr: ledger.ErrInvalidInstrument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := ledger.Instrument{Symbol: tt.symbol}
			if tt.isin != "" {
				in = instrument(t, tt.isin, tt.symbol)
			}
			if diff := cmp.Diff(tt.wantErr, in.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}
