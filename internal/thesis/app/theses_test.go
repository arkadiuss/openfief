package app_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/thesis"
	"github.com/arkadiuss/openfief/internal/thesis/app"
)

// memStore is an in-memory app.Store with the versions' rules of the real one.
type memStore struct {
	mu       sync.Mutex
	versions map[string][]thesis.Version
	order    []string
}

func (m *memStore) Save(ctx context.Context, id string, base int, at time.Time, d *thesis.Draft, within func(context.Context) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	have := m.versions[id]
	if err := versionsAgree(base, len(have)); err != nil {
		return err
	}
	if within != nil {
		if err := within(ctx); err != nil {
			return err
		}
	}
	if m.versions == nil {
		m.versions = map[string][]thesis.Version{}
	}
	if base == 0 {
		m.order = append(m.order, id)
	}
	m.versions[id] = append(have, thesis.Version{Number: base + 1, SavedAt: at, Draft: *d})
	return nil
}

// versionsAgree is the real store's rule: a new thesis has no versions, an edit
// is made on the latest one.
func versionsAgree(base, have int) error {
	switch {
	case base == 0 && have > 0, base > 0 && have != base && have > 0:
		return thesis.ErrStale
	case base > 0 && have == 0:
		return thesis.ErrNotFound
	}
	return nil
}

func (m *memStore) Get(_ context.Context, id string) (thesis.Thesis, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.versions[id]) == 0 {
		return thesis.Thesis{}, thesis.ErrNotFound
	}
	return thesis.Thesis{ID: id, Versions: slices.Clone(m.versions[id])}, nil
}

func (m *memStore) List(context.Context) ([]thesis.Summary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []thesis.Summary
	for _, id := range slices.Backward(m.order) {
		vs := m.versions[id]
		out = append(out, thesis.Summary{ID: id, Version: len(vs), CreatedAt: vs[0].SavedAt, UpdatedAt: vs[len(vs)-1].SavedAt, Draft: vs[len(vs)-1].Draft})
	}
	return out, nil
}

// owners admits the session token "owner" only.
type owners struct{}

func (owners) OwnerSession(_ context.Context, _, token string) (string, error) {
	switch token {
	case "owner":
		return "aaaaaaaaaaaaaaaa", nil
	case "read":
		return "", app.ErrNotOwner
	}
	return "", app.ErrNoSession
}

// audit records the changes it is asked to, or fails.
type audit struct {
	got  []app.Change
	fail error
}

func (a *audit) RecordChange(_ context.Context, c *app.Change) error {
	if a.fail != nil {
		return a.fail
	}
	a.got = append(a.got, *c)
	return nil
}

// day is the time every change is stamped with.
func day() time.Time { return time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC) }

func setup() (*app.Theses, *memStore, *audit) {
	store, au := &memStore{}, &audit{}
	n := 0
	ids := func() (string, error) {
		n++
		return fmt.Sprintf("%016x", n), nil
	}
	return app.New(store, owners{}, au, ids, day), store, au
}

func draft() *thesis.Draft {
	return &thesis.Draft{
		Statement: "Services revenue keeps growing.", Instruments: []string{"EXMP"}, Horizon: thesis.HorizonLongTerm,
		Invalidation: []string{"Growth stalls for two quarters."}, Sources: []string{"The holdings card."},
	}
}

func TestSaveAndEditKeepEveryVersionAndAuditEach(t *testing.T) {
	t.Parallel()
	theses, _, au := setup()
	ctx := t.Context()

	saved, err := theses.Save(ctx, "o", "owner", draft())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(saved.Versions) != 1 || saved.Versions[0].Number != 1 {
		t.Fatalf("Save() versions = %+v, want version 1 only", saved.Versions)
	}
	edit := draft()
	edit.Statement = "Services revenue keeps growing, and margins widen."
	edited, err := theses.Edit(ctx, "o", "owner", saved.ID, 1, edit)
	if err != nil {
		t.Fatalf("Edit() error = %v", err)
	}
	if len(edited.Versions) != 2 || edited.Latest().Draft.Statement != edit.Statement || edited.Versions[0].Draft.Statement != draft().Statement {
		t.Fatalf("Edit() versions = %+v, want the new text on top of the old", edited.Versions)
	}
	want := []app.Change{
		{At: day(), Session: "aaaaaaaaaaaaaaaa", Thesis: saved.ID, Version: 1, Edited: false},
		{At: day(), Session: "aaaaaaaaaaaaaaaa", Thesis: saved.ID, Version: 2, Edited: true},
	}
	if diff := cmp.Diff(want, au.got); diff != "" {
		t.Errorf("audit entries mismatch (-want +got):\n%s", diff)
	}
}

