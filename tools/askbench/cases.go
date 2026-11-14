package main

import (
	"bytes"
	_ "embed" // the cases are embedded so the bench runs from any directory
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/arkadiuss/openfief/internal/kernel"
)

//go:embed testdata/cases.yaml
var casesYAML []byte

// The kinds of case.
const (
	kindAnswer         = "answer"
	kindMulti          = "multi"
	kindNoTool         = "no-tool"
	kindAmbiguous      = "ambiguous"
	kindWhy            = "why"
	kindRefusalWrite   = "refusal-write"
	kindRefusalAdvice  = "refusal-advice"
	kindRefusalScope   = "refusal-scope"
	kindInjection      = "injection"
	langEnglish        = "en"
	langPolish         = "pl"
	pointFormatGroup   = ","
	pointFormatDecimal = "."
	commaFormatGroup   = " "
	commaFormatDecimal = ","
)

// ErrCases reports cases the bench can't read or that contradict themselves.
var ErrCases = errors.New("the bench's cases")

// Case is one question of the bench: what is asked, in which language, and what
// a pass looks like in terms code can decide.
type Case struct {
	ID       string `yaml:"id"`
	Lang     string `yaml:"lang"`
	Kind     string `yaml:"kind"`
	Safety   bool   `yaml:"safety"`
	Question string `yaml:"question"`
	Note     string `yaml:"note"`
	Expect   Expect `yaml:"expect"`
	// Cards are the tools whose card the answer must show.
	Cards []string `yaml:"cards"`
	// Checks are the checks beyond the kind's (defaultChecks).
	Checks []string `yaml:"checks"`
}

// Expect is what the model's tool calls must look like.
type Expect struct {
	// NoCalls says the model must call no tool.
	NoCalls bool `yaml:"noCalls"`
	// MaxCalls caps the calls the model asks for; 0 is no cap beyond D-208's.
	MaxCalls int `yaml:"maxCalls"`
	// Forbid lists tools the model must not call.
	Forbid []string `yaml:"forbid"`
	// Calls are the calls the model must make, in any order.
	Calls []ExpectedCall `yaml:"calls"`
}

// ExpectedCall is a call the model must make: a tool and its arguments (see
// the operators in testdata/cases.yaml), or alternatives of which one must.
type ExpectedCall struct {
	Tool  string         `yaml:"tool"`
	Args  map[string]any `yaml:"args"`
	AnyOf []ExpectedCall `yaml:"anyOf"`
}

// templateVars are what a case's text can name: dates of the bench's fixed
// clock and the month the seed's return fell in.
type templateVars struct {
	Today        string
	LastMonthEnd string
	DropMonthEN  string
	DropMonthPL  string
	DropPrevEnd  string
	DropEnd      string
}

// polishMonths are the months in the locative, as "w listopadzie 2025" wants.
func polishMonths() []string {
	return []string{"styczniu", "lutym", "marcu", "kwietniu", "maju", "czerwcu", "lipcu", "sierpniu", "wrześniu", "październiku", "listopadzie", "grudniu"}
}

// varsFor returns the template variables for the portfolio.
func varsFor(pf *Portfolio) (templateVars, error) {
	today := pf.Days[len(pf.Days)-1].Date
	lastEnd := today.AddDays(1 - today.Day()).AddDays(-1)
	v := templateVars{Today: today.String(), LastMonthEnd: lastEnd.String()}
	month, _, ok := pf.DropMonth()
	if !ok {
		return v, fmt.Errorf("%w: the seed has no month its return fell in", ErrCases)
	}
	var first, last kernel.Date
	for i := range pf.Days {
		d := pf.Days[i].Date
		if monthKey(d) != month {
			continue
		}
		if first.IsZero() {
			first = d
		}
		last = d
	}
	year := strconv.Itoa(first.Year())
	v.DropPrevEnd, v.DropEnd = first.AddDays(-1).String(), last.String()
	v.DropMonthEN = first.Month().String() + " " + year
	v.DropMonthPL = polishMonths()[int(first.Month())-1] + " " + year
	return v, nil
}

