package kernel_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestParseCurrency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		code    string
		want    string
		wantErr error
	}{
		{name: "PLN", code: "PLN", want: "PLN"},
		{name: "USD", code: "USD", want: "USD"},
		{name: "lower case", code: "pln", wantErr: kernel.ErrInvalidCurrency},
		{name: "two letters", code: "PL", wantErr: kernel.ErrInvalidCurrency},
		{name: "four letters", code: "PLNX", wantErr: kernel.ErrInvalidCurrency},
		{name: "digits", code: "985", wantErr: kernel.ErrInvalidCurrency},
		{name: "empty", code: "", wantErr: kernel.ErrInvalidCurrency},
		{name: "non-ASCII", code: "ZŁ", wantErr: kernel.ErrInvalidCurrency},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := kernel.ParseCurrency(tt.code)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseCurrency(%q) error (-want +got):\n%s", tt.code, diff)
			}
			if diff := cmp.Diff(tt.want, got.String()); diff != "" {
				t.Errorf("ParseCurrency(%q) (-want +got):\n%s", tt.code, diff)
			}
		})
	}
}
