package portfolio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// Errors of the model-portfolio maintainer (D-204). None of them quotes what
// a model wrote: a note and a reason are text the owner's figures shaped
// (GR-12).
var (
	// ErrInvalidAnswer reports a model's answer that can't be used at all: not
	// the JSON of the answer schema, or with no note, or a note too long.
	ErrInvalidAnswer = errors.New("invalid maintainer answer")
	// ErrInvalidProposal reports a set of target changes that breaks a rule of
	// CheckChanges: it is dropped whole, never in part.
	ErrInvalidProposal = errors.New("invalid target change proposal")
	// ErrStaleProposal reports a proposal whose "before" weights are no longer
	// the model portfolio's targets: the owner changed the model since.
	ErrStaleProposal = errors.New("the model portfolio changed since the proposal was made")
)

const (
	// MaxNoteRunes bounds a maintainer's note.
	MaxNoteRunes = 2000
	// MaxReasonRunes bounds the one-line reason of a change.
	MaxReasonRunes = 200
)

// Change is one proposed change of a target's weight, with the figures
// behind it (D-204): the target's current weight in the portfolio and its
// drift from the target, as OpenFief computed them when it asked.
type Change struct {
	// Target is the target's key (Target.Key): its kind and name.
	Target string
	// Before is the target's weight in the model, and After the proposed one,
	// in percent.
	Before, After kernel.Decimal
	// Reason is the model's one line.
	Reason string
	// Held is false for an index, which has no current weight; Current and
	// Drift are zero then.
	Held           bool
	Current, Drift kernel.Decimal
}

// Proposal is a set of changes that, applied together, leave the model
// portfolio's weights summing to 100%.
type Proposal struct {
	Changes []Change
}

// Answer is what the maintainer's model said, after checking it: the note,
// and the proposal when it passed every rule. Dropped is true when the model
// proposed changes and they broke a rule: the note stays, the changes are
// gone.
type Answer struct {
	Note     string
	Proposal *Proposal
	Dropped  bool
}

// answerJSON and changeJSON are the answer schema as the model writes it.
// Weights are decimal strings, never numbers, so no float touches them.
type answerJSON struct {
	Note    string       `json:"note"`
	Changes []changeJSON `json:"changes"`
}

type changeJSON struct {
	Target         string `json:"target"`
	CurrentTarget  string `json:"currentTargetPercent"`
	ProposedTarget string `json:"proposedTargetPercent"`
	Reason         string `json:"reason"`
}

// ResponseSchema is the JSON Schema the model's answer must follow, as
// strict structured output asks: a note, and a list of changes, usually empty.
func ResponseSchema() string {
	return `{"type":"object","additionalProperties":false,"required":["note","changes"],"properties":{` +
		`"note":{"type":"string","description":"A short plain-text note, at most 120 words."},` +
		`"changes":{"type":"array","description":"Proposed target changes; empty when none.","items":{` +
		`"type":"object","additionalProperties":false,` +
		`"required":["target","currentTargetPercent","proposedTargetPercent","reason"],"properties":{` +
		`"target":{"type":"string","description":"The target as listed: kind, a colon and its name."},` +
		`"currentTargetPercent":{"type":"string","description":"Its target weight now, as listed."},` +
		`"proposedTargetPercent":{"type":"string","description":"Its proposed target weight, a decimal with at most two places."},` +
		`"reason":{"type":"string","description":"One line, at most 200 characters."}}}}}}`
}

// ParseAnswer reads and checks content, the model's answer to a maintainer
// prompt over drift, whose targets are model's. An answer that is not the
// schema's JSON, or has no note or too long a note, is an error wrapping
// ErrInvalidAnswer: nothing of it is kept. A usable note with changes that
// break a rule of CheckChanges is an Answer with Dropped set and no
// proposal; with no changes, an Answer with the note alone.
func ParseAnswer(content string, model *Model, drift *Drift) (Answer, error) {
	var raw answerJSON
	dec := json.NewDecoder(strings.NewReader(content))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return Answer{}, fmt.Errorf("%w: not the answer schema's JSON", ErrInvalidAnswer)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return Answer{}, fmt.Errorf("%w: text after the JSON", ErrInvalidAnswer)
	}
	note := strings.TrimSpace(raw.Note)
	if note == "" || utf8.RuneCountInString(note) > MaxNoteRunes {
		return Answer{}, fmt.Errorf("%w: a note of 1 to %d characters is needed", ErrInvalidAnswer, MaxNoteRunes)
	}
	out := Answer{Note: note}
	if len(raw.Changes) == 0 {
		return out, nil
	}
	changes, ok := checkedChanges(raw.Changes, model, drift)
	if !ok {
		out.Dropped = true
		return out, nil
	}
	out.Proposal = &Proposal{Changes: changes}
	return out, nil
}

