package answer

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// The calc tool does exact decimal arithmetic on figures already shown
// (D-228). It reads no data: the numbers are the arguments, which come from
// the model, so its card says they are the figures given. Its parser is a
// small recursive-descent one over a fixed grammar, with no evaluation of
// anything else, and it is bounded in length, nesting, digits and the places
// of a quotient, so no expression a model writes can make it run long or grow
// a number without end.
//
//	expression = term { ("+" | "-") term }
//	term       = factor { ("*" | "/" | "×" | "÷") factor }
//	factor     = ("-" | "+") factor | scaled
//	primary    = number | "(" expression ")"
//	scaled     = primary { "%" }
//	number     = digits [ "." digits ]
//
// A "%" divides what it follows by 100, so 5% is 0.05 and 200 * 5% is 10.

// The bounds of an expression.
const (
	// maxExpressionRunes caps the length of an expression.
	maxExpressionRunes = 200
	// maxNesting caps how deep parentheses and unary signs nest.
	maxNesting = 20
	// maxLiteralDigits caps the digits of one number written in an expression.
	maxLiteralDigits = 30
	// maxResultDigits caps the digits of any intermediate result and of the
	// result: a product of long numbers grows without limit otherwise.
	maxResultDigits = 60
	// quotientPlaces is how many decimal places a division keeps when it does
	// not come out exactly; the quotient is rounded half-even there, once.
	quotientPlaces = 20
)

// CalcResult is the value of an expression: the number, and whether any
// division in it was rounded to quotientPlaces places, so the number is not
// exact.
type CalcResult struct {
	Value kernel.Decimal
	// Rounded reports that a division did not come out in quotientPlaces
	// places or fewer and was rounded half-even there.
	Rounded bool
}

// ErrCalc is every reason an expression is refused; a *CalcError says which.
var ErrCalc = errors.New("the expression cannot be computed")

// CalcError is why an expression was refused. Its text names a position, never
// the expression itself, which the person's figures may be in (GR-12).
type CalcError struct {
	Reason string
}

func (e *CalcError) Error() string { return "calc: " + e.Reason }

// Is reports that a CalcError is ErrCalc.
func (e *CalcError) Is(target error) bool { return target == ErrCalc }

func refused(format string, args ...any) error {
	return &CalcError{Reason: fmt.Sprintf(format, args...)}
}

// Calc evaluates expression exactly, in decimals: + - × ÷ (also * and /),
// parentheses, unary signs and % (divide by 100) on numbers written with a
// point. It never rounds except a division that doesn't end within
// quotientPlaces places, which it rounds half-even once there (CalcResult
// says so). It refuses an expression that is empty or longer than 200
// characters, nests deeper than 20 levels, writes a number of more than 30
// digits, divides by zero, or has a result of more than 60 digits, each with
// a CalcError that names no figure. It is pure: no I/O, no clock, no state.
func Calc(expression string) (CalcResult, error) {
	if strings.TrimSpace(expression) == "" {
		return CalcResult{}, refused("the expression is empty")
	}
	if utf8.RuneCountInString(expression) > maxExpressionRunes {
		return CalcResult{}, refused("the expression is longer than %d characters", maxExpressionRunes)
	}
	p := &calcParser{src: []rune(expression)}
	v, err := p.expression(0)
	if err != nil {
		return CalcResult{}, err
	}
	p.skipSpace()
	if p.pos < len(p.src) {
		return CalcResult{}, p.unexpected()
	}
	return CalcResult{Value: v, Rounded: p.rounded}, nil
}

// calcParser reads an expression left to right.
type calcParser struct {
	src     []rune
	pos     int
	rounded bool
}

func (p *calcParser) skipSpace() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

