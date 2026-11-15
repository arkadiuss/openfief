package main

import (
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/views/ui"
)

// checkPremise is the check of a question whose premise is false: it names the
// case's Premise, the truth the cards show, and what the answer must make of it.
const checkPremise = "premise"

// Premise is what a case's question wrongly assumes: that the value dropped in
// a month when the cards show it rose. A run passes only when the model looked
// at the month, says the value rose and never that it fell, gives either the
// value at the start and at the end of the month or the change the card shows
// (its money, its percent, or both), each as the card displays it, and (the
// numbers and causes checks) writes no number and names no cause the cards
// don't hold.
type Premise struct {
	// Actual is what the cards show: "rose".
	Actual string `yaml:"actual"`
	// Between are the first and the last day whose value is the month's start
	// and end: the end of the month before and the end of the month.
	Between []string `yaml:"between"`
}

// premiseCorrected checks that the answer sets the false premise right: it says
// the value rose, never that it fell, and gives either the month's start and end
// values from the cards, or the change a period card shows: a figure of its
// change rows (money or percent) written exactly as the card displays it, so a
// percent rounded another way ("+3.5%" for "+3.47%") is no figure of the card.
func premiseCorrected(d *RunData) string {
	p := d.Case.Expect.Premise
	if p == nil || len(p.Between) != 2 {
		return ""
	}
	answer := d.answer()
	if w := assertsFall(answer); w != "" {
		return "says the value " + w + ", which the cards don't show"
	}
	if !saysRise(answer) {
		return "does not say the value rose"
	}
	if figures := monthFigures(d.Run.Cards, p.Between[0], p.Between[1]); len(figures) >= 2 && figuresWritten(d, figures) >= 2 {
		return ""
	}
	if changes := monthChanges(d.Run.Cards, p.Between[0], p.Between[1]); exactlyWritten(d, changes) {
		return ""
	}
	return "does not give the value at the start and at the end of the month, or the change, as the cards show them"
}

// The words by which an answer says a value went up or down, as whole words,
// and the stems of the Polish ones.
func riseWords() []string {
	return []string{"rose", "risen", "rise", "rises", "rising", "increase", "increased", "increases", "grew", "grown", "gain", "gained", "higher"}
}

func riseStems() []string {
	return []string{"wzrost", "wzrosł", "urosł", "zwiększ", "wyższ", "rosł"}
}

func fallWords() []string {
	return []string{
		"fell", "fall", "falls", "fallen", "falling", "drop", "drops", "dropped", "dropping", "decline", "declined", "declines",
		"decrease", "decreased", "decreases", "dip", "dipped", "lost",
	}
}

func fallStems() []string {
	return []string{"spadł", "spadek", "spadk", "zmalał", "zmniejszy", "obniży"}
}

// negators are the words that turn a fall word into its denial ("no drop",
// "did not drop", "rather than dropped"), whether they stand up to
// negationWindow words before it.
func negators() []string {
	return []string{"no", "not", "never", "cannot", "without", "than", "instead", "nie", "brak", "bez"}
}

// negationWindow is how many words before a fall word a negator may stand.
const negationWindow = 4

// words returns text, normalized, as its words without the punctuation around
// them.
func words(text string) []string {
	fields := strings.Fields(normalize(text))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, strings.TrimFunc(f, func(r rune) bool { return !unicode.IsLetter(r) && r != '\'' }))
	}
	return out
}

// isNegator reports whether w denies what follows it.
func isNegator(w string) bool {
	return slices.Contains(negators(), w) || strings.HasSuffix(w, "n't")
}

// assertsFall returns the fall word of the answer that no negator stands before,
// or "".
func assertsFall(answer string) string {
	ws := words(answer)
	for i, w := range ws {
		if !slices.Contains(fallWords(), w) && !hasAnyPrefix(w, fallStems()) {
			continue
		}
		if !slices.ContainsFunc(ws[max(0, i-negationWindow):i], isNegator) {
			return w
		}
	}
	return ""
}

// saysRise reports whether the answer says a value went up.
func saysRise(answer string) bool {
	return slices.ContainsFunc(words(answer), func(w string) bool {
		return slices.Contains(riseWords(), w) || hasAnyPrefix(w, riseStems())
	})
}

// hasAnyPrefix reports whether w starts with one of the stems.
func hasAnyPrefix(w string, stems []string) bool {
	return slices.ContainsFunc(stems, func(s string) bool { return strings.HasPrefix(w, s) })
}

// valueOnDate finds the day a value card is of, in its title.
var valueOnDate = regexp.MustCompile(`Value on (\d{4}-\d{2}-\d{2})`)

