package app_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/attention/app"
)

// errFake is what a fake answers when told to fail.
var errFake = errors.New("fake failure")

// memStore is app.Store and app.Outbox in memory. InTx works on a copy and
// keeps it only when fn succeeds, like a database transaction; every
// appended change gets its event, as the PostgreSQL store writes it.
type memStore struct {
	mu        *sync.Mutex
	items     []attention.Item
	events    []app.Event
	published map[int64]time.Time
	// failAppend makes Append fail after storing the change.
	failAppend bool
}

func newMemStore() *memStore {
	return &memStore{mu: &sync.Mutex{}, published: map[int64]time.Time{}}
}

func (m *memStore) InTx(_ context.Context, fn func(tx app.Store) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	tx := &memStore{mu: &sync.Mutex{}, items: cloneItems(m.items), events: slices.Clone(m.events), published: m.published, failAppend: m.failAppend}
	if err := fn(tx); err != nil {
		return err
	}
	m.items, m.events = tx.items, tx.events
	return nil
}

func (*memStore) Lock(context.Context) error { return nil }

func (m *memStore) Items(_ context.Context, kind attention.Kind) ([]attention.Item, error) {
	var out []attention.Item
	for _, it := range cloneItems(m.items) {
		if kind == "" || it.Kind == kind {
			out = append(out, it)
		}
	}
	return out, nil
}

func (m *memStore) Item(_ context.Context, id int64) (attention.Item, bool, error) {
	for _, it := range cloneItems(m.items) {
		if it.ID == id {
			return it, true, nil
		}
	}
	return attention.Item{}, false, nil
}

func (m *memStore) Append(_ context.Context, step *attention.Step) (int64, error) {
	id := step.ItemID
	if id == 0 {
		id = int64(len(m.items) + 1)
		m.items = append(m.items, attention.Item{ID: id, Kind: step.Kind, Subject: step.Subject})
	}
	m.items[id-1].Changes = append(m.items[id-1].Changes, step.Change)
	if m.failAppend {
		return 0, errFake
	}
	m.events = append(m.events, app.Event{ID: int64(len(m.events) + 1), ItemID: id, Kind: step.Kind, Subject: step.Subject, Change: step.Change})
	return id, nil
}

func (m *memStore) Unpublished(_ context.Context, limit int) ([]app.Event, error) {
	var out []app.Event
	for _, e := range m.events {
		if _, ok := m.published[e.ID]; !ok && len(out) < limit {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *memStore) MarkPublished(_ context.Context, id int64, at time.Time) error {
	m.published[id] = at
	return nil
}

// cloneItems deep-copies items.
func cloneItems(items []attention.Item) []attention.Item {
	out := make([]attention.Item, len(items))
	for i, it := range items {
		it.Changes = slices.Clone(it.Changes)
		out[i] = it
	}
	return out
}

// fakeOwners accepts the session token "owner" only; "reader" is a read
// session.
type fakeOwners struct{}

func (fakeOwners) CheckOwner(_ context.Context, _, token string) error {
	switch token {
	case "owner":
		return nil
	case "reader":
		return app.ErrNotOwner
	}
	return app.ErrNoSession
}

// recorder is an app.Publisher that keeps what it got, failing from the
// failAt-th event on when failAt is set.
type recorder struct {
	got    []int64
	failAt int
}

func (r *recorder) Publish(_ context.Context, e *app.Event) error {
	if r.failAt > 0 && len(r.got)+1 >= r.failAt {
		return errFake
	}
	r.got = append(r.got, e.ID)
	return nil
}
