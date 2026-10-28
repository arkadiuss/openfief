package valuation_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/valuation"
)

func TestIndexLevelValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, index, date, level string
		wantErr                  error
	}{
		{name: "a synthetic index", index: "FAKE20", date: "2025-01-03", level: "1522.60"},
		{name: "the longest name", index: strings.Repeat("X", 64), date: "2025-01-03", level: "1"},
		{name: "no name", date: "2025-01-03", level: "1", wantErr: valuation.ErrInvalidIndexLevel},
		{name: "a name too long", index: strings.Repeat("X", 65), date: "2025-01-03", level: "1", wantErr: valuation.ErrInvalidIndexLevel},
		{name: "no date", index: "FAKE20", level: "1", wantErr: valuation.ErrInvalidIndexLevel},
		{name: "a zero level", index: "FAKE20", date: "2025-01-03", level: "0", wantErr: valuation.ErrInvalidIndexLevel},
		{name: "a negative level", index: "FAKE20", date: "2025-01-03", level: "-1", wantErr: valuation.ErrInvalidIndexLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := valuation.IndexLevel{Index: tt.index, Level: dec(t, tt.level)}
			if tt.date != "" {
				l.Date = date(t, tt.date)
			}
			if diff := cmp.Diff(tt.wantErr, l.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}
