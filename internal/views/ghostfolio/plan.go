package ghostfolio

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
)

// Actual is what Ghostfolio holds now, as far as the view looks.
type Actual struct {
	// Accounts are the IDs of every account the Ghostfolio user has.
	Accounts []string
	// Activities are the activities in owned accounts. Any other is left
	// out of the plan.
	Activities []Activity
}

// Plan is what one sync changes in Ghostfolio, in the order Syncer applies
// it (D-64): create the missing owned accounts, delete the stale
// activities an import would otherwise skip as duplicates, import the
// missing activities with every asset's full price history, then delete
// the remaining stale activities, one by one or a whole owned account at a
// time.
type Plan struct {
	CreateAccounts []Account
	// PreDelete are stale activities that a missing activity duplicates in
	// every field Ghostfolio's import compares, such as one moved to
	// another owned account. Ghostfolio would skip that import, so these go
	// before it; the import then brings back any asset the delete removed,
	// with its prices.
	PreDelete []Activity
	Import    []Activity
	// Assets are always sent, so every sync sends the full price history.
	Assets []AssetProfile
	Delete []Activity
	// Clear are owned accounts that should hold no activity but do: each is
	// emptied with one bulk delete filtered to it.
	Clear []string
}

// IsEmpty reports whether p changes nothing but the price history.
func (p *Plan) IsEmpty() bool {
	return len(p.CreateAccounts) == 0 && len(p.PreDelete) == 0 && len(p.Import) == 0 &&
		len(p.Delete) == 0 && len(p.Clear) == 0
}

// MakePlan compares desired with actual and returns what to change (D-4,
// D-64). Only accounts in desired are owned: an activity anywhere else is
// never touched. An activity is kept when one in actual agrees with it in
// every field, comment included; a desired activity nothing agrees with is
// imported, and an activity in an owned account that matches no desired
// one is stale and deleted. So a manual edit is reverted: the edited
// activity is stale and the desired one is imported again.
func MakePlan(desired *Desired, actual *Actual) Plan {
	owned := map[string]bool{}
	for _, a := range desired.Accounts {
		owned[a.ID] = true
	}
	p := Plan{Assets: desired.Assets}
	for _, a := range desired.Accounts {
		if !slices.Contains(actual.Accounts, a.ID) {
			p.CreateAccounts = append(p.CreateAccounts, a)
		}
	}
	var acts []Activity
	for i := range actual.Activities {
		if owned[actual.Activities[i].AccountID] {
			acts = append(acts, actual.Activities[i])
		}
	}
	slices.SortFunc(acts, func(a, b Activity) int { return cmp.Compare(a.ID, b.ID) })
	kept := make([]bool, len(acts))
	wanted := map[string]bool{}
	for i := range desired.Activities {
		d := &desired.Activities[i]
		wanted[d.AccountID] = true
		if !keep(acts, kept, d) {
			p.Import = append(p.Import, *d)
		}
	}
	p.splitStale(acts, kept, wanted)
	return p
}

// keep marks the first activity in acts not yet kept that agrees with d in
// every field, and reports whether there was one.
func keep(acts []Activity, kept []bool, d *Activity) bool {
	for j := range acts {
		if !kept[j] && acts[j].sameAs(d) {
			kept[j] = true
			return true
		}
	}
	return false
}

// splitStale sorts the stale activities, those in acts not kept, into
// PreDelete, Clear and Delete. wanted holds the owned accounts with a
// desired activity.
func (p *Plan) splitStale(acts []Activity, kept []bool, wanted map[string]bool) {
	blocked := map[string]bool{}
	var stale []Activity
	for i := range acts {
		if kept[i] {
			continue
		}
		s := acts[i]
		if slices.ContainsFunc(p.Import, func(d Activity) bool { return d.duplicateOf(&s) }) {
			p.PreDelete = append(p.PreDelete, s)
			blocked[s.AccountID] = true
			continue
		}
		stale = append(stale, s)
	}
	emptied := map[string]bool{}
	for i := range stale {
		if id := stale[i].AccountID; !wanted[id] && !blocked[id] {
			emptied[id] = true
		}
	}
	p.Clear = slices.Sorted(maps.Keys(emptied))
	for i := range stale {
		if !emptied[stale[i].AccountID] {
			p.Delete = append(p.Delete, stale[i])
		}
	}
}

// ErrPreviewMismatch reports an import whose dry run differs from what the
// view sent: Ghostfolio would skip an activity as a duplicate, or store it
// under another account or symbol, as it does when the owner's own asset
// has the name of the view's asset.
var ErrPreviewMismatch = errors.New("the dry run of a Ghostfolio import differs from it")

// ImportedActivity is one activity of an import's response: the activity as
// Ghostfolio would store it, and whether it would skip it as a duplicate.
type ImportedActivity struct {
	Activity  Activity
	Duplicate bool
}

// CheckPreview returns nil when the dry run preview of an import of sent
// shows every activity stored as sent: one previewed activity per comment,
// none a duplicate, each in the same account under the same data source
// and symbol. Otherwise it returns an error wrapping ErrPreviewMismatch
// that names the first comment that differs.
func CheckPreview(sent []Activity, preview []ImportedActivity) error {
	if len(sent) != len(preview) {
		return fmt.Errorf("%w: %d activities sent, %d previewed", ErrPreviewMismatch, len(sent), len(preview))
	}
	byComment := map[string]*ImportedActivity{}
	for i := range preview {
		byComment[preview[i].Activity.Comment] = &preview[i]
	}
	for i := range sent {
		s := &sent[i]
		pa, ok := byComment[s.Comment]
		switch {
		case !ok:
			return fmt.Errorf("%w: %q is not in the preview", ErrPreviewMismatch, s.Comment)
		case pa.Duplicate:
			return fmt.Errorf("%w: %q would be skipped as a duplicate", ErrPreviewMismatch, s.Comment)
		case pa.Activity.AccountID != s.AccountID || pa.Activity.DataSource != s.DataSource || pa.Activity.Symbol != s.Symbol:
			return fmt.Errorf("%w: %q would be stored under another account or symbol", ErrPreviewMismatch, s.Comment)
		}
	}
	return nil
}

// Batch is one import: activities and the assets they need.
type Batch struct {
	Activities []Activity
	Assets     []AssetProfile
}

// Batches splits an import of acts with assets into imports of at most size
// activities each. Each batch carries the assets its activities use, since
// Ghostfolio creates a new asset only with an activity that uses it; the
// first batch also carries every asset no activity uses, so their prices
// are sent too. With no activities there is one batch of assets alone.
func Batches(acts []Activity, assets []AssetProfile, size int) []Batch {
	used := map[string]bool{}
	var out []Batch
	for start := 0; start < len(acts) || start == 0; start += size {
		end := min(start+size, len(acts))
		b := Batch{Activities: acts[start:end]}
		syms := map[string]bool{}
		for i := range b.Activities {
			syms[b.Activities[i].Symbol] = true
		}
		for _, as := range assets {
			if syms[as.Symbol] {
				b.Assets = append(b.Assets, as)
				used[as.Symbol] = true
			}
		}
		out = append(out, b)
		if end == len(acts) {
			break
		}
	}
	for _, as := range assets {
		if !used[as.Symbol] {
			out[0].Assets = append(out[0].Assets, as)
		}
	}
	return out
}
