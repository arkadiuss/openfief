package answer

import (
	"errors"
	"strings"
	"testing"
)

func TestCalcExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, expr, want string
		rounded          bool
	}{
		{"binary floats would say 0.30000000000000004", "0.1+0.2", "0.3", false},
		{"a long decimal stays whole", "0.1234567890123456789 * 3", "0.3703703670370370367", false},
		{"two long integers multiply exactly", "12345678901234567890 * 98765432109876543210", "1219326311370217952237463801111263526900", false},
		{"precedence", "2 + 3 * 4", "14", false},
		{"parentheses", "(2 + 3) * 4", "20", false},
		{"unary signs", "-(2 - 5) + +1", "4", false},
		{"percent divides by 100", "200 * 5%", "10", false},
		{"percent alone", "5%", "0.05", false},
		{"percent after parentheses", "(10 + 15)%", "0.25", false},
		{"the multiplication and division signs", "6 × 7 ÷ 2", "21", false},
		{"an exact quotient is not rounded", "1 / 8", "0.125", false},
		{"a quotient that does not end is rounded half-even at 20 places", "1/3", "0.33333333333333333333", true},
		{"a tie at the 20th place goes to the even digit, down", "1 / 200000000000000000000", "0", true},
		{"a tie at the 20th place goes to the even digit, up", "3 / 200000000000000000000", "0.00000000000000000002", true},
		{"amounts to move to a 30/70 split", "(6000 + 3000 + 1000) * 30% - 6000", "-3000", false},
		{"spaces and tabs", " 1 +\t2 ", "3", false},
		{"subtraction below zero", "1 - 2.5", "-1.5", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := Calc(tt.expr)
			if err != nil {
				t.Fatalf("Calc(%q) error = %v", tt.expr, err)
			}
			if got.Value.String() != tt.want || got.Rounded != tt.rounded {
				t.Errorf("Calc(%q) = %s (rounded %v), want %s (rounded %v)", tt.expr, got.Value, got.Rounded, tt.want, tt.rounded)
			}
		})
	}
}

func TestCalcRefusals(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, expr, reason string
	}{
		{"empty", "  ", "empty"},
		{"too long", strings.Repeat("1+", 101), "longer than 200"},
		{"division by zero", "1 / (2 - 2)", "divides by zero"},
		{"percent of a zero divisor never happens, but 0 divisor does", "5 ÷ 0", "divides by zero"},
		{"nesting too deep", strings.Repeat("(", 25) + "1" + strings.Repeat(")", 25), "nests deeper than 20"},
		{"signs nest too", strings.Repeat("-", 25) + "1", "nests deeper than 20"},
		{"a literal with too many digits", "1" + strings.Repeat("0", 30), "more than 30 digits"},
		{"a result with too many digits", strings.Repeat("123456789012345678901234567890 * ", 2) + "123456789012345678901234567890", "more than 60 digits"},
		{"a decimal comma", "1,5 + 1", "decimal comma"},
		{"a thousands separator", "1 234 + 1", "not part of an expression"},
		{"a letter", "2 + x", "not part of an expression"},
		{"a function call", "abs(1)", "not part of an expression"},
		{"scientific notation", "1e5", "not part of an expression"},
		{"a missing operand", "2 +", "ends where a number"},
		{"an unclosed parenthesis", "(2 + 3", "ends where a number"},
		{"a stray closing parenthesis", "2 + 3)", "not part of an expression"},
		{"a trailing point", "2. + 1", "decimal point with no digit"},
		{"a second point", "1.2.3", "second point"},
		{"a code injection attempt", "__import__('os').system('id')", "not part of an expression"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Calc(tt.expr)
			ce, ok := errors.AsType[*CalcError](err)
			if !ok || !errors.Is(err, ErrCalc) {
				t.Fatalf("Calc(%q) error = %v, want a *CalcError", tt.expr, err)
			}
			if !strings.Contains(ce.Reason, tt.reason) {
				t.Errorf("Calc(%q) reason = %q, want it to contain %q", tt.expr, ce.Reason, tt.reason)
			}
			if strings.Contains(ce.Reason, "import") || strings.Contains(ce.Reason, "system") {
				t.Errorf("the reason quotes the expression: %q", ce.Reason)
			}
		})
	}
}

func TestCalcCard(t *testing.T) {
	t.Parallel()
	f := Separators(" ", ",")
	card := calcCard(" 1000.5 * 1000 ", f)
	if card.Title != "1000.5 * 1000 = 1000500" {
		t.Errorf("title = %q, want the expression and its result", card.Title)
	}
	if !strings.Contains(card.Text, "Computed from the figures given") {
		t.Errorf("text = %q, want it to say the figures are the ones given", card.Text)
	}
	if len(card.Rows) != 2 || card.Rows[1].Figure != "1"+NarrowSpace+"000"+NarrowSpace+"500" {
		t.Errorf("rows = %+v, want the result in the owner's number format", card.Rows)
	}
	if rounded := calcCard("1/3", f); !strings.Contains(rounded.Text, "rounded half-even") {
		t.Errorf("text = %q, want it to say the quotient was rounded", rounded.Text)
	}
	refused := calcCard("1/0", f)
	if refused.Title != "That can't be computed" || !strings.Contains(refused.Text, "divides by zero") || len(refused.Rows) != 0 {
		t.Errorf("card = %+v, want a refusal with its reason", refused)
	}
}
