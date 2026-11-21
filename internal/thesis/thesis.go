package thesis

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Errors about a draft thesis. The wrapped text names the field and never
// quotes its value (GR-12).
var (
	// ErrInvalid reports a draft that breaks a rule: a missing field, a
	// horizon that is neither long term nor short term, a text too long, or a
	// lease's size, which a thesis does not hold.
	ErrInvalid = errors.New("invalid thesis")
	// ErrNotFound reports a thesis, or a version of one, that does not exist.
	ErrNotFound = errors.New("thesis not found")
	// ErrStale reports an edit made on a version that is no longer the latest:
	// the owner edited the draft elsewhere in the meantime.
	ErrStale = errors.New("the thesis has a newer version")
)

// DataTier is the data tier of a thesis's text and instruments (D-206): T1.
const DataTier = "T1"

// Horizon is how long a thesis is meant to run (GR-5): mainly long term,
// short term allowed.
type Horizon string

// The horizons, a closed set.
const (
	HorizonLongTerm  Horizon = "long_term"
	HorizonShortTerm Horizon = "short_term"
)

// Horizons returns both horizons, long term first.
func Horizons() []Horizon { return []Horizon{HorizonLongTerm, HorizonShortTerm} }

// ParseHorizon parses a horizon as its constant spells it.
func ParseHorizon(s string) (Horizon, error) {
	h := Horizon(s)
	if !slices.Contains(Horizons(), h) {
		return "", fmt.Errorf("%w: the horizon is long_term or short_term", ErrInvalid)
	}
	return h, nil
}

// Words says a horizon in words, for the page.
func (h Horizon) Words() string {
	if h == HorizonShortTerm {
		return "Short-term"
	}
	return "Long-term"
}

// The limits of a draft, in characters (runes) and items.
const (
	MaxStatementRunes    = 1000
	MaxInstruments       = 10
	MaxInstrumentRunes   = 60
	MaxInvalidation      = 8
	MaxInvalidationRunes = 300
	MaxSources           = 10
	MaxSourceRunes       = 200
)

// Draft is what a thesis says: a statement, the instruments it is about, a
// horizon, the conditions that would invalidate it and the sources it rests
// on. It has no field for a lease's size or an amount: how much an agent may
// hold is the lease's, which the owner approves (GR-6).
type Draft struct {
	Statement    string
	Instruments  []string
	Horizon      Horizon
	Invalidation []string
	Sources      []string
}

// New returns d with its text tidied (surrounding space trimmed, runs of
// white space inside a line made one space, an instrument named twice kept
// once) when it is a complete draft: a statement, an instrument, a horizon of
// the two, an invalidation condition and a source, each within its limit, no
// control character, and no lease's size in any of its text. Otherwise it
// returns an error wrapping ErrInvalid that names the field.
func New(d *Draft) (Draft, error) {
	out := Draft{Statement: tidyStatement(d.Statement), Horizon: d.Horizon}
	out.Instruments = dedupe(tidyList(d.Instruments))
	out.Invalidation = tidyList(d.Invalidation)
	out.Sources = tidyList(d.Sources)
	if err := out.validate(); err != nil {
		return Draft{}, err
	}
	return out, nil
}

// Text returns every list item of d: its instruments, invalidation conditions
// and sources, in that order.
func (d *Draft) Text() []string {
	out := slices.Clone(d.Instruments)
	out = append(out, d.Invalidation...)
	return append(out, d.Sources...)
}

// validate checks a tidied draft.
func (d *Draft) validate() error {
	if d.Statement == "" || utf8.RuneCountInString(d.Statement) > MaxStatementRunes {
		return fmt.Errorf("%w: the statement is 1 to %d characters", ErrInvalid, MaxStatementRunes)
	}
	if _, err := ParseHorizon(string(d.Horizon)); err != nil {
		return err
	}
	if err := d.validateLists(); err != nil {
		return err
	}
	all := append(d.Text(), d.Statement)
	if slices.ContainsFunc(all, hasControl) {
		return fmt.Errorf("%w: no control characters", ErrInvalid)
	}
	if slices.ContainsFunc(all, MentionsSize) {
		return fmt.Errorf("%w: a thesis holds no lease size or amount", ErrInvalid)
	}
	return nil
}

