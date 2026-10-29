package jurisdiction_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
)

func TestNumberFormatKeys(t *testing.T) {
	t.Parallel()
	keys := make([]string, 0, len(jurisdiction.NumberFormats()))
	for _, f := range jurisdiction.NumberFormats() {
		if err := f.Validate(); err != nil {
			t.Errorf("%s: %v", f, err)
		}
		back, err := jurisdiction.ParseNumberFormat(f.String())
		if err != nil {
			t.Errorf("ParseNumberFormat(%q): %v", f, err)
		}
		if back != f {
			t.Errorf("ParseNumberFormat(%q) = %+v, want %+v", f, back, f)
		}
		keys = append(keys, f.String())
	}
	want := []string{"space-comma", "space-dot", "comma-dot", "dot-comma", "apostrophe-dot", "none-comma", "none-dot"}
	if diff := cmp.Diff(want, keys); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
}

func TestNumberFormatValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		format  jurisdiction.NumberFormat
		wantErr error
	}{
		{name: "Polish: space and comma", format: jurisdiction.NumberFormat{Group: jurisdiction.SeparatorSpace, Decimal: jurisdiction.SeparatorComma}},
		{name: "no grouping", format: jurisdiction.NumberFormat{Group: jurisdiction.SeparatorNone, Decimal: jurisdiction.SeparatorDot}},
		{name: "comma for both", format: jurisdiction.NumberFormat{Group: jurisdiction.SeparatorComma, Decimal: jurisdiction.SeparatorComma}, wantErr: jurisdiction.ErrInvalidNumberFormat},
		{name: "space as the decimal separator", format: jurisdiction.NumberFormat{Group: jurisdiction.SeparatorDot, Decimal: jurisdiction.SeparatorSpace}, wantErr: jurisdiction.ErrInvalidNumberFormat},
		{name: "no decimal separator", format: jurisdiction.NumberFormat{Group: jurisdiction.SeparatorDot}, wantErr: jurisdiction.ErrInvalidNumberFormat},
		{name: "an unknown group separator", format: jurisdiction.NumberFormat{Group: "_", Decimal: jurisdiction.SeparatorDot}, wantErr: jurisdiction.ErrInvalidNumberFormat},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.format.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseNumberFormatRejects(t *testing.T) {
	t.Parallel()
	for _, key := range []string{"", "space", "space-", "space-space", "dot-dot", "space-comma-dot", "tab-comma", "space-apostrophe"} {
		if _, err := jurisdiction.ParseNumberFormat(key); !cmp.Equal(jurisdiction.ErrInvalidNumberFormat, err, cmpopts.EquateErrors()) {
			t.Errorf("ParseNumberFormat(%q) = %v, want ErrInvalidNumberFormat", key, err)
		}
	}
}