// checkedChanges reads the changes of an answer and checks them
// (CheckChanges); false means a bad proposal, which the caller drops.
func checkedChanges(raw []changeJSON, model *Model, drift *Drift) ([]Change, bool) {
	changes, err := changesFrom(raw)
	if err != nil {
		return nil, false
	}
	return changes, CheckChanges(model, drift, changes) == nil
}

// changesFrom reads the changes of an answer, with their weights as
// decimals; the rules are CheckChanges'.
func changesFrom(raw []changeJSON) ([]Change, error) {
	out := make([]Change, 0, len(raw))
	for i := range raw {
		before, err := kernel.ParseDecimal(raw[i].CurrentTarget)
		if err != nil {
			return nil, fmt.Errorf("%w: change %d: a current weight that is not a decimal", ErrInvalidProposal, i+1)
		}
		after, err := kernel.ParseDecimal(raw[i].ProposedTarget)
		if err != nil {
			return nil, fmt.Errorf("%w: change %d: a proposed weight that is not a decimal", ErrInvalidProposal, i+1)
		}
		out = append(out, Change{Target: raw[i].Target, Before: before, After: after, Reason: strings.TrimSpace(raw[i].Reason)})
	}
	return out, nil
}

// CheckChanges checks changes against model and fills in each one's figures
// from drift. The rules, all of them or the set is invalid (an error
// wrapping ErrInvalidProposal):
//
//   - at least one change, and no more than the model has targets;
//   - each change names a target of the model, once: no target is added or
//     removed;
//   - its "before" weight is the target's weight in the model, and "after"
//     differs from it;
//   - its reason is one line of 1 to MaxReasonRunes characters;
//   - the model with every change applied is a valid model (Model.Validate):
//     each weight above 0% with at most WeightPlaces decimal places, the
//     weights summing to exactly 100%.
func CheckChanges(model *Model, drift *Drift, changes []Change) error {
	if len(changes) == 0 || len(changes) > len(model.Targets) {
		return fmt.Errorf("%w: a proposal changes between one and %d targets", ErrInvalidProposal, len(model.Targets))
	}
	figures := make(map[string]*Figure, len(drift.Figures))
	for i := range drift.Figures {
		figures[drift.Figures[i].Target.Key()] = &drift.Figures[i]
	}
	seen := map[string]bool{}
	for i := range changes {
		c := &changes[i]
		if err := checkChange(c, figures[c.Target], seen); err != nil {
			return fmt.Errorf("change %d: %w", i+1, err)
		}
		seen[c.Target] = true
		fig := figures[c.Target]
		c.Held, c.Current, c.Drift = fig.Held, fig.Current, fig.Drift
	}
	if _, err := applyChanges(model, changes); err != nil {
		return err
	}
	return nil
}

// checkChange checks one change of a proposal against the figure of the
// target it names, nil when the model has none, and the targets already
// changed.
func checkChange(c *Change, fig *Figure, seen map[string]bool) error {
	switch {
	case fig == nil:
		return fmt.Errorf("%w: it names a target that is not in the model", ErrInvalidProposal)
	case seen[c.Target]:
		return fmt.Errorf("%w: a target changed twice", ErrInvalidProposal)
	case !c.Before.Equal(fig.Target.Weight):
		return fmt.Errorf("%w: the current weight is not the model's", ErrInvalidProposal)
	case c.After.Equal(c.Before):
		return fmt.Errorf("%w: it changes nothing", ErrInvalidProposal)
	case !oneLine(c.Reason):
		return fmt.Errorf("%w: a reason is one line of 1 to %d characters", ErrInvalidProposal, MaxReasonRunes)
	}
	return nil
}