// validateLists checks that the instruments, the invalidation conditions and the
// sources are each between one item and their most, each item within its length.
func (d *Draft) validateLists() error {
	lists := []struct {
		name  string
		items []string
		max   int
		runes int
	}{
		{"instruments", d.Instruments, MaxInstruments, MaxInstrumentRunes},
		{"invalidation conditions", d.Invalidation, MaxInvalidation, MaxInvalidationRunes},
		{"sources", d.Sources, MaxSources, MaxSourceRunes},
	}
	for _, l := range lists {
		if len(l.items) == 0 || len(l.items) > l.max {
			return fmt.Errorf("%w: 1 to %d %s", ErrInvalid, l.max, l.name)
		}
		for _, it := range l.items {
			if utf8.RuneCountInString(it) > l.runes {
				return fmt.Errorf("%w: each of the %s is at most %d characters", ErrInvalid, l.name, l.runes)
			}
		}
	}
	return nil
}

// tidyStatement trims the statement and makes its line breaks "\n", keeping
// the lines (a statement may be a short paragraph) and tidying each.
func tidyStatement(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = tidyLine(l)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// tidyLine trims s and makes each run of white space one space.
func tidyLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// tidyList tidies each item of in and drops the empty ones.
func tidyList(in []string) []string {
	var out []string
	for _, s := range in {
		if t := tidyLine(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// dedupe drops an item that an earlier one equals without regard to case.
func dedupe(in []string) []string {
	var out []string
	for _, s := range in {
		if !slices.ContainsFunc(out, func(o string) bool { return strings.EqualFold(o, s) }) {
			out = append(out, s)
		}
	}
	return out
}

// hasControl reports whether s holds a control character other than the line
// break of a statement.
func hasControl(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r != '\n' && unicode.IsControl(r) }) >= 0
}

// sizePatterns are the patterns of what a lease's size looks like in words:
// the word "lease", a share of the portfolio, a sleeve or a budget, a position
// size, and an amount or a percentage that a verb of sizing (allocate, invest,
// put, stake, commit, spend, deploy) puts to work. A price or a figure of the
// company's own, such as "closed at 46.15 USD" or "grew 12%", is no lease size
// and passes. A function, not a variable: the domain holds no package-level
// state.
func sizePatterns() []*regexp.Regexp {
	const unit = `(?:%|pln|zł|zl|eur|usd|gbp|chf|€|\$|£)`
	const verb = `\b(?:allocat\w+|invest(?:ing|s|ed)?|put(?:ting)?|stak(?:e|es|ing)|commit(?:s|ting)?|spend(?:ing)?|deploy(?:s|ing)?|przeznacz\w*|zainwestuj\w*|zaalokuj\w*)`
	return []*regexp.Regexp{
		regexp.MustCompile(`(?i)\blease[sd]?\b|\bdzier[żz]aw\w*`),
		regexp.MustCompile(`(?i)` + verb + `[^.;\n]{0,40}?\d[\d .,]*\s?` + unit),
		regexp.MustCompile(`(?i)` + verb + `[^.;\n]{0,40}?(?:€|\$|£)\s?\d`),
		regexp.MustCompile(`(?i)\d+(?:[.,]\d+)?\s?%\s+(?:of\s+)?(?:the\s+|your\s+|my\s+|our\s+)?(?:portfolio|sleeve|satellite|fief|lease|budget|pool|capital|net worth|portfel\w*)`),
		regexp.MustCompile(`(?i)\bposition size\b|\bwielko[śs][ćc] pozycji\b`),
	}
}

// MentionsSize reports whether s names a lease's size or an amount of money:
// a thesis holds none (D-229), and a model's suggestion that does is no
// suggestion. A heuristic over fixed words, not a parser: it only ever
// refuses.
func MentionsSize(s string) bool {
	return slices.ContainsFunc(sizePatterns(), func(p *regexp.Regexp) bool { return p.MatchString(s) })
}

// Version is one saved text of a thesis: a draft with its number, from 1, and
// when it was saved.
type Version struct {
	Number  int
	SavedAt time.Time
	Draft   Draft
}

// Thesis is a draft thesis with every version it has had, oldest first.
type Thesis struct {
	ID       string
	Versions []Version
}

// Latest returns the newest version; the zero Version when there is none.
func (t *Thesis) Latest() Version {
	if len(t.Versions) == 0 {
		return Version{}
	}
	return t.Versions[len(t.Versions)-1]
}

// Summary is a draft thesis as the list shows it: its latest version, with
// when it was first saved.
type Summary struct {
	ID        string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Draft     Draft
}
