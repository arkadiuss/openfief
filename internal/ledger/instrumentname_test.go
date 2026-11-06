package ledger_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ledger"
)

// Synthetic ISINs with valid check digits; no real instrument is implied.
const (
	nameISINA = "PLTEST000010"
	nameISINB = "IETEST000026"
)

// isinOf parses s, failing t on an error.
func isinOf(t *testing.T, s string) ledger.ISIN {
	t.Helper()
	i, err := ledger.ParseISIN(s)
	if err != nil {
		t.Fatalf("ParseISIN(%q): %v", s, err)
	}
	return i
}

func TestCleanInstrumentName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, want string
		err            bool
	}{
		{name: "kept as it is", in: "Synthetic World Equity UCITS ETF", want: "Synthetic World Equity UCITS ETF"},
		{name: "space trimmed and collapsed", in: "  Synthetic \t Bond\nFund ", want: "Synthetic Bond Fund"},
		{name: "Polish letters", in: "Zażółć Gęślą SA", want: "Zażółć Gęślą SA"},
		{name: "empty", in: " \t ", err: true},
		{name: "a control character", in: "Synthetic\x00Fund", err: true},
		{name: "invalid UTF-8", in: "Synthetic \xff Fund", err: true},
		{name: "at the cap", in: strings.Repeat("ż", ledger.MaxInstrumentNameRunes), want: strings.Repeat("ż", ledger.MaxInstrumentNameRunes)},
		{name: "over the cap", in: strings.Repeat("a", ledger.MaxInstrumentNameRunes+1), err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ledger.CleanInstrumentName(tt.in)
			if tt.err != errors.Is(err, ledger.ErrInvalidInstrumentName) {
				t.Fatalf("CleanInstrumentName(%q) error = %v, want an error %v", tt.in, err, tt.err)
			}
			if got != tt.want {
				t.Errorf("CleanInstrumentName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCurrentInstrumentNames(t *testing.T) {
	t.Parallel()
	a, b := isinOf(t, nameISINA), isinOf(t, nameISINB)
	delivered := func(i ledger.ISIN, name string) ledger.InstrumentName {
		return ledger.InstrumentName{ISIN: i, Name: name, Source: "xtb-main"}
	}
	owners := func(i ledger.ISIN, name string) ledger.InstrumentName {
		return ledger.InstrumentName{ISIN: i, Name: name, ByOwner: true}
	}
	tests := []struct {
		name    string
		records []ledger.InstrumentName
		want    map[ledger.ISIN]string
	}{
		{name: "nothing recorded", want: map[ledger.ISIN]string{}},
		{
			name:    "the first delivered name wins",
			records: []ledger.InstrumentName{delivered(a, "Synthetic A"), delivered(a, "SYNTH A ETF"), delivered(b, "Synthetic B")},
			want:    map[ledger.ISIN]string{a: "Synthetic A", b: "Synthetic B"},
		},
		{
			name:    "the owner's latest rename wins over every delivered name",
			records: []ledger.InstrumentName{delivered(a, "Synthetic A"), owners(a, "My A"), delivered(a, "Later A"), owners(a, "My own A")},
			want:    map[ledger.ISIN]string{a: "My own A"},
		},
		{
			name:    "a rename before any delivered name",
			records: []ledger.InstrumentName{owners(b, "My B"), delivered(b, "Synthetic B")},
			want:    map[ledger.ISIN]string{b: "My B"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, ledger.CurrentInstrumentNames(tt.records), cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
				t.Errorf("CurrentInstrumentNames (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFirstDeliveredNames(t *testing.T) {
	t.Parallel()
	a, b := isinOf(t, nameISINA), isinOf(t, nameISINB)
	in := []ledger.InstrumentName{
		{ISIN: a, Name: "  Synthetic   A ", Source: "ibkr-main"},
		{ISIN: a, Name: "Synthetic A again", Source: "ibkr-main"},
		{ISIN: b, Name: "\x01", Source: "ibkr-main"},
		{Name: "No ISIN", Source: "ibkr-main"},
		{ISIN: b, Name: "Synthetic B", Source: "ibkr-main"},
	}
	want := []ledger.InstrumentName{
		{ISIN: a, Name: "Synthetic A", Source: "ibkr-main"},
		{ISIN: b, Name: "Synthetic B", Source: "ibkr-main"},
	}
	if diff := cmp.Diff(want, ledger.FirstDeliveredNames(in), cmp.Comparer(ledger.ISIN.Equal)); diff != "" {
		t.Errorf("FirstDeliveredNames (-want +got):\n%s", diff)
	}
}
