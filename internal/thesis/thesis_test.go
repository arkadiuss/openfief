package thesis_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/thesis"
)

// valid returns a complete draft that edit may break.
func valid(edit func(*thesis.Draft)) thesis.Draft {
	d := thesis.Draft{
		Statement:    "Example Corp's services revenue keeps growing faster than its hardware.",
		Instruments:  []string{"EXMP"},
		Horizon:      thesis.HorizonLongTerm,
		Invalidation: []string{"Services revenue growth falls below hardware growth for two quarters."},
		Sources:      []string{"The holdings card, Example Corp."},
	}
	if edit != nil {
		edit(&d)
	}
	return d
}

func TestNewTidiesAndAcceptsACompleteDraft(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   thesis.Draft
		want thesis.Draft
	}{
		{name: "a complete draft", in: valid(nil), want: valid(nil)},
		{
			name: "short term is a horizon", in: valid(func(d *thesis.Draft) { d.Horizon = thesis.HorizonShortTerm }),
			want: valid(func(d *thesis.Draft) { d.Horizon = thesis.HorizonShortTerm }),
		},
		{
			name: "white space is tidied and an instrument named twice is kept once",
			in: valid(func(d *thesis.Draft) {
				d.Statement = "  Growth \r\n\r\n  continues.  "
				d.Instruments = []string{" EXMP ", "exmp", "OTHR  Inc"}
			}),
			want: valid(func(d *thesis.Draft) {
				d.Statement = "Growth\n\ncontinues."
				d.Instruments = []string{"EXMP", "OTHR Inc"}
			}),
		},
		{
			name: "a price is no lease size", in: valid(func(d *thesis.Draft) { d.Sources = []string{"The market price card: closed at 46.15 USD."} }),
			want: valid(func(d *thesis.Draft) { d.Sources = []string{"The market price card: closed at 46.15 USD."} }),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := thesis.New(&tt.in)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("New() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewRefusesAnIncompleteOrOversizeDraftOrOneWithALeaseSize(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", thesis.MaxStatementRunes+1)
	tests := []struct {
		name string
		in   thesis.Draft
	}{
		{"no statement", valid(func(d *thesis.Draft) { d.Statement = "  " })},
		{"a statement too long", valid(func(d *thesis.Draft) { d.Statement = long })},
		{"no instrument", valid(func(d *thesis.Draft) { d.Instruments = []string{" "} })},
		{"too many instruments", valid(func(d *thesis.Draft) {
			d.Instruments = nil
			for i := range thesis.MaxInstruments + 1 {
				d.Instruments = append(d.Instruments, "INSTR"+strings.Repeat("x", i))
			}
		})},
		{"an instrument too long", valid(func(d *thesis.Draft) { d.Instruments = []string{strings.Repeat("x", thesis.MaxInstrumentRunes+1)} })},
		{"no horizon", valid(func(d *thesis.Draft) { d.Horizon = "" })},
		{"a horizon that is neither", valid(func(d *thesis.Draft) { d.Horizon = "medium_term" })},
		{"no invalidation condition", valid(func(d *thesis.Draft) { d.Invalidation = nil })},
		{"no source", valid(func(d *thesis.Draft) { d.Sources = nil })},
		{"a control character", valid(func(d *thesis.Draft) { d.Sources = []string{"a\x07b"} })},
		{"a lease size in the statement", valid(func(d *thesis.Draft) { d.Statement = "Put 5000 PLN into it." })},
		{"the word lease", valid(func(d *thesis.Draft) { d.Invalidation = []string{"Lease 10% more if it works."} })},
		{"a share of the portfolio", valid(func(d *thesis.Draft) { d.Statement = "Hold 5% of the portfolio in it." })},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := thesis.New(&tt.in)
			if !errors.Is(err, thesis.ErrInvalid) {
				t.Fatalf("New() error = %v, want ErrInvalid", err)
			}
			if got.Statement != "" {
				t.Errorf("New() returned a draft with its error")
			}
		})
	}
}

func TestNewErrorsNeverQuoteTheText(t *testing.T) {
	t.Parallel()
	secret := "ZZ-private-statement: put 5000 PLN in"
	d := valid(func(d *thesis.Draft) { d.Statement = secret })
	_, err := thesis.New(&d)
	if err == nil || strings.Contains(err.Error(), "ZZ-private") {
		t.Fatalf("New() error = %v, want an error that quotes nothing", err)
	}
}

func TestMentionsSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		text string
		want bool
	}{
		{"Allocate 2,500 EUR to it", true},
		{"Put 5000 PLN into it", true},
		{"invest about $300 in it", true},
		{"stake 10% on it", true},
		{"Zainwestuj 5000 zł w to", true},
		{"10 % of the satellite sleeve", true},
		{"5% portfela", true},
		{"a position size of one", true},
		{"The lease should be large", true},
		{"Dzierżawa w całości", true},
		{"The company closed at 46.15 USD on the day", false},
		{"Price $46.15, up from USD 41.10", false},
		{"Revenue grew 12% in 2026", false},
		{"Invest in growth through 2027", false},
		{"Dividend yield of 3.1% a year", false},
		{"Unleashed growth", false},
		{"EXMP and OTHR", false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			if got := thesis.MentionsSize(tt.text); got != tt.want {
				t.Errorf("MentionsSize(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestParseHorizon(t *testing.T) {
	t.Parallel()
	for _, h := range thesis.Horizons() {
		got, err := thesis.ParseHorizon(string(h))
		if err != nil || got != h {
			t.Errorf("ParseHorizon(%q) = %q, %v", h, got, err)
		}
	}
	if _, err := thesis.ParseHorizon("forever"); !errors.Is(err, thesis.ErrInvalid) {
		t.Errorf("ParseHorizon(forever) error = %v, want ErrInvalid", err)
	}
	if thesis.HorizonLongTerm.Words() != "Long-term" || thesis.HorizonShortTerm.Words() != "Short-term" {
		t.Errorf("a horizon's words are Long-term and Short-term")
	}
}

func TestLatest(t *testing.T) {
	t.Parallel()
	var empty thesis.Thesis
	if got := empty.Latest(); got.Number != 0 {
		t.Errorf("Latest() of no version = %d, want 0", got.Number)
	}
	th := thesis.Thesis{ID: "x", Versions: []thesis.Version{{Number: 1}, {Number: 2}}}
	if got := th.Latest(); got.Number != 2 {
		t.Errorf("Latest() = %d, want 2", got.Number)
	}
}