// oneLine reports whether s is one line of 1 to MaxReasonRunes characters.
func oneLine(s string) bool {
	return s != "" && utf8.RuneCountInString(s) <= MaxReasonRunes && !strings.ContainsAny(s, "\r\n")
}

// Apply returns model with p's changes made: each target's weight is its
// change's After. It changes no other field, so a model that is off stays
// off. It is an error wrapping ErrStaleProposal when a target is gone or
// its weight is no longer the change's Before, and ErrInvalidProposal when
// the result is not a valid model (Model.Validate).
func (p *Proposal) Apply(model *Model) (Model, error) {
	for i := range p.Changes {
		c := &p.Changes[i]
		idx := targetIndex(model, c.Target)
		if idx < 0 || !model.Targets[idx].Weight.Equal(c.Before) {
			return Model{}, ErrStaleProposal
		}
	}
	return applyChanges(model, p.Changes)
}

// applyChanges returns model with changes applied, validated.
func applyChanges(model *Model, changes []Change) (Model, error) {
	out := Model{Enabled: model.Enabled, Targets: append([]Target(nil), model.Targets...)}
	for i := range changes {
		idx := targetIndex(&out, changes[i].Target)
		if idx < 0 {
			return Model{}, fmt.Errorf("%w: a change names a target that is not in the model", ErrInvalidProposal)
		}
		out.Targets[idx].Weight = changes[i].After
	}
	if err := out.Validate(); err != nil {
		return Model{}, fmt.Errorf("%w: the changed model is not valid: %w", ErrInvalidProposal, err)
	}
	return out, nil
}

// targetIndex returns the position of the target with key in model, or -1.
func targetIndex(model *Model, key string) int {
	for i := range model.Targets {
		if model.Targets[i].Key() == key {
			return i
		}
	}
	return -1
}

// Prompt is what the maintainer sends a model: the instructions, which
// hold no data of the portfolio (data tier T2), and the figures, which are
// whole-portfolio data (T0, D-206).
type Prompt struct {
	Instructions string
	Figures      string
}

// instructions are the system prompt: the role, its limits and the answer's
// shape. They quote no figure.
const instructions = "You maintain a model portfolio for a self-hosted portfolio tracker. " +
	"You are given each target of the model portfolio with its target weight, the weight the owner holds now and the drift between them, " +
	"all computed by the tracker; trust them and do not recompute them. " +
	"You cannot trade, and nothing you write trades anything: the owner reads your note and applies or dismisses any change you propose. " +
	"Write a short plain-text note, at most 120 words, saying what stands out in the drift. " +
	"Propose changes to target weights only when you have a clear reason; most weeks you should propose none. " +
	"A change moves the weight of a target already listed, never adds or removes one; every weight stays above 0 with at most two decimal places, " +
	"and the new weights of all targets together sum to exactly 100. " +
	"Give each change a one-line reason. Answer only with the JSON object of the schema."

// BuildPrompt returns the prompt for drift as of asOf, with every weight in
// percent of the whole portfolio's value in reporting. The figures are
// weights only: no amount, account or institution.
func BuildPrompt(drift *Drift, asOf kernel.Date, reporting kernel.Currency) Prompt {
	var b strings.Builder
	fmt.Fprintf(&b, "Portfolio as of %s. Weights are percent of the whole portfolio's value in %s.\n", asOf, reporting)
	b.WriteString("Targets, one per line: key | target weight | weight held now | drift in percentage points.\n")
	for i := range drift.Figures {
		f := &drift.Figures[i]
		if f.Held {
			fmt.Fprintf(&b, "%s | %s | %s | %s\n", f.Target.Key(), f.Target.Weight, f.Current, signed(f.Drift))
		} else {
			fmt.Fprintf(&b, "%s | %s | not held | -\n", f.Target.Key(), f.Target.Weight)
		}
	}
	fmt.Fprintf(&b, "Held outside the targets (cash and other holdings): %s\n", drift.Other)
	return Prompt{Instructions: instructions, Figures: b.String()}
}

// signed writes d with a plus sign when it is above zero.
func signed(d kernel.Decimal) string {
	if d.Sign() > 0 {
		return "+" + d.String()
	}
	return d.String()
}
