package ibkrflex

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

// ErrMalformed reports input that is not a well-formed Flex Query statement.
var ErrMalformed = errors.New("malformed Flex statement")

// ErrTooLarge reports input over one of the Limits.
var ErrTooLarge = errors.New("flex statement exceeds a limit")

// Limits caps what a statement may contain, so that a hostile file cannot
// exhaust memory or time. The zero value of a field means its default.
type Limits struct {
	// MaxBytes caps the input size.
	MaxBytes int64
	// MaxDepth caps element nesting.
	MaxDepth int
	// MaxElements caps the number of elements.
	MaxElements int
	// MaxAttrs caps the attributes of one element.
	MaxAttrs int
	// MaxAttrLen caps the length of one attribute value.
	MaxAttrLen int
}

// The default limits: well above a year of daily activity in one account
// (IBKR's Flex rows have about a hundred attributes and nest five deep).
const (
	defaultMaxBytes    = 64 << 20
	defaultMaxDepth    = 16
	defaultMaxElements = 500_000
	defaultMaxAttrs    = 512
	defaultMaxAttrLen  = 4096
)

// withDefaults returns l with every zero field set to its default.
func (l Limits) withDefaults() Limits {
	if l.MaxBytes <= 0 {
		l.MaxBytes = defaultMaxBytes
	}
	if l.MaxDepth <= 0 {
		l.MaxDepth = defaultMaxDepth
	}
	if l.MaxElements <= 0 {
		l.MaxElements = defaultMaxElements
	}
	if l.MaxAttrs <= 0 {
		l.MaxAttrs = defaultMaxAttrs
	}
	if l.MaxAttrLen <= 0 {
		l.MaxAttrLen = defaultMaxAttrLen
	}
	return l
}

// The element depths of a Flex Query response.
const (
	depthResponse  = 1 // FlexQueryResponse
	depthStatement = 3 // FlexStatements > FlexStatement
	depthSection   = 4 // Trades, CashTransactions, ...
	depthRow       = 5 // Trade, CashTransaction, ...
)

// row is one element inside a statement's section, such as one Trade, with
// its attributes.
type row struct {
	elem  string
	index int // 1-based position in its section
	attrs map[string]string
}

// statement is one FlexStatement: one account over one period.
type statement struct {
	attrs    map[string]string
	sections map[string][]row
}

// decoder reads a Flex Query response within limits.
type decoder struct {
	limits     Limits
	depth      int
	elements   int
	statements []statement
	// inStatement is true inside a FlexStatement element.
	inStatement bool
	section     string
}

// decode reads every FlexStatement of a FlexQueryResponse from r.
func decode(r io.Reader, limits Limits) ([]statement, error) {
	d := decoder{limits: limits.withDefaults()}
	// One byte over the cap tells a file at the cap from one beyond it.
	lr := &io.LimitedReader{R: r, N: d.limits.MaxBytes + 1}
	x := xml.NewDecoder(lr)
	x.Strict = true
	for {
		tok, err := x.Token()
		if lr.N <= 0 {
			return nil, fmt.Errorf("%w: more than %d bytes", ErrTooLarge, d.limits.MaxBytes)
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
		}
		if err := d.token(tok); err != nil {
			return nil, err
		}
	}
	if d.elements == 0 {
		return nil, fmt.Errorf("%w: no FlexQueryResponse", ErrMalformed)
	}
	return d.statements, nil
}

// token handles one XML token.
func (d *decoder) token(tok xml.Token) error {
	switch t := tok.(type) {
	case xml.StartElement:
		return d.start(&t)
	case xml.EndElement:
		d.depth--
		return nil
	case xml.Directive:
		return fmt.Errorf("%w: DOCTYPE and other directives are not allowed", ErrMalformed)
	}
	return nil
}

// start handles an element's start tag.
func (d *decoder) start(t *xml.StartElement) error {
	d.depth++
	d.elements++
	if err := d.check(t); err != nil {
		return err
	}
	attrs, err := d.attrs(t)
	if err != nil {
		return err
	}
	switch d.depth {
	case depthResponse:
		if t.Name.Local != "FlexQueryResponse" {
			return fmt.Errorf("%w: root is <%s>, not <FlexQueryResponse>", ErrMalformed, t.Name.Local)
		}
	case depthStatement:
		d.inStatement = t.Name.Local == "FlexStatement"
		if d.inStatement {
			d.statements = append(d.statements, statement{attrs: attrs, sections: map[string][]row{}})
		}
	case depthSection:
		d.section = t.Name.Local
	case depthRow:
		d.addRow(t.Name.Local, attrs)
	}
	return nil
}

// check fails when the element that starts with t goes over a limit.
func (d *decoder) check(t *xml.StartElement) error {
	switch {
	case d.depth > d.limits.MaxDepth:
		return fmt.Errorf("%w: nested deeper than %d", ErrTooLarge, d.limits.MaxDepth)
	case d.elements > d.limits.MaxElements:
		return fmt.Errorf("%w: more than %d elements", ErrTooLarge, d.limits.MaxElements)
	case len(t.Attr) > d.limits.MaxAttrs:
		return fmt.Errorf("%w: <%s> has more than %d attributes", ErrTooLarge, t.Name.Local, d.limits.MaxAttrs)
	}
	return nil
}

// addRow appends a row to the current section of the current statement.
func (d *decoder) addRow(elem string, attrs map[string]string) {
	if !d.inStatement {
		return
	}
	s := &d.statements[len(d.statements)-1]
	rows := s.sections[d.section]
	s.sections[d.section] = append(rows, row{elem: elem, index: len(rows) + 1, attrs: attrs})
}

// attrs returns t's attributes by local name, checking their length.
func (d *decoder) attrs(t *xml.StartElement) (map[string]string, error) {
	m := make(map[string]string, len(t.Attr))
	for _, a := range t.Attr {
		if len(a.Value) > d.limits.MaxAttrLen {
			return nil, fmt.Errorf("%w: attribute %s of <%s> is longer than %d", ErrTooLarge, a.Name.Local, t.Name.Local, d.limits.MaxAttrLen)
		}
		m[a.Name.Local] = a.Value
	}
	return m, nil
}