func TestEditRefusals(t *testing.T) {
	t.Parallel()
	theses, store, au := setup()
	ctx := t.Context()
	saved, err := theses.Save(ctx, "o", "owner", draft())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := theses.Edit(ctx, "o", "owner", saved.ID, 1, draft()); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}
	broken := draft()
	broken.Horizon = "medium_term"
	tests := []struct {
		name string
		id   string
		base int
		d    *thesis.Draft
		want error
	}{
		{"an edit of a version that is no longer the latest", saved.ID, 1, draft(), thesis.ErrStale},
		{"an edit of a thesis that does not exist", "ffffffffffffffff", 1, draft(), thesis.ErrNotFound},
		{"an id of the wrong shape", "../etc/passwd", 1, draft(), thesis.ErrNotFound},
		{"a base below 1", saved.ID, 0, draft(), thesis.ErrNotFound},
		{"a draft that breaks a rule", saved.ID, 2, broken, thesis.ErrInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			before, entries := len(store.versions[saved.ID]), len(au.got)
			if _, err := theses.Edit(ctx, "o", "owner", tt.id, tt.base, tt.d); !errors.Is(err, tt.want) {
				t.Fatalf("Edit() error = %v, want %v", err, tt.want)
			}
			if len(store.versions[saved.ID]) != before || len(au.got) != entries {
				t.Errorf("a refused edit stored or audited something")
			}
		})
	}
}

func TestAnEntryTheLogCannotTakeKeepsTheVersionOut(t *testing.T) {
	t.Parallel()
	theses, store, au := setup()
	au.fail = kernel.ErrAuditUnavailable
	_, err := theses.Save(t.Context(), "o", "owner", draft())
	if !errors.Is(err, kernel.ErrAuditUnavailable) {
		t.Fatalf("Save() error = %v, want ErrAuditUnavailable", err)
	}
	if len(store.versions) != 0 {
		t.Errorf("a save whose entry failed was stored")
	}
}

func TestEveryCallIsTheOwnersAlone(t *testing.T) {
	t.Parallel()
	theses, _, _ := setup()
	ctx := t.Context()
	saved, err := theses.Save(ctx, "o", "owner", draft())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	calls := map[string]func(token string) error{
		"Save": func(tok string) error { return wrap(theses.Save(ctx, "o", tok, draft())) },
		"Edit": func(tok string) error { return wrap(theses.Edit(ctx, "o", tok, saved.ID, 1, draft())) },
		"Get":  func(tok string) error { return wrap(theses.Get(ctx, "o", tok, saved.ID)) },
		"List": func(tok string) error { return wrap(theses.List(ctx, "o", tok)) },
	}
	for name, call := range calls {
		t.Run(name+" refuses a read-scoped session", func(t *testing.T) {
			t.Parallel()
			if err := call("read"); !errors.Is(err, app.ErrNotOwner) {
				t.Errorf("%s() error = %v, want ErrNotOwner", name, err)
			}
		})
		t.Run(name+" refuses no session", func(t *testing.T) {
			t.Parallel()
			if err := call("nobody"); !errors.Is(err, app.ErrNoSession) {
				t.Errorf("%s() error = %v, want ErrNoSession", name, err)
			}
		})
	}
}

// wrap returns the error of a call, wrapped, whatever it returned.
func wrap[T any](_ T, err error) error {
	if err != nil {
		return fmt.Errorf("the call: %w", err)
	}
	return nil
}

func TestGetAndList(t *testing.T) {
	t.Parallel()
	theses, _, _ := setup()
	ctx := t.Context()
	a, err := theses.Save(ctx, "o", "owner", draft())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	other := draft()
	other.Statement = "Another claim."
	b, err := theses.Save(ctx, "o", "owner", other)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	list, err := theses.List(ctx, "o", "owner")
	if err != nil || len(list) != 2 || list[0].ID != b.ID || list[1].ID != a.ID {
		t.Fatalf("List() = %+v, %v, want the newest first", list, err)
	}
	if _, err := theses.Get(ctx, "o", "owner", "ffffffffffffffff"); !errors.Is(err, thesis.ErrNotFound) {
		t.Errorf("Get(unknown) error = %v, want ErrNotFound", err)
	}
}