// LoadCases reads the embedded cases with their dates filled in from pf.
func LoadCases(pf *Portfolio) ([]Case, error) {
	vars, err := varsFor(pf)
	if err != nil {
		return nil, err
	}
	return parseCases(casesYAML, &vars)
}

// parseCases fills raw in with vars and reads it, checking each case.
func parseCases(raw []byte, vars *templateVars) ([]Case, error) {
	tpl, err := template.New("cases").Option("missingkey=error").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCases, err)
	}
	var filled bytes.Buffer
	if execErr := tpl.Execute(&filled, vars); execErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrCases, execErr)
	}
	var cases []Case
	if yamlErr := yaml.Unmarshal(filled.Bytes(), &cases); yamlErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrCases, yamlErr)
	}
	seen := map[string]bool{}
	for i := range cases {
		c := &cases[i]
		if checkErr := c.check(); checkErr != nil {
			return nil, checkErr
		}
		if seen[c.ID] {
			return nil, fmt.Errorf("%w: %s twice", ErrCases, c.ID)
		}
		seen[c.ID] = true
	}
	return cases, nil
}

// check says what is wrong with c, if anything.
func (c *Case) check() error {
	kinds := []string{kindAnswer, kindMulti, kindNoTool, kindAmbiguous, kindWhy, kindRefusalWrite, kindRefusalAdvice, kindRefusalScope, kindInjection}
	switch {
	case c.ID == "" || strings.TrimSpace(c.Question) == "":
		return fmt.Errorf("%w: a case with no id or no question", ErrCases)
	case c.Lang != langEnglish && c.Lang != langPolish:
		return fmt.Errorf("%w: %s: language %q is en or pl", ErrCases, c.ID, c.Lang)
	case !slices.Contains(kinds, c.Kind):
		return fmt.Errorf("%w: %s: unknown kind %q", ErrCases, c.ID, c.Kind)
	case c.Expect.NoCalls && len(c.Expect.Calls) > 0:
		return fmt.Errorf("%w: %s: no calls and calls", ErrCases, c.ID)
	}
	return nil
}

// numberFormat returns the owner's number format a case's language implies:
// a Polish owner groups with a space and writes a decimal comma, an English one
// groups with a comma and writes a point (D-113).
func (c *Case) numberFormat() numberStyle {
	if c.Lang == langPolish {
		return numberStyle{Group: commaFormatGroup, Decimal: commaFormatDecimal}
	}
	return numberStyle{Group: pointFormatGroup, Decimal: pointFormatDecimal}
}

// numberStyle is the separators of an owner's number format.
type numberStyle struct {
	Group, Decimal string
}

// defaultChecks are the checks every case of a kind gets.
func defaultChecks(kind string) []string {
	switch kind {
	case kindAnswer, kindMulti, kindWhy:
		return []string{checkNumbers, checkEntities}
	case kindNoTool:
		return []string{checkNoNumbers}
	case kindRefusalWrite:
		return []string{checkRefusalWrite}
	case kindRefusalAdvice:
		return []string{checkRefusalAdvice, checkNoAdvice}
	case kindRefusalScope:
		return []string{checkRefusalScope, checkSuggests}
	case kindInjection:
		return []string{checkNumbers, checkEntities, checkNoEcho}
	case kindAmbiguous:
		return nil
	}
	return nil
}

// checksOf returns the names of the checks of c: its kind's, then its own.
func (c *Case) checksOf() []string {
	out := slices.Clone(defaultChecks(c.Kind))
	for _, name := range c.Checks {
		if !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	return out
}

// Clock is the fixed time source of a run: the bench has no wall clock beyond
// measuring how long a call took.
type Clock func() time.Time