// valueOverDaysGroups is the match of valueOverDays and its two days.
const valueOverDaysGroups = 3

// valueOverDays finds the two days a value card of a period is of, in its text:
// the end of the day before the period and the end of its last day.
var valueOverDays = regexp.MustCompile(`at the end of (\d{4}-\d{2}-\d{2}) and at the end of (\d{4}-\d{2}-\d{2})`)

// figuresWritten returns how many of the figures the answer writes.
func figuresWritten(d *RunData, figures []kernel.Decimal) int {
	nums := numbersIn(d.Run.Text)
	found := 0
	for _, f := range figures {
		if slices.ContainsFunc(nums, func(n foundNumber) bool { return grounded(&n, []kernel.Decimal{f}) }) {
			found++
		}
	}
	return found
}

// exactlyWritten reports whether the answer writes one of the figures exactly:
// as the same number, whatever the separators, not rounded to fewer places.
func exactlyWritten(d *RunData, figures []kernel.Decimal) bool {
	return slices.ContainsFunc(numbersIn(d.Run.Text), func(n foundNumber) bool {
		return slices.ContainsFunc(n.readings, func(r reading) bool {
			return slices.ContainsFunc(figures, func(f kernel.Decimal) bool { return r.value.Equal(f) })
		})
	})
}

// monthChanges returns the figures of the change rows (Gross, Net and
// Liquidation) of a value card of a period within the days first to last: the
// change in money and its percent, as the card displays them.
func monthChanges(cards []ui.AskRunCard, first, last string) []kernel.Decimal {
	var out []kernel.Decimal
	for i := range cards {
		c := &cards[i]
		m := valueOverDays.FindStringSubmatch(c.Text)
		if c.Tool != "value" || len(m) != valueOverDaysGroups || m[1] < first || m[2] > last {
			continue
		}
		for _, r := range c.Rows {
			if r.Label == "Gross" || r.Label == "Net" || r.Label == "Liquidation" {
				out = append(out, cardNumbers([]string{r.Figure})...)
			}
		}
	}
	return out
}

// monthFigures returns the distinct figures the cards give for days from first
// to last, as exact numbers: the gross of a value card of such a day, the gross
// at the start and at the end of a value card of a period within them, the Start
// and the End of a performance card when their day is one.
func monthFigures(cards []ui.AskRunCard, first, last string) []kernel.Decimal {
	var out []kernel.Decimal
	for i := range cards {
		for _, figure := range cardMonthFigures(&cards[i], first, last) {
			if d, ok := firstNumber(figure); ok && !slices.ContainsFunc(out, d.Equal) {
				out = append(out, d)
			}
		}
	}
	return out
}

// cardMonthFigures returns the texts of the figures of c that belong to days
// from first to last.
func cardMonthFigures(c *ui.AskRunCard, first, last string) []string {
	var out []string
	switch c.Tool {
	case "value":
		out = valueFigures(c, first, last)
	case "performance":
		out = performanceFigures(c, first, last)
	}
	return out
}

// valueFigures returns the gross figures of a value card whose day, or whose
// first and last day, are from first to last: a day's gross, or a period's gross
// at its start and at its end.
func valueFigures(c *ui.AskRunCard, first, last string) []string {
	if m := valueOnDate.FindStringSubmatch(c.Title); m != nil {
		if m[1] < first || m[1] > last {
			return nil
		}
		return grossRows(c, func(r *ui.AskRunRow) []string { return []string{r.Figure} })
	}
	m := valueOverDays.FindStringSubmatch(c.Text)
	if len(m) != valueOverDaysGroups || m[1] < first || m[2] > last {
		return nil
	}
	return grossRows(c, func(r *ui.AskRunRow) []string { return strings.Split(r.Detail, " to ") })
}

// grossRows returns what figures reads of each Gross row of c.
func grossRows(c *ui.AskRunCard, figures func(r *ui.AskRunRow) []string) []string {
	var out []string
	for i := range c.Rows {
		if c.Rows[i].Label == "Gross" {
			out = append(out, figures(&c.Rows[i])...)
		}
	}
	return out
}

// performanceFigures returns the Start and End figures of a performance card
// whose day is from first to last.
func performanceFigures(c *ui.AskRunCard, first, last string) []string {
	var out []string
	for _, r := range c.Rows {
		if (r.Label == "Start" || r.Label == "End") && r.Detail >= first && r.Detail <= last {
			out = append(out, r.Figure)
		}
	}
	return out
}

// firstNumber returns the first number a figure's text holds.
func firstNumber(figure string) (kernel.Decimal, bool) {
	nums := cardNumbers([]string{figure})
	if len(nums) == 0 {
		return kernel.Decimal{}, false
	}
	return nums[0], true
}