// peek returns the next character after any space, or 0 at the end.
func (p *calcParser) peek() rune {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

// unexpected is the refusal at the current position: 1-based, as a person
// counts characters.
func (p *calcParser) unexpected() error {
	if p.pos >= len(p.src) {
		return refused("the expression ends where a number or a closing parenthesis was expected")
	}
	return refused("character %d is not part of an expression: it holds numbers written with a point, + - * / × ÷ %%, and parentheses", p.pos+1)
}

// expression reads terms joined by + and -; depth is how deeply it nests.
func (p *calcParser) expression(depth int) (kernel.Decimal, error) {
	left, err := p.term(depth)
	if err != nil {
		return kernel.Decimal{}, err
	}
	for {
		op := p.peek()
		if op != '+' && op != '-' {
			return left, nil
		}
		p.pos++
		right, err := p.term(depth)
		if err != nil {
			return kernel.Decimal{}, err
		}
		if op == '+' {
			left = left.Add(right)
		} else {
			left = left.Sub(right)
		}
		if err := checkDigits(left); err != nil {
			return kernel.Decimal{}, err
		}
	}
}

// term reads factors joined by * and /.
func (p *calcParser) term(depth int) (kernel.Decimal, error) {
	left, err := p.factor(depth)
	if err != nil {
		return kernel.Decimal{}, err
	}
	for {
		op := p.peek()
		if op != '*' && op != '/' && op != '×' && op != '÷' {
			return left, nil
		}
		p.pos++
		right, err := p.factor(depth)
		if err != nil {
			return kernel.Decimal{}, err
		}
		if left, err = p.apply(op, left, right); err != nil {
			return kernel.Decimal{}, err
		}
	}
}

// apply returns left times right for a multiplication sign, else left divided
// by right, refused when it has more digits than a result may.
func (p *calcParser) apply(op rune, left, right kernel.Decimal) (kernel.Decimal, error) {
	if op == '*' || op == '×' {
		left = left.Mul(right)
	} else {
		var err error
		if left, err = p.divide(left, right); err != nil {
			return kernel.Decimal{}, err
		}
	}
	if err := checkDigits(left); err != nil {
		return kernel.Decimal{}, err
	}
	return left, nil
}

// divide returns a ÷ b: exact when it ends within quotientPlaces places,
// else rounded half-even there; a divisor of zero is refused.
func (p *calcParser) divide(a, b kernel.Decimal) (kernel.Decimal, error) {
	if b.IsZero() {
		return kernel.Decimal{}, refused("the expression divides by zero")
	}
	q, err := a.Quo(b, quotientPlaces, kernel.RoundHalfEven)
	if err != nil {
		return kernel.Decimal{}, refused("a division could not be computed")
	}
	if !q.Mul(b).Equal(a) {
		p.rounded = true
	}
	return q, nil
}

// factor reads a sign and what it applies to.
func (p *calcParser) factor(depth int) (kernel.Decimal, error) {
	if depth >= maxNesting {
		return kernel.Decimal{}, refused("the expression nests deeper than %d levels", maxNesting)
	}
	switch p.peek() {
	case '-':
		p.pos++
		v, err := p.factor(depth + 1)
		return v.Neg(), err
	case '+':
		p.pos++
		return p.factor(depth + 1)
	}
	return p.postfix(depth)
}

// postfix reads a primary and the percent signs after it.
func (p *calcParser) postfix(depth int) (kernel.Decimal, error) {
	v, err := p.primary(depth)
	if err != nil {
		return kernel.Decimal{}, err
	}
	for p.peek() == '%' {
		p.pos++
		if v, err = p.divide(v, kernel.NewDecimal(hundred, 0)); err != nil {
			return kernel.Decimal{}, err
		}
	}
	return v, nil
}

// primary reads a number or a parenthesised expression.
func (p *calcParser) primary(depth int) (kernel.Decimal, error) {
	c := p.peek()
	if c == '(' {
		p.pos++
		v, err := p.expression(depth + 1)
		if err != nil {
			return kernel.Decimal{}, err
		}
		if p.peek() != ')' {
			return kernel.Decimal{}, p.unexpected()
		}
		p.pos++
		return v, nil
	}
	if !isDigitRune(c) {
		return kernel.Decimal{}, p.unexpected()
	}
	return p.number()
}

// number reads digits with an optional point and digits.
func (p *calcParser) number() (kernel.Decimal, error) {
	start := p.pos
	digits, points := 0, 0
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case isDigitRune(c):
			digits++
		case c == '.' && points == 0:
			points++
			if p.pos+1 >= len(p.src) || !isDigitRune(p.src[p.pos+1]) {
				return kernel.Decimal{}, refused("character %d is a decimal point with no digit after it", p.pos+1)
			}
		case isSeparator(c):
			return kernel.Decimal{}, refused("character %d is a decimal comma, a second point or a thousands separator: write numbers with one point and no separators, such as 1234.5", p.pos+1)
		default:
			return p.parsed(start, digits)
		}
		p.pos++
	}
	return p.parsed(start, digits)
}

// isDigitRune reports whether c is an ASCII digit.
func isDigitRune(c rune) bool { return c >= '0' && c <= '9' }

// isSeparator reports whether c is a comma, a point or a no-break space, which a
// number written for the owner's format may hold and this one may not.
func isSeparator(c rune) bool { return c == ',' || c == '.' || c == '\u00a0' || c == '\u202f' }

// parsed turns the number read from start into a decimal.
func (p *calcParser) parsed(start, digits int) (kernel.Decimal, error) {
	if digits > maxLiteralDigits {
		return kernel.Decimal{}, refused("a number has more than %d digits", maxLiteralDigits)
	}
	d, err := kernel.ParseDecimal(string(p.src[start:p.pos]))
	if err != nil {
		return kernel.Decimal{}, refused("character %d starts a number that is not one", start+1)
	}
	return d, nil
}

// checkDigits refuses a value with more than maxResultDigits digits.
func checkDigits(d kernel.Decimal) error {
	digits := 0
	for _, c := range d.String() {
		if c >= '0' && c <= '9' {
			digits++
		}
	}
	if digits > maxResultDigits {
		return refused("a result has more than %d digits", maxResultDigits)
	}
	return nil
}

// calcCard answers a calc command: the value of expression computed exactly,
// shown with it, or why it can't be computed. The figures are the model's, not
// a record's, and the card says so.
func calcCard(expression string, f NumberFormat) Card {
	expression = strings.TrimSpace(expression)
	got, err := Calc(expression)
	if err != nil {
		reason := "The expression can't be computed."
		if ce, ok := errors.AsType[*CalcError](err); ok {
			reason = strings.ToUpper(ce.Reason[:1]) + ce.Reason[1:] + "."
		}
		return Card{Title: "That can't be computed", Text: reason}
	}
	value := got.Value.String()
	text := "Computed from the figures given, exactly in decimals: nothing here was read from your record, and a figure that was not on a card of yours is as good as its source."
	if got.Rounded {
		text += fmt.Sprintf(" A division did not come out in %d decimal places, so it was rounded half-even there and the result is not exact.", quotientPlaces)
	}
	return Card{
		Title: expression + " = " + value,
		Text:  text,
		Rows: []Row{
			{Label: "Expression", Figure: expression},
			{Label: "Result", Figure: f.Amount(value)},
		},
	}
}
